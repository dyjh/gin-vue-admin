<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="模型">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            placeholder="名称或模型标识"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="供应商">
          <el-select
            v-model="searchInfo.providerId"
            clearable
            filterable
            placeholder="全部"
            class="w-52"
          >
            <el-option
              v-for="provider in providers"
              :key="provider.id"
              :label="`${provider.name} · ${providerTypeLabel(provider.type)}`"
              :value="provider.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模型能力">
          <el-select
            v-model="searchInfo.capability"
            clearable
            placeholder="全部"
            class="w-36"
          >
            <el-option
              v-for="option in capabilityOptions"
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
    </div>

    <div class="gva-table-box">
      <div class="table-header">
        <div>
          <div class="text-lg font-medium">AI 模型</div>
          <div class="mt-1 text-sm text-gray-500">
            维护模型标识、文本/图片能力、上下文上限与人民币价格参数。
          </div>
        </div>
        <el-button v-if="canCreate" type="primary" @click="openCreate">
          新增模型
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
          <el-button link type="primary" @click="loadModels">重新加载</el-button>
        </template>
      </el-alert>

      <el-table
        v-loading="listLoading"
        :data="models"
        row-key="id"
        :empty-text="listError ? '加载失败' : '暂无模型'"
      >
        <el-table-column label="模型" min-width="210" fixed="left">
          <template #default="{ row }">
            <div class="font-medium">{{ row.name }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.modelKey }}</div>
          </template>
        </el-table-column>
        <el-table-column label="供应商" min-width="190">
          <template #default="{ row }">
            <div>{{ row.provider?.name || '—' }}</div>
            <div class="text-xs text-gray-400">
              {{ providerTypeLabel(row.provider?.type) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="模型能力" min-width="220">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-1">
              <el-tag
                v-for="capability in row.capabilities || []"
                :key="capability"
                effect="plain"
              >
                {{ capabilityLabel(capability) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="上下文" width="120" align="right">
          <template #default="{ row }">
            {{ row.contextLength ? formatInteger(row.contextLength) : '—' }}
          </template>
        </el-table-column>
        <el-table-column label="价格摘要" min-width="230">
          <template #default="{ row }">
            <div
              v-for="line in priceSummaryLines(row)"
              :key="line"
              class="text-xs text-gray-500"
            >
              {{ line }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="启用能力引用" width="130" align="right">
          <template #default="{ row }">{{ row.enabledCapabilityCount }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? '已启用' : '已停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">查看详情</el-button>
            <el-button v-if="canUpdate" link type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-dropdown
              v-if="canChangeStatus || canDelete"
              trigger="click"
              class="ml-3"
            >
              <el-button link type="primary">更多</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item
                    v-if="canChangeStatus"
                    :disabled="Boolean(modelStatusBlockedReason(row))"
                    :title="modelStatusBlockedReason(row)"
                    @click="changeStatus(row)"
                  >
                    {{ row.enabled ? '停用' : '启用' }}
                  </el-dropdown-item>
                  <el-dropdown-item
                    v-if="canDelete"
                    :divided="canChangeStatus"
                    :disabled="Boolean(modelDeleteBlockedReason(row))"
                    :title="modelDeleteBlockedReason(row)"
                    @click="removeModel(row)"
                  >
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
          @current-change="loadModels"
        />
      </div>
    </div>

    <el-drawer v-model="detailVisible" title="模型详情" size="680px" destroy-on-close>
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
            <el-descriptions-item label="模型名称">{{ detail.name }}</el-descriptions-item>
            <el-descriptions-item label="模型标识">
              {{ detail.modelKey }}
            </el-descriptions-item>
            <el-descriptions-item label="供应商">
              {{ detail.provider?.name || '—' }} ·
              {{ providerTypeLabel(detail.provider?.type) }}
            </el-descriptions-item>
            <el-descriptions-item label="模型能力">
              <div class="flex flex-wrap gap-1">
                <el-tag
                  v-for="capability in detail.capabilities || []"
                  :key="capability"
                  effect="plain"
                >
                  {{ capabilityLabel(capability) }}
                </el-tag>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="上下文上限">
              {{
                detail.contextLength
                  ? `${formatInteger(detail.contextLength)} tokens`
                  : '未设置'
              }}
            </el-descriptions-item>
            <el-descriptions-item label="输入价格">
              {{ tokenPriceLabel(detail.inputPricePerMillionTokensCny) }}
            </el-descriptions-item>
            <el-descriptions-item label="输出价格">
              {{ tokenPriceLabel(detail.outputPricePerMillionTokensCny) }}
            </el-descriptions-item>
            <el-descriptions-item label="图片价格">
              {{ imagePriceLabel(detail.imagePricePerUnitCny) }}
            </el-descriptions-item>
            <el-descriptions-item label="启用能力引用">
              {{ detail.enabledCapabilityCount }}
            </el-descriptions-item>
            <el-descriptions-item label="总引用">
              {{ detail.referenceCount }}
            </el-descriptions-item>
            <el-descriptions-item label="历史调用">
              {{ detail.usageCount }}
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              {{ detail.enabled ? '已启用' : '已停用' }}
            </el-descriptions-item>
            <el-descriptions-item label="备注">{{ detail.remark || '—' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(detail.createdAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatDateTime(detail.updatedAt) }}
            </el-descriptions-item>
              </el-descriptions>
              <div class="mt-4">
                <div class="mb-2 font-medium">引用该模型的能力</div>
                <el-table
                  :data="detail.referencedCapabilities || []"
                  size="small"
                  empty-text="暂无能力引用"
                >
                  <el-table-column prop="capabilityName" label="能力名称" />
                  <el-table-column prop="capabilityCode" label="能力编码" />
                </el-table>
              </div>
              <div class="mt-4 flex flex-wrap gap-2">
                <el-button
                  v-if="canReadProvider && detail.provider?.id"
                  @click="openModelProvider"
                >
                  查看供应商
                </el-button>
                <el-button v-if="canReadAIUsage" @click="openModelUsages">
                  查看调用记录
                </el-button>
              </div>
              <el-alert
                v-if="detail.enabledCapabilityCount > 0"
                :title="`该模型正被 ${detail.enabledCapabilityCount} 项启用能力引用，解除引用前不能停用或删除。`"
                type="warning"
                show-icon
                class="mt-4"
              />
            </el-tab-pane>
            <el-tab-pane v-if="canReadAudit" label="操作审计" lazy>
              <AuditPanel target-type="ai_model" :target-id="detail.id" />
            </el-tab-pane>
          </el-tabs>
        </template>
      </div>
    </el-drawer>

    <el-dialog
      v-model="editorVisible"
      :title="editorMode === 'create' ? '新增模型' : '编辑模型'"
      width="720px"
      destroy-on-close
      @closed="resetEditor"
    >
      <el-alert
        v-if="editorMode === 'edit'"
        title="编辑只更新模型配置，不改变启用状态；启停请使用列表中的独立操作。"
        type="info"
        show-icon
        class="mb-4"
      />
      <el-form
        ref="editorFormRef"
        :model="editorForm"
        :rules="editorRules"
        label-width="130px"
      >
        <el-form-item label="供应商" prop="providerId">
          <el-select
            v-model="editorForm.providerId"
            filterable
            class="w-full"
            placeholder="请选择已配置的供应商"
          >
            <el-option
              v-for="provider in providers"
              :key="provider.id"
              :label="`${provider.name} · ${providerTypeLabel(provider.type)}${
                provider.enabled ? '' : '（已停用）'
              }`"
              :value="provider.id"
              :disabled="
                editorMode === 'edit' && editorForm.wasEnabled && !provider.enabled
              "
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模型名称" prop="name">
          <el-input v-model.trim="editorForm.name" maxlength="60" show-word-limit />
        </el-form-item>
        <el-form-item label="模型标识" prop="modelKey">
          <el-input
            v-model.trim="editorForm.modelKey"
            maxlength="120"
            placeholder="供应商使用的模型 ID"
          />
        </el-form-item>
        <el-form-item label="模型能力" prop="capabilities">
          <el-checkbox-group v-model="editorForm.capabilities">
            <el-checkbox
              v-for="option in capabilityOptions"
              :key="option.value"
              :label="option.value"
            >
              {{ option.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="上下文上限" prop="contextLength">
          <el-input-number
            v-model="editorForm.contextLength"
            :min="1"
            :max="10000000"
            :step="1000"
            controls-position="right"
          />
          <span class="ml-2 text-sm text-gray-500">tokens</span>
        </el-form-item>
        <el-divider content-position="left">价格参数（人民币）</el-divider>
        <el-form-item label="输入价格" prop="inputPricePerMillionTokensCny">
          <el-input
            v-model.trim="editorForm.inputPricePerMillionTokensCny"
            placeholder="例如 2.00"
          >
            <template #append>元 / 百万 tokens</template>
          </el-input>
        </el-form-item>
        <el-form-item label="输出价格" prop="outputPricePerMillionTokensCny">
          <el-input
            v-model.trim="editorForm.outputPricePerMillionTokensCny"
            placeholder="例如 8.00"
          >
            <template #append>元 / 百万 tokens</template>
          </el-input>
        </el-form-item>
        <el-form-item label="图片价格" prop="imagePricePerUnitCny">
          <el-input
            v-model.trim="editorForm.imagePricePerUnitCny"
            placeholder="例如 0.20"
          >
            <template #append>元 / 张</template>
          </el-input>
        </el-form-item>
        <el-alert
          v-if="editorMode === 'create'"
          class="mb-4"
          type="info"
          :closable="false"
          title="新模型创建后默认停用，确认配置后再手动启用。"
        />
        <el-form-item label="备注">
          <el-input
            v-model.trim="editorForm.remark"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button type="primary" :loading="editorLoading" @click="submitEditor">
          保存
        </el-button>
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
  createAIModel,
  deleteAIModel,
  getAIModelDetail,
  getAIModelList,
  getAIProviderList,
  updateAIModel,
  updateAIModelStatus
} from '@/api/orderfood/ai'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodAIModel'
})

const route = useRoute()
const router = useRouter()
const providerTypeOptions = [
  { value: 'deepseek', label: 'DeepSeek' },
  { value: 'bailian', label: '千问（阿里云百炼）' },
  { value: 'openai', label: 'GPT（OpenAI）' }
]
const supportedProviderTypes = new Set(
  providerTypeOptions.map((option) => option.value)
)
const capabilityOptions = [
  { value: 'text', label: '文本' },
  { value: 'vision', label: '图片理解' },
  { value: 'image_generation', label: '图片生成' }
]

const btnAuth = useBtnAuth()
const hasBtnPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))
const canReadProvider = computed(() =>
  hasBtnPermission('orderfood:provider:read', 'providerRead')
)
const canCreate = computed(() =>
  hasBtnPermission('orderfood:model:create', 'modelCreate') &&
  canReadProvider.value
)
const canUpdate = computed(() =>
  hasBtnPermission('orderfood:model:update', 'modelUpdate') &&
  canReadProvider.value
)
const canChangeStatus = computed(() =>
  hasBtnPermission('orderfood:model:status', 'modelStatus')
)
const canDelete = computed(() =>
  hasBtnPermission('orderfood:model:delete', 'modelDelete')
)
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))
const canReadAIUsage = computed(() =>
  Boolean(btnAuth['orderfood:ai-usage:read'])
)

const providers = ref([])
const loadProviders = async () => {
  try {
    const data = unwrapOrderFoodResponse(
      await getAIProviderList({
        page: 1,
        pageSize: 100,
        sortBy: 'name',
        sortOrder: 'asc'
      }),
      '供应商选项加载失败'
    )
    providers.value = (data?.list || []).filter((provider) =>
      supportedProviderTypes.has(provider.type)
    )
  } catch (error) {
    providers.value = []
    if (canCreate.value || canUpdate.value) {
      ElMessage.warning(getOrderFoodErrorMessage(error, '供应商选项加载失败'))
    }
  }
}

const createDefaultSearch = () => ({
  keyword: '',
  providerId: String(route.query.providerId || ''),
  capability: '',
  enabled: undefined
})
const searchInfo = ref(createDefaultSearch())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const models = ref([])
const listLoading = ref(false)
const listError = ref('')

const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )

const loadModels = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getAIModelList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...searchInfo.value,
          sortBy: 'updatedAt',
          sortOrder: 'desc'
        })
      ),
      '模型列表加载失败'
    )
    models.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    models.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '模型列表加载失败')
  } finally {
    listLoading.value = false
  }
}

const onSubmit = () => {
  page.value = 1
  loadModels()
}
const onReset = () => {
  searchInfo.value = createDefaultSearch()
  page.value = 1
  loadModels()
}
const handleSizeChange = () => {
  page.value = 1
  loadModels()
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref(null)

const loadDetail = async (modelId) => {
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getAIModelDetail(modelId),
      '模型详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '模型详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const openDetail = (row) => {
  detailVisible.value = true
  loadDetail(row.id)
}

const openModelProvider = () => {
  if (!detail.value?.provider?.id) return
  router.push({
    name: 'OrderFoodAiProviders',
    query: { providerId: detail.value.provider.id }
  })
}

const openModelUsages = () => {
  if (!detail.value?.id) return
  router.push({
    name: 'OrderFoodAIUsages',
    query: { modelId: detail.value.id }
  })
}

const createDefaultEditor = () => ({
  id: '',
  providerId: '',
  name: '',
  modelKey: '',
  capabilities: ['text'],
  contextLength: undefined,
  inputPricePerMillionTokensCny: '',
  outputPricePerMillionTokensCny: '',
  imagePricePerUnitCny: '',
  remark: '',
  wasEnabled: false,
  expectedVersion: undefined
})
const editorVisible = ref(false)
const editorMode = ref('create')
const editorLoading = ref(false)
const editorFormRef = ref(null)
const editorForm = ref(createDefaultEditor())
const pricePattern = /^(0|[1-9][0-9]*)(\.[0-9]{1,8})?$/
const optionalPriceValidator = (_rule, value, callback) => {
  if (!value || pricePattern.test(value)) {
    callback()
    return
  }
  callback(new Error('请输入非负数，最多 8 位小数'))
}
const editorRules = {
  providerId: [{ required: true, message: '请选择供应商', trigger: 'change' }],
  name: [{ required: true, message: '请输入模型名称', trigger: 'blur' }],
  modelKey: [{ required: true, message: '请输入模型标识', trigger: 'blur' }],
  capabilities: [
    {
      type: 'array',
      required: true,
      min: 1,
      message: '至少选择一项模型能力',
      trigger: 'change'
    }
  ],
  inputPricePerMillionTokensCny: [
    { validator: optionalPriceValidator, trigger: 'blur' }
  ],
  outputPricePerMillionTokensCny: [
    { validator: optionalPriceValidator, trigger: 'blur' }
  ],
  imagePricePerUnitCny: [{ validator: optionalPriceValidator, trigger: 'blur' }]
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
      await getAIModelDetail(row.id),
      '模型详情加载失败'
    )
    editorForm.value = {
      id: data.id,
      providerId: data.provider?.id || '',
      name: data.name,
      modelKey: data.modelKey,
      capabilities: [...(data.capabilities || [])],
      contextLength: data.contextLength ?? undefined,
      inputPricePerMillionTokensCny: data.inputPricePerMillionTokensCny || '',
      outputPricePerMillionTokensCny: data.outputPricePerMillionTokensCny || '',
      imagePricePerUnitCny: data.imagePricePerUnitCny || '',
      remark: data.remark || '',
      wasEnabled: Boolean(data.enabled),
      expectedVersion: data.version
    }
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '模型详情加载失败'))
    editorVisible.value = false
  } finally {
    editorLoading.value = false
  }
}

const resetEditor = () => {
  editorForm.value = createDefaultEditor()
  editorFormRef.value?.clearValidate()
}

const normalizeOptional = (value) => (value === '' ? null : value)
const buildModelPayload = () => ({
  providerId: editorForm.value.providerId,
  name: editorForm.value.name,
  modelKey: editorForm.value.modelKey,
  capabilities: editorForm.value.capabilities,
  contextLength: editorForm.value.contextLength || null,
  inputPricePerMillionTokensCny: normalizeOptional(
    editorForm.value.inputPricePerMillionTokensCny
  ),
  outputPricePerMillionTokensCny: normalizeOptional(
    editorForm.value.outputPricePerMillionTokensCny
  ),
  imagePricePerUnitCny: normalizeOptional(editorForm.value.imagePricePerUnitCny),
  remark: editorForm.value.remark || null
})

const handleConflict = async (error, action) => {
  if (!isOrderFoodConflict(error)) return false
  ElMessage.warning(`${action}失败：模型配置或引用关系已变化，请刷新后重试`)
  await loadModels()
  return true
}

const submitEditor = async () => {
  const valid = await editorFormRef.value?.validate().catch(() => false)
  if (!valid) return
  editorLoading.value = true
  try {
    const payload = buildModelPayload()
    if (editorMode.value === 'create') {
      unwrapOrderFoodResponse(await createAIModel(payload), '模型创建失败')
      ElMessage.success('模型已创建')
    } else {
      // enabled 使用详情中读取的原值且页面不可编辑，普通更新不会改变启用状态。
      unwrapOrderFoodResponse(
        await updateAIModel(editorForm.value.id, {
          ...payload,
          expectedVersion: editorForm.value.expectedVersion
        }),
        '模型更新失败'
      )
      ElMessage.success('模型配置已保存，启用状态未改变')
    }
    editorVisible.value = false
    await loadModels()
  } catch (error) {
    if (!(await handleConflict(error, '保存'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '模型保存失败'))
    }
  } finally {
    editorLoading.value = false
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
    let current = row
    if (!nextEnabled) {
      current = unwrapOrderFoodResponse(
        await getAIModelDetail(row.id),
        '模型引用关系加载失败'
      )
      const references = current.referencedCapabilities || []
      if (references.length > 0) {
        await ElMessageBox.alert(
          references
            .map(
              (reference) =>
                `${reference.capabilityName}（${reference.capabilityCode}）`
            )
            .join('、'),
          '无法停用：请先解除以下能力引用',
          { confirmButtonText: '知道了', type: 'warning' }
        )
        return
      }
    }
    const reason = await requestReason(
      nextEnabled ? '启用模型' : '停用模型',
      nextEnabled
        ? `确认启用“${row.name}”吗？`
        : `确认停用“${row.name}”吗？停用后不能承接新的调用。`
    )
    unwrapOrderFoodResponse(
      await updateAIModelStatus(current.id, {
        enabled: nextEnabled,
        reason,
        expectedVersion: current.version
      }),
      nextEnabled ? '模型启用失败' : '模型停用失败'
    )
    ElMessage.success(nextEnabled ? '模型已启用' : '模型已停用')
    await loadModels()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '切换状态'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '模型状态更新失败'))
    }
  }
}

const removeModel = async (row) => {
  if (row.referenceCount > 0) {
    ElMessage.warning(
      `该模型仍有 ${row.enabledCapabilityCount || 0} 项能力引用和 ${
        row.usageCount || 0
      } 条历史调用，解除引用后才能删除`
    )
    return
  }
  try {
    const reason = await requestReason(
      '删除模型',
      `确认删除“${row.name}”吗？存在能力配置或调用记录引用时服务端会拒绝删除。`
    )
    unwrapOrderFoodResponse(
      await deleteAIModel(row.id, {
        reason,
        expectedVersion: row.version
      }),
      '模型删除失败'
    )
    ElMessage.success('模型已删除')
    await loadModels()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '删除'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '模型删除失败'))
    }
  }
}

const providerTypeLabel = (type) =>
  providerTypeOptions.find((option) => option.value === type)?.label || '—'
const capabilityLabel = (capability) =>
  capabilityOptions.find((option) => option.value === capability)?.label || capability
const tokenPriceLabel = (value) =>
  value === null || typeof value === 'undefined'
    ? '未设置'
    : `${value} 元 / 百万 tokens`
const imagePriceLabel = (value) =>
  value === null || typeof value === 'undefined' ? '未设置' : `${value} 元 / 张`
const priceSummaryLines = (row) => {
  const lines = []
  if (row.inputPricePerMillionTokensCny !== null) {
    lines.push(`输入 ${tokenPriceLabel(row.inputPricePerMillionTokensCny)}`)
  }
  if (row.outputPricePerMillionTokensCny !== null) {
    lines.push(`输出 ${tokenPriceLabel(row.outputPricePerMillionTokensCny)}`)
  }
  if (row.imagePricePerUnitCny !== null) {
    lines.push(`图片 ${imagePriceLabel(row.imagePricePerUnitCny)}`)
  }
  return lines.length > 0 ? lines : ['未设置价格']
}
const formatInteger = (value) => Number(value).toLocaleString('zh-CN')
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const modelStatusBlockedReason = (row) => {
  if (row.enabled) return ''
  if (!row.provider?.enabled) return '所属供应商已停用'
  if (!row.provider?.credentialConfigured) return '所属供应商未配置凭据'
  if (!row.provider?.lastConnectionTest?.success) {
    return '所属供应商当前配置尚未通过连接测试'
  }
  return ''
}
const modelDeleteBlockedReason = (row) =>
  Number(row.referenceCount || 0) > 0
    ? `仍有 ${row.referenceCount} 条能力配置或调用记录引用`
    : ''

Promise.all([loadProviders(), loadModels()])
if (route.query.modelId) {
  detailVisible.value = true
  loadDetail(String(route.query.modelId))
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

@media (max-width: 720px) {
  .table-header {
    flex-direction: column;
  }
}
</style>
