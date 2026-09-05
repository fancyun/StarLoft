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

// 管理后台API
export const adminAPI = {
  // 管理员登录
  login: (data: { username: string; password: string }) => {
    return request.post('/login', data)
  },

  // 获取用户列表
  getUsers: (params: any) => {
    return request.get('/users', { params })
  },

  // 更新用户状态
  updateUserStatus: (id: number, data: { status: number }) => {
    return request.put(`/users/${id}/status`, data)
  },

  // 删除用户
  deleteUser: (id: number) => {
    return request.delete(`/users/${id}`)
  },

  // 手动注册用户
  registerUser: (data: { phone: string; username: string; email: string; password: string }) => {
    return request.post('/users/register', data)
  },

  // 为用户手动充值（需银行流水单号）
  rechargeUser: (id: number, data: { amount: number; bank_serial_no: string; remark?: string }) => {
    return request.post(`/users/${id}/recharge`, data)
  },

  // 获取用户财务统计
  getUserFinanceStats: (id: number) => {
    return request.get(`/users/${id}/finance/stats`)
  },

  // 获取用户余额流水
  getUserBalanceLogs: (id: number, params?: { page?: number; page_size?: number }) => {
    return request.get(`/users/${id}/balance-logs`, { params })
  },

  // 获取用户认证订单
  getUserAuthOrders: (id: number, params?: { page?: number; page_size?: number }) => {
    return request.get(`/users/${id}/auth-orders`, { params })
  },

  // 获取认证订单列表
  getOrders: (params: any) => {
    return request.get('/orders', { params })
  },

  // 获取支付订单列表
  getPayments: (params: any) => {
    return request.get('/payments', { params })
  },

  // 获取统计概览
  getStatsOverview: (): Promise<StatsOverview> => {
    return request.get('/stats/overview')
  },

  // 获取订单统计
  getStatsOrders: (params: any): Promise<StatsOrdersResponse> => {
    return request.get('/stats/orders', { params })
  },

  // 获取收入统计
  getStatsRevenue: (params: any): Promise<StatsRevenueResponse> => {
    return request.get('/stats/revenue', { params })
  },

  // 获取最近订单
  getRecentOrders: (limit: number) => {
    return request.get(`/orders/recent?limit=${limit}`)
  },

  // 修改管理员密码
  changePassword: (data: { old_password: string; new_password: string }) => {
    return request.post('/change-password', data)
  },

  // 获取资源包列表
  getPacks: () => {
    return request.get('/packs')
  },

  // 创建资源包
  createPack: (data: { name: string; total_count: number; price: number; stock?: number; status?: number; description?: string }) => {
    return request.post('/packs', data)
  },

  // 更新资源包
  updatePack: (id: number, data: any) => {
    return request.put(`/packs/${id}`, data)
  },

  // 删除（下架）资源包
  deletePack: (id: number) => {
    return request.delete(`/packs/${id}`)
  }
}
