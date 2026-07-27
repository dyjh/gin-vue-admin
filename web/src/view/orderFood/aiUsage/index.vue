<template>
  <div class="order-food-list-page">
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="用户 ID">
          <el-input v-model.trim="searchInfo.userId" clearable />
        </el-form-item>
        <el-form-item label="能力">
          <el-select v-model="searchInfo.capabilityCode" clearable placeholder="全部" class="w-48">
            <el-option
              v-for="item in capabilityOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="供应商 ID">
          <el-input v-model.trim="searchInfo.providerId" clearable />
        </el-form-item>
        <el-form-item label="模型 ID">
          <el-input v-model.trim="searchInfo.modelId" clearable />
        </el-form-item>
        <el-form-item label="执行状态">
          <el-select v-model="searchInfo.executionStatus" clearable placeholder="全部" class="w-32">
            <el-option label="等待中" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="成功" value="succeeded" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-form-item>
        <el-form-item label="计费状态">
          <el-select v-model="searchInfo.billingStatus" clearable placeholder="全部" class="w-36">
            <el-option label="未扣积分" value="not_charged" />
            <el-option label="已扣积分" value="charged" />
            <el-option label="待退积分" value="refund_pending" />
            <el-option label="已退积分" value="refunded" />
          </el-select>
        </el-form-item>
        <el-form-item label="请求 ID">
          <el-input v-model.trim="searchInfo.requestId" clearable />
        </el-form-item>
        <el-form-item label="幂等键">
          <el-input v-model.trim="searchInfo.idempotencyKey" clearable />
        </el-form-item>
        <el-form-item label="创建时间">
          <el-date-picker
            v-model="createdRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
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
        <div class="text-lg font-medium">AI 调用记录</div>
        <div class="mt-1 text-sm text-gray-500">
          用于排查模型调用和积分状态；本页不提供重新执行或人工退积分操作。
        </div>
      </div>
      <el-alert v-if="listError" :title="listError" type="error" show-icon class="mb-4" />
      <el-table v-loading="listLoading" :data="items" row-key="id">
        <el-table-column label="调用" min-width="235">
          <template #default="{ row }">
            <div class="font-medium">{{ capabilityLabel(row.capabilityCode) }}</div>
            <div class="mt-1 text-xs text-gray-400">{{ row.id }}</div>
          </template>
        </el-table-column>
        <el-table-column label="用户" min-width="150">
          <template #default="{ row }">
            <div>{{ row.user?.nickname || '—' }}</div>
            <div class="text-xs text-gray-400">{{ row.user?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="供应商 / 模型" min-width="190">
          <template #default="{ row }">
            {{ row.providerName || '—' }} / {{ row.modelName || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="提示词" min-width="180">
          <template #default="{ row }">
            <div>{{ promptModeLabel(row.promptMode) }}</div>
            <div class="text-xs text-gray-400">{{ shortHash(row.promptHash) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="执行" width="100">
          <template #default="{ row }">
            <el-tag :type="executionTag(row.executionStatus)">
              {{ executionLabel(row.executionStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="计费" width="105">
          <template #default="{ row }">{{ billingLabel(row.billingStatus) }}</template>
        </el-table-column>
        <el-table-column label="耗时" width="105">
          <template #default="{ row }">
            {{ row.durationMs === null ? '—' : `${row.durationMs} ms` }}
          </template>
        </el-table-column>
        <el-table-column label="Token / 图片" min-width="145">
          <template #default="{ row }">
            {{ row.inputTokens ?? '—' }} / {{ row.outputTokens ?? '—' }}
            <span v-if="row.imageCount"> · 图 {{ row.imageCount }}</span>
          </template>
        </el-table-column>
        <el-table-column label="积分 / 成本" min-width="135">
          <template #default="{ row }">
            {{ row.pointCost }} 积分 / ¥{{ row.estimatedCostCny || '0.000000' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
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

    <el-drawer v-model="detailVisible" title="AI 调用详情" size="760px" destroy-on-close>
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon class="mb-4" />
        <template v-if="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="记录 ID" :span="2">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="能力">
              {{ capabilityLabel(detail.capabilityCode) }}
            </el-descriptions-item>
            <el-descriptions-item label="用户">
              {{ detail.user?.nickname || '—' }} · {{ detail.user?.id || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="供应商">
              {{ detail.providerName || '—' }} · {{ detail.providerId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="模型">
              {{ detail.modelName || '—' }} · {{ detail.modelId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="执行状态">
              {{ executionLabel(detail.executionStatus) }}
            </el-descriptions-item>
            <el-descriptions-item label="计费状态">
              {{ billingLabel(detail.billingStatus) }}
            </el-descriptions-item>
            <el-descriptions-item label="请求 ID" :span="2">
              {{ detail.requestId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="幂等键" :span="2">
              {{ detail.idempotencyKey || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="提示词">
              {{ promptModeLabel(detail.promptMode) }} · {{ shortHash(detail.promptHash) }}
            </el-descriptions-item>
            <el-descriptions-item label="耗时">
              {{ detail.durationMs === null ? '—' : `${detail.durationMs} ms` }}
            </el-descriptions-item>
            <el-descriptions-item label="Token">
              输入 {{ detail.inputTokens ?? '—' }} / 输出 {{ detail.outputTokens ?? '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="估算成本">
              ¥{{ detail.estimatedCostCny || '0.000000' }}
            </el-descriptions-item>
            <el-descriptions-item label="失败分类">
              {{ detail.failureCategory || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="失败说明">
              {{ detail.failureSummary || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="扣积分流水">
              {{ detail.chargedPointEntryId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="退款流水">
              {{ detail.refundPointEntryId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="敏感内容">
              {{ detail.sensitiveContentCleared ? '已永久清除' : '保留中' }}
            </el-descriptions-item>
            <el-descriptions-item label="记录版本">v{{ detail.version }}</el-descriptions-item>
          </el-descriptions>

          <div class="section-title">调用时间线</div>
          <el-timeline class="usage-timeline">
            <el-timeline-item
              v-for="event in detail.timeline || []"
              :key="`${event.type}-${event.occurredAt}`"
              :timestamp="formatDateTime(event.occurredAt)"
              placement="top"
            >
              {{ event.label }}
            </el-timeline-item>
          </el-timeline>

          <div class="detail-actions">
            <el-button v-if="canReadUsers" @click="openUser">
              查看用户
            </el-button>
            <el-button v-if="canReadCapabilities" @click="openCapability">
              查看能力配置
            </el-button>
            <el-button v-if="canReadPoints" @click="openPointEntries">
              查看积分流水
            </el-button>
            <el-button v-if="canReadProviders && detail.providerId" @click="openProvider">
              查看供应商
            </el-button>
            <el-button v-if="canReadModels && detail.modelId" @click="openModel">
              查看模型
            </el-button>
            <el-button :disabled="!detail.requestId" @click="copyRequestID">
              复制请求 ID
            </el-button>
            <el-button
              v-if="canReadSensitive && !detail.sensitiveContentCleared"
              :loading="sensitiveLoading"
              @click="loadSensitive"
            >
              查看原始内容
            </el-button>
            <el-button
              v-if="canDeleteSensitive && !detail.sensitiveContentCleared"
              type="danger"
              plain
              @click="clearSensitive"
            >
              清除敏感内容
            </el-button>
          </div>

          <el-alert
            v-if="sensitiveVisible"
            title="以下内容只用于故障排查，读取操作已记入审计日志。"
            type="warning"
            show-icon
            class="mt-4"
          />
          <template v-if="sensitiveVisible">
            <div class="json-title">原始输入</div>
            <pre class="json-view">{{ prettyJSON(detail.originalInput) }}</pre>
            <div class="json-title">图片元信息</div>
            <pre class="json-view">{{ prettyJSON(detail.imageMetadata) }}</pre>
            <div class="json-title">模型输出</div>
            <pre class="json-view">{{ prettyJSON(detail.modelOutput) }}</pre>
          </template>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  getShanghaiPresetRange,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import {
  deleteAIUsageSensitiveContent,
  getAIUsageDetail,
  getAIUsageList
} from '@/api/orderfood/operations'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodAIUsages' })

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canReadCapabilities = computed(() =>
  Boolean(btnAuth['orderfood:capability:read'])
)
const canReadPoints = computed(() => Boolean(btnAuth['orderfood:points:read']))
const canReadProviders = computed(() => Boolean(btnAuth['orderfood:provider:read']))
const canReadModels = computed(() => Boolean(btnAuth['orderfood:model:read']))
const canReadSensitive = computed(() =>
  Boolean(btnAuth['orderfood:ai-usage:sensitive-read'])
)
const canDeleteSensitive = computed(() =>
  Boolean(btnAuth['orderfood:ai-usage:sensitive-delete'])
)
const capabilityOptions = [
  { value: 'dish_text_extract', label: '菜品文本解析' },
  { value: 'recipe_image_extract', label: '菜谱长截图解析' },
  { value: 'dish_cover_create', label: '菜品封面生成' },
  { value: 'checkin_image_analyze', label: '打卡图片分析' },
  { value: 'meal_suggest', label: '饭局菜品建议' },
  { value: 'prep_sequence', label: '备菜顺序生成' }
]
const capabilityLabel = (value) =>
  capabilityOptions.find((item) => item.value === value)?.label || value || '—'
const executionLabels = {
  pending: '等待中',
  processing: '处理中',
  succeeded: '成功',
  failed: '失败'
}
const executionLabel = (value) => executionLabels[value] || value || '—'
const executionTag = (value) =>
  ({ succeeded: 'success', failed: 'danger', processing: 'warning' })[value] || 'info'
const billingLabels = {
  not_charged: '未扣积分',
  charged: '已扣积分',
  refund_pending: '待退积分',
  refunded: '已退积分'
}
const billingLabel = (value) => billingLabels[value] || value || '—'
const promptModeLabel = (value) =>
  ({ default: '平台默认', custom: '自定义' })[value] || value || '—'
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const shortHash = (value) => (value ? `${value.slice(0, 12)}…` : '—')
const prettyJSON = (value) => JSON.stringify(value ?? null, null, 2)
const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )
const rangeFromRoute = (rangeCode) => getShanghaiPresetRange(rangeCode)
const defaultSearch = () => ({
  userId: String(route.query.userId || ''),
  capabilityCode: String(route.query.capabilityCode || ''),
  providerId: String(route.query.providerId || ''),
  modelId: String(route.query.modelId || ''),
  executionStatus: String(route.query.executionStatus || ''),
  billingStatus: String(route.query.billingStatus || ''),
  requestId: String(route.query.requestId || ''),
  idempotencyKey: String(route.query.idempotencyKey || '')
})
const searchInfo = ref(defaultSearch())
const createdRange = ref(rangeFromRoute(String(route.query.range || '')))
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const items = ref([])
const listLoading = ref(false)
const listError = ref('')
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref(null)
const sensitiveLoading = ref(false)
const sensitiveVisible = ref(false)

const loadList = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const [createdFrom, createdTo] = createdRange.value || []
    const data = unwrapOrderFoodResponse(
      await getAIUsageList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...searchInfo.value,
          createdFrom: toShanghaiRFC3339(createdFrom),
          createdTo: toShanghaiRFC3339(createdTo),
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      'AI 调用记录加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
  } catch (error) {
    listError.value = getOrderFoodErrorMessage(error, 'AI 调用记录加载失败')
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
    capabilityCode: '',
    providerId: '',
    modelId: '',
    executionStatus: '',
    billingStatus: '',
    requestId: '',
    idempotencyKey: ''
  }
  createdRange.value = []
  onSubmit()
}
const handleSizeChange = () => {
  page.value = 1
  loadList()
}
const openDetail = async (usageId) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  detail.value = null
  sensitiveVisible.value = false
  try {
    detail.value = unwrapOrderFoodResponse(
      await getAIUsageDetail(usageId, { includeSensitiveContent: false }),
      'AI 调用详情加载失败'
    )
  } catch (error) {
    detailError.value = getOrderFoodErrorMessage(error, 'AI 调用详情加载失败')
  } finally {
    detailLoading.value = false
  }
}
const loadSensitive = async () => {
  if (!detail.value) return
  sensitiveLoading.value = true
  try {
    detail.value = unwrapOrderFoodResponse(
      await getAIUsageDetail(detail.value.id, { includeSensitiveContent: true }),
      '原始内容加载失败'
    )
    sensitiveVisible.value = true
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '原始内容加载失败'))
  } finally {
    sensitiveLoading.value = false
  }
}
const openUser = () => {
  if (!detail.value?.user?.id) return
  router.push({
    name: 'OrderFoodUsers',
    query: { userId: detail.value.user.id }
  })
}
const openPointEntries = () => {
  if (!detail.value?.id) return
  router.push({
    name: 'OrderFoodPointEntries',
    query: {
      userId: detail.value.user?.id || undefined,
      relatedObjectType: 'feature_usage',
      relatedObjectId: detail.value.id
    }
  })
}
const openCapability = () => {
  if (!detail.value?.capabilityCode) return
  router.push({
    name: 'OrderFoodAiCapabilityConfig',
    query: { capabilityCode: detail.value.capabilityCode }
  })
}
const openProvider = () => {
  if (!detail.value?.providerId) return
  router.push({
    name: 'OrderFoodAiProviders',
    query: { providerId: detail.value.providerId }
  })
}
const openModel = () => {
  if (!detail.value?.modelId) return
  router.push({
    name: 'OrderFoodAiModels',
    query: { modelId: detail.value.modelId }
  })
}
const copyRequestID = async () => {
  if (!detail.value?.requestId) return
  try {
    await navigator.clipboard.writeText(detail.value.requestId)
    ElMessage.success('请求 ID 已复制')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}
const clearSensitive = async () => {
  if (!detail.value) return
  try {
    const { value } = await ElMessageBox.prompt(
      '清除后原始输入、图片元信息和模型输出无法恢复，请填写原因。',
      '清除敏感内容',
      {
        confirmButtonText: '确认清除',
        cancelButtonText: '取消',
        inputPlaceholder: '请输入 4–200 个字符',
        inputValidator: (text) => {
          const length = String(text || '').trim().length
          return (length >= 4 && length <= 200) || '请输入 4–200 个字符'
        }
      }
    )
    unwrapOrderFoodResponse(
      await deleteAIUsageSensitiveContent(detail.value.id, {
        reason: String(value).trim(),
        expectedVersion: detail.value.version
      }),
      '敏感内容清除失败'
    )
    ElMessage.success('敏感内容已永久清除')
    await openDetail(detail.value.id)
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('记录已变化，已重新加载最新详情')
      await openDetail(detail.value.id)
      return
    }
    ElMessage.error(getOrderFoodErrorMessage(error, '敏感内容清除失败'))
  }
}

onMounted(async () => {
  await loadList()
  if (route.query.usageId) {
    await openDetail(String(route.query.usageId))
  }
})
</script>

<style scoped>
.detail-actions {
  display: flex;
  gap: 10px;
  margin-top: 16px;
}

.section-title {
  margin: 20px 0 12px;
  font-weight: 600;
}

.usage-timeline {
  padding-left: 4px;
}

.json-title {
  margin: 18px 0 8px;
  font-weight: 600;
}

.json-view {
  max-height: 320px;
  padding: 12px;
  overflow: auto;
  color: var(--el-text-color-primary);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--el-fill-color-light);
  border-radius: 6px;
}
</style>
