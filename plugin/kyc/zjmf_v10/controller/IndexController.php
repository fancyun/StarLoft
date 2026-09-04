<?php
namespace certification\zjmf_v10\controller;

use certification\zjmf_v10\ZjmfV10;

/**
 * StarLoft KYC 实名认证插件 - 外部回调控制器（智简魔方业务系统 v10）
 *
 * 按 v10 实名认证接口规范，插件根目录下创建 controller 目录用于外部访问（异步/同步回调）。
 * 访问地址：
 *   - 异步通知:  {域名}/certification/zjmf_v10/index/notifyHandle
 *   - 认证完成回跳: {域名}/certification/zjmf_v10/index/result
 *   - 状态查询(AJAX): {域名}/certification/zjmf_v10/index/status
 *
 * @author StarLoft
 * @version 1.0.0
 */
class IndexController
{
    /**
     * 异步通知处理（StarLoft KYC 平台认证结果通知）
     *
     * KYC 平台在认证终态（成功/失败/取消/超时）时向 notify_url 推送：
     *   POST application/json
     *   {
     *     "biz_no":          "全平台唯一流水号(即插件保存的 certify_id)",
     *     "status":          2,        // 0待认证 1认证中 2成功 3失败 4已取消 5超时
     *     "result_code":     "1000",
     *     "result_message":  "认证成功",
     *     "cost":            1.50,
     *     "sign":            "HMAC-SHA256 签名(防伪造,必校验)"
     *   }
     *
     * @return void 输出 JSON，KYC 平台收到 2xx 视为送达
     */
    public function notifyHandle()
    {
        // 允许跨域/设置 JSON 输出（v10 底层会包裹响应,这里显式 echo 兜底）
        header('Content-Type: application/json; charset=utf-8');

        $body = file_get_contents('php://input');
        $data = json_decode((string)$body, true);

        if (!is_array($data) || empty($data)) {
            // 兼容表单提交（application/x-www-form-urlencoded）
            $data = $_POST;
        }

        // 校验回调签名（HMAC-SHA256），防止伪造认证结果
        $sign = (string)($data['sign'] ?? '');
        $plugin = new ZjmfV10();
        if (!$plugin->verifyNotifySign($data, $sign)) {
            echo json_encode(['code' => 401, 'message' => 'signature verification failed']);
            return;
        }

        $bizNo = trim((string)($data['biz_no'] ?? ''));
        if ($bizNo === '') {
            echo json_encode(['code' => 400, 'message' => '缺少 biz_no']);
            return;
        }

        // 根据 KYC 订单状态映射为 v10 实名状态（1=通过 2=失败 4=认证中）
        $status = (int)($data['status'] ?? 0);
        $map = [
            0 => 4, // 待认证 -> 认证中
            1 => 4, // 认证中 -> 认证中
            2 => 1, // 成功 -> 通过
            3 => 2, // 失败 -> 失败
            4 => 2, // 已取消 -> 失败
            5 => 2, // 超时 -> 失败
        ];
        $localStatus = $map[$status] ?? 4;
        $resultMsg = trim((string)($data['result_message'] ?? $data['result_code'] ?? ''));

        // 定位并更新本地实名记录（按 certify_id = biz_no）
        $updated = $this->updateByCertifyId($bizNo, $localStatus, $resultMsg);

        if (!$updated) {
            echo json_encode(['code' => 404, 'message' => '未找到对应实名记录']);
            return;
        }

        echo json_encode(['code' => 0, 'message' => 'success']);
    }

    /**
     * 认证完成回跳页（用户在上游完成认证后返回）
     *
     * 展示认证结果，并引导用户返回会员中心实名页面。
     */
    public function result()
    {
        header('Content-Type: text/html; charset=utf-8');
        $certifyId = trim((string)($_GET['certify_id'] ?? $_POST['certify_id'] ?? ''));

        // 尝试查询当前认证状态
        $statusHtml = '<p>正在查询认证结果...</p>';
        if ($certifyId !== '') {
            try {
                $plugin = new ZjmfV10();
                $res = $plugin->getStatus(['certify_id' => $certifyId]);
                $s = (int)($res['status'] ?? 0);
                $msg = htmlspecialchars((string)($res['msg'] ?? ''), ENT_QUOTES, 'UTF-8');
                if ($s === 1) {
                    $statusHtml = '<div style="color:#19be6b;font-weight:bold;">认证已通过，请前往会员中心查看。</div>';
                } elseif ($s === 2) {
                    $statusHtml = '<div style="color:#f56c6c;font-weight:bold;">认证未通过：' . $msg . '</div>';
                } else {
                    $statusHtml = '<div style="color:#e6a23c;">认证处理中：' . $msg . '</div>';
                }
            } catch (\Throwable $e) {
                $statusHtml = '<div style="color:#f56c6c;">查询异常：' . htmlspecialchars($e->getMessage(), ENT_QUOTES, 'UTF-8') . '</div>';
            }
        }

        $certifyIdSafe = htmlspecialchars($certifyId, ENT_QUOTES, 'UTF-8');
        echo <<<HTML
<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>实名认证结果</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
  body { font-family: -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif; background:#f5f7fa; display:flex; align-items:center; justify-content:center; min-height:100vh; margin:0; }
  .card { background:#fff; border-radius:8px; padding:40px 32px; width:90%; max-width:420px; text-align:center; box-shadow:0 2px 12px rgba(0,0,0,.06); }
  .tip { color:#909399; font-size:13px; margin-top:8px; }
  a { display:inline-block; margin-top:24px; color:#fff; background:#409eff; padding:10px 28px; border-radius:4px; text-decoration:none; }
</style>
</head>
<body>
<div class="card">
  <h3>实名认证</h3>
  {$statusHtml}
  <div class="tip">任务流水号：{$certifyIdSafe}</div>
  <a href="/" onclick="history.back();return false;">返回会员中心</a>
</div>
</body>
</html>
HTML;
    }

    /**
     * 状态查询（AJAX，供认证页轮询使用）
     *
     * POST/GET 参数: certif_id / certify_id
     * 返回: {"status":0|1|2|4,"msg":"..."}
     */
    public function status()
    {
        header('Content-Type: application/json; charset=utf-8');
        $certifyId = trim((string)($_POST['certif_id'] ?? $_GET['certif_id'] ?? $_POST['certify_id'] ?? $_GET['certify_id'] ?? ''));

        if ($certifyId === '') {
            echo json_encode(['status' => 0, 'msg' => '缺少任务流水号']);
            return;
        }

        try {
            $plugin = new ZjmfV10();
            $res = $plugin->getStatus(['certify_id' => $certifyId]);
            echo json_encode($res);
        } catch (\Throwable $e) {
            echo json_encode(['status' => 4, 'msg' => '查询异常: ' . $e->getMessage()]);
        }
    }

    /**
     * 按 certify_id（全平台唯一流水号 biz_no）定位并更新本机实名记录状态
     *
     * 兼容 v10 常见认证表名；找不到记录时返回 false。
     *
     * @param string $certifyId 平台流水号
     * @param int $status 实名状态 1=通过 2=失败 4=认证中
     * @param string $resultMsg 结果说明
     * @return bool
     */
    protected function updateByCertifyId($certifyId, $status, $resultMsg = '')
    {
        try {
            if (!class_exists('think\Db')) return false;

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
                    $rec = \think\Db::name($tbl)
                        ->where('certify_id', $certifyId)
                        ->find();
                    if (!$rec) continue;

                    \think\Db::name($tbl)->where('id', $rec['id'])->update([
                        'status'     => $status,
                        'auth_fail'  => $resultMsg,
                        'update_time'=> date('Y-m-d H:i:s'),
                    ]);
                    return true;
                } catch (\Throwable $_) { /* 字段/表不匹配,试下一个 */ }
            }

            // 兜底: SHOW TABLES LIKE 后尝试按 certify_id 更新
            try {
                $like = \think\Db::query("SHOW TABLES LIKE '%certif%'");
                if (!empty($like)) {
                    foreach ($like as $r) {
                        $tblName = array_values($r)[0] ?? '';
                        if ($tblName === '') continue;
                        try {
                            $rec = \think\Db::name($tblName)->where('certify_id', $certifyId)->find();
                            if (!$rec) continue;
                            \think\Db::name($tblName)->where('id', $rec['id'])->update([
                                'status'     => $status,
                                'auth_fail'  => $resultMsg,
                                'update_time'=> date('Y-m-d H:i:s'),
                            ]);
                            return true;
                        } catch (\Throwable $_) {}
                    }
                }
            } catch (\Throwable $_) {}
        } catch (\Throwable $_) {}
        return false;
    }
}
