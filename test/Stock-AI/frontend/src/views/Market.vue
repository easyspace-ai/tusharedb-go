<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import {
  NCard,
  NGrid,
  NGi,
  NStatistic,
  NTabs,
  NTabPane,
  NDataTable,
  NTag,
  NTimeline,
  NTimelineItem,
  NSpin,
  NButton,
  NModal,
  NInput,
  NSpace,
  NScrollbar,
  NAlert,
  useMessage
} from 'naive-ui'
import { h } from 'vue'
import {
  GetMarketIndex,
  GetIndustryRank,
  GetMoneyFlow,
  GetNewsList,
  GetLongTigerRank,
  GetHotTopics,
  GetTradingTimeInfo,
  AIRecommendStream,
  AIChatStream,
  GetConfig,
  GetAShareSentiment,
  GetCachedMarketData,
  MarkFirstLoadComplete,
  IsFirstLoad
} from '../../wailsjs/go/main/App'
import SentimentGauge from '../components/SentimentGauge.vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const message = useMessage()
const indexes = ref([])
const industries = ref([])
const moneyFlow = ref([])
const newsList = ref([])
const longTiger = ref([])
const hotTopics = ref([])
const sentimentData = ref(null)
const loading = ref(false)
let quoteRefreshTimer = null
let newsRefreshTimer = null

// AI相关
const showAIModal = ref(false)
const aiEnabled = ref(false)
const aiLoading = ref(false)
const aiResponse = ref('')
const aiQuestion = ref('')
const aiMessages = ref([])
const aiScrollbarRef = ref(null)
const eventOffFns = []

const industryColumns = [
  { title: '行业', key: 'name', width: 120 },
  {
    title: '涨跌幅',
    key: 'changePercent',
    width: 100,
    render: (row) => {
      const val = row.changePercent
      if (val === undefined) return '-'
      const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
      return h('span', { style: { color, fontWeight: 'bold' } }, (val > 0 ? '+' : '') + val.toFixed(2) + '%')
    }
  },
  { title: '领涨股', key: 'leadStock', width: 100 }
]

const moneyColumns = [
  { title: '名称', key: 'name', width: 100 },
  { title: '代码', key: 'code', width: 80 },
  {
    title: '主力净流入',
    key: 'mainFlow',
    width: 120,
    render: (row) => {
      const val = row.mainFlow
      if (val === undefined) return '-'
      const color = val > 0 ? '#f5222d' : '#52c41a'
      return h('span', { style: { color } }, formatMoney(val))
    }
  },
  {
    title: '超大单净流入',
    key: 'superFlow',
    width: 120,
    render: (row) => {
      const val = row.superFlow
      if (val === undefined) return '-'
      const color = val > 0 ? '#f5222d' : '#52c41a'
      return h('span', { style: { color } }, formatMoney(val))
    }
  }
]

const longTigerColumns = [
  { title: '排名', key: 'rank', width: 60 },
  { title: '名称', key: 'name', width: 100 },
  { title: '代码', key: 'code', width: 80 },
  {
    title: '涨跌幅',
    key: 'changePercent',
    width: 80,
    render: (row) => {
      const val = row.changePercent
      const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
      return h('span', { style: { color } }, (val > 0 ? '+' : '') + val.toFixed(2) + '%')
    }
  },
  { title: '买入额', key: 'buyAmount', width: 100 },
  { title: '卖出额', key: 'sellAmount', width: 100 }
]

const hotTopicColumns = [
  { title: '排名', key: 'rank', width: 60 },
  { title: '话题', key: 'title', width: 200 },
  { title: '阅读数', key: 'readCount', width: 100, render: (row) => formatCount(row.readCount) },
  { title: '讨论数', key: 'postCount', width: 100, render: (row) => formatCount(row.postCount) }
]

const formatMoney = (val) => {
  if (val === undefined || val === null) return '-'
  const absVal = Math.abs(val)
  const sign = val >= 0 ? '' : '-'
  if (absVal >= 100000000) return sign + (absVal / 100000000).toFixed(2) + '亿'
  if (absVal >= 10000) return sign + (absVal / 10000).toFixed(2) + '万'
  return val.toFixed(2)
}

const formatCount = (val) => {
  if (!val) return '-'
  if (val >= 10000) return (val / 10000).toFixed(1) + '万'
  return val.toString()
}

const getImportanceType = (importance) => {
  if (importance === 'high') return 'error'
  if (importance === 'medium') return 'warning'
  return 'default'
}

// 检查AI配置
const checkAIConfig = async () => {
  try {
    const config = await GetConfig()
    aiEnabled.value = config.aiEnabled && config.aiApiKey
  } catch (e) {
    console.error('获取配置失败:', e)
  }
}

// 打开AI分析弹窗
const openAIAnalysis = () => {
  showAIModal.value = true
  aiMessages.value = []
  aiResponse.value = ''

  if (!aiEnabled.value) {
    message.warning('AI功能未启用，请在设置中配置')
    return
  }

  // 自动开始分析
  startAIRecommend()
}

// 开始AI市场分析
const startAIRecommend = async () => {
  if (aiLoading.value) return

  aiLoading.value = true
  aiResponse.value = ''

  // 添加用户消息
  aiMessages.value.push({
    role: 'user',
    content: '根据当前市场数据分析热点方向'
  })

  // 添加AI响应占位
  aiMessages.value.push({
    role: 'assistant',
    content: ''
  })

  scrollToBottom()

  try {
    await AIRecommendStream()
  } catch (e) {
    message.error('AI分析失败: ' + e)
    aiLoading.value = false
  }
}

// 发送AI问题
const sendAIQuestion = async () => {
  if (!aiQuestion.value.trim() || aiLoading.value) return

  const question = aiQuestion.value.trim()
  aiQuestion.value = ''

  aiLoading.value = true
  aiResponse.value = ''

  // 添加用户消息
  aiMessages.value.push({
    role: 'user',
    content: question
  })

  // 添加AI响应占位
  aiMessages.value.push({
    role: 'assistant',
    content: ''
  })

  scrollToBottom()

  try {
    await AIChatStream({
      message: question,
      sessionId: 'market',
      stockCode: ''
    })
  } catch (e) {
    message.error('发送失败: ' + e)
    aiLoading.value = false
  }
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (aiScrollbarRef.value) {
      aiScrollbarRef.value.scrollTo({ top: 999999, behavior: 'smooth' })
    }
  })
}

// 处理AI流式响应
const handleAIStream = (content) => {
  aiResponse.value += content
  if (aiMessages.value.length > 0) {
    aiMessages.value[aiMessages.value.length - 1].content = aiResponse.value
  }
  scrollToBottom()
}

const handleAIDone = () => {
  aiLoading.value = false
  aiResponse.value = ''
}

const handleAIError = (error) => {
  aiLoading.value = false
  message.error(error)
  if (aiMessages.value.length > 0 && aiMessages.value[aiMessages.value.length - 1].role === 'assistant') {
    aiMessages.value[aiMessages.value.length - 1].content = `错误: ${error}`
  }
}

// 格式化内容
const formatContent = (content) => {
  if (!content) return ''
  let html = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  html = html.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>')
  html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/^### (.+)$/gm, '<h4>$1</h4>')
  html = html.replace(/^## (.+)$/gm, '<h3>$1</h3>')
  html = html.replace(/^# (.+)$/gm, '<h2>$1</h2>')
  html = html.replace(/^- (.+)$/gm, '<li>$1</li>')
  html = html.replace(/\n/g, '<br>')
  return html
}

// 加载缓存数据（快速启动）
const loadCachedData = async () => {
  try {
    const cached = await GetCachedMarketData()
    if (cached && cached.hasCache) {
      console.log('[Market] 加载缓存数据，缓存时间:', cached.cacheTime)
      indexes.value = cached.marketIndex || []
      industries.value = cached.industryRank || []
      moneyFlow.value = cached.moneyFlow || []
      newsList.value = cached.newsList || []
      longTiger.value = cached.longTigerRank || []
      hotTopics.value = cached.hotTopics || []
      return true
    }
    return false
  } catch (e) {
    console.error('加载缓存数据失败:', e)
    return false
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const [indexData, industryData, flowData, newsData, tigerData, topicData, sentiment] = await Promise.all([
      GetMarketIndex(),
      GetIndustryRank(),
      GetMoneyFlow(),
      GetNewsList(),
      GetLongTigerRank(),
      GetHotTopics(),
      GetAShareSentiment()
    ])
    indexes.value = indexData || []
    industries.value = industryData || []
    moneyFlow.value = flowData || []
    newsList.value = newsData || []
    longTiger.value = tigerData || []
    hotTopics.value = topicData || []
    sentimentData.value = sentiment

    // 首次加载完成后，标记切换到轮询模式
    const isFirst = await IsFirstLoad()
    if (isFirst) {
      await MarkFirstLoadComplete()
      console.log('[Market] 首次加载完成，已切换到轮询模式')
    }
  } catch (e) {
    console.error('加载市场数据失败:', e)
  } finally {
    loading.value = false
  }
}

// 刷新行情数据（指数、行业、资金流向、情绪）- 只在交易时间刷新
const refreshQuoteData = async () => {
  try {
    const [indexData, industryData, flowData, sentiment] = await Promise.all([
      GetMarketIndex(),
      GetIndustryRank(),
      GetMoneyFlow(),
      GetAShareSentiment()
    ])
    indexes.value = indexData || []
    industries.value = industryData || []
    moneyFlow.value = flowData || []
    sentimentData.value = sentiment
  } catch (e) {
    console.error('刷新行情数据失败:', e)
  }
}

// 刷新资讯数据（快讯、龙虎榜、热门话题）- 全天刷新
const refreshNewsData = async () => {
  try {
    const [newsData, tigerData, topicData] = await Promise.all([
      GetNewsList(),
      GetLongTigerRank(),
      GetHotTopics()
    ])
    newsList.value = newsData || []
    longTiger.value = tigerData || []
    hotTopics.value = topicData || []
  } catch (e) {
    console.error('刷新资讯数据失败:', e)
  }
}

onMounted(async () => {
  await checkAIConfig()

  // 1. 先尝试加载缓存数据（毫秒级）
  const hasCached = await loadCachedData()
  if (hasCached) {
    console.log('[Market] 已显示缓存数据，后台刷新最新数据...')
    // 后台静默刷新最新数据
    loadData()
  } else {
    // 没有缓存，直接加载最新数据
    console.log('[Market] 无缓存，直接加载最新数据...')
    await loadData()
  }

  // 启动行情数据刷新（只在交易时间）
  startQuoteRefresh()
  // 启动资讯数据刷新（全天）
  startNewsRefresh()

  eventOffFns.push(EventsOn('ai-chat-stream', handleAIStream))
  eventOffFns.push(EventsOn('ai-chat-done', handleAIDone))
  eventOffFns.push(EventsOn('ai-chat-error', handleAIError))
})

// 行情数据刷新 - 只在交易时间刷新
const startQuoteRefresh = async () => {
  const checkAndRefresh = async () => {
    try {
      const timeInfo = await GetTradingTimeInfo()
      if (timeInfo.refreshInterval > 0) {
        await refreshQuoteData()
        quoteRefreshTimer = setTimeout(checkAndRefresh, timeInfo.refreshInterval * 1000)
      } else {
        // 非交易时间，每分钟检查一次是否进入交易时间
        quoteRefreshTimer = setTimeout(checkAndRefresh, 60000)
      }
    } catch (e) {
      console.error('刷新失败:', e)
      quoteRefreshTimer = setTimeout(checkAndRefresh, 30000)
    }
  }
  checkAndRefresh()
}

// 资讯数据刷新 - 全天刷新，非交易时间间隔长一些
const startNewsRefresh = async () => {
  const checkAndRefresh = async () => {
    try {
      const timeInfo = await GetTradingTimeInfo()
      await refreshNewsData()
      // 交易时间3分钟刷新一次，非交易时间5分钟刷新一次
      const interval = timeInfo.isTradingTime ? 180000 : 300000
      newsRefreshTimer = setTimeout(checkAndRefresh, interval)
    } catch (e) {
      console.error('刷新资讯失败:', e)
      newsRefreshTimer = setTimeout(checkAndRefresh, 300000)
    }
  }
  // 首次延迟30秒再刷新，避免和loadData重复
  newsRefreshTimer = setTimeout(checkAndRefresh, 30000)
}

onUnmounted(() => {
  if (quoteRefreshTimer) {
    clearTimeout(quoteRefreshTimer)
  }
  if (newsRefreshTimer) {
    clearTimeout(newsRefreshTimer)
  }
  eventOffFns.forEach((off) => typeof off === 'function' && off())
  eventOffFns.length = 0
})
</script>

<template>
  <div class="market-page">
    <n-spin :show="loading">
      <!-- 顶部区域：情绪仪表盘 + 指数概览 -->
      <div class="top-section">
        <!-- 左侧：情绪仪表盘 -->
        <div class="gauge-section">
          <n-card size="small" :bordered="false" class="gauge-card">
            <SentimentGauge :sentiment="sentimentData" label="A股市场情绪" />
          </n-card>
        </div>
        <!-- 右侧：指数卡片 -->
        <div class="index-section">
          <div class="index-list">
            <div v-for="idx in indexes" :key="idx.code" class="index-item">
              <span class="index-name">{{ idx.name }}</span>
              <span class="index-price">{{ idx.price?.toFixed(2) }}</span>
              <span class="index-change" :class="{ up: idx.changePercent > 0, down: idx.changePercent < 0 }">
                {{ (idx.changePercent > 0 ? '+' : '') + idx.changePercent?.toFixed(2) }}%
              </span>
            </div>
          </div>
        </div>
        <!-- AI按钮 -->
        <n-button type="warning" @click="openAIAnalysis" style="margin-left: 16px; flex-shrink: 0;">AI市场分析</n-button>
      </div>

      <!-- 主要内容 -->
      <n-grid :cols="2" :x-gap="16">
        <!-- 左侧：快讯 -->
        <n-gi>
          <n-card title="财经快讯" size="small" :bordered="false" style="height: 500px; overflow: auto;">
            <n-timeline>
              <n-timeline-item
                v-for="news in newsList"
                :key="news.id"
                :type="getImportanceType(news.importance)"
                :time="news.time"
              >
                <template #header>
                  <n-tag v-if="news.importance === 'high'" type="error" size="small" style="margin-right: 8px;">重要</n-tag>
                  <span>{{ news.title || news.content?.substring(0, 50) }}</span>
                </template>
                <div class="news-content">{{ news.content }}</div>
              </n-timeline-item>
            </n-timeline>
            <div v-if="newsList.length === 0" class="empty-tip">暂无快讯数据</div>
          </n-card>
        </n-gi>

        <!-- 右侧：数据面板 -->
        <n-gi>
          <n-card :bordered="false" size="small">
            <n-tabs type="line" animated>
              <n-tab-pane name="industry" tab="行业排行">
                <n-data-table
                  :columns="industryColumns"
                  :data="industries"
                  :bordered="false"
                  size="small"
                  :max-height="400"
                />
              </n-tab-pane>
              <n-tab-pane name="money" tab="资金流向">
                <n-data-table
                  :columns="moneyColumns"
                  :data="moneyFlow"
                  :bordered="false"
                  size="small"
                  :max-height="400"
                />
              </n-tab-pane>
              <n-tab-pane name="tiger" tab="龙虎榜">
                <n-data-table
                  :columns="longTigerColumns"
                  :data="longTiger"
                  :bordered="false"
                  size="small"
                  :max-height="400"
                />
              </n-tab-pane>
              <n-tab-pane name="hot" tab="热门话题">
                <n-data-table
                  :columns="hotTopicColumns"
                  :data="hotTopics"
                  :bordered="false"
                  size="small"
                  :max-height="400"
                />
              </n-tab-pane>
            </n-tabs>
          </n-card>
        </n-gi>
      </n-grid>
    </n-spin>

    <!-- AI分析弹窗 -->
    <n-modal v-model:show="showAIModal" preset="card" title="AI市场分析" style="width: 900px;">
      <div class="ai-analysis-container">
        <n-alert v-if="!aiEnabled" type="warning" style="margin-bottom: 16px;">
          AI功能未启用，请前往「设置」页面配置AI服务。
        </n-alert>

        <!-- 对话区域 -->
        <div class="ai-chat-container">
          <n-scrollbar ref="aiScrollbarRef" style="max-height: 400px;">
            <div class="ai-messages">
              <div v-if="aiMessages.length === 0" class="ai-empty-tip">
                点击下方按钮获取AI市场分析，或直接输入问题进行对话
              </div>
              <div
                v-for="(msg, index) in aiMessages"
                :key="index"
                :class="['ai-message', msg.role]"
              >
                <div class="ai-message-role">{{ msg.role === 'user' ? '我' : 'AI' }}</div>
                <div class="ai-message-content">
                  <n-spin v-if="msg.role === 'assistant' && aiLoading && index === aiMessages.length - 1 && !msg.content" size="small" />
                  <div v-else class="markdown-content" v-html="formatContent(msg.content)"></div>
                </div>
              </div>
            </div>
          </n-scrollbar>
        </div>

        <!-- 输入区域 -->
        <div class="ai-input-area">
          <n-space vertical style="width: 100%;">
            <n-space>
              <n-button type="primary" :disabled="!aiEnabled || aiLoading" @click="startAIRecommend">
                重新分析
              </n-button>
            </n-space>
            <n-input
              v-model:value="aiQuestion"
              type="textarea"
              placeholder="输入问题继续对话，如：今天哪些板块值得关注？"
              :autosize="{ minRows: 2, maxRows: 3 }"
              :disabled="!aiEnabled || aiLoading"
              @keyup.enter.exact="sendAIQuestion"
            />
            <n-button type="info" :loading="aiLoading" :disabled="!aiEnabled || !aiQuestion.trim()" @click="sendAIQuestion">
              发送问题
            </n-button>
          </n-space>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
.market-page {
  height: 100%;
}

.top-section {
  display: flex;
  align-items: flex-start;
  margin-bottom: 16px;
  gap: 16px;
}

.gauge-section {
  flex-shrink: 0;
  width: 200px;
}

.gauge-card {
  background: rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  justify-content: center;
}

.index-section {
  flex: 1;
}

.index-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.index-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  white-space: nowrap;
}

.index-name {
  font-size: 13px;
  color: #999;
}

.index-price {
  font-size: 16px;
  font-weight: bold;
  color: #fff;
}

.index-change {
  font-size: 13px;
  font-weight: bold;
}

.index-change.up {
  color: #f5222d;
}

.index-change.down {
  color: #52c41a;
}

.news-content {
  font-size: 13px;
  color: #999;
  line-height: 1.6;
  margin-top: 4px;
}

.empty-tip {
  text-align: center;
  color: #666;
  padding: 40px;
}

/* AI分析样式 */
.ai-analysis-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ai-chat-container {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  padding: 16px;
  min-height: 200px;
}

.ai-messages {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ai-empty-tip {
  text-align: center;
  color: #999;
  padding: 40px;
}

.ai-message {
  display: flex;
  gap: 12px;
}

.ai-message.user {
  flex-direction: row-reverse;
}

.ai-message-role {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  flex-shrink: 0;
}

.ai-message.user .ai-message-role {
  background: #18a058;
  color: white;
}

.ai-message.assistant .ai-message-role {
  background: #2080f0;
  color: white;
}

.ai-message-content {
  max-width: 80%;
  padding: 10px 14px;
  border-radius: 8px;
  line-height: 1.6;
  font-size: 14px;
}

.ai-message.user .ai-message-content {
  background: #18a058;
  color: white;
  border-bottom-right-radius: 0;
}

.ai-message.assistant .ai-message-content {
  background: rgba(255, 255, 255, 0.1);
  border-bottom-left-radius: 0;
}

.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4) {
  margin: 10px 0 6px 0;
}

.markdown-content :deep(strong) {
  color: #18a058;
}

.markdown-content :deep(li) {
  margin: 4px 0;
  margin-left: 16px;
}

.ai-input-area {
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}
</style>
