<script setup>
import { ref, onMounted, computed } from 'vue'
import {
  NCard,
  NDataTable,
  NButton,
  NSpin,
  useMessage
} from 'naive-ui'
import { h } from 'vue'
import { GetForexRates } from '../../wailsjs/go/main/App'

const props = defineProps({
  category: {
    type: String,
    required: true
  }
})

const message = useMessage()
const loading = ref(false)
const forexRates = ref([])

// 类别名称映射
const categoryNames = {
  'major': '主要货币对',
  'cross': '交叉货币对',
  'cny': '人民币相关'
}

// 过滤当前类别的外汇
const categoryForex = computed(() => {
  return forexRates.value.filter(f => f.category === props.category)
})

// 渲染涨跌幅
const renderChange = (val) => {
  if (val === undefined || val === null) return '-'
  const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
  return h('span', { style: { color, fontWeight: 'bold' } }, (val > 0 ? '+' : '') + val.toFixed(2) + '%')
}

// 列定义
const columns = [
  { title: '货币对', key: 'pair', width: 100 },
  { title: '名称', key: 'name', width: 150 },
  { title: '汇率', key: 'rate', width: 120, render: (row) => row.rate ? row.rate.toFixed(4) : '-' },
  { title: '涨跌', key: 'change', width: 100, render: (row) => {
    const val = row.change
    if (val === undefined || val === null) return '-'
    const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
    return h('span', { style: { color } }, (val > 0 ? '+' : '') + val.toFixed(4))
  }},
  { title: '涨跌幅', key: 'changePercent', width: 100, render: (row) => renderChange(row.changePercent) },
  { title: '最高', key: 'high', width: 100, render: (row) => row.high ? row.high.toFixed(4) : '-' },
  { title: '最低', key: 'low', width: 100, render: (row) => row.low ? row.low.toFixed(4) : '-' },
  { title: '更新时间', key: 'updateTime', width: 100 }
]

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const data = await GetForexRates()
    forexRates.value = data || []
  } catch (e) {
    console.error('加载外汇数据失败:', e)
    message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    const data = await GetForexRates()
    forexRates.value = data || []
    message.success('刷新成功')
  } catch (e) {
    message.error('刷新失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="forex-category-page">
    <n-spin :show="loading">
      <n-card :title="categoryNames[category]" :bordered="false">
        <template #header-extra>
          <n-button type="primary" @click="refreshData" :loading="loading">
            刷新数据
          </n-button>
        </template>

        <n-data-table
          :columns="columns"
          :data="categoryForex"
          :bordered="false"
          striped
          size="small"
          :max-height="600"
        />
        <div v-if="categoryForex.length === 0 && !loading" class="empty-tip">
          暂无{{ categoryNames[category] }}数据
        </div>
      </n-card>
    </n-spin>
  </div>
</template>

<style scoped>
.forex-category-page {
  height: 100%;
}

.empty-tip {
  text-align: center;
  color: #666;
  padding: 40px;
}
</style>
