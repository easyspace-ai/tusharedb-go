import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import naive from 'naive-ui'
import App from './App.vue'
import './style.css'

// 路由配置
const routes = [
  { path: '/', name: 'Stock', component: () => import('./views/Stock.vue') },
  // 市场行情 - 按国家分类
  { path: '/market', name: 'Market', component: () => import('./views/Market.vue') },
  { path: '/market/us', name: 'MarketUS', component: () => import('./views/MarketCountry.vue'), props: { country: 'us' } },
  { path: '/market/jp', name: 'MarketJP', component: () => import('./views/MarketCountry.vue'), props: { country: 'jp' } },
  { path: '/market/kr', name: 'MarketKR', component: () => import('./views/MarketCountry.vue'), props: { country: 'kr' } },
  { path: '/market/hk', name: 'MarketHK', component: () => import('./views/MarketCountry.vue'), props: { country: 'hk' } },
  { path: '/market/tw', name: 'MarketTW', component: () => import('./views/MarketCountry.vue'), props: { country: 'tw' } },
  { path: '/market/uk', name: 'MarketUK', component: () => import('./views/MarketCountry.vue'), props: { country: 'uk' } },
  { path: '/market/de', name: 'MarketDE', component: () => import('./views/MarketCountry.vue'), props: { country: 'de' } },
  { path: '/market/fr', name: 'MarketFR', component: () => import('./views/MarketCountry.vue'), props: { country: 'fr' } },
  { path: '/market/au', name: 'MarketAU', component: () => import('./views/MarketCountry.vue'), props: { country: 'au' } },
  { path: '/market/in', name: 'MarketIN', component: () => import('./views/MarketCountry.vue'), props: { country: 'in' } },
  { path: '/market/sg', name: 'MarketSG', component: () => import('./views/MarketCountry.vue'), props: { country: 'sg' } },
  { path: '/market/ca', name: 'MarketCA', component: () => import('./views/MarketCountry.vue'), props: { country: 'ca' } },
  // 期货市场
  { path: '/futures', name: 'Futures', component: () => import('./views/Futures.vue') },
  { path: '/futures/shfe', name: 'FuturesSHFE', component: () => import('./views/FuturesExchange.vue'), props: { exchange: 'SHFE' } },
  { path: '/futures/dce', name: 'FuturesDCE', component: () => import('./views/FuturesExchange.vue'), props: { exchange: 'DCE' } },
  { path: '/futures/czce', name: 'FuturesCZCE', component: () => import('./views/FuturesExchange.vue'), props: { exchange: 'CZCE' } },
  { path: '/futures/cffex', name: 'FuturesCFFEX', component: () => import('./views/FuturesExchange.vue'), props: { exchange: 'CFFEX' } },
  { path: '/futures/ine', name: 'FuturesINE', component: () => import('./views/FuturesExchange.vue'), props: { exchange: 'INE' } },
  // 外汇市场
  { path: '/forex', name: 'Forex', component: () => import('./views/Forex.vue') },
  { path: '/forex/major', name: 'ForexMajor', component: () => import('./views/ForexCategory.vue'), props: { category: 'major' } },
  { path: '/forex/cross', name: 'ForexCross', component: () => import('./views/ForexCategory.vue'), props: { category: 'cross' } },
  { path: '/forex/cny', name: 'ForexCNY', component: () => import('./views/ForexCategory.vue'), props: { category: 'cny' } },
  // 其他
  { path: '/fund', name: 'Fund', component: () => import('./views/Fund.vue') },
  { path: '/ai', name: 'AI', component: () => import('./views/AI.vue') },
  { path: '/ai-history', name: 'AIHistory', component: () => import('./views/AIHistory.vue') },
  { path: '/ai-analysis', name: 'AIAnalysis', component: () => import('./views/AIAnalysis.vue') },
  { path: '/plugin', name: 'Plugin', component: () => import('./views/Plugin.vue') },
  { path: '/prompt', name: 'Prompt', component: () => import('./views/Prompt.vue') },
  { path: '/settings', name: 'Settings', component: () => import('./views/Settings.vue') },
  { path: '/about', name: 'About', component: () => import('./views/About.vue') }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

const app = createApp(App)
app.use(router)
app.use(naive)
app.mount('#app')
