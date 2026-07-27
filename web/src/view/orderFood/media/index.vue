<template>
  <div>
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="文件 ID">
          <el-input v-model.trim="searchInfo.fileId" clearable placeholder="输入文件 ID" />
        </el-form-item>
        <el-form-item label="上传用户 ID">
          <el-input v-model.trim="searchInfo.uploaderUserId" clearable />
        </el-form-item>
        <el-form-item label="上传来源">
          <el-select v-model="searchInfo.uploadSource" clearable placeholder="全部" class="w-36">
            <el-option label="用户上传" value="user_upload" />
            <el-option label="管理端上传" value="admin_upload" />
            <el-option label="生成图片" value="ai_generated" />
          </el-select>
        </el-form-item>
        <el-form-item label="使用场景">
          <el-select v-model="searchInfo.sourceScene" clearable placeholder="全部" class="w-44">
            <el-option v-for="item in sceneOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="searchInfo.reviewStatus" clearable placeholder="全部" class="w-36">
            <el-option v-for="item in reviewOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源状态">
          <el-select v-model="searchInfo.resourceStatus" clearable placeholder="全部" class="w-32">
            <el-option label="正常" value="active" />
            <el-option label="文件缺失" value="missing" />
            <el-option label="已删除" value="deleted" />
          </el-select>
        </el-form-item>
        <el-form-item label="绑定对象">
          <div class="flex items-center gap-2">
            <el-select
              v-model="searchInfo.boundObjectType"
              clearable
              placeholder="对象类型"
              class="w-36"
            >
              <el-option label="用户菜品" value="user_dish" />
              <el-option label="官方菜品" value="official_dish" />
              <el-option label="做菜打卡" value="checkin" />
            </el-select>
            <el-input
              v-model.trim="searchInfo.boundObjectId"
              clearable
              placeholder="对象 ID"
              class="w-52"
            />
          </div>
        </el-form-item>
        <el-form-item label="创建时间">
          <el-date-picker
            v-model="searchInfo.createdRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </AdvancedSearchPanel>

    <div class="gva-table-box">
      <div class="mb-4">
        <div class="text-lg font-medium">图片资源</div>
        <div class="mt-1 text-sm text-gray-500">
          用于定位图片来源、业务绑定和审核结果；本页不提供绕过审核恢复或直接删除图片的操作。
        </div>
      </div>
      <el-alert v-if="listError" :title="listError" type="error" show-icon class="mb-4" />
      <el-table v-loading="listLoading" :data="items" row-key="id">
        <el-table-column label="图片" width="92">
          <template #default="{ row }">
            <el-image
              v-if="row.url"
              :src="row.url"
              fit="cover"
              class="thumbnail"
              :preview-src-list="[row.url]"
              preview-teleported
            />
            <div v-else class="thumbnail thumbnail-empty">不可用</div>
          </template>
        </el-table-column>
        <el-table-column label="文件" min-width="220">
          <template #default="{ row }">
            <div class="font-medium">{{ row.id }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.objectKeyMasked }}</div>
            <div class="text-xs text-gray-400">
              {{ formatDimensions(row) }} · {{ formatBytes(row.fileSize) }} ·
              {{ row.mimeType }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="来源" min-width="170">
          <template #default="{ row }">
            <div>{{ uploadSourceLabel(row.uploadSource) }}</div>
            <div class="text-xs text-gray-400">{{ sceneLabel(row.sourceScene) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="上传者" min-width="170">
          <template #default="{ row }">
            {{ row.uploader?.nickname || row.uploader?.username || '—' }}
            <div class="text-xs text-gray-400">{{ row.uploader?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="绑定对象" min-width="180">
          <template #default="{ row }">
            <div>{{ objectTypeLabel(row.boundObjectType) }}</div>
            <div class="text-xs text-gray-400">{{ row.boundObjectId || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="审核" width="105">
          <template #default="{ row }">
            <el-tag :type="reviewTagType(row.reviewStatus)">
              {{ reviewLabel(row.reviewStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="资源" width="95">
          <template #default="{ row }">
            <el-tag :type="row.resourceStatus === 'active' ? 'success' : 'danger'">
              {{ resourceLabel(row.resourceStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="上传时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">查看详情</el-button>
            <el-button
              v-if="canGovernObject(row)"
              link
              type="danger"
              @click="openGovernance(row)"
            >
              违规处理
            </el-button>
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

    <el-drawer v-model="detailVisible" title="图片资源详情" size="680px">
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon />
        <template v-if="detail">
          <el-image
            v-if="detail.url"
            :src="detail.url"
            fit="contain"
            class="detail-image"
            :preview-src-list="[detail.url]"
            preview-teleported
          />
          <el-descriptions :column="1" border>
            <el-descriptions-item label="文件 ID">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="对象键">{{ detail.objectKeyMasked }}</el-descriptions-item>
            <el-descriptions-item label="访问地址">
              <el-link
                v-if="detail.url"
                :href="detail.url"
                target="_blank"
                type="primary"
              >
                打开原图
              </el-link>
              <span v-else>不可访问</span>
            </el-descriptions-item>
            <el-descriptions-item label="尺寸">
              {{ detail.width || '—' }} × {{ detail.height || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="摘要">{{ detail.checksum || '—' }}</el-descriptions-item>
            <el-descriptions-item label="上传来源">
              {{ uploadSourceLabel(detail.uploadSource) }} / {{ sceneLabel(detail.sourceScene) }}
            </el-descriptions-item>
            <el-descriptions-item label="上传者">
              {{ detail.uploader?.nickname || detail.uploader?.username || '—' }}
              <span v-if="detail.uploader?.id"> · {{ detail.uploader.id }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="绑定对象">
              {{ objectTypeLabel(detail.boundObjectType) }} {{ detail.boundObjectId || '' }}
            </el-descriptions-item>
            <el-descriptions-item label="最近审核">
              <template v-if="detail.latestModerationRecord">
                <div>
                  {{ reviewLabel(detail.latestModerationRecord.status) }} ·
                  {{ formatDateTime(detail.latestModerationRecord.createdAt) }}
                </div>
                <div class="text-xs text-gray-500">
                  请求 {{ detail.latestModerationRecord.requestId }} · 风险标签：
                  {{ (detail.latestModerationRecord.riskLabels || []).join('、') || '无' }}
                </div>
              </template>
              <span v-else>无审核记录</span>
            </el-descriptions-item>
          </el-descriptions>
          <div class="mt-4 flex flex-wrap gap-2">
            <el-button
              v-if="canOpenUploader(detail)"
              @click="openUploader(detail.uploader.id)"
            >
              查看上传用户
            </el-button>
            <el-button
              v-if="canOpenBoundObject(detail)"
              @click="openBoundObject(detail)"
            >
              查看绑定对象
            </el-button>
            <el-button
              v-if="canGovernObject(detail)"
              type="danger"
              @click="openGovernance(detail)"
            >
              违规处理
            </el-button>
          </div>
        </template>
      </div>
    </el-drawer>

    <el-dialog
      v-model="governanceVisible"
      :title="`${objectTypeLabel(governanceTarget?.boundObjectType)}违规处理`"
      width="600px"
    >
      <el-form :model="governanceForm" label-width="90px">
        <el-form-item label="目标对象">
          {{ objectTypeLabel(governanceTarget?.boundObjectType) }} ·
          {{ governanceTarget?.boundObjectId }}
        </el-form-item>
        <el-form-item label="处理动作">
          <el-checkbox-group
            v-model="governanceForm.actions"
            @change="handleGovernanceActionsChange"
          >
            <el-checkbox
              v-for="option in governanceActionOptions"
              :key="option.value"
              :value="option.value"
              :disabled="option.disabled"
            >
              {{ option.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="违规类型">
          <el-input
            v-model.trim="governanceForm.violationType"
            maxlength="40"
            @input="clearGovernancePreview"
          />
        </el-form-item>
        <el-form-item label="严重程度">
          <el-radio-group v-model="governanceForm.severity" @change="clearGovernancePreview">
            <el-radio value="normal" :disabled="isCascadeGovernance">一般</el-radio>
            <el-radio value="serious">严重</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理原因">
          <el-input
            v-model.trim="governanceForm.reason"
            type="textarea"
            maxlength="300"
            @input="clearGovernancePreview"
          />
        </el-form-item>
        <el-form-item v-if="isCascadeGovernance" label="确认词">
          <div class="w-full">
            <el-input
              v-model="governanceForm.confirmText"
              maxlength="20"
              placeholder="请输入：确认处理复制链"
            />
            <div class="mt-1 text-xs text-red-500">
              必须完整输入“确认处理复制链”。
            </div>
          </div>
        </el-form-item>
      </el-form>
      <el-card v-if="governancePreview" shadow="never">
        <template #header>影响预览</template>
        <el-descriptions :column="2">
          <el-descriptions-item label="推荐记录">
            {{ governancePreview.recommendationCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="复制菜品">
            {{ governancePreview.copiedDishCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="影响用户">
            {{ governancePreview.affectedUserCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="通知数量">
            {{ governancePreview.notificationCount || 0 }}
          </el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-for="warning in governancePreview.warnings || []"
          :key="warning"
          :title="warning"
          type="warning"
          show-icon
          :closable="false"
          class="mt-2"
        />
      </el-card>
      <template #footer>
        <el-button @click="governanceVisible = false">取消</el-button>
        <el-button :loading="governanceLoading" @click="previewGovernance">预览影响</el-button>
        <el-button
          v-if="governancePreview"
          type="danger"
          :loading="governanceLoading"
          :disabled="isCascadeGovernance && !isCascadeConfirmValid"
          @click="executeGovernance"
        >
          确认执行
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  getShanghaiPresetRange,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import { getMediaDetail, getMediaList } from '@/api/orderfood/media'
import {
  executeOrderFoodGovernanceAction,
  previewOrderFoodGovernanceAction
} from '@/api/orderfood/governance'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodMedia' })

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canReadUserDishes = computed(() =>
  Boolean(btnAuth['orderfood:user-dish:private-read'])
)
const canReadOfficialDishes = computed(() =>
  Boolean(btnAuth['orderfood:official-dish:read'])
)
const canExecuteGovernance = computed(() =>
  Boolean(btnAuth['orderfood:governance:execute'])
)
const canCascadeGovernance = computed(
  () => Boolean(btnAuth['orderfood:governance:cascade'])
)
const sceneOptions = [
  { label: '用户头像', value: 'profile_avatar' },
  { label: '菜品封面', value: 'dish_cover' },
  { label: '菜品步骤', value: 'dish_step' },
  { label: '图片识菜', value: 'dish_extract' },
  { label: '做菜打卡', value: 'checkin' },
  { label: '生成封面', value: 'generated_cover' },
  { label: '官方菜品封面', value: 'official_dish_cover' }
]
const reviewOptions = [
  { label: '审核中', value: 'pending' },
  { label: '已通过', value: 'passed' },
  { label: '未通过', value: 'rejected' },
  { label: '审核失败', value: 'failed' },
  { label: '无需审核', value: 'not_required' }
]
const labelFrom = (options, value) =>
  options.find((item) => item.value === value)?.label || value || '—'
const sceneLabel = (value) => labelFrom(sceneOptions, value)
const reviewLabel = (value) => labelFrom(reviewOptions, value)
const uploadSourceLabel = (value) =>
  ({ user_upload: '用户上传', admin_upload: '管理端上传', ai_generated: '生成图片' })[
    value
  ] || value || '—'
const objectTypeLabel = (value) =>
  ({ user_dish: '用户菜品', official_dish: '官方菜品', checkin: '做菜打卡' })[
    value
  ] || value || '未绑定'
const resourceLabel = (value) =>
  ({ active: '正常', missing: '缺失', deleted: '已删除' })[value] || value || '—'
const reviewTagType = (value) =>
  ({ passed: 'success', rejected: 'danger', failed: 'danger', pending: 'warning', not_required: 'info' })[
    value
  ] || 'info'
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const formatDimensions = (item) =>
  item?.width && item?.height ? `${item.width} × ${item.height}` : '尺寸未知'
const formatBytes = (value) => {
  const size = Number(value || 0)
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}
const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )

const defaultSearch = () => ({
  fileId: route.query.fileId ? String(route.query.fileId) : '',
  uploaderUserId: String(route.query.uploaderUserId || ''),
  uploadSource: '',
  sourceScene: String(route.query.sourceScene || ''),
  reviewStatus: '',
  resourceStatus: '',
  boundObjectType: String(route.query.boundObjectType || ''),
  boundObjectId: String(route.query.boundObjectId || ''),
  createdRange: getShanghaiPresetRange(String(route.query.range || ''))
})
const searchInfo = ref(defaultSearch())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const items = ref([])
const listLoading = ref(false)
const listError = ref('')
const loadList = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const range = searchInfo.value.createdRange || []
    const data = unwrapOrderFoodResponse(
      await getMediaList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...searchInfo.value,
          createdRange: undefined,
          createdFrom: toShanghaiRFC3339(range[0]),
          createdTo: toShanghaiRFC3339(range[1]),
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      '图片资源加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
  } catch (error) {
    items.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '图片资源加载失败')
  } finally {
    listLoading.value = false
  }
}
const onSubmit = () => {
  page.value = 1
  loadList()
}
const onReset = () => {
  searchInfo.value = {
    fileId: '',
    uploaderUserId: '',
    uploadSource: '',
    sourceScene: '',
    reviewStatus: '',
    resourceStatus: '',
    boundObjectType: '',
    boundObjectId: '',
    createdRange: []
  }
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
const openDetail = async (row) => {
  const fileId = typeof row === 'string' ? row : row?.id
  if (!fileId) return
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getMediaDetail(fileId),
      '图片资源详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '图片资源详情加载失败')
  } finally {
    detailLoading.value = false
  }
}
const canOpenUploader = (item) =>
  canReadUsers.value &&
  item?.uploadSource === 'user_upload' &&
  Boolean(item?.uploader?.id)
const openUploader = (userId) =>
  router.push({ name: 'OrderFoodUsers', query: { userId } })
const canOpenBoundObject = (item) => {
  if (!item?.boundObjectId) return false
  if (item.boundObjectType === 'user_dish') return canReadUserDishes.value
  if (item.boundObjectType === 'official_dish') {
    return canReadOfficialDishes.value
  }
  return false
}
const openBoundObject = (item) => {
  if (!canOpenBoundObject(item)) return
  if (item.boundObjectType === 'user_dish') {
    router.push({
      name: 'OrderFoodUserDishes',
      query: { openDishId: item.boundObjectId }
    })
    return
  }
  router.push({
    name: 'OrderFoodOfficialDishes',
    query: { officialDishId: item.boundObjectId }
  })
}
const canGovernObject = (item) => {
  if (!canExecuteGovernance.value || !item?.boundObjectId || item.boundObjectDeleted) {
    return false
  }
  return (
    ['user_dish', 'official_dish', 'checkin'].includes(item.boundObjectType) &&
    Number(item.boundObjectVersion || 0) > 0
  )
}

const governanceVisible = ref(false)
const governanceLoading = ref(false)
const governanceTarget = ref(null)
const governancePreview = ref(null)
const governanceForm = ref({
  actions: ['disable_discoverability'],
  violationType: '',
  severity: 'normal',
  reason: '',
  confirmText: '',
  expectedVersion: 0
})
const openGovernance = async (row) => {
  if (!canGovernObject(row)) return
  governanceLoading.value = true
  try {
    const current = unwrapOrderFoodResponse(
      await getMediaDetail(row.id),
      '图片绑定对象加载失败'
    )
    if (!canGovernObject(current)) {
      ElMessage.warning('绑定对象已变化，当前不能继续处理')
      await loadList()
      return
    }
    const defaultAction =
      {
        user_dish: 'disable_discoverability',
        official_dish: 'soft_delete_official_dish',
        checkin: 'soft_delete_checkin'
      }[current.boundObjectType] || ''
    governanceTarget.value = current
    governanceForm.value = {
      actions: defaultAction ? [defaultAction] : [],
      violationType: '',
      severity: 'normal',
      reason: '',
      confirmText: '',
      expectedVersion: Number(current.boundObjectVersion || 0)
    }
    governancePreview.value = null
    governanceVisible.value = true
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '图片绑定对象加载失败'))
  } finally {
    governanceLoading.value = false
  }
}
const isCascadeGovernance = computed(() =>
  governanceForm.value.actions.includes('delete_copy_chain')
)
const isCascadeConfirmValid = computed(
  () => governanceForm.value.confirmText.trim() === '确认处理复制链'
)
const governanceActionOptions = computed(() => {
  const type = governanceTarget.value?.boundObjectType
  const hasOtherAction = governanceForm.value.actions.some(
    (action) => action !== 'delete_copy_chain'
  )
  if (type === 'user_dish') {
    const options = [
      {
        value: 'disable_discoverability',
        label: '关闭允许被发现',
        disabled: isCascadeGovernance.value
      },
      {
        value: 'soft_delete_dish',
        label: '软删除菜品',
        disabled: isCascadeGovernance.value
      }
    ]
    if (canCascadeGovernance.value) {
      options.push({
        value: 'delete_copy_chain',
        label: '严重违规：处理全部复制链',
        disabled: hasOtherAction
      })
    }
    return options
  }
  if (type === 'official_dish') {
    const options = [
      {
        value: 'soft_delete_official_dish',
        label: '软删除官方菜品',
        disabled: isCascadeGovernance.value
      }
    ]
    if (canCascadeGovernance.value) {
      options.push({
        value: 'delete_copy_chain',
        label: '严重违规：处理全部复制链',
        disabled: hasOtherAction
      })
    }
    return options
  }
  return [{ value: 'soft_delete_checkin', label: '软删除打卡', disabled: false }]
})
const clearGovernancePreview = () => {
  governancePreview.value = null
}
const handleGovernanceActionsChange = (actions) => {
  if (actions.includes('delete_copy_chain')) {
    governanceForm.value.actions = ['delete_copy_chain']
    governanceForm.value.severity = 'serious'
  } else {
    governanceForm.value.confirmText = ''
  }
  clearGovernancePreview()
}
const governancePayload = () => ({
  targetType:
    {
      user_dish: 'dish',
      official_dish: 'official_dish',
      checkin: 'checkin'
    }[governanceTarget.value.boundObjectType],
  targetId: governanceTarget.value.boundObjectId,
  actions: governanceForm.value.actions,
  violationType: governanceForm.value.violationType,
  severity: governanceForm.value.severity,
  reason: governanceForm.value.reason,
  expectedVersion: governanceForm.value.expectedVersion
})
const validateGovernance = () => {
  if (!governanceForm.value.actions.length) return '请选择处理动作'
  if (!governanceForm.value.violationType.trim()) return '请输入违规类型'
  if (governanceForm.value.reason.trim().length < 4) return '处理原因至少 4 个字符'
  if (isCascadeGovernance.value && governanceForm.value.severity !== 'serious') {
    return '处理复制链必须选择严重级别'
  }
  return ''
}
const previewGovernance = async () => {
  const errorMessage = validateGovernance()
  if (errorMessage) {
    ElMessage.warning(errorMessage)
    return
  }
  governanceLoading.value = true
  try {
    governancePreview.value = unwrapOrderFoodResponse(
      await previewOrderFoodGovernanceAction(governancePayload()),
      '影响预览失败'
    )
  } catch (error) {
    governancePreview.value = null
    ElMessage.error(getOrderFoodErrorMessage(error, '影响预览失败'))
  } finally {
    governanceLoading.value = false
  }
}
const executeGovernance = async () => {
  if (isCascadeGovernance.value && !isCascadeConfirmValid.value) {
    ElMessage.warning('请输入完整确认词“确认处理复制链”')
    return
  }
  governanceLoading.value = true
  try {
    unwrapOrderFoodResponse(
      await executeOrderFoodGovernanceAction({
        ...governancePayload(),
        previewToken: governancePreview.value.previewToken,
        confirmText: isCascadeGovernance.value
          ? governanceForm.value.confirmText
          : ''
      }),
      '违规处理失败'
    )
    ElMessage.success('违规处理已执行')
    governanceVisible.value = false
    await loadList()
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '违规处理失败'))
  } finally {
    governanceLoading.value = false
  }
}

loadList().then(() => {
  if (route.query.fileId) {
    openDetail(String(route.query.fileId))
  }
})
</script>

<style scoped>
.thumbnail {
  width: 64px;
  height: 48px;
  border-radius: 6px;
  background: #f2f3f5;
}

.thumbnail-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #a8abb2;
  font-size: 12px;
}

.detail-image {
  width: 100%;
  height: 260px;
  margin-bottom: 16px;
  border-radius: 8px;
  background: #f2f3f5;
}
</style>
