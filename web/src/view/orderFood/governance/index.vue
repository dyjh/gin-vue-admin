<template>
  <div>
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="目标类型">
          <el-select v-model="searchInfo.targetType" clearable placeholder="全部" class="w-36">
            <el-option v-for="item in targetOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标 ID">
          <el-input v-model.trim="searchInfo.targetId" clearable />
        </el-form-item>
        <el-form-item label="处理动作">
          <el-select v-model="searchInfo.action" clearable placeholder="全部" class="w-48">
            <el-option v-for="item in actionOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="违规类型">
          <el-input v-model.trim="searchInfo.violationType" clearable />
        </el-form-item>
        <el-form-item label="任务状态">
          <el-select v-model="searchInfo.jobStatus" clearable placeholder="全部" class="w-40">
            <el-option v-for="item in jobStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="管理员 ID">
          <el-input v-model.trim="searchInfo.administratorId" clearable />
        </el-form-item>
        <el-form-item label="执行时间">
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
        <div class="text-lg font-medium">违规处理记录</div>
        <div class="mt-1 text-sm text-gray-500">
          记录不可编辑或删除；复制链任务仅允许重试失败项，不重复处理已成功项目。
        </div>
      </div>
      <el-alert v-if="listError" :title="listError" type="error" show-icon class="mb-4" />
      <el-table v-loading="listLoading" :data="items" row-key="id">
        <el-table-column label="目标" min-width="220">
          <template #default="{ row }">
            <div class="font-medium">{{ row.targetLabel || '—' }}</div>
            <div class="mt-1 text-xs text-gray-400">
              {{ targetLabel(row.targetType) }} · {{ row.targetId }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="处理动作" min-width="220">
          <template #default="{ row }">
            <div class="flex flex-wrap gap-1">
              <el-tag v-for="action in row.actions || []" :key="action" size="small">
                {{ actionLabel(action) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="违规与原因" min-width="230">
          <template #default="{ row }">
            <div>{{ row.violationType }}</div>
            <div class="mt-1 text-xs text-gray-500">{{ row.reason }}</div>
          </template>
        </el-table-column>
        <el-table-column label="级别" width="95">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'serious' ? 'danger' : 'warning'">
              {{ row.severity === 'serious' ? '严重' : '一般' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="影响数量" width="100" align="right">
          <template #default="{ row }">{{ row.affectedCount }}</template>
        </el-table-column>
        <el-table-column label="任务进度" min-width="190">
          <template #default="{ row }">
            <el-tag v-if="row.jobStatus" :type="jobStatusTag(row.jobStatus)">
              {{ jobStatusLabel(row.jobStatus) }}
            </el-tag>
            <span v-else>同步完成</span>
            <div v-if="row.jobProgress" class="mt-1 text-xs text-gray-500">
              成功 {{ row.jobProgress.succeededCount }} / 失败
              {{ row.jobProgress.failedCount }} / 待处理
              {{ row.jobProgress.pendingCount }} / 共
              {{ row.jobProgress.totalCount }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="执行管理员" min-width="150">
          <template #default="{ row }">
            {{ row.administrator?.nickname || row.administrator?.username || '—' }}
            <div class="text-xs text-gray-400">{{ row.administrator?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="执行时间" min-width="170">
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

    <el-drawer v-model="detailVisible" title="违规处理详情" size="760px">
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon />
        <template v-if="detail">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="记录 ID">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="目标">
              {{ targetLabel(detail.targetType) }} · {{ detail.targetLabel }} · {{ detail.targetId }}
            </el-descriptions-item>
            <el-descriptions-item label="动作">
              {{ (detail.actions || []).map(actionLabel).join('、') }}
            </el-descriptions-item>
            <el-descriptions-item label="违规类型">{{ detail.violationType }}</el-descriptions-item>
            <el-descriptions-item label="原因">{{ detail.reason }}</el-descriptions-item>
            <el-descriptions-item label="请求 ID">{{ detail.requestId }}</el-descriptions-item>
            <el-descriptions-item label="幂等键">{{ detail.idempotencyKey }}</el-descriptions-item>
            <el-descriptions-item label="通知 ID">
              {{ (detail.notificationIds || []).join('、') || '—' }}
            </el-descriptions-item>
          </el-descriptions>
          <div class="mt-4">
            <el-button
              v-if="canAppendProcessing"
              type="primary"
              @click="openAppendProcessing"
            >
              对同一目标追加处理
            </el-button>
          </div>

          <el-card v-if="detail.job" shadow="never" class="mt-4">
            <template #header>
              <div class="flex items-center justify-between">
                <span>复制链任务</span>
                <el-button link type="primary" @click="refreshJob">刷新进度</el-button>
              </div>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="任务 ID">{{ activeJob.id }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                {{ jobStatusLabel(activeJob.status) }}
              </el-descriptions-item>
              <el-descriptions-item label="成功">{{ activeJob.succeededCount }}</el-descriptions-item>
              <el-descriptions-item label="失败">{{ activeJob.failedCount }}</el-descriptions-item>
              <el-descriptions-item label="待处理">{{ activeJob.pendingCount }}</el-descriptions-item>
              <el-descriptions-item label="总数">{{ activeJob.totalCount }}</el-descriptions-item>
              <el-descriptions-item label="最近错误" :span="2">
                {{ activeJob.lastErrorSummary || '—' }}
              </el-descriptions-item>
            </el-descriptions>
            <el-button
              v-if="canRetry && canCascade"
              class="mt-3"
              type="danger"
              :disabled="!activeJob.failedCount"
              @click="openRetry"
            >
              重试失败项
            </el-button>
            <el-table
              :data="activeJob.items || []"
              class="mt-4"
              row-key="id"
              empty-text="暂无任务项"
            >
              <el-table-column label="菜品 / 用户" min-width="210">
                <template #default="{ row }">
                  <div>{{ row.dishId }}</div>
                  <div class="text-xs text-gray-400">{{ row.userId }}</div>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="jobItemStatusTag(row.status)" size="small">
                    {{ jobItemStatusLabel(row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="attemptCount" label="执行次数" width="90" align="right" />
              <el-table-column label="最近结果" min-width="180">
                <template #default="{ row }">
                  <div>{{ row.lastErrorSummary || '—' }}</div>
                  <div class="text-xs text-gray-400">
                    {{ formatDateTime(row.processedAt || row.updatedAt) }}
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-tabs class="mt-4">
            <el-tab-pane label="影响快照">
              <pre class="json-box">{{ jsonText(detail.impactSnapshot) }}</pre>
            </el-tab-pane>
            <el-tab-pane label="处理前">
              <pre class="json-box">{{ jsonText(detail.beforeSummary) }}</pre>
            </el-tab-pane>
            <el-tab-pane label="处理后">
              <pre class="json-box">{{ jsonText(detail.afterSummary) }}</pre>
            </el-tab-pane>
            <el-tab-pane v-if="canReadAudit" label="审计记录" lazy>
              <el-button :loading="auditLoading" @click="loadAudits">加载审计记录</el-button>
              <el-table :data="audits" class="mt-3">
                <el-table-column prop="action" label="动作" min-width="160" />
                <el-table-column prop="requestId" label="请求 ID" min-width="190" />
                <el-table-column label="时间" min-width="170">
                  <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
                </el-table-column>
                <el-table-column label="操作" width="90" fixed="right">
                  <template #default="{ row }">
                    <el-button link type="primary" @click="openAuditDetail(row)">
                      查看
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
          </el-tabs>
        </template>
      </div>
    </el-drawer>

    <el-dialog v-model="auditDetailVisible" title="审计记录详情" width="760px">
      <div v-loading="auditDetailLoading">
        <el-alert
          v-if="auditDetailError"
          type="error"
          :title="auditDetailError"
          show-icon
          :closable="false"
          class="mb-4"
        />
        <template v-if="auditDetail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="动作">{{ auditDetail.action }}</el-descriptions-item>
            <el-descriptions-item label="管理员">
              {{ auditDetail.administrator?.nickname || auditDetail.administrator?.username || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="目标类型">{{ auditDetail.targetType }}</el-descriptions-item>
            <el-descriptions-item label="目标 ID">{{ auditDetail.targetId }}</el-descriptions-item>
            <el-descriptions-item label="请求 ID">{{ auditDetail.requestId }}</el-descriptions-item>
            <el-descriptions-item label="时间">{{ formatDateTime(auditDetail.createdAt) }}</el-descriptions-item>
            <el-descriptions-item label="来源 IP">{{ auditDetail.sourceIpMasked || '—' }}</el-descriptions-item>
            <el-descriptions-item label="操作原因">{{ auditDetail.reason || '—' }}</el-descriptions-item>
            <el-descriptions-item label="客户端摘要" :span="2">
              {{ auditDetail.userAgentSummary || '—' }}
            </el-descriptions-item>
          </el-descriptions>
          <el-tabs class="mt-4">
            <el-tab-pane label="变更前">
              <pre class="json-box">{{ jsonText(auditDetail.beforeSummary) }}</pre>
            </el-tab-pane>
            <el-tab-pane label="变更后">
              <pre class="json-box">{{ jsonText(auditDetail.afterSummary) }}</pre>
            </el-tab-pane>
          </el-tabs>
        </template>
      </div>
    </el-dialog>

    <el-dialog v-model="retryVisible" title="重试复制链失败项" width="520px">
      <el-alert
        :title="`将重试 ${activeJob?.failedCount || 0} 项，已成功的 ${activeJob?.succeededCount || 0} 项不会重复执行。`"
        type="warning"
        show-icon
        :closable="false"
        class="mb-4"
      />
      <el-form label-width="90px">
        <el-form-item label="重试原因">
          <el-input v-model.trim="retryReason" type="textarea" maxlength="200" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="retryVisible = false">取消</el-button>
        <el-button type="danger" :loading="retryLoading" @click="submitRetry">确认重试</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="appendVisible"
      title="对同一目标追加处理"
      width="680px"
      destroy-on-close
      @closed="resetAppendProcessing"
    >
      <el-alert
        title="追加处理会基于目标当前状态重新生成影响预览，不复用历史快照。"
        type="warning"
        show-icon
        :closable="false"
        class="mb-4"
      />
      <el-form
        ref="appendFormRef"
        :model="appendForm"
        :rules="appendRules"
        label-width="94px"
      >
        <el-form-item label="当前目标">
          {{ targetLabel(detail?.targetType) }} · {{ detail?.targetLabel }} ·
          {{ detail?.targetId }}
        </el-form-item>
        <el-form-item label="处理动作" prop="action">
          <el-radio-group v-model="appendForm.action" @change="handleAppendActionChange">
            <el-radio
              v-for="option in appendActionOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="违规类型" prop="violationType">
          <el-input
            v-model.trim="appendForm.violationType"
            maxlength="40"
            @input="clearAppendPreview"
          />
        </el-form-item>
        <el-form-item label="严重程度" prop="severity">
          <el-radio-group v-model="appendForm.severity" @change="clearAppendPreview">
            <el-radio-button value="normal" :disabled="appendIsCascade">一般</el-radio-button>
            <el-radio-button value="serious">严重</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理原因" prop="reason">
          <el-input
            v-model.trim="appendForm.reason"
            type="textarea"
            :rows="4"
            maxlength="300"
            show-word-limit
            placeholder="请输入 4-300 字处理原因"
            @input="clearAppendPreview"
          />
        </el-form-item>
        <el-form-item v-if="appendIsCascade" label="确认词">
          <div class="w-full">
            <el-input
              v-model="appendForm.confirmText"
              maxlength="20"
              placeholder="请输入：确认处理复制链"
            />
            <div class="mt-1 text-xs text-red-500">
              必须完整输入“确认处理复制链”。
            </div>
          </div>
        </el-form-item>
      </el-form>

      <el-card v-if="appendPreview" shadow="never">
        <template #header>最新影响预览</template>
        <el-descriptions :column="2">
          <el-descriptions-item label="推荐记录">
            {{ appendPreview.recommendationCount }}
          </el-descriptions-item>
          <el-descriptions-item label="复制菜品">
            {{ appendPreview.copiedDishCount }}
          </el-descriptions-item>
          <el-descriptions-item label="影响用户">
            {{ appendPreview.affectedUserCount }}
          </el-descriptions-item>
          <el-descriptions-item label="通知数量">
            {{ appendPreview.notificationCount }}
          </el-descriptions-item>
          <el-descriptions-item label="执行方式">
            {{ appendPreview.asynchronous ? '异步任务' : '同步事务' }}
          </el-descriptions-item>
          <el-descriptions-item label="有效期">
            {{ formatDateTime(appendPreview.expiresAt) }}
          </el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-for="warning in appendPreview.warnings || []"
          :key="warning"
          :title="warning"
          type="warning"
          show-icon
          :closable="false"
          class="mt-2"
        />
      </el-card>

      <template #footer>
        <el-button @click="appendVisible = false">取消</el-button>
        <el-button :loading="appendLoading" @click="previewAppendProcessing">
          {{ appendPreview ? '重新预览' : '预览影响' }}
        </el-button>
        <el-button
          v-if="appendPreview"
          type="danger"
          :loading="appendLoading"
          :disabled="appendIsCascade && !appendConfirmValid"
          @click="executeAppendProcessing"
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
import { useRoute } from 'vue-router'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import {
  executeOrderFoodGovernanceAction,
  getOrderFoodAuditLogDetail,
  getOrderFoodAuditLogList,
  getOrderFoodGovernanceJob,
  getOrderFoodGovernanceRecordDetail,
  getOrderFoodGovernanceRecordList,
  previewOrderFoodGovernanceAction,
  retryOrderFoodGovernanceJob
} from '@/api/orderfood/governance'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodGovernanceRecords' })

const route = useRoute()
const btnAuth = useBtnAuth()
const canExecute = computed(() =>
  Boolean(btnAuth['orderfood:governance:execute'])
)
const canRetry = computed(() => Boolean(btnAuth['orderfood:governance:retry']))
const canCascade = computed(() => Boolean(btnAuth['orderfood:governance:cascade']))
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))
const targetOptions = [
  { value: 'dish', label: '用户菜品' },
  { value: 'official_dish', label: '官方菜品' },
  { value: 'recipe', label: '菜谱' },
  { value: 'checkin', label: '打卡' }
]
const actionOptions = [
  { value: 'disable_discoverability', label: '关闭允许被发现' },
  { value: 'soft_delete_dish', label: '软删除用户菜品' },
  { value: 'soft_delete_official_dish', label: '软删除官方菜品' },
  { value: 'soft_delete_recipe', label: '软删除菜谱' },
  { value: 'soft_delete_checkin', label: '软删除打卡' },
  { value: 'delete_copy_chain', label: '处理复制链' }
]
const jobStatusOptions = [
  { value: 'abnormal', label: '异常（处理中/部分成功/失败）' },
  { value: 'pending', label: '待处理' },
  { value: 'processing', label: '处理中' },
  { value: 'partially_succeeded', label: '部分成功' },
  { value: 'succeeded', label: '已成功' },
  { value: 'failed', label: '失败' }
]
const labelFrom = (options, value) =>
  options.find((item) => item.value === value)?.label || value || '—'
const targetLabel = (value) => labelFrom(targetOptions, value)
const actionLabel = (value) => labelFrom(actionOptions, value)
const jobStatusLabel = (value) => labelFrom(jobStatusOptions, value)
const jobStatusTag = (value) =>
  ({ succeeded: 'success', failed: 'danger', partially_succeeded: 'warning', processing: 'primary', pending: 'info' })[
    value
  ] || 'info'
const jobItemStatusLabel = (value) =>
  ({ pending: '待处理', succeeded: '成功', failed: '失败' })[value] || value || '—'
const jobItemStatusTag = (value) =>
  ({ pending: 'info', succeeded: 'success', failed: 'danger' })[value] || 'info'
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const jsonText = (value) => JSON.stringify(value ?? null, null, 2)
const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(
      ([, value]) => value !== '' && value !== null && typeof value !== 'undefined'
    )
  )
const emptySearch = () => ({
  targetType: '',
  targetId: '',
  action: '',
  violationType: '',
  jobStatus: '',
  administratorId: '',
  createdRange: []
})
const defaultSearch = () => ({
  ...emptySearch(),
  targetType: String(route.query.targetType || ''),
  targetId: String(route.query.targetId || ''),
  jobStatus: String(route.query.jobStatus || '')
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
    const createdRange = searchInfo.value.createdRange || []
    const data = unwrapOrderFoodResponse(
      await getOrderFoodGovernanceRecordList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...searchInfo.value,
          createdRange: undefined,
          createdFrom: toShanghaiRFC3339(createdRange[0]),
          createdTo: toShanghaiRFC3339(createdRange[1]),
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      '违规处理记录加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
  } catch (error) {
    items.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '违规处理记录加载失败')
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
const canAppendProcessing = computed(
  () =>
    canExecute.value &&
    Boolean(detail.value?.targetExists) &&
    Number(detail.value?.targetVersion || 0) > 0
)
const activeJob = computed(() => detail.value?.job || null)
const openDetail = async (recordId) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  audits.value = []
  try {
    detail.value = unwrapOrderFoodResponse(
      await getOrderFoodGovernanceRecordDetail(recordId),
      '违规处理详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '违规处理详情加载失败')
  } finally {
    detailLoading.value = false
  }
}
const refreshJob = async () => {
  if (!activeJob.value) return
  try {
    detail.value.job = unwrapOrderFoodResponse(
      await getOrderFoodGovernanceJob(activeJob.value.id),
      '任务进度加载失败'
    )
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '任务进度加载失败'))
  }
}
const openAppendProcessing = () => {
  if (!canAppendProcessing.value) return
  appendForm.value = {
    action: '',
    violationType: detail.value.violationType || '',
    severity: 'normal',
    reason: '',
    confirmText: ''
  }
  appendPreview.value = null
  appendVisible.value = true
}

const audits = ref([])
const auditLoading = ref(false)
const auditDetailVisible = ref(false)
const auditDetailLoading = ref(false)
const auditDetailError = ref('')
const auditDetail = ref(null)
const loadAudits = async () => {
  if (!detail.value) return
  auditLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodAuditLogList({
        page: 1,
        pageSize: 100,
        targetType: detail.value.targetType,
        targetId: detail.value.targetId,
        sortOrder: 'desc'
      }),
      '审计记录加载失败'
    )
    audits.value = data?.list || []
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '审计记录加载失败'))
  } finally {
    auditLoading.value = false
  }
}

const openAuditDetail = async (row) => {
  auditDetailVisible.value = true
  auditDetailLoading.value = true
  auditDetailError.value = ''
  auditDetail.value = null
  try {
    auditDetail.value = unwrapOrderFoodResponse(
      await getOrderFoodAuditLogDetail(row.id),
      '审计记录详情加载失败'
    )
  } catch (error) {
    auditDetailError.value = getOrderFoodErrorMessage(
      error,
      '审计记录详情加载失败'
    )
  } finally {
    auditDetailLoading.value = false
  }
}

const retryVisible = ref(false)
const retryLoading = ref(false)
const retryReason = ref('')
const openRetry = () => {
  retryReason.value = ''
  retryVisible.value = true
}
const submitRetry = async () => {
  if (retryReason.value.trim().length < 4) {
    ElMessage.warning('重试原因至少 4 个字符')
    return
  }
  retryLoading.value = true
  try {
    detail.value.job = unwrapOrderFoodResponse(
      await retryOrderFoodGovernanceJob(activeJob.value.id, {
        reason: retryReason.value,
        expectedVersion: activeJob.value.version
      }),
      '任务重试失败'
    )
    retryVisible.value = false
    ElMessage.success('失败项重试已提交')
    await loadList()
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '任务重试失败'))
  } finally {
    retryLoading.value = false
  }
}

const appendActionMatrix = {
  dish: [
    { value: 'disable_discoverability', label: '关闭允许被发现' },
    { value: 'soft_delete_dish', label: '软删除用户菜品' },
    { value: 'delete_copy_chain', label: '严重违规：处理全部复制链', cascade: true }
  ],
  official_dish: [
    { value: 'soft_delete_official_dish', label: '软删除官方菜品' },
    { value: 'delete_copy_chain', label: '严重违规：处理全部复制链', cascade: true }
  ],
  recipe: [{ value: 'soft_delete_recipe', label: '软删除菜谱' }],
  checkin: [{ value: 'soft_delete_checkin', label: '软删除打卡' }]
}
const appendActionOptions = computed(() =>
  (appendActionMatrix[detail.value?.targetType] || []).filter(
    (option) => !option.cascade || canCascade.value
  )
)
const createDefaultAppendForm = () => ({
  action: '',
  violationType: '',
  severity: 'normal',
  reason: '',
  confirmText: ''
})
const appendVisible = ref(false)
const appendLoading = ref(false)
const appendFormRef = ref(null)
const appendForm = ref(createDefaultAppendForm())
const appendPreview = ref(null)
const appendRules = {
  action: [{ required: true, message: '请选择处理动作', trigger: 'change' }],
  violationType: [
    { required: true, message: '请输入违规类型', trigger: 'blur' },
    { max: 40, message: '违规类型最多 40 字', trigger: 'blur' }
  ],
  severity: [{ required: true, message: '请选择严重程度', trigger: 'change' }],
  reason: [
    { required: true, message: '请输入处理原因', trigger: 'blur' },
    { min: 4, max: 300, message: '原因长度应为 4-300 字', trigger: 'blur' }
  ]
}
const appendIsCascade = computed(
  () => appendForm.value.action === 'delete_copy_chain'
)
const appendConfirmValid = computed(
  () => appendForm.value.confirmText.trim() === '确认处理复制链'
)
const clearAppendPreview = () => {
  appendPreview.value = null
}
const handleAppendActionChange = () => {
  if (appendIsCascade.value) {
    appendForm.value.severity = 'serious'
  } else {
    appendForm.value.confirmText = ''
  }
  clearAppendPreview()
}
const resetAppendProcessing = () => {
  appendForm.value = createDefaultAppendForm()
  appendPreview.value = null
  appendFormRef.value?.clearValidate()
}
const appendPayload = () => ({
  targetType: detail.value.targetType,
  targetId: detail.value.targetId,
  actions: [appendForm.value.action],
  violationType: appendForm.value.violationType.trim(),
  severity: appendForm.value.severity,
  reason: appendForm.value.reason.trim(),
  expectedVersion: detail.value.targetVersion
})
const previewAppendProcessing = async () => {
  const valid = await appendFormRef.value?.validate().catch(() => false)
  if (!valid) return
  appendLoading.value = true
  try {
    appendPreview.value = unwrapOrderFoodResponse(
      await previewOrderFoodGovernanceAction(appendPayload()),
      '影响预览失败'
    )
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '影响预览失败'))
  } finally {
    appendLoading.value = false
  }
}
const executeAppendProcessing = async () => {
  if (!appendPreview.value) return
  if (appendIsCascade.value && !appendConfirmValid.value) {
    ElMessage.warning('请输入完整确认词“确认处理复制链”')
    return
  }
  appendLoading.value = true
  try {
    const result = unwrapOrderFoodResponse(
      await executeOrderFoodGovernanceAction({
        ...appendPayload(),
        previewToken: appendPreview.value.previewToken,
        confirmText: appendIsCascade.value ? appendForm.value.confirmText : ''
      }),
      '追加处理失败'
    )
    appendVisible.value = false
    detailVisible.value = false
    ElMessage.success(result?.jobId ? '追加处理已提交异步任务' : '追加处理已完成')
    await loadList()
    if (result?.recordId) {
      await openDetail(result.recordId)
    }
  } catch (error) {
    appendPreview.value = null
    ElMessage.error(getOrderFoodErrorMessage(error, '追加处理失败'))
  } finally {
    appendLoading.value = false
  }
}

loadList()
if (route.query.recordId) {
  openDetail(String(route.query.recordId))
}
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
