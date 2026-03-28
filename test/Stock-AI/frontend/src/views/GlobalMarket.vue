<script setup>
import { ref, onMounted, computed } from 'vue'
import {
  NCard,
  NTabs,
  NTabPane,
  NDataTable,
  NButton,
  NSpace,
  NTag,
  NInput,
  NModal,
  NForm,
  NFormItem,
  NSelect,
  NEmpty,
  useMessage
} from 'naive-ui'
import { h } from 'vue'
import {
  // 期货
  GetFuturesProducts,
  GetMainContracts,
  GetFuturesList,
  AddFutures,
  RemoveFutures,
  GetFuturesPrice,
  // 美股
  GetPopularUSStocks,
  GetUSStockList,
  AddUSStock,
  RemoveUSStock,
  GetUSStockPrice,
  // 港股
  GetPopularHKStocks,
  GetHKStockList,
  AddHKStock,
  RemoveHKStock,
  GetHKStockPrice,
  // 全球指数
  GetGlobalIndices,
  // 数字货币
  GetMainCryptos,
  GetCryptoList,
  AddCrypto,
  RemoveCrypto,
  GetCryptoPrice,
  // 外汇
  GetForexRates
} from '../../wailsjs/go/main/App'

const message = useMessage()
const loading = ref(false)
const activeTab = ref('indices')

// 数据
const globalIndices = ref([])
const mainContracts = ref([])
const myFutures = ref([])
const popularUSStocks = ref([])
const myUSStocks = ref([])
const popularHKStocks = ref([])
const myHKStocks = ref([])
const mainCryptos = ref([])
const myCryptos = ref([])
const forexRates = ref([])

// 弹窗
const showAddModal = ref(false)
const addType = ref('')
const addForm = ref({
  code: '',
  name: '',
  nameCN: '',
  exchange: ''
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

// 全球指数列表
const indicesColumns = [
  { title: '指数', key: 'nameCn', width: 150 },
  { title: '代码', key: 'code', width: 100 },
  { title: '最新', key: 'price', width: 100, render: (row) => renderPrice(row.price, 2) },
  { title: '涨跌', key: 'change', width: 100, render: (row) => {
    const val = row.change
    if (val === undefined || val === null) return '-'
    const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
    return h('span', { style: { color } }, (val > 0 ? '+' : '') + val.toFixed(2))
  }},
  { title: '涨跌幅', key: 'changePercent', width: 100, render: (row) => renderChange(row.changePercent) },
  { title: '地区', key: 'region', width: 80, render: (row) => {
    const regionMap = { asia: '亚洲', europe: '欧洲', america: '美洲', oceania: '大洋洲' }
    return regionMap[row.region] || row.region
  }}
]

// 期货列表
const futuresColumns = [
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

// 美股列表
const usStockColumns = [
  { title: '代码', key: 'symbol', width: 80 },
  { title: '名称', key: 'nameCn', width: 100 },
  { title: '英文名', key: 'name', width: 150, ellipsis: true },
  { title: '最新价', key: 'price', width: 100, render: (row) => row.price ? '$' + renderPrice(row.price, 2) : '-' },
  { title: '涨跌幅', key: 'changePercent', width: 100, render: (row) => renderChange(row.changePercent) },
  { title: '交易所', key: 'exchange', width: 80 }
]

// 港股列表
const hkStockColumns = [
  { title: '代码', key: 'code', width: 80 },
  { title: '名称', key: 'name', width: 150 },
  { title: '最新价', key: 'price', width: 100, render: (row) => row.price ? 'HK$' + renderPrice(row.price, 3) : '-' },
  { title: '涨跌幅', key: 'changePercent', width: 100, render: (row) => renderChange(row.changePercent) },
  { title: '成交量', key: 'volume', width: 100 }
]

// 数字货币列表
const cryptoColumns = [
  { title: '交易对', key: 'symbol', width: 100 },
  { title: '名称', key: 'nameCn', width: 80 },
  { title: '最新价', key: 'price', width: 120, render: (row) => row.price ? '$' + row.price.toFixed(row.price > 100 ? 2 : 4) : '-' },
  { title: '24h涨跌', key: 'changePercent', width: 100, render: (row) => renderChange(row.changePercent) },
  { title: '24h最高', key: 'high24h', width: 100, render: (row) => row.high24h ? '$' + row.high24h.toFixed(2) : '-' },
  { title: '24h最低', key: 'low24h', width: 100, render: (row) => row.low24h ? '$' + row.low24h.toFixed(2) : '-' },
  { title: '24h成交额', key: 'amount24h', width: 120, render: (row) => {
    if (!row.amount24h) return '-'
    if (row.amount24h > 1e9) return '$' + (row.amount24h / 1e9).toFixed(2) + 'B'
    if (row.amount24h > 1e6) return '$' + (row.amount24h / 1e6).toFixed(2) + 'M'
    return '$' + row.amount24h.toFixed(0)
  }}
]

// 外汇列表
const forexColumns = [
  { title: '货币对', key: 'pair', width: 100 },
  { title: '名称', key: 'name', width: 120 },
  { title: '汇率', key: 'rate', width: 100, render: (row) => renderPrice(row.rate, 4) },
  { title: '涨跌', key: 'change', width: 100, render: (row) => {
    const val = row.change
    if (val === undefined || val === null) return '-'
    const color = val > 0 ? '#f5222d' : val < 0 ? '#52c41a' : '#999'
    return h('span', { style: { color } }, (val > 0 ? '+' : '') + val.toFixed(4))
  }},
  { title: '涨跌幅', key: 'changePercent', width: 100, render: (row) => renderChange(row.changePercent) }
]

// 加载全球指数
const loadGlobalIndices = async () => {
  try {
    const data = await GetGlobalIndices()
    globalIndices.value = data || []
  } catch (e) {
    console.error('加载全球指数失败:', e)
  }
}

// 加载期货数据
const loadFutures = async () => {
  try {
    const [contracts, myList] = await Promise.all([
      GetMainContracts(),
      GetFuturesList()
    ])
    mainContracts.value = contracts || []
    myFutures.value = myList || []
  } catch (e) {
    console.error('加载期货数据失败:', e)
  }
}

// 加载美股数据
const loadUSStocks = async () => {
  try {
    const [popular, myList] = await Promise.all([
      GetPopularUSStocks(),
      GetUSStockList()
    ])
    popularUSStocks.value = popular || []
    myUSStocks.value = myList || []

    // 获取实时价格
    if (popular && popular.length > 0) {
      const symbols = popular.map(s => s.symbol)
      const prices = await GetUSStockPrice(symbols)
      if (prices) {
        popularUSStocks.value = popular.map(s => ({
          ...s,
          ...prices[s.symbol]
        }))
      }
    }
  } catch (e) {
    console.error('加载美股数据失败:', e)
  }
}

// 加载港股数据
const loadHKStocks = async () => {
  try {
    const [popular, myList] = await Promise.all([
      GetPopularHKStocks(),
      GetHKStockList()
    ])
    popularHKStocks.value = popular || []
    myHKStocks.value = myList || []

    // 获取实时价格
    if (popular && popular.length > 0) {
      const codes = popular.map(s => s.code)
      const prices = await GetHKStockPrice(codes)
      if (prices) {
        popularHKStocks.value = popular.map(s => ({
          ...s,
          ...prices[s.code]
        }))
      }
    }
  } catch (e) {
    console.error('加载港股数据失败:', e)
  }
}

// 加载数字货币数据
const loadCryptos = async () => {
  try {
    const [main, myList] = await Promise.all([
      GetMainCryptos(),
      GetCryptoList()
    ])
    mainCryptos.value = main || []
    myCryptos.value = myList || []

    // 获取实时价格
    if (main && main.length > 0) {
      const symbols = main.map(c => c.symbol)
      const prices = await GetCryptoPrice(symbols)
      if (prices) {
        mainCryptos.value = main.map(c => ({
          ...c,
          ...prices[c.symbol]
        }))
      }
    }
  } catch (e) {
    console.error('加载数字货币数据失败:', e)
  }
}

// 加载外汇数据
const loadForex = async () => {
  try {
    const data = await GetForexRates()
    forexRates.value = data || []
  } catch (e) {
    console.error('加载外汇数据失败:', e)
  }
}

// 加载所有数据
const loadAllData = async () => {
  loading.value = true
  try {
    await Promise.all([
      loadGlobalIndices(),
      loadFutures(),
      loadUSStocks(),
      loadHKStocks(),
      loadCryptos(),
      loadForex()
    ])
  } finally {
    loading.value = false
  }
}

// 刷新当前标签页数据
const refreshCurrentTab = async () => {
  loading.value = true
  try {
    switch (activeTab.value) {
      case 'indices':
        await loadGlobalIndices()
        break
      case 'futures':
        await loadFutures()
        break
      case 'us':
        await loadUSStocks()
        break
      case 'hk':
        await loadHKStocks()
        break
      case 'crypto':
        await loadCryptos()
        break
      case 'forex':
        await loadForex()
        break
    }
    message.success('刷新成功')
  } catch (e) {
    message.error('刷新失败')
  } finally {
    loading.value = false
  }
}

// 打开添加弹窗
const openAddModal = (type) => {
  addType.value = type
  addForm.value = { code: '', name: '', nameCN: '', exchange: '' }
  showAddModal.value = true
}

// 确认添加
const handleAdd = async () => {
  try {
    switch (addType.value) {
      case 'futures':
        await AddFutures(addForm.value.code, addForm.value.name, addForm.value.exchange)
        await loadFutures()
        break
      case 'us':
        await AddUSStock(addForm.value.code, addForm.value.name, addForm.value.nameCN, addForm.value.exchange)
        await loadUSStocks()
        break
      case 'hk':
        await AddHKStock(addForm.value.code, addForm.value.name, addForm.value.nameCN)
        await loadHKStocks()
        break
      case 'crypto':
        await AddCrypto(addForm.value.code, addForm.value.name, addForm.value.nameCN)
        await loadCryptos()
        break
    }
    message.success('添加成功')
    showAddModal.value = false
  } catch (e) {
    message.error('添加失败: ' + e)
  }
}

// 删除
const handleRemove = async (type, code) => {
  try {
    switch (type) {
      case 'futures':
        await RemoveFutures(code)
        await loadFutures()
        break
      case 'us':
        await RemoveUSStock(code)
        await loadUSStocks()
        break
      case 'hk':
        await RemoveHKStock(code)
        await loadHKStocks()
        break
      case 'crypto':
        await RemoveCrypto(code)
        await loadCryptos()
        break
    }
    message.success('删除成功')
  } catch (e) {
    message.error('删除失败: ' + e)
  }
}

// 添加弹窗标题
const addModalTitle = computed(() => {
  const titles = {
    futures: '添加期货',
    us: '添加美股',
    hk: '添加港股',
    crypto: '添加数字货币'
  }
  return titles[addType.value] || '添加'
})

onMounted(() => {
  loadAllData()
})
</script>

<template>
  <div class="global-market-page">
    <n-card title="全球市场" :bordered="false">
      <template #header-extra>
        <n-button type="primary" @click="refreshCurrentTab" :loading="loading">
          刷新数据
        </n-button>
      </template>

      <n-tabs v-model:value="activeTab" type="line" animated>
        <!-- 全球指数 -->
        <n-tab-pane name="indices" tab="全球指数">
          <n-data-table
            :columns="indicesColumns"
            :data="globalIndices"
            :loading="loading"
            :bordered="false"
            striped
            size="small"
          />
        </n-tab-pane>

        <!-- 期货 -->
        <n-tab-pane name="futures" tab="期货">
          <n-space vertical>
            <n-card title="主力合约" size="small">
              <n-data-table
                :columns="futuresColumns"
                :data="mainContracts"
                :loading="loading"
                :bordered="false"
                striped
                size="small"
                :max-height="400"
              />
            </n-card>
          </n-space>
        </n-tab-pane>

        <!-- 美股 -->
        <n-tab-pane name="us" tab="美股">
          <n-space vertical>
            <n-card title="热门美股" size="small">
              <n-data-table
                :columns="usStockColumns"
                :data="popularUSStocks"
                :loading="loading"
                :bordered="false"
                striped
                size="small"
                :max-height="500"
              />
            </n-card>
          </n-space>
        </n-tab-pane>

        <!-- 港股 -->
        <n-tab-pane name="hk" tab="港股">
          <n-space vertical>
            <n-card title="热门港股" size="small">
              <n-data-table
                :columns="hkStockColumns"
                :data="popularHKStocks"
                :loading="loading"
                :bordered="false"
                striped
                size="small"
                :max-height="500"
              />
            </n-card>
          </n-space>
        </n-tab-pane>

        <!-- 数字货币 -->
        <n-tab-pane name="crypto" tab="数字货币">
          <n-space vertical>
            <n-card title="主流数字货币" size="small">
              <n-data-table
                :columns="cryptoColumns"
                :data="mainCryptos"
                :loading="loading"
                :bordered="false"
                striped
                size="small"
                :max-height="500"
              />
            </n-card>
          </n-space>
        </n-tab-pane>

        <!-- 外汇 -->
        <n-tab-pane name="forex" tab="外汇">
          <n-data-table
            :columns="forexColumns"
            :data="forexRates"
            :loading="loading"
            :bordered="false"
            striped
            size="small"
          />
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 添加弹窗 -->
    <n-modal v-model:show="showAddModal" preset="dialog" :title="addModalTitle">
      <n-form>
        <n-form-item label="代码">
          <n-input v-model:value="addForm.code" placeholder="请输入代码" />
        </n-form-item>
        <n-form-item label="名称">
          <n-input v-model:value="addForm.name" placeholder="请输入名称" />
        </n-form-item>
        <n-form-item v-if="addType !== 'futures'" label="中文名">
          <n-input v-model:value="addForm.nameCN" placeholder="请输入中文名" />
        </n-form-item>
        <n-form-item v-if="addType === 'futures' || addType === 'us'" label="交易所">
          <n-input v-model:value="addForm.exchange" placeholder="请输入交易所" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" @click="handleAdd">确定</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.global-market-page {
  height: 100%;
}
</style>
