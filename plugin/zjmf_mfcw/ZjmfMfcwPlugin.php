<?php
namespace certification\zjmf_mfcw;

use app\admin\lib\Plugin;
use certification\zjmf_mfcw\logic\KycSdk;

/**
 * StarLoft KYC 实名认证插件
 *
 * 对接 StarLoft KYC 系统进行实名认证
 * 支持身份证三要素认证（姓名+身份证号+人脸识别）
 *
 * 修复要点:
 *  - personal(): 先查用户当前认证记录,若 status=4(认证中)且有 certify_id,直接复用原任务/查询后返回,不再重复创建扣费
 *  - personal(): 对 startKyc 的错误做精准分类(余额不足/签名错误/参数错误),余额不足直接中止,防止用户反复重试导致多扣
 *  - getStatus(): 对 queryResult 的错误做精准分类,余额不足/鉴权失败/参数错误 直接终态失败(status=2),不再"查询中..."无限轮询
 *  - getStatus(): 连续出现 NOT_FOUND 超过阈值后也终态失败,避免 certify_id 错误时卡死
 *  - getStatus(): 轮询计数改用文件存储(不再依赖 SESSION),并设置硬性上限,
 *                 到上限后强制返回 status=2(终态失败),杜绝平台系统以 2s/次无限轮询回调地址
 *
 * @author StarLoft
 * @version 1.1.1
 */
class ZjmfMfcwPlugin extends Plugin
{
    /**
     * 插件基本信息
     */
    public $info = [
        'name'        => 'ZjmfMfcw',
        'title'       => 'StarLoft KYC实名认证',
        'description' => 'StarLoft KYC 三要素实名认证（姓名+身份证+人脸识别）— 已修复重复创建/无限轮询问题',
        'status'      => 1,
        'author'      => 'StarLoft',
        'version'     => '1.1.1',
        'help_url'    => 'https://docs.starloft.cn/kyc/plugin'
    ];

    /** 轮询最大次数(系统按 2s/次轮询, 500 次约 16 分钟, 覆盖上游任务 15 分钟有效期)*/
    const MAX_POLL_COUNT = 500;

    /** 查询出错(网络/临时错误)时的轮询上限, 60 次约 2 分钟, 防止 KYC 后端不可达时无限轮询 */
    const MAX_ERROR_COUNT = 60;

    /** NOT_FOUND 连续多少次后判终态失败(防止 certify_id 写错或已被平台清理) */
    const MAX_NOT_FOUND = 5;

    public function install()
    {
        return true;
    }

    public function uninstall()
    {
        return true;
    }

    /**
     * 个人实名认证入口
     */
    public function personal($certifi)
    {
        try {
            $config = $this->getConfig();
            $sdk    = new KycSdk($config);

            $name   = trim($certifi['name']   ?? '');
            $idCard = trim($certifi['card']   ?? '');
            if ($name === '' || $idCard === '') {
                return $this->failHtml('姓名和身份证号不能为空,请返回重新填写');
            }

            // ============ 年龄限制校验(可配置 min_age,默认 16 周岁) ============
            $minAge = intval($config['min_age'] ?? 16);
            if ($minAge > 0) {
                $age = $this->getAgeFromIdCard($idCard);
                if ($age !== null && $age < $minAge) {
                    $msg = "根据身份证信息，您未满 {$minAge} 周岁，暂不支持实名认证";
                    updatePersonalCertifiStatus([
                        'status'     => 2,
                        'auth_fail'  => $msg,
                        'certify_id' => '',
                        'notes'      => $msg,
                    ]);
                    return $this->failHtml($msg);
                }
            }

            // ============ 修复1: 幂等检查 — 先拿当前用户已有的实名记录 ============
            $existing = $this->getCurrentUserCertiRecord();
            if (!empty($existing)) {
                $exStatus    = (int)($existing['certify_status'] ?? $existing['status'] ?? 0);
                $exCertifyId = trim((string)($existing['certify_id'] ?? ''));
                $exNotes     = (string)($existing['notes'] ?? '');

                // status=4 表示已经在认证中,且有平台任务号 -> 直接复用,不重复创建扣费
                if ($exStatus === 4 && $exCertifyId !== '') {
                    $needRecreate = false;
                    // 先尝试查一下这个任务号还活不活着(可能 KYC 平台侧任务已过期/已出结果但魔方没收到通知)
                    $query = $sdk->queryResult(['platform_biz_no' => $exCertifyId]);
                    $cat   = KycSdk::classifyError($query);

                    if ($cat === KycSdk::ERR_CAT_SUCCESS) {
                        $st = (int)($query['data']['status'] ?? 0);
                        if ($st === 2) {
                            // 已经成功过了,直接写终态 + 显示成功
                            updatePersonalCertifiStatus([
                                'status'     => 1,
                                'auth_fail'  => '',
                                'certify_id' => $exCertifyId,
                            ]);
                            return $this->successHtml();
                        }
                        if ($st === 3) {
                            $msg = $query['data']['result_message'] ?? '实名认证失败';
                            updatePersonalCertifiStatus([
                                'status'     => 2,
                                'auth_fail'  => $msg,
                                'certify_id' => $exCertifyId,
                            ]);
                            return $this->failHtml($msg);
                        }
                        // st===1 仍在认证中 -> 继续走下面复用
                    } elseif ($cat === KycSdk::ERR_CAT_NO_BALANCE
                              || $cat === KycSdk::ERR_CAT_AUTH
                              || $cat === KycSdk::ERR_CAT_NOT_FOUND) {
                        // 任务已不存在/鉴权失败/余额不足 -> 历史任务不可用,需要重建
                        $needRecreate = true;
                    } else {
                        // 其它临时错误/未知 -> 倾向复用,先继续用原来的任务号轮询
                    }

                    if (!$needRecreate) {
                        // 复用任务: 返回跳转 HTML(但 auth_url 不再从新任务拿,而是回显一个"继续轮询"的页)
                        return $this->continuePollHtml($exCertifyId, '您有一份正在进行的实名任务,已为您恢复进度,正在等待核验结果...');
                    }
                    // 需要重建,继续往下
                }
            }

            // ============ 真正创建新任务 ============
            $uid   = $this->resolveCurrentUid();
            // 稳定幂等业务单号：同一用户+同一身份证固定生成，后端据此去重，避免重复创建任务/重复扣费
            $bizNo = 'ZJMF_' . $uid . '_' . substr(md5($idCard), 0, 8);
            $domain = $this->resolveDomain();
            $notifyUrl = $domain . '/certification/zjmf_mfcw/callback?uid=' . $uid;
            $returnUrl = $domain . '/certification/zjmf_mfcw/result?uid=' . $uid;

            $result = $sdk->startKyc([
                'biz_no'        => $bizNo,
                'name'          => $name,
                'id_card'       => $idCard,
                'return_url'    => $returnUrl,
                'notify_url'    => $notifyUrl,
                'biz_extra_data'=> json_encode(['uid' => $uid]),
            ]);

            $cat = KycSdk::classifyError($result);

            // ============ 修复2: startKyc 错误精准分类,余额不足等硬错误立刻终止 ============
            if ($cat !== KycSdk::ERR_CAT_SUCCESS) {
                $msg = (string)($result['message'] ?? '实名认证接口请求失败,请稍后再试');

                switch ($cat) {
                    case KycSdk::ERR_CAT_NO_BALANCE:
                        $displayMsg = '实名平台商户余额/额度不足,请联系管理员充值后重试。(返回:' . $msg . ')';
                        break;
                    case KycSdk::ERR_CAT_AUTH:
                        $displayMsg = '实名平台鉴权失败(AppKey/签名配置错误),请联系管理员检查插件配置。(返回:' . $msg . ')';
                        break;
                    case KycSdk::ERR_CAT_PARAM:
                        $displayMsg = '实名信息有误:' . $msg;
                        break;
                    case KycSdk::ERR_CAT_TEMP:
                    default:
                        $displayMsg = '实名接口暂时不可用,请稍后再试。(' . $msg . ')';
                        break;
                }

                // 对硬错误直接置 status=2 失败,避免用户反复重试触发多次请求
                $failHard = in_array($cat, [KycSdk::ERR_CAT_NO_BALANCE, KycSdk::ERR_CAT_AUTH, KycSdk::ERR_CAT_PARAM], true);
                $data = [
                    'status'    => $failHard ? 2 : 4, // 临时错误 -> 保留"认证中",让用户稍后再试不至于立刻失败
                    'auth_fail' => $displayMsg,
                    'certify_id'=> '',
                    'notes'     => "startKyc错误[{$cat}]: {$msg}\n时间: " . date('Y-m-d H:i:s'),
                ];
                updatePersonalCertifiStatus($data);
                return $this->failHtml($displayMsg);
            }

            // 创建成功 -> 写库
            $orderData = $result['data'];
            $data = [
                'status'     => 4,
                'auth_fail'  => '',
                'certify_id' => $orderData['platform_biz_no'],
                'notes'      => "KYC平台流水号: {$orderData['platform_biz_no']}\n业务订单号: {$bizNo}\n创建时间: " . date('Y-m-d H:i:s'),
            ];
            updatePersonalCertifiStatus($data);

            $authUrl = $orderData['auth_url'];
            return <<<HTML
<div class="kyc-auth-container" style="text-align: center; padding: 20px;">
    <h5 class="pt-2 font-weight-bold h5 py-4">正在跳转到实名认证页面...</h5>
    <p>如果页面未自动跳转,请点击下方按钮(<b>请勿反复刷新或多次点击</b>,以免重复创建实名任务扣费)</p>
    <a href="{$authUrl}" class="btn btn-primary" target="_blank">前往认证</a>
    <script>
        (function(){
            var opened = false;
            function tryOpen(){ if(!opened){ opened=true; window.open('{$authUrl}', '_blank'); } }
            setTimeout(tryOpen, 800);
            // 兜底: 点击任何空白处再开一次,避免浏览器拦截
            document.addEventListener && document.addEventListener('click', function once(){
                tryOpen();
                document.removeEventListener('click', once);
            }, {once:true});
        })();
    </script>
</div>
HTML;

        } catch (\Exception $e) {
            $errorMsg = '系统错误: ' . $e->getMessage();
            $data = [
                'status'    => 4, // 系统异常不直接判失败,保留下次重试空间
                'auth_fail' => $errorMsg,
                'certify_id'=> '',
                'notes'     => 'Exception: ' . $e->getMessage() . "\n" . $e->getFile() . ':' . $e->getLine(),
            ];
            try { updatePersonalCertifiStatus($data); } catch (\Throwable $_) {}
            return $this->failHtml($errorMsg);
        }
    }

    /**
     * 企业实名认证（暂不支持）
     */
    public function company($certifi)
    {
        $data = [
            'status'     => 2,
            'auth_fail'  => '当前版本暂不支持企业认证',
            'certify_id' => '',
            'notes'      => '企业认证功能开发中',
        ];
        updateCompanyCertifiStatus($data);

        return "<h3 class=\"pt-2 font-weight-bold h2 py-4\" style=\"color: #e6a23c;\">
            <i class=\"fa fa-info-circle\"></i> 当前版本暂不支持企业认证
        </h3>";
    }

    public function collectionInfo()
    {
        return [];
    }

    /**
     * 查询认证状态(前端轮询入口)
     *
     * 核心修复:
     *  - 不再一视同仁返回"status=4 查询中...": 余额不足/鉴权错误等硬错误立刻结束(status=2)
     *  - NOT_FOUND 连续超阈值后也结束,避免 certify_id 错误死循环
     *  - 轮询次数上限(超上限给用户提示)
     */
    public function getStatus($certifi)
    {
        $certifyId = trim((string)($certifi['certify_id'] ?? ''));

        try {
            $config = $this->getConfig();
            $sdk    = new KycSdk($config);
            $result = $sdk->queryResult(['platform_biz_no' => $certifyId]);
            $cat    = KycSdk::classifyError($result);

            // --- 成功响应(顶层 result_code=1000/SUCCESS) -> 按 data.result_code(订单自身状态)分 ---
            // 订单 code 读取顺序: data.result_code / data.status_code / data.code(SDK 已为我们做了别名双写)
            if ($cat === KycSdk::ERR_CAT_SUCCESS) {
                $orderData    = is_array($result['data'] ?? null) ? $result['data'] : [];
                $orderCode    = (int)($orderData['result_code'] ?? $orderData['status_code'] ?? $orderData['code'] ?? 0);
                $orderMessage = (string)($orderData['result_message'] ?? $orderData['message'] ?? '');

                // 1000 SUCCESS = 验证通过 -> status=1 已认证
                if ($orderCode === 1000 || $orderMessage === 'SUCCESS') {
                    return ['status' => 1, 'msg' => '实名认证通过'];
                }
                // 2000/3000(计费的 5 种)/4000 = 终态不通过 -> status=2
                $rejectMsgs = [
                    'PASS_LIVING_NOT_THE_SAME',
                    'NO_ID_CARD_NUMBER','ID_NUMBER_NAME_NOT_MATCH','NO_FACE_FOUND','NO_ID_PHOTO','PHOTO_FORMAT_ERROR',
                    'FAIL_LIVING_FACE_ATTACK',
                    'FAILED','CANCELLED','TIMEOUT',
                ];
                if (in_array($orderCode, [2000, 4000], true)
                    || ($orderCode === 3000 && in_array($orderMessage, $rejectMsgs, true))
                    || ($orderCode === 6000 && in_array($orderMessage, ['FAILED','CANCELLED','TIMEOUT'], true))) {
                    $failMsg = $this->translateResultMsg($orderMessage) ?: ($orderData['result_message'] ?? '实名认证失败');
                    updatePersonalCertifiStatus([
                        'status'     => 2,
                        'auth_fail'  => $failMsg,
                        'certify_id' => $certifyId,
                    ]);
                    return ['status' => 2, 'msg' => $failMsg];
                }
                // 6000 NOT_STARTED / PROCESSING = 还在进行中
                if ($orderCode === 6000 && in_array($orderMessage, ['NOT_STARTED','PROCESSING'], true)) {
                    $tipMap = [
                        'NOT_STARTED' => '等待您打开认证页面完成操作...',
                        'PROCESSING'  => '认证处理中,请稍候...',
                    ];
                    return $this->trackPollAndReturn($certifyId, 4, $tipMap[$orderMessage] ?? '认证处理中,请稍候...');
                }
                // 6100 = webRTC/权限问题(提示用户检查浏览器权限)
                if ($orderCode === 6100) {
                    $tipMap = [
                        'SUPPORT_ERROR'     => '当前浏览器不支持人脸核身(需要支持 webRTC),请改用 Chrome/Edge 最新版后重试。',
                        'PERMISSIONS_ERROR' => '摄像头权限被拒绝,请允许浏览器使用摄像头后刷新页面重新进入。',
                        'OTHER_ERROR'       => 'webRTC 连接异常,请检查网络/摄像头后刷新重试。',
                    ];
                    $tip = $tipMap[$orderMessage] ?? '认证遇到问题,请检查浏览器摄像头权限后重试。';
                    return $this->trackPollAndReturn($certifyId, 4, $tip);
                }
                // 3000 DATA_SOURCE_ERROR/INTERNAL_ERROR = 临时错误,继续等(出错路径用较短上限,防止无限轮询)
                if ($orderCode === 3000 && in_array($orderMessage, ['DATA_SOURCE_ERROR','INTERNAL_ERROR'], true)) {
                    return $this->trackPollAndReturn($certifyId, 4, '服务端临时异常,持续重试中...(' . $orderMessage . ')', 'err', self::MAX_ERROR_COUNT);
                }
                // 兜底 -> 继续轮询
                return $this->trackPollAndReturn($certifyId, 4, '认证处理中,请稍候...(order_code=' . $orderCode . ')');
            }

            // --- 非成功响应,按错误分类分 ---
            $msg = (string)($result['message'] ?? $result['result_message'] ?? '查询失败');

            // 新增:实名"业务不通过/终态失败"(活体攻击/证号不匹配/已取消/超时等)-> 直接 status=2 结束
            if ($cat === KycSdk::ERR_CAT_REJECT) {
                $failMsg = $this->translateResultMsg((string)($result['result_message'] ?? '')) ?: $msg;
                updatePersonalCertifiStatus([
                    'status'     => 2,
                    'auth_fail'  => '实名认证未通过: ' . $failMsg,
                    'certify_id' => $certifyId,
                ]);
                return ['status' => 2, 'msg' => '实名认证未通过: ' . $failMsg];
            }
            // 新增:用户侧可恢复(未开始/进行中/不支持webRTC/权限被拒) -> 保持 4 继续轮询
            if ($cat === KycSdk::ERR_CAT_USER_ACTION) {
                $rm = (string)($result['result_message'] ?? '');
                $tipMap = [
                    'NOT_STARTED'       => '等待您打开认证页面完成操作...',
                    'PROCESSING'        => '认证处理中,请稍候...',
                    'SUPPORT_ERROR'     => '当前浏览器不支持人脸核身(需要支持 webRTC),请改用 Chrome/Edge 最新版后重试。',
                    'PERMISSIONS_ERROR' => '摄像头权限被拒绝,请允许浏览器使用摄像头后刷新页面重新进入。',
                    'OTHER_ERROR'       => 'webRTC 连接异常,请检查网络/摄像头后刷新重试。',
                ];
                $tip = $tipMap[$rm] ?? ($msg ?: '认证处理中,请稍候...');
                return $this->trackPollAndReturn($certifyId, 4, $tip);
            }

            // 修复: 余额不足 / 鉴权配置错 / 参数错 => 终态失败,不再继续轮询!
            if ($cat === KycSdk::ERR_CAT_NO_BALANCE) {
                // 余额不足时同步写库为终态失败,避免下次再创建
                updatePersonalCertifiStatus([
                    'status'     => 2,
                    'auth_fail'  => '实名平台商户余额/额度不足,请联系管理员充值后重试。(' . $msg . ')',
                    'certify_id' => $certifyId,
                ]);
                return ['status' => 2, 'msg' => '实名平台商户余额/额度不足,请联系管理员充值后重试。'];
            }
            if ($cat === KycSdk::ERR_CAT_AUTH) {
                updatePersonalCertifiStatus([
                    'status'     => 2,
                    'auth_fail'  => '实名平台鉴权失败,请联系管理员检查配置。(' . $msg . ')',
                    'certify_id' => $certifyId,
                ]);
                return ['status' => 2, 'msg' => '实名平台鉴权失败,请联系管理员检查配置。'];
            }
            if ($cat === KycSdk::ERR_CAT_PARAM) {
                updatePersonalCertifiStatus([
                    'status'     => 2,
                    'auth_fail'  => '实名参数错误: ' . $msg,
                    'certify_id' => $certifyId,
                ]);
                return ['status' => 2, 'msg' => '实名参数错误: ' . $msg];
            }

            // NOT_FOUND: 可能 certify_id 传错了 / 任务已过期被清理 -> 连续计数,超阈值判终态
            if ($cat === KycSdk::ERR_CAT_NOT_FOUND) {
                $cnt = $this->incPollCounter($certifyId, 'notfound');
                if ($cnt >= self::MAX_NOT_FOUND) {
                    updatePersonalCertifiStatus([
                        'status'     => 2,
                        'auth_fail'  => '查询不到实名任务(已超过重试次数),请稍后重新发起实名。(' . $msg . ')',
                        'certify_id' => $certifyId,
                    ]);
                    return ['status' => 2, 'msg' => '查询不到实名任务,请稍后重新发起实名。'];
                }
                return ['status' => 4, 'msg' => '任务查询中,请稍候...(未找到,剩余重试 ' . (self::MAX_NOT_FOUND - $cnt) . ')'];
            }

            // 临时错误 / 未知: 保持 4(继续轮询),但用较短上限强制终止,避免后端不可达时无限轮询
            return $this->trackPollAndReturn($certifyId, 4, '查询中,请稍候...(' . $msg . ')', 'err', self::MAX_ERROR_COUNT);

        } catch (\Exception $e) {
            // 本地异常: 短暂继续轮询,超过上限强制终态失败
            return $this->trackPollAndReturn($certifyId, 4, '查询中,请稍候...(异常:' . $e->getMessage() . ')', 'err', self::MAX_ERROR_COUNT);
        }
    }

    // =========================================================
    // 辅助: 取当前用户实名记录(魔方财务无标准方法,用 Db 兜底查)
    //   - 表名依次覆盖:认证系统最常见的 im_host_user_certification / im_certification_personal
    //     以及不带 im_ 前缀的老版本;另外用 Db::query('SHOW TABLES') 做一次模糊匹配兜底
    // =========================================================
    protected function getCurrentUserCertiRecord()
    {
        try {
            $uid = $this->resolveCurrentUid();
            if ($uid <= 0) return [];

            if (class_exists('think\Db')) {
                // 候选表名(覆盖魔方财务/顺戴财务常见惯例)
                $tables = [
                    'im_host_user_certification',
                    'host_user_certification',
                    'im_certification_personal',
                    'certification_personal',
                    'im_user_certification',
                    'user_certification',
                    'certification_user_personal',
                ];
                foreach ($tables as $tbl) {
                    try {
                        $row = \think\Db::name($tbl)
                            ->where('uid', $uid)
                            ->order('id desc')
                            ->find();
                        if ($row) return $row;
                    } catch (\Throwable $_) { /* 表不存在,试下一个 */ }
                }

                // 表名全猜不中,用 SHOW TABLES LIKE 兜底搜一次
                try {
                    $like = \think\Db::query("SHOW TABLES LIKE '%certif%'");
                    if (!empty($like)) {
                        foreach ($like as $r) {
                            $tblName = array_values($r)[0] ?? '';
                            if ($tblName === '') continue;
                            try {
                                $row = \think\Db::name($tblName)
                                    ->where('uid', $uid)
                                    ->order('id desc')
                                    ->find();
                                if ($row) return $row;
                            } catch (\Throwable $_) {}
                        }
                    }
                } catch (\Throwable $_) {}
            }
        } catch (\Throwable $_) {}
        return [];
    }

    /**
     * 根据18位身份证号计算周岁年龄
     *
     * @param string $idCard 身份证号
     * @return int|null 年龄；身份证号格式不合法时返回 null
     */
    protected function getAgeFromIdCard($idCard)
    {
        if (!is_string($idCard) || strlen($idCard) !== 18) {
            return null;
        }

        $year  = (int)substr($idCard, 6, 4);
        $month = (int)substr($idCard, 10, 2);
        $day   = (int)substr($idCard, 12, 2);

        if ($year < 1900 || !checkdate($month, $day, $year)) {
            return null;
        }

        $age = (int)date('Y') - $year;
        $birthMd = sprintf('%02d%02d', $month, $day);
        if (date('md') < $birthMd) {
            $age--;
        }
        return $age;
    }

    /**
     * 魔方财务通用取 UID:
     *  - 优先 \think\Session::get('user_id') / session('uid') / session('user_id')
     *  - 再 request()->uid / input('uid/d')
     *  - 再 auth('user')->id(TP5 auth 惯例)
     * 至少一个成功就不会让防重检查空跑
     */
    protected function resolveCurrentUid()
    {
        $uid = 0;
        try {
            if (class_exists('think\Session')) {
                $uid = (int)\think\Session::get('user_id');
                if ($uid <= 0) $uid = (int)\think\Session::get('uid');
                if ($uid <= 0) $uid = (int)\think\Session::get('user.id');
            }
            if ($uid <= 0 && isset($_SESSION)) {
                $uid = (int)($_SESSION['user_id'] ?? $_SESSION['uid'] ?? 0);
            }
            if ($uid <= 0 && function_exists('session')) {
                $uid = (int)(session('user_id') ?: session('uid') ?: 0);
            }
            if ($uid <= 0) {
                $uid = (int)(\request()->uid ?? input('uid/d', 0));
            }
            if ($uid <= 0 && function_exists('auth')) {
                try {
                    $user = auth('user');
                    if (is_object($user) && isset($user->id)) $uid = (int)$user->id;
                } catch (\Throwable $_) {}
            }
        } catch (\Throwable $_) {}
        return $uid;
    }

    /**
     * 安全的获取当前域名(先取 HTTP_X_FORWARDED_PROTO/HTTPS 再拼 SERVER_NAME)
     * 避免 CLI 或某些异常场景下 request()->domain() 直接 Fatal
     */
    protected function resolveDomain()
    {
        try {
            if (function_exists('request') && is_object(request())) {
                $dom = request()->domain();
                if ($dom) return rtrim($dom, '/');
            }
        } catch (\Throwable $_) {}
        $scheme = 'http';
        if (!empty($_SERVER['HTTPS']) && strtolower($_SERVER['HTTPS']) !== 'off') $scheme = 'https';
        if (!empty($_SERVER['HTTP_X_FORWARDED_PROTO']) && strtolower($_SERVER['HTTP_X_FORWARDED_PROTO']) === 'https') $scheme = 'https';
        $host = $_SERVER['HTTP_HOST'] ?? $_SERVER['SERVER_NAME'] ?? '';
        return $host ? ($scheme . '://' . $host) : '';
    }

    /**
     * LeafSM result_message(英文常量) -> 中文用户可读提示
     */
    protected function translateResultMsg($resultMessage)
    {
        static $map = [
            'SUCCESS'                       => '验证成功',
            'PASS_LIVING_NOT_THE_SAME'      => '活体检测通过,但照片与身份证信息比对非同一人,认证未通过。',
            'NO_ID_CARD_NUMBER'             => '身份证号码不存在,请核对后重试。',
            'ID_NUMBER_NAME_NOT_MATCH'      => '身份证号与姓名不匹配,请核对后重试。',
            'NO_FACE_FOUND'                 => '上传的照片中未检测到人脸,请上传清晰正脸照后重试。',
            'NO_ID_PHOTO'                   => '系统未查询到该身份证对应的参考照片,请稍后重试。',
            'PHOTO_FORMAT_ERROR'            => '参考照片格式错误,请更换清晰 JPG/PNG 照片后重试。',
            'DATA_SOURCE_ERROR'             => '公安数据源临时异常,请稍后重试。',
            'INTERNAL_ERROR'                => '服务器内部错误,请稍后重试。',
            'FAIL_LIVING_FACE_ATTACK'       => '活体检测未通过(存在照片翻拍/攻击特征),请正对摄像头配合动作后重试。',
            'NOT_STARTED'                   => '验证尚未开始,请点击上方按钮进入认证页面。',
            'PROCESSING'                    => '验证进行中,请稍候...',
            'FAILED'                        => '验证流程异常结束,请重新发起实名。',
            'CANCELLED'                     => '您已主动取消本次认证,可重新发起。',
            'TIMEOUT'                       => '等待认证超时,请重新发起实名并尽快完成。',
            'SUPPORT_ERROR'                 => '当前浏览器不支持人脸核身(需支持 webRTC),请改用最新版 Chrome/Edge。',
            'PERMISSIONS_ERROR'             => '摄像头权限被禁止,请允许浏览器访问摄像头后刷新重试。',
            'OTHER_ERROR'                   => '摄像头/WebRTC 连接异常,请检查网络与摄像头后重试。',
        ];
        $k = (string)$resultMessage;
        return isset($map[$k]) ? $map[$k] : '';
    }

    // =========================================================
    // 辅助: 简易 "轮询次数" 跟踪
    // 修复: 改用文件计数(而非 SESSION),避免平台系统轮询与浏览器会话不一致,
    //       导致计数永远达不到上限而无限轮询; 到上限后强制返回终态失败,保证轮询终止。
    // =========================================================
    protected function trackPollAndReturn($certifyId, $status, $msg, $type = 'total', $max = self::MAX_POLL_COUNT)
    {
        $cnt = $this->incPollCounter($certifyId, $type);
        if ($cnt >= $max) {
            // 到达轮询上限仍未出结果 -> 强制终态失败,不再无限轮询
            return [
                'status' => 2,
                'msg'    => '认证状态查询超时,请稍后重新发起实名认证。',
            ];
        }
        return ['status' => $status, 'msg' => $msg];
    }

    protected function incPollCounter($certifyId, $type)
    {
        // 计数文件存于系统临时目录(24 小时有效),跨请求稳定
        $dir  = rtrim(sys_get_temp_dir(), '/\\') . '/starloft_kyc_poll';
        if (!is_dir($dir)) { @mkdir($dir, 0777, true); }
        $file = $dir . '/' . $type . '_' . md5((string)$certifyId) . '.txt';

        $now  = time();
        $cnt  = 0;
        if (is_file($file)) {
            $raw   = @file_get_contents($file);
            $parts = explode('|', (string)$raw);
            if (count($parts) === 2 && (int)$parts[1] >= $now) {
                $cnt = (int)$parts[0];
            }
        }
        $cnt++;
        @file_put_contents($file, $cnt . '|' . ($now + 86400));
        return $cnt;
    }

    // =========================================================
    // 辅助: 通用 HTML 输出
    // =========================================================
    protected function failHtml($msg)
    {
        $m = htmlspecialchars($msg, ENT_QUOTES, 'UTF-8');
        return "<h3 class=\"pt-2 font-weight-bold h2 py-4\" style=\"color: #f56c6c;\"><i class=\"fa fa-exclamation-circle\"></i> {$m}</h3>";
    }
    protected function successHtml()
    {
        return "<h3 class=\"pt-2 font-weight-bold h2 py-4\" style=\"color: #19be6b;\"><i class=\"fa fa-check-circle\"></i> 实名认证已通过,结果同步中...</h3>";
    }
    protected function continuePollHtml($certifyId, $tip)
    {
        $t   = htmlspecialchars($tip, ENT_QUOTES, 'UTF-8');
        $cid = htmlspecialchars($certifyId, ENT_QUOTES, 'UTF-8');
        // Bug 修复:personal() 走"恢复旧任务"分支时不触发前端轮询钩子会卡死。
        // 这里额外注入一段 JS:优先调用魔方财务/顺戴财务已存在的全局轮询函数,
        // 否则自己每 3 秒 AJAX POST /certification/zjmf_mfcw/getStatus
        return <<<HTML
<div class="kyc-auth-container" style="text-align:center;padding:20px;">
    <h5 class="pt-2 font-weight-bold h5 py-4">
        <i class="fa fa-spinner fa-spin"></i> {$t}
    </h5>
    <p class="text-muted small">任务流水号: {$cid}</p>
    <p id="starloft-kyc-progress-tip" class="small"></p>
</div>
<script>
(function(){
    var tipEl  = document.getElementById('starloft-kyc-progress-tip');
    var say    = function(m){ if(tipEl) tipEl.innerText = m; };
    var stop   = false;
    var tick   = 0;
    var doPoll = function kycLoop(){
        if (stop) return;
        // 1. 优先沿用魔方财务已存在的"认证轮询"函数(不同版本函数名不同,都尝试一下)
        if (typeof window.pollCertifyStatus === 'function') { try{ window.pollCertifyStatus(); setTimeout(kycLoop, 3000); return; }catch(e){} }
        if (typeof window.checkCertifyStatus === 'function') { try{ window.checkCertifyStatus(); setTimeout(kycLoop, 3000); return; }catch(e){} }
        // 2. 否则自己 AJAX(兼容 jQuery / 原生)
        var onResp = function(json){
            try{
                var d = (typeof json === 'string') ? JSON.parse(json) : json;
                if (!d || typeof d.status === 'undefined') { say('查询失败,3 秒后重试...'); setTimeout(kycLoop, 3000); return; }
                if (d.status == 1) { stop = true; say('认证通过,正在跳转...'); location.reload(); return; }
                if (d.status == 2) { stop = true; say(d.msg || '认证失败'); location.reload(); return; }
                tick++;
                if (tick > 180) { stop = true; say('长时间无响应,请稍后刷新重试。'); return; }
                if (d.msg) say(d.msg);
                setTimeout(kycLoop, 3000);
            }catch(e){ say('解析异常,3 秒后重试...'); setTimeout(kycLoop, 3000); }
        };
        if (typeof window.jQuery !== 'undefined' && typeof window.jQuery.ajax === 'function') {
            window.jQuery.ajax({
                url:  '/certification/zjmf_mfcw/getStatus',
                type: 'POST',
                dataType: 'json',
                data: { certif_id: '{$cid}' },
                success: onResp,
                error:   function(){ say('网络异常,3 秒后重试...'); setTimeout(kycLoop, 3000); }
            });
        } else {
            var x = new XMLHttpRequest();
            x.open('POST', '/certification/zjmf_mfcw/getStatus', true);
            x.setRequestHeader('Content-Type','application/x-www-form-urlencoded;charset=UTF-8');
            x.onload = function(){ onResp(x.responseText); };
            x.onerror= function(){ say('网络异常,3 秒后重试...'); setTimeout(kycLoop, 3000); };
            x.send('certif_id=' + encodeURIComponent('{$cid}'));
        }
    };
    // DOM ready 后启动
    if (document.readyState === 'complete' || document.readyState === 'interactive') {
        setTimeout(doPoll, 500);
    } else {
        document.addEventListener('DOMContentLoaded', function(){ setTimeout(doPoll, 500); });
    }
})();
</script>
HTML;
    }
}
