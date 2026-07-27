<template>
  <div class="dashboard-page">
    <section class="dashboard-hero">
      <div class="hero-copy">
        <div class="hero-eyebrow">
          <el-icon><DataAnalysis /></el-icon>
          <span>OPERATIONS OVERVIEW</span>
        </div>
        <h1>运营工作台</h1>
        <p>聚焦用户增长、内容供给与关键服务状态，快速定位需要关注的业务变化。</p>
      </div>

      <div class="hero-toolbar">
        <div class="live-status">
          <span class="live-status-dot" />
          <span>{{ rangeLabel }}</span>
        </div>
        <el-radio-group v-model="range" size="large" @change="loadDashboard">
          <el-radio-button value="today">今日</el-radio-button>
          <el-radio-button value="last7Days">近 7 天</el-radio-button>
          <el-radio-button value="last30Days">近 30 天</el-radio-button>
        </el-radio-group>
        <el-button
          class="hero-refresh"
          :icon="Refresh"
          :loading="loading"
          size="large"
          @click="loadDashboard"
        >
          刷新
        </el-button>
      </div>
    </section>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      description="页面保留上次成功获取的数据，可稍后重新刷新。"
      type="error"
      show-icon
      class="dashboard-error"
    />

    <div v-loading="loading && !dashboard" class="dashboard-body">
      <section v-if="dashboard" class="summary-grid" aria-label="核心指标">
        <button
          v-for="item in summaryMetrics"
          :key="item.key"
          type="button"
          class="summary-card"
          :class="[`tone-${item.tone}`, { clickable: canOpenTarget(item.metric.targetPage) }]"
          :disabled="!canOpenTarget(item.metric.targetPage)"
          @click="openTarget(item.metric.targetPage, item.metric.targetQuery)"
        >
          <span class="summary-card-top">
            <span class="summary-icon">
              <el-icon><component :is="item.icon" /></el-icon>
            </span>
            <el-icon v-if="canOpenTarget(item.metric.targetPage)" class="summary-arrow">
              <Right />
            </el-icon>
          </span>
          <span class="summary-label">{{ item.metric.label }}</span>
          <span class="summary-value">
            {{ metricValue(item.metric) }}
            <small v-if="item.metric.unit">{{ item.metric.unit }}</small>
          </span>
          <span class="summary-description">{{ item.description }}</span>
        </button>
      </section>

      <div v-if="dashboard" class="content-layout">
        <main class="insight-grid">
          <section
            v-for="section in detailSections"
            :key="section.key"
            class="insight-panel"
            :class="{ wide: section.wide }"
          >
            <header class="panel-header">
              <span class="panel-icon" :class="`tone-${section.tone}`">
                <el-icon><component :is="section.icon" /></el-icon>
              </span>
              <span>
                <strong>{{ section.title }}</strong>
                <small>{{ section.description }}</small>
              </span>
            </header>

            <div v-if="section.key === 'business'" class="business-chart-layout">
              <button
                type="button"
                class="conversion-gauge"
                :class="{ clickable: canOpenTarget(conversionMetric.targetPage) }"
                :disabled="!canOpenTarget(conversionMetric.targetPage)"
                @click="openTarget(conversionMetric.targetPage, conversionMetric.targetQuery)"
              >
                <span class="gauge-ring" :style="{ '--gauge-progress': `${conversionRate * 3.6}deg` }">
                  <span class="gauge-center">
                    <strong>{{ metricValue(conversionMetric) }}</strong>
                    <small>%</small>
                  </span>
                </span>
                <span class="gauge-label">{{ conversionMetric.label }}</span>
                <span class="gauge-hint">新增用户转化</span>
              </button>

              <div class="chart-canvas business-stage-chart">
                <VChart :option="businessChartOption" autoresize />
              </div>
            </div>

            <div v-else-if="section.key === 'content'" class="content-chart">
              <button
                v-for="item in contentChartItems"
                :key="item.metric.key"
                type="button"
                class="content-bar-item"
                :class="{ clickable: canOpenTarget(item.metric.targetPage) }"
                :disabled="!canOpenTarget(item.metric.targetPage)"
                @click="openTarget(item.metric.targetPage, item.metric.targetQuery)"
              >
                <span class="content-bar-heading">
                  <span>{{ item.metric.label }}</span>
                  <strong>
                    {{ metricValue(item.metric) }}
                    <small>{{ item.metric.unit }}</small>
                  </strong>
                </span>
                <span class="content-bar-track">
                  <span :class="`tone-${item.tone}`" :style="{ width: `${item.percent}%` }" />
                </span>
              </button>
            </div>

            <div v-else-if="section.key === 'points'" class="chart-canvas points-chart">
              <VChart :option="pointsChartOption" autoresize />
            </div>

            <div v-else-if="section.key === 'ai'" class="ai-chart-layout">
              <div class="chart-canvas ai-donut-chart">
                <VChart :option="aiChartOption" autoresize />
              </div>

              <div class="ai-metric-board">
                <button
                  v-for="(metric, index) in aiRateMetrics"
                  :key="metric.key"
                  type="button"
                  class="ai-rate-row"
                  :class="{ clickable: canOpenTarget(metric.targetPage) }"
                  :disabled="!canOpenTarget(metric.targetPage)"
                  @click="openTarget(metric.targetPage, metric.targetQuery)"
                >
                  <i :class="index === 0 ? 'success' : 'danger'" />
                  <span>{{ metric.label }}</span>
                  <strong>{{ metricValue(metric) }}<small>%</small></strong>
                </button>

                <div class="ai-stat-grid">
                  <button
                    v-for="metric in aiStatMetrics"
                    :key="metric.key"
                    type="button"
                    class="ai-stat-card"
                    :class="{ clickable: canOpenTarget(metric.targetPage) }"
                    :disabled="!canOpenTarget(metric.targetPage)"
                    @click="openTarget(metric.targetPage, metric.targetQuery)"
                  >
                    <span>{{ metric.label }}</span>
                    <strong>
                      {{ metricValue(metric) }}
                      <small>{{ metric.unit }}</small>
                    </strong>
                  </button>
                </div>
              </div>
            </div>
          </section>
        </main>

        <aside class="status-column">
          <section class="status-panel task-panel">
            <header class="status-panel-header">
              <div>
                <span class="status-kicker">ACTION CENTER</span>
                <h2>需要处理</h2>
              </div>
              <span class="task-count" :class="{ empty: !alertCount }">{{ alertCount }}</span>
            </header>

            <div v-if="!(dashboard.alerts || []).length" class="all-clear">
              <span class="all-clear-icon">
                <el-icon><CircleCheckFilled /></el-icon>
              </span>
              <div>
                <strong>运行正常</strong>
                <p>当前没有需要人工介入的异常</p>
              </div>
            </div>

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
                <span class="alert-icon" :class="{ critical: alert.level === 'critical' }">
                  <el-icon><WarningFilled /></el-icon>
                </span>
                <span class="alert-copy">
                  <strong>{{ alertLabel(alert.type) }}</strong>
                  <small>{{ alert.level === 'critical' ? '建议立即处理' : '请及时关注' }}</small>
                </span>
                <span class="alert-number">{{ alert.count }}</span>
                <el-icon v-if="canOpenTarget(alert.targetPage)" class="alert-arrow">
                  <Right />
                </el-icon>
              </button>
            </div>
          </section>

          <section class="status-panel capability-panel">
            <header class="capability-header">
              <div>
                <span class="status-kicker">PLATFORM STATUS</span>
                <h2>平台服务状态</h2>
              </div>
              <span class="health-badge" :class="`tone-${platformHealth.tone}`">
                <i />
                {{ platformHealth.label }}
              </span>
            </header>

            <p class="health-hint">{{ platformHealth.hint }}</p>

            <div class="capability-stats">
              <div>
                <span>平台总开关</span>
                <strong>{{ dashboard.capabilityState?.platformDefaultEnabled ? '开启' : '关闭' }}</strong>
              </div>
              <div>
                <span>紧急停用</span>
                <strong :class="{ danger: dashboard.capabilityState?.emergencyDisabled }">
                  {{ dashboard.capabilityState?.emergencyDisabled ? '已启用' : '未启用' }}
                </strong>
              </div>
            </div>

            <div class="coverage-block">
              <div class="coverage-heading">
                <span>能力覆盖用户</span>
                <strong>
                  {{ dashboard.capabilityState?.effectiveEnabledUserCount || 0 }}
                  <small>/ {{ dashboard.capabilityState?.normalUserCount || 0 }} 人</small>
                </strong>
              </div>
              <div class="coverage-track">
                <span :style="{ width: `${effectiveRatio}%` }" />
              </div>
              <div class="coverage-foot">
                <span>覆盖率 {{ effectiveRatio }}%</span>
                <span>策略版本 v{{ dashboard.capabilityState?.policyVersion || 0 }}</span>
              </div>
            </div>

            <button
              v-if="canOpenTarget('OrderFoodPlatformPolicy')"
              type="button"
              class="policy-link"
              @click="openTarget('OrderFoodPlatformPolicy', {})"
            >
              <span>查看平台策略</span>
              <el-icon><Right /></el-icon>
            </button>
          </section>
        </aside>
      </div>

      <div v-if="!dashboard && !loading" class="empty-dashboard">
        <el-empty description="暂无概览数据" />
      </div>
      <div v-if="dashboard" class="dashboard-footer">
        <span>数据按 {{ dashboard.timezone }} 时区统计</span>
        <span>最近更新 {{ formatDateTime(dashboard.generatedAt) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  CircleCheckFilled,
  Coin,
  DataAnalysis,
  Dish,
  MagicStick,
  Refresh,
  Right,
  TrendCharts,
  User,
  WarningFilled
} from '@element-plus/icons-vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart, PieChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import { GridComponent, TitleComponent, TooltipComponent } from 'echarts/components'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import { getOrderFoodDashboard } from '@/api/orderfood/operations'
import {
  getOrderFoodErrorMessage,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

use([BarChart, PieChart, CanvasRenderer, GridComponent, TitleComponent, TooltipComponent])

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
const rangeLabels = {
  today: '今日实时数据',
  last7Days: '近 7 天数据',
  last30Days: '近 30 天数据'
}
const rangeLabel = computed(() => rangeLabels[range.value] || '运营数据')
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

const summaryMetricConfigs = [
  {
    key: 'new_users',
    icon: User,
    tone: 'blue',
    description: '所选周期内完成注册'
  },
  {
    key: 'active_users',
    icon: TrendCharts,
    tone: 'teal',
    description: '产生过关键业务行为'
  },
  {
    key: 'meals_created',
    icon: Dish,
    tone: 'amber',
    description: '所选周期新建饭局'
  },
  {
    key: 'usable_dishes',
    icon: CircleCheckFilled,
    tone: 'cyan',
    description: '当前可正常使用的菜品'
  }
]

const detailSectionConfigs = [
  {
    key: 'business',
    title: '业务转化',
    description: '从新增用户到饭局履约的核心链路',
    icon: TrendCharts,
    tone: 'blue',
    wide: true,
    metricKeys: [
      'first_dish_creation_rate',
      'meal_participants_joined',
      'meals_confirmed',
      'shopping_lists_created',
      'checkins'
    ]
  },
  {
    key: 'content',
    title: '内容供给',
    description: '菜品发现与平台推荐',
    icon: Dish,
    tone: 'teal',
    metricKeys: ['discoverable_dishes', 'recommended_dishes']
  },
  {
    key: 'points',
    title: '积分流转',
    description: '周期内积分发放与消耗',
    icon: Coin,
    tone: 'amber',
    metricKeys: ['points_earned', 'points_spent', 'points_refunded']
  },
  {
    key: 'ai',
    title: 'AI 服务',
    description: '调用质量、耗时与成本估算',
    icon: MagicStick,
    tone: 'cyan',
    wide: true,
    metricKeys: [
      'ai_success_rate',
      'ai_failure_rate',
      'ai_average_duration_ms',
      'ai_estimated_cost_cny'
    ]
  }
]

const metricMap = computed(() =>
  Object.fromEntries((dashboard.value?.metrics || []).map((metric) => [metric.key, metric]))
)
const summaryMetrics = computed(() =>
  summaryMetricConfigs
    .map((config) => ({ ...config, metric: metricMap.value[config.key] }))
    .filter((item) => item.metric)
)
const detailSections = computed(() =>
  detailSectionConfigs.map((section) => ({
    ...section,
    metrics: section.metricKeys.map((key) => metricMap.value[key]).filter(Boolean)
  }))
)
const metricNumber = (metric) => {
  const value = Number(metric?.value)
  return Number.isFinite(value) ? value : 0
}
const conversionMetric = computed(
  () =>
    metricMap.value.first_dish_creation_rate || {
      key: 'first_dish_creation_rate',
      label: '新用户首次创建菜品率',
      value: null,
      unit: '%'
    }
)
const conversionRate = computed(() =>
  Math.max(0, Math.min(100, metricNumber(conversionMetric.value)))
)
const businessStageMetrics = computed(() =>
  [
    'meal_participants_joined',
    'meals_confirmed',
    'shopping_lists_created',
    'checkins'
  ]
    .map((key) => metricMap.value[key])
    .filter(Boolean)
)
const businessChartOption = computed(() => {
  const shortLabels = {
    meal_participants_joined: '加入饭局',
    meals_confirmed: '确认菜单',
    shopping_lists_created: '生成采购单',
    checkins: '完成打卡'
  }
  const metrics = [...businessStageMetrics.value].reverse()
  return {
    animationDuration: 500,
    grid: { left: 76, right: 48, top: 8, bottom: 8 },
    tooltip: {
      trigger: 'item',
      formatter: ({ name, value }) => `${name}<br/><strong>${value}</strong> 次`
    },
    xAxis: {
      type: 'value',
      show: false,
      max: ({ max }) => Math.max(max, 1)
    },
    yAxis: {
      type: 'category',
      data: metrics.map((metric) => shortLabels[metric.key] || metric.label),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#8492a6', fontSize: 11 }
    },
    series: [
      {
        type: 'bar',
        data: metrics.map((metric) => metricNumber(metric)),
        barWidth: 12,
        showBackground: true,
        backgroundStyle: {
          color: 'rgba(132, 146, 166, 0.10)',
          borderRadius: 8
        },
        itemStyle: {
          color: '#2f6fec',
          borderRadius: 8
        },
        label: {
          show: true,
          position: 'right',
          color: '#64748b',
          fontSize: 11,
          formatter: '{c} 次'
        }
      }
    ]
  }
})
const contentChartItems = computed(() => {
  const tones = ['blue', 'teal', 'amber']
  const metrics = ['usable_dishes', 'discoverable_dishes', 'recommended_dishes']
    .map((key) => metricMap.value[key])
    .filter(Boolean)
  const maximum = Math.max(...metrics.map(metricNumber), 1)
  return metrics.map((metric, index) => ({
    metric,
    tone: tones[index],
    percent: metricNumber(metric) ? Math.max(3, (metricNumber(metric) / maximum) * 100) : 0
  }))
})
const pointChartMetrics = computed(() =>
  ['points_earned', 'points_spent', 'points_refunded']
    .map((key) => metricMap.value[key])
    .filter(Boolean)
)
const pointsChartOption = computed(() => {
  const colors = ['#d98521', '#2f6fec', '#159184']
  return {
    animationDuration: 500,
    grid: { left: 8, right: 8, top: 28, bottom: 26, containLabel: true },
    tooltip: {
      trigger: 'item',
      formatter: ({ name, value }) => `${name}<br/><strong>${value}</strong> 积分`
    },
    xAxis: {
      type: 'category',
      data: pointChartMetrics.value.map((metric) => metric.label.replace('积分', '')),
      axisLine: { lineStyle: { color: 'rgba(132, 146, 166, 0.18)' } },
      axisTick: { show: false },
      axisLabel: { color: '#8492a6', fontSize: 11 }
    },
    yAxis: {
      type: 'value',
      show: false,
      max: ({ max }) => Math.max(max, 1)
    },
    series: [
      {
        type: 'bar',
        barMaxWidth: 34,
        showBackground: true,
        backgroundStyle: {
          color: 'rgba(132, 146, 166, 0.08)',
          borderRadius: [7, 7, 0, 0]
        },
        data: pointChartMetrics.value.map((metric, index) => ({
          value: metricNumber(metric),
          itemStyle: {
            color: colors[index],
            borderRadius: [7, 7, 0, 0]
          }
        })),
        label: {
          show: true,
          position: 'top',
          color: '#64748b',
          fontSize: 11
        }
      }
    ]
  }
})
const aiRateMetrics = computed(() =>
  ['ai_success_rate', 'ai_failure_rate'].map((key) => metricMap.value[key]).filter(Boolean)
)
const aiStatMetrics = computed(() =>
  ['ai_average_duration_ms', 'ai_estimated_cost_cny']
    .map((key) => metricMap.value[key])
    .filter(Boolean)
)
const aiChartOption = computed(() => {
  const success = metricNumber(metricMap.value.ai_success_rate)
  const failure = metricNumber(metricMap.value.ai_failure_rate)
  const hasCalls = success + failure > 0
  return {
    animationDuration: 550,
    title: {
      text: hasCalls ? `${success}%` : '—',
      subtext: hasCalls ? '调用成功率' : '暂无调用',
      left: 'center',
      top: '34%',
      textStyle: { color: '#64748b', fontSize: 23, fontWeight: 700 },
      subtextStyle: { color: '#8492a6', fontSize: 10, lineHeight: 20 }
    },
    tooltip: {
      show: hasCalls,
      trigger: 'item',
      formatter: '{b}<br/><strong>{c}%</strong>'
    },
    series: [
      {
        type: 'pie',
        radius: ['64%', '82%'],
        center: ['50%', '50%'],
        silent: !hasCalls,
        label: { show: false },
        emphasis: { scaleSize: 4 },
        data: hasCalls
          ? [
              { value: success, name: '成功', itemStyle: { color: '#159184' } },
              { value: failure, name: '失败', itemStyle: { color: '#d9594c' } }
            ]
          : [{ value: 1, name: '暂无调用', itemStyle: { color: '#e5e9ef' } }]
      }
    ]
  }
})
const alertCount = computed(() =>
  (dashboard.value?.alerts || []).reduce((total, alert) => total + Number(alert.count || 0), 0)
)
const effectiveRatio = computed(() => {
  const state = dashboard.value?.capabilityState
  if (!state?.normalUserCount) return 0
  return Math.min(
    100,
    Math.round((Number(state.effectiveEnabledUserCount || 0) / Number(state.normalUserCount)) * 100)
  )
})
const platformHealth = computed(() => {
  const state = dashboard.value?.capabilityState
  if (state?.emergencyDisabled) {
    return {
      label: '紧急停用',
      hint: '平台能力已被紧急停用，请确认当前策略是否符合预期。',
      tone: 'danger'
    }
  }
  if (state?.platformDefaultEnabled) {
    return {
      label: '服务正常',
      hint: '平台默认能力已开启，当前未触发紧急停用。',
      tone: 'success'
    }
  }
  return {
    label: '能力关闭',
    hint: '平台默认能力当前处于关闭状态，仅展示历史运营数据。',
    tone: 'muted'
  }
})

const metricValue = (metric) => {
  if (metric.value === null || typeof metric.value === 'undefined') return '—'
  if (typeof metric.value === 'boolean') return metric.value ? '开启' : '关闭'
  const numericValue =
    typeof metric.value === 'string' && metric.value.trim() !== ''
      ? Number(metric.value)
      : metric.value
  if (typeof numericValue === 'number' && Number.isFinite(numericValue)) {
    return numericValue.toLocaleString('zh-CN', {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    })
  }
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
.dashboard-page {
  --dashboard-ink: var(--el-text-color-primary);
  --dashboard-muted: var(--el-text-color-secondary);
  --dashboard-panel: var(--el-bg-color-overlay);
  --dashboard-line: var(--el-border-color-lighter);
  min-height: 100%;
  padding: 14px 16px 24px;
  color: var(--dashboard-ink);
  font-family: "HarmonyOS Sans SC", "Microsoft YaHei", sans-serif;
}

.dashboard-hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 32px;
  min-height: 154px;
  padding: 28px 30px;
  overflow: hidden;
  color: #fff;
  background:
    radial-gradient(circle at 78% -70%, rgb(87 193 198 / 45%) 0, transparent 52%),
    linear-gradient(118deg, #102f4f 0%, #174d63 62%, #17656b 100%);
  border: 1px solid rgb(255 255 255 / 10%);
  border-radius: 18px;
  box-shadow: 0 14px 38px rgb(15 47 79 / 16%);
}

.dashboard-hero::after {
  position: absolute;
  right: 23%;
  bottom: -72px;
  width: 220px;
  height: 220px;
  border: 1px solid rgb(255 255 255 / 9%);
  border-radius: 50%;
  box-shadow:
    0 0 0 34px rgb(255 255 255 / 3%),
    0 0 0 72px rgb(255 255 255 / 2%);
  content: "";
  pointer-events: none;
}

.hero-copy,
.hero-toolbar {
  position: relative;
  z-index: 1;
}

.hero-eyebrow {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #9fdfe0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
}

.hero-copy h1 {
  margin: 9px 0 7px;
  font-size: 29px;
  font-weight: 700;
  letter-spacing: 0.03em;
  line-height: 1.25;
}

.hero-copy p {
  margin: 0;
  color: rgb(255 255 255 / 68%);
  font-size: 14px;
}

.hero-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.live-status {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-right: 4px;
  color: rgb(255 255 255 / 72%);
  font-size: 12px;
  white-space: nowrap;
}

.live-status-dot {
  width: 7px;
  height: 7px;
  background: #5ee4b2;
  border-radius: 50%;
  box-shadow: 0 0 0 4px rgb(94 228 178 / 12%);
}

.hero-toolbar :deep(.el-radio-button__inner) {
  padding-right: 17px;
  padding-left: 17px;
  color: rgb(255 255 255 / 72%);
  background: rgb(255 255 255 / 8%);
  border-color: rgb(255 255 255 / 12%);
  box-shadow: none;
}

.hero-toolbar :deep(.el-radio-button:first-child .el-radio-button__inner) {
  border-left-color: rgb(255 255 255 / 12%);
}

.hero-toolbar :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  color: #123849;
  background: #fff;
  border-color: #fff;
  box-shadow: -1px 0 0 0 #fff;
}

.hero-refresh {
  color: #fff;
  background: rgb(255 255 255 / 8%);
  border-color: rgb(255 255 255 / 16%);
}

.hero-refresh:hover,
.hero-refresh:focus {
  color: #fff;
  background: rgb(255 255 255 / 16%);
  border-color: rgb(255 255 255 / 28%);
}

.dashboard-error {
  margin-top: 14px;
}

.dashboard-body {
  min-height: 360px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  margin-top: 14px;
}

.summary-card {
  --accent: #2f6fec;
  --accent-soft: rgb(47 111 236 / 10%);
  position: relative;
  display: flex;
  min-height: 174px;
  padding: 18px 19px;
  overflow: hidden;
  color: var(--dashboard-ink);
  text-align: left;
  background: var(--dashboard-panel);
  border: 1px solid var(--dashboard-line);
  border-radius: 15px;
  flex-direction: column;
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease,
    transform 180ms ease;
}

.summary-card::after {
  position: absolute;
  top: -34px;
  right: -34px;
  width: 108px;
  height: 108px;
  background: var(--accent-soft);
  border-radius: 50%;
  content: "";
  pointer-events: none;
}

.summary-card.tone-teal,
.panel-icon.tone-teal {
  --accent: #159184;
  --accent-soft: rgb(21 145 132 / 11%);
}

.summary-card.tone-amber,
.panel-icon.tone-amber {
  --accent: #d98521;
  --accent-soft: rgb(217 133 33 / 12%);
}

.summary-card.tone-cyan,
.panel-icon.tone-cyan {
  --accent: #1687a7;
  --accent-soft: rgb(22 135 167 / 11%);
}

.summary-card:disabled,
.compact-metric:disabled,
.alert-item:disabled {
  cursor: default;
  opacity: 1;
}

.summary-card.clickable,
.compact-metric.clickable,
.alert-item.clickable {
  cursor: pointer;
}

.summary-card.clickable:hover {
  border-color: color-mix(in srgb, var(--accent) 42%, var(--dashboard-line));
  box-shadow: 0 13px 28px rgb(31 52 75 / 10%);
  transform: translateY(-2px);
}

.summary-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 34px;
}

.summary-icon,
.panel-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  background: var(--accent-soft);
}

.summary-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  font-size: 17px;
}

.summary-arrow {
  position: relative;
  z-index: 1;
  color: var(--el-text-color-placeholder);
  font-size: 15px;
}

.summary-label {
  margin-top: 14px;
  color: var(--dashboard-muted);
  font-size: 13px;
}

.summary-value {
  margin-top: 3px;
  color: var(--dashboard-ink);
  font-size: 31px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1.2;
}

.summary-value small {
  margin-left: 5px;
  color: var(--dashboard-muted);
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0;
}

.summary-description {
  margin-top: auto;
  padding-top: 8px;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}

.content-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 350px;
  gap: 14px;
  margin-top: 14px;
}

.insight-grid {
  display: grid;
  align-content: start;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.insight-panel,
.status-panel {
  background: var(--dashboard-panel);
  border: 1px solid var(--dashboard-line);
  border-radius: 15px;
}

.insight-panel {
  min-width: 0;
  padding: 18px;
}

.insight-panel.wide {
  grid-column: 1 / -1;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 11px;
  margin-bottom: 15px;
}

.panel-icon {
  --accent: #2f6fec;
  --accent-soft: rgb(47 111 236 / 10%);
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  border-radius: 10px;
}

.panel-header > span:last-child {
  display: grid;
  gap: 2px;
}

.panel-header strong {
  font-size: 15px;
  font-weight: 650;
}

.panel-header small {
  color: var(--dashboard-muted);
  font-size: 11px;
}

.compact-metric-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  overflow: hidden;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--dashboard-line);
  border-radius: 11px;
}

.compact-metric-grid.count-3 {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.insight-panel.wide .compact-metric-grid {
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
}

.compact-metric {
  position: relative;
  min-width: 0;
  min-height: 87px;
  padding: 14px;
  color: inherit;
  text-align: left;
  background: transparent;
  border: 0;
  border-right: 1px solid var(--dashboard-line);
  transition: background 160ms ease;
}

.compact-metric:last-child {
  border-right: 0;
}

.compact-metric.clickable:hover {
  background: var(--el-fill-color);
}

.compact-label,
.compact-value {
  display: block;
}

.compact-label {
  overflow: hidden;
  color: var(--dashboard-muted);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.compact-value {
  margin-top: 7px;
  font-size: 22px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.compact-value small {
  margin-left: 4px;
  color: var(--dashboard-muted);
  font-size: 11px;
  font-weight: 400;
}

.compact-arrow {
  position: absolute;
  right: 10px;
  bottom: 11px;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}

.business-chart-layout {
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr);
  gap: 18px;
  min-height: 214px;
}

.conversion-gauge {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 14px;
  color: inherit;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--dashboard-line);
  border-radius: 12px;
  flex-direction: column;
  transition:
    border-color 160ms ease,
    background 160ms ease;
}

.conversion-gauge:disabled,
.content-bar-item:disabled,
.ai-rate-row:disabled,
.ai-stat-card:disabled {
  cursor: default;
  opacity: 1;
}

.conversion-gauge.clickable,
.content-bar-item.clickable,
.ai-rate-row.clickable,
.ai-stat-card.clickable {
  cursor: pointer;
}

.conversion-gauge.clickable:hover,
.content-bar-item.clickable:hover,
.ai-rate-row.clickable:hover,
.ai-stat-card.clickable:hover {
  background: var(--el-fill-color);
  border-color: var(--el-border-color);
}

.gauge-ring {
  --gauge-progress: 0deg;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 116px;
  height: 116px;
  background: conic-gradient(#2f6fec var(--gauge-progress), var(--el-fill-color-dark) 0);
  border-radius: 50%;
}

.gauge-ring::after {
  position: absolute;
  width: 86px;
  height: 86px;
  background: var(--dashboard-panel);
  border-radius: 50%;
  box-shadow: inset 0 0 0 1px var(--dashboard-line);
  content: "";
}

.gauge-center {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: baseline;
  color: var(--dashboard-ink);
}

.gauge-center strong {
  font-size: 25px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
}

.gauge-center small {
  margin-left: 2px;
  color: var(--dashboard-muted);
  font-size: 11px;
}

.gauge-label {
  margin-top: 11px;
  font-size: 11px;
  font-weight: 600;
}

.gauge-hint {
  margin-top: 3px;
  color: var(--el-text-color-placeholder);
  font-size: 9px;
}

.chart-canvas {
  min-width: 0;
}

.business-stage-chart {
  height: 214px;
}

.content-chart {
  display: grid;
  gap: 17px;
  padding: 9px 1px 8px;
}

.content-bar-item {
  display: grid;
  gap: 8px;
  padding: 0;
  color: inherit;
  text-align: left;
  background: transparent;
  border: 0;
}

.content-bar-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: var(--dashboard-muted);
  font-size: 11px;
}

.content-bar-heading strong {
  color: var(--dashboard-ink);
  font-size: 15px;
  font-variant-numeric: tabular-nums;
}

.content-bar-heading small {
  color: var(--dashboard-muted);
  font-size: 9px;
  font-weight: 400;
}

.content-bar-track {
  display: block;
  height: 9px;
  overflow: hidden;
  background: var(--el-fill-color-dark);
  border-radius: 99px;
}

.content-bar-track > span {
  display: block;
  height: 100%;
  min-width: 0;
  background: #2f6fec;
  border-radius: inherit;
  transition: width 420ms ease;
}

.content-bar-track > span.tone-teal {
  background: #159184;
}

.content-bar-track > span.tone-amber {
  background: #d98521;
}

.points-chart {
  height: 188px;
}

.ai-chart-layout {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 20px;
  min-height: 220px;
}

.ai-donut-chart {
  height: 220px;
}

.ai-metric-board {
  display: grid;
  align-content: center;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.ai-rate-row {
  display: grid;
  align-items: center;
  grid-template-columns: auto 1fr auto;
  gap: 9px;
  padding: 13px;
  color: inherit;
  text-align: left;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--dashboard-line);
  border-radius: 10px;
}

.ai-rate-row i {
  width: 8px;
  height: 8px;
  background: #159184;
  border-radius: 50%;
}

.ai-rate-row i.danger {
  background: #d9594c;
}

.ai-rate-row span {
  color: var(--dashboard-muted);
  font-size: 11px;
}

.ai-rate-row strong {
  font-size: 17px;
  font-variant-numeric: tabular-nums;
}

.ai-rate-row small {
  margin-left: 2px;
  color: var(--dashboard-muted);
  font-size: 9px;
}

.ai-stat-grid {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.ai-stat-card {
  display: grid;
  gap: 8px;
  padding: 15px;
  color: inherit;
  text-align: left;
  background: transparent;
  border: 1px solid var(--dashboard-line);
  border-radius: 10px;
}

.ai-stat-card > span {
  color: var(--dashboard-muted);
  font-size: 10px;
}

.ai-stat-card strong {
  font-size: 20px;
  font-variant-numeric: tabular-nums;
}

.ai-stat-card small {
  margin-left: 4px;
  color: var(--dashboard-muted);
  font-size: 9px;
  font-weight: 400;
}

.status-column {
  display: grid;
  align-content: start;
  gap: 14px;
}

.status-panel {
  padding: 18px;
}

.status-panel-header,
.capability-header,
.coverage-heading,
.coverage-foot,
.policy-link,
.alert-item,
.all-clear {
  display: flex;
  align-items: center;
}

.status-panel-header,
.capability-header {
  justify-content: space-between;
  gap: 12px;
}

.status-kicker {
  display: block;
  margin-bottom: 3px;
  color: var(--el-text-color-placeholder);
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.16em;
}

.status-panel h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 650;
}

.task-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 30px;
  height: 30px;
  padding: 0 8px;
  color: #b84238;
  background: rgb(216 74 62 / 10%);
  border-radius: 10px;
  font-size: 13px;
  font-weight: 700;
}

.task-count.empty {
  color: #16835b;
  background: rgb(22 131 91 / 10%);
}

.all-clear {
  gap: 12px;
  min-height: 84px;
  margin-top: 15px;
  padding: 14px;
  background: linear-gradient(135deg, rgb(28 156 108 / 9%), rgb(28 156 108 / 3%));
  border: 1px solid rgb(28 156 108 / 14%);
  border-radius: 11px;
}

.all-clear-icon {
  display: inline-flex;
  color: #1c9c6c;
  font-size: 29px;
}

.all-clear strong {
  font-size: 14px;
}

.all-clear p {
  margin: 4px 0 0;
  color: var(--dashboard-muted);
  font-size: 11px;
}

.alert-list {
  display: grid;
  gap: 8px;
  margin-top: 15px;
}

.alert-item {
  gap: 10px;
  width: 100%;
  padding: 11px;
  color: inherit;
  text-align: left;
  background: var(--el-fill-color-lighter);
  border: 1px solid transparent;
  border-radius: 10px;
}

.alert-item.clickable:hover {
  background: var(--el-fill-color);
  border-color: var(--dashboard-line);
}

.alert-icon {
  display: inline-flex;
  color: #d98a20;
  font-size: 17px;
}

.alert-icon.critical {
  color: #d84a3e;
}

.alert-copy {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.alert-copy strong {
  overflow: hidden;
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.alert-copy small {
  color: var(--dashboard-muted);
  font-size: 10px;
}

.alert-number {
  margin-left: auto;
  font-size: 17px;
  font-weight: 700;
}

.alert-arrow {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}

.capability-panel {
  color: #f3f8fb;
  background:
    radial-gradient(circle at 92% 0%, rgb(44 154 156 / 22%) 0, transparent 40%),
    #112c41;
  border-color: rgb(255 255 255 / 7%);
}

.capability-panel .status-kicker {
  color: rgb(197 224 230 / 52%);
}

.health-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border-radius: 8px;
  font-size: 10px;
  font-weight: 600;
  white-space: nowrap;
}

.health-badge i {
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 50%;
}

.health-badge.tone-success {
  color: #6ae2b0;
  background: rgb(106 226 176 / 10%);
}

.health-badge.tone-danger {
  color: #ff8b7c;
  background: rgb(255 139 124 / 10%);
}

.health-badge.tone-muted {
  color: #b6c6cf;
  background: rgb(182 198 207 / 10%);
}

.health-hint {
  margin: 13px 0 16px;
  color: rgb(222 236 240 / 62%);
  font-size: 11px;
  line-height: 1.65;
}

.capability-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  overflow: hidden;
  border: 1px solid rgb(255 255 255 / 8%);
  border-radius: 10px;
}

.capability-stats > div {
  display: grid;
  gap: 5px;
  padding: 12px;
}

.capability-stats > div + div {
  border-left: 1px solid rgb(255 255 255 / 8%);
}

.capability-stats span {
  color: rgb(222 236 240 / 52%);
  font-size: 10px;
}

.capability-stats strong {
  color: #f4fbfd;
  font-size: 14px;
}

.capability-stats strong.danger {
  color: #ff8b7c;
}

.coverage-block {
  margin-top: 16px;
}

.coverage-heading,
.coverage-foot {
  justify-content: space-between;
}

.coverage-heading > span {
  color: rgb(222 236 240 / 58%);
  font-size: 11px;
}

.coverage-heading strong {
  font-size: 17px;
  font-variant-numeric: tabular-nums;
}

.coverage-heading small {
  color: rgb(222 236 240 / 52%);
  font-size: 10px;
  font-weight: 400;
}

.coverage-track {
  height: 6px;
  margin: 10px 0 7px;
  overflow: hidden;
  background: rgb(255 255 255 / 8%);
  border-radius: 99px;
}

.coverage-track span {
  display: block;
  height: 100%;
  background: linear-gradient(90deg, #34bca6, #70ddbd);
  border-radius: inherit;
  transition: width 400ms ease;
}

.coverage-foot {
  color: rgb(222 236 240 / 43%);
  font-size: 9px;
}

.policy-link {
  justify-content: space-between;
  width: 100%;
  margin-top: 17px;
  padding: 10px 0 0;
  color: #a6e2df;
  background: transparent;
  border: 0;
  border-top: 1px solid rgb(255 255 255 / 8%);
  cursor: pointer;
  font-size: 11px;
}

.policy-link:hover {
  color: #fff;
}

.empty-dashboard {
  margin-top: 14px;
  background: var(--dashboard-panel);
  border: 1px solid var(--dashboard-line);
  border-radius: 15px;
}

.dashboard-footer {
  display: flex;
  justify-content: flex-end;
  gap: 18px;
  margin-top: 12px;
  color: var(--el-text-color-placeholder);
  font-size: 10px;
}

@media (max-width: 1280px) {
  .dashboard-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .hero-toolbar {
    width: 100%;
  }

  .live-status {
    margin-right: auto;
  }

  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .content-layout {
    grid-template-columns: minmax(0, 1fr) 320px;
  }
}

@media (max-width: 980px) {
  .content-layout {
    grid-template-columns: 1fr;
  }

  .status-column {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .dashboard-page {
    padding: 10px;
  }

  .dashboard-hero {
    min-height: auto;
    padding: 22px 18px;
    border-radius: 14px;
  }

  .hero-copy h1 {
    font-size: 24px;
  }

  .hero-toolbar {
    align-items: stretch;
    flex-wrap: wrap;
  }

  .live-status {
    width: 100%;
  }

  .hero-toolbar :deep(.el-radio-group) {
    display: flex;
    flex: 1;
  }

  .hero-toolbar :deep(.el-radio-button) {
    flex: 1;
  }

  .hero-toolbar :deep(.el-radio-button__inner) {
    width: 100%;
    padding-right: 10px;
    padding-left: 10px;
  }

  .summary-grid,
  .insight-grid,
  .status-column {
    grid-template-columns: 1fr;
  }

  .business-chart-layout,
  .ai-chart-layout {
    grid-template-columns: 1fr;
  }

  .conversion-gauge {
    min-height: 196px;
  }

  .ai-metric-board {
    grid-template-columns: 1fr;
  }

  .ai-stat-grid {
    grid-column: auto;
  }

  .summary-card {
    min-height: 158px;
  }

  .insight-panel.wide {
    grid-column: auto;
  }

  .insight-panel.wide .compact-metric-grid,
  .compact-metric-grid,
  .compact-metric-grid.count-3 {
    grid-template-columns: 1fr 1fr;
  }

  .compact-metric:nth-child(even) {
    border-right: 0;
  }

  .compact-metric:nth-child(n + 3) {
    border-top: 1px solid var(--dashboard-line);
  }

  .dashboard-footer {
    align-items: flex-end;
    flex-direction: column;
    gap: 2px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .summary-card,
  .coverage-track span {
    transition: none;
  }
}
</style>
