<script setup>
import { ref, onMounted } from 'vue'
import {
  NCard,
  NSpace,
  NButton,
  NTag,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NTimeline,
  NTimelineItem,
  useMessage
} from 'naive-ui'
import { GetVersion, CheckUpdate, OpenURL } from '../../wailsjs/go/main/App'

const message = useMessage()
const version = ref('1.0.0')
const buildTime = ref('')
const hasUpdate = ref(false)
const latestVersion = ref('')
const checking = ref(false)
const MANUAL_UPDATE_EVENT = 'stock-ai:show-update-dialog'

const changelog = [
  { version: 'v1.0.0', date: '2025-11-15', content: '首个正式版本发布', type: 'success' },
]

const features = [
  { title: '实时行情', desc: '支持A股、港股、美股实时行情获取，数据来源新浪/腾讯财经' },
  { title: '自选管理', desc: '添加自选股票和基金，支持分组管理和排序' },
  { title: '市场概览', desc: '主要指数、行业排行、资金流向、龙虎榜一览' },
  { title: '财经快讯', desc: '财联社实时快讯推送，重要消息高亮显示' },
  { title: '研报公告', desc: '个股研报和公告查询，掌握最新动态' },
  { title: 'K线图表', desc: '日K/周K/月K线图表展示，支持分时走势' },
  { title: 'AI分析', desc: '接入DeepSeek/OpenAI/Ollama等AI模型，智能分析股票' },
  { title: '本地存储', desc: 'SQLite本地数据库，数据安全不上传' }
]

const loadVersion = async () => {
  try {
    const info = await GetVersion()
    if (info) {
      version.value = info.version || '1.0.0'
      buildTime.value = info.buildTime || ''
    }
  } catch (e) {
    console.error('获取版本信息失败:', e)
  }
}

const checkUpdate = async () => {
  checking.value = true
  try {
    const result = await CheckUpdate()
    if (result && result.hasUpdate) {
      hasUpdate.value = true
      latestVersion.value = result.version
      message.info('发现新版本: ' + result.version + '，请在弹窗中选择是否更新')
      window.dispatchEvent(new CustomEvent(MANUAL_UPDATE_EVENT, { detail: result }))
    } else if (result?.skipped && result?.skipVersion) {
      message.info('您已选择跳过版本 ' + result.skipVersion + '，等待新版本发布')
    } else {
      message.success('当前已是最新版本')
    }
  } catch (e) {
    message.error('检查更新失败: ' + e)
  } finally {
    checking.value = false
  }
}

const openQQ = () => {
  OpenURL('tencent://message/?uin=3946808002&Site=qq&Menu=yes')
}

onMounted(() => {
  loadVersion()
})
</script>

<template>
  <div class="about-page">
    <n-card :bordered="false">
      <n-space vertical size="large">
        <!-- Logo和标题 -->
        <div class="app-header">
          <div class="app-logo">S</div>
          <div class="app-title">
            <h1>Stock AI</h1>
            <p>智能股票分析助手</p>
          </div>
        </div>

        <!-- 版本信息 -->
        <n-descriptions label-placement="left" :column="2" bordered>
          <n-descriptions-item label="当前版本">
            <n-tag type="info" size="large">v{{ version }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="构建时间">
            {{ buildTime || '2025-11-15' }}
          </n-descriptions-item>
          <n-descriptions-item label="技术栈">
            Go 1.23 + Wails v2 + Vue3 + Naive UI
          </n-descriptions-item>
          <n-descriptions-item label="数据存储">
            SQLite (WAL模式)
          </n-descriptions-item>
          <n-descriptions-item label="数据来源">
            新浪财经 / 腾讯财经 / 东方财富 / 财联社
          </n-descriptions-item>
          <n-descriptions-item label="运行平台">
            Windows 10/11
          </n-descriptions-item>
        </n-descriptions>

        <!-- 操作按钮 -->
        <n-space>
          <n-button type="primary" :loading="checking" @click="checkUpdate">检查更新</n-button>
        </n-space>

        <div v-if="hasUpdate" class="update-info">
          <n-tag type="warning" size="large">
            发现新版本: {{ latestVersion }}，请在弹窗中选择更新策略
          </n-tag>
        </div>

        <n-divider title-placement="left">功能特性</n-divider>

        <!-- 功能列表 -->
        <div class="feature-grid">
          <div v-for="(feature, index) in features" :key="index" class="feature-item">
            <div class="feature-title">{{ feature.title }}</div>
            <div class="feature-desc">{{ feature.desc }}</div>
          </div>
        </div>

        <n-divider title-placement="left">更新日志</n-divider>

        <!-- 更新日志 -->
        <n-timeline>
          <n-timeline-item
            v-for="log in changelog"
            :key="log.version"
            :type="log.type"
            :title="log.version"
            :time="log.date"
          >
            {{ log.content }}
          </n-timeline-item>
        </n-timeline>

        <n-divider title-placement="left">免责声明</n-divider>

        <div class="disclaimer">
          <n-alert type="warning" :bordered="false">
            <template #header>重要声明</template>
            <ol>
              <li><strong>非投资建议</strong>：本软件提供的所有信息、数据、分析结果仅供学习研究和参考，不构成任何投资建议、投资咨询或证券推荐。</li>
              <li><strong>无资质声明</strong>：本软件及其开发者不具备证券投资咨询业务资格，不提供任何形式的证券投资咨询服务。</li>
              <li><strong>AI 生成内容</strong>：软件中的 AI 分析结果由人工智能模型生成，可能存在错误或偏差，仅供参考，不应作为投资决策的依据。</li>
              <li><strong>数据来源</strong>：本软件数据来源于公开渠道，不保证数据的准确性、完整性和及时性。</li>
              <li><strong>投资风险</strong>：股市有风险，投资需谨慎。任何投资决策应基于您自己的独立判断，并建议咨询专业的持牌投资顾问。</li>
              <li><strong>免责条款</strong>：使用本软件造成的任何直接或间接损失，开发者不承担任何责任。</li>
            </ol>
          </n-alert>
        </div>

        <n-divider />

        <n-divider title-placement="left">联系开发者</n-divider>
        <div class="contact-info">
          <div class="contact-text">
            QQ 咨询： 
            <a href="javascript:;" @click.prevent="openQQ">3946808002</a>
            （点击自动拉起 QQ）
          </div>
          <n-button type="success" tertiary @click="openQQ">打开 QQ 联系</n-button>
        </div>
        <p class="contact-warning">
          <span>仅供个人学习自用，任何售卖、商业分发或收费服务前请务必取得作者书面许可，违者将依法追责。</span>
        </p>

        <div class="copyright">
          <p>Copyright 2025 Stock AI. 基于 MIT 协议开源。</p>
          <p style="color: #f0a020; font-weight: bold;">本软件仅供学习研究使用，不构成任何投资建议。投资有风险，入市需谨慎。</p>
        </div>
      </n-space>
    </n-card>
  </div>
</template>

<style scoped>
.about-page {
  max-width: 900px;
}

.app-header {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 20px 0;
}

.app-logo {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #18a058, #36ad6a);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40px;
  font-weight: bold;
  color: white;
}

.app-title h1 {
  font-size: 28px;
  color: #18a058;
  margin: 0 0 8px 0;
}

.app-title p {
  color: #999;
  margin: 0;
  font-size: 14px;
}

.update-info {
  padding: 10px 0;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.feature-item {
  padding: 16px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
}

.feature-title {
  font-size: 15px;
  font-weight: bold;
  margin-bottom: 8px;
  color: #18a058;
}

.feature-desc {
  font-size: 13px;
  color: #999;
  line-height: 1.5;
}

.contact-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  padding-bottom: 12px;
}

.contact-text a {
  color: #18a058;
  font-weight: 600;
}

.contact-warning {
  color: #ff4d4f;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
}

.copyright {
  text-align: center;
  color: #666;
  font-size: 12px;
}

.copyright p {
  margin: 4px 0;
}

.disclaimer {
  margin-bottom: 16px;
}

.disclaimer ol {
  margin: 8px 0 0 0;
  padding-left: 20px;
}

.disclaimer li {
  margin-bottom: 8px;
  line-height: 1.6;
}
</style>
