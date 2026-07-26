<template>
  <div>
    <div class="gva-table-box">
      <div class="table-header">
        <div>
          <div class="text-lg font-medium">AI 能力配置</div>
          <div class="mt-1 text-sm text-gray-500">
            每项能力选择自己的主模型和提示词；平台是否开放 AI 由“平台整体能力策略”统一控制。
          </div>
        </div>
        <el-button :loading="listLoading" @click="loadList">刷新</el-button>
      </div>
      <el-alert v-if="listError" :title="listError" type="error" show-icon class="mb-4" />
      <el-table v-loading="listLoading" :data="items" row-key="code">
        <el-table-column label="能力" min-width="220">
          <template #default="{ row }">
            <div class="font-medium">{{ row.name }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.code }}</div>
          </template>
        </el-table-column>
        <el-table-column label="模型类型" width="120">
          <template #default="{ row }">
            {{ modelCapabilityLabel(row.requiredModelCapability) }}
          </template>
        </el-table-column>
        <el-table-column label="主模型" min-width="180">
          <template #default="{ row }">{{ row.primaryModelName || '未配置' }}</template>
        </el-table-column>
        <el-table-column label="积分" width="95">
          <template #default="{ row }">{{ row.pointCost }} / 次</template>
        </el-table-column>
        <el-table-column label="免费额度" width="105">
          <template #default="{ row }">{{ row.freeQuotaPerDay }} / 日</template>
        </el-table-column>
        <el-table-column label="每日上限" width="105">
          <template #default="{ row }">
            {{ row.dailyLimitPerUser ? `${row.dailyLimitPerUser} / 日` : '不限制' }}
          </template>
        </el-table-column>
        <el-table-column label="提示词" width="120">
          <template #default="{ row }">
            <div>{{ promptModeLabel(row.promptMode) }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ shortHash(row.promptHash) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="95">
          <template #default="{ row }">
            <el-tag :type="row.ready ? 'success' : 'warning'">
              {{ row.ready ? '可用' : '未就绪' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openWorkspace(row.code)">
              {{ canUpdateConfig ? '编辑能力配置' : '查看配置' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-drawer
      v-model="workspaceVisible"
      :title="workspace?.summary?.name || '能力配置'"
      size="900px"
      destroy-on-close
    >
      <div v-loading="workspaceLoading" class="min-h-64">
        <el-alert
          v-if="workspaceError"
          :title="workspaceError"
          type="error"
          show-icon
          class="mb-4"
        />
        <template v-if="workspace">
          <el-alert
            title="配置保存后立即生效；是否向小程序开放仍由平台总开关统一决定。"
            type="info"
            show-icon
            :closable="false"
            class="mb-4"
          />
          <div v-if="canReadAIUsage" class="mb-4">
            <el-button @click="openCapabilityUsages">查看调用记录</el-button>
          </div>
          <el-tabs v-model="activeTab">
            <el-tab-pane label="模型与调用规则" name="config">
              <div class="section-toolbar">
                <el-descriptions :column="2" border class="flex-1">
                  <el-descriptions-item label="当前版本">
                    v{{ workspace.config?.version || 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="更新时间">
                    {{ formatDateTime(workspace.config?.updatedAt) }}
                  </el-descriptions-item>
                  <el-descriptions-item label="更新人">
                    {{ administratorLabel(workspace.config?.updatedBy) }}
                  </el-descriptions-item>
                  <el-descriptions-item label="修改原因">
                    {{ workspace.config?.reason || '—' }}
                  </el-descriptions-item>
                </el-descriptions>
                <el-button
                  v-if="canUpdateConfig && !configEditing"
                  type="primary"
                  @click="beginConfigEdit"
                >
                  编辑能力配置
                </el-button>
              </div>
              <el-form ref="configFormRef" :model="configForm" :rules="configRules" label-width="150px">
                <el-form-item label="主模型" prop="primaryModelId">
                  <el-select
                    v-model="configForm.primaryModelId"
                    filterable
                    class="w-full"
                    placeholder="请选择已启用且模态匹配的模型"
                    :disabled="!configEditing"
                  >
                    <el-option
                      v-for="model in modelOptions"
                      :key="model.id"
                      :label="`${model.name}（${model.provider?.name || '未知供应商'}）`"
                      :value="model.id"
                    />
                  </el-select>
                  <div class="form-hint">
                    当前能力要求{{ modelCapabilityLabel(requiredCapability(activeCode)) }}模型，不配置备用模型。
                  </div>
                </el-form-item>
                <el-form-item label="每次消耗积分">
                  <el-input-number v-model="configForm.pointCost" :min="0" :max="100000" :disabled="!configEditing" />
                </el-form-item>
                <el-form-item label="每用户每日上限">
                  <el-input-number v-model="configForm.dailyLimitPerUser" :min="0" :max="100000" :disabled="!configEditing" />
                  <span class="form-inline-hint">0 表示不单独限制</span>
                </el-form-item>
                <el-form-item label="每日免费额度">
                  <el-input-number v-model="configForm.freeQuotaPerDay" :min="0" :max="100000" :disabled="!configEditing" />
                </el-form-item>
                <el-form-item label="调用超时">
                  <el-input-number v-model="configForm.timeoutMs" :min="1000" :max="120000" :step="1000" :disabled="!configEditing" />
                  <span class="form-inline-hint">毫秒</span>
                </el-form-item>
                <el-form-item label="固定结果校验">
                  <div class="fixed-validation-rules">
                    <div
                      v-for="rule in fixedValidationRules"
                      :key="rule"
                      class="fixed-validation-rule"
                    >
                      {{ rule }}
                    </div>
                    <el-empty
                      v-if="fixedValidationRules.length === 0"
                      description="暂无校验说明"
                      :image-size="48"
                    />
                    <div class="form-hint">
                      由服务端固定执行，不受提示词和本页配置影响，也不能在管理端关闭。
                    </div>
                  </div>
                </el-form-item>
                <el-form-item v-if="configEditing" label="修改原因" prop="reason">
                  <el-input
                    v-model.trim="configForm.reason"
                    type="textarea"
                    :rows="3"
                    maxlength="200"
                    show-word-limit
                  />
                </el-form-item>
                <el-form-item v-if="configEditing">
                  <el-button type="primary" :loading="configSaving" @click="saveConfig">
                    保存并立即生效
                  </el-button>
                  <el-button @click="discardConfigChanges">放弃本地修改</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <el-tab-pane label="提示词" name="prompt">
              <el-alert
                v-if="!canReadPrompt"
                title="当前角色没有查看提示词正文的权限。"
                type="warning"
                show-icon
              />
              <div v-else-if="!promptWorkspace" class="prompt-load">
                <p>提示词正文属于受限配置，按需加载并记录查看审计。</p>
                <el-button type="primary" :loading="promptLoading" @click="loadPromptWorkspace">
                  查看完整提示词
                </el-button>
              </div>
              <template v-else>
                <el-alert
                  v-if="!workspace.config"
                  title="请先保存模型与调用规则，再配置提示词。"
                  type="warning"
                  show-icon
                  :closable="false"
                  class="mb-4"
                />
                <div class="section-toolbar">
                  <el-descriptions :column="2" border class="flex-1">
                    <el-descriptions-item label="提示词来源">
                      {{ promptModeLabel(promptWorkspace.current?.mode || 'default') }}
                    </el-descriptions-item>
                    <el-descriptions-item label="当前摘要">
                      <code>{{ promptWorkspace.effectivePromptHash || '—' }}</code>
                    </el-descriptions-item>
                    <el-descriptions-item label="固定结果结构" :span="2">
                      <pre class="schema-preview">{{ fixedOutputSchemaText }}</pre>
                    </el-descriptions-item>
                  </el-descriptions>
                  <el-button
                    v-if="canUpdatePrompt && !promptEditing"
                    type="primary"
                    :disabled="!workspace.config"
                    @click="beginPromptEdit"
                  >
                    编辑提示词
                  </el-button>
                </div>
                <el-form ref="promptFormRef" :model="promptForm" :rules="promptRules" label-width="130px">
                  <el-form-item label="当前版本">
                    v{{ promptWorkspace.current?.version || 0 }}
                  </el-form-item>
                  <el-form-item label="提示词来源">
                    <el-radio-group
                      v-model="promptForm.mode"
                      :disabled="!promptEditing"
                      @change="handlePromptModeChange"
                    >
                      <el-radio value="default">平台默认</el-radio>
                      <el-radio value="custom">自定义</el-radio>
                    </el-radio-group>
                  </el-form-item>
                  <el-form-item label="允许变量">
                    <el-tag
                      v-for="variable in allowedVariables"
                      :key="variable.name"
                      class="mr-2 mb-2 prompt-variable"
                      :class="{ clickable: promptEditing && promptForm.mode === 'custom' }"
                      @click="insertPromptVariable(variable.name)"
                    >
                      {{ promptVariableLabel(variable) }}
                    </el-tag>
                  </el-form-item>
                  <el-form-item label="系统提示词" prop="systemPrompt">
                    <el-input
                      v-model="promptForm.systemPrompt"
                      type="textarea"
                      :rows="8"
                      maxlength="12000"
                      show-word-limit
                      :disabled="promptForm.mode === 'default' || !promptEditing"
                      @focus="promptInsertTarget = 'systemPrompt'"
                    />
                  </el-form-item>
                  <el-form-item label="用户提示词模板" prop="userPromptTemplate">
                    <el-input
                      v-model="promptForm.userPromptTemplate"
                      type="textarea"
                      :rows="8"
                      maxlength="12000"
                      show-word-limit
                      :disabled="promptForm.mode === 'default' || !promptEditing"
                      @focus="promptInsertTarget = 'userPromptTemplate'"
                    />
                  </el-form-item>
                  <el-form-item label="测试变量 JSON">
                    <el-input v-model="testVariablesText" type="textarea" :rows="6" />
                  </el-form-item>
                  <el-form-item v-if="promptEditing" label="修改原因" prop="reason">
                    <el-input
                      v-model.trim="promptForm.reason"
                      type="textarea"
                      :rows="3"
                      maxlength="200"
                      show-word-limit
                    />
                  </el-form-item>
                </el-form>
                <div class="action-row">
                  <el-button
                    v-if="promptEditing"
                    @click="restoreDefaultPrompt"
                  >
                    恢复平台默认
                  </el-button>
                  <el-button
                    v-if="promptEditing"
                    :loading="promptValidating"
                    @click="validatePrompt"
                  >
                    校验
                  </el-button>
                  <el-button :loading="renderLoading" @click="renderPrompt">
                    预览变量替换
                  </el-button>
                  <el-button
                    v-if="canTestPrompt"
                    :loading="testLoading"
                    :disabled="!workspace.config"
                    @click="testPrompt"
                  >
                    测试当前填写内容
                  </el-button>
                  <el-button
                    v-if="promptEditing"
                    type="primary"
                    :loading="promptSaving"
                    :disabled="!workspace.config"
                    @click="savePrompt"
                  >
                    保存并立即生效
                  </el-button>
                  <el-button v-if="promptEditing" @click="discardPromptChanges">
                    放弃本地修改
                  </el-button>
                </div>
                <el-alert
                  v-if="validationResult"
                  :title="validationResult.valid ? '提示词校验通过' : '提示词校验未通过'"
                  :type="validationResult.valid ? 'success' : 'error'"
                  show-icon
                  class="mt-4"
                >
                  <template #default>
                    <div v-for="message in validationResult.errors || []" :key="message">
                      {{ message }}
                    </div>
                  </template>
                </el-alert>
                <el-descriptions v-if="renderResult" :column="1" border class="mt-4">
                  <el-descriptions-item label="系统提示词">
                    <pre class="prompt-preview">{{ renderResult.systemPrompt }}</pre>
                  </el-descriptions-item>
                  <el-descriptions-item label="用户提示词">
                    <pre class="prompt-preview">{{ renderResult.userPrompt }}</pre>
                  </el-descriptions-item>
                </el-descriptions>
                <el-descriptions v-if="testResult" :column="2" border class="mt-4">
                  <el-descriptions-item label="测试结果">
                    <el-tag :type="testResult.success ? 'success' : 'danger'">
                      {{ testResult.success ? '成功' : '失败' }}
                    </el-tag>
                  </el-descriptions-item>
                  <el-descriptions-item label="耗时">{{ testResult.durationMs }} ms</el-descriptions-item>
                  <el-descriptions-item label="说明" :span="2">
                    {{ testResult.safeSummary || '—' }}
                  </el-descriptions-item>
                </el-descriptions>
                <el-descriptions
                  v-if="promptWorkspace.lastSuccessfulTest"
                  title="最近一次成功测试"
                  :column="2"
                  border
                  class="mt-4"
                >
                  <el-descriptions-item label="模型">
                    {{ promptWorkspace.lastSuccessfulTest.modelName || '—' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="供应商">
                    {{ promptWorkspace.lastSuccessfulTest.providerName || '—' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="提示词摘要">
                    {{ promptWorkspace.lastSuccessfulTest.promptHash || '—' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="测试时间">
                    {{ formatDateTime(promptWorkspace.lastSuccessfulTest.testedAt) }}
                  </el-descriptions-item>
                </el-descriptions>
              </template>
            </el-tab-pane>
            <el-tab-pane v-if="canReadAudit" label="操作审计" name="audit" lazy>
              <AuditPanel target-type="ai_capability" :target-id="activeCode" />
            </el-tab-pane>
          </el-tabs>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import AuditPanel from '@/view/orderFood/components/AuditPanel.vue'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import {
  getAICapabilityDetail,
  getAICapabilityList,
  getAICapabilityPromptWorkspace,
  getAIModelList,
  renderAICapabilityPrompt,
  testAICapabilityPrompt,
  updateAICapability,
  updateAICapabilityPrompt,
  validateAICapabilityPrompt
} from '@/api/orderfood/ai'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodAiCapabilityConfig' })

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const canUpdateConfig = computed(
  () =>
    Boolean(btnAuth['orderfood:capability:update']) &&
    Boolean(btnAuth['orderfood:model:read'])
)
const canReadPrompt = computed(() => Boolean(btnAuth['orderfood:prompt:read']))
const canUpdatePrompt = computed(() => Boolean(btnAuth['orderfood:prompt:update']))
const canTestPrompt = computed(() => Boolean(btnAuth['orderfood:prompt:test']))
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))
const canReadAIUsage = computed(() =>
  Boolean(btnAuth['orderfood:ai-usage:read'])
)
const requiredCapabilities = {
  dish_text_extract: 'text',
  recipe_image_extract: 'vision',
  dish_cover_create: 'image_generation',
  checkin_image_analyze: 'vision',
  meal_suggest: 'text',
  prep_sequence: 'text'
}
const requiredCapability = (code) => requiredCapabilities[code] || 'text'
const modelCapabilityLabel = (value) =>
  ({ text: '文字生成', vision: '图片理解', image_generation: '图片生成' })[value] || value
const promptModeLabel = (value) => (value === 'custom' ? '自定义' : value ? '平台默认' : '—')
const promptVariableLabel = (variable) =>
  `${variable.label || variable.name}（{{${variable.name}}}）${variable.required ? ' *' : ''}`
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const administratorLabel = (administrator) =>
  administrator?.nickname || administrator?.username || administrator?.id || '—'
const shortHash = (value) => (value ? `${String(value).slice(0, 10)}…` : '—')

const listLoading = ref(false)
const listError = ref('')
const items = ref([])
const workspaceVisible = ref(false)
const workspaceLoading = ref(false)
const workspaceError = ref('')
const workspace = ref(null)
const promptWorkspace = ref(null)
const promptLoading = ref(false)
const activeCode = ref('')
const activeTab = ref('config')
const modelOptions = ref([])
const configFormRef = ref()
const configSaving = ref(false)
const configEditing = ref(false)
const configForm = ref({})
const configRules = {
  primaryModelId: [{ required: true, message: '请选择主模型', trigger: 'change' }],
  reason: [
    { required: true, message: '请输入修改原因', trigger: 'blur' },
    { min: 2, max: 200, message: '原因长度为 2-200 字', trigger: 'blur' }
  ]
}
const promptFormRef = ref()
const promptSaving = ref(false)
const promptEditing = ref(false)
const promptValidating = ref(false)
const renderLoading = ref(false)
const testLoading = ref(false)
const validationResult = ref(null)
const renderResult = ref(null)
const testResult = ref(null)
const testVariablesText = ref('{}')
const promptForm = ref({})
const promptInsertTarget = ref('userPromptTemplate')
const promptRules = {
  reason: [
    { required: true, message: '请输入修改原因', trigger: 'blur' },
    { min: 2, max: 200, message: '原因长度为 2-200 字', trigger: 'blur' }
  ]
}
const allowedVariables = computed(() => promptWorkspace.value?.allowedVariables || [])
const fixedOutputSchemaText = computed(() =>
  JSON.stringify(promptWorkspace.value?.fixedOutputSchema || {}, null, 2)
)
const fixedValidationRules = computed(
  () => workspace.value?.config?.fixedValidationRules || []
)
const selectedModelLabel = computed(() => {
  const selected = modelOptions.value.find(
    (item) => item.id === configForm.value.primaryModelId
  )
  if (!selected) return configForm.value.primaryModelId || '未配置'
  return `${selected.name}（${selected.provider?.name || '未知供应商'}）`
})

const promptPayload = () => ({
  mode: promptForm.value.mode,
  ...(promptForm.value.mode === 'custom'
    ? {
        systemPrompt: promptForm.value.systemPrompt,
        userPromptTemplate: promptForm.value.userPromptTemplate
      }
    : {})
})
const parseVariables = () => {
  const parsed = JSON.parse(testVariablesText.value)
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    throw new Error('测试变量必须是 JSON 对象')
  }
  return Object.fromEntries(
    Object.entries(parsed).map(([key, value]) => [key, String(value)])
  )
}

const loadList = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getAICapabilityList({ sortBy: 'sortOrder', sortOrder: 'asc' }),
      'AI 能力列表加载失败'
    )
    items.value = data?.list || []
  } catch (error) {
    listError.value = getOrderFoodErrorMessage(error, 'AI 能力列表加载失败')
  } finally {
    listLoading.value = false
  }
}
const applyWorkspace = () => {
  const config = workspace.value?.config || {}
  configForm.value = {
    primaryModelId: config.primaryModelId || '',
    pointCost: Number(config.pointCost || 0),
    dailyLimitPerUser: Number(config.dailyLimitPerUser || 0),
    timeoutMs: Number(config.timeoutMs || 60000),
    freeQuotaPerDay: Number(config.freeQuotaPerDay || 0),
    reason: ''
  }
  configFormRef.value?.clearValidate()
}
const beginConfigEdit = () => {
  applyWorkspace()
  configEditing.value = true
}
const discardConfigChanges = () => {
  applyWorkspace()
  configEditing.value = false
}
const applyPromptWorkspace = () => {
  const current = promptWorkspace.value?.current
  const mode = current?.mode === 'custom' ? 'custom' : 'default'
  promptForm.value = {
    mode,
    systemPrompt: mode === 'custom'
      ? current?.systemPrompt || ''
      : promptWorkspace.value?.defaultSystemPrompt || '',
    userPromptTemplate: mode === 'custom'
      ? current?.userPromptTemplate || ''
      : promptWorkspace.value?.defaultUserPromptTemplate || '',
    reason: ''
  }
  const variables = {}
  for (const variable of allowedVariables.value) {
    if (variable.required) variables[variable.name] = variable.example || `测试${variable.name}`
  }
  testVariablesText.value = JSON.stringify(variables, null, 2)
  promptFormRef.value?.clearValidate()
}
const beginPromptEdit = () => {
  applyPromptWorkspace()
  promptEditing.value = true
}
const discardPromptChanges = () => {
  applyPromptWorkspace()
  promptEditing.value = false
}
const restoreDefaultPrompt = () => {
  promptForm.value.mode = 'default'
  promptForm.value.systemPrompt = promptWorkspace.value?.defaultSystemPrompt || ''
  promptForm.value.userPromptTemplate =
    promptWorkspace.value?.defaultUserPromptTemplate || ''
}
const handlePromptModeChange = (mode) => {
  if (mode === 'default') {
    restoreDefaultPrompt()
    return
  }
  const current = promptWorkspace.value?.current
  if (current?.mode === 'custom') {
    promptForm.value.systemPrompt = current.systemPrompt || ''
    promptForm.value.userPromptTemplate = current.userPromptTemplate || ''
  }
}
const insertPromptVariable = (name) => {
  if (!promptEditing.value || promptForm.value.mode !== 'custom') return
  const field = promptInsertTarget.value || 'userPromptTemplate'
  const token = `{{${name}}}`
  const current = String(promptForm.value[field] || '')
  promptForm.value[field] = `${current}${current && !current.endsWith('\n') ? '\n' : ''}${token}`
}
const loadPromptWorkspace = async () => {
  if (!canReadPrompt.value || !activeCode.value) return
  promptLoading.value = true
  try {
    promptWorkspace.value = unwrapOrderFoodResponse(
      await getAICapabilityPromptWorkspace(activeCode.value),
      '提示词加载失败'
    )
    applyPromptWorkspace()
    promptEditing.value = false
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '提示词加载失败'))
  } finally {
    promptLoading.value = false
  }
}
const loadWorkspace = async () => {
  workspaceLoading.value = true
  workspaceError.value = ''
  try {
    const [detailResponse, modelResponse] = await Promise.all([
      getAICapabilityDetail(activeCode.value),
      getAIModelList({
        page: 1,
        pageSize: 100,
        enabled: true,
        capability: requiredCapability(activeCode.value),
        sortBy: 'name',
        sortOrder: 'asc'
      })
    ])
    workspace.value = unwrapOrderFoodResponse(detailResponse, '能力配置加载失败')
    modelOptions.value = unwrapOrderFoodResponse(modelResponse, '模型列表加载失败')?.list || []
    applyWorkspace()
    configEditing.value = false
  } catch (error) {
    workspaceError.value = getOrderFoodErrorMessage(error, '能力配置加载失败')
  } finally {
    workspaceLoading.value = false
  }
}
const openWorkspace = async (code) => {
  activeCode.value = code
  activeTab.value = 'config'
  workspaceVisible.value = true
  workspace.value = null
  promptWorkspace.value = null
  configEditing.value = false
  promptEditing.value = false
  validationResult.value = null
  renderResult.value = null
  testResult.value = null
  await loadWorkspace()
}

const openCapabilityUsages = () => {
  if (!activeCode.value) return
  router.push({
    name: 'OrderFoodAIUsages',
    query: { capabilityCode: activeCode.value }
  })
}

const saveConfig = async () => {
  const valid = await configFormRef.value?.validate().catch(() => false)
  if (!valid) return
  const summary = [
    `主模型：${selectedModelLabel.value}`,
    `提示词：${promptModeLabel(workspace.value?.config?.promptMode)}，摘要 ${shortHash(workspace.value?.config?.promptHash)}`,
    `积分：${configForm.value.pointCost} / 次`,
    `每日免费：${configForm.value.freeQuotaPerDay} 次`,
    `每日上限：${configForm.value.dailyLimitPerUser || '不限制'}`,
    `超时：${configForm.value.timeoutMs} ms`
  ].join('\n')
  const confirmed = await ElMessageBox.confirm(
    summary,
    '确认保存并立即生效',
    {
      confirmButtonText: '确认保存',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => true).catch(() => false)
  if (!confirmed) return
  configSaving.value = true
  try {
    unwrapOrderFoodResponse(
      await updateAICapability(activeCode.value, {
        primaryModelId: configForm.value.primaryModelId,
        pointCost: configForm.value.pointCost,
        dailyLimitPerUser: configForm.value.dailyLimitPerUser,
        timeoutMs: configForm.value.timeoutMs,
        freeQuotaPerDay: configForm.value.freeQuotaPerDay,
        reason: configForm.value.reason,
        expectedVersion: workspace.value?.config?.version || 0
      }),
      '保存能力配置失败'
    )
    ElMessage.success('能力配置已保存并生效')
    await Promise.all([loadWorkspace(), loadList()])
    if (canReadPrompt.value) await loadPromptWorkspace()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('配置已发生变化，已重新加载')
      await loadWorkspace()
      if (canReadPrompt.value) await loadPromptWorkspace()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存能力配置失败'))
    }
  } finally {
    configSaving.value = false
  }
}
const validatePrompt = async () => {
  promptValidating.value = true
  try {
    validationResult.value = unwrapOrderFoodResponse(
      await validateAICapabilityPrompt(activeCode.value, { prompt: promptPayload() }),
      '提示词校验失败'
    )
    if (validationResult.value.valid) ElMessage.success('提示词校验通过')
    return validationResult.value
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '提示词校验失败'))
    return null
  } finally {
    promptValidating.value = false
  }
}
const renderPrompt = async () => {
  renderLoading.value = true
  try {
    renderResult.value = unwrapOrderFoodResponse(
      await renderAICapabilityPrompt(activeCode.value, {
        prompt: promptPayload(),
        variables: parseVariables()
      }),
      '提示词预览失败'
    )
    ElMessage.success('变量替换预览已生成')
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '提示词预览失败'))
  } finally {
    renderLoading.value = false
  }
}
const testPrompt = async () => {
  testLoading.value = true
  try {
    testResult.value = unwrapOrderFoodResponse(
      await testAICapabilityPrompt(activeCode.value, {
        prompt: promptPayload(),
        variables: parseVariables()
      }),
      '提示词测试失败'
    )
    ElMessage[testResult.value.success ? 'success' : 'warning'](
      testResult.value.safeSummary || '提示词测试完成'
    )
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '提示词测试失败'))
  } finally {
    testLoading.value = false
  }
}
const savePrompt = async () => {
  const valid = await promptFormRef.value?.validate().catch(() => false)
  if (!valid || !workspace.value?.config) return
  const validation = await validatePrompt()
  if (!validation?.valid) {
    ElMessage.warning('请先处理提示词校验问题')
    return
  }
  const confirmed = await ElMessageBox.confirm(
    [
      `主模型：${workspace.value.summary?.primaryModelName || '未配置'}`,
      `提示词模式：${promptModeLabel(promptForm.value.mode)}`,
      `提示词摘要：${validation.promptHash || '—'}`,
      `积分：${workspace.value.config.pointCost} / 次`,
      `每日免费：${workspace.value.config.freeQuotaPerDay} 次`,
      `每日上限：${workspace.value.config.dailyLimitPerUser || '不限制'}`
    ].join('\n'),
    '确认保存提示词并立即生效',
    {
      confirmButtonText: '确认保存',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => true).catch(() => false)
  if (!confirmed) return
  promptSaving.value = true
  try {
    unwrapOrderFoodResponse(
      await updateAICapabilityPrompt(activeCode.value, {
        prompt: promptPayload(),
        reason: promptForm.value.reason,
        expectedVersion: workspace.value.config.version
      }),
      '保存提示词失败'
    )
    ElMessage.success('提示词已保存并生效')
    await Promise.all([loadWorkspace(), loadList()])
    await loadPromptWorkspace()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('配置已发生变化，已重新加载')
      await loadWorkspace()
      await loadPromptWorkspace()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存提示词失败'))
    }
  } finally {
    promptSaving.value = false
  }
}

onMounted(async () => {
  await loadList()
  if (route.query.capabilityCode) {
    await openWorkspace(String(route.query.capabilityCode))
  }
})
</script>

<style scoped>
.table-header,
.action-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.action-row {
  justify-content: flex-start;
  margin: 0;
}

.section-toolbar {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 18px;
}

.flex-1 {
  flex: 1;
}

.prompt-load {
  padding: 36px;
  color: var(--el-text-color-secondary);
  text-align: center;
}

.prompt-variable.clickable {
  cursor: pointer;
}

.prompt-variable.clickable:hover {
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}

.form-hint {
  width: 100%;
  margin-top: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.form-inline-hint {
  margin-left: 10px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.fixed-validation-rules {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  background: var(--el-fill-color-light);
}

.fixed-validation-rule + .fixed-validation-rule {
  margin-top: 8px;
}

.fixed-validation-rule::before {
  margin-right: 8px;
  color: var(--el-color-success);
  content: '✓';
}

.prompt-preview {
  margin: 0;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.schema-preview {
  max-height: 220px;
  margin: 0;
  overflow: auto;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}
</style>
