import request from '@/utils/request'

// 类型定义
interface StatsOrdersResponse {
  dates: string[]
  counts: number[]
}

// 公开API（无需登录）
export const publicAPI = {
  // 获取系统配置
  getConfig: () => {
    return request.get('/public/config')
  }
}

// 用户相关API
export const userAPI = {
  // 发送短信验证码
  sendSMSCode: (data: { phone: string; captcha_ticket: string; captcha_randstr: string; scene: string }) => {
    return request.post('/user/send-code', data)
  },

  // 发送邮箱验证码
  sendEmailCode: (data: { email: string; captcha_ticket: string; captcha_randstr: string; scene: string }) => {
    return request.post('/user/send-email-code', data)
  },

  // 用户注册（手机号+用户名+邮箱+双验证码）
  register: (data: { phone: string; username: string; email: string; sms_code: string; email_code: string; password: string; captcha_ticket: string; captcha_randstr: string }) => {
    return request.post('/user/register', data)
  },

  // 用户登录（支持用户名/手机号/邮箱）
  login: (data: { account?: string; password?: string; sms_code?: string; login_type: string; captcha_ticket: string; captcha_randstr: string }) => {
    return request.post('/user/login', data)
  },

  // 获取用户信息
  getProfile: () => {
    return request.get('/user/profile')
  },

  // 查询KYC认证状态
  getKYCStatus: () => {
    return request.get('/user/kyc/status')
  },

  // Web端发起KYC认证
  startKYC: (data: { name: string; id_card: string; return_url: string }) => {
    return request.post('/user/kyc', data)
  },

  // 取消当前进行中的认证
  cancelKYC: () => {
    return request.delete('/user/kyc')
  },

  // 同步上游认证结果
  syncKYC: () => {
    return request.post('/user/kyc/sync')
  },

  // 查询认证记录
  getKYCRecords: (params: { page?: number; page_size?: number; status?: string }) => {
    return request.get('/user/records', { params })
  },

  // 查询认证调用统计（近30天，按天计数）
  getCallStats: (): Promise<StatsOrdersResponse> => {
    return request.get('/user/stats/calls')
  },

  // 发起充值
  createRecharge: (data: { amount: number; channel: string }) => {
    return request.post('/user/recharge', data)
  },

  // 查询充值结果（轮询）
  getRechargeResult: (params: { pay_order_no: string }) => {
    return request.get('/user/recharge/result', { params })
  },

  // 重置API密钥
  resetAPIKey: () => {
    return request.post('/user/api-key/reset')
  },

  // 修改密码
  changePassword: (data: { sms_code: string; new_password: string; captcha_ticket: string; captcha_randstr: string }) => {
    return request.post('/user/change-password', data)
  },

  // 获取在售资源包列表
  listPacks: () => {
    return request.get('/user/packs')
  },

  // 使用余额购买资源包
  purchasePack: (id: number) => {
    return request.post(`/user/packs/${id}/purchase`)
  },

  // 我的资源包列表
  myPacks: () => {
    return request.get('/user/packs/mine')
  }
}
