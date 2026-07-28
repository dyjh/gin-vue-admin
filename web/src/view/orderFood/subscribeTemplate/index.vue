<template>
  <div class="order-food-list-page">
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="关键词">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            placeholder="模板名称或用途"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="业务场景">
          <el-select v-model="searchInfo.scene" clearable placeholder="全部" class="w-40">
            <el-option label="饭局最终结果" value="meal_status" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用状态">
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
          <div class="text-lg font-medium">订阅消息模板</div>
          <div class="mt-1 text-sm text-gray-500">
            第一版仅用于参与者主动授权后的饭局最终确认或取消结果。
          </div>
        </div>
        <el-button v-if="canCreate" type="primary" @click="openCreate">
          新增模板
        </el-button>
      </div>

      <el-alert
        title="新建模板固定为停用；字段映射有效后，使用独立启用操作。用户禁用、关闭点单和 AI 退积分不会发送订阅消息。"
        type="info"
        show-icon
        class="mb-4"
      />
      <el-alert
        v-if="listError"
        :title="listError"
        type="error"
        show-icon
        class="mb-4"
      >
        <template #default>
          <el-button link type="primary" @click="loadList">重新加载</el-button>
        </template>
      </el-alert>

      <el-table
        v-loading="listLoading"
        :data="tableData"
        row-key="id"
        :empty-text="listError ? '加载失败' : '暂无模板'"
      >
        <el-table-column label="模板" min-width="220" fixed="left">
          <template #default="{ row }">
            <div class="font-medium">{{ row.name }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.id }}</div>
          </template>
        </el-table-column>
        <el-table-column label="微信模板 ID" min-width="190">
          <template #default="{ row }">{{ row.wechatTemplateIdMasked }}</template>
        </el-table-column>
        <el-table-column label="场景" width="130">
          <template #default>饭局最终结果</template>
        </el-table-column>
        <el-table-column label="用途" min-width="200">
          <template #default="{ row }">{{ row.purpose }}</template>
        </el-table-column>
        <el-table-column label="字段映射摘要" min-width="260">
          <template #default="{ row }">
            {{ (row.fieldMappingSummary || []).join('、') || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="启用状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? '已启用' : '已停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发送数" prop="sendCount" width="90" align="right" />
        <el-table-column label="失败数" prop="failureCount" width="90" align="right" />
        <el-table-column label="版本" prop="version" width="80" align="right" />
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="310" fixed="right">
          <template #default="{ row }">
            <div class="table-row-actions">
              <el-button link type="primary" @click="openDetail(row.id)">查看详情</el-button>
            <el-button v-if="canUpdate" link type="primary" @click="openEdit(row.id)">
              编辑
            </el-button>
            <el-button
              v-if="canReadLogs"
              link
              type="primary"
              @click="openLogs(row.id)"
            >
              发送记录
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
                    @click="changeStatus(row)"
                  >
                    {{ row.enabled ? '停用' : '启用' }}
                  </el-dropdown-item>
                  <el-dropdown-item
                    v-if="canDelete"
                    :divided="canChangeStatus"
                    :disabled="row.enabled || row.logCount > 0"
                    :title="
                      row.enabled
                        ? '请先停用模板'
                        : row.logCount > 0
                          ? '已有发送记录，只能停用'
                          : ''
                    "
                    @click="removeTemplate(row)"
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
          @current-change="loadList"
        />
      </div>
    </div>

    <el-drawer
      v-model="detailVisible"
      title="订阅消息模板详情"
      size="700px"
      destroy-on-close
      @closed="detail = null"
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
          <el-descriptions :column="1" border>
            <el-descriptions-item label="模板名称">{{ detail.name }}</el-descriptions-item>
            <el-descriptions-item label="模板 ID">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="微信模板 ID">
              <span class="break-all">{{ detail.wechatTemplateId }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="业务场景">饭局最终结果</el-descriptions-item>
            <el-descriptions-item label="用途">{{ detail.purpose }}</el-descriptions-item>
            <el-descriptions-item label="启用状态">
              {{ detail.enabled ? '已启用' : '已停用' }}
            </el-descriptions-item>
            <el-descriptions-item label="字段映射">
              <el-table :data="mappingEntries(detail.fieldMappings)" size="small">
                <el-table-column label="业务字段" prop="businessField" min-width="150" />
                <el-table-column label="微信模板字段" prop="wechatField" min-width="180" />
              </el-table>
            </el-descriptions-item>
            <el-descriptions-item label="发送 / 失败">
              {{ detail.sendCount }} / {{ detail.failureCount }}
            </el-descriptions-item>
            <el-descriptions-item label="发送记录总数">
              {{ detail.logCount }}
            </el-descriptions-item>
            <el-descriptions-item label="版本">{{ detail.version }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(detail.createdAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatDateTime(detail.updatedAt) }}
            </el-descriptions-item>
          </el-descriptions>
          <div class="mt-4">
            <el-button
              v-if="canReadLogs"
              type="primary"
              plain
              @click="openLogs(detail.id)"
            >
              查看发送记录
            </el-button>
          </div>
        </template>
      </div>
    </el-drawer>

    <el-dialog
      v-model="editorVisible"
      :title="editorMode === 'create' ? '新增订阅消息模板' : '编辑订阅消息模板'"
      width="760px"
      destroy-on-close
      @closed="resetEditor"
    >
      <el-alert
        title="业务字段固定用于组装饭局最终结果；右侧填写微信公众平台模板中的字段名，例如 thing1、phrase2、time3。"
        type="info"
        show-icon
        class="mb-4"
      />
      <el-form
        ref="editorFormRef"
        :model="editorForm"
        :rules="editorRules"
        label-width="120px"
      >
        <el-form-item label="模板名称" prop="name">
          <el-input v-model.trim="editorForm.name" maxlength="60" show-word-limit />
        </el-form-item>
        <el-form-item label="微信模板 ID" prop="wechatTemplateId">
          <el-input
            v-model.trim="editorForm.wechatTemplateId"
            maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="业务场景">
          <el-input model-value="饭局最终结果（meal_status）" disabled />
        </el-form-item>
        <el-form-item label="用途" prop="purpose">
          <el-input
            v-model.trim="editorForm.purpose"
            type="textarea"
            :rows="3"
            maxlength="160"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="字段映射" prop="mappingRows">
          <div class="w-full">
            <div
              v-for="(row, index) in editorForm.mappingRows"
              :key="`${row.businessField}-${index}`"
              class="mapping-row"
            >
              <el-input
                v-model.trim="row.businessField"
                :disabled="row.required"
                maxlength="64"
                placeholder="业务字段"
              />
              <span class="text-gray-400">→</span>
              <el-input
                v-model.trim="row.wechatField"
                maxlength="40"
                placeholder="微信模板字段"
              />
              <el-button
                v-if="!row.required"
                link
                type="danger"
                @click="removeMapping(index)"
              >
                删除
              </el-button>
              <span v-else class="required-label">必填</span>
            </div>
            <el-button
              class="mt-2"
              :disabled="editorForm.mappingRows.length >= 20"
              @click="addMapping"
            >
              新增扩展字段
            </el-button>
          </div>
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
import {
  createSubscribeTemplate,
  deleteSubscribeTemplate,
  getSubscribeTemplateDetail,
  getSubscribeTemplateList,
  updateSubscribeTemplate,
  updateSubscribeTemplateStatus
} from '@/api/orderfood/message'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodSubscribeTemplates'
})

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const hasBtnPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))
const canCreate = computed(() =>
  hasBtnPermission('orderfood:subscribe-template:create', 'subscribeTemplateCreate')
)
const canUpdate = computed(() =>
  hasBtnPermission('orderfood:subscribe-template:update', 'subscribeTemplateUpdate')
)
const canChangeStatus = computed(() =>
  hasBtnPermission('orderfood:subscribe-template:status', 'subscribeTemplateStatus')
)
const canDelete = computed(() =>
  hasBtnPermission('orderfood:subscribe-template:delete', 'subscribeTemplateDelete')
)
const canReadLogs = computed(() =>
  hasBtnPermission('orderfood:subscribe-log:read', 'subscribeLogRead')
)

const createDefaultSearch = () => ({
  keyword: '',
  scene: '',
  enabled: undefined
})
const searchInfo = ref(createDefaultSearch())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const tableData = ref([])
const listLoading = ref(false)
const listError = ref('')

const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )

const loadList = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getSubscribeTemplateList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...searchInfo.value,
          sortBy: 'updatedAt',
          sortOrder: 'desc'
        })
      ),
      '订阅消息模板加载失败'
    )
    tableData.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    tableData.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '订阅消息模板加载失败')
  } finally {
    listLoading.value = false
  }
}

const onSubmit = () => {
  page.value = 1
  loadList()
}
const onReset = () => {
  searchInfo.value = createDefaultSearch()
  page.value = 1
  loadList()
}
const handleSizeChange = () => {
  page.value = 1
  loadList()
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref(null)
const openDetail = async (templateId) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getSubscribeTemplateDetail(templateId),
      '模板详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '模板详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const requiredMappingRows = () => [
  {
    businessField: 'mealName',
    label: '饭局名称',
    wechatField: '',
    required: true
  },
  {
    businessField: 'result',
    label: '最终结果',
    wechatField: '',
    required: true
  },
  {
    businessField: 'resultAt',
    label: '结果时间',
    wechatField: '',
    required: true
  }
]
const createDefaultEditor = () => ({
  id: '',
  name: '',
  wechatTemplateId: '',
  scene: 'meal_status',
  purpose: '',
  mappingRows: requiredMappingRows(),
  expectedVersion: undefined
})
const editorVisible = ref(false)
const editorMode = ref('create')
const editorLoading = ref(false)
const editorFormRef = ref(null)
const editorForm = ref(createDefaultEditor())

const mappingValidator = (_rule, rows, callback) => {
  if (!Array.isArray(rows) || rows.length < 3) {
    callback(new Error('请配置必需字段映射'))
    return
  }
  const businessFields = rows.map((row) => row.businessField?.trim())
  const wechatFields = rows.map((row) => row.wechatField?.trim())
  if (businessFields.some((value) => !value) || wechatFields.some((value) => !value)) {
    callback(new Error('业务字段和微信模板字段均不能为空'))
    return
  }
  if (new Set(businessFields).size !== businessFields.length) {
    callback(new Error('业务字段不能重复'))
    return
  }
  if (new Set(wechatFields).size !== wechatFields.length) {
    callback(new Error('同一个微信模板字段不能重复映射'))
    return
  }
  callback()
}
const editorRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  wechatTemplateId: [
    { required: true, message: '请输入微信模板 ID', trigger: 'blur' }
  ],
  purpose: [{ required: true, message: '请输入模板用途', trigger: 'blur' }],
  mappingRows: [{ validator: mappingValidator, trigger: 'blur' }]
}

const openCreate = () => {
  editorMode.value = 'create'
  editorForm.value = createDefaultEditor()
  editorVisible.value = true
}
const openEdit = async (templateId) => {
  editorMode.value = 'edit'
  editorVisible.value = true
  editorLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getSubscribeTemplateDetail(templateId),
      '模板详情加载失败'
    )
    const mappings = data.fieldMappings || {}
    const required = requiredMappingRows().map((row) => ({
      ...row,
      wechatField: mappings[row.businessField] || ''
    }))
    const requiredKeys = new Set(required.map((row) => row.businessField))
    const extensions = Object.entries(mappings)
      .filter(([key]) => !requiredKeys.has(key))
      .map(([businessField, wechatField]) => ({
        businessField,
        wechatField,
        required: false
      }))
    editorForm.value = {
      id: data.id,
      name: data.name,
      wechatTemplateId: data.wechatTemplateId,
      scene: data.scene,
      purpose: data.purpose,
      mappingRows: [...required, ...extensions],
      expectedVersion: data.version
    }
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '模板详情加载失败'))
    editorVisible.value = false
  } finally {
    editorLoading.value = false
  }
}
const resetEditor = () => {
  editorForm.value = createDefaultEditor()
  editorFormRef.value?.clearValidate()
}
const addMapping = () => {
  editorForm.value.mappingRows.push({
    businessField: '',
    wechatField: '',
    required: false
  })
}
const removeMapping = (index) => {
  editorForm.value.mappingRows.splice(index, 1)
}
const buildMappings = () =>
  Object.fromEntries(
    editorForm.value.mappingRows.map((row) => [
      row.businessField.trim(),
      row.wechatField.trim()
    ])
  )

const handleConflict = async (error, action) => {
  if (!isOrderFoodConflict(error)) return false
  ElMessage.warning(`${action}失败：模板版本或引用状态已变化，请刷新后重试`)
  await loadList()
  return true
}
const submitEditor = async () => {
  const valid = await editorFormRef.value?.validate().catch(() => false)
  if (!valid) return
  editorLoading.value = true
  try {
    const payload = {
      name: editorForm.value.name,
      wechatTemplateId: editorForm.value.wechatTemplateId,
      scene: 'meal_status',
      purpose: editorForm.value.purpose,
      fieldMappings: buildMappings()
    }
    if (editorMode.value === 'create') {
      unwrapOrderFoodResponse(
        await createSubscribeTemplate(payload),
        '模板创建失败'
      )
      ElMessage.success('模板已创建，当前为停用状态')
    } else {
      unwrapOrderFoodResponse(
        await updateSubscribeTemplate(editorForm.value.id, {
          ...payload,
          expectedVersion: editorForm.value.expectedVersion
        }),
        '模板更新失败'
      )
      ElMessage.success('模板配置已保存，启用状态未改变')
    }
    editorVisible.value = false
    await loadList()
  } catch (error) {
    if (!(await handleConflict(error, '保存'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '模板保存失败'))
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
  const enabled = !row.enabled
  try {
    const reason = await requestReason(
      enabled ? '启用订阅消息模板' : '停用订阅消息模板',
      enabled
        ? `确认启用“${row.name}”吗？同一场景只能启用一个模板。`
        : `确认停用“${row.name}”吗？停用后不再创建新的微信发送任务。`
    )
    unwrapOrderFoodResponse(
      await updateSubscribeTemplateStatus(row.id, {
        enabled,
        reason,
        expectedVersion: row.version
      }),
      enabled ? '模板启用失败' : '模板停用失败'
    )
    ElMessage.success(enabled ? '模板已启用' : '模板已停用')
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '切换状态'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '模板状态更新失败'))
    }
  }
}
const removeTemplate = async (row) => {
  if (row.enabled || Number(row.logCount) > 0) {
    ElMessage.warning('启用中或已有发送记录的模板不能删除')
    return
  }
  try {
    const reason = await requestReason(
      '删除订阅消息模板',
      `确认删除“${row.name}”吗？该操作仅允许无发送记录引用的停用模板。`
    )
    unwrapOrderFoodResponse(
      await deleteSubscribeTemplate(row.id, {
        reason,
        expectedVersion: row.version
      }),
      '模板删除失败'
    )
    ElMessage.success('模板已删除')
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '删除'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '模板删除失败'))
    }
  }
}

const openLogs = (templateId) => {
  router.push({ name: 'OrderFoodSubscribeLogs', query: { templateId } })
}
const mappingEntries = (mappings) =>
  Object.entries(mappings || {}).map(([businessField, wechatField]) => ({
    businessField,
    wechatField
  }))
const formatDateTime = (value) => (value ? formatDate(value) : '—')

loadList().then(() => {
  if (route.query.templateId) {
    openDetail(String(route.query.templateId))
  }
})
</script>

<style scoped>
.table-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.mapping-row {
  display: grid;
  grid-template-columns: minmax(160px, 1fr) auto minmax(180px, 1fr) 56px;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.required-label {
  color: var(--el-color-danger);
  font-size: 12px;
  text-align: center;
}

@media (max-width: 720px) {
  .table-header {
    flex-direction: column;
  }

  .mapping-row {
    grid-template-columns: 1fr;
  }
}
</style>
