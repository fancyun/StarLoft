import request from '@/utils/request'

// 类型定义
interface StatsOverview {
  total_users: number
  today_orders: number
  today_revenue: number
  month_revenue: number
}

interface StatsOrdersResponse {
  dates: string[]
  counts: number[]
}

interface StatsRevenueResponse {
  dates: string[]
  amounts: number[]
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

// 管理后台API
export const adminAPI = {
  // 管理员登录
  login: (data: { username: string; password: string }) => {
    return request.post('/admin/login', data)
  },

  // 获取用户列表
  getUsers: (params: any) => {
    return request.get('/admin/users', { params })
  },

  // 更新用户状态
  updateUserStatus: (id: number, data: { status: number }) => {
    return request.put(`/admin/users/${id}/status`, data)
  },

  // 删除用户
  deleteUser: (id: number) => {
    return request.delete(`/admin/users/${id}`)
  },

  // 手动注册用户
  registerUser: (data: { phone: string; username: string; email: string; password: string }) => {
    return request.post('/admin/users/register', data)
  },

  // 为用户手动充值（需银行流水单号）
  rechargeUser: (id: number, data: { amount: number; bank_serial_no: string; remark?: string }) => {
    return request.post(`/admin/users/${id}/recharge`, data)
  },

  // 获取用户财务统计
  getUserFinanceStats: (id: number) => {
    return request.get(`/admin/users/${id}/finance/stats`)
  },

  // 获取用户余额流水
  getUserBalanceLogs: (id: number, params?: { page?: number; page_size?: number }) => {
    return request.get(`/admin/users/${id}/balance-logs`, { params })
  },

  // 获取用户认证订单
  getUserAuthOrders: (id: number, params?: { page?: number; page_size?: number }) => {
    return request.get(`/admin/users/${id}/auth-orders`, { params })
  },

  // 获取认证订单列表
  getOrders: (params: any) => {
    return request.get('/admin/orders', { params })
  },

  // 获取支付订单列表
  getPayments: (params: any) => {
    return request.get('/admin/payments', { params })
  },

  // 获取系统配置
  getConfig: () => {
    return request.get('/admin/config')
  },

  // 更新系统配置
  updateConfig: (data: any) => {
    return request.put('/admin/config', data)
  },

  // 获取统计概览
  getStatsOverview: (): Promise<StatsOverview> => {
    return request.get('/admin/stats/overview')
  },

  // 获取订单统计
  getStatsOrders: (params: any): Promise<StatsOrdersResponse> => {
    return request.get('/admin/stats/orders', { params })
  },

  // 获取收入统计
  getStatsRevenue: (params: any): Promise<StatsRevenueResponse> => {
    return request.get('/admin/stats/revenue', { params })
  },

  // 获取最近订单
  getRecentOrders: (limit: number) => {
    return request.get(`/admin/orders/recent?limit=${limit}`)
  },

  // 修改管理员密码
  changePassword: (data: { old_password: string; new_password: string }) => {
    return request.post('/admin/change-password', data)
  },

  // 获取资源包列表
  getPacks: () => {
    return request.get('/admin/packs')
  },

  // 创建资源包
  createPack: (data: { name: string; total_count: number; price: number; stock?: number; status?: number; description?: string }) => {
    return request.post('/admin/packs', data)
  },

  // 更新资源包
  updatePack: (id: number, data: any) => {
    return request.put(`/admin/packs/${id}`, data)
  },

  // 删除（下架）资源包
  deletePack: (id: number) => {
    return request.delete(`/admin/packs/${id}`)
  }
}
