<template>
  <div>
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="请求 ID">
          <el-input v-model.trim="searchInfo.requestId" clearable />
        </el-form-item>
        <el-form-item label="文件 ID">
          <el-input v-model.trim="searchInfo.fileId" clearable />
        </el-form-item>
        <el-form-item label="用户 ID">
          <el-input v-model.trim="searchInfo.userId" clearable />
        </el-form-item>
        <el-form-item label="对象类型">
          <el-select v-model="searchInfo.objectType" clearable placeholder="全部" class="w-36">
            <el-option label="用户菜品" value="user_dish" />
            <el-option label="官方菜品" value="official_dish" />
            <el-option label="菜品" value="dish" />
            <el-option label="用户" value="user" />
            <el-option label="做菜打卡" value="checkin" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable placeholder="全部" class="w-36">
            <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="风险标签">
          <el-input v-model.trim="searchInfo.riskLabel" clearable />
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
        <div class="text-lg font-medium">图片审核记录</div>
        <div class="mt-1 text-sm text-gray-500">
          记录用户图片进入业务前的审核结果。管理端官方菜品封面不走外部审核，因此不会产生这类记录。
        </div>
      </div>
      <el-alert v-if="listError" :title="listError" type="error" show-icon class="mb-4" />
      <el-table v-loading="listLoading" :data="items" row-key="id">
        <el-table-column label="审核记录" min-width="220">
          <template #default="{ row }">
            <div class="font-medium">{{ row.id }}</div>
            <div class="mt-1 text-xs text-gray-400">请求 {{ row.requestId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="文件 ID" prop="fileId" min-width="190" />
        <el-table-column label="用户" min-width="170">
          <template #default="{ row }">
            <div>{{ row.user?.nickname || '—' }}</div>
            <div class="text-xs text-gray-400">{{ row.user?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="业务对象" min-width="180">
          <template #default="{ row }">
            <div>{{ objectTypeLabel(row.objectType) }}</div>
            <div class="text-xs text-gray-400">{{ row.objectId || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="105">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="风险" min-width="180">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-1">
              <el-tag v-for="label in row.riskLabels || []" :key="label" type="danger" size="small">
                {{ label }}
              </el-tag>
              <span v-if="!(row.riskLabels || []).length">—</span>
            </div>
            <div v-if="row.riskLevel" class="mt-1 text-xs text-gray-400">{{ row.riskLevel }}</div>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">{{ row.durationMs == null ? '—' : `${row.durationMs} ms` }}</template>
        </el-table-column>
        <el-table-column label="错误码" width="130">
          <template #default="{ row }">{{ row.errorCode || '—' }}</template>
        </el-table-column>
        <el-table-column label="供应商请求 ID" min-width="190">
          <template #default="{ row }">{{ row.providerRequestId || '—' }}</template>
        </el-table-column>
        <el-table-column label="时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">查看详情</el-button>
            <el-button v-if="canReadMedia" link type="primary" @click="openMedia(row.fileId)">
              查看图片
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

    <el-drawer v-model="detailVisible" title="图片审核详情" size="700px">
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon />
        <template v-if="detail">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="记录 ID">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="请求 ID">{{ detail.requestId }}</el-descriptions-item>
            <el-descriptions-item label="文件 ID">{{ detail.fileId }}</el-descriptions-item>
            <el-descriptions-item label="用户">
              {{ detail.user?.nickname || '—' }}
              <span v-if="detail.user?.id"> · {{ detail.user.id }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="审核状态">{{ statusLabel(detail.status) }}</el-descriptions-item>
            <el-descriptions-item label="风险标签">
              {{ (detail.riskLabels || []).join('、') || '无' }}
            </el-descriptions-item>
            <el-descriptions-item label="风险等级">
              {{ detail.riskLevel || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="审核耗时">
              {{ detail.durationMs == null ? '—' : `${detail.durationMs} ms` }}
            </el-descriptions-item>
            <el-descriptions-item label="业务对象">
              {{ objectTypeLabel(detail.objectType) }} {{ detail.objectId || '' }}
            </el-descriptions-item>
            <el-descriptions-item label="供应商请求">
              {{ detail.providerRequestId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="端点">{{ detail.endpointLabel || '—' }}</el-descriptions-item>
            <el-descriptions-item label="错误码">{{ detail.errorCode || '—' }}</el-descriptions-item>
            <el-descriptions-item label="错误摘要">{{ detail.errorSummary || '—' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(detail.createdAt) }}
            </el-descriptions-item>
          </el-descriptions>
          <div v-if="detail.providerResponseSummary && canReadSensitive" class="mt-5">
            <el-button @click="providerSummaryVisible = !providerSummaryVisible">
              {{ providerSummaryVisible ? '收起供应商响应摘要' : '查看供应商响应摘要' }}
            </el-button>
            <div v-if="providerSummaryVisible" class="mt-3">
              <div class="mb-2 font-medium">供应商响应脱敏摘要</div>
              <pre class="json-box">{{ JSON.stringify(detail.providerResponseSummary, null, 2) }}</pre>
            </div>
          </div>
          <el-alert
            v-else-if="!canReadSensitive"
            class="mt-4"
            title="当前角色没有查看供应商响应摘要的权限。"
            type="info"
            show-icon
            :closable="false"
          />
          <el-alert
            v-else-if="canReadSensitive"
            class="mt-4"
            title="本条记录没有可展示的供应商响应摘要。"
            type="info"
            show-icon
            :closable="false"
          />
          <div class="mt-4 flex flex-wrap gap-2">
            <el-button v-if="canReadMedia" @click="openMedia(detail.fileId)">
              查看图片
            </el-button>
            <el-button
              v-if="canOpenObject(detail)"
              @click="openObject(detail)"
            >
              查看关联对象
            </el-button>
          </div>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  getShanghaiPresetRange,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import {
  getModerationRecordDetail,
  getModerationRecordList
} from '@/api/orderfood/media'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodModerationRecords' })

const router = useRouter()
const route = useRoute()
const btnAuth = useBtnAuth()
const canReadSensitive = computed(() =>
  Boolean(btnAuth['orderfood:moderation:sensitive-read'])
)
const canReadMedia = computed(() => Boolean(btnAuth['orderfood:media:read']))
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canReadUserDishes = computed(() =>
  Boolean(btnAuth['orderfood:user-dish:private-read'])
)
const canReadOfficialDishes = computed(() =>
  Boolean(btnAuth['orderfood:official-dish:read'])
)
const statusOptions = [
  { label: '未通过或审核失败', value: 'rejected_or_failed' },
  { label: '审核中', value: 'pending' },
  { label: '已通过', value: 'passed' },
  { label: '未通过', value: 'rejected' },
  { label: '审核失败', value: 'failed' },
  { label: '无需审核', value: 'not_required' }
]
const statusLabel = (value) =>
  statusOptions.find((item) => item.value === value)?.label || value || '—'
const statusTagType = (value) =>
  ({ passed: 'success', rejected: 'danger', failed: 'danger', pending: 'warning', not_required: 'info' })[
    value
  ] || 'info'
const objectTypeLabel = (value) =>
  ({
    dish: '菜品',
    user_dish: '用户菜品',
    official_dish: '官方菜品',
    user: '用户',
    checkin: '做菜打卡'
  })[value] || value || '—'
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )
const emptySearch = () => ({
  requestId: '',
  fileId: '',
  userId: '',
  objectType: '',
  status: '',
  riskLabel: '',
  createdRange: []
})
const defaultSearch = () => ({
  ...emptySearch(),
  fileId: String(route.query.fileId || ''),
  userId: String(route.query.userId || ''),
  objectType: String(route.query.objectType || ''),
  status: String(route.query.status || ''),
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
      await getModerationRecordList(
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
      '审核记录加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
  } catch (error) {
    items.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '审核记录加载失败')
  } finally {
    listLoading.value = false
  }
}
const onSubmit = () => {
  page.value = 1
  loadList()
}
const onReset = () => {
  searchInfo.value = emptySearch()
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
const providerSummaryVisible = ref(false)
const openDetail = async (row) => {
  const recordId = typeof row === 'string' ? row : row?.id
  if (!recordId) return
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  providerSummaryVisible.value = false
  try {
    detail.value = unwrapOrderFoodResponse(
      await getModerationRecordDetail(recordId),
      '审核详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '审核详情加载失败')
  } finally {
    detailLoading.value = false
  }
}
const openMedia = (fileId) =>
  router.push({ name: 'OrderFoodMedia', query: { fileId } })
const canOpenObject = (item) => {
  if (!item?.objectId) return false
  if (item.objectType === 'user') return canReadUsers.value
  if (item.objectType === 'dish' || item.objectType === 'user_dish') {
    return canReadUserDishes.value
  }
  if (item.objectType === 'official_dish') {
    return canReadOfficialDishes.value
  }
  return false
}
const openObject = (item) => {
  if (!canOpenObject(item)) return
  if (item.objectType === 'user') {
    router.push({ name: 'OrderFoodUsers', query: { userId: item.objectId } })
    return
  }
  if (item.objectType === 'official_dish') {
    router.push({
      name: 'OrderFoodOfficialDishes',
      query: { officialDishId: item.objectId }
    })
    return
  }
  router.push({
    name: 'OrderFoodUserDishes',
    query: { openDishId: item.objectId }
  })
}

loadList().then(() => {
  if (route.query.recordId) {
    openDetail(String(route.query.recordId))
  }
})
</script>

<style scoped>
.json-box {
  overflow: auto;
  max-height: 360px;
  padding: 12px;
  border-radius: 6px;
  color: var(--el-text-color-primary);
  background: var(--el-fill-color-light);
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
