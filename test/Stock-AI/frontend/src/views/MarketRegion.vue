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
import { GetGlobalIndices } from '../../wailsjs/go/main/App'

const props = defineProps({
  region: {
    type: String,
    required: true
  }
})

const message = useMessage()
const loading = ref(false)
const globalIndices = ref([])

// 地区名称映射
const regionNames = {
  asia: '亚洲',
  europe: '欧洲',
  america: '美洲',
  oceania: '大洋洲'
}

// 过滤当前地区的指数
const regionIndices = computed(() => {
  return globalIndices.value.filter(idx => idx.region === props.region)
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

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const data = await GetGlobalIndices()
    globalIndices.value = data || []
  } catch (e) {
    console.error('加载全球指数失败:', e)
    message.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    const data = await GetGlobalIndices()
    globalIndices.value = data || []
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
  <div class="market-region-page">
    <n-spin :show="loading">
      <n-card :title="regionNames[region] + '股市'" :bordered="false">
        <template #header-extra>
          <n-button type="primary" @click="refreshData" :loading="loading">
            刷新数据
          </n-button>
        </template>

        <n-data-table
          :columns="columns"
          :data="regionIndices"
          :bordered="false"
          striped
          size="small"
          :max-height="600"
        />
        <div v-if="regionIndices.length === 0 && !loading" class="empty-tip">
          暂无{{ regionNames[region] }}股市数据
        </div>
      </n-card>
    </n-spin>
  </div>
</template>

<style scoped>
.market-region-page {
  height: 100%;
}

.empty-tip {
  text-align: center;
  color: #666;
  padding: 40px;
}
</style>
