import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAdminStore = defineStore('admin', () => {
  // 管理员端
  const adminToken = ref(localStorage.getItem('admin_token') || '')
  const adminInfo = ref<any>(null)

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
    adminToken,
    adminInfo,
    setAdminToken,
    setAdminInfo,
    clearAdminInfo
  }
})
