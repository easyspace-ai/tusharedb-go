<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import {
  NModal,
  NCard,
  NButton,
  NSpace,
  NTag,
  NAlert,
  NProgress,
  useMessage
} from 'naive-ui'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import {
  CheckUpdate,
  OpenURL,
  DownloadAndInstallUpdate,
  SkipUpdateVersion
} from '../../wailsjs/go/main/App'

const MANUAL_UPDATE_EVENT = 'stock-ai:show-update-dialog'

const message = useMessage()
const showModal = ref(false)
const updateInfo = ref(null)
const autoUpdating = ref(false)
const skipping = ref(false)
const progress = ref(0)
const progressMessage = ref('')
const updateStatus = ref('')
const hasError = ref(false)
const downloadSpeed = ref(0)
const etaSeconds = ref(-1)
const eventUnbinders = []

const progressStatus = computed(() => {
  if (hasError.value) return 'error'
  if (updateStatus.value === 'restarting') return 'success'
  return 'default'
})

const formattedSpeed = computed(() => {
  if (downloadSpeed.value > 0) {
    return `${downloadSpeed.value.toFixed(2)} MB/s`
  }
  return '计算中...'
})

const formattedEta = computed(() => {
  if (etaSeconds.value < 0) {
    return '估算中...'
  }
  const total = Math.round(etaSeconds.value)
  const minutes = Math.floor(total / 60)
  const seconds = total % 60
  if (minutes > 0) {
    return `${minutes}分${seconds}秒`
  }
  return `${seconds}秒`
})

const openUpdateModal = (info) => {
  updateInfo.value = info
  showModal.value = true
  autoUpdating.value = false
  skipping.value = false
  hasError.value = false
  progress.value = 0
  progressMessage.value = ''
  updateStatus.value = ''
  downloadSpeed.value = 0
  etaSeconds.value = -1
}

const manualHandler = (event) => {
  const info = event?.detail
  if (info?.hasUpdate) {
    openUpdateModal(info)
  } else if (info?.skipped && info?.skipVersion) {
    message.info(`您已选择跳过版本 ${info.skipVersion}`)
  }
}

const startAutoUpdate = async () => {
  if (autoUpdating.value || !updateInfo.value?.hasUpdate) return
  autoUpdating.value = true
  hasError.value = false
  progress.value = 0
  progressMessage.value = '正在准备下载更新...'
  updateStatus.value = 'downloading'

  try {
    await DownloadAndInstallUpdate()
  } catch (err) {
    hasError.value = true
    const errMsg = err?.message || err
    progressMessage.value = `自动更新失败: ${errMsg}`
    updateStatus.value = 'error'
    autoUpdating.value = false
  }
}

const bindUpdateEvents = () => {
  eventUnbinders.push(
    EventsOn('update:progress', (payload) => {
      if (typeof payload?.percent === 'number') {
        progress.value = Math.min(100, Math.max(0, payload.percent))
      }
      if (payload?.message) {
        progressMessage.value = payload.message
      }
      if (typeof payload?.speed === 'number') {
        downloadSpeed.value = payload.speed
      }
      if (typeof payload?.etaSeconds === 'number') {
        etaSeconds.value = payload.etaSeconds
      } else {
        etaSeconds.value = -1
      }
    })
  )

  eventUnbinders.push(
    EventsOn('update:status', (payload) => {
      if (payload?.status) {
        updateStatus.value = payload.status
      }
      if (payload?.message) {
        progressMessage.value = payload.message
      }
      if (payload?.status === 'error') {
        hasError.value = true
        autoUpdating.value = false
      }
    })
  )
}

const checkForUpdate = async () => {
  try {
    const info = await CheckUpdate()
    if (info && info.hasUpdate) {
      openUpdateModal(info)
    }
  } catch (e) {
    console.error('检查更新失败:', e)
  }
}

const openDownload = () => {
  if (updateInfo.value?.downloadUrl) {
    OpenURL(updateInfo.value.downloadUrl)
  }
}

const openRelease = () => {
  if (updateInfo.value?.releaseUrl) {
    OpenURL(updateInfo.value.releaseUrl)
  }
}

const ignoreThisTime = () => {
  if (!autoUpdating.value) {
    showModal.value = false
  }
}

const skipThisVersion = async () => {
  if (!updateInfo.value?.version || autoUpdating.value || skipping.value) return
  skipping.value = true
  try {
    await SkipUpdateVersion(updateInfo.value.version)
    message.success(`已跳过版本 ${updateInfo.value.version}`)
    showModal.value = false
  } catch (err) {
    const errMsg = err?.message || err
    message.error(`跳过版本失败: ${errMsg}`)
  } finally {
    skipping.value = false
  }
}

onMounted(() => {
  bindUpdateEvents()
  window.addEventListener(MANUAL_UPDATE_EVENT, manualHandler)
  // 延迟 2 秒检查更新，避免影响启动速度
  setTimeout(checkForUpdate, 2000)
})

onUnmounted(() => {
  window.removeEventListener(MANUAL_UPDATE_EVENT, manualHandler)
  eventUnbinders.forEach((off) => {
    if (typeof off === 'function') {
      off()
    }
  })
})
</script>

<template>
  <n-modal v-model:show="showModal" :mask-closable="false">
    <n-card
      title="发现新版本"
      :bordered="false"
      style="width: 480px;"
      role="dialog"
    >
      <template #header-extra>
        <n-tag type="success">{{ updateInfo?.version }}</n-tag>
      </template>

      <div class="update-content">
        <div class="version-compare">
          <span class="current">当前版本: {{ updateInfo?.currentVersion }}</span>
          <span class="arrow">→</span>
          <span class="new">最新版本: {{ updateInfo?.version }}</span>
        </div>

        <div class="auto-update" v-if="autoUpdating || hasError">
          <n-progress
            v-if="autoUpdating"
            type="line"
            :percentage="Math.round(progress)"
            :status="progressStatus"
            :show-indicator="false"
          />
          <p class="progress-text">
            {{ progressMessage || (autoUpdating ? '正在准备更新...' : '更新状态已停止') }}
          </p>
          <div class="progress-meta" v-if="autoUpdating">
            <span>速度: {{ formattedSpeed }}</span>
            <span>预计剩余: {{ formattedEta }}</span>
          </div>
          <n-alert v-if="hasError" type="error" style="margin-top: 12px;">
            {{ progressMessage || '自动更新失败，请稍后重试' }}
            <template #action>
              <n-button text type="error" @click="startAutoUpdate">重试更新</n-button>
            </template>
          </n-alert>
        </div>

        <n-alert type="info" style="margin: 16px 0;" v-if="updateInfo?.releaseDate">
          发布日期: {{ updateInfo.releaseDate }}
        </n-alert>

        <div class="description" v-if="updateInfo?.description">
          <h4>更新内容:</h4>
          <p>{{ updateInfo.description }}</p>
        </div>
      </div>

      <template #footer>
        <div class="footer-actions">
          <n-space>
            <n-button quaternary @click="ignoreThisTime" :disabled="autoUpdating">
              忽视本次提醒
            </n-button>
            <n-button
              quaternary
              @click="skipThisVersion"
              :disabled="autoUpdating"
              :loading="skipping"
            >
              跳过此版本
            </n-button>
          </n-space>
          <n-space justify="end">
            <n-button @click="openRelease" :disabled="autoUpdating" v-if="updateInfo?.releaseUrl">
              查看详情
            </n-button>
            <n-button @click="openDownload" :disabled="autoUpdating">
              手动下载
            </n-button>
            <n-button type="primary" @click="startAutoUpdate" :loading="autoUpdating">
              立即更新
            </n-button>
          </n-space>
        </div>
      </template>
    </n-card>
  </n-modal>
</template>

<style scoped>
.update-content {
  padding: 8px 0;
}

.version-compare {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  margin-bottom: 16px;
}

.version-compare .current {
  color: #999;
}

.version-compare .arrow {
  color: #18a058;
  font-size: 20px;
}

.version-compare .new {
  color: #18a058;
  font-weight: 600;
}

.auto-update {
  margin-bottom: 16px;
}

.progress-text {
  margin-top: 8px;
  color: #fff;
  font-size: 14px;
}

.progress-meta {
  margin-top: 8px;
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #ccc;
}

.description h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #fff;
}

.description p {
  margin: 0;
  color: #999;
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
}

.footer-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}
</style>
