<script setup>
import { ref, onMounted } from 'vue'
import { NModal, NCard, NButton, NCheckbox, NSpace, NAlert, NScrollbar } from 'naive-ui'

const showModal = ref(false)
const agreed = ref(false)

const DISCLAIMER_KEY = 'stock_ai_disclaimer_agreed'
const DISCLAIMER_VERSION = '1.0' // 更新版本号可以让用户重新确认

onMounted(() => {
  const agreedVersion = localStorage.getItem(DISCLAIMER_KEY)
  if (agreedVersion !== DISCLAIMER_VERSION) {
    showModal.value = true
  }
})

const handleAgree = () => {
  if (agreed.value) {
    localStorage.setItem(DISCLAIMER_KEY, DISCLAIMER_VERSION)
    showModal.value = false
  }
}
</script>

<template>
  <n-modal
    v-model:show="showModal"
    :mask-closable="false"
    :close-on-esc="false"
    preset="card"
    title="用户协议与免责声明"
    style="width: 600px; max-width: 90vw;"
  >
    <n-scrollbar style="max-height: 60vh;">
      <div class="disclaimer-content">
        <n-alert type="warning" :bordered="false" style="margin-bottom: 16px;">
          <strong>请仔细阅读以下声明，继续使用本软件即表示您已理解并同意以下条款。</strong>
        </n-alert>

        <h3>一、软件性质</h3>
        <p>本软件是一个<strong>开源的股票数据展示和AI接口调用工具</strong>，仅供学习研究使用。软件本身不提供任何AI服务，AI分析功能需要用户自行配置第三方AI服务（如DeepSeek、OpenAI等）的API密钥。</p>

        <h3>二、非投资建议声明</h3>
        <p>本软件提供的所有信息、数据、分析结果<strong>仅供学习研究和参考</strong>，不构成任何投资建议、投资咨询或证券推荐。</p>

        <h3>三、无资质声明</h3>
        <p>本软件及其开发者<strong>不具备证券投资咨询业务资格</strong>，不提供任何形式的证券投资咨询服务。</p>

        <h3>四、AI生成内容声明</h3>
        <p>软件中的AI分析结果由用户自行配置的第三方AI服务生成，可能存在错误或偏差，<strong>仅供参考，不应作为投资决策的依据</strong>。开发者对AI生成的内容不承担任何责任。</p>

        <h3>五、数据来源声明</h3>
        <p>本软件数据来源于公开渠道（新浪财经、腾讯财经、东方财富等），<strong>不保证数据的准确性、完整性和及时性</strong>。</p>

        <h3>六、投资风险提示</h3>
        <p><strong style="color: #f0a020;">股市有风险，投资需谨慎。</strong>任何投资决策应基于您自己的独立判断，并建议咨询专业的持牌投资顾问。</p>

        <h3>七、免责条款</h3>
        <p>使用本软件造成的任何直接或间接损失，开发者<strong>不承担任何责任</strong>。用户应自行承担使用本软件的全部风险。</p>

        <h3>八、开源协议</h3>
        <p>本软件基于MIT协议开源，您可以自由使用、修改和分发本软件，但需保留原始版权声明和免责声明，并默认仅限个人学习或内部研究使用。</p>

        <h3>九、商业使用限制</h3>
        <p style="color: #ff4d4f; font-weight: bold;">
          未经作者书面许可，禁止将本软件或其修改版本用于任何形式的售卖、商业分发、收费托管、付费服务或变相盈利行为。若需商业化使用，请通过作者联系方式获取授权。
        </p>
      </div>
    </n-scrollbar>

    <template #footer>
      <n-space vertical style="width: 100%;">
        <n-checkbox v-model:checked="agreed">
          我已阅读并理解上述声明，同意继续使用本软件
        </n-checkbox>
        <n-button
          type="primary"
          block
          :disabled="!agreed"
          @click="handleAgree"
        >
          同意并继续
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<style scoped>
.disclaimer-content {
  line-height: 1.8;
  color: #e0e0e0;
}

.disclaimer-content h3 {
  margin: 16px 0 8px 0;
  color: #18a058;
  font-size: 15px;
}

.disclaimer-content p {
  margin: 0 0 12px 0;
  font-size: 14px;
  text-align: justify;
}
</style>
