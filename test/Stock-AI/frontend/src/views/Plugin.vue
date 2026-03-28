<script setup>
import { ref, onMounted, computed } from 'vue'
import {
  NCard,
  NButton,
  NSpace,
  NTabs,
  NTabPane,
  NList,
  NListItem,
  NThing,
  NSwitch,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NAlert,
  NEmpty,
  NTag,
  NPopconfirm,
  NSpin,
  NDescriptions,
  NDescriptionsItem,
  NCode,
  useMessage
} from 'naive-ui'
import {
  GetPlugins,
  GetNotificationTemplates,
  CreatePluginFromTemplate,
  DeletePlugin,
  TogglePlugin,
  TestNotification,
  UpdatePlugin,
  ImportPlugin,
  ExportPlugin,
  OpenPluginsDir,
  RefreshPlugins,
  GetPluginsDir
} from '../../wailsjs/go/main/App'

const message = useMessage()
const loading = ref(false)
const plugins = ref([])
const templates = ref([])

// 添加插件弹窗
const showAddModal = ref(false)
const addForm = ref({
  templateId: '',
  name: '',
  params: {}
})

// 编辑插件弹窗
const showEditModal = ref(false)
const editForm = ref({
  id: '',
  name: '',
  type: '',
  description: '',
  enabled: true,
  config: {}
})

// 导入插件弹窗
const showImportModal = ref(false)
const importJson = ref('')

// 导出插件弹窗
const showExportModal = ref(false)
const exportJson = ref('')

// 插件目录
const pluginsDir = ref('')

// 当前选中的模板
const selectedTemplate = computed(() => {
  return templates.value.find(t => t.id === addForm.value.templateId)
})

// 按类型分组的插件
const notificationPlugins = computed(() => {
  return plugins.value.filter(p => p.type === 'notification')
})

const datasourcePlugins = computed(() => {
  return plugins.value.filter(p => p.type === 'datasource')
})

const aiPlugins = computed(() => {
  return plugins.value.filter(p => p.type === 'ai')
})


// 加载插件列表
const loadPlugins = async () => {
  loading.value = true
  try {
    const data = await GetPlugins()
    plugins.value = data || []
  } catch (e) {
    console.error('加载插件失败:', e)
    message.error('加载插件失败')
  } finally {
    loading.value = false
  }
}

// 加载通知模板
const loadTemplates = async () => {
  try {
    const data = await GetNotificationTemplates()
    templates.value = data || []
  } catch (e) {
    console.error('加载模板失败:', e)
  }
}

// 打开添加弹窗
const openAddModal = () => {
  addForm.value = {
    templateId: '',
    name: '',
    params: {}
  }
  showAddModal.value = true
}

// 选择模板后初始化参数
const onTemplateChange = (templateId) => {
  const template = templates.value.find(t => t.id === templateId)
  if (template) {
    addForm.value.name = template.name
    // 初始化参数
    const params = {}
    if (template.config.params) {
      Object.keys(template.config.params).forEach(key => {
        params[key] = ''
      })
    }
    // URL参数
    if (template.config.url) {
      addForm.value.params.url = template.config.url
    }
    addForm.value.params = params
  }
}

// 获取模板需要的参数列表
const getTemplateParams = (template) => {
  if (!template) return []
  const params = []

  // 从URL中提取参数
  if (template.config.url) {
    const urlParams = template.config.url.match(/\{(\w+)\}/g)
    if (urlParams) {
      urlParams.forEach(p => {
        const key = p.replace(/[{}]/g, '')
        if (!['stockCode', 'stockName', 'alertType', 'currentPrice', 'condition', 'targetValue', 'triggerTime', 'change', 'changePercent'].includes(key)) {
          params.push({ key, label: getParamLabel(key) })
        }
      })
    }
  }

  // 从params中提取
  if (template.config.params) {
    Object.keys(template.config.params).forEach(key => {
      if (!params.find(p => p.key === key)) {
        params.push({ key, label: getParamLabel(key) })
      }
    })
  }

  return params
}

// 获取参数标签
const getParamLabel = (key) => {
  const labels = {
    'url': 'Webhook URL',
    'deviceKey': '设备Key',
    'sendKey': 'SendKey',
    'token': 'Token',
    'access_token': 'Access Token'
  }
  return labels[key] || key
}

// 添加插件
const addPlugin = async () => {
  if (!addForm.value.templateId) {
    message.warning('请选择模板')
    return
  }
  if (!addForm.value.name) {
    message.warning('请输入名称')
    return
  }

  // 检查必填参数
  const template = selectedTemplate.value
  const requiredParams = getTemplateParams(template)
  for (const param of requiredParams) {
    if (param.key === 'url' && !addForm.value.params.url) {
      message.warning('请输入Webhook URL')
      return
    }
  }

  loading.value = true
  try {
    await CreatePluginFromTemplate(
      addForm.value.templateId,
      addForm.value.name,
      addForm.value.params
    )
    message.success('添加成功')
    showAddModal.value = false
    await loadPlugins()
  } catch (e) {
    message.error('添加失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 删除插件
const deletePluginHandler = async (id) => {
  try {
    await DeletePlugin(id)
    message.success('删除成功')
    await loadPlugins()
  } catch (e) {
    message.error('删除失败: ' + e)
  }
}

// 切换插件状态
const togglePluginHandler = async (id, enabled) => {
  try {
    await TogglePlugin(id, enabled)
    message.success(enabled ? '已启用' : '已禁用')
  } catch (e) {
    message.error('操作失败: ' + e)
    await loadPlugins()
  }
}

// 测试通知
const testNotificationHandler = async (id) => {
  loading.value = true
  try {
    await TestNotification(id)
    message.success('测试消息已发送')
  } catch (e) {
    message.error('发送失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 打开编辑弹窗
const openEditModal = (plugin) => {
  editForm.value = {
    id: plugin.id,
    name: plugin.name,
    type: plugin.type,
    description: plugin.description,
    enabled: plugin.enabled,
    config: typeof plugin.config === 'string' ? JSON.parse(plugin.config) : plugin.config
  }
  showEditModal.value = true
}

// 更新插件
const updatePluginHandler = async () => {
  loading.value = true
  try {
    await UpdatePlugin(editForm.value)
    message.success('更新成功')
    showEditModal.value = false
    await loadPlugins()
  } catch (e) {
    message.error('更新失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 获取插件类型标签
const getPluginTypeTag = (type) => {
  const types = {
    'notification': { label: '通知', type: 'success' },
    'datasource': { label: '数据源', type: 'info' },
    'ai': { label: 'AI模型', type: 'warning' }
  }
  return types[type] || { label: type, type: 'default' }
}

// 解析插件配置
const parseConfig = (config) => {
  if (typeof config === 'string') {
    try {
      return JSON.parse(config)
    } catch {
      return {}
    }
  }
  return config || {}
}

// 打开导入弹窗
const openImportModal = () => {
  importJson.value = ''
  showImportModal.value = true
}

// 导入插件
const importPluginHandler = async () => {
  if (!importJson.value.trim()) {
    message.warning('请输入插件JSON')
    return
  }
  loading.value = true
  try {
    await ImportPlugin(importJson.value)
    message.success('导入成功')
    showImportModal.value = false
    await loadPlugins()
  } catch (e) {
    message.error('导入失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 导出插件
const exportPluginHandler = async (id) => {
  try {
    const json = await ExportPlugin(id)
    exportJson.value = json
    showExportModal.value = true
  } catch (e) {
    message.error('导出失败: ' + e)
  }
}

// 复制到剪贴板
const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(exportJson.value)
    message.success('已复制到剪贴板')
  } catch (e) {
    message.error('复制失败')
  }
}

// 打开插件目录
const openPluginsDirHandler = async () => {
  try {
    await OpenPluginsDir()
  } catch (e) {
    message.error('打开目录失败: ' + e)
  }
}

// 刷新插件（扫描目录中的新插件）
const refreshPluginsHandler = async () => {
  loading.value = true
  try {
    const [imported, errors] = await RefreshPlugins()
    if (imported > 0) {
      message.success(`成功导入 ${imported} 个插件`)
    }
    if (errors && errors.length > 0) {
      errors.forEach(err => message.warning(err))
    }
    if (imported === 0 && (!errors || errors.length === 0)) {
      message.info('没有发现新插件')
    }
    await loadPlugins()
  } catch (e) {
    message.error('刷新失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 加载插件目录路径
const loadPluginsDir = async () => {
  try {
    pluginsDir.value = await GetPluginsDir()
  } catch (e) {
    console.error('获取插件目录失败:', e)
  }
}

onMounted(() => {
  loadPlugins()
  loadTemplates()
  loadPluginsDir()
})
</script>

<template>
  <div class="plugin-page">
    <n-card title="插件管理" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-button @click="openPluginsDirHandler">打开插件目录</n-button>
          <n-button @click="refreshPluginsHandler">刷新</n-button>
          <n-button @click="openImportModal">导入插件</n-button>
          <n-button type="primary" @click="openAddModal">添加插件</n-button>
        </n-space>
      </template>

      <n-spin :show="loading">
        <n-tabs type="line" animated>
          <!-- 通知插件 -->
          <n-tab-pane name="notification" tab="通知插件">
            <n-alert type="info" style="margin-bottom: 16px;">
              通知插件可以在股票提醒触发时，将消息推送到钉钉、企业微信、飞书等平台。
              <br />
              <span style="color: #999;">插件目录: {{ pluginsDir }}</span>
            </n-alert>

            <n-list v-if="notificationPlugins.length > 0" bordered>
              <n-list-item v-for="plugin in notificationPlugins" :key="plugin.id">
                <n-thing>
                  <template #header>
                    <n-space align="center">
                      <span>{{ plugin.name }}</span>
                      <n-tag v-if="plugin.version" size="small">v{{ plugin.version }}</n-tag>
                      <n-tag :type="plugin.enabled ? 'success' : 'default'" size="small">
                        {{ plugin.enabled ? '已启用' : '已禁用' }}
                      </n-tag>
                    </n-space>
                  </template>
                  <template #header-extra>
                    <n-space>
                      <n-switch
                        :value="plugin.enabled"
                        @update:value="(val) => togglePluginHandler(plugin.id, val)"
                      />
                      <n-button size="small" @click="testNotificationHandler(plugin.id)" :disabled="!plugin.enabled">
                        测试
                      </n-button>
                      <n-button size="small" @click="exportPluginHandler(plugin.id)">
                        导出
                      </n-button>
                      <n-button size="small" @click="openEditModal(plugin)">
                        编辑
                      </n-button>
                      <n-popconfirm @positive-click="deletePluginHandler(plugin.id)">
                        <template #trigger>
                          <n-button size="small" type="error">删除</n-button>
                        </template>
                        确定要删除这个插件吗？
                      </n-popconfirm>
                    </n-space>
                  </template>
                  <template #description>
                    {{ plugin.description }}
                  </template>
                  <n-descriptions :column="2" label-placement="left" size="small" style="margin-top: 8px;">
                    <n-descriptions-item label="URL">
                      {{ parseConfig(plugin.config).url || '-' }}
                    </n-descriptions-item>
                    <n-descriptions-item label="方法">
                      {{ parseConfig(plugin.config).method || 'POST' }}
                    </n-descriptions-item>
                  </n-descriptions>
                </n-thing>
              </n-list-item>
            </n-list>
            <n-empty v-else description="暂无通知插件，点击上方按钮添加" />
          </n-tab-pane>

          <!-- 数据源插件 -->
          <n-tab-pane name="datasource" tab="数据源插件">
            <n-alert type="info" style="margin-bottom: 16px;">
              数据源插件可以接入自定义的行情数据API，获取股票实时行情。
            </n-alert>

            <n-list v-if="datasourcePlugins.length > 0" bordered>
              <n-list-item v-for="plugin in datasourcePlugins" :key="plugin.id">
                <n-thing>
                  <template #header>
                    <n-space align="center">
                      <span>{{ plugin.name }}</span>
                      <n-tag v-if="plugin.version" size="small">v{{ plugin.version }}</n-tag>
                      <n-tag :type="plugin.enabled ? 'success' : 'default'" size="small">
                        {{ plugin.enabled ? '已启用' : '已禁用' }}
                      </n-tag>
                    </n-space>
                  </template>
                  <template #header-extra>
                    <n-space>
                      <n-switch :value="plugin.enabled" @update:value="(val) => togglePluginHandler(plugin.id, val)" />
                      <n-button size="small" @click="exportPluginHandler(plugin.id)">导出</n-button>
                      <n-button size="small" @click="openEditModal(plugin)">编辑</n-button>
                      <n-popconfirm @positive-click="deletePluginHandler(plugin.id)">
                        <template #trigger>
                          <n-button size="small" type="error">删除</n-button>
                        </template>
                        确定要删除这个插件吗？
                      </n-popconfirm>
                    </n-space>
                  </template>
                  <template #description>{{ plugin.description }}</template>
                  <n-descriptions :column="2" label-placement="left" size="small" style="margin-top: 8px;">
                    <n-descriptions-item label="BaseURL">{{ parseConfig(plugin.config).baseUrl || '-' }}</n-descriptions-item>
                    <n-descriptions-item label="行情端点">{{ parseConfig(plugin.config).endpoints?.quote || '-' }}</n-descriptions-item>
                  </n-descriptions>
                </n-thing>
              </n-list-item>
            </n-list>
            <n-empty v-else description="暂无数据源插件，可通过导入JSON添加" />
          </n-tab-pane>

          <!-- AI模型插件 -->
          <n-tab-pane name="ai" tab="AI模型插件">
            <n-alert type="info" style="margin-bottom: 16px;">
              AI模型插件可以接入自定义的AI大模型，用于股票分析。
            </n-alert>

            <n-list v-if="aiPlugins.length > 0" bordered>
              <n-list-item v-for="plugin in aiPlugins" :key="plugin.id">
                <n-thing>
                  <template #header>
                    <n-space align="center">
                      <span>{{ plugin.name }}</span>
                      <n-tag v-if="plugin.version" size="small">v{{ plugin.version }}</n-tag>
                      <n-tag :type="plugin.enabled ? 'success' : 'default'" size="small">
                        {{ plugin.enabled ? '已启用' : '已禁用' }}
                      </n-tag>
                    </n-space>
                  </template>
                  <template #header-extra>
                    <n-space>
                      <n-switch :value="plugin.enabled" @update:value="(val) => togglePluginHandler(plugin.id, val)" />
                      <n-button size="small" @click="exportPluginHandler(plugin.id)">导出</n-button>
                      <n-button size="small" @click="openEditModal(plugin)">编辑</n-button>
                      <n-popconfirm @positive-click="deletePluginHandler(plugin.id)">
                        <template #trigger>
                          <n-button size="small" type="error">删除</n-button>
                        </template>
                        确定要删除这个插件吗？
                      </n-popconfirm>
                    </n-space>
                  </template>
                  <template #description>{{ plugin.description }}</template>
                  <n-descriptions :column="2" label-placement="left" size="small" style="margin-top: 8px;">
                    <n-descriptions-item label="BaseURL">{{ parseConfig(plugin.config).baseUrl || '-' }}</n-descriptions-item>
                    <n-descriptions-item label="模型">{{ parseConfig(plugin.config).model || '-' }}</n-descriptions-item>
                  </n-descriptions>
                </n-thing>
              </n-list-item>
            </n-list>
            <n-empty v-else description="暂无AI模型插件，可通过导入JSON添加" />
          </n-tab-pane>
        </n-tabs>
      </n-spin>
    </n-card>

    <!-- 添加插件弹窗 -->
    <n-modal
      v-model:show="showAddModal"
      preset="card"
      title="添加通知插件"
      style="width: 500px;"
    >
      <n-form label-placement="left" label-width="100">
        <n-form-item label="选择模板">
          <n-select
            v-model:value="addForm.templateId"
            :options="templates.map(t => ({ label: t.name, value: t.id }))"
            placeholder="请选择通知模板"
            @update:value="onTemplateChange"
          />
        </n-form-item>

        <n-form-item label="插件名称">
          <n-input v-model:value="addForm.name" placeholder="请输入插件名称" />
        </n-form-item>

        <template v-if="selectedTemplate">
          <n-alert type="info" style="margin-bottom: 16px;">
            {{ selectedTemplate.description }}
          </n-alert>

          <!-- Webhook URL -->
          <n-form-item label="Webhook URL" v-if="selectedTemplate.id !== 'bark'">
            <n-input
              v-model:value="addForm.params.url"
              placeholder="请输入Webhook URL"
              type="textarea"
              :rows="2"
            />
          </n-form-item>

          <!-- Bark特殊处理 -->
          <template v-if="selectedTemplate.id === 'bark'">
            <n-form-item label="设备Key">
              <n-input v-model:value="addForm.params.deviceKey" placeholder="请输入Bark设备Key" />
            </n-form-item>
          </template>

          <!-- Server酱特殊处理 -->
          <template v-if="selectedTemplate.id === 'serverchan'">
            <n-form-item label="SendKey">
              <n-input v-model:value="addForm.params.sendKey" placeholder="请输入Server酱SendKey" />
            </n-form-item>
          </template>

          <!-- PushPlus特殊处理 -->
          <template v-if="selectedTemplate.id === 'pushplus'">
            <n-form-item label="Token">
              <n-input v-model:value="addForm.params.token" placeholder="请输入PushPlus Token" />
            </n-form-item>
          </template>
        </template>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" @click="addPlugin" :loading="loading">确定</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 编辑插件弹窗 -->
    <n-modal
      v-model:show="showEditModal"
      preset="card"
      title="编辑插件"
      style="width: 500px;"
    >
      <n-form label-placement="left" label-width="100">
        <n-form-item label="插件名称">
          <n-input v-model:value="editForm.name" placeholder="请输入插件名称" />
        </n-form-item>

        <n-form-item label="Webhook URL">
          <n-input
            v-model:value="editForm.config.url"
            placeholder="请输入Webhook URL"
            type="textarea"
            :rows="2"
          />
        </n-form-item>

        <n-form-item label="启用状态">
          <n-switch v-model:value="editForm.enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" @click="updatePluginHandler" :loading="loading">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 导入插件弹窗 -->
    <n-modal
      v-model:show="showImportModal"
      preset="card"
      title="导入插件"
      style="width: 600px;"
    >
      <n-alert type="info" style="margin-bottom: 16px;">
        粘贴插件的 JSON 配置，或将插件文件放入插件目录后点击"刷新"按钮。
        <br />
        <a href="https://github.com/xxx/stock-ai-plugins" target="_blank" style="color: #63e2b7;">查看插件开发文档</a>
      </n-alert>
      <n-input
        v-model:value="importJson"
        type="textarea"
        placeholder='{"id": "my-plugin", "name": "我的插件", "type": "notification", ...}'
        :rows="12"
        style="font-family: monospace;"
      />
      <template #footer>
        <n-space justify="end">
          <n-button @click="showImportModal = false">取消</n-button>
          <n-button type="primary" @click="importPluginHandler" :loading="loading">导入</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 导出插件弹窗 -->
    <n-modal
      v-model:show="showExportModal"
      preset="card"
      title="导出插件"
      style="width: 600px;"
    >
      <n-alert type="info" style="margin-bottom: 16px;">
        复制下方 JSON 配置分享给其他用户，或保存为 .json 文件。
      </n-alert>
      <n-input
        v-model:value="exportJson"
        type="textarea"
        :rows="12"
        readonly
        style="font-family: monospace;"
      />
      <template #footer>
        <n-space justify="end">
          <n-button @click="showExportModal = false">关闭</n-button>
          <n-button type="primary" @click="copyToClipboard">复制到剪贴板</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.plugin-page {
  padding: 16px;
}
</style>
