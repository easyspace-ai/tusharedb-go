<script setup>
import { computed } from 'vue'
import { NTooltip, NPopover } from 'naive-ui'

const props = defineProps({
  // 情绪数据对象（来自后端API）
  sentiment: {
    type: Object,
    default: null
  },
  // 简单模式：只传value
  value: {
    type: Number,
    default: null
  },
  label: {
    type: String,
    default: '市场情绪'
  },
  loading: {
    type: Boolean,
    default: false
  }
})

// 获取情绪值
const sentimentValue = computed(() => {
  if (props.sentiment?.value !== undefined) {
    return props.sentiment.value
  }
  if (props.value !== null) {
    return props.value
  }
  return 50
})

// 计算指针角度 (0-100 映射到 -90度 到 90度)
const needleRotation = computed(() => {
  const clampedValue = Math.max(0, Math.min(100, sentimentValue.value))
  return (clampedValue - 50) * 1.8
})

// 情绪状态文字
const sentimentText = computed(() => {
  if (props.sentiment?.levelCn) {
    return props.sentiment.levelCn
  }
  const val = sentimentValue.value
  if (val < 20) return '极度恐慌'
  if (val < 40) return '恐慌'
  if (val < 60) return '中性'
  if (val < 80) return '贪婪'
  return '极度贪婪'
})

// 情绪颜色
const sentimentColor = computed(() => {
  const val = sentimentValue.value
  if (val < 20) return '#f5222d'
  if (val < 40) return '#fa8c16'
  if (val < 60) return '#fadb14'
  if (val < 80) return '#a0d911'
  return '#52c41a'
})

// 情绪描述
const sentimentDescription = computed(() => {
  if (props.sentiment?.description) {
    return props.sentiment.description
  }
  return ''
})

// 分项指标
const components = computed(() => {
  return props.sentiment?.components || []
})

// 更新时间
const updateTime = computed(() => {
  return props.sentiment?.updateTime || ''
})
</script>

<template>
  <div class="sentiment-gauge">
    <n-popover trigger="hover" placement="bottom" :disabled="components.length === 0">
      <template #trigger>
        <div class="gauge-container">
          <svg viewBox="0 0 200 120" class="gauge-svg">
            <!-- 背景弧线 -->
            <defs>
              <linearGradient id="gaugeGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" style="stop-color:#f5222d" />
                <stop offset="25%" style="stop-color:#fa8c16" />
                <stop offset="50%" style="stop-color:#fadb14" />
                <stop offset="75%" style="stop-color:#a0d911" />
                <stop offset="100%" style="stop-color:#52c41a" />
              </linearGradient>
            </defs>

            <!-- 外圈弧线 -->
            <path
              d="M 20 100 A 80 80 0 0 1 180 100"
              fill="none"
              stroke="url(#gaugeGradient)"
              stroke-width="8"
              stroke-linecap="round"
            />

            <!-- 刻度线 -->
            <g stroke="rgba(255,255,255,0.3)" stroke-width="1">
              <line x1="25" y1="95" x2="30" y2="90" />
              <line x1="40" y1="65" x2="47" y2="68" />
              <line x1="70" y1="40" x2="75" y2="47" />
              <line x1="100" y1="25" x2="100" y2="35" />
              <line x1="130" y1="40" x2="125" y2="47" />
              <line x1="160" y1="65" x2="153" y2="68" />
              <line x1="175" y1="95" x2="170" y2="90" />
            </g>

            <!-- 标签文字 -->
            <text x="25" y="108" fill="#f5222d" font-size="9" text-anchor="middle">恐慌</text>
            <text x="100" y="18" fill="#fadb14" font-size="9" text-anchor="middle">中性</text>
            <text x="175" y="108" fill="#52c41a" font-size="9" text-anchor="middle">贪婪</text>

            <!-- 指针 -->
            <g :style="{ transform: `rotate(${needleRotation}deg)`, transformOrigin: '100px 100px', transition: 'transform 0.5s ease-out' }">
              <line x1="100" y1="100" x2="100" y2="35" stroke="#fff" stroke-width="2" stroke-linecap="round" />
              <circle cx="100" cy="100" r="6" fill="#fff" />
              <circle cx="100" cy="100" r="3" :fill="sentimentColor" />
            </g>

            <!-- 中心数值 -->
            <text x="100" y="78" fill="#fff" font-size="22" font-weight="bold" text-anchor="middle">
              {{ sentimentValue.toFixed(0) }}
            </text>

            <!-- 情绪状态 -->
            <text x="100" y="94" :fill="sentimentColor" font-size="11" font-weight="bold" text-anchor="middle">
              {{ sentimentText }}
            </text>
          </svg>

          <!-- 底部标签 -->
          <div class="gauge-label">{{ label }}</div>
          <div v-if="updateTime" class="gauge-time">{{ updateTime }}</div>
        </div>
      </template>

      <!-- 详细指标弹窗 -->
      <div class="sentiment-details">
        <div class="details-title">情绪指标详情</div>
        <div v-if="sentimentDescription" class="details-desc">{{ sentimentDescription }}</div>
        <div class="details-list">
          <div v-for="comp in components" :key="comp.name" class="detail-item">
            <div class="detail-header">
              <span class="detail-name">{{ comp.nameCn }}</span>
              <span class="detail-weight">(权重{{ (comp.weight * 100).toFixed(0) }}%)</span>
            </div>
            <div class="detail-bar">
              <div class="detail-bar-fill" :style="{ width: comp.value + '%', backgroundColor: getBarColor(comp.value) }"></div>
            </div>
            <div class="detail-info">
              <span class="detail-value">{{ comp.value.toFixed(0) }}</span>
              <span class="detail-data">{{ comp.data }}</span>
            </div>
          </div>
        </div>
      </div>
    </n-popover>
  </div>
</template>

<script>
export default {
  methods: {
    getBarColor(value) {
      if (value < 20) return '#f5222d'
      if (value < 40) return '#fa8c16'
      if (value < 60) return '#fadb14'
      if (value < 80) return '#a0d911'
      return '#52c41a'
    }
  }
}
</script>

<style scoped>
.sentiment-gauge {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.gauge-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
}

.gauge-svg {
  width: 180px;
  height: 110px;
}

.gauge-label {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.gauge-time {
  font-size: 10px;
  color: #666;
}

/* 详情弹窗样式 */
.sentiment-details {
  min-width: 280px;
  padding: 8px;
}

.details-title {
  font-size: 14px;
  font-weight: bold;
  margin-bottom: 8px;
  color: #fff;
}

.details-desc {
  font-size: 12px;
  color: #999;
  margin-bottom: 12px;
  padding: 8px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
}

.details-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-item {
  padding: 8px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 4px;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 4px;
}

.detail-name {
  font-size: 12px;
  color: #fff;
}

.detail-weight {
  font-size: 10px;
  color: #666;
}

.detail-bar {
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 4px;
}

.detail-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.detail-info {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
}

.detail-value {
  color: #fff;
  font-weight: bold;
}

.detail-data {
  color: #999;
}
</style>
