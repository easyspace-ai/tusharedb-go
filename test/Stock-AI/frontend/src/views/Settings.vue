<script setup>
import { ref, onMounted } from 'vue'
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSwitch,
  NButton,
  NSpace,
  NDivider,
  NSelect,
  NAlert,
  NCollapse,
  NCollapseItem,
  NColorPicker,
  useMessage
} from 'naive-ui'
import { GetConfig, SaveConfig, GetDataPipelineStatus, TestAlertPush } from '../../wailsjs/go/main/App'

const message = useMessage()
const loading = ref(false)

const config = ref({
  id: 1,
  refreshInterval: 15,
  proxyUrl: '',
  proxyPoolEnabled: false,
  proxyProvider: 'kuaidaili',
  proxyApiUrl: '',
  proxyApiKey: '',
  proxyApiSecret: '',
  proxyRegion: '',
  proxyPoolList: '',
  proxyPoolProtocol: 'http',
  proxyPoolTTL: 60,
  proxyPoolSize: 5,
  aiEnabled: false,
  aiModel: 'deepseek',
  aiApiKey: '',
  aiApiUrl: '',
  browserPath: '',
  // 付费API配置
  paidApiEnabled: false,
  paidApiProvider: '',
  paidApiKey: '',
  paidApiSecret: '',
  paidApiUrl: '',
  // Tushare配置
  tushareToken: '',
  tushareEnabled: false,
  // AKShare配置
  akshareEnabled: false,
  // 数据源优先级
  dataSourcePriority: 'tushare',
  theme: 'dark',
  customPrimary: '#18a058',
  alertPushEnabled: false,
  wecomWebhook: '',
  dingtalkWebhook: '',
  emailPushEnabled: false,
  emailSmtp: '',
  emailPort: 465,
  emailUser: '',
  emailPassword: '',
  emailTo: ''
})

const aiModelOptions = [
  { label: 'DeepSeek', value: 'deepseek' },
  { label: 'OpenAI (GPT-4)', value: 'openai' },
  { label: 'Ollama (本地部署)', value: 'ollama' },
  { label: '硅基流动', value: 'siliconflow' },
  { label: '阿里云百炼', value: 'aliyun' },
  { label: '火山方舟', value: 'volcengine' }
]

const paidApiOptions = [
  { label: '东方财富开放平台', value: 'eastmoney' },
  { label: '同花顺iFinD', value: 'ths' },
  { label: '聚合数据', value: 'juhe' },
  { label: '万得Wind', value: 'wind' },
  { label: '自定义API', value: 'custom' }
]

const dataSourcePriorityOptions = [
  { label: 'Tushare 优先', value: 'tushare' },
  { label: 'AKShare 优先', value: 'akshare' }
]

const themeOptions = [
  { label: '深色模式', value: 'dark' },
  { label: '浅色模式', value: 'light' },
  { label: '跟随系统', value: 'system' }
]

const proxyProviderOptions = [
  { label: '快代理（Kuaidaili）', value: 'kuaidaili' },
  { label: '青果网络（Qingguo）', value: 'qingguo' },
  { label: '通用 HTTP API', value: 'custom_api' },
  { label: '自定义列表', value: 'custom_list' }
]

const proxyProtocolOptions = [
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
  { label: 'SOCKS5', value: 'socks5' }
]

const pipelineStatus = ref(null)
const statusLoading = ref(false)

const loadPipelineStatus = async () => {
  statusLoading.value = true
  try {
    const status = await GetDataPipelineStatus()
    pipelineStatus.value = status
  } catch (e) {
    console.error('加载数据源状态失败:', e)
  } finally {
    statusLoading.value = false
  }
}

const testPush = async (channel, label) => {
  try {
    await TestAlertPush(channel)
    message.success(`${label} 推送测试成功`)
  } catch (e) {
    message.error(`${label} 推送失败: ${e}`)
  }
}

const statusTagType = (status) => {
  if (status === 'healthy') return 'success'
  if (status === 'degraded') return 'warning'
  if (status === 'disabled') return 'error'
  return 'default'
}

const statusText = (status) => {
  if (status === 'healthy') return '正常'
  if (status === 'degraded') return '关注'
  if (status === 'disabled') return '暂停'
  return '未知'
}

const loadConfig = async () => {
  try {
    const data = await GetConfig()
    if (data) {
      config.value = { ...config.value, ...data }
    }
  } catch (e) {
    console.error('加载配置失败:', e)
  }
}

const saveConfig = async () => {
  loading.value = true
  try {
    await SaveConfig(config.value)
    message.success('保存成功')
    window.dispatchEvent(
      new CustomEvent('stock-ai:theme-updated', {
        detail: {
          theme: config.value.theme,
          customPrimary: config.value.customPrimary
        }
      })
    )
  } catch (e) {
    message.error('保存失败: ' + e)
  } finally {
    loading.value = false
  }
}

const resetConfig = () => {
  config.value = {
    id: config.value.id,
    refreshInterval: 15,
    proxyUrl: '',
    proxyPoolEnabled: false,
    proxyProvider: 'kuaidaili',
    proxyApiUrl: '',
    proxyApiKey: '',
    proxyApiSecret: '',
    proxyRegion: '',
    proxyPoolList: '',
    proxyPoolProtocol: 'http',
    proxyPoolTTL: 60,
    proxyPoolSize: 5,
    aiEnabled: false,
    aiModel: 'deepseek',
    aiApiKey: '',
    aiApiUrl: '',
    browserPath: '',
    paidApiEnabled: false,
    paidApiProvider: '',
    paidApiKey: '',
    paidApiSecret: '',
    paidApiUrl: '',
    tushareToken: '',
    tushareEnabled: false,
    akshareEnabled: false,
    dataSourcePriority: 'tushare',
    theme: 'dark',
    customPrimary: '#18a058',
    alertPushEnabled: false,
    wecomWebhook: '',
    dingtalkWebhook: '',
    emailPushEnabled: false,
    emailSmtp: '',
    emailPort: 465,
    emailUser: '',
    emailPassword: '',
    emailTo: ''
  }
  message.info('已重置为默认值，请点击保存')
}

onMounted(() => {
  loadConfig()
  loadPipelineStatus()
})
</script>

<template>
  <div class="settings-page">
    <n-card title="系统设置" :bordered="false">
      <div class="status-wrapper">
        <div class="status-header">
          <div>
            <strong>数据通道监控</strong>
            <span class="status-update">最后更新：{{ pipelineStatus?.generatedAt || '加载中...' }}</span>
          </div>
          <n-button size="small" :loading="statusLoading" @click="loadPipelineStatus">刷新</n-button>
        </div>
        <div class="status-grid" v-if="pipelineStatus">
          <div class="status-column">
            <h4>行情数据源</h4>
            <div v-if="pipelineStatus.marketSources?.length">
              <div v-for="source in pipelineStatus.marketSources" :key="source.key" class="status-item">
                <div class="status-item-header">
                  <span>{{ source.name }}</span>
                  <n-tag size="small" :type="statusTagType(source.status)">{{ statusText(source.status) }}</n-tag>
                </div>
                <div class="status-item-meta">
                  <span>响应：{{ source.latencyLabel }}</span>
                  <span>最近成功：{{ source.lastSuccess || '未知' }}</span>
                </div>
              </div>
            </div>
            <div v-else class="status-empty">暂无数据</div>
          </div>
          <div class="status-column">
            <h4>财务数据</h4>
            <div v-if="pipelineStatus.financial">
              <div
                v-for="(status, key) in pipelineStatus.financial"
                :key="key"
                class="status-item"
              >
                <div class="status-item-header">
                  <span>{{ key.toUpperCase() }}</span>
                  <n-tag
                    size="small"
                    :type="status?.available ? 'success' : status?.isDisabled ? 'error' : 'warning'"
                  >
                    {{ status?.available ? '可用' : status?.isDisabled ? '暂停' : '未启用' }}
                  </n-tag>
                </div>
                <div class="status-item-meta">
                  <span>失败次数：{{ status?.failCount ?? 0 }}</span>
                  <span v-if="status?.disabledUntil">恢复：{{ status.disabledUntil }}</span>
                </div>
              </div>
            </div>
            <div v-else class="status-empty">暂无数据</div>
          </div>
          <div class="status-column">
            <h4>代理池</h4>
            <div v-if="pipelineStatus.proxy">
              <div class="status-item">
                <div class="status-item-header">
                  <span>代理状态</span>
                  <n-tag
                    size="small"
                    :type="pipelineStatus.proxy.activeProxies > 0 ? 'success' : 'warning'"
                  >
                    {{ pipelineStatus.proxy.activeProxies > 0 ? '已缓存' : '等待刷新' }}
                  </n-tag>
                </div>
                <div class="status-item-meta">
                  <span>启用：{{ pipelineStatus.proxy.enabled ? '是' : '否' }}</span>
                  <span>数量：{{ pipelineStatus.proxy.activeProxies }}</span>
                </div>
                <div class="status-item-meta">
                  <span>供应商：{{ pipelineStatus.proxy.provider || '未配置' }}</span>
                  <span>到期：{{ pipelineStatus.proxy.expiresAt || '未知' }}</span>
                </div>
                <div class="status-item-meta">
                  <span>最后获取：{{ pipelineStatus.proxy.lastFetch || '未知' }}</span>
                  <span v-if="pipelineStatus.proxy.lastError">错误：{{ pipelineStatus.proxy.lastError }}</span>
                </div>
              </div>
            </div>
            <div v-else class="status-empty">暂无数据</div>
          </div>
        </div>
      </div>
      <n-form label-placement="left" label-width="140">
        <n-divider title-placement="left">基础设置</n-divider>

        <n-form-item label="行情刷新间隔">
          <n-input-number v-model:value="config.refreshInterval" :min="5" :max="120" style="width: 200px;">
            <template #suffix>秒</template>
          </n-input-number>
          <span style="margin-left: 12px; color: #999;">建议15秒以上，避免请求过于频繁</span>
        </n-form-item>

        <n-form-item label="HTTP代理">
          <n-input v-model:value="config.proxyUrl" placeholder="如 http://127.0.0.1:7890 或 socks5://127.0.0.1:1080" style="width: 400px;" />
        </n-form-item>

        <n-divider title-placement="left">代理池（快代理 / 青果网络 / 通用API）</n-divider>

        <n-alert type="info" style="margin-bottom: 16px;">
          <p>支持快代理、青果网络等主流代理商，也可使用任意HTTP API或自定义代理列表。</p>
          <p>启用后系统会自动轮询多个代理，自动限流并记录日志（位于 %USERPROFILE%\.stock-ai\logs\update.log）。</p>
        </n-alert>

        <n-form-item label="启用代理池">
          <n-switch v-model:value="config.proxyPoolEnabled" />
          <span style="margin-left: 12px; color: #999;">提升访问稳定性，建议与HTTP代理互为备份</span>
        </n-form-item>

        <template v-if="config.proxyPoolEnabled">
          <n-form-item label="服务商">
            <n-select v-model:value="config.proxyProvider" :options="proxyProviderOptions" style="width: 260px;" />
          </n-form-item>

          <template v-if="config.proxyProvider === 'custom_list'">
            <n-form-item label="代理列表">
              <n-input
                type="textarea"
                v-model:value="config.proxyPoolList"
                :rows="4"
                style="width: 520px;"
                placeholder="每行一个，如 1.1.1.1:8000 或 socks5://1.1.1.1:8000"
              />
            </n-form-item>
            <n-form-item label="协议">
              <n-select v-model:value="config.proxyPoolProtocol" :options="proxyProtocolOptions" style="width: 200px;" />
              <span style="margin-left: 12px; color: #999;">未带协议的地址会自动补全为该协议</span>
            </n-form-item>
          </template>
          <template v-else>
            <n-form-item label="API 地址">
              <n-input
                v-model:value="config.proxyApiUrl"
                style="width: 520px;"
                placeholder="留空使用内置模板，支持 {apiKey}、{apiSecret}、{num}、{region} 占位符"
              />
            </n-form-item>
            <n-form-item label="API Key / Secret ID">
              <n-input v-model:value="config.proxyApiKey" style="width: 400px;" placeholder="快代理 SecretId / 青果 Key" />
            </n-form-item>
            <n-form-item label="API Secret / Token">
              <n-input v-model:value="config.proxyApiSecret" style="width: 400px;" placeholder="快代理 Signature / 青果 Secret" />
            </n-form-item>
            <n-form-item label="线路地区（可选）">
              <n-input v-model:value="config.proxyRegion" style="width: 260px;" placeholder="如 cn、hk、us 等" />
            </n-form-item>
            <n-form-item label="每次拉取数量">
              <n-input-number v-model:value="config.proxyPoolSize" :min="1" :max="50" style="width: 160px;" />
              <span style="margin-left: 12px; color: #999;">系统会轮询这些代理，建议 5~10 个</span>
            </n-form-item>
            <n-form-item label="代理生命周期">
              <n-input-number v-model:value="config.proxyPoolTTL" :min="30" :max="600" style="width: 160px;">
                <template #suffix>秒</template>
              </n-input-number>
              <span style="margin-left: 12px; color: #999;">超过此时间自动重新拉取新的代理</span>
            </n-form-item>
            <n-form-item label="代理协议">
              <n-select v-model:value="config.proxyPoolProtocol" :options="proxyProtocolOptions" style="width: 200px;" />
            </n-form-item>
            <n-alert type="success" style="margin-bottom: 16px;">
              <p><strong>快代理示例：</strong>SecretId/Signature 后可直接留空 API 地址，系统将使用 <code>https://dps.kdlapi.com/getdps</code> 模板。</p>
              <p><strong>青果网络示例：</strong>填写 Key/Secret，留空 API 地址，系统将使用 <code>https://proxy.qg.net/allocate</code> 模板。</p>
              <p>其他供应商可填写自定义 API 地址，例如 <code>https://api.xxx.com/get?token={apiKey}&amp;num={num}</code>。</p>
            </n-alert>
          </template>
        </template>

        <n-form-item label="浏览器路径">
          <n-input v-model:value="config.browserPath" placeholder="用于抓取网页内容，留空自动检测Edge/Chrome/Firefox/QQ/360/夸克等浏览器" style="width: 400px;" />
        </n-form-item>

        <n-alert type="info" style="margin-bottom: 16px;">
          <p>1. 行情数据仅在交易时间（周一至周五 9:30-11:30, 13:00-15:00）自动刷新</p>
          <p>2. 配置HTTP代理可以提高数据获取的稳定性</p>
          <p>3. 浏览器路径用于抓取部分需要JavaScript渲染的页面数据</p>
        </n-alert>

        <n-divider title-placement="left">界面外观</n-divider>

        <n-form-item label="主题模式">
          <n-select v-model:value="config.theme" :options="themeOptions" style="width: 220px;" />
        </n-form-item>

        <n-form-item label="主色调">
          <n-color-picker v-model:value="config.customPrimary" :show-alpha="false" modes="hex" />
        </n-form-item>

        <n-divider title-placement="left">财务数据源配置（Tushare / AKShare）</n-divider>

        <n-alert type="info" style="margin-bottom: 16px;">
          <p>Tushare 和 AKShare 是两个免费的财务数据源，用于获取股票的财务指标、资产负债表、现金流等数据。</p>
          <p>系统会自动在两个数据源之间切换，确保数据获取的稳定性。</p>
        </n-alert>

        <n-form-item label="启用 Tushare">
          <n-switch v-model:value="config.tushareEnabled" />
          <span style="margin-left: 12px; color: #999;">Tushare Pro 提供专业的财务数据</span>
        </n-form-item>

        <template v-if="config.tushareEnabled">
          <n-form-item label="Tushare Token">
            <n-input v-model:value="config.tushareToken" type="password" show-password-on="click" placeholder="访问 tushare.pro 注册获取Token" style="width: 400px;" />
          </n-form-item>
        </template>

        <n-form-item label="启用 AKShare">
          <n-switch v-model:value="config.akshareEnabled" />
          <span style="margin-left: 12px; color: #999;">AKShare 开源免费，无需Token（需要Python环境）</span>
        </n-form-item>

        <template v-if="config.tushareEnabled || config.akshareEnabled">
          <n-form-item label="数据源优先级">
            <n-select v-model:value="config.dataSourcePriority" :options="dataSourcePriorityOptions" style="width: 200px;" />
            <span style="margin-left: 12px; color: #999;">优先使用哪个数据源获取财务数据</span>
          </n-form-item>
        </template>

        <n-collapse style="margin-bottom: 16px;">
          <n-collapse-item title="Tushare / AKShare 配置说明" name="financial-guide">
            <div class="api-guide">
              <p><strong>Tushare Pro：</strong></p>
              <p>1. 访问 <a href="https://tushare.pro" target="_blank">tushare.pro</a> 注册账号</p>
              <p>2. 完成实名认证后获取 Token</p>
              <p>3. 免费用户每分钟可调用 60 次，足够日常使用</p>
              <p style="margin-top: 8px;"><strong>AKShare：</strong></p>
              <p>1. 需要本地安装 Python 3.7+ 环境</p>
              <p>2. 首次使用会自动安装 akshare 库</p>
              <p>3. 完全免费，但请求频率需要控制</p>
              <p style="margin-top: 8px; color: #18a058;"><strong>防封禁说明：</strong>系统已内置智能限流机制，会自动控制请求频率、添加随机延迟，最大程度保护您的IP不被封禁。</p>
            </div>
          </n-collapse-item>
        </n-collapse>

        <n-divider title-placement="left">付费数据API（可选）</n-divider>

        <n-form-item label="启用付费API">
          <n-switch v-model:value="config.paidApiEnabled" />
          <span style="margin-left: 12px; color: #999;">使用付费API可获得更稳定、更全面的数据</span>
        </n-form-item>

        <template v-if="config.paidApiEnabled">
          <n-form-item label="数据服务商">
            <n-select v-model:value="config.paidApiProvider" :options="paidApiOptions" placeholder="选择数据服务商" style="width: 300px;" />
          </n-form-item>

          <n-form-item label="API Key / Token">
            <n-input v-model:value="config.paidApiKey" type="password" show-password-on="click" placeholder="填写API Key或Token" style="width: 400px;" />
          </n-form-item>

          <n-form-item label="API Secret (可选)">
            <n-input v-model:value="config.paidApiSecret" type="password" show-password-on="click" placeholder="部分服务需要Secret" style="width: 400px;" />
          </n-form-item>

          <n-form-item label="API URL (可选)">
            <n-input v-model:value="config.paidApiUrl" placeholder="自定义API地址，留空使用默认地址" style="width: 400px;" />
          </n-form-item>

          <n-collapse>
            <n-collapse-item title="各服务商申请指南" name="guide">
              <div class="api-guide">
                <p><strong>东方财富开放平台：</strong>访问 data.eastmoney.com 注册申请</p>
                <p><strong>同花顺iFinD：</strong>访问 ft.10jqka.com.cn 申请机构账号</p>
                <p><strong>Tushare Pro：</strong>访问 tushare.pro 注册获取Token（有免费额度）</p>
                <p><strong>AKShare：</strong>开源免费，无需API Key，但建议配合代理使用</p>
                <p><strong>聚合数据：</strong>访问 juhe.cn 注册申请股票数据API</p>
                <p><strong>万得Wind：</strong>需要购买Wind终端或API服务</p>
              </div>
            </n-collapse-item>
          </n-collapse>

          <n-alert type="success" style="margin: 16px 0;">
            使用付费API的优势：请求频率限制更宽松、数据更全面准确、不易被封禁IP
          </n-alert>
        </template>

        <n-divider title-placement="left">AI 智能分析</n-divider>

        <n-form-item label="启用AI分析">
          <n-switch v-model:value="config.aiEnabled" />
          <span style="margin-left: 12px; color: #999;">开启后可使用AI分析股票</span>
        </n-form-item>

        <template v-if="config.aiEnabled">
          <n-form-item label="AI模型">
            <n-select v-model:value="config.aiModel" :options="aiModelOptions" style="width: 300px;" />
          </n-form-item>

          <n-form-item label="API Key">
            <n-input v-model:value="config.aiApiKey" type="password" show-password-on="click" placeholder="填写对应AI服务的API Key" style="width: 400px;" />
          </n-form-item>

          <n-form-item label="API URL (可选)">
            <n-input v-model:value="config.aiApiUrl" placeholder="自定义API地址，留空使用默认地址" style="width: 400px;" />
          </n-form-item>
        </template>

        <n-alert type="warning" style="margin-bottom: 16px;">
          所有API Key和密钥均存储在本地数据库中，不会上传到任何服务器。请妥善保管您的密钥。
        </n-alert>

        <n-divider title-placement="left">告警推送</n-divider>

        <n-form-item label="开启推送">
          <n-switch v-model:value="config.alertPushEnabled" />
          <span style="margin-left: 12px; color: #999;">触发价格提醒时自动推送至企业微信、钉钉或邮箱</span>
        </n-form-item>

        <div v-if="config.alertPushEnabled" class="push-grid">
          <n-card title="企业微信机器人" size="small">
            <n-form-item label="Webhook 地址">
              <n-input v-model:value="config.wecomWebhook" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..." />
            </n-form-item>
            <n-button size="small" @click="testPush('wecom', '企业微信')">测试推送</n-button>
          </n-card>

          <n-card title="钉钉机器人" size="small">
            <n-form-item label="Webhook 地址">
              <n-input v-model:value="config.dingtalkWebhook" placeholder="https://oapi.dingtalk.com/robot/send?access_token=..." />
            </n-form-item>
            <n-button size="small" @click="testPush('dingtalk', '钉钉')">测试推送</n-button>
          </n-card>

          <n-card title="邮件通知" size="small">
            <n-form-item label="启用邮件">
              <n-switch v-model:value="config.emailPushEnabled" />
            </n-form-item>
            <template v-if="config.emailPushEnabled">
              <n-form-item label="SMTP服务器">
                <n-input v-model:value="config.emailSmtp" placeholder="smtp.qq.com" />
              </n-form-item>
              <n-form-item label="端口">
                <n-input-number v-model:value="config.emailPort" :min="1" :max="65535" style="width: 160px;" />
              </n-form-item>
              <n-form-item label="发件邮箱">
                <n-input v-model:value="config.emailUser" placeholder="user@example.com" />
              </n-form-item>
              <n-form-item label="授权码/密码">
                <n-input v-model:value="config.emailPassword" type="password" placeholder="邮箱授权码" />
              </n-form-item>
              <n-form-item label="收件地址">
                <n-input v-model:value="config.emailTo" placeholder="多个地址用逗号分隔" />
              </n-form-item>
              <n-button size="small" @click="testPush('email', '邮件')">测试邮件</n-button>
            </template>
          </n-card>
        </div>

        <n-divider />

        <n-form-item>
          <n-space>
            <n-button type="primary" :loading="loading" @click="saveConfig">保存设置</n-button>
            <n-button @click="resetConfig">恢复默认</n-button>
          </n-space>
        </n-form-item>
      </n-form>
    </n-card>
  </div>
</template>

<style scoped>
.settings-page {
  max-width: 800px;
}

.api-guide {
  font-size: 13px;
  line-height: 2;
  color: #999;
}

.api-guide p {
  margin: 4px 0;
}

.api-guide strong {
  color: #18a058;
}

.status-wrapper {
  border: 1px solid rgba(24, 160, 88, 0.2);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 24px;
  background: rgba(24, 160, 88, 0.04);
}

.status-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.status-update {
  margin-left: 12px;
  font-size: 12px;
  color: #999;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}

.status-column {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 12px;
}

.status-item {
  padding: 8px 0;
  border-bottom: 1px dashed rgba(255, 255, 255, 0.12);
}

.status-item:last-child {
  border-bottom: none;
}

.status-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.status-item-meta {
  font-size: 12px;
  color: #999;
  display: flex;
  justify-content: space-between;
}

.status-empty {
  font-size: 12px;
  color: #999;
}

.push-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}
</style>
