<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="清单 ID">
          <el-input v-model.trim="searchInfo.shoppingListId" clearable />
        </el-form-item>
        <el-form-item label="饭局 ID">
          <el-input v-model.trim="searchInfo.mealId" clearable />
        </el-form-item>
        <el-form-item label="创建者 ID">
          <el-input v-model.trim="searchInfo.creatorId" clearable />
        </el-form-item>
        <el-form-item label="分享状态">
          <el-select v-model="searchInfo.shareStatus" clearable placeholder="全部" class="w-36">
            <el-option label="可用" value="active" />
            <el-option label="已过期" value="expired" />
            <el-option label="已撤销" value="revoked" />
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
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="mb-4">
        <div class="text-lg font-medium">采购清单</div>
        <div class="mt-1 text-sm text-gray-500">
          管理端仅用于排查清单生成结果和分享状态，不展示可直接访问的完整分享令牌。
        </div>
      </div>
      <el-alert v-if="listError" :title="listError" type="error" show-icon class="mb-4" />
      <el-table v-loading="listLoading" :data="items" row-key="id">
        <el-table-column label="采购清单" min-width="230">
          <template #default="{ row }">
            <div class="font-medium">{{ row.id }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.mealName }} · {{ row.mealId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="创建者" min-width="160">
          <template #default="{ row }">
            <div>{{ row.creator?.nickname || '—' }}</div>
            <div class="text-xs text-gray-400">{{ row.creator?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="采购进度" width="230">
          <template #default="{ row }">
            共 {{ row.totalCount }} / 待采购 {{ row.pendingCount }} / 已完成 {{ row.completedCount }}
          </template>
        </el-table-column>
        <el-table-column label="分享状态" width="105">
          <template #default="{ row }">
            <el-tag :type="shareStatusTag(row.shareStatus)">
              {{ shareStatusLabel(row.shareStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分享过期时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.shareExpiresAt) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="175" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row.id)">查看详情</el-button>
            <el-button
              v-if="canReadMeals"
              link
              type="primary"
              @click="openMeal(row.mealId)"
            >
              查看饭局
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

    <el-drawer v-model="detailVisible" title="采购清单详情" size="760px">
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon />
        <template v-if="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="清单 ID" :span="2">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="饭局">
              <el-button
                v-if="canReadMeals"
                link
                type="primary"
                @click="openMeal(detail.mealId)"
              >
                {{ detail.mealName }} · {{ detail.mealId }}
              </el-button>
              <span v-else>{{ detail.mealName }} · {{ detail.mealId }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="创建者">
              {{ detail.creator?.nickname || '—' }} · {{ detail.creator?.id || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="分享状态">{{ shareStatusLabel(detail.shareStatus) }}</el-descriptions-item>
            <el-descriptions-item label="分享令牌摘要">{{ detail.shareTokenMasked || '未生成' }}</el-descriptions-item>
            <el-descriptions-item label="分享过期时间">
              {{ formatDateTime(detail.shareExpiresAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="分享撤销时间">
              {{ formatDateTime(detail.shareRevokedAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="清单项总数">{{ detail.totalCount }}</el-descriptions-item>
            <el-descriptions-item label="采购进度">
              待采购 {{ detail.pendingCount }} / 已完成 {{ detail.completedCount }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatDateTime(detail.createdAt) }}</el-descriptions-item>
            <el-descriptions-item label="更新时间">{{ formatDateTime(detail.updatedAt) }}</el-descriptions-item>
          </el-descriptions>

          <div class="mt-4 flex flex-wrap gap-2">
            <el-button
              v-if="canReadMeals && detail.mealId"
              @click="openMeal(detail.mealId)"
            >
              查看饭局
            </el-button>
            <el-button
              v-if="canReadUsers && detail.creator?.id"
              @click="openUser(detail.creator.id)"
            >
              查看创建者
            </el-button>
            <el-button @click="copyShoppingListID(detail.id)">复制清单 ID</el-button>
          </div>

          <div class="mt-5 mb-2 font-medium">清单项</div>
          <el-table :data="detail.items || []">
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.completed ? 'success' : 'warning'">
                  {{ row.completed ? '已完成' : '待采购' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="食材" min-width="150" />
            <el-table-column prop="amount" label="用量" width="110" />
            <el-table-column label="来源菜品" min-width="210">
              <template #default="{ row }">{{ (row.sourceDishNames || []).join('、') || '手动添加' }}</template>
            </el-table-column>
            <el-table-column label="备注" min-width="150">
              <template #default="{ row }">{{ row.note || '—' }}</template>
            </el-table-column>
          </el-table>

          <el-collapse class="mt-4">
            <el-collapse-item title="生成来源快照 ID">
              <div v-if="detail.generatedFromSnapshotIds?.length" class="snapshot-list">
                <code v-for="id in detail.generatedFromSnapshotIds" :key="id">{{ id }}</code>
              </div>
              <span v-else>无</span>
            </el-collapse-item>
          </el-collapse>
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
import {
  getShoppingListAdminDetail,
  getShoppingListAdminList
} from '@/api/orderfood/meal'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodShoppingLists' })

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const canReadMeals = computed(() => Boolean(btnAuth['orderfood:meal:read']))
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const shareStatusLabel = (value) =>
  ({ active: '可用', expired: '已过期', revoked: '已撤销' })[value] || value || '—'
const shareStatusTag = (value) =>
  ({ active: 'success', expired: 'warning', revoked: 'danger' })[value] || 'info'
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )
const defaultSearch = () => ({
  shoppingListId: String(route.query.shoppingListId || ''),
  mealId: String(route.query.mealId || ''),
  creatorId: '',
  shareStatus: '',
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
    const { createdRange, ...filters } = searchInfo.value
    const data = unwrapOrderFoodResponse(
      await getShoppingListAdminList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...filters,
          createdFrom: toShanghaiRFC3339(createdRange?.[0]),
          createdTo: toShanghaiRFC3339(createdRange?.[1]),
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      '采购清单加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
  } catch (error) {
    items.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '采购清单加载失败')
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
    shoppingListId: '',
    mealId: '',
    creatorId: '',
    shareStatus: '',
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
const openDetail = async (shoppingListId) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getShoppingListAdminDetail(shoppingListId),
      '采购清单详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '采购清单详情加载失败')
  } finally {
    detailLoading.value = false
  }
}
const openMeal = (mealId) =>
  router.push({ name: 'OrderFoodMeals', query: { mealId, openMealId: mealId } })
const openUser = (userId) =>
  router.push({ name: 'OrderFoodUsers', query: { userId } })
const copyShoppingListID = async (shoppingListId) => {
  try {
    await navigator.clipboard.writeText(shoppingListId)
    ElMessage.success('清单 ID 已复制')
  } catch {
    ElMessage.warning('复制失败，请手动复制')
  }
}

loadList()
if (route.query.openShoppingListId) {
  openDetail(String(route.query.openShoppingListId))
}
</script>

<style scoped>
.snapshot-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
