<script setup>
import { ref, onMounted, onUnmounted, nextTick, h } from 'vue'
import {
  NCard,
  NDataTable,
  NButton,
  NInput,
  NInputNumber,
  NSpace,
  NModal,
  NForm,
  NFormItem,
  NTag,
  NSpin,
  NAlert,
  NList,
  NListItem,
  NThing,
  NSelect,
  NRadioGroup,
  NRadio,
  NEmpty,
  NScrollbar,
  useMessage
} from 'naive-ui'
import * as echarts from 'echarts'
import {
  GetFundList,
  AddFund,
  RemoveFund,
  GetFundPrice,
  GetFundOverview,
  GetFundPosition,
  AddFundPosition,
  UpdateFundPosition,
  DeleteFundPosition,
  GetFundAlerts,
  AddFundAlert,
  DeleteFundAlert,
  ToggleFundAlert,
  ResetFundAlert,
  CheckFundAlerts,
  AIAnalyzeFundStream,
  AIChatStream,
  AISummarizeContentStream,
  OpenURL
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const message = useMessage()
const funds = ref([])
const loading = ref(false)
const priceRefreshing = ref(false)
const showAddModal = ref(false)
const newFundCode = ref('')

const selectedFund = ref(null)
const fundOverview = ref(null)
const showDetailModal = ref(false)
const detailLoading = ref(false)
const historyChartRef = ref(null)
let historyChart = null

const showPositionModal = ref(false)
const positionLoading = ref(false)
const currentPosition = ref(null)
const positionForm = ref({
  fundCode: '',
  fundName: '',
  buyNav: null,
  buyDate: '',
  share: null,
  costNav: null,
  targetNav: null,
  stopLossNav: null,
  notes: ''
})

const showAlertModal = ref(false)
const fundAlerts = ref([])
const alertLoading = ref(false)
const alertForm = ref({
  alertType: 'nav',
  condition: 'above',
  targetValue: null
})

const showAIModal = ref(false)
const aiMessages = ref([])
const aiLoading = ref(false)
const aiResponse = ref('')
const aiQuestion = ref('')
const aiScrollbarRef = ref(null)

const showSummaryModal = ref(false)
const summaryTitle = ref('')
const summaryContent = ref('')
const summaryLoading = ref(false)
const summaryUrl = ref('')
const summaryType = ref('notice')
const summaryInfoCode = ref('')
const summaryArtCode = ref('')
const summaryScrollbarRef = ref(null)
const manualContent = ref('')
const showManualInput = ref(false)

const eventOffFns = []
let alertTimer = null

const columns = [
  { title: '代码', key: 'code', width: 100 },
  { title: '名称', key: 'name', width: 180 },
  { title: '类型', key: 'type', width: 160 },
  { title: '净值', key: 'nav', width: 80, render: row => (row.nav ? row.nav.toFixed(4) : '-') },
  { title: '估值', key: 'estimate', width: 80, render: row => (row.estimate ? row.estimate.toFixed(4) : '-') },
  {
    title: '估算涨跌',
    key: 'changePercent',
    width: 100,
    render: row => {
      if (row.changePercent === undefined || row.changePercent === null) return '-'
      const val = row.changePercent
      const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
      return h('span', { style: { color, fontWeight: 'bold' } }, (val > 0 ? '+' : '') + val.toFixed(2) + '%')
    }
  },
  { title: '更新时间', key: 'updateTime', width: 150 },
  {
    title: '操作',
    key: 'actions',
    width: 360,
    render: row =>
      h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, { size: 'small', type: 'info', onClick: () => openDetail(row) }, { default: () => '详情' }),
          h(NButton, { size: 'small', type: 'tertiary', onClick: () => openPositionModal(row) }, { default: () => '持仓' }),
          h(NButton, { size: 'small', type: 'primary', onClick: () => openAlertModal(row) }, { default: () => '提醒' }),
          h(NButton, { size: 'small', type: 'warning', onClick: () => openFundAIModal(row) }, { default: () => 'AI分析' }),
          h(NButton, { size: 'small', type: 'error', onClick: () => handleRemove(row.code) }, { default: () => '删除' })
        ]
      })
  }
]

const loadFunds = async () => {
  loading.value = true
  try {
    const list = await GetFundList()
    funds.value = list || []
    if (funds.value.length) {
      await refreshPrices()
    }
  } catch (e) {
    message.error(`加载基金列表失败: ${e}`)
  } finally {
    loading.value = false
  }
}

const refreshPrices = async () => {
  if (!funds.value.length) return
  priceRefreshing.value = true
  try {
    const codes = funds.value.map(f => f.code)
    const prices = await GetFundPrice(codes)
    if (prices) {
      funds.value = funds.value.map(f => ({
        ...f,
        ...(prices[f.code] || {})
      }))
    }
  } catch (e) {
    message.error(`刷新估值失败: ${e}`)
  } finally {
    priceRefreshing.value = false
  }
}

const handleAddFund = async () => {
  if (!newFundCode.value.trim()) {
    message.warning('请输入基金代码')
    return
  }
  try {
    await AddFund(newFundCode.value.trim())
    message.success('添加成功')
    showAddModal.value = false
    newFundCode.value = ''
    await loadFunds()
  } catch (e) {
    message.error(`添加失败: ${e}`)
  }
}

const handleRemove = async (code) => {
  try {
    await RemoveFund(code)
    message.success('已删除')
    await loadFunds()
  } catch (e) {
    message.error(`删除失败: ${e}`)
  }
}

const openDetail = async (fund) => {
  selectedFund.value = fund
  showDetailModal.value = true
  detailLoading.value = true
  try {
    const overview = await GetFundOverview(fund.code)
    fundOverview.value = overview
    await nextTick()
    renderHistoryChart()
  } catch (e) {
    message.error(`获取基金详情失败: ${e}`)
  } finally {
    detailLoading.value = false
  }
}

const renderHistoryChart = () => {
  if (!historyChartRef.value || !fundOverview.value?.history?.length) return
  if (!historyChart) {
    historyChart = echarts.init(historyChartRef.value)
  }
  const dates = fundOverview.value.history.map(item => item.date)
  const navs = fundOverview.value.history.map(item => item.nav)
  historyChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: dates },
    yAxis: { type: 'value' },
    grid: { left: 40, right: 20, top: 20, bottom: 40 },
    series: [{
      name: '净值',
      type: 'line',
      smooth: true,
      areaStyle: {},
      data: navs
    }]
  })
}

const openPositionModal = async (fund) => {
  selectedFund.value = fund
  showPositionModal.value = true
  positionLoading.value = true
  positionForm.value = {
    fundCode: fund.code,
    fundName: fund.name,
    buyNav: null,
    buyDate: new Date().toISOString().split('T')[0],
    share: null,
    costNav: null,
    targetNav: null,
    stopLossNav: null,
    notes: ''
  }

  try {
    const position = await GetFundPosition(fund.code)
    if (position && position.id) {
      currentPosition.value = position
      positionForm.value = {
        fundCode: position.fundCode,
        fundName: position.fundName,
        buyNav: position.buyNav,
        buyDate: position.buyDate,
        share: position.share,
        costNav: position.costNav,
        targetNav: position.targetNav,
        stopLossNav: position.stopLossNav,
        notes: position.notes
      }
    } else {
      currentPosition.value = null
    }
  } catch (e) {
    currentPosition.value = null
  } finally {
    positionLoading.value = false
  }
}

const saveFundPosition = async () => {
  if (!positionForm.value.buyNav || !positionForm.value.share) {
    message.warning('请填写买入净值和份额')
    return
  }
  try {
    if (currentPosition.value && currentPosition.value.id) {
      await UpdateFundPosition({ ...currentPosition.value, ...positionForm.value })
      message.success('持仓已更新')
    } else {
      await AddFundPosition({ ...positionForm.value })
      message.success('持仓已保存')
    }
    showPositionModal.value = false
    currentPosition.value = null
  } catch (e) {
    message.error(`保存失败: ${e}`)
  }
}

const deleteFundPosition = async () => {
  if (!currentPosition.value?.id) return
  try {
    await DeleteFundPosition(currentPosition.value.id)
    message.success('已删除持仓')
    showPositionModal.value = false
    currentPosition.value = null
  } catch (e) {
    message.error(`删除失败: ${e}`)
  }
}

const openAlertModal = async (fund) => {
  selectedFund.value = fund
  showAlertModal.value = true
  alertLoading.value = true
  alertForm.value = {
    alertType: 'nav',
    condition: 'above',
    targetValue: fund.nav || fund.estimate || null
  }
  try {
    const alerts = await GetFundAlerts(fund.code)
    fundAlerts.value = alerts || []
  } catch (e) {
    message.error(`获取提醒失败: ${e}`)
  } finally {
    alertLoading.value = false
  }
}

const addFundAlert = async () => {
  if (!alertForm.value.targetValue) {
    message.warning('请填写目标值')
    return
  }
  try {
    await AddFundAlert({
      fundCode: selectedFund.value.code,
      fundName: selectedFund.value.name,
      alertType: alertForm.value.alertType,
      condition: alertForm.value.condition,
      targetValue: alertForm.value.targetValue
    })
    message.success('提醒已添加')
    const alerts = await GetFundAlerts(selectedFund.value.code)
    fundAlerts.value = alerts || []
  } catch (e) {
    message.error(`添加提醒失败: ${e}`)
  }
}

const removeFundAlert = async (id) => {
  try {
    await DeleteFundAlert(id)
    const alerts = await GetFundAlerts(selectedFund.value.code)
    fundAlerts.value = alerts || []
  } catch (e) {
    message.error(`删除提醒失败: ${e}`)
  }
}

const toggleFundAlert = async (alert) => {
  try {
    await ToggleFundAlert(alert.id, !alert.enabled)
    alert.enabled = !alert.enabled
  } catch (e) {
    message.error(`更新提醒失败: ${e}`)
  }
}

const resetFundAlert = async (alert) => {
  try {
    await ResetFundAlert(alert.id)
    const alerts = await GetFundAlerts(selectedFund.value.code)
    fundAlerts.value = alerts || []
  } catch (e) {
    message.error(`重置提醒失败: ${e}`)
  }
}

const openFundAIModal = (fund) => {
  selectedFund.value = fund
  aiMessages.value = []
  aiResponse.value = ''
  showAIModal.value = true
  startFundAIAnalysis()
}

const startFundAIAnalysis = async () => {
  if (!selectedFund.value || aiLoading.value) return
  aiLoading.value = true
  aiMessages.value.push({
    role: 'assistant',
    content: '正在分析基金数据，请稍候...'
  })
  try {
    await AIAnalyzeFundStream(selectedFund.value.code)
  } catch (e) {
    aiLoading.value = false
    message.error(`AI分析失败: ${e}`)
  }
}

const sendAIQuestion = async () => {
  if (!aiQuestion.value.trim() || aiLoading.value || !selectedFund.value) return
  const question = aiQuestion.value.trim()
  aiQuestion.value = ''
  aiLoading.value = true
  aiMessages.value.push({ role: 'user', content: question })
  aiMessages.value.push({ role: 'assistant', content: '' })
  scrollAIToBottom()
  try {
    await AIChatStream({
      message: question,
      sessionId: 'fund-' + selectedFund.value.code,
      fundCode: selectedFund.value.code
    })
  } catch (e) {
    aiLoading.value = false
    message.error(`发送失败: ${e}`)
  }
}

const handleAIStream = (content) => {
  if (!aiMessages.value.length) return
  const last = aiMessages.value[aiMessages.value.length - 1]
  if (last.role !== 'assistant') return
  last.content = (last.content || '') + content
  scrollAIToBottom()
}

const handleAIDone = () => {
  aiLoading.value = false
}

const handleAIError = (err) => {
  aiLoading.value = false
  message.error(err)
}

const scrollAIToBottom = () => {
  nextTick(() => {
    if (aiScrollbarRef.value) {
      aiScrollbarRef.value.scrollTo({ top: 999999, behavior: 'smooth' })
    }
  })
}

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

const openSummaryModal = (notice, type = 'notice') => {
  if (!selectedFund.value) return
  summaryTitle.value = notice.title
  summaryType.value = type
  summaryUrl.value = notice.url || ''
  summaryInfoCode.value = notice.infoCode || ''
  summaryArtCode.value = notice.artCode || notice.id || ''
  summaryContent.value = ''
  summaryLoading.value = true
  showSummaryModal.value = true
  manualContent.value = ''
  showManualInput.value = false
  AISummarizeContentStream(
    notice.title,
    summaryType.value,
    summaryUrl.value,
    summaryInfoCode.value,
    summaryArtCode.value,
    selectedFund.value?.code || '',
    ''
  )
    .catch(e => {
      summaryLoading.value = false
      message.error(`AI解读失败: ${e}`)
    })
}

const analyzeManualSummary = () => {
  if (!manualContent.value.trim()) {
    message.warning('请先粘贴内容')
    return
  }
  summaryContent.value = ''
  summaryLoading.value = true
  showManualInput.value = false
  AISummarizeContentStream(
    summaryTitle.value,
    summaryType.value,
    '',
    summaryInfoCode.value,
    summaryArtCode.value,
    selectedFund.value?.code || '',
    manualContent.value.trim()
  )
    .catch(e => {
      summaryLoading.value = false
      message.error(`AI解读失败: ${e}`)
    })
}

const handleSummaryStream = (content) => {
  summaryContent.value += content
  nextTick(() => {
    if (summaryScrollbarRef.value) {
      summaryScrollbarRef.value.scrollTo({ top: 999999, behavior: 'smooth' })
    }
  })
}

const handleSummaryDone = () => {
  summaryLoading.value = false
}

const handleSummaryError = (err) => {
  summaryLoading.value = false
  summaryContent.value = `错误：${err}`
}

const startAlertTimer = () => {
  stopAlertTimer()
  alertTimer = setInterval(async () => {
    try {
      await CheckFundAlerts()
    } catch (e) {
      // ignore
    }
  }, 5000)
}

const stopAlertTimer = () => {
  if (alertTimer) {
    clearInterval(alertTimer)
    alertTimer = null
  }
}

const goNoticeDetail = (url) => {
  if (url) {
    OpenURL(url)
  }
}

onMounted(async () => {
  await loadFunds()
  eventOffFns.push(EventsOn('ai-chat-stream', handleAIStream))
  eventOffFns.push(EventsOn('ai-chat-done', handleAIDone))
  eventOffFns.push(EventsOn('ai-chat-error', handleAIError))
  eventOffFns.push(EventsOn('ai-summary-stream', handleSummaryStream))
  eventOffFns.push(EventsOn('ai-summary-done', handleSummaryDone))
  eventOffFns.push(EventsOn('ai-summary-error', handleSummaryError))
  startAlertTimer()
})

onUnmounted(() => {
  eventOffFns.forEach(off => typeof off === 'function' && off())
  eventOffFns.length = 0
  stopAlertTimer()
  if (historyChart) {
    historyChart.dispose()
    historyChart = null
  }
})
</script>

<template>
  <div class="fund-page">
    <n-card title="自选基金" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-button type="primary" @click="showAddModal = true">添加基金</n-button>
          <n-button :loading="priceRefreshing" @click="refreshPrices">刷新估值</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="funds"
        :loading="loading"
        :bordered="false"
        :single-line="false"
        striped
      />
    </n-card>

    <n-modal v-model:show="showAddModal" preset="dialog" title="添加基金">
      <n-form>
        <n-form-item label="基金代码">
          <n-input v-model:value="newFundCode" placeholder="请输入基金代码，如 000001" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" @click="handleAddFund">确定</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showDetailModal" preset="card" :title="selectedFund ? selectedFund.name : '基金详情'" style="width: 960px;">
      <n-spin :show="detailLoading">
        <div v-if="fundOverview">
          <div class="detail-basic">
            <n-card size="small" class="basic-card">
              <div class="info-row">
                <div>代码：{{ fundOverview.detail?.code }}</div>
                <div>类型：{{ fundOverview.detail?.type }}</div>
                <div>风险等级：{{ fundOverview.detail?.riskLevel || '-' }}</div>
              </div>
              <div class="info-row">
                <div>基金经理：{{ fundOverview.detail?.manager || '-' }}</div>
                <div>基金公司：{{ fundOverview.detail?.company || '-' }}</div>
                <div>成立日期：{{ fundOverview.detail?.inceptionDate || '-' }}</div>
              </div>
              <div class="info-row">
                <div>最新净值：{{ fundOverview.price?.nav?.toFixed(4) || '-' }}</div>
                <div>估算净值：{{ fundOverview.price?.estimate?.toFixed(4) || '-' }}</div>
                <div>估算涨跌：{{ fundOverview.price?.changePercent?.toFixed(2) || '-' }}%</div>
              </div>
            </n-card>
          </div>

          <div class="detail-chart">
            <n-card size="small" title="净值走势">
              <div ref="historyChartRef" class="history-chart"></div>
            </n-card>
          </div>

          <div class="detail-grid">
            <n-card size="small" title="前十大重仓">
              <n-empty v-if="!fundOverview.stockHoldings?.length" description="暂无数据" />
              <n-list v-else>
                <n-list-item v-for="item in fundOverview.stockHoldings.slice(0, 10)" :key="item.code">
                  <n-thing>
                    <template #header>
                      <span>{{ item.name }} ({{ item.code }})</span>
                    </template>
                    <template #description>
                      占比 {{ item.ratio.toFixed(2) }}% | 行业 {{ item.industry || '-' }} | 变动 {{ item.trend || '-' }} {{ item.change?.toFixed(2) || '0.00' }}%
                    </template>
                  </n-thing>
                </n-list-item>
              </n-list>
            </n-card>
            <n-card size="small" title="最近公告">
              <n-empty v-if="!fundOverview.notices?.length" description="暂无公告" />
              <n-list v-else>
                <n-list-item v-for="item in fundOverview.notices.slice(0, 8)" :key="item.id">
                  <n-thing>
                    <template #header>
                      <span class="notice-link" @click="goNoticeDetail(item.url)">{{ item.title }}</span>
                    </template>
                    <template #description>
                      <div class="notice-desc">
                        <span class="notice-date">{{ item.date }}</span>
                        <n-button
                          size="tiny"
                          type="warning"
                          tertiary
                          class="notice-ai-btn"
                          @click="openSummaryModal(item)"
                        >
                          AI解读
                        </n-button>
                      </div>
                    </template>
                  </n-thing>
                </n-list-item>
              </n-list>
            </n-card>
          </div>
        </div>
      </n-spin>
    </n-modal>

    <n-modal v-model:show="showPositionModal" preset="card" :title="'持仓管理 - ' + (selectedFund?.name || '')" style="width: 600px;">
      <n-spin :show="positionLoading">
        <n-form label-width="80px">
          <n-form-item label="买入净值">
            <n-input-number v-model:value="positionForm.buyNav" :precision="4" :min="0" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="买入日期">
            <n-input v-model:value="positionForm.buyDate" placeholder="YYYY-MM-DD" />
          </n-form-item>
          <n-form-item label="持有份额">
            <n-input-number v-model:value="positionForm.share" :min="0" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="成本净值">
            <n-input-number v-model:value="positionForm.costNav" :precision="4" :min="0" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="目标净值">
            <n-input-number v-model:value="positionForm.targetNav" :precision="4" :min="0" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="止损净值">
            <n-input-number v-model:value="positionForm.stopLossNav" :precision="4" :min="0" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="备注">
            <n-input type="textarea" v-model:value="positionForm.notes" :autosize="{ minRows: 2, maxRows: 4 }" />
          </n-form-item>
        </n-form>
        <n-space justify="space-between">
          <n-button v-if="currentPosition" type="error" @click="deleteFundPosition">删除持仓</n-button>
          <div></div>
          <n-space>
            <n-button @click="showPositionModal = false">取消</n-button>
            <n-button type="primary" @click="saveFundPosition">{{ currentPosition ? '更新持仓' : '保存持仓' }}</n-button>
          </n-space>
        </n-space>
      </n-spin>
    </n-modal>

    <n-modal v-model:show="showAlertModal" preset="card" :title="'提醒设置 - ' + (selectedFund?.name || '')" style="width: 640px;">
      <n-spin :show="alertLoading">
        <n-card size="small" title="新增提醒" class="alert-form">
          <n-form label-width="80px">
            <n-form-item label="提醒类型">
              <n-radio-group v-model:value="alertForm.alertType">
                <n-radio value="nav">净值提醒</n-radio>
                <n-radio value="change">涨跌提醒</n-radio>
              </n-radio-group>
            </n-form-item>
            <n-form-item label="触发条件">
              <n-space align="center">
                <n-select v-model:value="alertForm.condition" :options="alertForm.alertType === 'nav' ? [
                  { label: '高于', value: 'above' },
                  { label: '低于', value: 'below' }
                ] : [
                  { label: '涨幅达到', value: 'above' },
                  { label: '跌幅达到', value: 'below' }
                ]" style="width: 120px;" />
                <n-input-number
                  v-model:value="alertForm.targetValue"
                  :precision="2"
                  :min="0"
                  style="width: 140px;"
                />
                <span>{{ alertForm.alertType === 'nav' ? '' : '%' }}</span>
              </n-space>
            </n-form-item>
            <n-form-item>
              <n-button type="primary" @click="addFundAlert">添加提醒</n-button>
            </n-form-item>
          </n-form>
        </n-card>
        <n-card size="small" title="已设置的提醒">
          <n-empty v-if="!fundAlerts.length" description="暂无提醒" />
          <n-list v-else>
            <n-list-item v-for="alert in fundAlerts" :key="alert.id">
              <n-thing>
                <template #header>
                  <n-space align="center">
                    <n-tag size="small" :type="alert.alertType === 'change' ? 'warning' : 'info'">
                      {{ alert.alertType === 'change' ? '涨跌' : '净值' }}
                    </n-tag>
                    <span>{{ alert.condition === 'above' ? '高于' : '低于' }} {{ alert.targetValue }}</span>
                    <n-tag v-if="alert.triggered" size="small" type="success">已触发</n-tag>
                  </n-space>
                </template>
                <template #header-extra>
                  <n-space>
                    <n-switch :value="alert.enabled" size="small" @update:value="() => toggleFundAlert(alert)" />
                    <n-button v-if="alert.triggered" size="tiny" @click="resetFundAlert(alert)">重置</n-button>
                    <n-button size="tiny" type="error" @click="removeFundAlert(alert.id)">删除</n-button>
                  </n-space>
                </template>
                <template #description>
                  <span style="color: #999;">创建时间：{{ new Date(alert.createdAt).toLocaleString() }}</span>
                </template>
              </n-thing>
            </n-list-item>
          </n-list>
        </n-card>
      </n-spin>
    </n-modal>

    <n-modal v-model:show="showAIModal" preset="card" :title="'AI分析 - ' + (selectedFund?.name || '')" style="width: 900px;">
      <div class="ai-analysis-container">
        <n-alert type="warning" :bordered="false" class="disclaimer-alert">
          <template #icon>
            <span style="font-size: 16px;">⚠️</span>
          </template>
          <div class="disclaimer-text">
            <strong>免责声明：</strong>AI分析仅供学习研究参考，不构成任何投资建议。投资有风险，入市需谨慎。
          </div>
        </n-alert>

        <div class="ai-chat-container">
          <n-scrollbar ref="aiScrollbarRef" style="max-height: 400px;">
            <div class="ai-messages">
              <div v-if="aiMessages.length === 0" class="ai-empty-tip">
                <n-spin v-if="aiLoading" size="small" />
                <div v-else>点击下方按钮开始AI分析，或直接输入问题进行对话</div>
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

        <div class="ai-input-area">
          <n-space vertical style="width: 100%;">
            <n-space>
              <n-button type="primary" :disabled="aiLoading" @click="startFundAIAnalysis">
                重新分析
              </n-button>
            </n-space>
            <n-input
              v-model:value="aiQuestion"
              type="textarea"
              placeholder="输入问题继续对话，如：这只基金适合定投吗？"
              :autosize="{ minRows: 2, maxRows: 3 }"
              :disabled="aiLoading"
              @keyup.enter.exact="sendAIQuestion"
            />
            <n-button type="info" :loading="aiLoading" :disabled="aiLoading || !aiQuestion.trim()" @click="sendAIQuestion">
              发送问题
            </n-button>
          </n-space>
        </div>
      </div>
    </n-modal>

    <n-modal v-model:show="showSummaryModal" preset="card" :title="'AI解读 - ' + (summaryType === 'report' ? '研报' : '公告')" style="width: 700px;">
      <div class="ai-summary-container">
        <n-alert type="warning" :bordered="false" class="disclaimer-alert">
          <template #icon>
            <span style="font-size: 16px;">⚠️</span>
          </template>
          <div class="disclaimer-text">
            <strong>免责声明：</strong>以下内容由AI生成，仅供学习研究参考，不构成任何投资建议。投资有风险，入市需谨慎，盈亏自负。
          </div>
        </n-alert>

        <div class="summary-title">{{ summaryTitle }}</div>
        <div class="summary-content-area">
          <n-scrollbar ref="summaryScrollbarRef" style="max-height: 320px;">
            <n-spin v-if="summaryLoading && !summaryContent" size="small" style="display: block; margin: 40px auto;" />
            <div v-else class="markdown-content" v-html="formatContent(summaryContent)"></div>
          </n-scrollbar>
        </div>
        <div v-if="summaryLoading" class="summary-loading-tip">AI正在解读中，请稍候...</div>

        <n-space justify="space-between">
          <n-button text size="small" v-if="summaryUrl" @click="OpenURL(summaryUrl)">打开原文</n-button>
          <span></span>
        </n-space>

        <div class="manual-input-section">
          <n-button
            v-if="!showManualInput && !summaryLoading"
            size="small"
            type="info"
            dashed
            @click="showManualInput = true"
          >
            内容抓取不完整？点击手动粘贴
          </n-button>
          <div v-if="showManualInput" class="manual-input-area">
            <n-input
              v-model:value="manualContent"
              type="textarea"
              placeholder="从浏览器复制公告/研报内容粘贴到这里，AI会基于文本重新解读"
              :autosize="{ minRows: 4, maxRows: 8 }"
            />
            <n-space style="margin-top: 8px;" justify="end">
              <n-button size="small" @click="showManualInput = false; manualContent = ''">取消</n-button>
              <n-button type="primary" size="small" :disabled="!manualContent.trim()" @click="analyzeManualSummary">重新解读</n-button>
            </n-space>
          </div>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
.fund-page {
  height: 100%;
}

.detail-basic {
  margin-bottom: 12px;
}

.basic-card .info-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
  font-size: 13px;
}

.detail-chart {
  margin-bottom: 12px;
}

.history-chart {
  width: 100%;
  height: 300px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 12px;
}

.notice-link {
  color: #18a058;
  cursor: pointer;
}

.notice-desc {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}

.notice-date {
  color: rgba(255, 255, 255, 0.65);
}

.notice-ai-btn :deep(.n-button__content) {
  font-weight: 600;
}

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

.disclaimer-alert {
  background: rgba(250, 173, 20, 0.1) !important;
  border: 1px solid rgba(250, 173, 20, 0.3) !important;
}

.disclaimer-text {
  font-size: 12px;
  color: #faad14;
}

.ai-summary-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary-title {
  font-size: 16px;
  font-weight: 600;
}

.summary-content-area {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  padding: 12px;
  min-height: 200px;
}

.summary-loading-tip {
  font-size: 13px;
  color: #999;
}

.manual-input-section {
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.manual-input-area {
  width: 100%;
}

.alert-form {
  margin-bottom: 12px;
}
</style>
