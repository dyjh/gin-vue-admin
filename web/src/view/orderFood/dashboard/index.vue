<template>
  <div>
    <div class="gva-search-box dashboard-toolbar">
      <div>
        <div class="text-lg font-medium">数据概览</div>
        <div class="mt-1 text-sm text-gray-500">
          查看核心业务运行情况；金额是运营估算，不作为财务结算依据。
        </div>
      </div>
      <div class="toolbar-actions">
        <el-radio-group v-model="range" @change="loadDashboard">
          <el-radio-button value="today">今日</el-radio-button>
          <el-radio-button value="last7Days">近 7 天</el-radio-button>
          <el-radio-button value="last30Days">近 30 天</el-radio-button>
        </el-radio-group>
        <el-button :loading="loading" @click="loadDashboard">刷新</el-button>
      </div>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      description="页面保留上次成功获取的数据，可稍后重新刷新。"
      type="error"
      show-icon
      class="mb-4"
    />

    <div v-loading="loading && !dashboard" class="dashboard-content">
      <div v-if="dashboard" class="metric-grid">
        <button
          v-for="metric in dashboard.metrics || []"
          :key="metric.key"
          type="button"
          class="metric-card"
          :class="{ clickable: canOpenTarget(metric.targetPage) }"
          :disabled="!canOpenTarget(metric.targetPage)"
          @click="openTarget(metric.targetPage, metric.targetQuery)"
        >
          <span class="metric-label">{{ metric.label }}</span>
          <span class="metric-value">
            {{ metricValue(metric) }}
            <small v-if="metric.unit">{{ metric.unit }}</small>
          </span>
          <span v-if="canOpenTarget(metric.targetPage)" class="metric-link">查看明细</span>
        </button>
      </div>

      <div v-if="dashboard" class="dashboard-panels">
        <div class="gva-table-box panel">
          <div class="panel-title">平台整体能力</div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="平台总开关">
              <el-tag :type="dashboard.capabilityState?.platformDefaultEnabled ? 'success' : 'info'">
                {{ dashboard.capabilityState?.platformDefaultEnabled ? '已开启' : '已关闭' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="紧急停用">
              <el-tag :type="dashboard.capabilityState?.emergencyDisabled ? 'danger' : 'success'">
                {{ dashboard.capabilityState?.emergencyDisabled ? '已停用' : '未停用' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="实际生效用户">
              {{ dashboard.capabilityState?.effectiveEnabledUserCount || 0 }} /
              {{ dashboard.capabilityState?.normalUserCount || 0 }}
            </el-descriptions-item>
            <el-descriptions-item label="策略版本">
              v{{ dashboard.capabilityState?.policyVersion || 0 }}
            </el-descriptions-item>
          </el-descriptions>
          <el-button
            v-if="canOpenTarget('OrderFoodPlatformPolicy')"
            class="mt-4"
            type="primary"
            plain
            @click="openTarget('OrderFoodPlatformPolicy', {})"
          >
            查看平台策略
          </el-button>
        </div>

        <div class="gva-table-box panel">
          <div class="panel-title">需要处理</div>
          <el-empty
            v-if="!(dashboard.alerts || []).length"
            description="当前没有异常提醒"
            :image-size="72"
          />
          <div v-else class="alert-list">
            <button
              v-for="alert in dashboard.alerts"
              :key="alert.type"
              type="button"
              class="alert-item"
              :class="{ clickable: canOpenTarget(alert.targetPage) }"
              :disabled="!canOpenTarget(alert.targetPage)"
              @click="openTarget(alert.targetPage, alert.targetQuery)"
            >
              <span>
                <el-tag :type="alert.level === 'critical' ? 'danger' : 'warning'">
                  {{ alertLabel(alert.type) }}
                </el-tag>
              </span>
              <strong>{{ alert.count }}</strong>
              <span v-if="canOpenTarget(alert.targetPage)" class="alert-action">去处理</span>
            </button>
          </div>
        </div>
      </div>

      <el-empty v-if="!dashboard && !loading" description="暂无概览数据" />
      <div v-if="dashboard" class="generated-at">
        统计时区 {{ dashboard.timezone }} · 最近更新 {{ formatDateTime(dashboard.generatedAt) }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import { getOrderFoodDashboard } from '@/api/orderfood/operations'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodDashboard' })

const router = useRouter()
const btnAuth = useBtnAuth()
const targetPermissions = {
  OrderFoodUsers: 'orderfood:user:read',
  OrderFoodUserDishes: 'orderfood:user-dish:read',
  OrderFoodDiscoverableDishes: 'orderfood:discoverable-dish:read',
  OrderFoodRecommendations: 'orderfood:recommendation:read',
  OrderFoodMeals: 'orderfood:meal:read',
  OrderFoodShoppingLists: 'orderfood:shopping:read',
  OrderFoodMedia: 'orderfood:media:read',
  OrderFoodPointEntries: 'orderfood:points:read',
  OrderFoodAIUsages: 'orderfood:ai-usage:read',
  OrderFoodPlatformPolicy: 'orderfood:platform-policy:read',
  OrderFoodGovernanceRecords: 'orderfood:governance:read',
  OrderFoodModerationRecords: 'orderfood:moderation:read',
  OrderFoodSubscribeLogs: 'orderfood:subscribe-log:read'
}
const canOpenTarget = (name) =>
  Boolean(name && targetPermissions[name] && btnAuth[targetPermissions[name]])
const range = ref('today')
const dashboard = ref(null)
const loading = ref(false)
const errorMessage = ref('')
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const alertLabels = {
  governance_job_abnormal: '违规处理任务异常',
  moderation_failed: '图片审核未通过',
  subscribe_failed: '订阅消息发送失败',
  refund_pending: '待退积分',
  ai_provider_consecutive_failure: 'AI 供应商连续失败',
  ai_model_consecutive_failure: 'AI 模型连续失败'
}
const alertLabel = (value) => alertLabels[value] || value
const metricValue = (metric) => {
  if (metric.value === null || typeof metric.value === 'undefined') return '—'
  if (typeof metric.value === 'boolean') return metric.value ? '开启' : '关闭'
  return metric.value
}
const openTarget = (name, query = {}) => {
  if (!canOpenTarget(name)) return
  router.push({ name, query: query || {} })
}
const loadDashboard = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    dashboard.value = unwrapOrderFoodResponse(
      await getOrderFoodDashboard({ range: range.value }),
      '运营概览加载失败'
    )
  } catch (error) {
    errorMessage.value = getOrderFoodErrorMessage(error, '运营概览加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboard)
</script>

<style scoped>
.dashboard-toolbar,
.toolbar-actions,
.metric-grid,
.dashboard-panels,
.alert-item {
  display: flex;
}

.dashboard-toolbar {
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.toolbar-actions {
  align-items: center;
  gap: 12px;
}

.metric-grid {
  flex-wrap: wrap;
  gap: 14px;
}

.metric-card {
  position: relative;
  width: calc(25% - 11px);
  min-height: 126px;
  padding: 18px;
  color: inherit;
  text-align: left;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
}

.metric-card:disabled {
  cursor: default;
  opacity: 1;
}

.metric-card.clickable {
  cursor: pointer;
}

.metric-card.clickable:hover {
  border-color: var(--el-color-primary-light-5);
  box-shadow: var(--el-box-shadow-light);
}

.metric-label,
.metric-value,
.metric-link {
  display: block;
}

.metric-label {
  color: var(--el-text-color-secondary);
  font-size: 14px;
}

.metric-value {
  margin-top: 14px;
  font-size: 28px;
  font-weight: 600;
}

.metric-value small {
  margin-left: 4px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  font-weight: 400;
}

.metric-link {
  margin-top: 10px;
  color: var(--el-color-primary);
  font-size: 12px;
}

.dashboard-panels {
  gap: 16px;
  margin-top: 16px;
}

.panel {
  width: 50%;
}

.panel-title {
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
}

.alert-list {
  display: grid;
  gap: 10px;
}

.alert-item {
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px;
  color: inherit;
  background: var(--el-fill-color-light);
  border: 0;
  border-radius: 8px;
}

.alert-item.clickable {
  cursor: pointer;
}

.alert-item strong {
  margin-left: auto;
  font-size: 20px;
}

.alert-action {
  color: var(--el-color-primary);
}

.generated-at {
  margin-top: 12px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  text-align: right;
}

@media (max-width: 1100px) {
  .metric-card {
    width: calc(50% - 7px);
  }
}
</style>
