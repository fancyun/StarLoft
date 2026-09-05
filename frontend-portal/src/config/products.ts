// 平台产品目录配置
// 每新增一个产品，在此登记 key（URL 路径段）、展示信息与进入控制台的地址，
// 再在 router 中增加对应路由（复用 ProductPage.vue）即可上线。
export interface ProductFeature {
  title: string
  desc: string
}

export interface Product {
  /** URL 路径段，如 kyc、cs、sms（对应 /product/xxx 门户页） */
  key: string
  name: string
  english: string
  tagline: string
  description: string
  features: ProductFeature[]
  scenarios: string[]
  /** 控制台内对应产品页地址 */
  consolePath: string
  /** available 正常使用 / coming-soon 建设中 */
  status: 'available' | 'coming-soon'
  /** 卡片图标（内联 SVG 字符串） */
  icon: string
}

const shieldIcon =
  '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>'
const serverIcon =
  '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/><line x1="6" y1="6.5" x2="6.01" y2="6.5"/><line x1="6" y1="17.5" x2="6.01" y2="17.5"/></svg>'
const smsIcon =
  '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/></svg>'

export const products: Product[] = [
  {
    key: 'kyc',
    name: '实名认证',
    english: 'Real-name Authentication',
    tagline: '三要素核验 · 活体识别 · API 开放',
    description:
      '提供身份证三要素实名认证服务（姓名 + 身份证号 + 人脸活体识别），支持 Web 与 API 两种调用方式，认证结果异步通知、安全回流，适用于各类实名场景。',
    features: [
      { title: '三要素核验', desc: '姓名、身份证号与人脸活体多维度交叉核验，结果可信' },
      { title: '活体检测', desc: '集成活体识别能力，有效防范照片、视频等攻击手段' },
      { title: 'API 开放', desc: 'RESTful 接口 + HMAC-SHA256 签名鉴权，快速接入下游系统' },
      { title: '灵活计费', desc: '资源包 + 余额组合计费，先扣资源包、余额兜底' },
      { title: '结果回流', desc: '异步通知 + 同步跳转双通道，认证结果安全送达' }
    ],
    scenarios: ['账号实名', '电商交易', '金融开户', '内容平台', '直播平台', '业务风控'],
    consolePath: 'https://console.starloft.cn/kyc',
    status: 'available',
    icon: shieldIcon
  },
  {
    key: 'cs',
    name: '云服务器',
    english: 'Cloud Server',
    tagline: '弹性计算 · 秒级交付',
    description:
      '提供弹性云服务器实例，按需选择配置与地域，快速部署与弹性扩容，满足业务上云需求。产品建设中，敬请期待。',
    features: [
      { title: '弹性配置', desc: '多档 CPU / 内存规格，按业务需求灵活选择' },
      { title: '秒级交付', desc: '下单后快速开通实例，即刻投入使用' },
      { title: '安全隔离', desc: '实例网络隔离与安全组防护，保障数据安全' },
      { title: '灵活计费', desc: '支持包年包月与按量计费多种模式' }
    ],
    scenarios: ['网站部署', '应用托管', '数据处理', '业务测试'],
    consolePath: 'https://console.starloft.cn/cs',
    status: 'coming-soon',
    icon: serverIcon
  },
  {
    key: 'sms',
    name: '短信服务',
    english: 'Short Message Service',
    tagline: '验证码短信 · 通知短信',
    description:
      '提供短信发送能力，覆盖验证码、通知、营销等场景，高到达率、实时状态回执。产品建设中，敬请期待。',
    features: [
      { title: '验证码短信', desc: '登录、注册、找回密码等验证码场景快速触达' },
      { title: '状态回执', desc: '实时获取短信发送状态，失败可追踪' },
      { title: '内容审核', desc: '发送内容安全审核，规避违规风险' },
      { title: '灵活计费', desc: '按条计费，支持资源包预购' }
    ],
    scenarios: ['注册验证', '登录验证', '通知提醒', '营销触达'],
    consolePath: 'https://console.starloft.cn/sms',
    status: 'coming-soon',
    icon: smsIcon
  }
]

export const productByKey = (key: string): Product | undefined =>
  products.find((p) => p.key === key)
