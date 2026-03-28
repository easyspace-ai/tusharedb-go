<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NLayout,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NConfigProvider,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
  darkTheme,
  zhCN
} from 'naive-ui'
import {
  TrendingUpOutline,
  StatsChartOutline,
  WalletOutline,
  SettingsOutline,
  InformationCircleOutline,
  SparklesOutline,
  TimeOutline,
  CashOutline,
  ExtensionPuzzleOutline,
  DocumentTextOutline,
  AnalyticsOutline
} from '@vicons/ionicons5'
import { h } from 'vue'
import AISidebar from './components/AISidebar.vue'
import UpdateChecker from './components/UpdateChecker.vue'
import DisclaimerModal from './components/DisclaimerModal.vue'
import { GetConfig } from '../wailsjs/go/main/App'

const router = useRouter()
const route = useRoute()
const collapsed = ref(false)
const currentTheme = ref('dark')
const customPrimary = ref('#18a058')
const aiSidebarCollapsed = ref(false)
const prefersDark = window.matchMedia ? window.matchMedia('(prefers-color-scheme: dark)') : null
const systemDark = ref(prefersDark ? prefersDark.matches : true)

const isDarkTheme = computed(() => {
  if (currentTheme.value === 'system') {
    return systemDark.value
  }
  return currentTheme.value === 'dark'
})

const themeOverrides = computed(() => ({
  common: {
    primaryColor: customPrimary.value,
    primaryColorHover: customPrimary.value,
    primaryColorSuppl: customPrimary.value
  }
}))

// 需要显示 AI 侧边栏的路由
const showAISidebarRoutes = ['/', '/fund', '/market']

// 是否显示 AI 侧边栏
const showAISidebar = computed(() => {
  const path = route.path
  return showAISidebarRoutes.some(r => path === r || path.startsWith('/market'))
})

const menuOptions = [
  {
    label: '自选股票',
    key: '/',
    icon: () => h(TrendingUpOutline)
  },
  {
    label: '市场行情',
    key: 'market',
    icon: () => h(StatsChartOutline),
    children: [
      { label: 'A股市场', key: '/market' },
      { label: '美国', key: '/market/us' },
      { label: '日本', key: '/market/jp' },
      { label: '韩国', key: '/market/kr' },
      { label: '中国香港', key: '/market/hk' },
      { label: '中国台湾', key: '/market/tw' },
      { label: '英国', key: '/market/uk' },
      { label: '德国', key: '/market/de' },
      { label: '法国', key: '/market/fr' },
      { label: '澳大利亚', key: '/market/au' },
      { label: '印度', key: '/market/in' },
      { label: '新加坡', key: '/market/sg' },
      { label: '加拿大', key: '/market/ca' }
    ]
  },
  {
    label: '基金管理',
    key: '/fund',
    icon: () => h(WalletOutline)
  },
  {
    label: 'AI 历史',
    key: '/ai-history',
    icon: () => h(TimeOutline)
  },
  {
    label: 'AI分析',
    key: '/ai-analysis',
    icon: () => h(AnalyticsOutline)
  },
  {
    label: '插件管理',
    key: '/plugin',
    icon: () => h(ExtensionPuzzleOutline)
  },
  {
    label: 'AI提示词',
    key: '/prompt',
    icon: () => h(DocumentTextOutline)
  },
  {
    label: '系统设置',
    key: '/settings',
    icon: () => h(SettingsOutline)
  },
  {
    label: '关于软件',
    key: '/about',
    icon: () => h(InformationCircleOutline)
  }
]

const handleMenuUpdate = (key) => {
  if (key.startsWith('/')) {
    router.push(key)
  }
}

const applyTheme = (cfg) => {
  if (!cfg) return
  if (cfg.theme) {
    currentTheme.value = cfg.theme
  }
  if (cfg.customPrimary) {
    customPrimary.value = cfg.customPrimary
  }
}

const loadThemeFromConfig = async () => {
  try {
    const cfg = await GetConfig()
    applyTheme(cfg)
  } catch (e) {
    console.error('加载主题配置失败:', e)
  }
}

const handleThemeEvent = (event) => {
  applyTheme(event?.detail || {})
}

const handleSystemThemeChange = (event) => {
  systemDark.value = event.matches
}

onMounted(() => {
  loadThemeFromConfig()
  prefersDark?.addEventListener('change', handleSystemThemeChange)
  window.addEventListener('stock-ai:theme-updated', handleThemeEvent)
})

onUnmounted(() => {
  prefersDark?.removeEventListener('change', handleSystemThemeChange)
  window.removeEventListener('stock-ai:theme-updated', handleThemeEvent)
})
</script>

<template>
  <n-config-provider :theme="isDarkTheme ? darkTheme : null" :theme-overrides="themeOverrides" :locale="zhCN">
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <n-layout has-sider style="height: 100vh;">
            <n-layout-sider
              bordered
              collapse-mode="width"
              :collapsed-width="64"
              :width="200"
              :collapsed="collapsed"
              show-trigger
              @collapse="collapsed = true"
              @expand="collapsed = false"
            >
              <div class="logo" :class="{ collapsed }">
                <span v-if="!collapsed">Stock AI</span>
                <span v-else>S</span>
              </div>
              <n-menu
                :collapsed="collapsed"
                :collapsed-width="64"
                :collapsed-icon-size="22"
                :options="menuOptions"
                :default-value="'/'"
                @update:value="handleMenuUpdate"
              />
            </n-layout-sider>
            <n-layout has-sider sider-placement="right">
              <n-layout-content content-style="padding: 16px; overflow: auto;">
                <router-view />
              </n-layout-content>
              <n-layout-sider
                v-if="showAISidebar"
                bordered
                :width="300"
                :collapsed-width="0"
                :collapsed="aiSidebarCollapsed"
                collapse-mode="width"
                show-trigger="bar"
                trigger-style="right: -12px;"
                collapsed-trigger-style="right: -12px;"
                @collapse="aiSidebarCollapsed = true"
                @expand="aiSidebarCollapsed = false"
              >
                <AISidebar />
              </n-layout-sider>
            </n-layout>
          </n-layout>
          <!-- 更新检测组件 -->
          <UpdateChecker />
          <!-- 免责声明弹窗 -->
          <DisclaimerModal />
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  width: 100%;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: bold;
  color: #18a058;
  border-bottom: 1px solid rgba(255, 255, 255, 0.09);
}

.logo.collapsed {
  font-size: 24px;
}
</style>
