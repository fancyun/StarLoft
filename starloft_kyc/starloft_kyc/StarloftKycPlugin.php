<?php
namespace certification\starloft_kyc;

use app\admin\lib\Plugin;
use certification\starloft_kyc\logic\KycSdk;

/**
 * StarLoft KYC 实名认证插件
 * 
 * 对接 StarLoft KYC 系统进行实名认证
 * 支持身份证三要素认证（姓名+身份证号+人脸识别）
 * 
 * @author StarLoft
 * @version 1.0.0
 */
class StarloftKycPlugin extends Plugin
{
    /**
     * 插件基本信息
     */
    public $info = [
        'name'        => 'StarloftKyc',
        'title'       => 'StarLoft KYC实名认证',
        'description' => 'StarLoft KYC 三要素实名认证（姓名+身份证+人脸识别）',
        'status'      => 1,
        'author'      => 'StarLoft',
        'version'     => '1.0.0',
        'help_url'    => 'https://docs.starloft.cn/kyc/plugin'
    ];

    /**
     * 插件安装
     */
    public function install()
    {
        return true;
    }

    /**
     * 插件卸载
     */
    public function uninstall()
    {
        return true;
    }

    /**
     * 个人实名认证
     * 
     * @param array $certifi 认证信息
     *   - name: 姓名
     *   - card: 身份证号
     * @return string HTML内容
     */
    public function personal($certifi)
    {
        try {
            // 防重复发起：刷新页面或重复提交时，若上次认证仍在进行中，直接返回提示，不重复创建订单
            $uid = \request()->uid;
            $pendingKey = 'starloft_kyc_pending_' . $uid;
            $pending = session($pendingKey);
            if (is_array($pending) && !empty($pending['time']) && (time() - intval($pending['time'])) < 900) {
                return "<h3 class=\"pt-2 font-weight-bold h2 py-4\" style=\"color: #e6a23c;\">
                    <i class=\"fa fa-info-circle\"></i> 您有认证正在进行中，请勿重复提交，请点击页面上的“前往认证”按钮继续完成认证
                </h3>";
            }

            // 获取插件配置
            $config = $this->getConfig();
            
            // 初始化SDK
            $sdk = new KycSdk($config);
            
            // 生成业务订单号
            $bizNo = 'ZJMF' . date('YmdHis') . mt_rand(1000, 9999);
            
            // 构建回调URL（$uid 已在防重复检查处获取）
            $notifyUrl = request()->domain() . '/certification/starloft_kyc/callback?uid=' . $uid;
            $returnUrl = request()->domain() . '/certification/starloft_kyc/result?uid=' . $uid;
            
            // 调用KYC系统创建认证订单
            $result = $sdk->startKyc([
                'biz_no' => $bizNo,
                'name' => $certifi['name'],
                'id_card' => $certifi['card'],
                'return_url' => $returnUrl,
                'notify_url' => $notifyUrl,
                'biz_extra_data' => json_encode(['uid' => $uid])
            ]);
            
            if ($result['code'] === 0) {
                $orderData = $result['data'];

                // 记录会话标记，防止刷新页面重复发起（与订单有效期15分钟一致）
                session($pendingKey, ['biz_no' => $bizNo, 'time' => time()]);

                // 更新认证状态
                $data = [
                    'status' => 4, // 4: 认证中
                    'auth_fail' => '',
                    'certify_id' => $orderData['platform_biz_no'],
                    'notes' => "KYC平台流水号: {$orderData['platform_biz_no']}\n" .
                               "业务订单号: {$bizNo}\n" .
                               "创建时间: " . date('Y-m-d H:i:s')
                ];
                
                updatePersonalCertifiStatus($data);
                
                // 返回认证页面
                $authUrl = $orderData['auth_url'];
                return <<<HTML
<div class="kyc-auth-container" style="text-align: center; padding: 20px;">
    <h5 class="pt-2 font-weight-bold h5 py-4">正在跳转到实名认证页面...</h5>
    <p>如果页面未自动跳转，请点击下方按钮</p>
    <a href="{$authUrl}" class="btn btn-primary" target="_blank">前往认证</a>
    <script>
        setTimeout(function() {
            window.open('{$authUrl}', '_blank');
        }, 1000);
    </script>
</div>
HTML;
            } else {
                // 认证失败
                $errorMsg = $result['message'] ?? '实名认证接口配置错误，请联系管理员';
                $data = [
                    'status' => 2, // 2: 未通过
                    'auth_fail' => $errorMsg,
                    'certify_id' => '',
                    'notes' => "错误信息: {$errorMsg}\n失败时间: " . date('Y-m-d H:i:s')
                ];
                updatePersonalCertifiStatus($data);
                
                return "<h3 class=\"pt-2 font-weight-bold h2 py-4\" style=\"color: #f56c6c;\">
                    <i class=\"fa fa-exclamation-circle\"></i> {$errorMsg}
                </h3>";
            }
        } catch (\Exception $e) {
            $errorMsg = '系统错误: ' . $e->getMessage();
            $data = [
                'status' => 2,
                'auth_fail' => $errorMsg,
                'certify_id' => '',
                'notes' => $errorMsg
            ];
            updatePersonalCertifiStatus($data);
            
            return "<h3 class=\"pt-2 font-weight-bold h2 py-4\" style=\"color: #f56c6c;\">
                <i class=\"fa fa-exclamation-circle\"></i> {$errorMsg}
            </h3>";
        }
    }

    /**
     * 企业实名认证（暂不支持）
     */
    public function company($certifi)
    {
        $data = [
            'status' => 2,
            'auth_fail' => '当前版本暂不支持企业认证',
            'certify_id' => '',
            'notes' => '企业认证功能开发中'
        ];
        updateCompanyCertifiStatus($data);
        
        return "<h3 class=\"pt-2 font-weight-bold h2 py-4\" style=\"color: #e6a23c;\">
            <i class=\"fa fa-info-circle\"></i> 当前版本暂不支持企业认证
        </h3>";
    }

    /**
     * 前台自定义字段（不需要额外字段）
     */
    public function collectionInfo()
    {
        return [];
    }

    /**
     * 查询认证状态（用于轮询）
     * 
     * @param array $certifi
     *   - certify_id: 认证证书（订单号）
     * @return array|bool
     */
    public function getStatus($certifi)
    {
        try {
            $config = $this->getConfig();
            $sdk = new KycSdk($config);
            
            // 查询认证结果
            $result = $sdk->queryResult([
                'platform_biz_no' => $certifi['certify_id']
            ]);
            
            if ($result['code'] === 0) {
                $orderData = $result['data'];
                
                switch ($orderData['status']) {
                    case 2: // 认证成功
                        $this->clearPendingMark();
                        return [
                            'status' => 1,
                            'msg' => '实名认证通过'
                        ];
                    case 3: // 认证失败
                        $this->clearPendingMark();
                        return [
                            'status' => 2,
                            'msg' => $orderData['result_message'] ?? '实名认证失败'
                        ];
                    case 1: // 认证中
                    default:
                        return [
                            'status' => 4,
                            'msg' => '认证处理中，请稍候...'
                        ];
                }
            }
            
            // 查询失败，保持认证中状态
            return [
                'status' => 4,
                'msg' => '查询中...'
            ];
        } catch (\Exception $e) {
            // 查询异常，保持认证中状态
            return [
                'status' => 4,
                'msg' => '查询中...'
            ];
        }
    }

    /**
     * 清除当前用户“认证进行中”的会话标记
     * 
     * 认证有最终结果（成功/失败）后调用，允许用户重新发起认证
     */
    private function clearPendingMark()
    {
        $uid = \request()->uid;
        if ($uid) {
            session('starloft_kyc_pending_' . $uid, null);
        }
    }
}
