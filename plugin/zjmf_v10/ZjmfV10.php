<?php
namespace certification\zjmf_v10;

use certification\zjmf_v10\logic\KycSdk;

/**
 * StarLoft KYC 实名认证插件（智简魔方业务系统 v10）
 *
 * 对接 StarLoft KYC 系统进行实名认证，支持身份证三要素认证（姓名+身份证号+人脸识别）。
 * 按智简魔方业务系统 v10 实名认证接口规范开发：
 *   - 入口文件：目录名大驼峰 + .php（ZjmfV10.php），命名空间 certification\zjmf_v10
 *   - 基础信息：$info
 *   - 必选方法：install() / uninstall()
 *   - 可选方法：ZjmfV10CollectionInfo($type) 前台自定义字段
 *              ZjmfV10Person($certifi)    个人实名认证（返回 html）
 *              ZjmfV10Company($certifi)   企业实名认证（返回 html）
 *   - 状态查询：getStatus($certifi)       前台轮询认证状态
 *   - 异步回调：controller/IndexController.php 的 notifyHandle 方法
 *
 * @author StarLoft
 * @version 1.0.0
 */
class ZjmfV10
{
    /**
     * 插件基本信息
     */
    public $info = [
        'name'        => 'zjmf_v10',
        'title'       => 'StarLoft KYC实名认证',
        'description' => 'StarLoft KYC 三要素实名认证（姓名+身份证+人脸识别）',
        'status'      => 1,
        'author'      => 'StarLoft',
        'version'     => '1.0.0',
        'help_url'    => 'https://docs.starloft.cn/kyc/plugin/v10',
    ];

    /** 轮询最大次数(系统按 2s/次轮询, 500 次约 16 分钟, 覆盖上游任务 15 分钟有效期) */
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
     * 前台自定义字段（个人认证）
     *
     * @param string $type 实名认证类型 person|company
     * @return array
     */
    public function ZjmfV10CollectionInfo($type)
    {
        return [
            'name' => [
                'title'    => '姓名',
                'type'     => 'text',
                'value'    => '',
                'tip'      => '',
                'required' => true,
            ],
            'card' => [
                'title'    => '身份证号码',
                'type'     => 'text',
                'value'    => '',
                'tip'      => '',
                'required' => true,
            ],
        ];
    }

    /**
     * 个人实名认证入口
     *
     * @param array $certifi 认证信息
     *   - card_type: 证件类型 1身份证（仅支持身份证）
     *   - name: 姓名
     *   - card: 身份证号
     * @return string html
     */
    public function ZjmfV10Person($certifi)
    {
        try {
            // 证件类型校验：仅支持身份证（card_type=1）；部分环境不传 card_type，按身份证处理
            $cardType = (int)($certifi['card_type'] ?? 1);
            if ($cardType !== 1) {
                return $this->failHtml('当前版本仅支持中国大陆居民身份证实名认证');
            }

            $config = $this->getPluginConfig();
            $sdk    = new KycSdk($config);

            $name   = trim((string)($certifi['name'] ?? ''));
            $idCard = trim((string)($certifi['card'] ?? ''));
            if ($name === '' || $idCard === '') {
                return $this->failHtml('姓名和身份证号不能为空,请返回重新填写');
            }

            // ============ 年龄限制校验(可配置 min_age,默认 16 周岁) ============
            $minAge = intval($config['min_age'] ?? 16);
            if ($minAge > 0) {
                $age = $this->getAgeFromIdCard($idCard);
                if ($age !== null && $age < $minAge) {
                    $msg = "根据身份证信息，您未满 {$minAge} 周岁，暂不支持实名认证";
                    $this->updateLocalCertiStatus($certifi, [
                        'status'     => 2,
                        'auth_fail'  => $msg,
                        'certify_id' => '',
                        'notes'      => $msg,
                    ]);
                    return $this->failHtml($msg);
                }
            }

            // ============ 幂等检查: 已认证中且有任务号 -> 复用,避免重复发单扣费 ============
            $existing = $this->getCurrentUserCertiRecord($certifi);
            if (!empty($existing)) {
                $exStatus    = (int)($existing['certify_status'] ?? $existing['status'] ?? 0);
                $exCertifyId = $this->resolveCertifyId($existing);

                if ($exStatus === 4 && $exCertifyId !== '') {
                    $needRecreate = false;
                    $query = $sdk->queryResult(['platform_biz_no' => $exCertifyId]);
                    $cat   = KycSdk::classifyError($query);

                    if ($cat === KycSdk::ERR_CAT_SUCCESS) {
                        $st = (int)($query['data']['status'] ?? 0);
                        if ($st === 2) {
                            // 已经成功过了,直接写终态 + 显示成功
                            $this->updateLocalCertiStatus($certifi, [
                                'status'     => 1,
                                'auth_fail'  => '',
                                'certify_id' => $exCertifyId,
                            ]);
                            return $this->successHtml();
                        }
                        if ($st === 3) {
                            $msg = $query['data']['result_message'] ?? '实名认证失败';
                            $this->updateLocalCertiStatus($certifi, [
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
                        return $this->continuePollHtml($exCertifyId, '您有一份正在进行的实名任务,已为您恢复进度,正在等待核验结果...');
                    }
                }
            }

            // ============ 真正创建新任务 ============
            $uid   = $this->resolveCurrentUid($certifi);
            // 随机业务单号：每次认证唯一（同一用户允许多次实名），纯随机数无前缀
            $bizNo = '';
            for ($i = 0; $i < 20; $i++) {
                $bizNo .= random_int(0, 9);
            }
            $domain = $this->resolveDomain();
            $notifyUrl = $domain . '/certification/zjmf_v10/index/notifyHandle';
            $returnUrl = !empty($config['return_url'])
                ? $config['return_url']
                : $domain . '/certification/zjmf_v10/index/result';

            $result = $sdk->startKyc([
                'biz_no'         => $bizNo,
                'name'           => $name,
                'id_card'        => $idCard,
                'return_url'     => $returnUrl,
                'notify_url'     => $notifyUrl,
                'biz_extra_data' => json_encode(['uid' => $uid]),
            ]);

            $cat = KycSdk::classifyError($result);

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
                $this->updateLocalCertiStatus($certifi, [
                    'status'    => $failHard ? 2 : 4,
                    'auth_fail' => $displayMsg,
                    'certify_id'=> '',
                    'notes'     => "startKyc错误[{$cat}]: {$msg}\n时间: " . date('Y-m-d H:i:s'),
                ]);
                return $this->failHtml($displayMsg);
            }

            // 创建成功 -> 写库
            $orderData = is_array($result['data'] ?? null) ? $result['data'] : [];
            $platformBizNo = (string)($orderData['platform_biz_no'] ?? '');
            $authUrl = (string)($orderData['auth_url'] ?? '');

            $this->updateLocalCertiStatus($certifi, [
                'status'     => 4,
                'auth_fail'  => '',
                'certify_id' => $platformBizNo,
                'notes'      => "KYC平台流水号: {$platformBizNo}\n业务订单号: {$bizNo}\n创建时间: " . date('Y-m-d H:i:s'),
            ]);

            if ($authUrl === '') {
                return $this->failHtml('实名平台未返回认证地址,请稍后重试或联系管理员。');
            }

            return $this->buildAuthHtml($authUrl);

        } catch (\Exception $e) {
            $errorMsg = '系统错误: ' . $e->getMessage();
            $this->updateLocalCertiStatus($certifi, [
                'status'    => 4,
                'auth_fail' => $errorMsg,
                'certify_id'=> '',
                'notes'     => 'Exception: ' . $e->getMessage() . "\n" . $e->getFile() . ':' . $e->getLine(),
            ]);
            return $this->failHtml($errorMsg);
        }
    }

    /**
     * 企业实名认证（暂不支持）
     *
     * @param array $certifi 认证信息
     * @return string html
     */
    public function ZjmfV10Company($certifi)
    {
        $this->updateLocalCertiStatus($certifi, [
            'status'     => 2,
            'auth_fail'  => '当前版本暂不支持企业认证',
            'certify_id' => '',
            'notes'      => '企业认证功能开发中',
        ]);

        return "<h3 class=\"pt-2 font-weight-bold h2 py-4\" style=\"color: #e6a23c;\">
            <i class=\"fa fa-info-circle\"></i> 当前版本暂不支持企业认证
        </h3>";
    }

    /**
     * 查询认证状态（前台轮询入口）
     *
     * @param array $certifi 认证信息
     * @return array ['status' => 0|1|2|4, 'msg' => ...]
     *   0=待认证 1=认证通过 2=认证失败 4=认证中
     */
    public function getStatus($certifi)
    {
        $certifyId = $this->resolveCertifyId($certifi);

        // 任务号为空则兜底从本地认证记录补齐,避免误报查询失败
        if ($certifyId === '') {
            $rec = $this->getCurrentUserCertiRecord($certifi);
            if (!empty($rec)) {
                $certifyId = $this->resolveCertifyId($rec);
            }
        }

        if ($certifyId === '') {
            return ['status' => 0, 'msg' => '尚未发起实名认证'];
        }

        try {
            $config = $this->getPluginConfig();
            $sdk    = new KycSdk($config);
            $result = $sdk->queryResult(['platform_biz_no' => $certifyId]);
            $cat    = KycSdk::classifyError($result);

            // --- 成功响应 -> 按 data 订单自身状态分 ---
            if ($cat === KycSdk::ERR_CAT_SUCCESS) {
                $orderData    = is_array($result['data'] ?? null) ? $result['data'] : [];
                $orderCode    = (int)($orderData['result_code'] ?? $orderData['status_code'] ?? $orderData['code'] ?? 0);
                $orderMessage = (string)($orderData['result_message'] ?? $orderData['message'] ?? '');

                // 1000 SUCCESS = 验证通过 -> status=1 已认证
                if ($orderCode === 1000 || $orderMessage === 'SUCCESS') {
                    $this->updateLocalCertiStatus($certifi, [
                        'status'     => 1,
                        'auth_fail'  => '',
                        'certify_id' => $certifyId,
                    ]);
                    return ['status' => 1, 'msg' => '实名认证通过'];
                }

                // 终态不通过 -> status=2
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
                    $this->updateLocalCertiStatus($certifi, [
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

                // 6100 = webRTC/权限问题
                if ($orderCode === 6100) {
                    $tipMap = [
                        'SUPPORT_ERROR'     => '当前浏览器不支持人脸核身(需要支持 webRTC),请改用 Chrome/Edge 最新版后重试。',
                        'PERMISSIONS_ERROR' => '摄像头权限被拒绝,请允许浏览器使用摄像头后刷新页面重新进入。',
                        'OTHER_ERROR'       => 'webRTC 连接异常,请检查网络/摄像头后刷新重试。',
                    ];
                    $tip = $tipMap[$orderMessage] ?? '认证遇到问题,请检查浏览器摄像头权限后重试。';
                    return $this->trackPollAndReturn($certifyId, 4, $tip);
                }

                // 3000 DATA_SOURCE_ERROR/INTERNAL_ERROR = 临时错误,继续等
                if ($orderCode === 3000 && in_array($orderMessage, ['DATA_SOURCE_ERROR','INTERNAL_ERROR'], true)) {
                    return $this->trackPollAndReturn($certifyId, 4, '服务端临时异常,持续重试中...(' . $orderMessage . ')', 'err', self::MAX_ERROR_COUNT);
                }

                // 兜底 -> 继续轮询
                return $this->trackPollAndReturn($certifyId, 4, '认证处理中,请稍候...(order_code=' . $orderCode . ')');
            }

            // --- 非成功响应,按错误分类分 ---
            $msg = (string)($result['message'] ?? $result['result_message'] ?? '查询失败');

            if ($cat === KycSdk::ERR_CAT_REJECT) {
                $failMsg = $this->translateResultMsg((string)($result['result_message'] ?? '')) ?: $msg;
                $this->updateLocalCertiStatus($certifi, [
                    'status'     => 2,
                    'auth_fail'  => '实名认证未通过: ' . $failMsg,
                    'certify_id' => $certifyId,
                ]);
                return ['status' => 2, 'msg' => '实名认证未通过: ' . $failMsg];
            }

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

            if ($cat === KycSdk::ERR_CAT_NO_BALANCE) {
                $this->updateLocalCertiStatus($certifi, [
                    'status'     => 2,
                    'auth_fail'  => '实名平台商户余额/额度不足,请联系管理员充值后重试。(' . $msg . ')',
                    'certify_id' => $certifyId,
                ]);
                return ['status' => 2, 'msg' => '实名平台商户余额/额度不足,请联系管理员充值后重试。'];
            }
            if ($cat === KycSdk::ERR_CAT_AUTH) {
                $this->updateLocalCertiStatus($certifi, [
                    'status'     => 2,
                    'auth_fail'  => '实名平台鉴权失败,请联系管理员检查配置。(' . $msg . ')',
                    'certify_id' => $certifyId,
                ]);
                return ['status' => 2, 'msg' => '实名平台鉴权失败,请联系管理员检查配置。'];
            }
            if ($cat === KycSdk::ERR_CAT_PARAM) {
                $this->updateLocalCertiStatus($certifi, [
                    'status'     => 2,
                    'auth_fail'  => '实名参数错误: ' . $msg,
                    'certify_id' => $certifyId,
                ]);
                return ['status' => 2, 'msg' => '实名参数错误: ' . $msg];
            }

            // NOT_FOUND: 连续计数,超阈值判终态
            if ($cat === KycSdk::ERR_CAT_NOT_FOUND) {
                $cnt = $this->incPollCounter($certifyId, 'notfound');
                if ($cnt >= self::MAX_NOT_FOUND) {
                    $this->updateLocalCertiStatus($certifi, [
                        'status'     => 2,
                        'auth_fail'  => '查询不到实名任务(已超过重试次数),请稍后重新发起实名。(' . $msg . ')',
                        'certify_id' => $certifyId,
                    ]);
                    return ['status' => 2, 'msg' => '查询不到实名任务,请稍后重新发起实名。'];
                }
                return ['status' => 4, 'msg' => '任务查询中,请稍候...(未找到,剩余重试 ' . (self::MAX_NOT_FOUND - $cnt) . ')'];
            }

            // 临时错误 / 未知: 保持 4 继续轮询,用较短上限强制终止
            return $this->trackPollAndReturn($certifyId, 4, '查询中,请稍候...(' . $msg . ')', 'err', self::MAX_ERROR_COUNT);

        } catch (\Exception $e) {
            return $this->trackPollAndReturn($certifyId, 4, '查询中,请稍候...(异常:' . $e->getMessage() . ')', 'err', self::MAX_ERROR_COUNT);
        }
    }

    // =========================================================
    // 辅助: 插件配置读取
    // 优先使用框架基类提供的 getConfig()；不存在则回退 config.php 默认值。
    // 若 v10 已将配置注入到 $this->config，这里也会直接返回。
    // =========================================================
    protected function getPluginConfig()
    {
        // 1) 框架基类提供 getConfig() 时优先使用
        if (method_exists($this, 'getConfig')) {
            try {
                $c = $this->getConfig();
                if (is_array($c) && !empty($c)) {
                    return $c;
                }
            } catch (\Throwable $e) {}
        }

        // 2) 读取 config.php 默认值兜底
        $defaults = [];
        $cfgFile = __DIR__ . '/config.php';
        if (is_file($cfgFile)) {
            $arr = include $cfgFile;
            if (is_array($arr)) {
                foreach ($arr as $k => $v) {
                    if (is_array($v)) {
                        $defaults[$k] = $v['value'] ?? ($v['default'] ?? '');
                    } elseif ($v !== null) {
                        $defaults[$k] = $v;
                    }
                }
            }
        }
        return $defaults;
    }

    /**
     * 校验 KYC 平台异步回调签名（HMAC-SHA256）
     *
     * 与后端 buildNotifySign 算法一致：
     * 对固定字段(biz_no/cost/platform_biz_no/result_code/result_message/status)按 key 字典序
     * 拼接为 "k=v&k=v..." 的原始字符串（不做 URL 编码），再以插件自己的 api_secret
     * 计算 HMAC-SHA256 十六进制小写签名。可有效防止伪造回调。
     *
     * @param array  $data 回调 JSON 数据（含 sign 字段）
     * @param string $sign 回调携带的签名
     * @return bool
     */
    public function verifyNotifySign($data, $sign)
    {
        if (!is_array($data) || !is_string($sign) || $sign === '') {
            return false;
        }
        $config = $this->getPluginConfig();
        $secret = (string)($config['api_secret'] ?? '');
        if ($secret === '') {
            return false;
        }

        $fields = [
            'biz_no'          => (string)($data['biz_no'] ?? ''),
            'cost'            => sprintf('%.2f', (float)($data['cost'] ?? 0)),
            'platform_biz_no' => (string)($data['platform_biz_no'] ?? ''),
            'result_code'     => (string)($data['result_code'] ?? ''),
            'result_message'  => (string)($data['result_message'] ?? ''),
            'status'          => (string)(int)($data['status'] ?? 0),
        ];
        ksort($fields);

        $canonical = '';
        foreach ($fields as $k => $v) {
            $canonical .= ($canonical === '' ? '' : '&') . $k . '=' . $v;
        }

        $expect = hash_hmac('sha256', $canonical, $secret);
        return hash_equals($expect, strtolower(trim((string)$sign)));
    }

    // =========================================================
    // 辅助: 本机实名记录读写（v10 无统一全局函数,直接用 think\Db 读写认证表）
    // 表名依次覆盖 v10 常见惯例: host_certification_person / host_certification / im_host_user_certification
    // 以及不带前缀的老表; 另用 SHOW TABLES LIKE '%certif%' 做一次模糊匹配兜底。
    // =========================================================
    protected function getCurrentUserCertiRecord($certifi = [])
    {
        try {
            $uid = $this->resolveCurrentUid($certifi);
            if ($uid <= 0) return [];

            $found = $this->findCertiRecordWithTable($uid);
            return $found ? $found['row'] : [];
        } catch (\Throwable $_) {}
        return [];
    }

    /**
     * 定位用户实名记录及其所在表（同一 uid 只取最新一条）
     *
     * @param int $uid 用户ID
     * @return array|null ['table' => 表名, 'row' => 记录] 或 null
     */
    protected function findCertiRecordWithTable($uid)
    {
        if (!class_exists('think\Db')) return null;

        $tables = [
            'host_certification_person',
            'host_certification',
            'certification_person',
            'certification',
            'im_host_user_certification',
            'host_user_certification',
            'im_certification_personal',
            'certification_personal',
        ];
        foreach ($tables as $tbl) {
            try {
                $row = \think\Db::name($tbl)
                    ->where('uid', $uid)
                    ->order('id desc')
                    ->find();
                if ($row) return ['table' => $tbl, 'row' => $row];
            } catch (\Throwable $_) { /* 表不存在,试下一个 */ }
        }

        // 表名全猜不中,用 SHOW TABLES LIKE 兜底搜一次
        try {
            $like = \think\Db::query("SHOW TABLES LIKE '%certif%'");
            if (!empty($like)) {
                foreach ($like as $r) {
                    $tblName = array_values($r)[0] ?? '';
                    if ($tblName === '' || in_array($tblName, $tables, true)) continue;
                    try {
                        $row = \think\Db::name($tblName)
                            ->where('uid', $uid)
                            ->order('id desc')
                            ->find();
                        if ($row) return ['table' => $tblName, 'row' => $row];
                    } catch (\Throwable $_) {}
                }
            }
        } catch (\Throwable $_) {}

        return null;
    }

    /**
     * 写入/更新本机实名记录状态
     *
     * 优先按 uid 更新已有记录；无记录时尝试插入。
     * 表字段兼容: status / certify_status, auth_fail / auth_fail_reason / reason, certify_id, notes
     */
    protected function updateLocalCertiStatus($certifi = [], $data = [])
    {
        try {
            if (empty($data) || !class_exists('think\Db')) return;
            $uid = $this->resolveCurrentUid($certifi);
            if ($uid <= 0) return;

            // 已有记录 -> 定位其所在表后更新（避免跨表错位）
            $found = $this->findCertiRecordWithTable($uid);
            if ($found) {
                $update = [];
                if (isset($data['status'])) $update['status'] = $data['status'];
                if (isset($data['auth_fail'])) $update['auth_fail'] = $data['auth_fail'];
                if (array_key_exists('certify_id', $data)) $update['certify_id'] = $data['certify_id'];
                if (isset($data['notes'])) $update['notes'] = $data['notes'];
                if (isset($data['name'])) $update['name'] = $data['name'];
                if (isset($data['card'])) $update['card'] = $data['card'];
                if (!empty($update)) {
                    try {
                        \think\Db::name($found['table'])->where('id', $found['row']['id'])->update($update);
                    } catch (\Throwable $_) { /* 字段不匹配则忽略 */ }
                }
                return;
            }

            // 无记录 -> 尝试插入到首个可用的认证表
            $tables = [
                'host_certification_person',
                'host_certification',
                'certification_person',
                'certification',
                'im_host_user_certification',
                'host_user_certification',
                'im_certification_personal',
                'certification_personal',
            ];
            $insert = [
                'uid'         => $uid,
                'status'      => $data['status'] ?? 4,
                'auth_fail'   => $data['auth_fail'] ?? '',
                'certify_id'  => $data['certify_id'] ?? '',
                'notes'       => $data['notes'] ?? '',
                'create_time' => date('Y-m-d H:i:s'),
                'update_time' => date('Y-m-d H:i:s'),
            ];
            foreach ($tables as $tbl) {
                try {
                    \think\Db::name($tbl)->insert($insert);
                    return;
                } catch (\Throwable $_) { /* 表不存在/字段不匹配,试下一个 */ }
            }
        } catch (\Throwable $_) {}
    }

    /**
     * 解析认证任务号（平台流水号 platform_biz_no）
     * 兼容字段名: certify_id / certif_id / certifi_id / certifyId / certifiId / certifId
     */
    protected function resolveCertifyId($certifi)
    {
        $candidateKeys = ['certify_id', 'certif_id', 'certifi_id', 'certifyId', 'certifiId', 'certifId'];

        if (is_array($certifi)) {
            foreach ($candidateKeys as $k) {
                $v = trim((string)($certifi[$k] ?? ''));
                if ($v !== '') return $v;
            }
            $nested = $certifi['certifi'] ?? null;
            if (is_array($nested)) {
                foreach ($candidateKeys as $k) {
                    $v = trim((string)($nested[$k] ?? ''));
                    if ($v !== '') return $v;
                }
            }
        }
        return '';
    }

    /**
     * 根据18位身份证号计算周岁年龄
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
     * 通用取 UID: 优先 $certifi 记录里的 uid/user_id, 再尝试 Session/Cookie/session()/request/auth('user')
     */
    protected function resolveCurrentUid($certifi = [])
    {
        $uid = 0;

        if (is_array($certifi)) {
            $uid = (int)($certifi['uid'] ?? $certifi['user_id'] ?? 0);
            if ($uid <= 0 && is_array($certifi['certifi'] ?? null)) {
                $uid = (int)($certifi['certifi']['uid'] ?? $certifi['certifi']['user_id'] ?? 0);
            }
            if ($uid > 0) return $uid;
        }

        try {
            if (class_exists('think\Session')) {
                foreach (['user_id', 'uid', 'userid', 'user.id', 'userinfo.id', 'login_uid'] as $k) {
                    $uid = (int)\think\Session::get($k);
                    if ($uid > 0) return $uid;
                }
            }
            if (isset($_SESSION)) {
                foreach (['user_id', 'uid', 'userid', 'userinfo.id'] as $k) {
                    if (isset($_SESSION[$k])) {
                        $uid = (int)$_SESSION[$k];
                        if ($uid > 0) return $uid;
                    }
                }
            }
        } catch (\Throwable $_) {}

        try {
            if (class_exists('think\Cookie')) {
                foreach (['user_id', 'uid', 'userid'] as $k) {
                    $uid = (int)\think\Cookie::get($k);
                    if ($uid > 0) return $uid;
                }
            }
        } catch (\Throwable $_) {}

        if ($uid <= 0 && function_exists('session')) {
            foreach (['user_id', 'uid', 'userid'] as $k) {
                $v = \session($k);
                if ($v) {
                    $uid = (int)$v;
                    if ($uid > 0) return $uid;
                }
            }
        }

        if ($uid <= 0 && function_exists('request')) {
            try {
                $uid = (int)(\request()->uid ?? \input('uid/d', 0));
            } catch (\Throwable $_) {}
        }

        if ($uid <= 0 && function_exists('auth')) {
            try {
                $user = \auth('user');
                if (is_object($user) && isset($user->id)) $uid = (int)$user->id;
            } catch (\Throwable $_) {}
        }

        return $uid;
    }

    /**
     * 安全的获取当前域名
     */
    protected function resolveDomain()
    {
        try {
            if (function_exists('request') && is_object(\request())) {
                $dom = \request()->domain();
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
     * 上游 result_message(英文常量) -> 中文用户可读提示
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
    // 辅助: 简易 "轮询次数" 跟踪（文件计数,跨请求稳定）
    // =========================================================
    protected function trackPollAndReturn($certifyId, $status, $msg, $type = 'total', $max = self::MAX_POLL_COUNT)
    {
        $cnt = $this->incPollCounter($certifyId, $type);
        if ($cnt >= $max) {
            return [
                'status' => 2,
                'msg'    => '认证状态查询超时,请稍后重新发起实名认证。',
            ];
        }
        return ['status' => $status, 'msg' => $msg];
    }

    protected function incPollCounter($certifyId, $type)
    {
        $dir  = rtrim(sys_get_temp_dir(), '/\\') . '/starloft_kyc_poll_v10';
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
    protected function buildAuthHtml($authUrl)
    {
        $url = htmlspecialchars($authUrl, ENT_QUOTES, 'UTF-8');
        return <<<HTML
<div class="kyc-auth-container" style="text-align: center; padding: 20px;">
    <h5 class="pt-2 font-weight-bold h5 py-4">正在跳转到实名认证页面...</h5>
    <p>如果页面未自动跳转,请点击下方按钮(<b>请勿反复刷新或多次点击</b>,以免重复创建实名任务扣费)</p>
    <a href="{$url}" class="btn btn-primary" target="_blank">前往认证</a>
    <script>
        (function(){
            var opened = false;
            function tryOpen(){ if(!opened){ opened=true; window.open('{$url}', '_blank'); } }
            setTimeout(tryOpen, 800);
            document.addEventListener && document.addEventListener('click', function once(){
                tryOpen();
                document.removeEventListener('click', once);
            }, {once:true});
        })();
    </script>
</div>
HTML;
    }

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
        return <<<HTML
<div class="kyc-auth-container" style="text-align:center;padding:20px;">
    <h5 class="pt-2 font-weight-bold h5 py-4">
        <i class="fa fa-spinner fa-spin"></i> {$t}
    </h5>
    <p class="text-muted small">任务流水号: {$cid}</p>
    <p id="starloft-kyc-v10-progress-tip" class="small"></p>
</div>
<script>
(function(){
    var tipEl  = document.getElementById('starloft-kyc-v10-progress-tip');
    var say    = function(m){ if(tipEl) tipEl.innerText = m; };
    var stop   = false;
    var tick   = 0;
    var doPoll = function kycLoop(){
        if (stop) return;
        if (typeof window.pollCertifyStatus === 'function') { try{ window.pollCertifyStatus(); setTimeout(kycLoop, 3000); return; }catch(e){} }
        if (typeof window.checkCertifyStatus === 'function') { try{ window.checkCertifyStatus(); setTimeout(kycLoop, 3000); return; }catch(e){} }
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
                url:  '/certification/zjmf_v10/index/status',
                type: 'POST',
                dataType: 'json',
                data: { certif_id: '{$cid}' },
                success: onResp,
                error:   function(){ say('网络异常,3 秒后重试...'); setTimeout(kycLoop, 3000); }
            });
        } else {
            var x = new XMLHttpRequest();
            x.open('POST', '/certification/zjmf_v10/index/status', true);
            x.setRequestHeader('Content-Type','application/x-www-form-urlencoded;charset=UTF-8');
            x.onload = function(){ onResp(x.responseText); };
            x.onerror= function(){ say('网络异常,3 秒后重试...'); setTimeout(kycLoop, 3000); };
            x.send('certif_id=' + encodeURIComponent('{$cid}'));
        }
    };
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
