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

  // 用户注册
  register: (data: { phone: string; sms_code: string; password: string; captcha_ticket: string; captcha_randstr: string }) => {
    return request.post('/user/register', data)
  },

  // 用户登录
  login: (data: { phone: string; password?: string; sms_code?: string; login_type: string; captcha_ticket: string; captcha_randstr: string }) => {
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

  // 更换实名：标记最新实名记录为已更换
  replaceKYC: () => {
    return request.post('/user/kyc/replace')
  },

  // 同步上游认证结果
  syncKYC: () => {
    return request.post('/user/kyc/sync')
  },

  // 查询认证记录
  getKYCRecords: (params: { page?: number; page_size?: number; status?: string }) => {
    return request.get('/user/records', { params })
  },

  // 查询认证记录（别名，兼容旧代码）
  getRecords: (params: { page: number; page_size: number; start_date?: string; end_date?: string }) => {
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

  // 查询充值结果
  getRechargeResult: (pay_order_no: string) => {
    return request.get('/user/recharge/result', { params: { pay_order_no } })
  },

  // 重置API密钥
  resetAPIKey: () => {
    return request.post('/user/api-key/reset')
  },

  // 修改密码
  changePassword: (data: { sms_code: string; new_password: string; captcha_ticket: string; captcha_randstr: string }) => {
    return request.post('/user/change-password', data)
  }
}

// KYC认证API（API Key调用）
export const kycAPI = {
  // 发起认证请求
  start: (data: any, headers: any) => {
    return request.post('/kyc/start', data, { headers })
  },

  // 查询认证结果
  getResult: (data: any, headers: any) => {
    return request.post('/kyc/result', data, { headers })
  },

  // 查询余额
  queryBalance: (headers: any) => {
    return request.post('/kyc/balance/query', {}, { headers })
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

  // 获取用户详情
  getUserDetail: (id: number) => {
    return request.get(`/admin/users/${id}`)
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
  registerUser: (data: { phone: string; password: string; kyc_price?: number }) => {
    return request.post('/admin/users/register', data)
  },

  // 修改用户KYC单价
  updateUserKycPrice: (id: number, data: { kyc_price: number }) => {
    return request.put(`/admin/users/${id}/discount`, data)
  },

  // 为用户手动充值
  rechargeUser: (id: number, data: { amount: number; remark?: string }) => {
    return request.post(`/admin/users/${id}/recharge`, data)
  },

  // 为用户赠送余额
  giftUser: (id: number, data: { amount: number; reason: string }) => {
    return request.post(`/admin/users/${id}/gift`, data)
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

  // 获取订单详情
  getOrderDetail: (id: number) => {
    return request.get(`/admin/orders/${id}`)
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
  }
}
