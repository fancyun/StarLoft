import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  // 用户端
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<any>(null)
  const isKYCVerified = ref(false)

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

  return {
    token,
    userInfo,
    isKYCVerified,
    setToken,
    setUserInfo,
    logout,
    clearAuth
  }
})
