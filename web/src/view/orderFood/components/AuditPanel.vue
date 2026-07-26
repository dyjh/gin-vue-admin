<template>
  <div>
    <div class="audit-toolbar">
      <div>
        <div class="font-medium">操作审计</div>
        <div class="mt-1 text-xs text-gray-400">
          仅展示当前对象的管理操作；详情包含请求 ID 和变更前后摘要。
        </div>
      </div>
      <el-button :loading="loading" @click="loadAudits">
        {{ loaded ? '刷新' : '加载审计记录' }}
      </el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
      class="mt-3"
    />
    <el-table
      v-if="loaded || loading"
      v-loading="loading"
      :data="items"
      class="mt-3"
      empty-text="暂无操作审计"
    >
      <el-table-column label="动作" prop="action" min-width="170" />
      <el-table-column label="管理员" min-width="130">
        <template #default="{ row }">
          {{ row.administrator?.nickname || row.administrator?.username || '—' }}
        </template>
      </el-table-column>
      <el-table-column label="原因" prop="reason" min-width="180">
        <template #default="{ row }">{{ row.reason || '—' }}</template>
      </el-table-column>
      <el-table-column label="时间" min-width="170">
        <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="90" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row.id)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div v-if="loaded && total > pageSize" class="audit-pagination">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="loadAudits"
      />
    </div>

    <el-dialog v-model="detailVisible" title="审计记录详情" width="760px" append-to-body>
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert
          v-if="detailError"
          :title="detailError"
          type="error"
          show-icon
          :closable="false"
        />
        <template v-else-if="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="动作">{{ detail.action }}</el-descriptions-item>
            <el-descriptions-item label="管理员">
              {{ detail.administrator?.nickname || detail.administrator?.username || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="目标类型">{{ detail.targetType }}</el-descriptions-item>
            <el-descriptions-item label="目标 ID">{{ detail.targetId }}</el-descriptions-item>
            <el-descriptions-item label="请求 ID">{{ detail.requestId || '—' }}</el-descriptions-item>
            <el-descriptions-item label="幂等键">
              {{ detail.idempotencyKey || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="来源 IP">
              {{ detail.sourceIpMasked || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="时间">
              {{ formatDateTime(detail.createdAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="操作原因" :span="2">
              {{ detail.reason || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="用户代理" :span="2">
              {{ detail.userAgentSummary || '—' }}
            </el-descriptions-item>
          </el-descriptions>
          <div class="audit-summary-grid">
            <el-card shadow="never">
              <template #header>变更前摘要</template>
              <pre>{{ jsonText(detail.beforeSummary) }}</pre>
            </el-card>
            <el-card shadow="never">
              <template #header>变更后摘要</template>
              <pre>{{ jsonText(detail.afterSummary) }}</pre>
            </el-card>
          </div>
        </template>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import {
  getOrderFoodAuditLogDetail,
  getOrderFoodAuditLogList
} from '@/api/orderfood/governance'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'
import { formatOrderFoodDateTime } from '@/view/orderFood/utils/time'

defineOptions({ name: 'OrderFoodAuditPanel' })

const props = defineProps({
  targetType: {
    type: String,
    required: true
  },
  targetId: {
    type: String,
    default: ''
  }
})

const page = ref(1)
const pageSize = 20
const total = ref(0)
const items = ref([])
const loading = ref(false)
const loaded = ref(false)
const errorMessage = ref('')
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref(null)

// resetAudits 在面板切换到另一个业务对象时清除上一对象的记录。
const resetAudits = () => {
  page.value = 1
  total.value = 0
  items.value = []
  loaded.value = false
  errorMessage.value = ''
  detailVisible.value = false
  detail.value = null
}

// loadAudits 按当前业务对象分页加载管理操作审计。
const loadAudits = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodAuditLogList({
        page: page.value,
        pageSize,
        targetType: props.targetType,
        targetId: props.targetId || undefined,
        sortBy: 'createdAt',
        sortOrder: 'desc'
      }),
      '审计记录加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
    loaded.value = true
  } catch (error) {
    items.value = []
    total.value = 0
    errorMessage.value = getOrderFoodErrorMessage(error, '审计记录加载失败')
  } finally {
    loading.value = false
  }
}

// openDetail 加载一条审计记录的请求信息和变更摘要。
const openDetail = async (auditLogId) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  detail.value = null
  try {
    detail.value = unwrapOrderFoodResponse(
      await getOrderFoodAuditLogDetail(auditLogId),
      '审计详情加载失败'
    )
  } catch (error) {
    detailError.value = getOrderFoodErrorMessage(error, '审计详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

// formatDateTime 统一按点餐业务的上海时区显示审计时间。
const formatDateTime = (value) => formatOrderFoodDateTime(value) || '—'

// jsonText 将审计摘要格式化为便于核对的只读 JSON。
const jsonText = (value) => JSON.stringify(value ?? null, null, 2)

watch(
  () => [props.targetType, props.targetId],
  resetAudits
)
</script>

<style scoped>
.audit-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.audit-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}

.audit-summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.audit-summary-grid pre {
  max-height: 320px;
  margin: 0;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
}

@media (max-width: 720px) {
  .audit-summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
