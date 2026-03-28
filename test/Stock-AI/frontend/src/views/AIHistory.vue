<script setup>
import { ref, onMounted, computed } from 'vue'
import {
  NCard,
  NTabs,
  NTabPane,
  NList,
  NListItem,
  NThing,
  NButton,
  NSpace,
  NEmpty,
  NAlert,
  NModal,
  NScrollbar,
  NTag,
  NPopconfirm,
  NDropdown,
  useMessage,
  useDialog
} from 'naive-ui'
import {
  GetAIChatHistory,
  GetAIAnalysisHistory,
  GetAIDataCleanupInfo,
  ExportAIChatHistory,
  ExportAIAnalysisHistory,
  DeleteAIChatSession,
  DeleteAIAnalysisResult,
  ClearOldAIData
} from '../../wailsjs/go/main/App'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const chatSessions = ref([])
const analysisResults = ref([])
const cleanupInfo = ref({})
const showDetailModal = ref(false)
const detailContent = ref('')
const detailTitle = ref('')

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const [chats, analyses, cleanup] = await Promise.all([
      GetAIChatHistory(),
      GetAIAnalysisHistory(),
      GetAIDataCleanupInfo()
    ])
    chatSessions.value = chats || []
    analysisResults.value = analyses || []
    cleanupInfo.value = cleanup || {}
  } catch (e) {
    console.error('加载数据失败:', e)
    message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 查看聊天详情
const viewChatDetail = (session) => {
  detailTitle.value = `聊天记录 - ${session.sessionId}`
  let content = ''
  for (const msg of session.messages || []) {
    const role = msg.role === 'user' ? '用户' : (msg.role === 'assistant' ? 'AI助手' : '系统')
    const time = new Date(msg.createdAt).toLocaleString()
    content += `【${role}】${time}\n${msg.content}\n\n`
  }
  detailContent.value = content
  showDetailModal.value = true
}

// 查看分析详情
const viewAnalysisDetail = (result) => {
  detailTitle.value = `${result.stockName} (${result.stockCode}) 分析报告`
  let content = `分析时间：${new Date(result.createdAt).toLocaleString()}\n`
  if (result.suggestion) {
    content += `分析观点：${result.suggestion}\n`
  }
  content += `\n${result.analysis}`
  content += `\n\n---\n免责声明：以上分析由AI生成，仅供学习研究参考，不构成任何投资建议。`
  detailContent.value = content
  showDetailModal.value = true
}

// 导出下拉菜单选项
const exportOptions = [
  { label: '导出为 TXT', key: 'txt' },
  { label: '导出为 Markdown', key: 'md' }
]

// 导出聊天记录
const handleExportChat = async (key) => {
  try {
    const content = await ExportAIChatHistory('', key)
    downloadFile(content, `AI聊天记录_${formatDate(new Date())}.${key}`)
    message.success('导出成功')
  } catch (e) {
    message.error('导出失败: ' + e)
  }
}

// 导出分析记录
const handleExportAnalysis = async (key) => {
  try {
    const content = await ExportAIAnalysisHistory(key)
    downloadFile(content, `AI分析记录_${formatDate(new Date())}.${key}`)
    message.success('导出成功')
  } catch (e) {
    message.error('导出失败: ' + e)
  }
}

// 下载文件
const downloadFile = (content, filename) => {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// 格式化日期
const formatDate = (date) => {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}${m}${d}`
}

// 删除聊天会话
const handleDeleteChat = async (sessionId) => {
  try {
    await DeleteAIChatSession(sessionId)
    message.success('删除成功')
    loadData()
  } catch (e) {
    message.error('删除失败: ' + e)
  }
}

// 删除分析结果
const handleDeleteAnalysis = async (id) => {
  try {
    await DeleteAIAnalysisResult(id)
    message.success('删除成功')
    loadData()
  } catch (e) {
    message.error('删除失败: ' + e)
  }
}

// 手动清理过期数据
const handleClearOldData = () => {
  dialog.warning({
    title: '确认清理',
    content: '将清理30天前的聊天记录和7天前的分析结果，此操作不可恢复。确定要继续吗？',
    positiveText: '确定清理',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const result = await ClearOldAIData()
        message.success(`清理完成：删除了 ${result[0]} 条聊天记录和 ${result[1]} 条分析结果`)
        loadData()
      } catch (e) {
        message.error('清理失败: ' + e)
      }
    }
  })
}

// 计算清理提示信息
const cleanupWarning = computed(() => {
  const info = cleanupInfo.value
  if (!info.chatCount && !info.analysisCount) {
    return null
  }
  return {
    chat: `聊天记录保留 ${info.chatRetentionDays || 30} 天，当前共 ${info.chatCount || 0} 条`,
    analysis: `分析结果保留 ${info.analysisRetentionDays || 7} 天，当前共 ${info.analysisCount || 0} 条`
  }
})

onMounted(() => {
  loadData()
})

// 获取建议类型对应的标签颜色
const getSuggestionType = (suggestion) => {
  if (!suggestion) return 'default'
  const s = suggestion.toLowerCase()
  if (s.includes('买') || s.includes('buy')) return 'success'
  if (s.includes('卖') || s.includes('sell')) return 'error'
  if (s.includes('持') || s.includes('hold')) return 'warning'
  return 'info'
}
</script>

<template>
  <div class="ai-history-page">
    <n-card title="AI 历史记录" :bordered="false">
      <!-- 数据清理警告 -->
      <n-alert type="warning" style="margin-bottom: 16px;" v-if="cleanupWarning">
        <template #header>数据自动清理提示</template>
        <p>{{ cleanupWarning.chat }}</p>
        <p>{{ cleanupWarning.analysis }}</p>
        <p style="color: #f0a020; margin-top: 8px;">如需长期保存，请尽快下载导出！</p>
      </n-alert>

      <n-tabs type="line" animated>
        <!-- 聊天记录 Tab -->
        <n-tab-pane name="chat" tab="聊天记录">
          <n-space justify="end" style="margin-bottom: 16px;">
            <n-dropdown :options="exportOptions" @select="handleExportChat">
              <n-button type="primary" :disabled="chatSessions.length === 0">
                导出聊天记录
              </n-button>
            </n-dropdown>
            <n-button @click="handleClearOldData">清理过期数据</n-button>
          </n-space>

          <n-empty v-if="chatSessions.length === 0" description="暂无聊天记录" />

          <n-list v-else hoverable clickable>
            <n-list-item v-for="session in chatSessions" :key="session.sessionId">
              <n-thing>
                <template #header>
                  <span style="cursor: pointer;" @click="viewChatDetail(session)">
                    {{ session.lastMessage || '(无内容)' }}
                  </span>
                </template>
                <template #header-extra>
                  <n-space>
                    <n-tag size="small" type="info">{{ session.messageCount }} 条消息</n-tag>
                    <n-popconfirm @positive-click="handleDeleteChat(session.sessionId)">
                      <template #trigger>
                        <n-button size="small" type="error" quaternary>删除</n-button>
                      </template>
                      确定删除这个会话吗？
                    </n-popconfirm>
                  </n-space>
                </template>
                <template #description>
                  <span style="color: #999; font-size: 12px;">
                    会话ID: {{ session.sessionId }} | 创建时间: {{ session.createdAt }}
                  </span>
                </template>
              </n-thing>
            </n-list-item>
          </n-list>
        </n-tab-pane>

        <!-- 分析记录 Tab -->
        <n-tab-pane name="analysis" tab="分析记录">
          <n-space justify="end" style="margin-bottom: 16px;">
            <n-dropdown :options="exportOptions" @select="handleExportAnalysis">
              <n-button type="primary" :disabled="analysisResults.length === 0">
                导出分析记录
              </n-button>
            </n-dropdown>
            <n-button @click="handleClearOldData">清理过期数据</n-button>
          </n-space>

          <n-empty v-if="analysisResults.length === 0" description="暂无分析记录" />

          <n-list v-else hoverable clickable>
            <n-list-item v-for="result in analysisResults" :key="result.id">
              <n-thing>
                <template #header>
                  <span style="cursor: pointer;" @click="viewAnalysisDetail(result)">
                    {{ result.stockName }} ({{ result.stockCode }})
                  </span>
                </template>
                <template #header-extra>
                  <n-space>
                    <n-tag v-if="result.suggestion" size="small" :type="getSuggestionType(result.suggestion)">
                      {{ result.suggestion }}
                    </n-tag>
                    <n-popconfirm @positive-click="handleDeleteAnalysis(result.id)">
                      <template #trigger>
                        <n-button size="small" type="error" quaternary>删除</n-button>
                      </template>
                      确定删除这条分析记录吗？
                    </n-popconfirm>
                  </n-space>
                </template>
                <template #description>
                  <span style="color: #999; font-size: 12px;">
                    分析时间: {{ new Date(result.createdAt).toLocaleString() }}
                  </span>
                </template>
              </n-thing>
            </n-list-item>
          </n-list>
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 详情弹窗 -->
    <n-modal v-model:show="showDetailModal" preset="card" :title="detailTitle" style="width: 800px; max-width: 90vw;">
      <n-scrollbar style="max-height: 60vh;">
        <pre style="white-space: pre-wrap; word-wrap: break-word; font-family: inherit; margin: 0;">{{ detailContent }}</pre>
      </n-scrollbar>
    </n-modal>
  </div>
</template>

<style scoped>
.ai-history-page {
  max-width: 1000px;
}

pre {
  line-height: 1.6;
}
</style>
