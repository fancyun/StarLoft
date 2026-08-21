<?php
namespace certification\starloft_kyc\logic;

/**
 * StarLoft KYC SDK
 * 
 * 用于对接 StarLoft KYC 实名认证系统的 SDK 类
 * 
 * @author StarLoft
 * @version 1.0.0
 */
class KycSdk
{
    /** @var string API基础URL */
    private $apiUrl;
    
    /** @var string API Key */
    private $apiKey;
    
    /** @var string API Secret */
    private $apiSecret;
    
    /** @var int 请求超时时间（秒） */
    private $timeout = 30;

    /**
     * 构造函数
     * 
     * @param array $config 配置信息
     *   - api_url: API地址
     *   - api_key: API Key
     *   - api_secret: API Secret
     */
    public function __construct($config)
    {
        $this->apiUrl = rtrim($config['api_url'] ?? '', '/');
        $this->apiKey = $config['api_key'] ?? '';
        $this->apiSecret = $config['api_secret'] ?? '';
        
        if (empty($this->apiUrl) || empty($this->apiKey) || empty($this->apiSecret)) {
            throw new \Exception('KYC API配置不完整，请检查插件配置');
        }
    }

    /**
     * 生成 HMAC-SHA256 签名
     * 
     * 签名算法：hex(HMAC-SHA256(api_secret, 原始请求体))
     * 
     * @param string $body 原始请求体（POST 为 JSON 字符串，GET 为空字符串）
     * @return string 小写十六进制签名
     */
    private function generateSign($body)
    {
        return hash_hmac('sha256', $body, $this->apiSecret);
    }

    /**
     * 发送HTTP请求
     * 
     * @param string $method 请求方法 GET|POST
     * @param string $endpoint API端点
     * @param array $data 请求数据
     * @return array 响应数据
     */
    private function request($method, $endpoint, $data = [])
    {
        $url = $this->apiUrl . $endpoint;

        // 序列化请求体（签名与实际发送必须完全一致）
        // POST 无参数时发送空 JSON 对象 {}，与文档保持一致（空数组 [] 易被误判为非法 JSON）
        if ($method === 'POST') {
            $body = empty($data) ? '{}' : json_encode($data);
        } else {
            $body = '';
        }

        // 时间戳：Unix 秒，后端校验允许 ±5 分钟
        $timestamp = (string)time();
        // 签名：hex(HMAC-SHA256(api_secret, 原始请求体))
        $sign = $this->generateSign($body);

        $headers = [
            'Content-Type: application/json',
            'X-Api-Key: ' . $this->apiKey,
            'X-Sign: ' . $sign,
            'X-Sign-Version: hmac_sha256',
            'X-Timestamp: ' . $timestamp,
        ];
        
        $ch = curl_init();
        
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, $this->timeout);
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, true);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, 2);
        
        if ($method === 'POST') {
            curl_setopt($ch, CURLOPT_POST, true);
            curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
        } elseif ($method === 'GET' && !empty($data)) {
            $url .= '?' . http_build_query($data);
            curl_setopt($ch, CURLOPT_URL, $url);
        }
        
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        
        curl_close($ch);
        
        if ($error) {
            return [
                'code' => -1,
                'message' => '网络请求失败: ' . $error
            ];
        }
        
        $result = json_decode($response, true);

        // 响应体不是合法 JSON：通常是 404/502 等网关或路由返回的 HTML，
        // 附带 HTTP 状态码与响应片段，便于定位（常见原因：API地址缺少 /api/v1 前缀）
        if (json_last_error() !== JSON_ERROR_NONE) {
            $snippet = substr(trim((string)$response), 0, 200);
            $message = 'API响应解析失败 (HTTP ' . $httpCode . ')';
            if ($snippet !== '') {
                $message .= ': ' . $snippet;
            }
            if ($httpCode == 404) {
                $message .= ' [提示: 请确认 API地址 包含 /api/v1 前缀]';
            }
            return [
                'code' => -1,
                'message' => $message
            ];
        }

        if ($httpCode !== 200) {
            return [
                'code' => $httpCode,
                'message' => $result['message'] ?? ('HTTP错误: ' . $httpCode)
            ];
        }

        return $result;
    }

    /**
     * 创建实名认证订单
     * 
     * @param array $params 参数
     *   - biz_no: 业务订单号
     *   - name: 真实姓名
     *   - id_card: 身份证号
     *   - return_url: 前端回调地址
     *   - notify_url: 后端回调地址
     *   - biz_extra_data: 业务扩展数据（可选）
     * @return array
     */
    public function startKyc($params)
    {
        return $this->request('POST', '/kyc/start', $params);
    }

    /**
     * 查询认证结果
     * 
     * @param array $params 参数
     *   - biz_no: 业务订单号（可选）
     *   - platform_biz_no: 平台订单号（可选）
     * @return array
     */
    public function queryResult($params)
    {
        return $this->request('POST', '/kyc/result', $params);
    }

    /**
     * 查询用户余额
     * 
     * @return array
     */
    public function queryBalance()
    {
        return $this->request('POST', '/kyc/balance/query');
    }

    /**
     * 测试API连接
     * 
     * @return array
     */
    public function testConnection()
    {
        return $this->queryBalance();
    }
}
