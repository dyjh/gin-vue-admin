<template>
  <div>
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item :label="`${title}名称`">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            maxlength="40"
            :placeholder="`输入${title}名称`"
            @keyup.enter="onSubmit"
          />
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
          <div class="text-lg font-medium">{{ title }}管理</div>
          <div class="mt-1 text-sm text-gray-500">
            停用后不再出现在新建表单中，历史数据仍保留原值；有业务引用时不能删除。
          </div>
        </div>
        <div class="flex gap-2">
          <el-button
            v-if="canSort && !sortMode"
            :loading="sortLoading"
            @click="enterSortMode"
          >
            排序模式
          </el-button>
          <el-button
            v-if="sortMode"
            @click="cancelSortMode"
          >
            取消排序
          </el-button>
          <el-button
            v-if="sortMode"
            type="primary"
            :disabled="changedSortItems.length === 0"
            :loading="sortSaving"
            @click="saveSortOrder"
          >
            保存排序
          </el-button>
          <el-button v-if="canCreate && !sortMode" type="primary" @click="openCreate">
            新增{{ title }}
          </el-button>
        </div>
      </div>

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

      <div v-if="sortMode" v-loading="sortLoading" class="sort-panel">
        <div class="sort-help">
          按住左侧拖拽柄调整顺序，保存后整批生效；取消排序不会保存任何改动。
        </div>
        <div
          v-for="(item, index) in sortDraft"
          :key="item.id"
          class="sort-row"
          :class="{ 'is-dragging': draggedSortID === item.id }"
          draggable="true"
          @dragstart="startSortDrag($event, item.id)"
          @dragover.prevent
          @drop="dropSortItem(item.id)"
          @dragend="finishSortDrag"
        >
          <span class="sort-handle" title="拖拽排序">⋮⋮</span>
          <span class="sort-rank">{{ index + 1 }}</span>
          <div class="sort-content">
            <div class="font-medium">{{ item.name }}</div>
            <div class="text-xs text-gray-400">{{ item.id }}</div>
          </div>
          <el-tag :type="item.enabled ? 'success' : 'info'" size="small">
            {{ item.enabled ? '已启用' : '已停用' }}
          </el-tag>
          <span class="sort-reference">引用 {{ item.referenceCount }}</span>
        </div>
      </div>

      <el-table
        v-else
        v-loading="listLoading"
        :data="items"
        row-key="id"
        :empty-text="listError ? '加载失败' : `暂无${title}`"
      >
        <el-table-column :label="title" min-width="220" fixed="left">
          <template #default="{ row }">
            <div class="font-medium">{{ row.name }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.id }}</div>
          </template>
        </el-table-column>
        <el-table-column label="排序" width="150">
          <template #default="{ row }">
            <span>{{ row.sortOrder }}</span>
          </template>
        </el-table-column>
        <el-table-column label="业务引用" width="110" align="right">
          <template #default="{ row }">
            <el-tag :type="row.referenceCount > 0 ? 'warning' : 'info'" effect="plain">
              {{ row.referenceCount }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? '已启用' : '已停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="版本" prop="version" width="90" align="right" />
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <div class="catalog-actions">
              <el-button v-if="canUpdate" link type="primary" @click="openEdit(row)">
                编辑
              </el-button>
              <el-button
                v-if="canUpdate"
                link
                :type="row.enabled ? 'warning' : 'success'"
                @click="changeStatus(row)"
              >
                {{ row.enabled ? '停用' : '启用' }}
              </el-button>
              <el-tooltip
                v-if="canDelete"
                :disabled="row.referenceCount === 0"
                content="已有业务数据引用，只能停用"
                placement="top"
              >
                <span>
                  <el-button
                    link
                    type="danger"
                    :disabled="row.referenceCount > 0"
                    @click="removeItem(row)"
                  >
                    删除
                  </el-button>
                </span>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!sortMode" class="flex justify-end pt-4">
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

    <el-dialog
      v-model="editorVisible"
      :title="editorMode === 'create' ? `新增${title}` : `编辑${title}`"
      width="520px"
      destroy-on-close
      @closed="resetEditor"
    >
      <el-form
        ref="editorFormRef"
        v-loading="editorLoading"
        :model="editorForm"
        :rules="editorRules"
        label-width="90px"
      >
        <el-form-item
          :label="`${title}名称`"
          prop="name"
          :error="nameServerError"
        >
          <el-input
            ref="nameInputRef"
            v-model.trim="editorForm.name"
            maxlength="40"
            show-word-limit
            :placeholder="`请输入${title}名称`"
            @input="nameServerError = ''"
          />
        </el-form-item>
        <el-form-item label="排序值" prop="sortOrder">
          <el-input-number
            v-model="editorForm.sortOrder"
            :min="1"
            :max="999999"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item label="启用状态" prop="enabled">
          <el-switch
            v-model="editorForm.enabled"
          />
          <span class="ml-3 text-xs text-gray-400">
            停用后仅从新建表单中隐藏
          </span>
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
import { computed, nextTick, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import {
  createCatalogItem,
  deleteCatalogItem,
  getCatalogItemList,
  updateCatalogItem,
  updateCatalogSortOrder
} from '@/api/orderfood/catalog'
import {
  getOrderFoodErrorCode,
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

const props = defineProps({
  resource: {
    type: String,
    required: true
  },
  title: {
    type: String,
    required: true
  },
  permissionPrefix: {
    type: String,
    required: true
  }
})

const btnAuth = useBtnAuth()
const hasPermission = (suffix) =>
  Boolean(btnAuth[`${props.permissionPrefix}:${suffix}`])
const canCreate = computed(() => hasPermission('create'))
const canUpdate = computed(() => hasPermission('update'))
const canDelete = computed(() => hasPermission('delete'))
const canSort = computed(() => hasPermission('sort'))

const searchInfo = ref({
  keyword: '',
  enabled: undefined
})
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const items = ref([])
const listLoading = ref(false)
const listError = ref('')
const sortMode = ref(false)
const sortDraft = ref([])
const sortOriginal = ref({})
const sortLoading = ref(false)
const draggedSortID = ref('')
const sortSaving = ref(false)

const formatDateTime = (value) => (value ? formatDate(value) : '—')
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
      await getCatalogItemList(
        props.resource,
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...searchInfo.value,
          sortBy: 'sortOrder',
          sortOrder: 'asc'
        })
      ),
      `${props.title}列表加载失败`
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    items.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(
      error,
      `${props.title}列表加载失败`
    )
  } finally {
    listLoading.value = false
  }
}

const onSubmit = () => {
  cancelSortMode()
  page.value = 1
  loadList()
}
const onReset = () => {
  cancelSortMode()
  searchInfo.value = { keyword: '', enabled: undefined }
  page.value = 1
  loadList()
}
const handleSizeChange = () => {
  page.value = 1
  loadList()
}

const createDefaultEditor = () => ({
  id: '',
  name: '',
  sortOrder: 1,
  enabled: true,
  expectedVersion: undefined
})
const editorVisible = ref(false)
const editorMode = ref('create')
const editorLoading = ref(false)
const editorFormRef = ref(null)
const nameInputRef = ref(null)
const nameServerError = ref('')
const editorForm = ref(createDefaultEditor())
const editorRules = {
  name: [
    { required: true, message: `请输入${props.title}名称`, trigger: 'blur' },
    { max: 40, message: '名称最多 40 个字符', trigger: 'blur' }
  ],
  sortOrder: [
    { required: true, message: '请输入排序值', trigger: 'change' },
    { type: 'number', min: 1, message: '排序值必须大于 0', trigger: 'change' }
  ]
}

const openCreate = () => {
  editorMode.value = 'create'
  editorForm.value = createDefaultEditor()
  nameServerError.value = ''
  editorVisible.value = true
}
const openEdit = (row) => {
  editorMode.value = 'edit'
  editorForm.value = {
    id: row.id,
    name: row.name,
    sortOrder: row.sortOrder,
    enabled: row.enabled,
    expectedVersion: row.version
  }
  nameServerError.value = ''
  editorVisible.value = true
}
const resetEditor = () => {
  editorForm.value = createDefaultEditor()
  nameServerError.value = ''
  editorFormRef.value?.clearValidate()
}

const handleConflict = async (error, action) => {
  if (!isOrderFoodConflict(error)) return false
  ElMessage.warning(`${action}失败：数据版本已变化，请刷新后重试`)
  await loadList()
  return true
}

const submitEditor = async () => {
  const valid = await editorFormRef.value?.validate().catch(() => false)
  if (!valid) return
  editorLoading.value = true
  try {
    const payload = {
      name: editorForm.value.name.trim(),
      sortOrder: editorForm.value.sortOrder,
      enabled: editorForm.value.enabled
    }
    if (editorMode.value === 'create') {
      unwrapOrderFoodResponse(
        await createCatalogItem(props.resource, payload),
        `${props.title}创建失败`
      )
      ElMessage.success(`${props.title}已创建`)
    } else {
      unwrapOrderFoodResponse(
        await updateCatalogItem(props.resource, editorForm.value.id, {
          ...payload,
          expectedVersion: editorForm.value.expectedVersion
        }),
        `${props.title}保存失败`
      )
      ElMessage.success(`${props.title}已保存`)
    }
    editorVisible.value = false
    await loadList()
  } catch (error) {
    if (getOrderFoodErrorCode(error) === 20903) {
      nameServerError.value = `该${props.title}名称已存在`
      await nextTick()
      nameInputRef.value?.focus()
    } else if (!(await handleConflict(error, '保存'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, `${props.title}保存失败`))
    }
  } finally {
    editorLoading.value = false
  }
}

const changeStatus = async (row) => {
  const nextEnabled = !row.enabled
  try {
    await ElMessageBox.confirm(
      nextEnabled
        ? `确认启用“${row.name}”吗？`
        : `确认停用“${row.name}”吗？历史数据仍会显示该${props.title}。`,
      nextEnabled ? `启用${props.title}` : `停用${props.title}`,
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: nextEnabled ? 'success' : 'warning'
      }
    )
    unwrapOrderFoodResponse(
      await updateCatalogItem(props.resource, row.id, {
        name: row.name,
        sortOrder: row.sortOrder,
        enabled: nextEnabled,
        expectedVersion: row.version
      }),
      `${props.title}状态更新失败`
    )
    ElMessage.success(nextEnabled ? `${props.title}已启用` : `${props.title}已停用`)
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '更新状态'))) {
      ElMessage.error(
        getOrderFoodErrorMessage(error, `${props.title}状态更新失败`)
      )
    }
  }
}

const reasonValidator = (value) => {
  const length = value?.trim().length || 0
  return length >= 4 && length <= 200 ? true : '请输入 4-200 字删除原因'
}
const removeItem = async (row) => {
  if (row.referenceCount > 0) {
    ElMessage.warning(`该${props.title}已有业务引用，只能停用`)
    return
  }
  try {
    const result = await ElMessageBox.prompt(
      `确认删除“${row.name}”吗？删除后不可恢复。`,
      `删除${props.title}`,
      {
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputPlaceholder: '请输入 4-200 字删除原因',
        inputValidator: reasonValidator
      }
    )
    unwrapOrderFoodResponse(
      await deleteCatalogItem(props.resource, row.id, {
        reason: result.value.trim(),
        expectedVersion: row.version
      }),
      `${props.title}删除失败`
    )
    ElMessage.success(`${props.title}已删除`)
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '删除'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, `${props.title}删除失败`))
    }
  }
}

const changedSortItems = computed(() =>
  sortDraft.value
    .filter(
      (item) =>
        Number(item.sortOrder) !== Number(sortOriginal.value[item.id]?.sortOrder)
    )
    .map((item) => ({
      id: item.id,
      sortOrder: Number(item.sortOrder),
      expectedVersion: sortOriginal.value[item.id]?.version
    }))
)

const enterSortMode = async () => {
  sortLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getCatalogItemList(props.resource, {
        page: 1,
        pageSize: 100,
        sortBy: 'sortOrder',
        sortOrder: 'asc'
      }),
      `${props.title}排序数据加载失败`
    )
    if (Number(data?.total || 0) > 100) {
      ElMessage.warning(`${props.title}超过 100 项，暂不能整批拖拽排序`)
      return
    }
    sortDraft.value = (data?.list || []).map((item) => ({ ...item }))
    sortOriginal.value = Object.fromEntries(
      sortDraft.value.map((item) => [
        item.id,
        { sortOrder: item.sortOrder, version: item.version }
      ])
    )
    sortMode.value = true
  } catch (error) {
    ElMessage.error(
      getOrderFoodErrorMessage(error, `${props.title}排序数据加载失败`)
    )
  } finally {
    sortLoading.value = false
  }
}

const cancelSortMode = () => {
  sortMode.value = false
  sortDraft.value = []
  sortOriginal.value = {}
  draggedSortID.value = ''
}

const startSortDrag = (event, id) => {
  draggedSortID.value = id
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/plain', id)
}

const dropSortItem = (targetID) => {
  const sourceID = draggedSortID.value
  if (!sourceID || sourceID === targetID) return
  const sourceIndex = sortDraft.value.findIndex((item) => item.id === sourceID)
  const targetIndex = sortDraft.value.findIndex((item) => item.id === targetID)
  if (sourceIndex < 0 || targetIndex < 0) return
  const next = [...sortDraft.value]
  const [moved] = next.splice(sourceIndex, 1)
  next.splice(targetIndex, 0, moved)
  sortDraft.value = next.map((item, index) => ({
    ...item,
    sortOrder: index + 1
  }))
}

const finishSortDrag = () => {
  draggedSortID.value = ''
}

const saveSortOrder = async () => {
  if (changedSortItems.value.length === 0) return
  if (changedSortItems.value.some((item) => !Number.isInteger(item.sortOrder) || item.sortOrder < 1)) {
    ElMessage.warning('排序值必须是大于 0 的整数')
    return
  }
  sortSaving.value = true
  try {
    const result = unwrapOrderFoodResponse(
      await updateCatalogSortOrder(props.resource, {
        items: changedSortItems.value
      }),
      `${props.title}排序保存失败`
    )
    ElMessage.success(`已更新 ${result?.updated || changedSortItems.value.length} 条排序`)
    cancelSortMode()
    await loadList()
  } catch (error) {
    if (!(await handleConflict(error, '保存排序'))) {
      ElMessage.error(
        getOrderFoodErrorMessage(error, `${props.title}排序保存失败`)
      )
    }
  } finally {
    sortSaving.value = false
  }
}

loadList()
</script>

<style scoped>
.table-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.catalog-actions {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  white-space: nowrap;
}

.catalog-actions :deep(.el-button) {
  margin-left: 0;
}

.sort-panel {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  overflow: hidden;
}

.sort-help {
  padding: 12px 16px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color);
}

.sort-row {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 64px;
  padding: 8px 16px;
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-lighter);
  cursor: grab;
}

.sort-row:last-child {
  border-bottom: 0;
}

.sort-row.is-dragging {
  opacity: 0.45;
}

.sort-handle {
  color: var(--el-text-color-placeholder);
  font-size: 20px;
  line-height: 1;
  letter-spacing: -4px;
  user-select: none;
}

.sort-rank {
  width: 32px;
  color: var(--el-text-color-secondary);
  text-align: right;
}

.sort-content {
  flex: 1;
  min-width: 0;
}

.sort-reference {
  width: 72px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  text-align: right;
}
</style>
