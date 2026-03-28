<script setup>
import { ref, onMounted, computed } from 'vue'
import {
  NCard,
  NButton,
  NSpace,
  NSelect,
  NAlert,
  NEmpty,
  NTag,
  NSpin,
  NResult,
  NDivider,
  NText,
  useMessage
} from 'naive-ui'
import {
  ListPrompts,
  ExecuteScreenerPrompt,
  ExecuteReviewPrompt
} from '../../wailsjs/go/main/App'

const message = useMessage()
const loading = ref(false)
const activeTab = ref('screener')

// 选股相关
const screenerPrompts = ref([])
const selectedScreener = ref(null)
const screenerResult = ref(null)

// 复盘相关
const reviewPrompts = ref([])
const selectedReview = ref(null)
const reviewResult = ref(null)

// 加载选股提示词
const loadScreenerPrompts = async () => {
  try {
    const data = await ListPrompts('screener')
    screenerPrompts.value = data || []
  } catch (e) {
    console.error('加载选股提示词失败:', e)
  }
}

// 加载复盘提示词
const loadReviewPrompts = async () => {
  try {
    const data = await ListPrompts('review')
    reviewPrompts.value = data || []
  } catch (e) {
    console.error('加载复盘提示词失败:', e)
  }
}

// 执行选股
const runScreener = async () => {
  if (!selectedScreener.value) {
    message.warning('请选择选股提示词')
    return
  }

  loading.value = true
  screenerResult.value = null
  try {
    const result = await ExecuteScreenerPrompt(selectedScreener.value)
    screenerResult.value = result
    message.success('选股分析完成')
  } catch (e) {
    message.error('选股失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 执行复盘
const runReview = async () => {
  if (!selectedReview.value) {
    message.warning('请选择复盘提示词')
    return
  }

  loading.value = true
  reviewResult.value = null
  try {
    const result = await ExecuteReviewPrompt(selectedReview.value)
    reviewResult.value = result
    message.success('复盘分析完成')
  } catch (e) {
    message.error('复盘失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 选股提示词选项
const screenerOptions = computed(() => {
  return screenerPrompts.value.map(p => ({
    label: p.name,
    value: p.name
  }))
})

// 复盘提示词选项
const reviewOptions = computed(() => {
  return reviewPrompts.value.map(p => ({
    label: p.name,
    value: p.name
  }))
})

onMounted(() => {
  loadScreenerPrompts()
  loadReviewPrompts()
})
</script>

<template>
  <div class="ai-analysis-page">
    <n-card title="AI智能分析" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-tag type="info">基于AI提示词</n-tag>
        </n-space>
      </template>

      <n-spin :show="loading">
        <!-- AI选股 -->
        <n-card title="AI选股" size="small" style="margin-bottom: 16px;">
          <n-alert type="info" style="margin-bottom: 16px;">
            选择一个选股提示词，AI将根据您的自选股列表进行分析筛选。
            <br />
            <n-text depth="3">提示：在「AI提示词」页面创建选股提示词</n-text>
          </n-alert>

          <n-space vertical>
            <n-space>
              <n-select
                v-model:value="selectedScreener"
                :options="screenerOptions"
                placeholder="选择选股提示词"
                style="width: 300px;"
                :disabled="screenerPrompts.length === 0"
              />
              <n-button
                type="primary"
                @click="runScreener"
                :disabled="!selectedScreener || loading"
                :loading="loading && activeTab === 'screener'"
              >
                开始选股
              </n-button>
            </n-space>

            <n-empty v-if="screenerPrompts.length === 0" description="暂无选股提示词，请先创建" />

            <!-- 选股结果 -->
            <div v-if="screenerResult" class="result-box">
              <n-divider>选股结果</n-divider>
              <div class="result-content">
                <div class="result-summary">{{ screenerResult.summary }}</div>
                <n-divider dashed>详细分析</n-divider>
                <pre class="result-raw">{{ screenerResult.raw }}</pre>
              </div>
            </div>
          </n-space>
        </n-card>

        <!-- AI复盘 -->
        <n-card title="AI复盘" size="small">
          <n-alert type="info" style="margin-bottom: 16px;">
            选择一个复盘提示词，AI将根据您的持仓数据进行分析复盘。
            <br />
            <n-text depth="3">提示：在「AI提示词」页面创建复盘提示词</n-text>
          </n-alert>

          <n-space vertical>
            <n-space>
              <n-select
                v-model:value="selectedReview"
                :options="reviewOptions"
                placeholder="选择复盘提示词"
                style="width: 300px;"
                :disabled="reviewPrompts.length === 0"
              />
              <n-button
                type="primary"
                @click="runReview"
                :disabled="!selectedReview || loading"
                :loading="loading && activeTab === 'review'"
              >
                开始复盘
              </n-button>
            </n-space>

            <n-empty v-if="reviewPrompts.length === 0" description="暂无复盘提示词，请先创建" />

            <!-- 复盘结果 -->
            <div v-if="reviewResult" class="result-box">
              <n-divider>复盘结果</n-divider>
              <div class="result-content">
                <div class="result-summary">{{ reviewResult.summary }}</div>
                <n-divider dashed>详细分析</n-divider>
                <pre class="result-raw">{{ reviewResult.raw }}</pre>
              </div>
            </div>
          </n-space>
        </n-card>
      </n-spin>
    </n-card>
  </div>
</template>

<style scoped>
.ai-analysis-page {
  padding: 16px;
}

.result-box {
  margin-top: 16px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.04);
  border-radius: 8px;
}

.result-content {
  color: #e0e0e0;
}

.result-summary {
  font-size: 14px;
  line-height: 1.6;
  margin-bottom: 16px;
}

.result-raw {
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  background: rgba(0, 0, 0, 0.2);
  padding: 12px;
  border-radius: 4px;
  max-height: 400px;
  overflow-y: auto;
}
</style>
