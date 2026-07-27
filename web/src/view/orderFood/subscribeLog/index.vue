<template>
  <div>
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="用户 ID">
          <el-input
            v-model.trim="searchInfo.userId"
            clearable
            placeholder="输入用户 ID"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="模板 ID">
          <el-input
            v-model.trim="searchInfo.templateId"
            clearable
            placeholder="输入模板 ID"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="业务场景">
          <el-select
            v-model="searchInfo.scene"
            clearable
            placeholder="全部"
            class="w-40"
          >
            <el-option label="饭局最终结果" value="meal_status" />
          </el-select>
        </el-form-item>
        <el-form-item label="发送状态">
          <el-select v-model="searchInfo.status" clearable placeholder="全部" class="w-32">
            <el-option
              v-for="option in statusOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="请求 ID">
          <el-input
            v-model.trim="searchInfo.requestId"
            clearable
            placeholder="精确请求 ID"
            @keyup.enter="onSubmit"
          />
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
      <div class="mb-5">
        <div class="text-lg font-medium">订阅消息记录</div>
        <div class="mt-1 text-sm text-gray-500">
          只读排查微信订阅消息的授权、模板和发送结果，第一版不提供人工重发。
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
        :empty-text="listError ? '加载失败' : '暂无发送记录'"
      >
        <el-table-column label="发送记录" min-width="220" fixed="left">
          <template #default="{ row }">
            <div class="font-medium">{{ row.id }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.requestId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="用户" min-width="180">
          <template #default="{ row }">
            <div>{{ row.user?.nickname || '未命名用户' }}</div>
            <div class="text-xs text-gray-400">{{ row.user?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="模板" min-width="190">
          <template #default="{ row }">
            <div>{{ row.templateName || '未知模板' }}</div>
            <div class="text-xs text-gray-400">{{ row.templateId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="场景" width="120">
          <template #default>饭局最终结果</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="重试次数" width="100" align="right">
          <template #default="{ row }">{{ row.retryCount }}</template>
        </el-table-column>
        <el-table-column label="微信返回码" min-width="130">
          <template #default="{ row }">{{ row.wechatErrorCode || '—' }}</template>
        </el-table-column>
        <el-table-column label="发送时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.sentAt) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
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
      title="订阅消息发送详情"
      size="760px"
      destroy-on-close
      @closed="detail = null"
    >
      <div v-loading="detailLoading" class="min-h-56">
        <el-alert
          v-if="detailError"
          :title="detailError"
          type="error"
          show-icon
          class="mb-4"
        />
        <template v-if="detail">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="记录 ID">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="请求 ID">
              <div class="flex items-center justify-between gap-3">
                <span class="break-all">{{ detail.requestId }}</span>
                <el-button
                  link
                  type="primary"
                  @click="copyText(detail.requestId, '请求 ID')"
                >
                  复制
                </el-button>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="接收用户">
              {{ detail.user?.nickname || '未命名用户' }}（{{ detail.user?.id }}）
            </el-descriptions-item>
            <el-descriptions-item label="订阅模板">
              {{ detail.templateName || '未知模板' }}（{{ detail.templateId }}）
            </el-descriptions-item>
            <el-descriptions-item label="发送状态">
              {{ statusLabel(detail.status) }}
            </el-descriptions-item>
            <el-descriptions-item label="授权检查">
              {{
                detail.authorizationChecked
                  ? detail.authorizationAvailable
                    ? '已检查，授权可用'
                    : '已检查，授权不可用'
                  : '尚未检查'
              }}
            </el-descriptions-item>
            <el-descriptions-item label="跳转页面">
              {{ detail.targetPage || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="脱敏载荷摘要">
              <pre class="safe-json">{{ formatJSON(detail.safePayloadSummary) }}</pre>
            </el-descriptions-item>
            <el-descriptions-item label="失败摘要">
              {{ detail.errorSummary || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="微信返回码">
              {{ detail.wechatErrorCode || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="创建 / 发送时间">
              {{ formatDateTime(detail.createdAt) }} / {{ formatDateTime(detail.sentAt) }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="mt-5 text-base font-medium">发送尝试</div>
          <el-table :data="detail.attempts || []" class="mt-3" empty-text="尚无发送尝试">
            <el-table-column label="次数" prop="attempt" width="70" />
            <el-table-column label="结果" width="90">
              <template #default="{ row }">{{ statusLabel(row.status) }}</template>
            </el-table-column>
            <el-table-column label="微信返回码" min-width="130">
              <template #default="{ row }">{{ row.wechatErrorCode || '—' }}</template>
            </el-table-column>
            <el-table-column label="失败摘要" min-width="200">
              <template #default="{ row }">{{ row.errorSummary || '—' }}</template>
            </el-table-column>
            <el-table-column label="开始时间" min-width="165">
              <template #default="{ row }">{{ formatDateTime(row.startedAt) }}</template>
            </el-table-column>
            <el-table-column label="结束时间" min-width="165">
              <template #default="{ row }">{{ formatDateTime(row.finishedAt) }}</template>
            </el-table-column>
          </el-table>

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
              v-if="canReadTemplates && detail.templateId"
              type="primary"
              plain
              @click="openTemplate(detail.templateId)"
            >
              查看模板
            </el-button>
            <el-button
              v-if="canReadMeals && detail.relatedMealId"
              type="primary"
              plain
              @click="openMeal(detail.relatedMealId)"
            >
              查看关联饭局
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
  getShanghaiPresetRange,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import {
  getSubscribeLogDetail,
  getSubscribeLogList
} from '@/api/orderfood/message'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodSubscribeLogs'
})

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const canReadUsers = Boolean(btnAuth['orderfood:user:read'])
const canReadTemplates = Boolean(btnAuth['orderfood:subscribe-template:read'])
const canReadMeals = Boolean(btnAuth['orderfood:meal:read'])
const statusOptions = [
  { value: 'pending', label: '等待发送' },
  { value: 'sending', label: '发送中' },
  { value: 'sent', label: '已发送' },
  { value: 'failed', label: '发送失败' }
]
const createDefaultSearch = () => ({
  userId: String(route.query.userId || ''),
  templateId: String(route.query.templateId || ''),
  scene: String(route.query.scene || ''),
  status: String(route.query.status || ''),
  requestId: String(route.query.requestId || ''),
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
      await getSubscribeLogList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          userId: searchInfo.value.userId,
          templateId: searchInfo.value.templateId,
          scene: searchInfo.value.scene,
          status: searchInfo.value.status,
          requestId: searchInfo.value.requestId,
          createdFrom: toShanghaiRFC3339(range[0]),
          createdTo: toShanghaiRFC3339(range[1]),
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      '订阅消息记录加载失败'
    )
    tableData.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    tableData.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '订阅消息记录加载失败')
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
    templateId: '',
    scene: '',
    status: '',
    requestId: '',
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
const openDetail = async (logId) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getSubscribeLogDetail(logId),
      '发送详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '发送详情加载失败')
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
const openTemplate = (templateId) => {
  router.push({ name: 'OrderFoodSubscribeTemplates', query: { templateId } })
}
const openMeal = (mealId) => {
  router.push({ name: 'OrderFoodMeals', query: { openMealId: mealId } })
}
const statusLabel = (status) =>
  statusOptions.find((option) => option.value === status)?.label ||
  (status === 'sent' ? '已发送' : status === 'failed' ? '发送失败' : status || '—')
const statusTagType = (status) =>
  ({ sent: 'success', failed: 'danger', sending: 'warning', pending: 'info' })[
    status
  ] || 'info'
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const formatJSON = (value) => (value ? JSON.stringify(value, null, 2) : '—')

loadList().then(() => {
  if (route.query.logId) {
    openDetail(String(route.query.logId))
  }
})
</script>

<style scoped>
.safe-json {
  margin: 0;
  max-height: 260px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
  line-height: 1.6;
}
</style>
