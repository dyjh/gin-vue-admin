<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="菜品">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            maxlength="60"
            placeholder="输入菜品名称"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="来源">
          <el-select
            v-model="searchInfo.sourceType"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="用户菜品" value="creator" />
            <el-option label="官方菜品" value="official" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源菜品 ID">
          <el-input
            v-model.trim="searchInfo.sourceDishId"
            clearable
            placeholder="输入来源菜品 ID"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchInfo.status"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="草稿" value="draft" />
            <el-option label="已发布" value="published" />
            <el-option label="已下线" value="offline" />
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
          <div class="text-lg font-medium">推荐精选</div>
          <div class="mt-1 text-sm text-gray-500">
            首页精选最多同时发布 5 条；来源菜品创建后不可替换，下线不影响用户已经复制的副本。
          </div>
        </div>
        <div class="flex gap-2">
          <el-button
            v-if="canSort"
            :disabled="changedSortItems.length === 0"
            :loading="sortSaving"
            @click="saveSortOrder"
          >
            保存排序
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

      <el-table
        v-loading="listLoading"
        :data="items"
        row-key="id"
        :empty-text="listError ? '加载失败' : '暂无推荐记录'"
      >
        <el-table-column label="推荐菜品" min-width="270" fixed="left">
          <template #default="{ row }">
            <div class="dish-cell">
              <el-image :src="row.coverUrl" fit="cover" class="dish-cover">
                <template #error>
                  <div class="cover-fallback">无图</div>
                </template>
              </el-image>
              <div class="min-w-0">
                <div class="font-medium">{{ row.dishName }}</div>
                <div class="mt-1 text-xs text-gray-400">{{ row.id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="来源" min-width="170">
          <template #default="{ row }">
            <el-tag :type="row.sourceType === 'official' ? 'success' : 'info'">
              {{ sourceTypeLabel(row.sourceType) }}
            </el-tag>
            <div class="mt-1 text-xs text-gray-500">
              {{ row.sourceAuthorLabel || '—' }}
            </div>
            <div class="text-xs text-gray-400">{{ row.sourceDishId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="推荐位置" width="110">
          <template #default>首页精选</template>
        </el-table-column>
        <el-table-column label="排序" width="150">
          <template #default="{ row }">
            <el-input-number
              v-if="canSort"
              v-model="sortInputs[row.id]"
              :min="1"
              :max="999999"
              controls-position="right"
              class="sort-input"
            />
            <span v-else>{{ row.sortOrder }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="复制次数" prop="copyCount" width="100" align="right" />
        <el-table-column label="发布时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.publishedAt) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="350" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">查看</el-button>
            <el-button v-if="canUpdate" link type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-button
              v-if="canGovernance"
              link
              type="danger"
              @click="openGovernance(row)"
            >
              违规处理
            </el-button>
            <el-tooltip
              v-if="canPublish"
              :disabled="!publishBlockedReason(row)"
              :content="publishBlockedReason(row)"
              placement="top"
            >
              <span>
                <el-button
                  link
                  type="success"
                  :disabled="Boolean(publishBlockedReason(row))"
                  @click="publishItem(row)"
                >
                  {{ row.status === 'offline' ? '重新发布' : '发布' }}
                </el-button>
              </span>
            </el-tooltip>
            <el-tooltip
              v-if="canOffline"
              :disabled="!offlineBlockedReason(row)"
              :content="offlineBlockedReason(row)"
              placement="top"
            >
              <span>
                <el-button
                  link
                  type="warning"
                  :disabled="Boolean(offlineBlockedReason(row))"
                  @click="offlineItem(row)"
                >
                  下线
                </el-button>
              </span>
            </el-tooltip>
            <el-tooltip
              v-if="canDelete"
              :disabled="!deleteBlockedReason(row)"
              :content="deleteBlockedReason(row)"
              placement="top"
            >
              <span>
                <el-button
                  link
                  type="danger"
                  :disabled="Boolean(deleteBlockedReason(row))"
                  @click="removeItem(row)"
                >
                  删除
                </el-button>
              </span>
            </el-tooltip>
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

    <el-drawer v-model="detailVisible" title="推荐详情" size="720px">
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon />
        <template v-if="detail">
          <div class="dish-cell mb-4">
            <el-image :src="detail.coverUrl" fit="cover" class="detail-cover" />
            <div>
              <div class="text-xl font-medium">{{ detail.dishName }}</div>
              <div class="mt-1 text-sm text-gray-500">
                {{ sourceTypeLabel(detail.sourceType) }} ·
                {{ detail.sourceAuthorLabel || '—' }}
              </div>
              <div class="mt-2 flex gap-2">
                <el-tag :type="statusTagType(detail.status)">
                  {{ statusLabel(detail.status) }}
                </el-tag>
                <el-tag :type="detail.sourceAvailable ? 'success' : 'danger'">
                  {{ detail.sourceAvailable ? '来源可用' : '来源不可用' }}
                </el-tag>
              </div>
            </div>
          </div>
          <el-alert
            v-if="!detail.sourceAvailable"
            :title="detail.sourceUnavailableReason || '来源当前不可发布'"
            type="warning"
            show-icon
            class="mb-4"
          />
          <el-descriptions :column="1" border>
            <el-descriptions-item label="推荐 ID">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="来源菜品 ID">
              {{ detail.sourceDishId }}
            </el-descriptions-item>
            <el-descriptions-item label="展示说明">
              {{ detail.displayNote || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="排序值">{{ detail.sortOrder }}</el-descriptions-item>
            <el-descriptions-item label="复制次数">{{ detail.copyCount }}</el-descriptions-item>
            <el-descriptions-item label="创建人">
              {{ detail.createdBy?.nickname || detail.createdBy?.username || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="更新人">
              {{ detail.updatedBy?.nickname || detail.updatedBy?.username || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="发布时间">
              {{ formatDateTime(detail.publishedAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="下线时间">
              {{ formatDateTime(detail.offlineAt) }}
            </el-descriptions-item>
          </el-descriptions>
        </template>
      </div>
    </el-drawer>

    <el-dialog
      v-model="governanceVisible"
      title="推荐及源头违规处理"
      width="680px"
      destroy-on-close
      @closed="resetGovernance"
    >
      <el-alert
        title="推荐仅作为处理入口；确认后将直接处理其源菜品，并同步下线该源头关联的全部在线推荐。"
        type="warning"
        show-icon
        :closable="false"
        class="mb-4"
      />
      <el-form label-width="110px">
        <el-form-item label="入口推荐">
          {{ governanceTarget?.dishName }} · {{ governanceTarget?.entryRecommendationId }}
        </el-form-item>
        <el-form-item label="实际处理源头">
          {{ sourceTypeLabel(governanceTarget?.sourceType) }} ·
          {{ governanceTarget?.sourceDishId }}
        </el-form-item>
        <el-form-item label="严重程度">
          <el-radio-group
            v-model="governanceForm.severity"
            @change="handleGovernanceSeverityChange"
          >
            <el-radio-button value="normal">一般</el-radio-button>
            <el-radio-button
              value="serious"
              :disabled="!canGovernanceCascade"
            >
              严重
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          v-if="governanceTarget?.sourceType === 'creator' && governanceForm.severity === 'normal'"
          label="处理源头"
        >
          <el-radio-group
            v-model="governanceForm.normalAction"
            @change="clearGovernancePreview"
          >
            <el-radio value="disable_discoverability">关闭允许被发现</el-radio>
            <el-radio value="soft_delete_dish">软删除源菜品</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-else label="处理源头">
          <span>
            {{
              governanceForm.severity === 'serious'
                ? '处理源菜品及全部用户复制菜品'
                : '软删除官方源菜品'
            }}
          </span>
        </el-form-item>
        <el-form-item label="违规类型">
          <el-input
            v-model.trim="governanceForm.violationType"
            maxlength="40"
            placeholder="例如：违规内容、侵权、垃圾信息"
            @input="clearGovernancePreview"
          />
        </el-form-item>
        <el-form-item label="处理原因">
          <el-input
            v-model.trim="governanceForm.reason"
            type="textarea"
            :rows="4"
            minlength="4"
            maxlength="300"
            show-word-limit
            @input="clearGovernancePreview"
          />
        </el-form-item>
        <el-form-item v-if="isCascadeGovernance" label="确认词">
          <el-input
            v-model.trim="governanceForm.confirmText"
            placeholder="请输入：确认处理复制链"
          />
        </el-form-item>
      </el-form>
      <el-card v-if="governancePreview" shadow="never">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="联动下线推荐">
            {{ governancePreview.recommendationCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="处理用户复制菜品">
            {{ governancePreview.copiedDishCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="影响用户">
            {{ governancePreview.affectedUserCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="站内通知">
            {{ governancePreview.notificationCount || 0 }}
          </el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-for="warning in governancePreview.warnings || []"
          :key="warning"
          :title="warning"
          type="warning"
          :closable="false"
          class="mt-3"
        />
      </el-card>
      <template #footer>
        <el-button @click="governanceVisible = false">取消</el-button>
        <el-button :loading="governanceLoading" @click="previewGovernance">
          {{ governancePreview ? '重新预览' : '预览影响' }}
        </el-button>
        <el-button
          v-if="governancePreview"
          type="danger"
          :loading="governanceLoading"
          :disabled="isCascadeGovernance && !isCascadeConfirmValid"
          @click="executeGovernance"
        >
          一键处理
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="editorVisible"
      title="编辑推荐"
      width="560px"
      destroy-on-close
      @closed="resetEditor"
    >
      <el-form
        ref="editorFormRef"
        v-loading="editorLoading"
        :model="editorForm"
        :rules="editorRules"
        label-width="100px"
      >
        <el-form-item label="来源菜品">
          <el-input :model-value="`${editorForm.dishName} · ${editorForm.sourceDishId}`" disabled />
        </el-form-item>
        <el-form-item label="推荐位置">
          <el-input model-value="首页精选" disabled />
        </el-form-item>
        <el-form-item label="排序值" prop="sortOrder">
          <el-input-number
            v-model="editorForm.sortOrder"
            :min="1"
            :max="999999"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item label="展示说明" prop="displayNote">
          <el-input
            v-model.trim="editorForm.displayNote"
            type="textarea"
            :rows="3"
            maxlength="160"
            show-word-limit
            placeholder="可选，小程序详情页展示"
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
import { computed, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import {
  deleteRecommendation,
  getRecommendationDetail,
  getRecommendationList,
  offlineRecommendation,
  publishRecommendation,
  updateRecommendation,
  updateRecommendationSortOrder
} from '@/api/orderfood/recommendation'
import {
  executeOrderFoodGovernanceAction,
  previewOrderFoodGovernanceAction
} from '@/api/orderfood/governance'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodRecommendations'
})

const route = useRoute()
const btnAuth = useBtnAuth()
const hasPermission = (permission) => Boolean(btnAuth[permission])
const canUpdate = computed(() => hasPermission('orderfood:recommendation:update'))
const canDelete = computed(() => hasPermission('orderfood:recommendation:delete'))
const canPublish = computed(() => hasPermission('orderfood:recommendation:publish'))
const canOffline = computed(() => hasPermission('orderfood:recommendation:offline'))
const canSort = computed(() => hasPermission('orderfood:recommendation:sort'))
const canGovernance = computed(() => hasPermission('orderfood:governance:execute'))
const canGovernanceCascade = computed(() =>
  hasPermission('orderfood:governance:cascade')
)

const createDefaultSearch = () => ({
  keyword: '',
  sourceType: String(route.query.sourceType || ''),
  sourceDishId: String(route.query.sourceDishId || ''),
  status: String(route.query.status || '')
})
const searchInfo = ref(createDefaultSearch())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const items = ref([])
const listLoading = ref(false)
const listError = ref('')
const sortInputs = reactive({})
const sortSaving = ref(false)

const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const sourceTypeLabel = (value) =>
  ({ creator: '用户菜品', official: '官方菜品' })[value] || value || '—'
const statusLabel = (value) =>
  ({ draft: '草稿', published: '已发布', offline: '已下线' })[value] || value
const statusTagType = (value) =>
  ({ draft: 'info', published: 'success', offline: 'warning' })[value] || 'info'
const publishBlockedReason = (row) => {
  if (row.status === 'published') return '当前推荐已发布，无需重复发布'
  if (row.sourceAvailable === false) {
    return row.sourceUnavailableReason || '来源菜品当前不可用，不能发布'
  }
  return ''
}
const offlineBlockedReason = (row) =>
  row.status === 'published' ? '' : '只有已发布的推荐可以下线'
const deleteBlockedReason = (row) =>
  row.status === 'draft' ? '' : '只有草稿可以删除；已发布推荐需先下线'

const resetSortInputs = () => {
  Object.keys(sortInputs).forEach((key) => delete sortInputs[key])
  items.value.forEach((item) => {
    sortInputs[item.id] = item.sortOrder
  })
}
const loadList = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getRecommendationList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...searchInfo.value,
          position: 'home_featured',
          sortBy: 'sortOrder',
          sortOrder: 'asc'
        })
      ),
      '推荐列表加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
    resetSortInputs()
  } catch (error) {
    items.value = []
    total.value = 0
    resetSortInputs()
    listError.value = getOrderFoodErrorMessage(error, '推荐列表加载失败')
  } finally {
    listLoading.value = false
  }
}
const onSubmit = () => {
  page.value = 1
  loadList()
}
const onReset = () => {
  searchInfo.value = { keyword: '', sourceType: '', sourceDishId: '', status: '' }
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
const loadDetail = async (id) => {
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getRecommendationDetail(id),
      '推荐详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '推荐详情加载失败')
  } finally {
    detailLoading.value = false
  }
}
const openDetail = (row) => {
  detailVisible.value = true
  loadDetail(row.id)
}

const createDefaultGovernanceForm = () => ({
  severity: 'normal',
  normalAction: 'disable_discoverability',
  violationType: '',
  reason: '',
  confirmText: ''
})
const governanceVisible = ref(false)
const governanceLoading = ref(false)
const governanceTarget = ref(null)
const governancePreview = ref(null)
const governanceForm = ref(createDefaultGovernanceForm())
const isCascadeGovernance = computed(
  () => governanceForm.value.severity === 'serious'
)
const isCascadeConfirmValid = computed(
  () => governanceForm.value.confirmText.trim() === '确认处理复制链'
)
const clearGovernancePreview = () => {
  governancePreview.value = null
}
const handleGovernanceSeverityChange = () => {
  governanceForm.value.confirmText = ''
  clearGovernancePreview()
}
const resetGovernance = () => {
  governanceTarget.value = null
  governancePreview.value = null
  governanceForm.value = createDefaultGovernanceForm()
}
const openGovernance = async (row) => {
  governanceLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getRecommendationDetail(row.id),
      '推荐源头加载失败'
    )
    if (!data?.dish || !data?.sourceDishId || !data?.dish?.version) {
      ElMessage.warning('来源菜品已不存在，不能重复处理')
      return
    }
    governanceTarget.value = {
      entryRecommendationId: data.id,
      dishName: data.dishName,
      sourceType: data.sourceType,
      sourceDishId: data.sourceDishId,
      sourceVersion: data.dish.version
    }
    governanceForm.value = createDefaultGovernanceForm()
    governancePreview.value = null
    governanceVisible.value = true
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '推荐源头加载失败'))
  } finally {
    governanceLoading.value = false
  }
}
const governanceActions = () => {
  if (isCascadeGovernance.value) return ['delete_copy_chain']
  if (governanceTarget.value?.sourceType === 'official') {
    return ['soft_delete_official_dish']
  }
  return [governanceForm.value.normalAction]
}
const governancePayload = () => ({
  targetType:
    governanceTarget.value.sourceType === 'official' ? 'official_dish' : 'dish',
  targetId: governanceTarget.value.sourceDishId,
  entryRecommendationId: governanceTarget.value.entryRecommendationId,
  actions: governanceActions(),
  violationType: governanceForm.value.violationType,
  severity: governanceForm.value.severity,
  reason: governanceForm.value.reason,
  expectedVersion: governanceTarget.value.sourceVersion
})
const validateGovernance = (requireConfirm = false) => {
  if (!governanceTarget.value) return '推荐源头未加载'
  if (!governanceForm.value.violationType.trim()) return '请输入违规类型'
  const reasonLength = governanceForm.value.reason.trim().length
  if (reasonLength < 4 || reasonLength > 300) return '请输入 4-300 字处理原因'
  if (requireConfirm && isCascadeGovernance.value && !isCascadeConfirmValid.value) {
    return '请输入完整确认词“确认处理复制链”'
  }
  return ''
}
const previewGovernance = async () => {
  const validationMessage = validateGovernance()
  if (validationMessage) {
    ElMessage.warning(validationMessage)
    return
  }
  governanceLoading.value = true
  try {
    governancePreview.value = unwrapOrderFoodResponse(
      await previewOrderFoodGovernanceAction(governancePayload()),
      '违规处理影响预览失败'
    )
  } catch (error) {
    governancePreview.value = null
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('推荐源头版本已变化，请刷新列表后重新处理')
      await loadList()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '违规处理影响预览失败'))
    }
  } finally {
    governanceLoading.value = false
  }
}
const executeGovernance = async () => {
  if (!governancePreview.value) return
  const validationMessage = validateGovernance(true)
  if (validationMessage) {
    ElMessage.warning(validationMessage)
    return
  }
  governanceLoading.value = true
  try {
    const result = unwrapOrderFoodResponse(
      await executeOrderFoodGovernanceAction({
        ...governancePayload(),
        previewToken: governancePreview.value.previewToken,
        confirmText: isCascadeGovernance.value
          ? governanceForm.value.confirmText.trim()
          : undefined
      }),
      '违规处理执行失败'
    )
    ElMessage.success(
      result?.status === 'pending'
        ? `源头已处理，复制链任务已创建：${result.jobId || '等待调度'}`
        : `推荐和源头已一并处理，共影响 ${result?.affectedCount || 0} 条源内容`
    )
    governanceVisible.value = false
    detailVisible.value = false
    await loadList()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('推荐源头或影响预览已失效，请重新预览')
      clearGovernancePreview()
      await loadList()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '违规处理执行失败'))
    }
  } finally {
    governanceLoading.value = false
  }
}

const createDefaultEditor = () => ({
  id: '',
  dishName: '',
  sourceDishId: '',
  sortOrder: 1,
  displayNote: '',
  expectedVersion: undefined
})
const editorVisible = ref(false)
const editorLoading = ref(false)
const editorFormRef = ref(null)
const editorForm = ref(createDefaultEditor())
const editorRules = {
  sortOrder: [
    { required: true, message: '请输入排序值', trigger: 'change' },
    { type: 'number', min: 1, message: '排序值必须大于 0', trigger: 'change' }
  ]
}
const openEdit = (row) => {
  editorForm.value = {
    id: row.id,
    dishName: row.dishName,
    sourceDishId: row.sourceDishId,
    sortOrder: row.sortOrder,
    displayNote: '',
    expectedVersion: row.version
  }
  editorVisible.value = true
  editorLoading.value = true
  getRecommendationDetail(row.id)
    .then((response) => {
      const data = unwrapOrderFoodResponse(response, '推荐详情加载失败')
      editorForm.value.displayNote = data.displayNote || ''
      editorForm.value.expectedVersion = data.version
    })
    .catch((error) => {
      ElMessage.error(getOrderFoodErrorMessage(error, '推荐详情加载失败'))
      editorVisible.value = false
    })
    .finally(() => {
      editorLoading.value = false
    })
}
const resetEditor = () => {
  editorForm.value = createDefaultEditor()
  editorFormRef.value?.clearValidate()
}
const handleConflict = async (error, action) => {
  if (!isOrderFoodConflict(error)) return false
  ElMessage.warning(`${action}失败：推荐状态或版本已变化，请刷新后重试`)
  await loadList()
  return true
}
const submitEditor = async () => {
  const valid = await editorFormRef.value?.validate().catch(() => false)
  if (!valid) return
  editorLoading.value = true
  try {
    const base = {
      position: 'home_featured',
      sortOrder: editorForm.value.sortOrder,
      displayNote: editorForm.value.displayNote || null
    }
    unwrapOrderFoodResponse(
      await updateRecommendation(editorForm.value.id, {
        ...base,
        expectedVersion: editorForm.value.expectedVersion
      }),
      '推荐保存失败'
    )
    ElMessage.success('推荐信息已保存')
    editorVisible.value = false
    await loadList()
  } catch (error) {
    if (!(await handleConflict(error, '保存'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '推荐保存失败'))
    }
  } finally {
    editorLoading.value = false
  }
}

const publishItem = async (row) => {
  if (publishBlockedReason(row)) {
    ElMessage.warning(publishBlockedReason(row))
    return
  }
  try {
    await ElMessageBox.confirm(
      row.status === 'offline'
        ? `确认重新发布“${row.dishName}”吗？服务端会重新校验来源和首页容量。`
        : `确认发布“${row.dishName}”到首页精选吗？`,
      row.status === 'offline' ? '重新发布推荐' : '发布推荐',
      {
        confirmButtonText: '确认发布',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    unwrapOrderFoodResponse(
      await publishRecommendation(row.id, { expectedVersion: row.version }),
      '推荐发布失败'
    )
    ElMessage.success('推荐已发布')
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '发布'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '推荐发布失败'))
    }
  }
}
const reasonValidator = (value) => {
  const length = value?.trim().length || 0
  return length >= 4 && length <= 200 ? true : '请输入 4-200 字下线原因'
}
const offlineItem = async (row) => {
  if (offlineBlockedReason(row)) {
    ElMessage.warning(offlineBlockedReason(row))
    return
  }
  try {
    const result = await ElMessageBox.prompt(
      `确认下线“${row.dishName}”吗？用户已经复制的副本不受影响。`,
      '下线推荐',
      {
        confirmButtonText: '确认下线',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputPlaceholder: '请输入 4-200 字下线原因',
        inputValidator: reasonValidator
      }
    )
    unwrapOrderFoodResponse(
      await offlineRecommendation(row.id, {
        reason: result.value.trim(),
        expectedVersion: row.version
      }),
      '推荐下线失败'
    )
    ElMessage.success('推荐已下线')
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '下线'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '推荐下线失败'))
    }
  }
}
const removeItem = async (row) => {
  if (deleteBlockedReason(row)) {
    ElMessage.warning(deleteBlockedReason(row))
    return
  }
  try {
    await ElMessageBox.confirm(
      `确认删除“${row.dishName}”的推荐草稿吗？`,
      '删除推荐草稿',
      {
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    unwrapOrderFoodResponse(
      await deleteRecommendation(row.id, { expectedVersion: row.version }),
      '推荐草稿删除失败'
    )
    ElMessage.success('推荐草稿已删除')
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '删除'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '推荐草稿删除失败'))
    }
  }
}

const changedSortItems = computed(() =>
  items.value
    .filter((item) => Number(sortInputs[item.id]) !== Number(item.sortOrder))
    .map((item) => ({
      id: item.id,
      sortOrder: Number(sortInputs[item.id]),
      expectedVersion: item.version
    }))
)
const saveSortOrder = async () => {
  if (changedSortItems.value.length === 0) return
  if (
    changedSortItems.value.some(
      (item) => !Number.isInteger(item.sortOrder) || item.sortOrder < 1
    )
  ) {
    ElMessage.warning('排序值必须是大于 0 的整数')
    return
  }
  sortSaving.value = true
  try {
    const result = unwrapOrderFoodResponse(
      await updateRecommendationSortOrder({ items: changedSortItems.value }),
      '推荐排序保存失败'
    )
    ElMessage.success(`已更新 ${result?.updated || changedSortItems.value.length} 条排序`)
    await loadList()
  } catch (error) {
    if (!(await handleConflict(error, '保存排序'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '推荐排序保存失败'))
    }
  } finally {
    sortSaving.value = false
  }
}

loadList()
if (route.query.recommendationId) {
  detailVisible.value = true
  loadDetail(String(route.query.recommendationId))
}
</script>

<style scoped>
.table-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.dish-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dish-cover {
  width: 56px;
  height: 56px;
  flex: none;
  border-radius: 8px;
  background: #f2f3f5;
}

.detail-cover {
  width: 120px;
  height: 90px;
  flex: none;
  border-radius: 10px;
  background: #f2f3f5;
}

.cover-fallback {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  color: #a8abb2;
  font-size: 12px;
}

.sort-input {
  width: 120px;
}
</style>
