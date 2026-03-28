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
  NModal,
  NForm,
  NFormItem,
  NInput,
  NAlert,
  NEmpty,
  NTag,
  NPopconfirm,
  NSpin,
  useMessage
} from 'naive-ui'
import {
  GetPromptTypes,
  ListAllPrompts,
  GetPrompt,
  CreatePrompt,
  UpdatePrompt,
  DeletePrompt,
  RenamePrompt,
  ExportPrompt,
  GetPromptsDir
} from '../../wailsjs/go/main/App'

const message = useMessage()
const loading = ref(false)
const promptTypes = ref([])
const allPrompts = ref({})
const promptsDir = ref('')
const activeTab = ref('indicator')

// 新建/编辑弹窗
const showEditModal = ref(false)
const editMode = ref('create') // create | edit
const editForm = ref({
  type: '',
  name: '',
  content: '',
  originalName: '' // 用于编辑时记录原名称
})

// 导出弹窗
const showExportModal = ref(false)
const exportContent = ref('')
const exportName = ref('')

// 导入弹窗
const showImportModal = ref(false)
const importForm = ref({
  type: '',
  name: '',
  content: ''
})

// 获取类型的中文名称
const getTypeName = (type) => {
  const found = promptTypes.value.find(t => t.type === type)
  return found ? found.name : type
}

// 获取类型的描述
const getTypeDescription = (type) => {
  const found = promptTypes.value.find(t => t.type === type)
  return found ? found.description : ''
}

// 加载提示词类型
const loadPromptTypes = async () => {
  try {
    const types = await GetPromptTypes()
    promptTypes.value = types || []
    if (types && types.length > 0) {
      activeTab.value = types[0].type
    }
  } catch (e) {
    console.error('加载提示词类型失败:', e)
  }
}

// 加载所有提示词
const loadAllPrompts = async () => {
  loading.value = true
  try {
    const data = await ListAllPrompts()
    allPrompts.value = data || {}
  } catch (e) {
    console.error('加载提示词失败:', e)
    message.error('加载提示词失败')
  } finally {
    loading.value = false
  }
}

// 加载提示词目录
const loadPromptsDir = async () => {
  try {
    promptsDir.value = await GetPromptsDir()
  } catch (e) {
    console.error('获取提示词目录失败:', e)
  }
}

// 打开新建弹窗
const openCreateModal = (type) => {
  editMode.value = 'create'
  editForm.value = {
    type: type,
    name: '',
    content: '',
    originalName: ''
  }
  showEditModal.value = true
}

// 打开编辑弹窗
const openEditModal = async (type, name) => {
  loading.value = true
  try {
    const prompt = await GetPrompt(type, name)
    editMode.value = 'edit'
    editForm.value = {
      type: type,
      name: prompt.name,
      content: prompt.content,
      originalName: prompt.name
    }
    showEditModal.value = true
  } catch (e) {
    message.error('获取提示词失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 保存提示词
const savePrompt = async () => {
  if (!editForm.value.name.trim()) {
    message.warning('请输入名称')
    return
  }
  if (!editForm.value.content.trim()) {
    message.warning('请输入提示词内容')
    return
  }

  loading.value = true
  try {
    if (editMode.value === 'create') {
      await CreatePrompt(editForm.value.type, editForm.value.name, editForm.value.content)
      message.success('创建成功')
    } else {
      // 如果名称改变了，先重命名
      if (editForm.value.name !== editForm.value.originalName) {
        await RenamePrompt(editForm.value.type, editForm.value.originalName, editForm.value.name)
      }
      await UpdatePrompt(editForm.value.type, editForm.value.name, editForm.value.content)
      message.success('保存成功')
    }
    showEditModal.value = false
    await loadAllPrompts()
  } catch (e) {
    message.error('保存失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 删除提示词
const deletePromptHandler = async (type, name) => {
  try {
    await DeletePrompt(type, name)
    message.success('删除成功')
    await loadAllPrompts()
  } catch (e) {
    message.error('删除失败: ' + e)
  }
}

// 导出提示词
const exportPromptHandler = async (type, name) => {
  try {
    const content = await ExportPrompt(type, name)
    exportContent.value = content
    exportName.value = name
    showExportModal.value = true
  } catch (e) {
    message.error('导出失败: ' + e)
  }
}

// 复制到剪贴板
const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(exportContent.value)
    message.success('已复制到剪贴板')
  } catch (e) {
    message.error('复制失败')
  }
}

// 下载为TXT文件
const downloadAsTxt = () => {
  const blob = new Blob([exportContent.value], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = exportName.value + '.txt'
  a.click()
  URL.revokeObjectURL(url)
  message.success('下载成功')
}

// 打开导入弹窗
const openImportModal = (type) => {
  importForm.value = {
    type: type,
    name: '',
    content: ''
  }
  showImportModal.value = true
}

// 导入提示词
const importPromptHandler = async () => {
  if (!importForm.value.name.trim()) {
    message.warning('请输入名称')
    return
  }
  if (!importForm.value.content.trim()) {
    message.warning('请输入或粘贴提示词内容')
    return
  }

  loading.value = true
  try {
    await CreatePrompt(importForm.value.type, importForm.value.name, importForm.value.content)
    message.success('导入成功')
    showImportModal.value = false
    await loadAllPrompts()
  } catch (e) {
    message.error('导入失败: ' + e)
  } finally {
    loading.value = false
  }
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  loadPromptTypes()
  loadAllPrompts()
  loadPromptsDir()
})
</script>

<template>
  <div class="prompt-page">
    <n-card title="AI提示词管理" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-tag type="info">目录: {{ promptsDir }}</n-tag>
        </n-space>
      </template>

      <n-alert type="info" style="margin-bottom: 16px;">
        AI提示词是纯文本文件（.txt），用于指导AI分析股票。您可以创建、编辑、导入导出提示词，无需编写代码。
      </n-alert>

      <n-spin :show="loading">
        <n-tabs v-model:value="activeTab" type="line" animated>
          <n-tab-pane
            v-for="pType in promptTypes"
            :key="pType.type"
            :name="pType.type"
            :tab="pType.name"
          >
            <n-alert type="default" style="margin-bottom: 16px;">
              {{ pType.description }}
            </n-alert>

            <n-space style="margin-bottom: 16px;">
              <n-button type="primary" @click="openCreateModal(pType.type)">
                新建{{ pType.name }}
              </n-button>
              <n-button @click="openImportModal(pType.type)">
                导入
              </n-button>
            </n-space>

            <n-list v-if="allPrompts[pType.type] && allPrompts[pType.type].length > 0" bordered>
              <n-list-item v-for="prompt in allPrompts[pType.type]" :key="prompt.name">
                <n-thing>
                  <template #header>
                    <n-space align="center">
                      <span style="font-weight: 500;">{{ prompt.name }}</span>
                      <n-tag size="small" type="success">TXT</n-tag>
                    </n-space>
                  </template>
                  <template #header-extra>
                    <n-space>
                      <n-button size="small" @click="openEditModal(pType.type, prompt.name)">
                        编辑
                      </n-button>
                      <n-button size="small" @click="exportPromptHandler(pType.type, prompt.name)">
                        导出
                      </n-button>
                      <n-popconfirm @positive-click="deletePromptHandler(pType.type, prompt.name)">
                        <template #trigger>
                          <n-button size="small" type="error">删除</n-button>
                        </template>
                        确定要删除「{{ prompt.name }}」吗？
                      </n-popconfirm>
                    </n-space>
                  </template>
                  <template #description>
                    <div class="prompt-preview">
                      {{ prompt.content.length > 100 ? prompt.content.substring(0, 100) + '...' : prompt.content }}
                    </div>
                    <div class="prompt-meta">
                      更新时间: {{ formatTime(prompt.updatedAt) }}
                    </div>
                  </template>
                </n-thing>
              </n-list-item>
            </n-list>
            <n-empty v-else :description="`暂无${pType.name}，点击上方按钮创建`" />
          </n-tab-pane>
        </n-tabs>
      </n-spin>
    </n-card>

    <!-- 新建/编辑弹窗 -->
    <n-modal
      v-model:show="showEditModal"
      preset="card"
      :title="editMode === 'create' ? '新建提示词' : '编辑提示词'"
      style="width: 700px;"
    >
      <n-form label-placement="top">
        <n-form-item label="名称" required>
          <n-input
            v-model:value="editForm.name"
            placeholder="请输入提示词名称（将作为文件名）"
          />
        </n-form-item>

        <n-form-item label="类型">
          <n-tag type="info">{{ getTypeName(editForm.type) }}</n-tag>
        </n-form-item>

        <n-form-item label="提示词内容" required>
          <n-input
            v-model:value="editForm.content"
            type="textarea"
            placeholder="请输入提示词内容...

可用变量：
{code} - 股票代码
{name} - 股票名称
{price} - 当前价格
{change} - 涨跌额
{changePercent} - 涨跌幅
{volume} - 成交量
{high} - 最高价
{low} - 最低价
{open} - 开盘价
{preClose} - 昨收价
{klines} - K线数据"
            :rows="15"
            style="font-family: monospace;"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" @click="savePrompt" :loading="loading">
            {{ editMode === 'create' ? '创建' : '保存' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 导出弹窗 -->
    <n-modal
      v-model:show="showExportModal"
      preset="card"
      title="导出提示词"
      style="width: 600px;"
    >
      <n-alert type="info" style="margin-bottom: 16px;">
        复制下方内容分享给其他用户，或下载为TXT文件。
      </n-alert>
      <n-input
        v-model:value="exportContent"
        type="textarea"
        :rows="12"
        readonly
        style="font-family: monospace;"
      />
      <template #footer>
        <n-space justify="end">
          <n-button @click="showExportModal = false">关闭</n-button>
          <n-button @click="downloadAsTxt">下载TXT</n-button>
          <n-button type="primary" @click="copyToClipboard">复制到剪贴板</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 导入弹窗 -->
    <n-modal
      v-model:show="showImportModal"
      preset="card"
      title="导入提示词"
      style="width: 600px;"
    >
      <n-form label-placement="top">
        <n-form-item label="名称" required>
          <n-input
            v-model:value="importForm.name"
            placeholder="请输入提示词名称"
          />
        </n-form-item>

        <n-form-item label="提示词内容" required>
          <n-input
            v-model:value="importForm.content"
            type="textarea"
            placeholder="粘贴提示词内容..."
            :rows="12"
            style="font-family: monospace;"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showImportModal = false">取消</n-button>
          <n-button type="primary" @click="importPromptHandler" :loading="loading">导入</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.prompt-page {
  padding: 16px;
}

.prompt-preview {
  color: #666;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  margin-bottom: 8px;
}

.prompt-meta {
  color: #999;
  font-size: 12px;
}
</style>
