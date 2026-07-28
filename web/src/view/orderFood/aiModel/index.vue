<template>
  <div class="order-food-list-page">
    <AdvancedSearchPanel>
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
    </AdvancedSearchPanel>

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
            <div class="table-row-actions">
              <el-button link type="primary" @click="openDetail(row)">查看详情</el-button>
            <el-button v-if="canUpdate" link type="primary" @click="openEdit(row)">
              编辑
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
      width="780px"
      top="4vh"
      class="model-editor-dialog"
      destroy-on-close
      @closed="resetEditor"
    >
      <template #header>
        <div class="model-editor-header">
          <div class="model-editor-title">
            {{ editorMode === 'create' ? '新增模型' : '编辑模型' }}
          </div>
          <div class="model-editor-subtitle">
            {{
              editorMode === 'create'
                ? '从供应商目录选择模型，并补充能力与计费信息'
                : '调整模型能力、上下文与计费参数'
            }}
          </div>
        </div>
      </template>

      <div class="model-editor-content">
        <div
          class="model-editor-status"
          :class="{ 'is-edit': editorMode === 'edit' }"
          role="status"
        >
          <span class="model-editor-status-icon" aria-hidden="true">
            {{ editorMode === 'create' ? '停' : '启' }}
          </span>
          <div>
            <div class="model-editor-status-title">
              {{
                editorMode === 'create'
                  ? '创建后默认为停用状态'
                  : '本次编辑不会改变启用状态'
              }}
            </div>
            <div class="model-editor-status-description">
              {{
                editorMode === 'create'
                  ? '保存并确认配置无误后，再从模型列表中手动启用。'
                  : '如需启用或停用模型，请返回列表使用独立的状态操作。'
              }}
            </div>
          </div>
        </div>

        <el-form
          ref="editorFormRef"
          :model="editorForm"
          :rules="editorRules"
          label-position="top"
          class="model-editor-form"
        >
          <section class="model-form-section">
            <div class="model-form-section-heading">
              <div class="model-form-section-title">基础配置</div>
              <div class="model-form-section-description">
                先选择供应商，再从实时目录中选择模型
              </div>
            </div>
            <div class="model-form-section-body model-basic-grid">
              <el-form-item label="供应商" prop="providerId">
                <el-select
                  v-model="editorForm.providerId"
                  filterable
                  class="w-full"
                  placeholder="请选择已配置的供应商"
                  @change="onEditorProviderChange"
                >
                  <el-option
                    v-for="provider in providers"
                    :key="provider.id"
                    :label="`${provider.name} · ${providerTypeLabel(provider.type)}${
                      provider.enabled ? '' : '（已停用）'
                    }`"
                    :value="provider.id"
                    :disabled="
                      editorMode === 'edit' &&
                      editorForm.wasEnabled &&
                      !provider.enabled
                    "
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="模型" prop="modelKey">
                <div class="w-full">
                  <el-select
                    v-model="editorForm.modelKey"
                    filterable
                    class="w-full"
                    placeholder="请先选择供应商，再选择模型"
                    :loading="providerModelOptionsLoading"
                    :disabled="!editorForm.providerId"
                    @change="onEditorModelChange"
                  >
                    <el-option
                      v-for="option in providerModelOptions"
                      :key="option.modelKey"
                      :label="providerModelOptionLabel(option)"
                      :value="option.modelKey"
                      :disabled="providerModelOptionDisabled(option)"
                    />
                  </el-select>
                  <div
                    v-if="providerModelOptionsError"
                    class="model-field-message is-error"
                  >
                    <span>{{ providerModelOptionsError }}</span>
                    <el-button
                      link
                      type="primary"
                      @click="loadProviderModelOptions"
                    >
                      重新加载
                    </el-button>
                  </div>
                  <div
                    v-else-if="
                      editorForm.providerId &&
                      !providerModelOptionsLoading &&
                      providerModelOptions.length === 0
                    "
                    class="model-field-message"
                  >
                    供应商未返回可用模型，请先检查 API Key、基础地址和连接状态。
                  </div>
                </div>
              </el-form-item>
            </div>
          </section>

          <section class="model-form-section">
            <div class="model-form-section-heading">
              <div class="model-form-section-title">能力参数</div>
              <div class="model-form-section-description">
                标记模型用途，并设置可选的上下文容量
              </div>
            </div>
            <div class="model-form-section-body model-capability-grid">
              <el-form-item label="模型能力" prop="capabilities">
                <el-checkbox-group
                  v-model="editorForm.capabilities"
                  class="model-capability-options"
                >
                  <el-checkbox
                    v-for="option in capabilityOptions"
                    :key="option.value"
                    :label="option.value"
                    border
                  >
                    {{ option.label }}
                  </el-checkbox>
                </el-checkbox-group>
              </el-form-item>
              <el-form-item label="上下文上限" prop="contextLength">
                <div class="model-context-control">
                  <el-input-number
                    v-model="editorForm.contextLength"
                    :min="1"
                    :max="10000000"
                    :step="1000"
                    controls-position="right"
                    class="model-context-input"
                    placeholder="选填"
                  />
                  <span class="model-field-unit">tokens</span>
                </div>
                <div class="model-field-hint">不确定时可留空</div>
              </el-form-item>
            </div>
          </section>

          <section class="model-form-section">
            <div class="model-form-section-heading">
              <div class="model-form-section-title">计费参数</div>
              <div class="model-form-section-description">
                统一按人民币填写，未配置的价格可留空
              </div>
            </div>
            <div class="model-form-section-body model-price-grid">
              <el-form-item
                label="输入价格"
                prop="inputPricePerMillionTokensCny"
              >
                <el-input
                  v-model.trim="editorForm.inputPricePerMillionTokensCny"
                  placeholder="例如 2.00"
                >
                  <template #append>元 / 百万 tokens</template>
                </el-input>
              </el-form-item>
              <el-form-item
                label="输出价格"
                prop="outputPricePerMillionTokensCny"
              >
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
            </div>
          </section>

          <section class="model-form-section">
            <div class="model-form-section-heading">
              <div class="model-form-section-title">补充信息</div>
              <div class="model-form-section-description">
                记录模型版本、用途或配置说明
              </div>
            </div>
            <div class="model-form-section-body">
              <el-form-item label="备注" class="model-remark-field">
                <el-input
                  v-model.trim="editorForm.remark"
                  type="textarea"
                  :rows="3"
                  maxlength="300"
                  show-word-limit
                  resize="none"
                  placeholder="选填，便于团队理解该模型的配置用途"
                />
              </el-form-item>
            </div>
          </section>
        </el-form>
      </div>
      <template #footer>
        <div class="model-editor-footer">
          <el-button @click="editorVisible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="editorLoading"
            @click="submitEditor"
          >
            {{ editorMode === 'create' ? '创建模型' : '保存修改' }}
          </el-button>
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
  createAIModel,
  deleteAIModel,
  getAIModelDetail,
  getAIModelList,
  getAIModelProviderOptions,
  getAIProviderModelOptions,
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
  hasBtnPermission('orderfood:model:create', 'modelCreate')
)
const canUpdate = computed(() =>
  hasBtnPermission('orderfood:model:update', 'modelUpdate')
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
      await getAIModelProviderOptions(),
      '供应商选项加载失败'
    )
    providers.value = (Array.isArray(data) ? data : []).filter((provider) =>
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
  originalProviderId: '',
  originalName: '',
  originalModelKey: '',
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
const providerModelOptions = ref([])
const providerModelOptionsLoading = ref(false)
const providerModelOptionsError = ref('')
let providerModelOptionsRequestID = 0

const currentProviderModelOption = (option) =>
  editorMode.value === 'edit' &&
  editorForm.value.providerId === editorForm.value.originalProviderId &&
  option.modelKey === editorForm.value.originalModelKey

const providerModelOptionDisabled = (option) =>
  Boolean(option.configured) && !currentProviderModelOption(option)

const providerModelOptionLabel = (option) => {
  const label =
    option.name && option.name !== option.modelKey
      ? `${option.name} · ${option.modelKey}`
      : option.modelKey
  if (currentProviderModelOption(option)) return `${label}（当前配置）`
  return option.configured ? `${label}（已配置）` : label
}

const mergeCurrentProviderModelOption = (options) => {
  const result = [...options]
  if (
    editorMode.value === 'edit' &&
    editorForm.value.providerId === editorForm.value.originalProviderId &&
    editorForm.value.originalModelKey &&
    !result.some((option) => option.modelKey === editorForm.value.originalModelKey)
  ) {
    result.unshift({
      name: editorForm.value.originalName || editorForm.value.originalModelKey,
      modelKey: editorForm.value.originalModelKey,
      configured: true
    })
  }
  return result
}

const loadProviderModelOptions = async () => {
  const providerId = editorForm.value.providerId
  const requestID = ++providerModelOptionsRequestID
  providerModelOptions.value = []
  providerModelOptionsError.value = ''
  if (!providerId) {
    providerModelOptionsLoading.value = false
    return
  }
  providerModelOptionsLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getAIProviderModelOptions(providerId),
      '模型选项加载失败'
    )
    if (requestID !== providerModelOptionsRequestID) return
    providerModelOptions.value = mergeCurrentProviderModelOption(
      Array.isArray(data) ? data : []
    )
  } catch (error) {
    if (requestID !== providerModelOptionsRequestID) return
    providerModelOptions.value = mergeCurrentProviderModelOption([])
    providerModelOptionsError.value = getOrderFoodErrorMessage(
      error,
      '模型选项加载失败'
    )
  } finally {
    if (requestID === providerModelOptionsRequestID) {
      providerModelOptionsLoading.value = false
    }
  }
}

const onEditorProviderChange = () => {
  editorForm.value.name = ''
  editorForm.value.modelKey = ''
  editorFormRef.value?.clearValidate('modelKey')
  loadProviderModelOptions()
}

const onEditorModelChange = (modelKey) => {
  const isOriginal =
    editorForm.value.providerId === editorForm.value.originalProviderId &&
    modelKey === editorForm.value.originalModelKey
  editorForm.value.name = isOriginal
    ? editorForm.value.originalName
    : modelKey
}

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
  modelKey: [{ required: true, message: '请选择模型', trigger: 'change' }],
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
  providerModelOptionsRequestID++
  providerModelOptions.value = []
  providerModelOptionsError.value = ''
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
      originalProviderId: data.provider?.id || '',
      originalName: data.name,
      originalModelKey: data.modelKey,
      capabilities: [...(data.capabilities || [])],
      contextLength: data.contextLength ?? undefined,
      inputPricePerMillionTokensCny: data.inputPricePerMillionTokensCny || '',
      outputPricePerMillionTokensCny: data.outputPricePerMillionTokensCny || '',
      imagePricePerUnitCny: data.imagePricePerUnitCny || '',
      remark: data.remark || '',
      wasEnabled: Boolean(data.enabled),
      expectedVersion: data.version
    }
    await loadProviderModelOptions()
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '模型详情加载失败'))
    editorVisible.value = false
  } finally {
    editorLoading.value = false
  }
}

const resetEditor = () => {
  providerModelOptionsRequestID++
  providerModelOptions.value = []
  providerModelOptionsError.value = ''
  providerModelOptionsLoading.value = false
  editorForm.value = createDefaultEditor()
  editorFormRef.value?.clearValidate()
}

const normalizeOptional = (value) => (value === '' ? null : value)
const buildModelPayload = () => ({
  providerId: editorForm.value.providerId,
  name: editorForm.value.name || editorForm.value.modelKey,
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

:deep(.model-editor-dialog) {
  max-width: calc(100vw - 32px);
  border-radius: 12px;
  overflow: hidden;
}

:deep(.model-editor-dialog .el-dialog__header) {
  margin-right: 0;
  padding: 22px 28px 18px;
  border-bottom: 1px solid #e8edf4;
}

:deep(.model-editor-dialog .el-dialog__headerbtn) {
  top: 12px;
  right: 14px;
}

:deep(.model-editor-dialog .el-dialog__body) {
  max-height: calc(100vh - 210px);
  padding: 0;
  overflow-y: auto;
  background: #f6f8fb;
}

:deep(.model-editor-dialog .el-dialog__footer) {
  padding: 14px 28px;
  border-top: 1px solid #e8edf4;
  background: #ffffff;
}

.model-editor-header {
  padding-right: 36px;
}

.model-editor-title {
  color: #172033;
  font-size: 18px;
  font-weight: 600;
  line-height: 26px;
}

.model-editor-subtitle {
  margin-top: 3px;
  color: #778197;
  font-size: 13px;
  line-height: 20px;
}

.model-editor-content {
  padding: 20px 28px 24px;
}

.model-editor-status {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
  padding: 13px 15px;
  border: 1px solid #d8e5ff;
  border-radius: 9px;
  background: #f1f6ff;
}

.model-editor-status.is-edit {
  border-color: #dfe5ec;
  background: #f4f6f8;
}

.model-editor-status-icon {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 7px;
  background: #dce9ff;
  color: #3268c9;
  font-size: 12px;
  font-weight: 600;
}

.model-editor-status.is-edit .model-editor-status-icon {
  background: #e4e8ed;
  color: #596579;
}

.model-editor-status-title {
  color: #26354f;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.model-editor-status-description {
  margin-top: 1px;
  color: #68758b;
  font-size: 12px;
  line-height: 18px;
}

.model-editor-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.model-form-section {
  border: 1px solid #e4e9f0;
  border-radius: 10px;
  background: #ffffff;
}

.model-form-section-heading {
  display: flex;
  align-items: baseline;
  gap: 10px;
  padding: 14px 16px 10px;
}

.model-form-section-title {
  flex: 0 0 auto;
  color: #28364d;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.model-form-section-description {
  color: #8a94a6;
  font-size: 12px;
  line-height: 18px;
}

.model-form-section-body {
  padding: 2px 16px 0;
}

.model-basic-grid,
.model-price-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 16px;
}

.model-capability-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(220px, 0.8fr);
  column-gap: 22px;
}

.model-price-grid > :last-child {
  grid-column: 1 / 2;
}

:deep(.model-editor-form .el-form-item) {
  margin-bottom: 16px;
}

:deep(.model-editor-form .el-form-item__label) {
  height: auto;
  margin-bottom: 6px;
  padding: 0;
  color: #4a566b;
  font-size: 13px;
  font-weight: 500;
  line-height: 20px;
}

:deep(.model-editor-form .el-input__wrapper),
:deep(.model-editor-form .el-select__wrapper),
:deep(.model-editor-form .el-textarea__inner) {
  box-shadow: 0 0 0 1px #dce2ea inset;
}

:deep(.model-editor-form .el-input__wrapper:hover),
:deep(.model-editor-form .el-select__wrapper:hover),
:deep(.model-editor-form .el-textarea__inner:hover) {
  box-shadow: 0 0 0 1px #aeb9c8 inset;
}

.model-field-message {
  margin-top: 6px;
  color: #7d8797;
  font-size: 12px;
  line-height: 18px;
}

.model-field-message.is-error {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e25050;
}

.model-capability-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

:deep(.model-capability-options .el-checkbox) {
  height: 34px;
  margin-right: 0;
  padding: 0 12px;
  border-radius: 7px;
}

:deep(.model-capability-options .el-checkbox.is-bordered.is-checked) {
  border-color: #78a2ef;
  background: #f2f6ff;
}

.model-context-control {
  display: flex;
  align-items: center;
  width: 100%;
}

.model-context-input {
  width: 100%;
}

.model-field-unit {
  margin-left: 8px;
  color: #6f7a8d;
  font-size: 12px;
}

.model-field-hint {
  margin-top: 5px;
  color: #97a0af;
  font-size: 12px;
  line-height: 18px;
}

.model-remark-field {
  margin-bottom: 16px;
}

.model-editor-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 720px) {
  .table-header {
    flex-direction: column;
  }

  .model-editor-content {
    padding: 16px;
  }

  .model-form-section-heading {
    display: block;
  }

  .model-form-section-description {
    margin-top: 2px;
  }

  .model-basic-grid,
  .model-capability-grid,
  .model-price-grid {
    grid-template-columns: 1fr;
  }

  .model-price-grid > :last-child {
    grid-column: auto;
  }

  :deep(.model-editor-dialog .el-dialog__header) {
    padding-right: 20px;
    padding-left: 20px;
  }

  :deep(.model-editor-dialog .el-dialog__footer) {
    padding-right: 20px;
    padding-left: 20px;
  }
}
</style>
