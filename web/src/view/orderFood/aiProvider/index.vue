<template>
  <div class="order-food-list-page">
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="供应商">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            placeholder="名称"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchInfo.type" clearable placeholder="全部" class="w-52">
            <el-option
              v-for="option in providerTypeOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchInfo.enabled"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="已启用" :value="true" />
            <el-option label="已停用" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </AdvancedSearchPanel>

    <div class="gva-table-box">
      <div class="table-header">
        <div>
          <div class="text-lg font-medium">AI 供应商</div>
          <div class="mt-1 text-sm text-gray-500">
            首期仅接入 DeepSeek、千问（阿里云百炼）和 GPT（OpenAI）。
          </div>
        </div>
        <el-button v-if="canCreate" type="primary" @click="openCreate">
          新增供应商
        </el-button>
      </div>

      <el-alert
        v-if="listError"
        :title="listError"
        type="error"
        show-icon
        class="mb-4"
      >
        <template #default>
          <el-button link type="primary" @click="loadProviders">重新加载</el-button>
        </template>
      </el-alert>

      <el-table
        v-loading="listLoading"
        :data="providers"
        row-key="id"
        class="provider-table"
        :empty-text="listError ? '加载失败' : '暂无供应商'"
      >
        <el-table-column label="供应商" min-width="190" fixed="left">
          <template #default="{ row }">
            <div class="provider-name-cell">
              <div class="provider-name-cell__name">{{ row.name }}</div>
              <div class="provider-name-cell__type">{{ providerTypeLabel(row.type) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="基础地址" min-width="300">
          <template #default="{ row }">
            <el-tooltip
              :content="safeBaseURL(row.baseUrl)"
              placement="top"
              :show-after="300"
            >
              <span class="safe-url">{{ safeBaseURL(row.baseUrl) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="API Key" width="100">
          <template #default="{ row }">
            <el-tag :type="row.credentialConfigured ? 'success' : 'warning'">
              {{ row.credentialConfigured ? '已配置' : '未配置' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="请求策略" width="110">
          <template #default="{ row }">
            {{ row.timeoutMs }} ms
          </template>
        </el-table-column>
        <el-table-column label="引用" min-width="210">
          <template #default="{ row }">
            <div class="provider-reference-summary">
              <span>模型 <b>{{ row.modelCount }}</b></span>
              <span>能力 <b>{{ row.enabledCapabilityCount }}</b></span>
              <span>调用 <b>{{ row.usageCount }}</b></span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? '已启用' : '已停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近连接测试" width="130">
          <template #default="{ row }">
            <el-tag
              v-if="row.lastConnectionTest"
              :type="row.lastConnectionTest.success ? 'success' : 'danger'"
            >
              {{ row.lastConnectionTest.success ? '成功' : '失败' }}
            </el-tag>
            <span v-else class="text-gray-400">尚未测试</span>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="292" fixed="right">
          <template #default="{ row }">
            <div class="table-row-actions">
              <el-button link type="primary" @click="openDetail(row)">查看详情</el-button>
              <el-button v-if="canUpdate" link type="primary" @click="openEdit(row)">
                编辑
              </el-button>
              <el-button
                v-if="canTest"
                link
                type="primary"
                :loading="activeActionId === row.id && actionType === 'test'"
                @click="testConnection(row)"
              >
                测试连接
              </el-button>
              <el-dropdown
                v-if="canChangeStatus || canDelete"
                trigger="click"
              >
                <el-button link type="primary">更多</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-if="canChangeStatus"
                      :disabled="!row.enabled && Boolean(providerStatusBlockedReason(row))"
                      :title="providerStatusBlockedReason(row)"
                      @click="changeStatus(row)"
                    >
                      {{ row.enabled ? '停用' : '启用' }}
                    </el-dropdown-item>
                    <el-dropdown-item
                      v-if="canDelete"
                      :divided="canChangeStatus"
                      :disabled="Boolean(providerDeleteBlockedReason(row))"
                      :title="providerDeleteBlockedReason(row)"
                      @click="removeProvider(row)"
                    >
                      删除
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="flex justify-end pt-4">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="loadProviders"
        />
      </div>
    </div>

    <el-drawer
      v-model="detailVisible"
      title="供应商详情"
      size="620px"
      destroy-on-close
    >
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert
          v-if="detailError"
          :title="detailError"
          type="error"
          show-icon
          class="mb-4"
        />
        <template v-if="detail">
          <el-tabs>
            <el-tab-pane label="详情">
              <el-descriptions :column="1" border>
            <el-descriptions-item label="名称">{{ detail.name }}</el-descriptions-item>
            <el-descriptions-item label="类型">
              {{ providerTypeLabel(detail.type) }}
            </el-descriptions-item>
            <el-descriptions-item label="基础地址">
              {{ safeBaseURL(detail.baseUrl) }}
            </el-descriptions-item>
            <el-descriptions-item label="API Key">
              <el-tag :type="detail.credentialConfigured ? 'success' : 'warning'">
                {{ detail.credentialConfigured ? '已配置' : '未配置' }}
              </el-tag>
              <span class="ml-2 text-xs text-gray-400">完整密钥不会在页面或接口中回显</span>
            </el-descriptions-item>
            <el-descriptions-item label="超时">
              {{ detail.timeoutMs }} ms
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              {{ detail.enabled ? '已启用' : '已停用' }}
            </el-descriptions-item>
            <el-descriptions-item label="模型引用">
              {{ detail.modelCount }}
            </el-descriptions-item>
            <el-descriptions-item label="启用能力引用">
              {{ detail.enabledCapabilityCount }}
            </el-descriptions-item>
            <el-descriptions-item label="历史调用引用">
              {{ detail.usageCount }}
            </el-descriptions-item>
            <el-descriptions-item label="备注">{{ detail.remark || '—' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(detail.createdAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatDateTime(detail.updatedAt) }}
            </el-descriptions-item>
              </el-descriptions>

              <div class="mt-4 flex flex-wrap gap-2">
                <el-button v-if="canReadModels" @click="openProviderModels">
                  查看引用模型
                </el-button>
                <el-button v-if="canReadAIUsage" @click="openProviderUsages">
                  查看调用记录
                </el-button>
              </div>

              <div class="mt-5 mb-2 font-medium">引用模型</div>
              <el-table :data="detail.referencedModels || []" size="small" border>
                <el-table-column prop="name" label="模型名称" min-width="150" />
                <el-table-column prop="modelKey" label="模型标识" min-width="180" />
                <el-table-column label="状态" width="90">
                  <template #default="{ row }">
                    <el-tag :type="row.enabled ? 'success' : 'info'">
                      {{ row.enabled ? '已启用' : '已停用' }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>

              <div class="mt-5 mb-2 font-medium">引用能力</div>
              <el-table :data="detail.referencedCapabilities || []" size="small" border>
                <el-table-column label="能力" min-width="170">
                  <template #default="{ row }">
                    {{ capabilityLabel(row.capabilityCode, row.capabilityName) }}
                  </template>
                </el-table-column>
                <el-table-column prop="modelName" label="主模型" min-width="160" />
              </el-table>

              <div class="mt-5 mb-2 font-medium">最近一次连接测试</div>
              <el-empty
                v-if="!detail.lastConnectionTest"
                description="尚未测试"
                :image-size="72"
              />
              <el-descriptions v-else :column="1" border>
            <el-descriptions-item label="结果">
              <el-tag :type="detail.lastConnectionTest.success ? 'success' : 'danger'">
                {{ detail.lastConnectionTest.success ? '连接成功' : '连接失败' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="耗时">
              {{ detail.lastConnectionTest.durationMs }} ms
            </el-descriptions-item>
            <el-descriptions-item label="说明">
              {{ detail.lastConnectionTest.safeMessage || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="测试时间">
              {{ formatDateTime(detail.lastConnectionTest.testedAt) }}
            </el-descriptions-item>
              </el-descriptions>
            </el-tab-pane>
            <el-tab-pane v-if="canReadAudit" label="操作审计" lazy>
              <AuditPanel target-type="ai_provider" :target-id="detail.id" />
            </el-tab-pane>
          </el-tabs>
        </template>
      </div>
    </el-drawer>

    <el-dialog
      v-model="editorVisible"
      :title="editorMode === 'create' ? '新增供应商' : '编辑供应商'"
      width="680px"
      top="6vh"
      class="provider-editor-dialog"
      destroy-on-close
      :close-on-click-modal="!editorLoading"
      :close-on-press-escape="!editorLoading"
      @closed="resetEditor"
    >
      <el-form
        ref="editorFormRef"
        :model="editorForm"
        :rules="editorRules"
        label-position="top"
        class="provider-editor-form"
      >
        <section class="provider-editor-section">
          <div class="provider-editor-section__heading">
            <span>基础信息</span>
            <small>设置供应商类型、显示名称和接口地址</small>
          </div>
          <div class="provider-editor-grid">
            <el-form-item label="供应商类型" prop="type">
              <el-select
                v-model="editorForm.type"
                :disabled="editorMode === 'edit'"
                class="w-full"
                @change="applyProviderDefaults"
              >
                <el-option
                  v-for="option in providerTypeOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <div class="form-tip">创建后不可修改</div>
            </el-form-item>
            <el-form-item label="名称" prop="name">
              <el-input
                v-model.trim="editorForm.name"
                maxlength="60"
                show-word-limit
                placeholder="用于后台识别"
              />
            </el-form-item>
            <el-form-item
              label="基础地址"
              prop="baseUrl"
              class="provider-editor-grid__full"
            >
              <el-input
                v-model.trim="editorForm.baseUrl"
                maxlength="300"
                placeholder="例如 https://api.deepseek.com"
              />
            </el-form-item>
          </div>
        </section>

        <section class="provider-editor-section">
          <div class="provider-editor-section__heading">
            <span>连接设置</span>
            <small>配置访问凭证和接口超时时间</small>
          </div>
          <div class="provider-editor-grid">
            <el-form-item
              v-if="editorMode === 'create' || canWriteCredential"
              :label="editorMode === 'create' ? 'API Key' : '更换 API Key'"
              :prop="editorMode === 'create' ? 'apiKey' : undefined"
            >
              <el-input
                v-model.trim="editorForm.apiKey"
                type="password"
                show-password
                autocomplete="new-password"
                maxlength="500"
                placeholder="请输入供应商 API Key"
              />
              <div class="form-tip">
                {{
                  editorMode === 'create'
                    ? '密钥将明文保存到数据库，保存后不再回显'
                    : '留空表示保持现有密钥；当前密钥不会回显'
                }}
              </div>
            </el-form-item>
            <el-form-item
              label="连接超时"
              prop="timeoutMs"
            >
              <div class="provider-timeout-field">
                <el-input-number
                  v-model="editorForm.timeoutMs"
                  :min="1000"
                  :max="120000"
                  :step="1000"
                  controls-position="right"
                />
                <span>毫秒</span>
              </div>
              <div class="form-tip">可设置 1,000–120,000 毫秒</div>
            </el-form-item>
          </div>

          <el-alert
            v-if="editorMode === 'create'"
            class="provider-state-note"
            type="info"
            show-icon
            :closable="false"
            title="创建后默认停用，请先完成连接测试，再从列表中启用。"
          />

          <el-form-item label="备注" prop="remark" class="provider-remark-field">
            <el-input
              v-model.trim="editorForm.remark"
              type="textarea"
              :rows="3"
              maxlength="300"
              show-word-limit
              placeholder="选填，可记录账号用途或配置说明"
            />
          </el-form-item>
        </section>
      </el-form>
      <template #footer>
        <div class="provider-editor-footer">
          <span>保存配置不会自动启用供应商</span>
          <div class="provider-editor-footer__actions">
            <el-button :disabled="editorLoading" @click="editorVisible = false">
              取消
            </el-button>
            <el-button type="primary" :loading="editorLoading" @click="submitEditor">
              保存
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import AuditPanel from '@/view/orderFood/components/AuditPanel.vue'
import {
  createAIProvider,
  deleteAIProvider,
  getAIProviderDetail,
  getAIProviderList,
  testAIProviderConnection,
  updateAIProvider,
  updateAIProviderStatus
} from '@/api/orderfood/ai'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodAIProvider'
})

const route = useRoute()
const router = useRouter()
const providerTypeOptions = [
  {
    value: 'deepseek',
    label: 'DeepSeek',
    defaultName: 'DeepSeek',
    defaultBaseURL: 'https://api.deepseek.com'
  },
  {
    value: 'bailian',
    label: '千问（阿里云百炼）',
    defaultName: '千问（阿里云百炼）',
    defaultBaseURL: 'https://dashscope.aliyuncs.com/compatible-mode/v1'
  },
  {
    value: 'openai',
    label: 'GPT（OpenAI）',
    defaultName: 'GPT（OpenAI）',
    defaultBaseURL: 'https://api.openai.com/v1'
  }
]
const supportedProviderTypes = new Set(
  providerTypeOptions.map((option) => option.value)
)

const btnAuth = useBtnAuth()
const hasBtnPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))
const canWriteCredential = computed(() =>
  hasBtnPermission(
    'orderfood:provider:credential-write',
    'providerCredentialWrite'
  )
)
const canCreate = computed(
  () =>
    hasBtnPermission('orderfood:provider:create', 'providerCreate') &&
    canWriteCredential.value
)
const canUpdate = computed(() =>
  hasBtnPermission('orderfood:provider:update', 'providerUpdate')
)
const canTest = computed(() =>
  hasBtnPermission('orderfood:provider:test', 'providerTest')
)
const canChangeStatus = computed(() =>
  hasBtnPermission('orderfood:provider:status', 'providerStatus')
)
const canDelete = computed(() =>
  hasBtnPermission('orderfood:provider:delete', 'providerDelete')
)
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))
const canReadModels = computed(() => Boolean(btnAuth['orderfood:model:read']))
const canReadAIUsage = computed(() =>
  Boolean(btnAuth['orderfood:ai-usage:read'])
)

const createDefaultSearch = () => ({
  keyword: '',
  type: '',
  enabled: undefined
})
const searchInfo = ref(createDefaultSearch())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const providers = ref([])
const listLoading = ref(false)
const listError = ref('')

const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )

const loadProviders = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getAIProviderList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...searchInfo.value,
          sortBy: 'updatedAt',
          sortOrder: 'desc'
        })
      ),
      '供应商列表加载失败'
    )
    providers.value = (data?.list || []).filter((item) =>
      supportedProviderTypes.has(item.type)
    )
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    providers.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '供应商列表加载失败')
  } finally {
    listLoading.value = false
  }
}

const onSubmit = () => {
  page.value = 1
  loadProviders()
}
const onReset = () => {
  searchInfo.value = createDefaultSearch()
  page.value = 1
  loadProviders()
}
const handleSizeChange = () => {
  page.value = 1
  loadProviders()
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref(null)

const loadDetail = async (providerId) => {
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getAIProviderDetail(providerId),
      '供应商详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '供应商详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const openDetail = (row) => {
  detailVisible.value = true
  loadDetail(row.id)
}

const openProviderModels = () => {
  if (!detail.value?.id) return
  router.push({
    name: 'OrderFoodAiModels',
    query: { providerId: detail.value.id }
  })
}

const openProviderUsages = () => {
  if (!detail.value?.id) return
  router.push({
    name: 'OrderFoodAIUsages',
    query: { providerId: detail.value.id }
  })
}

const createDefaultEditor = () => ({
  id: '',
  name: '',
  type: 'deepseek',
  baseUrl: 'https://api.deepseek.com',
  apiKey: '',
  timeoutMs: 30000,
  remark: '',
  expectedVersion: undefined
})
const editorVisible = ref(false)
const editorMode = ref('create')
const editorLoading = ref(false)
const editorFormRef = ref(null)
const editorForm = ref(createDefaultEditor())

const editorRules = {
  type: [{ required: true, message: '请选择供应商类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  baseUrl: [
    { required: true, message: '请输入基础地址', trigger: 'blur' },
    { type: 'url', message: '请输入有效的 HTTP 或 HTTPS 地址', trigger: 'blur' }
  ],
  apiKey: [
    { required: true, message: '请输入 API Key', trigger: 'blur' },
    { max: 500, message: 'API Key 不能超过 500 个字符', trigger: 'blur' },
    { pattern: /^\S+$/, message: 'API Key 不能包含空白字符', trigger: 'blur' }
  ],
  timeoutMs: [{ required: true, message: '请输入超时时间', trigger: 'change' }]
}

const applyProviderDefaults = (type) => {
  if (editorMode.value !== 'create') return
  const option = providerTypeOptions.find((item) => item.value === type)
  if (!option) return
  editorForm.value.name = option.defaultName
  editorForm.value.baseUrl = option.defaultBaseURL
}

const openCreate = () => {
  editorMode.value = 'create'
  editorForm.value = createDefaultEditor()
  editorVisible.value = true
}

const openEdit = async (row) => {
  editorMode.value = 'edit'
  editorVisible.value = true
  editorLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getAIProviderDetail(row.id),
      '供应商详情加载失败'
    )
    editorForm.value = {
      id: data.id,
      name: data.name,
      type: data.type,
      baseUrl: data.baseUrl,
      apiKey: '',
      timeoutMs: data.timeoutMs,
      remark: data.remark || '',
      expectedVersion: data.version
    }
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '供应商详情加载失败'))
    editorVisible.value = false
  } finally {
    editorLoading.value = false
  }
}

const resetEditor = () => {
  editorForm.value = createDefaultEditor()
  editorFormRef.value?.clearValidate()
}

const handleConflict = async (error, action) => {
  if (!isOrderFoodConflict(error)) return false
  ElMessage.warning(`${action}失败：供应商配置已变化，请刷新后重试`)
  await loadProviders()
  return true
}

const submitEditor = async () => {
  const valid = await editorFormRef.value?.validate().catch(() => false)
  if (!valid) return
  editorLoading.value = true
  try {
    if (editorMode.value === 'create') {
      unwrapOrderFoodResponse(
        await createAIProvider({
          name: editorForm.value.name,
          type: editorForm.value.type,
          baseUrl: editorForm.value.baseUrl,
          apiKey: editorForm.value.apiKey,
          timeoutMs: editorForm.value.timeoutMs,
          remark: editorForm.value.remark || null
        }),
        '供应商创建失败'
      )
      ElMessage.success('供应商已创建')
    } else {
      const payload = {
        name: editorForm.value.name,
        baseUrl: editorForm.value.baseUrl,
        timeoutMs: editorForm.value.timeoutMs,
        remark: editorForm.value.remark || null,
        expectedVersion: editorForm.value.expectedVersion
      }
      if (canWriteCredential.value && editorForm.value.apiKey) {
        payload.apiKey = editorForm.value.apiKey
      }
      unwrapOrderFoodResponse(
        await updateAIProvider(editorForm.value.id, payload),
        '供应商更新失败'
      )
      ElMessage.success('供应商配置已保存')
    }
    editorVisible.value = false
    await loadProviders()
  } catch (error) {
    if (!(await handleConflict(error, '保存'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '供应商保存失败'))
    }
  } finally {
    editorLoading.value = false
  }
}

const activeActionId = ref('')
const actionType = ref('')

const runRowAction = async (row, type, action) => {
  activeActionId.value = row.id
  actionType.value = type
  try {
    await action()
    await loadProviders()
    if (detailVisible.value && detail.value?.id === row.id) {
      await loadDetail(row.id)
    }
  } finally {
    activeActionId.value = ''
    actionType.value = ''
  }
}

const reasonValidator = (value) => {
  const length = value?.trim().length || 0
  return length >= 4 && length <= 200 ? true : '请输入 4-200 字操作原因'
}

const requestReason = async (title, message) => {
  const result = await ElMessageBox.prompt(message, title, {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputPlaceholder: '请输入 4-200 字操作原因',
    inputValidator: reasonValidator
  })
  return result.value.trim()
}

const changeStatus = async (row) => {
  const nextEnabled = !row.enabled
  try {
    if (!nextEnabled) {
      const current = unwrapOrderFoodResponse(
        await getAIProviderDetail(row.id),
        '供应商停用影响加载失败'
      )
      const enabledModels = (current.referencedModels || []).filter((item) => item.enabled)
      const capabilities = current.referencedCapabilities || []
      if (enabledModels.length || capabilities.length) {
        const modelNames = enabledModels.map((item) => item.name).join('、') || '无'
        const capabilityNames =
          capabilities
            .map((item) => capabilityLabel(item.capabilityCode, item.capabilityName))
            .join('、') || '无'
        await ElMessageBox.alert(
          `请先解除以下影响：启用模型：${modelNames}；能力：${capabilityNames}。`,
          '当前不能停用供应商',
          { confirmButtonText: '知道了', type: 'warning' }
        )
        return
      }
    }
    const reason = await requestReason(
      nextEnabled ? '启用供应商' : '停用供应商',
      nextEnabled
        ? `确认启用“${row.name}”吗？`
        : `确认停用“${row.name}”吗？停用后不能承接新的调用。`
    )
    await runRowAction(row, 'status', async () => {
      unwrapOrderFoodResponse(
        await updateAIProviderStatus(row.id, {
          enabled: nextEnabled,
          reason,
          expectedVersion: row.version
        }),
        nextEnabled ? '供应商启用失败' : '供应商停用失败'
      )
      ElMessage.success(nextEnabled ? '供应商已启用' : '供应商已停用')
    })
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '切换状态'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '供应商状态更新失败'))
    }
  }
}

const removeProvider = async (row) => {
  if (row.modelCount > 0 || row.enabledCapabilityCount > 0 || row.usageCount > 0) {
    ElMessage.warning('该供应商仍有模型、能力或历史调用引用，不能删除')
    return
  }
  try {
    const reason = await requestReason(
      '删除供应商',
      `确认删除“${row.name}”吗？只有从未被模型、能力或调用引用时才能删除。`
    )
    await runRowAction(row, 'delete', async () => {
      unwrapOrderFoodResponse(
        await deleteAIProvider(row.id, {
          reason,
          expectedVersion: row.version
        }),
        '供应商删除失败'
      )
      ElMessage.success('供应商已删除')
    })
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '删除'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '供应商删除失败'))
    }
  }
}

const testConnection = async (row) => {
  try {
    await ElMessageBox.confirm(
      `将使用“${row.name}”当前已保存的 API Key 测试连接，不会展示请求头或完整响应。`,
      '测试连接',
      {
        confirmButtonText: '开始测试',
        cancelButtonText: '取消',
        type: 'info'
      }
    )
    await runRowAction(row, 'test', async () => {
      const result = unwrapOrderFoodResponse(
        await testAIProviderConnection(row.id, {
          timeoutMs: Math.min(row.timeoutMs, 30000),
          expectedVersion: row.version
        }),
        '连接测试失败'
      )
      if (result.success) {
        ElMessage.success(`连接成功，耗时 ${result.durationMs} ms`)
      } else {
        ElMessage.error(result.safeMessage || '连接失败')
      }
    })
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '测试连接'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '连接测试失败'))
    }
  }
}

const providerTypeLabel = (type) =>
  providerTypeOptions.find((option) => option.value === type)?.label || '不支持的类型'
const capabilityLabel = (code, fallback) =>
  ({
    dish_text_extract: '菜品文本解析',
    recipe_image_extract: '菜谱长截图解析',
    dish_cover_create: '菜品封面生成',
    checkin_image_analyze: '打卡图片分析',
    preference_profile_summarize: '打卡偏好画像整理',
    meal_suggest: '不知道吃什么',
    prep_sequence: '饭局备菜顺序'
  })[code] || fallback || code || '—'
const safeBaseURL = (value) => {
  if (!value) return '—'
  try {
    const url = new URL(value)
    return `${url.protocol}//${url.host}${url.pathname}`
  } catch {
    return '地址格式异常'
  }
}
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const providerStatusBlockedReason = (row) => {
  if (row.enabled) {
    if (Number(row.enabledModelCount || 0) > 0) {
      return `仍有 ${row.enabledModelCount} 个已启用模型，请先停用模型`
    }
    if (Number(row.enabledCapabilityCount || 0) > 0) {
      return `仍被 ${row.enabledCapabilityCount} 项能力引用，请先解除引用`
    }
    return ''
  }
  if (!row.credentialConfigured) return '请先配置 API Key'
  if (!row.lastConnectionTest?.success) return '请先完成当前配置的连接测试并确保成功'
  return ''
}
const providerDeleteBlockedReason = (row) =>
  Number(row.modelCount || 0) > 0
    ? `仍有 ${row.modelCount} 个模型引用，只能保留或停用`
    : Number(row.usageCount || 0) > 0
      ? `已有 ${row.usageCount} 条历史调用引用，不能删除`
      : ''

loadProviders()
if (route.query.providerId) {
  detailVisible.value = true
  loadDetail(String(route.query.providerId))
}
</script>

<style scoped>
.table-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.safe-url {
  display: block;
  overflow: hidden;
  color: var(--el-text-color-regular);
  line-height: 1.55;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-name-cell {
  display: grid;
  gap: 5px;
  min-width: 0;
}

.provider-name-cell__name {
  overflow: hidden;
  color: var(--el-text-color-primary);
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-name-cell__type {
  overflow: hidden;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-reference-summary {
  display: flex;
  align-items: center;
  gap: 14px;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}

.provider-reference-summary b {
  margin-left: 2px;
  color: var(--el-text-color-primary);
  font-weight: 600;
}

.provider-table :deep(.el-table__body td.el-table__cell) {
  padding-block: 13px;
}

.provider-editor-section + .provider-editor-section {
  margin-top: 4px;
  padding-top: 20px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.provider-editor-section__heading {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 16px;
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
}

.provider-editor-section__heading small {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  font-weight: 400;
}

.provider-editor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 18px;
}

.provider-editor-grid__full {
  grid-column: 1 / -1;
}

.provider-editor-form :deep(.el-form-item) {
  margin-bottom: 18px;
}

.provider-editor-form :deep(.el-form-item__label) {
  height: auto;
  margin-bottom: 8px;
  padding: 0;
  color: var(--el-text-color-regular);
  font-weight: 500;
  line-height: 1.3;
}

.provider-editor-form :deep(.el-input__wrapper),
.provider-editor-form :deep(.el-select__wrapper),
.provider-editor-form :deep(.el-textarea__inner),
.provider-editor-form :deep(.el-input-number) {
  border-radius: 6px;
}

.provider-timeout-field {
  display: flex;
  align-items: center;
  gap: 10px;
}

.provider-timeout-field :deep(.el-input-number) {
  width: 210px;
}

.provider-timeout-field > span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.provider-state-note {
  margin: 0 0 18px;
  border: 1px solid var(--el-color-primary-light-8);
  background: var(--el-color-primary-light-9);
}

.provider-state-note :deep(.el-alert__title) {
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.5;
}

.provider-remark-field {
  margin-bottom: 0 !important;
}

.provider-editor-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.provider-editor-footer > span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.provider-editor-footer__actions {
  display: flex;
  gap: 10px;
}

.provider-editor-footer__actions :deep(.el-button) {
  min-width: 72px;
  margin-left: 0;
}
.form-tip {
  width: 100%;
  margin-top: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}

:global(.provider-editor-dialog) {
  max-width: calc(100vw - 32px);
  overflow: hidden;
  border-radius: 10px;
}

:global(.provider-editor-dialog .el-dialog__header) {
  margin-right: 0;
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

:global(.provider-editor-dialog .el-dialog__title) {
  color: var(--el-text-color-primary);
  font-size: 18px;
  font-weight: 600;
}

:global(.provider-editor-dialog .el-dialog__headerbtn) {
  top: 13px;
  right: 14px;
}

:global(.provider-editor-dialog .el-dialog__body) {
  max-height: calc(88vh - 132px);
  overflow-y: auto;
  padding: 20px 24px 4px;
}

:global(.provider-editor-dialog .el-dialog__footer) {
  padding: 14px 24px 16px;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-extra-light);
}

@media (max-width: 720px) {
  .table-header {
    flex-direction: column;
  }

  .provider-editor-grid {
    grid-template-columns: 1fr;
  }

  .provider-editor-grid__full {
    grid-column: auto;
  }

  .provider-editor-section__heading,
  .provider-editor-footer {
    align-items: flex-start;
    flex-direction: column;
  }

  .provider-editor-footer {
    gap: 12px;
  }

  .provider-editor-footer__actions {
    justify-content: flex-end;
    width: 100%;
  }

  :global(.provider-editor-dialog .el-dialog__body) {
    padding-inline: 18px;
  }

  :global(.provider-editor-dialog .el-dialog__header),
  :global(.provider-editor-dialog .el-dialog__footer) {
    padding-inline: 18px;
  }
}
</style>
