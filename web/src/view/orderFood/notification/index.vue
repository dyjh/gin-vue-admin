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
        <el-form-item label="通知类型">
          <el-select v-model="searchInfo.type" clearable placeholder="全部" class="w-40">
            <el-option
              v-for="option in typeOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="已读状态">
          <el-select
            v-model="searchInfo.read"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="已读" :value="true" />
            <el-option label="未读" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标对象">
          <div class="flex items-center gap-2">
            <el-input
              v-model.trim="searchInfo.targetType"
              clearable
              placeholder="目标类型"
              class="w-32"
            />
            <el-input
              v-model.trim="searchInfo.targetId"
              clearable
              placeholder="目标 ID"
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
    </div>

    <div class="gva-table-box">
      <div class="mb-5">
        <div class="text-lg font-medium">站内通知</div>
        <div class="mt-1 text-sm text-gray-500">
          只读查看已写入用户通知中心的消息，不提供新建、撤回、重发或修改已读状态。
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
        :data="tableData"
        row-key="id"
        :empty-text="listError ? '加载失败' : '暂无通知'"
      >
        <el-table-column label="通知" min-width="260" fixed="left">
          <template #default="{ row }">
            <div class="font-medium">{{ row.title }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.id }}</div>
          </template>
        </el-table-column>
        <el-table-column label="用户" min-width="190">
          <template #default="{ row }">
            <div>{{ row.user?.nickname || '未命名用户' }}</div>
            <div class="text-xs text-gray-400">{{ row.user?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="130">
          <template #default="{ row }">
            <el-tag effect="plain">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="目标" min-width="190">
          <template #default="{ row }">
            <div>{{ row.targetType || '—' }}</div>
            <div class="text-xs text-gray-400">{{ row.targetId || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="用户已读" width="100">
          <template #default="{ row }">
            <el-tag :type="row.read ? 'success' : 'info'">
              {{ row.read ? '已读' : '未读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="读取时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.readAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row.id)">
              查看详情
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

    <el-drawer
      v-model="detailVisible"
      title="通知详情"
      size="680px"
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
            <el-descriptions-item label="通知 ID">
              <div class="flex items-center justify-between gap-3">
                <span class="break-all">{{ detail.id }}</span>
                <el-button link type="primary" @click="copyText(detail.id, '通知 ID')">
                  复制
                </el-button>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="接收用户">
              {{ detail.user?.nickname || '未命名用户' }}（{{ detail.user?.id }}）
            </el-descriptions-item>
            <el-descriptions-item label="通知类型">
              {{ typeLabel(detail.type) }}
            </el-descriptions-item>
            <el-descriptions-item label="生成来源">
              {{ detail.generationSource }}
            </el-descriptions-item>
            <el-descriptions-item label="标题">{{ detail.title }}</el-descriptions-item>
            <el-descriptions-item label="正文">
              <div class="whitespace-pre-wrap break-words">{{ detail.content }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="目标对象">
              {{ detail.targetType || '—' }} / {{ detail.targetId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="订阅消息">
              {{
                detail.subscribeRequired
                  ? `已关联发送记录 ${detail.subscribeLogId || '—'}`
                  : '不需要订阅消息'
              }}
            </el-descriptions-item>
            <el-descriptions-item label="用户已读">
              {{ detail.read ? `已读于 ${formatDateTime(detail.readAt)}` : '未读' }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(detail.createdAt) }}
            </el-descriptions-item>
          </el-descriptions>
          <div class="mt-4 flex flex-wrap gap-2">
            <el-button
              v-if="canReadUsers && detail.user?.id"
              type="primary"
              plain
              @click="openUser(detail.user.id)"
            >
              查看用户
            </el-button>
            <el-button
              v-if="canOpenTarget(detail.type, detail.targetType, detail.targetId)"
              type="primary"
              plain
              @click="openTarget(detail.type, detail.targetType, detail.targetId)"
            >
              查看目标对象
            </el-button>
            <el-button
              v-if="canReadSubscribeLogs && detail.subscribeLogId"
              type="primary"
              plain
              @click="openSubscribeLog(detail.subscribeLogId)"
            >
              查看订阅发送记录
            </el-button>
          </div>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import {
  getOrderFoodNotificationDetail,
  getOrderFoodNotificationList
} from '@/api/orderfood/message'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodNotifications'
})

const router = useRouter()
const route = useRoute()
const btnAuth = useBtnAuth()
const canReadUsers = Boolean(btnAuth['orderfood:user:read'])
const canReadSubscribeLogs = Boolean(btnAuth['orderfood:subscribe-log:read'])
const typeOptions = [
  { value: 'governance', label: '内容处理' },
  { value: 'discoverability', label: '公开状态' },
  { value: 'points', label: '积分' },
  { value: 'feature_refund', label: '功能退积分' },
  { value: 'meal', label: '饭局' }
]
const createDefaultSearch = () => ({
  userId: String(route.query.userId || ''),
  type: '',
  read: undefined,
  targetType: '',
  targetId: '',
  createdRange: []
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
      await getOrderFoodNotificationList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          userId: searchInfo.value.userId,
          type: searchInfo.value.type,
          read: searchInfo.value.read,
          targetType: searchInfo.value.targetType,
          targetId: searchInfo.value.targetId,
          createdFrom: toShanghaiRFC3339(range[0]),
          createdTo: toShanghaiRFC3339(range[1]),
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      '通知列表加载失败'
    )
    tableData.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    tableData.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '通知列表加载失败')
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
const openDetail = async (notificationId) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getOrderFoodNotificationDetail(notificationId),
      '通知详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '通知详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const copyText = async (value, label) => {
  try {
    await navigator.clipboard.writeText(value)
    ElMessage.success(`${label}已复制`)
  } catch {
    ElMessage.warning('复制失败，请手动复制')
  }
}
const openUser = (userId) => {
  router.push({ name: 'OrderFoodUsers', query: { userId } })
}
const openSubscribeLog = (logId) => {
  router.push({ name: 'OrderFoodSubscribeLogs', query: { logId } })
}
const targetRoute = (notificationType, targetType, targetId) => {
  const id = String(targetId || '').trim()
  if (!id) return null
  if (notificationType === 'governance') {
    return btnAuth['orderfood:governance:read']
      ? {
          name: 'OrderFoodGovernanceRecords',
          query: { targetType, targetId: id }
        }
      : null
  }
  if (notificationType === 'points') {
    return btnAuth['orderfood:points:read']
      ? {
          name: 'OrderFoodPointEntries',
          query: { userId: id }
        }
      : null
  }
  const routes = {
    meal: {
      permission: 'orderfood:meal:read',
      name: 'OrderFoodMeals',
      query: { openMealId: id }
    },
    feature_usage: {
      permission: 'orderfood:ai-usage:read',
      name: 'OrderFoodAIUsages',
      query: { usageId: id }
    },
    dish: {
      permission: 'orderfood:user-dish:private-read',
      name: 'OrderFoodUserDishes',
      query: { openDishId: id }
    },
    recipe: {
      permission: 'orderfood:user-recipe:private-read',
      name: 'OrderFoodUserRecipes',
      query: { openRecipeId: id }
    },
    recommendation: {
      permission: 'orderfood:recommendation:read',
      name: 'OrderFoodRecommendations',
      query: { recommendationId: id }
    },
    user: {
      permission: 'orderfood:user:read',
      name: 'OrderFoodUsers',
      query: { userId: id }
    }
  }
  const target = routes[targetType]
  return target && btnAuth[target.permission] ? target : null
}
const canOpenTarget = (notificationType, targetType, targetId) =>
  Boolean(targetRoute(notificationType, targetType, targetId))
const openTarget = (notificationType, targetType, targetId) => {
  const target = targetRoute(notificationType, targetType, targetId)
  if (target) router.push({ name: target.name, query: target.query })
}
const typeLabel = (type) =>
  typeOptions.find((option) => option.value === type)?.label || type || '—'
const formatDateTime = (value) => (value ? formatDate(value) : '—')

loadList()
</script>
