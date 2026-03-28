<script setup>
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import {
  NCard,
  NDataTable,
  NButton,
  NSpin,
  NGrid,
  NGi,
  NTimeline,
  NTimelineItem,
  NTag,
  NModal,
  NInput,
  NSpace,
  NScrollbar,
  NAlert,
  useMessage
} from 'naive-ui'
import { h } from 'vue'
import {
  GetGlobalIndices,
  GetGlobalNews,
  GetGlobalMarketSentiment,
  AIChatStream,
  GetConfig,
  GetCachedGlobalMarketData,
  MarkFirstLoadComplete,
  IsFirstLoad
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import SentimentGauge from '../components/SentimentGauge.vue'

const props = defineProps({
  country: {
    type: String,
    required: true
  }
})

const message = useMessage()
const loading = ref(false)
const globalIndices = ref([])
const newsList = ref([])
const sentimentData = ref(null)

// AI相关
const showAIModal = ref(false)
const aiEnabled = ref(false)
const aiLoading = ref(false)
const aiResponse = ref('')
const aiQuestion = ref('')
const aiMessages = ref([])
const aiScrollbarRef = ref(null)
const eventOffFns = []

// 国家名称映射
const countryNames = {
  'us': '美国',
  'jp': '日本',
  'kr': '韩国',
  'hk': '中国香港',
  'tw': '中国台湾',
  'uk': '英国',
  'de': '德国',
  'fr': '法国',
  'au': '澳大利亚',
  'in': '印度',
  'sg': '新加坡',
  'ca': '加拿大'
}

// 过滤当前国家的指数
const countryIndices = computed(() => {
  return globalIndices.value.filter(idx => idx.country === props.country)
})

// 主要指数（前3个用于顶部展示）
const mainIndices = computed(() => {
  return countryIndices.value.slice(0, 3)
})

// 渲染涨跌幅
const renderChange = (val) => {
  if (val === undefined || val === null) return '-'
  const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
  return h('span', { style: { color, fontWeight: 'bold' } }, (val > 0 ? '+' : '') + val.toFixed(2) + '%')
}

// 列定义
const columns = [
  { title: '指数', key: 'nameCn', width: 180 },
  { title: '英文名', key: 'name', width: 200, ellipsis: true },
  { title: '代码', key: 'code', width: 100 },
  { title: '最新价', key: 'price', width: 120, render: (row) => row.price ? row.price.toFixed(2) : '-' },
  { title: '涨跌', key: 'change', width: 100, render: (row) => {
    const val = row.change
    if (val === undefined || val === null) return '-'
    const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
    return h('span', { style: { color } }, (val > 0 ? '+' : '') + val.toFixed(2))
  }},
  { title: '涨跌幅', key: 'changePercent', width: 100, render: (row) => renderChange(row.changePercent) },
  { title: '更新时间', key: 'updateTime', width: 100 }
]

// 获取新闻重要性类型
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

// 加载缓存数据（快速启动）
const loadCachedData = async () => {
  try {
    const cached = await GetCachedGlobalMarketData(props.country)
    if (cached && cached.hasCache) {
      console.log('[MarketCountry] 加载缓存数据，缓存时间:', cached.cacheTime)
      globalIndices.value = cached.globalIndices || []
      newsList.value = cached.news || []
      sentimentData.value = cached.sentiment
      return true
    }
    return false
  } catch (e) {
    console.error('加载缓存数据失败:', e)
    return false
  }
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const [indicesData, newsData, sentiment] = await Promise.all([
      GetGlobalIndices(),
      GetGlobalNews(props.country),
      GetGlobalMarketSentiment(props.country)
    ])
    globalIndices.value = indicesData || []
    newsList.value = newsData || []
    sentimentData.value = sentiment

    // 首次加载完成后，标记切换到轮询模式
    const isFirst = await IsFirstLoad()
    if (isFirst) {
      await MarkFirstLoadComplete()
      console.log('[MarketCountry] 首次加载完成，已切换到轮询模式')
    }
  } catch (e) {
    console.error('加载数据失败:', e)
    message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    const [indicesData, newsData, sentiment] = await Promise.all([
      GetGlobalIndices(),
      GetGlobalNews(props.country),
      GetGlobalMarketSentiment(props.country)
    ])
    globalIndices.value = indicesData || []
    newsList.value = newsData || []
    sentimentData.value = sentiment
    message.success('刷新成功')
  } catch (e) {
    message.error('刷新失败')
  } finally {
    loading.value = false
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
  startAIAnalysis()
}

// 开始AI分析
const startAIAnalysis = async () => {
  if (aiLoading.value) return

  aiLoading.value = true
  aiResponse.value = ''

  const countryName = countryNames[props.country] || props.country

  // 添加用户消息
  aiMessages.value.push({
    role: 'user',
    content: `分析${countryName}股市当前的市场状况和投资机会`
  })

  // 添加AI响应占位
  aiMessages.value.push({
    role: 'assistant',
    content: ''
  })

  scrollToBottom()

  try {
    // 构建上下文
    const indices = countryIndices.value.map(idx =>
      `${idx.nameCn || idx.name}: ${idx.price?.toFixed(2)} (${idx.changePercent > 0 ? '+' : ''}${idx.changePercent?.toFixed(2)}%)`
    ).join('\n')

    const sentimentInfo = sentimentData.value
      ? `市场情绪指数: ${sentimentData.value.value?.toFixed(0)} (${sentimentData.value.levelCn})`
      : ''

    const contextMessage = `请分析${countryName}股市当前状况。

当前指数数据:
${indices}

${sentimentInfo}

请从以下几个方面进行分析:
1. 市场整体表现
2. 主要指数走势分析
3. 市场情绪解读
4. 投资建议和风险提示`

    await AIChatStream({
      message: contextMessage,
      sessionId: `market_${props.country}`,
      stockCode: ''
    })
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

  aiMessages.value.push({
    role: 'user',
    content: question
  })

  aiMessages.value.push({
    role: 'assistant',
    content: ''
  })

  scrollToBottom()

  try {
    await AIChatStream({
      message: question,
      sessionId: `market_${props.country}`,
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

onMounted(async () => {
  await checkAIConfig()

  // 1. 先尝试加载缓存数据（毫秒级）
  const hasCached = await loadCachedData()
  if (hasCached) {
    console.log('[MarketCountry] 已显示缓存数据，后台刷新最新数据...')
    // 后台静默刷新最新数据
    loadData()
  } else {
    // 没有缓存，直接加载最新数据
    console.log('[MarketCountry] 无缓存，直接加载最新数据...')
    await loadData()
  }

  eventOffFns.push(EventsOn('ai-chat-stream', handleAIStream))
  eventOffFns.push(EventsOn('ai-chat-done', handleAIDone))
  eventOffFns.push(EventsOn('ai-chat-error', handleAIError))
})

// 监听国家变化，自动刷新数据
watch(() => props.country, async () => {
  // 先尝试加载缓存
  const hasCached = await loadCachedData()
  if (hasCached) {
    loadData()
  } else {
    await loadData()
  }
})

onUnmounted(() => {
  eventOffFns.forEach((off) => typeof off === 'function' && off())
  eventOffFns.length = 0
})
</script>

<template>
  <div class="market-country-page">
    <n-spin :show="loading">
      <!-- 顶部区域：指数概览 + 情绪仪表盘 -->
      <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px;">
        <n-grid :cols="4" :x-gap="12" :y-gap="12" style="flex: 1;">
          <!-- 情绪仪表盘 -->
          <n-gi>
            <n-card size="small" :bordered="false" class="gauge-card">
              <SentimentGauge :sentiment="sentimentData" :label="countryNames[country] + '市场情绪'" />
            </n-card>
          </n-gi>
          <!-- 主要指数 -->
          <n-gi v-for="idx in mainIndices" :key="idx.code">
            <n-card size="small" :bordered="false" class="index-card">
              <div class="index-name">{{ idx.nameCn || idx.name }}</div>
              <div class="index-price">{{ idx.price?.toFixed(2) }}</div>
              <div class="index-change" :class="{ up: idx.changePercent > 0, down: idx.changePercent < 0 }">
                {{ (idx.changePercent > 0 ? '+' : '') + idx.changePercent?.toFixed(2) }}%
              </div>
            </n-card>
          </n-gi>
        </n-grid>
        <n-button type="warning" @click="openAIAnalysis" style="margin-left: 16px;">AI市场分析</n-button>
      </div>

      <!-- 主要内容区域 -->
      <n-grid :cols="2" :x-gap="16">
        <!-- 左侧：财经快讯 -->
        <n-gi>
          <n-card :title="countryNames[country] + '财经快讯'" size="small" :bordered="false" style="height: 500px; overflow: auto;">
            <template #header-extra>
              <n-button type="primary" size="small" @click="refreshData" :loading="loading">
                刷新
              </n-button>
            </template>
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
            <div v-if="newsList.length === 0 && !loading" class="empty-tip">
              暂无快讯数据
            </div>
          </n-card>
        </n-gi>

        <!-- 右侧：指数详情 -->
        <n-gi>
          <n-card :title="countryNames[country] + '股市指数'" size="small" :bordered="false">
            <n-data-table
              :columns="columns"
              :data="countryIndices"
              :bordered="false"
              striped
              size="small"
              :max-height="420"
            />
            <div v-if="countryIndices.length === 0 && !loading" class="empty-tip">
              暂无{{ countryNames[country] }}股市数据
            </div>
          </n-card>
        </n-gi>
      </n-grid>
    </n-spin>

    <!-- AI分析弹窗 -->
    <n-modal v-model:show="showAIModal" preset="card" :title="countryNames[country] + 'AI市场分析'" style="width: 900px;">
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
              <n-button type="primary" :disabled="!aiEnabled || aiLoading" @click="startAIAnalysis">
                重新分析
              </n-button>
            </n-space>
            <n-input
              v-model:value="aiQuestion"
              type="textarea"
              :placeholder="`输入问题继续对话，如：${countryNames[country]}股市近期有哪些投资机会？`"
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
.market-country-page {
  height: 100%;
}

.index-card {
  text-align: center;
  background: rgba(255, 255, 255, 0.05);
}

.gauge-card {
  background: rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  justify-content: center;
}

.index-name {
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}

.index-price {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 4px;
}

.index-change {
  font-size: 14px;
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
