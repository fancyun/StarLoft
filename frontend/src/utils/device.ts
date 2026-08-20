/**
 * 设备检测工具
 * 基于 User-Agent 判断当前设备类型
 */

export function isMobileDevice(): boolean {
  const ua = navigator.userAgent
  return /Android|webOS|iPhone|iPod|BlackBerry|IEMobile|Opera Mini|Mobile|mobile|CriOS/i.test(ua)
}

export function isTabletDevice(): boolean {
  const ua = navigator.userAgent
  return /iPad|Android(?!.*Mobile)/i.test(ua)
}

/** 是否为移动设备（手机或平板） */
export function isMobileOrTablet(): boolean {
  return isMobileDevice() || isTabletDevice()
}