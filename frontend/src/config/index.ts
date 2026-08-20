/**
 * 前端配置文件（公开配置,不含敏感信息）
 */

export const APP_CONFIG = {
  // API 基础地址（开发/生产环境）
  API_BASE_URL: (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8080',
  
  // 腾讯云验证码配置
  captchaAppId: (import.meta as any).env?.VITE_CAPTCHA_APP_ID || '',
  
  // 其他前端配置...
}

/**
 * 注意：
 * - 所有密钥和敏感配置（包括 CaptchaAppId）都应该通过 API 从后端获取
 * - 不要在前端硬编码任何配置参数
 */
