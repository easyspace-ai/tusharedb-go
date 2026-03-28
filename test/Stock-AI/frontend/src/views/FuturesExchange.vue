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
import { GetFuturesProducts, GetMainContracts } from '../../wailsjs/go/main/App'

const props = defineProps({
  exchange: {
    type: String,
    required: true
  }
})

const message = useMessage()
const loading = ref(false)
const mainContracts = ref([])
const futuresProducts = ref([])

// 交易所名称映射
const exchangeNames = {
  'SHFE': '上海期货交易所',
  'DCE': '大连商品交易所',
  'CZCE': '郑州商品交易所',
  'CFFEX': '中国金融期货交易所',
  'INE': '上海国际能源交易中心'
}

// 过滤当前交易所的产品
const exchangeProducts = computed(() => {
  return futuresProducts.value.filter(p => p.exchange === props.exchange)
})

// 过滤当前交易所的主力合约
const exchangeContracts = computed(() => {
  return mainContracts.value.filter(c => {
    // 根据合约代码判断交易所
    const code = c.code || ''
    const exchangeMap = {
      'SHFE': ['CU', 'AL', 'ZN', 'PB', 'NI', 'SN', 'AU', 'AG', 'RB', 'WR', 'HC', 'SS', 'BU', 'RU', 'SP', 'FU'],
      'DCE': ['C', 'CS', 'A', 'B', 'M', 'Y', 'P', 'FB', 'BB', 'JD', 'L', 'V', 'PP', 'J', 'JM', 'I', 'EG', 'EB', 'PG', 'LH'],
      'CZCE': ['SR', 'CF', 'CY', 'PM', 'WH', 'RI', 'LR', 'JR', 'RS', 'OI', 'RM', 'TA', 'MA', 'FG', 'SF', 'SM', 'ZC', 'AP', 'CJ', 'UR', 'SA', 'PF', 'PK'],
      'CFFEX': ['IF', 'IC', 'IH', 'IM', 'T', 'TF', 'TS'],
      'INE': ['SC', 'LU', 'NR', 'BC']
    }
    const prefixes = exchangeMap[props.exchange] || []
    return prefixes.some(prefix => code.toUpperCase().startsWith(prefix))
  })
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
const contractColumns = [
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
  { title: '品种', key: 'name', width: 120 },
  { title: '代码', key: 'code', width: 80 },
  { title: '类别', key: 'category', width: 100 },
  { title: '交易单位', key: 'unit', width: 120 },
  { title: '最小变动', key: 'minMove', width: 100 }
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
    message.error('加载数据失败')
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

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="futures-exchange-page">
    <n-spin :show="loading">
      <n-card :title="exchangeNames[exchange]" :bordered="false">
        <template #header-extra>
          <n-button type="primary" @click="refreshData" :loading="loading">
            刷新数据
          </n-button>
        </template>

        <!-- 主力合约 -->
        <n-card title="主力合约" size="small" style="margin-bottom: 16px;">
          <n-data-table
            :columns="contractColumns"
            :data="exchangeContracts"
            :bordered="false"
            striped
            size="small"
            :max-height="400"
          />
          <div v-if="exchangeContracts.length === 0 && !loading" class="empty-tip">
            暂无主力合约数据
          </div>
        </n-card>

        <!-- 期货品种 -->
        <n-card title="期货品种" size="small">
          <n-data-table
            :columns="productColumns"
            :data="exchangeProducts"
            :bordered="false"
            striped
            size="small"
            :max-height="300"
          />
          <div v-if="exchangeProducts.length === 0 && !loading" class="empty-tip">
            暂无期货品种数据
          </div>
        </n-card>
      </n-card>
    </n-spin>
  </div>
</template>

<style scoped>
.futures-exchange-page {
  height: 100%;
}

.empty-tip {
  text-align: center;
  color: #666;
  padding: 40px;
}
</style>
