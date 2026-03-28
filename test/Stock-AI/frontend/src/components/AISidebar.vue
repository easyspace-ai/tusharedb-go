<script setup>
import { ref, onMounted, onUnmounted, nextTick, defineProps, watch } from 'vue'
import {
  NCard,
  NInput,
  NButton,
  NSpace,
  NSpin,
  NTag,
  NScrollbar,
  NEmpty,
  useMessage
} from 'naive-ui'
import {
  AIChatStream,
  AIAnalyzeStockStream,
  GetConfig
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const props = defineProps({
  // 当前选中的股票代码
  stockCode: {
    type: String,
    default: ''
  },
  // 当前选中的股票名称
  stockName: {
    type: String,
    default: ''
  }
})

const message = useMessage()
const inputMessage = ref('')
const messages = ref([])
const loading = ref(false)
const aiEnabled = ref(false)
const currentResponse = ref('')
const scrollbarRef = ref(null)
const eventOffFns = []

// 检查AI配置
const checkAIConfig = async () => {
  try {
    const config = await GetConfig()
    aiEnabled.value = config.aiEnabled && config.aiApiKey
  } catch (e) {
    console.error('获取配置失败:', e)
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
      sessionId: 'sidebar',
      stockCode: props.stockCode || ''
    })
  } catch (e) {
    message.error('发送失败: ' + e)
    loading.value = false
  }
}

// 分析当前股票
const analyzeCurrentStock = async () => {
  if (!props.stockCode) {
    message.warning('请先选择要分析的股票')
    return
  }

  if (loading.value) return

  // 添加用户消息
  messages.value.push({
    role: 'user',
    content: `分析股票 ${props.stockName || props.stockCode}`
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
    await AIAnalyzeStockStream(props.stockCode)
  } catch (e) {
    message.error('分析失败: ' + e)
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

// 格式化内容（简单的Markdown转换）
const formatContent = (content) => {
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

onMounted(async () => {
  await checkAIConfig()

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
  <div class="ai-sidebar">
    <div class="ai-sidebar-header">
      <span class="title">AI 助手</span>
      <n-space>
        <n-tag :type="aiEnabled ? 'success' : 'warning'" size="small">
          {{ aiEnabled ? '已启用' : '未配置' }}
        </n-tag>
        <n-button size="tiny" quaternary @click="clearMessages">清空</n-button>
      </n-space>
    </div>

    <!-- 快捷操作 -->
    <div class="quick-actions" v-if="stockCode">
      <n-button
        size="small"
        type="primary"
        block
        :disabled="!aiEnabled || loading"
        @click="analyzeCurrentStock"
      >
        分析 {{ stockName || stockCode }}
      </n-button>
    </div>

    <!-- 对话区域 -->
    <div class="chat-container">
      <n-scrollbar ref="scrollbarRef" style="height: 100%;">
        <div class="messages">
          <n-empty v-if="messages.length === 0" description="输入问题开始对话" size="small" />

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
        placeholder="输入问题..."
        :autosize="{ minRows: 2, maxRows: 3 }"
        :disabled="!aiEnabled || loading"
        @keyup.enter.exact="sendMessage"
      />
      <n-button
        type="primary"
        size="small"
        :loading="loading"
        :disabled="!aiEnabled || !inputMessage.trim()"
        style="margin-top: 8px;"
        block
        @click="sendMessage"
      >
        发送
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.ai-sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: rgba(0, 0, 0, 0.2);
  border-left: 1px solid rgba(255, 255, 255, 0.09);
}

.ai-sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.09);
}

.ai-sidebar-header .title {
  font-weight: 500;
  font-size: 14px;
}

.quick-actions {
  padding: 8px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.09);
}

.chat-container {
  flex: 1;
  overflow: hidden;
  padding: 8px;
}

.messages {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 4px;
}

.message {
  display: flex;
  gap: 8px;
}

.message.user {
  flex-direction: row-reverse;
}

.message-role {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
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
  max-width: 85%;
  padding: 8px 10px;
  border-radius: 6px;
  line-height: 1.5;
  font-size: 13px;
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
  margin: 8px 0 4px 0;
  font-size: 13px;
}

.markdown-content :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  padding: 8px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 4px 0;
  font-size: 12px;
}

.markdown-content :deep(code) {
  background: rgba(0, 0, 0, 0.2);
  padding: 1px 4px;
  border-radius: 2px;
  font-family: monospace;
  font-size: 12px;
}

.markdown-content :deep(pre code) {
  background: none;
  padding: 0;
}

.markdown-content :deep(li) {
  margin: 2px 0;
  margin-left: 16px;
}

.markdown-content :deep(strong) {
  color: #18a058;
}

.input-area {
  padding: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.09);
}
</style>
