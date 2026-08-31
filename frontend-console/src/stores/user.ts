import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  // 用户端
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<any>(null)
  const isKYCVerified = ref(false)

  // 管理员端
  const adminToken = ref(localStorage.getItem('admin_token') || '')
  const adminInfo = ref<any>(null)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUserInfo = (info: any) => {
    userInfo.value = info
    isKYCVerified.value = info.is_kyc_verified === 1
  }

  const logout = () => {
    token.value = ''
    userInfo.value = null
    isKYCVerified.value = false
    localStorage.removeItem('token')
  }

  const clearAuth = () => {
    logout()
  }

  // 管理员相关
  const setAdminToken = (newToken: string) => {
    adminToken.value = newToken
    localStorage.setItem('admin_token', newToken)
  }

  const setAdminInfo = (info: any) => {
    adminInfo.value = info
  }

  const clearAdminInfo = () => {
    adminToken.value = ''
    adminInfo.value = null
    localStorage.removeItem('admin_token')
  }

  return {
    token,
    userInfo,
    isKYCVerified,
    adminToken,
    adminInfo,
    setToken,
    setUserInfo,
    logout,
    clearAuth,
    setAdminToken,
    setAdminInfo,
    clearAdminInfo
  }
})
