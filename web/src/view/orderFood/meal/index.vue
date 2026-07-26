<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="饭局">
          <el-input v-model.trim="searchInfo.keyword" clearable placeholder="名称或饭局 ID" />
        </el-form-item>
        <el-form-item label="创建者 ID">
          <el-input v-model.trim="searchInfo.creatorId" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable placeholder="全部" class="w-36">
            <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="关闭原因">
          <el-select v-model="searchInfo.closeReason" clearable placeholder="全部" class="w-36">
            <el-option label="发起人关闭" value="manual" />
            <el-option label="截止时间到达" value="deadline" />
          </el-select>
        </el-form-item>
        <el-form-item label="取消原因">
          <el-select v-model="searchInfo.cancelReason" clearable placeholder="全部" class="w-40">
            <el-option label="手动取消" value="manual" />
            <el-option label="发起人被禁用" value="creator_disabled" />
          </el-select>
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
        <el-form-item label="点单截止">
          <el-date-picker
            v-model="searchInfo.deadlineRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item label="参与时间">
          <el-date-picker
            v-model="searchInfo.joinedRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item label="确认菜单">
          <el-date-picker
            v-model="searchInfo.confirmedRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item label="完成饭局">
          <el-date-picker
            v-model="searchInfo.completedRange"
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
      <div class="mb-4">
        <div class="text-lg font-medium">饭局列表</div>
        <div class="mt-1 text-sm text-gray-500">
          本页用于查看饭局进度、候选菜和最终菜单，不提供代替发起人关闭或完成饭局的操作。
        </div>
      </div>
      <el-alert v-if="listError" :title="listError" type="error" show-icon class="mb-4" />
      <el-table v-loading="listLoading" :data="items" row-key="id">
        <el-table-column label="饭局" min-width="220">
          <template #default="{ row }">
            <div class="font-medium">{{ row.name }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.id }} · 邀请码 {{ row.codeMasked }}</div>
          </template>
        </el-table-column>
        <el-table-column label="创建者" min-width="160">
          <template #default="{ row }">
            <div>{{ row.creator?.nickname || '—' }}</div>
            <div class="text-xs text-gray-400">{{ row.creator?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="115">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="原因" min-width="145">
          <template #default="{ row }">
            {{ reasonLabel(row) }}
          </template>
        </el-table-column>
        <el-table-column label="参与 / 候选 / 最终" width="145">
          <template #default="{ row }">
            {{ row.participantCount }} / {{ row.candidateCount }} / {{ row.finalDishCount }}
          </template>
        </el-table-column>
        <el-table-column label="点单截止" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.deadlineAt) }}</template>
        </el-table-column>
        <el-table-column label="确认菜单" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.confirmedAt) }}</template>
        </el-table-column>
        <el-table-column label="完成饭局" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.completedAt) }}</template>
        </el-table-column>
        <el-table-column label="采购清单" min-width="160">
          <template #default="{ row }">
            <el-button
              v-if="canReadShopping && row.shoppingListId"
              link
              type="primary"
              @click="openShopping(row.shoppingListId)"
            >
              {{ row.shoppingListId }}
            </el-button>
            <span v-else>未生成</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="105" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row.id)">查看详情</el-button>
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

    <el-drawer v-model="detailVisible" title="饭局详情" size="880px">
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon />
        <template v-if="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="饭局">{{ detail.name }}（{{ detail.id }}）</el-descriptions-item>
            <el-descriptions-item label="状态">{{ statusLabel(detail.status) }}</el-descriptions-item>
            <el-descriptions-item label="创建者">
              {{ detail.creator?.nickname || '—' }} · {{ detail.creator?.id || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="邀请码">{{ detail.codeMasked }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatDateTime(detail.createdAt) }}</el-descriptions-item>
            <el-descriptions-item label="点单截止">{{ formatDateTime(detail.deadlineAt) }}</el-descriptions-item>
            <el-descriptions-item label="关闭点单">{{ formatDateTime(detail.closedAt) }}</el-descriptions-item>
            <el-descriptions-item label="关闭来源">
              {{ closeSourceLabel(detail.closeSource) }}
            </el-descriptions-item>
            <el-descriptions-item label="确认菜单">{{ formatDateTime(detail.confirmedAt) }}</el-descriptions-item>
            <el-descriptions-item label="完成饭局">{{ formatDateTime(detail.completedAt) }}</el-descriptions-item>
            <el-descriptions-item label="取消饭局">{{ formatDateTime(detail.cancelledAt) }}</el-descriptions-item>
            <el-descriptions-item label="取消前状态">
              {{ statusLabel(detail.cancelledFromStatus) }}
            </el-descriptions-item>
            <el-descriptions-item label="采购清单生成">
              {{ shoppingGenerationLabel(detail) }}
            </el-descriptions-item>
          </el-descriptions>

          <el-alert
            v-if="detail.cancelReason === 'creator_disabled'"
            type="warning"
            show-icon
            :closable="false"
            class="mt-4"
            :title="disabledCreatorRetentionSummary(detail)"
          />

          <div class="mt-4 flex flex-wrap gap-2">
            <el-button
              v-if="canReadUsers && detail.creator?.id"
              @click="openUser(detail.creator.id)"
            >
              查看创建者
            </el-button>
            <el-button
              v-if="canReadShopping && detail.shoppingListSummary?.id"
              type="primary"
              @click="openShopping(detail.shoppingListSummary.id)"
            >
              查看采购清单
            </el-button>
            <el-button @click="copyMealID(detail.id)">复制饭局 ID</el-button>
          </div>

          <el-tabs class="mt-4">
            <el-tab-pane :label="`参与者（${detail.participants?.length || 0}）`">
              <el-table :data="detail.participants || []">
                <el-table-column prop="id" label="用户 ID" min-width="200" />
                <el-table-column prop="nickname" label="昵称" min-width="160" />
                <el-table-column label="状态" width="110">
                  <template #default="{ row }">{{ userStatusLabel(row.status) }}</template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            <el-tab-pane :label="`候选菜（${detail.candidates?.length || 0}）`">
              <el-table :data="detail.candidates || []">
                <el-table-column prop="name" label="菜品" min-width="180" />
                <el-table-column prop="id" label="候选菜 ID" min-width="200" />
                <el-table-column prop="dishId" label="菜品 ID" min-width="180" />
                <el-table-column prop="voteCount" label="点选人数" width="100" />
                <el-table-column label="候选状态" width="120">
                  <template #default="{ row }">{{ candidateStatusLabel(row.status) }}</template>
                </el-table-column>
                <el-table-column label="最终菜单" width="120">
                  <template #default="{ row }">
                    {{ row.selected ? `${row.finalServings || 0} 份` : '未选中' }}
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            <el-tab-pane :label="`最终菜单（${detail.finalDishSnapshots?.length || 0}）`">
              <el-table :data="detail.finalDishSnapshots || []">
                <el-table-column type="expand">
                  <template #default="{ row }">
                    <el-descriptions :column="1" border>
                      <el-descriptions-item label="菜品 ID">
                        {{ row.dishId || '—' }}
                      </el-descriptions-item>
                      <el-descriptions-item label="封面">
                        <el-link
                          v-if="row.coverUrl"
                          :href="row.coverUrl"
                          target="_blank"
                          type="primary"
                        >
                          查看快照封面
                        </el-link>
                        <span v-else>—</span>
                      </el-descriptions-item>
                      <el-descriptions-item label="食材快照">
                        <pre class="snapshot-json">{{ formatSnapshot(row.ingredients) }}</pre>
                      </el-descriptions-item>
                      <el-descriptions-item label="步骤快照">
                        <pre class="snapshot-json">{{ formatSnapshot(row.steps) }}</pre>
                      </el-descriptions-item>
                    </el-descriptions>
                  </template>
                </el-table-column>
                <el-table-column prop="name" label="菜品" min-width="180" />
                <el-table-column prop="candidateId" label="候选菜 ID" min-width="200" />
                <el-table-column prop="finalServings" label="份数" width="90" />
                <el-table-column label="快照时间" min-width="170">
                  <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            <el-tab-pane label="进度记录">
              <el-timeline>
                <el-timeline-item
                  v-for="event in detail.timeline || []"
                  :key="`${event.type}-${event.occurredAt}`"
                  :timestamp="formatDateTime(event.occurredAt)"
                  placement="top"
                >
                  <div class="font-medium">{{ event.summary }}</div>
                  <div class="text-xs text-gray-400">
                    {{ event.actorType === 'system' ? '系统' : event.actorId || '用户' }}
                  </div>
                </el-timeline-item>
              </el-timeline>
            </el-tab-pane>
          </el-tabs>

        </template>
      </div>
    </el-drawer>
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
import { getMealAdminDetail, getMealAdminList } from '@/api/orderfood/meal'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodMeals' })

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canReadShopping = computed(() =>
  Boolean(btnAuth['orderfood:shopping:read'])
)
const statusOptions = [
  { value: 'collecting', label: '点单中' },
  { value: 'closed', label: '已关闭点单' },
  { value: 'confirmed', label: '采购进行中' },
  { value: 'completed', label: '已完成' },
  { value: 'cancelled', label: '已取消' }
]
const statusLabel = (value) =>
  statusOptions.find((item) => item.value === value)?.label || value || '—'
const statusTag = (value) =>
  ({ collecting: 'primary', closed: 'warning', confirmed: 'success', completed: 'info', cancelled: 'danger' })[
    value
  ] || 'info'
const candidateStatusLabel = (value) =>
  ({ active: '正常', source_deleted: '来源菜品已删除', removed: '已移除' })[value] || value || '—'
const userStatusLabel = (value) =>
  ({ normal: '正常', disabled: '已禁用', cancelled: '已注销' })[value] || value || '—'
const closeSourceLabel = (value) =>
  ({
    creator_action: '发起人主动关闭',
    minute_scan: '每分钟截止扫描',
    service_guard: '小程序业务请求截止守卫',
    legacy: '历史数据，准确来源未记录'
  })[value] ||
  value ||
  '—'
const reasonLabel = (row) => {
  if (row.cancelReason === 'creator_disabled') return '发起人被禁用'
  if (row.cancelReason === 'manual') return '手动取消'
  if (row.closeReason === 'deadline') {
    return `截止时间到达 · ${closeSourceLabel(row.closeSource)}`
  }
  if (row.closeReason === 'manual') {
    return `发起人关闭 · ${closeSourceLabel(row.closeSource)}`
  }
  return '—'
}
const shoppingGenerationLabel = (meal) => {
  if (meal?.shoppingListSummary?.id || meal?.shoppingListId) return '已生成'
  if (meal?.status === 'confirmed' || meal?.status === 'completed') return '应已生成但未找到清单'
  if (
    meal?.status === 'cancelled' &&
    (meal?.cancelledFromStatus === 'confirmed' || meal?.cancelledFromStatus === 'completed')
  ) {
    return '历史采购清单未找到'
  }
  return '尚未确认最终菜单'
}
const shareRevocationLabel = (value) =>
  ({ revoked: '外部采购分享已撤销', no_share: '未生成外部采购分享', not_revoked: '外部采购分享尚未撤销' })[
    value
  ] ||
  value ||
  '撤销结果未记录'
const disabledCreatorRetentionSummary = (meal) =>
  `取消前状态：${statusLabel(meal?.cancelledFromStatus)}。保留只读内容：参与者 ${
    meal?.participants?.length || 0
  } 人、候选菜 ${meal?.candidates?.length || 0} 个、最终菜单快照 ${
    meal?.finalDishSnapshots?.length || 0
  } 份；${shareRevocationLabel(meal?.shareRevocationStatus)}。`
const formatSnapshot = (value) => {
  if (value === null || typeof value === 'undefined') return '—'
  try {
    return JSON.stringify(value, null, 2)
  } catch (_error) {
    return String(value)
  }
}
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )
const routeRange = getShanghaiPresetRange(String(route.query.range || ''))
const defaultSearch = () => ({
  keyword: String(route.query.mealId || ''),
  creatorId: '',
  status: '',
  closeReason: '',
  cancelReason: '',
  createdRange: route.query.timeField === 'created' ? routeRange : [],
  deadlineRange: [],
  joinedRange: route.query.timeField === 'joined' ? routeRange : [],
  confirmedRange: route.query.timeField === 'confirmed' ? routeRange : [],
  completedRange: route.query.timeField === 'completed' ? routeRange : []
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
    const {
      createdRange,
      deadlineRange,
      joinedRange,
      confirmedRange,
      completedRange,
      ...filters
    } = searchInfo.value
    const data = unwrapOrderFoodResponse(
      await getMealAdminList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...filters,
          createdFrom: toShanghaiRFC3339(createdRange?.[0]),
          createdTo: toShanghaiRFC3339(createdRange?.[1]),
          deadlineFrom: toShanghaiRFC3339(deadlineRange?.[0]),
          deadlineTo: toShanghaiRFC3339(deadlineRange?.[1]),
          joinedFrom: toShanghaiRFC3339(joinedRange?.[0]),
          joinedTo: toShanghaiRFC3339(joinedRange?.[1]),
          confirmedFrom: toShanghaiRFC3339(confirmedRange?.[0]),
          confirmedTo: toShanghaiRFC3339(confirmedRange?.[1]),
          completedFrom: toShanghaiRFC3339(completedRange?.[0]),
          completedTo: toShanghaiRFC3339(completedRange?.[1]),
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      '饭局列表加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
  } catch (error) {
    items.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '饭局列表加载失败')
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
    keyword: '',
    creatorId: '',
    status: '',
    closeReason: '',
    cancelReason: '',
    createdRange: [],
    deadlineRange: [],
    joinedRange: [],
    confirmedRange: [],
    completedRange: []
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
const openDetail = async (mealId) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getMealAdminDetail(mealId),
      '饭局详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '饭局详情加载失败')
  } finally {
    detailLoading.value = false
  }
}
const openShopping = (shoppingListId) =>
  router.push({
    name: 'OrderFoodShoppingLists',
    query: { shoppingListId, openShoppingListId: shoppingListId }
  })
const openUser = (userId) =>
  router.push({ name: 'OrderFoodUsers', query: { userId } })
const copyMealID = async (mealId) => {
  try {
    await navigator.clipboard.writeText(mealId)
    ElMessage.success('饭局 ID 已复制')
  } catch {
    ElMessage.warning('复制失败，请手动复制')
  }
}

loadList()
if (route.query.openMealId) {
  openDetail(String(route.query.openMealId))
}
</script>

<style scoped>
.snapshot-json {
  max-height: 280px;
  margin: 0;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
