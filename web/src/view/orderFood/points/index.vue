<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="用户 ID">
          <el-input
            v-model.trim="searchInfo.userId"
            clearable
            placeholder="输入用户 ID"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="流水类型">
          <el-select v-model="searchInfo.type" clearable placeholder="全部" class="w-32">
            <el-option
              v-for="option in typeOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="业务场景">
          <el-input
            v-model.trim="searchInfo.scene"
            clearable
            placeholder="例如 daily_checkin"
          />
        </el-form-item>
        <el-form-item label="关联对象">
          <div class="flex items-center gap-2">
            <el-input
              v-model.trim="searchInfo.relatedObjectType"
              clearable
              placeholder="对象类型"
              class="w-32"
            />
            <el-input
              v-model.trim="searchInfo.relatedObjectId"
              clearable
              placeholder="对象 ID"
              class="w-52"
            />
          </div>
        </el-form-item>
        <el-form-item label="排障标识">
          <div class="flex items-center gap-2">
            <el-input
              v-model.trim="searchInfo.requestId"
              clearable
              placeholder="请求 ID"
              class="w-48"
            />
            <el-input
              v-model.trim="searchInfo.idempotencyKey"
              clearable
              placeholder="幂等键"
              class="w-48"
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
    </div>

    <div class="gva-table-box">
      <div class="table-header">
        <div>
          <div class="text-lg font-medium">积分流水</div>
          <div class="mt-1 text-sm text-gray-500">
            流水只读且不可撤销或删除；消耗与退款通过关联流水 ID 追踪。
          </div>
        </div>
        <el-button v-if="canAdjust" type="primary" @click="openAdjustment()">
          去调整积分
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
          <el-button link type="primary" @click="loadList">重新加载</el-button>
        </template>
      </el-alert>

      <el-table
        v-loading="listLoading"
        :data="tableData"
        row-key="id"
        :empty-text="listError ? '加载失败' : '暂无积分流水'"
      >
        <el-table-column label="流水" min-width="220" fixed="left">
          <template #default="{ row }">
            <div class="font-medium">{{ row.title }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.id }}</div>
          </template>
        </el-table-column>
        <el-table-column label="用户" min-width="180">
          <template #default="{ row }">
            <div>{{ row.user?.nickname || '未命名用户' }}</div>
            <div class="text-xs text-gray-400">{{ row.user?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.type)" effect="plain">
              {{ typeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="积分变化" width="110" align="right">
          <template #default="{ row }">
            <span :class="row.amount >= 0 ? 'text-green-600' : 'text-red-500'">
              {{ row.amount >= 0 ? '+' : '' }}{{ row.amount }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="变化后余额" prop="balanceAfter" width="110" align="right" />
        <el-table-column label="场景" prop="scene" min-width="150" />
        <el-table-column label="关联对象" min-width="190">
          <template #default="{ row }">
            <div>{{ row.relatedObjectType || '—' }}</div>
            <div class="text-xs text-gray-400">{{ row.relatedObjectId || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="关联流水" min-width="180">
          <template #default="{ row }">{{ row.relatedEntryId || '—' }}</template>
        </el-table-column>
        <el-table-column label="请求 ID" min-width="190">
          <template #default="{ row }">{{ row.requestId || '—' }}</template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="canReadUsers"
              link
              type="primary"
              @click="openUser(row.user?.id)"
            >
              查看用户
            </el-button>
            <el-button
              v-if="canReadAIUsage && isAIUsageEntry(row)"
              link
              type="primary"
              @click="openRelatedRecord(row)"
            >
              查看关联记录
            </el-button>
            <el-button
              v-if="canAdjust"
              link
              type="primary"
              @click="openAdjustment(row.user?.id)"
            >
              调整积分
            </el-button>
            <el-button link type="primary" @click="copyEntryID(row.id)">
              复制 ID
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
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  getShanghaiPresetRange,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import { getPointEntryList } from '@/api/orderfood/points'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodPointEntries'
})

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canReadAIUsage = computed(() =>
  Boolean(btnAuth['orderfood:ai-usage:read'])
)
const canAdjust = computed(() =>
  Boolean(btnAuth['orderfood:points:adjust'] || btnAuth.pointsAdjust)
)
const typeOptions = [
  { value: 'earned', label: '获得' },
  { value: 'spent', label: '消耗' },
  { value: 'refund', label: '退款' },
  { value: 'adjustment', label: '人工调整' }
]
const createDefaultSearch = () => ({
  userId: String(route.query.userId || ''),
  type: String(route.query.type || ''),
  scene: String(route.query.scene || ''),
  relatedObjectType: String(route.query.relatedObjectType || ''),
  relatedObjectId: String(route.query.relatedObjectId || ''),
  requestId: String(route.query.requestId || ''),
  idempotencyKey: String(route.query.idempotencyKey || ''),
  createdRange: getShanghaiPresetRange(String(route.query.range || ''))
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
    const range = searchInfo.value.createdRange || []
    const data = unwrapOrderFoodResponse(
      await getPointEntryList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          userId: searchInfo.value.userId,
          type: searchInfo.value.type,
          scene: searchInfo.value.scene,
          relatedObjectType: searchInfo.value.relatedObjectType,
          relatedObjectId: searchInfo.value.relatedObjectId,
          requestId: searchInfo.value.requestId,
          idempotencyKey: searchInfo.value.idempotencyKey,
          createdFrom: toShanghaiRFC3339(range[0]),
          createdTo: toShanghaiRFC3339(range[1]),
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      '积分流水加载失败'
    )
    tableData.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    tableData.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '积分流水加载失败')
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
    userId: '',
    type: '',
    scene: '',
    relatedObjectType: '',
    relatedObjectId: '',
    requestId: '',
    idempotencyKey: '',
    createdRange: []
  }
  page.value = 1
  loadList()
}
const handleSizeChange = () => {
  page.value = 1
  loadList()
}
const openUser = (userId) => {
  if (userId) {
    router.push({ name: 'OrderFoodUsers', query: { userId } })
  }
}
const isAIUsageEntry = (row) =>
  row?.relatedObjectType === 'feature_usage' && Boolean(row?.relatedObjectId)
const openRelatedRecord = (row) => {
  if (!isAIUsageEntry(row)) return
  router.push({
    name: 'OrderFoodAIUsages',
    query: { usageId: row.relatedObjectId }
  })
}
const openAdjustment = (userId = '') => {
  router.push({
    name: 'OrderFoodPointAdjustments',
    query: userId ? { userId } : {}
  })
}
const copyEntryID = async (entryId) => {
  try {
    await navigator.clipboard.writeText(entryId)
    ElMessage.success('流水 ID 已复制')
  } catch {
    ElMessage.warning('复制失败，请手动复制')
  }
}
const typeLabel = (type) =>
  typeOptions.find((option) => option.value === type)?.label || type || '—'
const typeTagType = (type) =>
  ({ earned: 'success', spent: 'danger', refund: 'warning', adjustment: 'info' })[
    type
  ] || 'info'
const formatDateTime = (value) => (value ? formatDate(value) : '—')

loadList()
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
