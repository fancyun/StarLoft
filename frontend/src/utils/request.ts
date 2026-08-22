import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    
    // 根据请求路径判断使用哪个token
    if (config.url?.startsWith('/admin/') && userStore.adminToken) {
      // 管理员接口使用adminToken
      config.headers.Authorization = `Bearer ${userStore.adminToken}`
    } else if (userStore.token) {
      // 普通用户接口使用token
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const { code, message, data } = response.data
    if (code === 0) {
      return data
    } else {
      // 业务错误：显示错误消息并 reject
      ElMessage.error(message || '请求失败')
      return Promise.reject(new Error(message || '请求失败'))
    }
  },
  (error) => {
    // HTTP 错误（网络错误、服务器错误等）
    const backendMsg = error.response?.data?.message
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      // 根据请求路径判断是管理员接口还是用户接口，跳转到对应的登录页
      const isAdminReq = error.config?.url?.startsWith('/admin/')
      if (isAdminReq) {
        userStore.clearAdminInfo()
        window.location.href = '/admin/login'
      } else {
        userStore.logout()
        window.location.href = '/login'
      }
      ElMessage.error(backendMsg || '登录已过期，请重新登录')
    } else {
      ElMessage.error(backendMsg || error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

// 响应拦截器已解包 data 返回业务数据（而非 AxiosResponse），
// 因此对 request 实例做类型收窄，避免 vue-tsc 误判返回类型。
const typedRequest = request as unknown as {
  get<T = any>(url: string, config?: any): Promise<T>
  post<T = any>(url: string, data?: any, config?: any): Promise<T>
  put<T = any>(url: string, data?: any, config?: any): Promise<T>
  delete<T = any>(url: string, config?: any): Promise<T>
}

export default typedRequest
