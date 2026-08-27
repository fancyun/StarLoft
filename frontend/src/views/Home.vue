<template>
  <div class="home-container">
    <!-- 顶部导航栏 -->
    <header class="header">
      <div class="header-content">
        <div class="logo">
          <span class="brand-logo">SL</span>
          <div class="logo-text">
            <h1>星楼KYC</h1>
            <span class="subtitle">KYC Authentication Platform</span>
          </div>
        </div>
        <nav class="nav">
          <router-link to="/login" class="nav-btn primary">登录</router-link>
          <router-link to="/register" class="nav-btn">注册</router-link>
        </nav>
      </div>
    </header>

    <!-- 主横幅区域 -->
    <section class="hero">
      <div class="hero-content">
        <h2 class="hero-title">安全、快速、可靠的实名认证服务</h2>
        <p class="hero-subtitle">为您的业务提供专业的实名认证解决方案，支持银联实名认证</p>
        <div class="hero-actions">
          <router-link to="/register" class="btn-large primary">立即开始</router-link>
          <a href="#features" class="btn-large secondary">了解更多</a>
        </div>
        <div class="hero-stats">
          <div class="stat-item">
            <div class="stat-number">99.9%</div>
            <div class="stat-label">服务可用性</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">&lt;3s</div>
            <div class="stat-label">平均响应时间</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">7×24</div>
            <div class="stat-label">技术支持</div>
          </div>
        </div>
      </div>
    </section>

    <!-- 功能特性 -->
    <section id="features" class="features">
      <div class="section-content">
        <h2 class="section-title">核心功能</h2>
        <p class="section-subtitle">全方位满足您的实名认证需求</p>
        
        <div class="feature-grid">
          <div class="feature-card">
            <div class="feature-icon">🔐</div>
            <h3>银联实名认证</h3>
            <p>对接银联权威数据源，支持身份证实名认证，准确率高达99%以上</p>
          </div>
          
          <div class="feature-card">
            <div class="feature-icon">⚡</div>
            <h3>快速响应</h3>
            <p>秒级响应，平均认证时间小于3秒，满足高并发业务场景需求</p>
          </div>
          
          <div class="feature-card">
            <div class="feature-icon">🛡️</div>
            <h3>数据安全</h3>
            <p>采用银行级加密技术，确保用户隐私数据安全，符合国家信息安全标准</p>
          </div>
          
          <div class="feature-card">
            <div class="feature-icon">📊</div>
            <h3>数据统计</h3>
            <p>提供详细的认证记录和数据统计，帮助您更好地了解业务情况</p>
          </div>
          
          <div class="feature-card">
            <div class="feature-icon">🔌</div>
            <h3>简单接入</h3>
            <p>提供完善的API文档和SDK，支持多种编程语言，快速完成接入</p>
          </div>
          
          <div class="feature-card">
            <div class="feature-icon">💰</div>
            <h3>灵活计费</h3>
            <p>按需充值，按次计费，支持折扣优惠，降低您的运营成本</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 使用流程 -->
    <section class="process">
      <div class="section-content">
        <h2 class="section-title">使用流程</h2>
        <p class="section-subtitle">四步即可开始使用</p>
        
        <div class="process-steps">
          <div class="step">
            <div class="step-number">1</div>
            <h3>注册账号</h3>
            <p>填写手机号完成注册</p>
          </div>
          
          <div class="step-arrow">→</div>
          
          <div class="step">
            <div class="step-number">2</div>
            <h3>账户充值</h3>
            <p>选择充值金额进行充值</p>
          </div>
          
          <div class="step-arrow">→</div>
          
          <div class="step">
            <div class="step-number">3</div>
            <h3>获取密钥</h3>
            <p>在个人中心获取API密钥</p>
          </div>
          
          <div class="step-arrow">→</div>
          
          <div class="step">
            <div class="step-number">4</div>
            <h3>开始使用</h3>
            <p>调用API进行实名认证</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 价格方案 -->
    <section class="pricing">
      <div class="section-content">
        <h2 class="section-title">价格方案</h2>
        <p class="section-subtitle">透明的计费标准，无隐藏费用</p>
        
        <div class="pricing-card">
          <div class="price-header">
            <h3>实名认证服务</h3>
            <div class="price">
              <span class="price-number">0.50</span>
              <span class="price-unit">元/次</span>
            </div>
          </div>
          <ul class="price-features">
            <li>✓ 银联实名认证</li>
            <li>✓ 秒级响应</li>
            <li>✓ 99.9%准确率</li>
            <li>✓ 7×24技术支持</li>
            <li>✓ 详细数据统计</li>
            <li>✓ API接口调用</li>
          </ul>
          <p class="price-note">* 支持批量购买折扣，充值越多优惠越多</p>
        </div>
      </div>
    </section>

    <!-- 底部 -->
    <footer class="footer">
      <div class="footer-content">
        <div class="footer-info">
          <h3>星楼KYC</h3>
          <p>为您的业务提供专业的实名认证解决方案</p>
        </div>
        <div class="footer-links">
          <div class="link-group">
            <h4>产品</h4>
            <a href="#features">核心功能</a>
            <a href="#process">使用流程</a>
            <a href="#pricing">价格方案</a>
          </div>
          <div class="link-group">
            <h4>支持</h4>
            <a href="#">API文档</a>
            <a href="#">常见问题</a>
            <a href="#">联系我们</a>
          </div>
          <div class="link-group">
            <h4>关于</h4>
            <router-link to="/login">用户登录</router-link>
            <router-link to="/register">注册账号</router-link>
          </div>
        </div>
      </div>
      <div class="footer-bottom">
        <p>&copy; 2024 星楼KYC. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
// 首页无需额外逻辑
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.home-container {
  min-height: 100vh;
  background: #ffffff;
}

/* 头部导航 */
.header {
  position: sticky;
  top: 0;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--border-color);
  z-index: 100;
  padding: 1rem 0;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo h1 {
  font-size: 1.5rem;
  color: #1d2129;
  margin-bottom: 0.25rem;
}

.logo .subtitle {
  font-size: 0.75rem;
  color: #86909c;
}

.nav {
  display: flex;
  gap: 1rem;
}

.nav-btn {
  padding: 0.5rem 1.5rem;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.3s;
}

.nav-btn.primary {
  background: #006eff;
  color: white;
}

.nav-btn.primary:hover {
  background: #0052d9;
}

.nav-btn:not(.primary) {
  color: #4e5969;
  border: 1px solid #e5e6eb;
}

.nav-btn:not(.primary):hover {
  background: #f2f3f5;
}

/* 主横幅 */
.hero {
  background: linear-gradient(135deg, #0052d9 0%, #006eff 55%, #3a8dff 100%);
  color: white;
  padding: 6rem 2rem;
  text-align: center;
}

.hero-content {
  max-width: 1200px;
  margin: 0 auto;
}

.hero-title {
  font-size: 3rem;
  font-weight: 700;
  margin-bottom: 1.5rem;
  line-height: 1.2;
}

.hero-subtitle {
  font-size: 1.25rem;
  margin-bottom: 2.5rem;
  opacity: 0.9;
}

.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  margin-bottom: 4rem;
}

.btn-large {
  padding: 1rem 2.5rem;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 600;
  font-size: 1.1rem;
  transition: all 0.3s;
  display: inline-block;
}

.btn-large.primary {
  background: white;
  color: #006eff;
}

.btn-large.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(0,0,0,0.2);
}

.btn-large.secondary {
  background: rgba(255,255,255,0.2);
  color: white;
  border: 2px solid white;
}

.btn-large.secondary:hover {
  background: rgba(255,255,255,0.3);
}

.hero-stats {
  display: flex;
  justify-content: center;
  gap: 4rem;
}

.stat-item {
  text-align: center;
}

.stat-number {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 0.9rem;
  opacity: 0.8;
}

/* 通用区块样式 */
.section-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 5rem 2rem;
}

.section-title {
  font-size: 2.5rem;
  text-align: center;
  margin-bottom: 1rem;
  color: #1d2129;
}

.section-subtitle {
  text-align: center;
  color: #86909c;
  font-size: 1.1rem;
  margin-bottom: 3rem;
}

/* 功能特性 */
.features {
  background: #f5f7fa;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.feature-card {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  transition: all 0.3s;
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
}

.feature-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.feature-card h3 {
  font-size: 1.25rem;
  margin-bottom: 1rem;
  color: #1d2129;
}

.feature-card p {
  color: #4e5969;
  line-height: 1.6;
}

/* 使用流程 */
.process-steps {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 2rem;
}

.step {
  text-align: center;
  flex: 0 0 200px;
}

.step-number {
  width: 60px;
  height: 60px;
  background: #006eff;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0 auto 1rem;
}

.step h3 {
  font-size: 1.1rem;
  margin-bottom: 0.5rem;
  color: #1d2129;
}

.step p {
  color: #4e5969;
  font-size: 0.9rem;
}

.step-arrow {
  font-size: 2rem;
  color: #c9cdd4;
  flex: 0 0 auto;
}

/* 价格方案 */
.pricing {
  background: #f5f7fa;
}

.pricing-card {
  max-width: 500px;
  margin: 0 auto;
  background: white;
  border-radius: 12px;
  padding: 3rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.06);
}

.price-header {
  text-align: center;
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 2px solid var(--border-color);
}

.price-header h3 {
  font-size: 1.5rem;
  margin-bottom: 1.5rem;
  color: #1d2129;
}

.price {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 0.5rem;
}

.price-number {
  font-size: 3.5rem;
  font-weight: 700;
  color: #006eff;
}

.price-unit {
  font-size: 1.2rem;
  color: #86909c;
}

.price-features {
  list-style: none;
  margin-bottom: 2rem;
}

.price-features li {
  padding: 0.75rem 0;
  color: #4e5969;
  font-size: 1rem;
}

.price-note {
  text-align: center;
  color: #86909c;
  font-size: 0.875rem;
}

/* 底部 */
.footer {
  background: #1f2937;
  color: #9ca3af;
  padding: 3rem 0 1rem;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem 2rem;
  display: grid;
  grid-template-columns: 2fr 3fr;
  gap: 3rem;
}

.footer-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 1rem;
}

.footer-info h3 {
  color: white;
  margin-bottom: 0;
}

.footer-links {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}

.link-group h4 {
  color: white;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.link-group a {
  display: block;
  color: #9ca3af;
  text-decoration: none;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  transition: color 0.3s;
}

.link-group a:hover {
  color: white;
}

.footer-bottom {
  text-align: center;
  padding-top: 2rem;
  border-top: 1px solid #374151;
  font-size: 0.875rem;
}

/* 响应式 */
@media (max-width: 768px) {
  .hero-title {
    font-size: 2rem;
  }
  
  .hero-subtitle {
    font-size: 1rem;
  }
  
  .hero-stats {
    flex-direction: column;
    gap: 2rem;
  }
  
  .hero-actions {
    flex-direction: column;
  }
  
  .feature-grid {
    grid-template-columns: 1fr;
  }
  
  .process-steps {
    flex-direction: column;
  }
  
  .step-arrow {
    transform: rotate(90deg);
  }
  
  .footer-content {
    grid-template-columns: 1fr;
  }
  
  .footer-links {
    grid-template-columns: 1fr;
  }
}
</style>
