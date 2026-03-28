<script setup>
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import {
  NCard,
  NTabs,
  NTabPane,
  NDataTable,
  NButton,
  NCollapse,
  NCollapseItem,
  NSpin,
  NModal,
  NInput,
  NSpace,
  NScrollbar,
  NAlert,
  useMessage
} from 'naive-ui'
import { h } from 'vue'
import {
  GetFuturesProducts,
  GetMainContracts,
  AIChatStream,
  GetConfig
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const message = useMessage()
const loading = ref(false)
const mainContracts = ref([])
const futuresProducts = ref([])

// AI相关
const showAIModal = ref(false)
const aiEnabled = ref(false)
const aiLoading = ref(false)
const aiResponse = ref('')
const aiQuestion = ref('')
const aiMessages = ref([])
const aiScrollbarRef = ref(null)
const eventOffFns = []

// 期货产品按交易所分组
const productsByExchange = computed(() => {
  const exchangeMap = {
    'SHFE': { name: '上海期货交易所', products: [] },
    'DCE': { name: '大连商品交易所', products: [] },
    'CZCE': { name: '郑州商品交易所', products: [] },
    'CFFEX': { name: '中国金融期货交易所', products: [] },
    'INE': { name: '上海国际能源交易中心', products: [] }
  }
  futuresProducts.value.forEach(p => {
    if (exchangeMap[p.exchange]) {
      exchangeMap[p.exchange].products.push(p)
    }
  })
  return Object.entries(exchangeMap).filter(([_, v]) => v.products.length > 0)
})

// 渲染涨跌幅
const renderChange = (val) => {
  if (val === undefined || val === null) return '-'
  const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
  return h('span', { style: { color, fontWeight: 'bold' } }, (val > 0 ? '+' : '') + val.toFixed(2) + '%')
}

// 渲染价格
const renderPrice = (val, digits = 2) => {
  if (val === undefined || val === null) return '-'
  return val.toFixed(digits)
}

// 主力合约列表
const mainContractColumns = [
  { title: '合约', key: 'name', width: 120 },
  { title: '代码', key: 'code', width: 100 },
  { title: '最新价', key: 'price', width: 100, render: (row) => renderPrice(row.price, 2) },
  { title: '涨跌幅', key: 'changePercent', width: 100, render: (row) => renderChange(row.changePercent) },
  { title: '开盘', key: 'open', width: 80, render: (row) => renderPrice(row.open, 2) },
  { title: '最高', key: 'high', width: 80, render: (row) => renderPrice(row.high, 2) },
  { title: '最低', key: 'low', width: 80, render: (row) => renderPrice(row.low, 2) },
  { title: '成交量', key: 'volume', width: 100 },
  { title: '持仓量', key: 'openInterest', width: 100 }
]

// 期货产品列表
const productColumns = [
  { title: '品种', key: 'name', width: 100 },
  { title: '代码', key: 'code', width: 80 },
  { title: '类别', key: 'category', width: 80 },
  { title: '交易单位', key: 'unit', width: 100 },
  { title: '最小变动', key: 'minMove', width: 80 }
]

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const [contracts, products] = await Promise.all([
      GetMainContracts(),
      GetFuturesProducts()
    ])
    mainContracts.value = contracts || []
    futuresProducts.value = products || []
  } catch (e) {
    console.error('加载期货数据失败:', e)
    message.error('加载期货数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    const contracts = await GetMainContracts()
    mainContracts.value = contracts || []
    message.success('刷新成功')
  } catch (e) {
    message.error('刷新失败')
  } finally {
    loading.value = false
  }
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
  startAIAnalysis()
}

// 开始AI分析
const startAIAnalysis = async () => {
  if (aiLoading.value) return

  aiLoading.value = true
  aiResponse.value = ''

  // 添加用户消息
  aiMessages.value.push({
    role: 'user',
    content: '分析当前期货市场的整体状况和投资机会'
  })

  // 添加AI响应占位
  aiMessages.value.push({
    role: 'assistant',
    content: ''
  })

  scrollToBottom()

  try {
    // 构建上下文
    const contracts = mainContracts.value.slice(0, 20).map(c =>
      `${c.name}: ${c.price?.toFixed(2)} (${c.changePercent > 0 ? '+' : ''}${c.changePercent?.toFixed(2)}%)`
    ).join('\n')

    const contextMessage = `请分析当前期货市场状况。

当前主力合约数据:
${contracts}

请从以下几个方面进行分析:
1. 市场整体表现
2. 各板块走势分析（能源、金属、农产品等）
3. 主要品种分析
4. 投资建议和风险提示`

    await AIChatStream({
      message: contextMessage,
      sessionId: 'futures_market',
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
      sessionId: 'futures_market',
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
  loadData()

  eventOffFns.push(EventsOn('ai-chat-stream', handleAIStream))
  eventOffFns.push(EventsOn('ai-chat-done', handleAIDone))
  eventOffFns.push(EventsOn('ai-chat-error', handleAIError))
})

onUnmounted(() => {
  eventOffFns.forEach((off) => typeof off === 'function' && off())
  eventOffFns.length = 0
})
</script>

<template>
  <div class="futures-page">
    <n-spin :show="loading">
      <n-card title="期货市场" :bordered="false">
        <template #header-extra>
          <n-space>
            <n-button type="warning" @click="openAIAnalysis">AI市场分析</n-button>
            <n-button type="primary" @click="refreshData" :loading="loading">
              刷新数据
            </n-button>
          </n-space>
        </template>

        <n-tabs type="line" animated>
          <!-- 主力合约 -->
          <n-tab-pane name="main" tab="主力合约">
            <n-data-table
              :columns="mainContractColumns"
              :data="mainContracts"
              :bordered="false"
              striped
              size="small"
              :max-height="600"
            />
            <div v-if="mainContracts.length === 0" class="empty-tip">暂无主力合约数据</div>
          </n-tab-pane>

          <!-- 期货品种（按交易所分类） -->
          <n-tab-pane name="products" tab="期货品种">
            <n-collapse>
              <n-collapse-item v-for="[exchange, data] in productsByExchange" :key="exchange" :title="data.name" :name="exchange">
                <n-data-table
                  :columns="productColumns"
                  :data="data.products"
                  :bordered="false"
                  striped
                  size="small"
                  :pagination="false"
                />
              </n-collapse-item>
            </n-collapse>
            <div v-if="productsByExchange.length === 0" class="empty-tip">暂无期货品种数据</div>
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-spin>

    <!-- AI分析弹窗 -->
    <n-modal v-model:show="showAIModal" preset="card" title="AI期货市场分析" style="width: 900px;">
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
              placeholder="输入问题继续对话，如：原油期货近期走势如何？"
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
.futures-page {
  height: 100%;
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
