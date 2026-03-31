<script setup lang="ts">
import { ref } from 'vue'

const settings = ref({
  centerUrl: 'http://localhost:8080',
  wsUrl: 'ws://localhost:8080/ws',
  autoRefresh: true,
  refreshInterval: 10,
  theme: 'light',
  language: 'zh-CN'
})

const saved = ref(false)

function saveSettings() {
  localStorage.setItem('ofa-settings', JSON.stringify(settings.value))
  saved.value = true
  setTimeout(() => { saved.value = false }, 2000)
}

function resetSettings() {
  settings.value = {
    centerUrl: 'http://localhost:8080',
    wsUrl: 'ws://localhost:8080/ws',
    autoRefresh: true,
    refreshInterval: 10,
    theme: 'light',
    language: 'zh-CN'
  }
  localStorage.removeItem('ofa-settings')
}

// Load settings from localStorage
const savedSettings = localStorage.getItem('ofa-settings')
if (savedSettings) {
  try {
    settings.value = { ...settings.value, ...JSON.parse(savedSettings) }
  } catch (e) {
    console.error('Failed to load settings:', e)
  }
}
</script>

<template>
  <div>
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1>系统设置</h1>
        <p class="text-muted">配置 Dashboard 参数</p>
      </div>
    </div>

    <div class="settings-layout">
      <!-- Connection Settings -->
      <div class="settings-card">
        <div class="card-header">
          <h3>
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
            </svg>
            连接设置
          </h3>
        </div>
        <div class="card-body">
          <div class="form-group">
            <label class="form-label">Center 服务地址</label>
            <input type="text" class="form-input" v-model="settings.centerUrl" placeholder="http://localhost:8080">
            <span class="form-hint">REST API 服务地址</span>
          </div>

          <div class="form-group">
            <label class="form-label">WebSocket 地址</label>
            <input type="text" class="form-input" v-model="settings.wsUrl" placeholder="ws://localhost:8080/ws">
            <span class="form-hint">实时通信 WebSocket 地址</span>
          </div>
        </div>
      </div>

      <!-- Display Settings -->
      <div class="settings-card">
        <div class="card-header">
          <h3>
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            显示设置
          </h3>
        </div>
        <div class="card-body">
          <div class="form-group">
            <label class="form-label">界面语言</label>
            <select class="form-select" v-model="settings.language">
              <option value="zh-CN">简体中文</option>
              <option value="en-US">English</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">主题</label>
            <select class="form-select" v-model="settings.theme">
              <option value="light">浅色模式</option>
              <option value="dark">深色模式</option>
              <option value="auto">跟随系统</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Data Settings -->
      <div class="settings-card">
        <div class="card-header">
          <h3>
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <ellipse cx="12" cy="5" rx="9" ry="3"/>
              <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
              <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
            </svg>
            数据设置
          </h3>
        </div>
        <div class="card-body">
          <div class="toggle-group">
            <div class="toggle-item">
              <div class="toggle-info">
                <span class="toggle-label">自动刷新</span>
                <span class="toggle-desc">定期自动更新数据</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="settings.autoRefresh">
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <div class="form-group" v-if="settings.autoRefresh">
            <label class="form-label">刷新间隔 (秒)</label>
            <input type="number" class="form-input" v-model="settings.refreshInterval" min="5" max="60">
          </div>
        </div>
      </div>

      <!-- About -->
      <div class="settings-card">
        <div class="card-header">
          <h3>
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="16" x2="12" y2="12"/>
              <line x1="12" y1="8" x2="12.01" y2="8"/>
            </svg>
            关于
          </h3>
        </div>
        <div class="card-body">
          <div class="about-item">
            <span class="about-label">OFA Dashboard</span>
            <span class="about-value">v1.0.0</span>
          </div>
          <div class="about-item">
            <span class="about-label">Vue</span>
            <span class="about-value">3.4.x</span>
          </div>
          <div class="about-item">
            <span class="about-label">构建工具</span>
            <span class="about-value">Vite 5.x</span>
          </div>
          <div class="about-links">
            <a href="https://github.com/taomic2035/OFA" target="_blank" class="about-link">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
              </svg>
              GitHub
            </a>
            <a href="https://github.com/taomic2035/OFA/wiki" target="_blank" class="about-link">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
                <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
              </svg>
              文档
            </a>
          </div>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="settings-actions">
      <button class="btn btn-outline" @click="resetSettings">重置默认</button>
      <button class="btn btn-primary" @click="saveSettings">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/>
          <polyline points="17 21 17 13 7 13 7 21"/>
          <polyline points="7 3 7 8 15 8"/>
        </svg>
        {{ saved ? '已保存' : '保存设置' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.page-header {
  margin-bottom: 1.5rem;
}

.page-header h1 { margin: 0; font-size: 1.5rem; }
.page-header p { margin: 0.25rem 0 0 0; }

.settings-layout {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

@media (max-width: 768px) {
  .settings-layout { grid-template-columns: 1fr; }
}

.settings-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.card-header {
  padding: 1rem 1.25rem;
  background: var(--gray-50);
  border-bottom: 1px solid var(--gray-100);
}

.card-header h3 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  font-size: 1rem;
}

.card-body { padding: 1.25rem; }

.form-hint {
  display: block;
  font-size: 0.75rem;
  color: var(--gray-400);
  margin-top: 0.375rem;
}

.toggle-group { display: flex; flex-direction: column; gap: 1rem; }

.toggle-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toggle-info { display: flex; flex-direction: column; gap: 0.125rem; }
.toggle-label { font-weight: 500; }
.toggle-desc { font-size: 0.75rem; color: var(--gray-500); }

.toggle {
  position: relative;
  width: 48px;
  height: 26px;
  cursor: pointer;
}

.toggle input { opacity: 0; width: 0; height: 0; }

.toggle-slider {
  position: absolute;
  inset: 0;
  background: var(--gray-300);
  border-radius: 13px;
  transition: all 0.2s;
}

.toggle-slider::before {
  content: '';
  position: absolute;
  width: 20px;
  height: 20px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: all 0.2s;
}

.toggle input:checked + .toggle-slider { background: var(--primary); }
.toggle input:checked + .toggle-slider::before { transform: translateX(22px); }

.about-item {
  display: flex;
  justify-content: space-between;
  padding: 0.625rem 0;
  border-bottom: 1px solid var(--gray-100);
}

.about-item:last-of-type { border-bottom: none; }
.about-label { color: var(--gray-600); }
.about-value { font-weight: 500; }

.about-links {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--gray-100);
}

.about-link {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  color: var(--primary);
  text-decoration: none;
  font-size: 0.875rem;
}

.about-link:hover { text-decoration: underline; }

.settings-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
</style>