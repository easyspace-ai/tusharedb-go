<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import {
  NCard,
  NInput,
  NButton,
  NSpace,
  NSpin,
  NAlert,
  NTag,
  NSelect,
  NDivider,
  NScrollbar,
  useMessage
} from 'naive-ui'
import {
  AIChatStream,
  AIAnalyzeStockStream,
  AIRecommendStream,
  GetStockList,
  GetConfig,
  ListPrompts,
  GetActivePersona,
  SetActivePersona
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const message = useMessage()
const inputMessage = ref('')
const messages = ref([])
const loading = ref(false)
const aiEnabled = ref(false)
const currentResponse = ref('')
const scrollbarRef = ref(null)
const stocks = ref([])
const selectedStock = ref(null)
const eventOffFns = []

// 人设相关
const personaPrompts = ref([])
const selectedPersona = ref(null)
const activePersonaContent = ref('')

// 快捷功能
const quickActions = [
  { label: '分析股票', value: 'analyze', desc: '对选中的股票进行AI分析' },
  { label: '市场分析', value: 'recommend', desc: '根据市场数据分析热点方向（仅供参考）' },
]

// 检查AI配置
const checkAIConfig = async () => {
  try {
    const config = await GetConfig()
    aiEnabled.value = config.aiEnabled && config.aiApiKey
  } catch (e) {
    console.error('获取配置失败:', e)
  }
}

// 加载股票列表
const loadStocks = async () => {
  try {
    const list = await GetStockList()
    stocks.value = (list || []).map(s => ({
      label: `${s.name} (${s.code})`,
      value: s.code
    }))
  } catch (e) {
    console.error('加载股票列表失败:', e)
  }
}

// 加载人设列表
const loadPersonas = async () => {
  try {
    const data = await ListPrompts('persona')
    personaPrompts.value = data || []

    // 获取当前激活的人设
    const activePersona = await GetActivePersona()
    if (activePersona) {
      activePersonaContent.value = activePersona
      // 找到对应的人设名称
      const found = personaPrompts.value.find(p => p.content === activePersona)
      if (found) {
        selectedPersona.value = found.name
      }
    }
  } catch (e) {
    console.error('加载人设失败:', e)
  }
}

// 切换人设
const onPersonaChange = async (personaName) => {
  try {
    await SetActivePersona(personaName || '')
    if (personaName) {
      const found = personaPrompts.value.find(p => p.name === personaName)
      activePersonaContent.value = found ? found.content : ''
      message.success(`已切换到人设: ${personaName}`)
    } else {
      activePersonaContent.value = ''
      message.info('已清除人设')
    }
  } catch (e) {
    message.error('切换人设失败: ' + e)
  }
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (scrollbarRef.value) {
      scrollbarRef.value.scrollTo({ top: 999999, behavior: 'smooth' })
    }
  })
}

// 发送消息
const sendMessage = async () => {
  if (!inputMessage.value.trim() || loading.value) return

  const userMessage = inputMessage.value.trim()
  inputMessage.value = ''

  // 添加用户消息
  messages.value.push({
    role: 'user',
    content: userMessage
  })

  // 添加AI响应占位
  messages.value.push({
    role: 'assistant',
    content: ''
  })

  loading.value = true
  currentResponse.value = ''
  scrollToBottom()

  try {
    await AIChatStream({
      message: userMessage,
      sessionId: 'default',
      stockCode: selectedStock.value || ''
    })
  } catch (e) {
    message.error('发送失败: ' + e)
    loading.value = false
  }
}

// 分析股票
const analyzeStock = async () => {
  if (!selectedStock.value) {
    message.warning('请先选择要分析的股票')
    return
  }

  if (loading.value) return

  // 添加用户消息
  messages.value.push({
    role: 'user',
    content: `分析股票 ${selectedStock.value}`
  })

  // 添加AI响应占位
  messages.value.push({
    role: 'assistant',
    content: ''
  })

  loading.value = true
  currentResponse.value = ''
  scrollToBottom()

  try {
    await AIAnalyzeStockStream(selectedStock.value)
  } catch (e) {
    message.error('分析失败: ' + e)
    loading.value = false
  }
}

// 市场分析
const getRecommend = async () => {
  if (loading.value) return

  // 添加用户消息
  messages.value.push({
    role: 'user',
    content: '请根据当前市场数据分析热点方向'
  })

  // 添加AI响应占位
  messages.value.push({
    role: 'assistant',
    content: ''
  })

  loading.value = true
  currentResponse.value = ''
  scrollToBottom()

  try {
    await AIRecommendStream()
  } catch (e) {
    message.error('获取分析失败: ' + e)
    loading.value = false
  }
}

// 清空对话
const clearMessages = () => {
  messages.value = []
  currentResponse.value = ''
}

// 处理流式响应
const handleStreamResponse = (content) => {
  currentResponse.value += content
  if (messages.value.length > 0) {
    messages.value[messages.value.length - 1].content = currentResponse.value
  }
  scrollToBottom()
}

// 处理响应完成
const handleStreamDone = () => {
  loading.value = false
  currentResponse.value = ''
}

// 处理错误
const handleStreamError = (error) => {
  loading.value = false
  message.error(error)
  if (messages.value.length > 0 && messages.value[messages.value.length - 1].role === 'assistant') {
    messages.value[messages.value.length - 1].content = `错误: ${error}`
  }
}

onMounted(async () => {
  await checkAIConfig()
  await loadStocks()
  await loadPersonas()

  eventOffFns.push(EventsOn('ai-chat-stream', handleStreamResponse))
  eventOffFns.push(EventsOn('ai-chat-done', handleStreamDone))
  eventOffFns.push(EventsOn('ai-chat-error', handleStreamError))
})

onUnmounted(() => {
  eventOffFns.forEach((off) => typeof off === 'function' && off())
  eventOffFns.length = 0
})
</script>

<template>
  <div class="ai-page">
    <n-card title="AI 智能助手" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-tag :type="aiEnabled ? 'success' : 'warning'">
            {{ aiEnabled ? 'AI已启用' : 'AI未配置' }}
          </n-tag>
          <n-button size="small" @click="clearMessages">清空对话</n-button>
        </n-space>
      </template>

      <!-- AI未配置提示 -->
      <n-alert v-if="!aiEnabled" type="warning" style="margin-bottom: 16px;">
        AI功能未启用或未配置API Key，请前往「设置」页面配置AI服务。
        <br />
        支持的AI服务：DeepSeek、通义千问、智谱GLM、文心一言、硅基流动、Ollama等
      </n-alert>

      <!-- 快捷功能 -->
      <div class="quick-actions">
        <n-space>
          <n-select
            v-model:value="selectedStock"
            :options="stocks"
            placeholder="选择股票"
            filterable
            clearable
            style="width: 200px;"
          />
          <n-button type="primary" :disabled="!aiEnabled || loading || !selectedStock" @click="analyzeStock">
            分析股票
          </n-button>
          <n-button type="info" :disabled="!aiEnabled || loading" @click="getRecommend">
            市场分析
          </n-button>
          <n-divider vertical />
          <n-select
            v-model:value="selectedPersona"
            :options="personaPrompts.map(p => ({ label: p.name, value: p.name }))"
            placeholder="选择AI人设"
            clearable
            style="width: 150px;"
            @update:value="onPersonaChange"
          />
          <n-tag v-if="selectedPersona" type="success" size="small">
            人设: {{ selectedPersona }}
          </n-tag>
        </n-space>
      </div>

      <n-divider />

      <!-- 对话区域 -->
      <div class="chat-container">
        <n-scrollbar ref="scrollbarRef" style="max-height: calc(100vh - 380px);">
          <div class="messages">
            <div v-if="messages.length === 0" class="empty-tip">
              <p>欢迎使用AI智能助手！</p>
              <p>你可以：</p>
              <ul>
                <li>选择股票后点击「分析股票」获取AI分析报告</li>
                <li>点击「市场分析」获取市场热点方向参考</li>
                <li>直接输入问题与AI对话</li>
              </ul>
              <p style="color: #f0a020; margin-top: 12px; font-size: 12px;">
                <strong>免责声明：</strong>AI分析结果仅供参考，不构成任何投资建议。投资有风险，入市需谨慎。
              </p>
            </div>

            <div
              v-for="(msg, index) in messages"
              :key="index"
              :class="['message', msg.role]"
            >
              <div class="message-role">
                {{ msg.role === 'user' ? '我' : 'AI' }}
              </div>
              <div class="message-content">
                <n-spin v-if="msg.role === 'assistant' && loading && index === messages.length - 1 && !msg.content" size="small" />
                <div v-else class="markdown-content" v-html="formatContent(msg.content)"></div>
              </div>
            </div>
          </div>
        </n-scrollbar>
      </div>

      <!-- 输入区域 -->
      <div class="input-area">
        <n-input
          v-model:value="inputMessage"
          type="textarea"
          placeholder="输入问题，按Enter发送..."
          :autosize="{ minRows: 2, maxRows: 4 }"
          :disabled="!aiEnabled || loading"
          @keyup.enter.exact="sendMessage"
        />
        <n-button
          type="primary"
          :loading="loading"
          :disabled="!aiEnabled || !inputMessage.trim()"
          style="margin-top: 8px;"
          @click="sendMessage"
        >
          发送
        </n-button>
      </div>
    </n-card>
  </div>
</template>

<script>
// 格式化内容（简单的Markdown转换）
function formatContent(content) {
  if (!content) return ''

  // 转义HTML
  let html = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

  // 处理代码块
  html = html.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')

  // 处理行内代码
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>')

  // 处理粗体
  html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')

  // 处理标题
  html = html.replace(/^### (.+)$/gm, '<h4>$1</h4>')
  html = html.replace(/^## (.+)$/gm, '<h3>$1</h3>')
  html = html.replace(/^# (.+)$/gm, '<h2>$1</h2>')

  // 处理列表
  html = html.replace(/^- (.+)$/gm, '<li>$1</li>')
  html = html.replace(/^(\d+)\. (.+)$/gm, '<li>$2</li>')

  // 处理换行
  html = html.replace(/\n/g, '<br>')

  return html
}

export default {
  methods: {
    formatContent
  }
}
</script>

<style scoped>
.ai-page {
  height: 100%;
}

.quick-actions {
  margin-bottom: 8px;
}

.chat-container {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  padding: 16px;
  min-height: 300px;
}

.messages {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.empty-tip {
  text-align: center;
  color: #999;
  padding: 40px;
}

.empty-tip ul {
  text-align: left;
  display: inline-block;
  margin-top: 16px;
}

.empty-tip li {
  margin: 8px 0;
}

.message {
  display: flex;
  gap: 12px;
}

.message.user {
  flex-direction: row-reverse;
}

.message-role {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  flex-shrink: 0;
}

.message.user .message-role {
  background: #18a058;
  color: white;
}

.message.assistant .message-role {
  background: #2080f0;
  color: white;
}

.message-content {
  max-width: 80%;
  padding: 12px 16px;
  border-radius: 8px;
  line-height: 1.6;
}

.message.user .message-content {
  background: #18a058;
  color: white;
  border-bottom-right-radius: 0;
}

.message.assistant .message-content {
  background: rgba(255, 255, 255, 0.1);
  border-bottom-left-radius: 0;
}

.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4) {
  margin: 12px 0 8px 0;
}

.markdown-content :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 8px 0;
}

.markdown-content :deep(code) {
  background: rgba(0, 0, 0, 0.2);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

.markdown-content :deep(pre code) {
  background: none;
  padding: 0;
}

.markdown-content :deep(li) {
  margin: 4px 0;
  margin-left: 20px;
}

.markdown-content :deep(strong) {
  color: #18a058;
}

.input-area {
  margin-top: 16px;
}
</style>
