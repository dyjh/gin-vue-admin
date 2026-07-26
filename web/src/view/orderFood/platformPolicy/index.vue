<template>
  <div class="order-food-page" v-loading="loading">
    <div class="page-header">
      <div>
        <h2 class="page-title">平台整体能力策略</h2>
        <p class="page-subtitle">
          一个总开关统一控制全部 AI 功能，不提供单项开关或用户白名单。保存后立即生效。
        </p>
      </div>
      <div class="header-actions">
        <el-button
          v-if="canUpdate && !editing"
          type="primary"
          @click="beginEdit"
        >
          编辑当前配置
        </el-button>
        <el-button @click="refreshWorkspace">刷新</el-button>
      </div>
    </div>

    <el-alert
      v-if="pageError"
      :title="pageError"
      type="error"
      show-icon
      :closable="false"
      class="section-card"
    />

    <template v-if="workspace">
      <div class="metric-grid">
        <el-card shadow="never">
          <div class="metric-label">平台 AI 总开关</div>
          <el-tag :type="config.platformDefaultEnabled ? 'success' : 'info'" size="large">
            {{ config.platformDefaultEnabled ? '全部开放' : '全部关闭' }}
          </el-tag>
        </el-card>
        <el-card shadow="never">
          <div class="metric-label">紧急状态</div>
          <el-tag :type="config.emergencyDisabled ? 'danger' : 'success'" size="large">
            {{ config.emergencyDisabled ? '已紧急停用' : '正常' }}
          </el-tag>
        </el-card>
        <el-card shadow="never">
          <div class="metric-label">当前生效用户</div>
          <strong class="metric-value">
            {{ workspace.userCounts?.effectiveEnabledUserCount ?? 0 }}
          </strong>
          <span class="metric-note">
            正常用户 {{ workspace.userCounts?.normalUserCount ?? 0 }} 人
          </span>
        </el-card>
        <el-card shadow="never">
          <div class="metric-label">能力就绪</div>
          <strong class="metric-value">
            {{ workspace.readiness?.readyCount ?? 0 }}/{{ workspace.readiness?.totalCount ?? 0 }}
          </strong>
          <span class="metric-note">
            {{ workspace.readiness?.allReady ? '全部可用' : '仍有能力未完成配置' }}
          </span>
        </el-card>
      </div>

      <el-alert
        v-if="!workspace.readiness?.allReady"
        type="warning"
        :title="`以下能力尚未就绪：${workspace.readiness?.unreadyCapabilities?.join('、') || '未知'}`"
        description="需要先完成供应商、模型、能力和提示词配置，才能打开平台总开关。"
        show-icon
        :closable="false"
        class="section-card"
      />

      <el-card
        v-if="canUpdate"
        shadow="never"
        class="section-card emergency-card"
        :class="{ active: config.emergencyDisabled }"
      >
        <div>
          <strong>{{ config.emergencyDisabled ? '所有 AI 功能已紧急停用' : '全平台 AI 紧急停用' }}</strong>
          <p>此操作只改变紧急状态并立即生效，不会修改下面的总开关和入口文案。</p>
        </div>
        <el-button
          :type="config.emergencyDisabled ? 'success' : 'danger'"
          @click="openEmergencyDialog"
        >
          {{ config.emergencyDisabled ? '解除紧急停用' : '立即紧急停用' }}
        </el-button>
      </el-card>

      <el-card shadow="never" class="section-card">
        <template #header>
          <div class="section-header">
            <div>
              <strong>当前配置</strong>
              <span class="section-note">
                v{{ config.policyVersion }} · {{ formatDateTime(config.updatedAt) }}
              </span>
            </div>
            <el-tag type="success">已生效</el-tag>
          </div>
        </template>

        <el-descriptions :column="2" border class="current-summary">
          <el-descriptions-item label="更新人">
            {{ administratorLabel(config.updatedBy) }}
          </el-descriptions-item>
          <el-descriptions-item label="更新时间">
            {{ formatDateTime(config.updatedAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="修改原因" :span="2">
            {{ config.reason || '—' }}
          </el-descriptions-item>
        </el-descriptions>

        <el-alert
          v-if="editing"
          :title="impactTitle"
          :description="impactDescription"
          type="info"
          show-icon
          :closable="false"
          class="impact-alert"
        />

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          :disabled="!canUpdate || !editing"
        >
          <el-form-item label="平台 AI 总开关" prop="platformDefaultEnabled">
            <el-switch
              v-model="form.platformDefaultEnabled"
              active-text="全部 AI 功能开放"
              inactive-text="全部 AI 功能关闭"
            />
          </el-form-item>

          <div class="feature-list">
            <div
              v-for="(feature, index) in form.featureLabels"
              :key="feature.code"
              class="feature-item"
            >
              <div class="feature-heading">
                <strong>{{ featureCodeLabel(feature.code) }}</strong>
                <code>{{ feature.code }}</code>
              </div>
              <div class="feature-fields">
                <el-form-item
                  label="标题"
                  :prop="`featureLabels.${index}.title`"
                  :rules="requiredRule('请输入功能标题')"
                >
                  <el-input v-model.trim="feature.title" maxlength="40" />
                </el-form-item>
                <el-form-item
                  label="按钮文案"
                  :prop="`featureLabels.${index}.actionLabel`"
                  :rules="requiredRule('请输入按钮文案')"
                >
                  <el-input v-model.trim="feature.actionLabel" maxlength="20" />
                </el-form-item>
                <el-form-item
                  label="功能说明"
                  :prop="`featureLabels.${index}.description`"
                  :rules="requiredRule('请输入功能说明')"
                  class="wide"
                >
                  <el-input v-model.trim="feature.description" maxlength="160" />
                </el-form-item>
                <el-form-item
                  label="排序"
                  :prop="`featureLabels.${index}.sortOrder`"
                  :rules="sortOrderRules"
                >
                  <el-input-number
                    v-model="feature.sortOrder"
                    :min="1"
                    :max="5"
                    :step="1"
                    controls-position="right"
                  />
                </el-form-item>
                <el-form-item label="积分提示（实时）">
                  <span class="billing-hint">
                    {{ feature.costHint || '对应能力尚未配置' }}
                  </span>
                </el-form-item>
                <el-form-item label="免费次数提示（实时）">
                  <span class="billing-hint">
                    {{ feature.freeQuotaHint || '对应能力尚未配置' }}
                  </span>
                </el-form-item>
              </div>
            </div>
          </div>

          <el-form-item v-if="editing" label="修改原因" prop="reason">
            <el-input
              v-model.trim="form.reason"
              type="textarea"
              :rows="3"
              maxlength="200"
              show-word-limit
              placeholder="说明本次配置修改原因"
            />
          </el-form-item>
          <el-form-item v-if="canUpdate && editing">
            <el-button
              type="primary"
              :loading="saving"
              :disabled="form.platformDefaultEnabled && !workspace.readiness?.allReady"
              @click="saveConfig"
            >
              保存并立即生效
            </el-button>
            <el-button @click="discardChanges">放弃本地修改</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card v-if="canReadAudit" shadow="never" class="section-card">
        <AuditPanel target-type="platform_policy" target-id="platform" />
      </el-card>
    </template>

    <el-dialog
      v-model="emergencyVisible"
      :title="config?.emergencyDisabled ? '解除紧急停用' : '全平台 AI 紧急停用'"
      width="520px"
    >
      <el-alert
        :type="config?.emergencyDisabled ? 'success' : 'error'"
        :title="config?.emergencyDisabled ? '解除后将按平台总开关恢复 AI 功能。' : '确认后所有 AI 功能将立即不可用。'"
        show-icon
        :closable="false"
        class="dialog-alert"
      />
      <el-form ref="emergencyFormRef" :model="emergencyForm" :rules="emergencyRules" label-width="90px">
        <el-form-item label="操作原因" prop="reason">
          <el-input
            v-model.trim="emergencyForm.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="emergencyVisible = false">取消</el-button>
        <el-button
          :type="config?.emergencyDisabled ? 'success' : 'danger'"
          :loading="emergencyLoading"
          @click="submitEmergencyStatus"
        >
          确认立即生效
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import AuditPanel from '@/view/orderFood/components/AuditPanel.vue'
import {
  getPlatformCapabilityPolicy,
  updatePlatformCapabilityEmergencyStatus,
  updatePlatformCapabilityPolicy
} from '@/api/orderfood/ai'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodPlatformPolicy' })

const featureDefaults = [
  { code: 'dish_extract', title: '菜品资料整理', actionLabel: '整理菜品', description: '从文字或菜谱图片整理菜品信息', sortOrder: 1 },
  { code: 'cover_create', title: '生成菜品封面', actionLabel: '生成封面', description: '根据菜品内容生成封面图片', sortOrder: 2 },
  { code: 'meal_suggest', title: '不知道吃什么', actionLabel: '帮我推荐', description: '根据人数和偏好生成候选菜品', sortOrder: 3 },
  { code: 'prep_sequence', title: '备菜顺序', actionLabel: '生成顺序', description: '根据饭局菜品整理备菜顺序', sortOrder: 4 },
  { code: 'taste_profile', title: '偏好画像', actionLabel: '查看偏好', description: '根据打卡记录生成偏好画像', sortOrder: 5 }
]

const btnAuth = useBtnAuth()
const canUpdate = computed(() =>
  Boolean(btnAuth['orderfood:platform-policy:update'] || btnAuth.platformPolicyUpdate)
)
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))
const workspace = ref(null)
const config = computed(() => workspace.value?.config || null)
const loading = ref(false)
const pageError = ref('')
const saving = ref(false)
const editing = ref(false)
const formRef = ref()
const form = ref({ platformDefaultEnabled: false, featureLabels: [], reason: '' })
const rules = {
  reason: [
    { required: true, message: '请输入修改原因', trigger: 'blur' },
    { min: 2, max: 200, message: '原因长度为 2-200 字', trigger: 'blur' }
  ]
}
const requiredRule = (message) => [
  { required: true, message, trigger: 'blur' }
]
const sortOrderRules = [
  { required: true, message: '请输入排序值', trigger: 'change' },
  { type: 'number', min: 1, max: 5, message: '排序值必须为 1-5', trigger: 'change' }
]

const normalizeFeatures = (items = []) => {
  const current = new Map(items.map((item) => [item.code, item]))
  return featureDefaults
    .map((fallback) => ({
      ...fallback,
      ...(current.get(fallback.code) || {}),
      costHint: current.get(fallback.code)?.costHint || '',
      freeQuotaHint: current.get(fallback.code)?.freeQuotaHint || ''
    }))
    .sort((left, right) => left.sortOrder - right.sortOrder)
}
const hydrateForm = () => {
  form.value = {
    platformDefaultEnabled: Boolean(config.value?.platformDefaultEnabled),
    featureLabels: normalizeFeatures(config.value?.featureLabels),
    reason: ''
  }
  formRef.value?.clearValidate()
}
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const administratorLabel = (administrator) =>
  administrator?.nickname || administrator?.username || administrator?.id || '—'
const featureCodeLabel = (code) =>
  ({
    dish_extract: '菜品资料整理',
    cover_create: '菜品封面生成',
    meal_suggest: '不知道吃什么',
    prep_sequence: '饭局备菜顺序',
    taste_profile: '打卡偏好画像'
  })[code] || code
const affectedEntryNames = computed(() =>
  [...form.value.featureLabels]
    .sort((left, right) => left.sortOrder - right.sortOrder)
    .map((item) => featureCodeLabel(item.code))
)
const proposedEnabledUserCount = computed(() =>
  form.value.platformDefaultEnabled && !config.value?.emergencyDisabled
    ? Number(workspace.value?.userCounts?.normalUserCount || 0)
    : 0
)
const impactTitle = computed(() =>
  form.value.platformDefaultEnabled
    ? `保存后预计 ${proposedEnabledUserCount.value} 名正常用户可使用全部能力`
    : '保存后预计全部用户都无法使用 AI 功能'
)
const impactDescription = computed(() => {
  const entries = `涉及入口：${affectedEntryNames.value.join('、')}。`
  if (config.value?.emergencyDisabled && form.value.platformDefaultEnabled) {
    return `${entries} 当前处于紧急停用状态，保存后实际生效用户仍为 0；解除紧急停用后才会整体开放。`
  }
  return `${entries} 不存在单项开关或用户例外。`
})

const loadWorkspace = async () => {
  loading.value = true
  pageError.value = ''
  try {
    workspace.value = unwrapOrderFoodResponse(
      await getPlatformCapabilityPolicy(),
      '加载平台策略失败'
    )
    hydrateForm()
    editing.value = false
  } catch (error) {
    pageError.value = getOrderFoodErrorMessage(error, '加载平台策略失败')
  } finally {
    loading.value = false
  }
}

const beginEdit = () => {
  hydrateForm()
  editing.value = true
}

const discardChanges = () => {
  hydrateForm()
  editing.value = false
}

const refreshWorkspace = async () => {
  if (editing.value) {
    const confirmed = await ElMessageBox.confirm(
      '刷新会放弃当前未保存修改，是否继续？',
      '确认刷新',
      {
        confirmButtonText: '放弃并刷新',
        cancelButtonText: '继续编辑',
        type: 'warning'
      }
    ).then(() => true).catch(() => false)
    if (!confirmed) return
  }
  await loadWorkspace()
}

const saveConfig = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid || !config.value) return
  const uniqueSortOrders = new Set(
    form.value.featureLabels.map((item) => Number(item.sortOrder))
  )
  if (uniqueSortOrders.size !== form.value.featureLabels.length) {
    ElMessage.warning('5 个入口的排序值不能重复')
    return
  }
  const confirmed = await ElMessageBox.confirm(
    `${impactTitle.value}。${impactDescription.value}`,
    '确认保存并立即生效',
    {
      confirmButtonText: '确认保存',
      cancelButtonText: '取消',
      type: form.value.platformDefaultEnabled ? 'warning' : 'info'
    }
  ).then(() => true).catch(() => false)
  if (!confirmed) return
  saving.value = true
  try {
    const next = unwrapOrderFoodResponse(
      await updatePlatformCapabilityPolicy({
        platformDefaultEnabled: form.value.platformDefaultEnabled,
        featureLabels: form.value.featureLabels.map((item) => ({
          code: item.code,
          title: item.title,
          actionLabel: item.actionLabel,
          description: item.description,
          sortOrder: item.sortOrder
        })),
        reason: form.value.reason,
        expectedVersion: config.value.policyVersion
      }),
      '保存平台策略失败'
    )
    workspace.value = { ...workspace.value, config: next }
    editing.value = false
    hydrateForm()
    ElMessage.success('平台策略已保存并生效')
    await loadWorkspace()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('配置已被其他管理员更新，已重新加载最新值')
      await loadWorkspace()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存平台策略失败'))
    }
  } finally {
    saving.value = false
  }
}

const emergencyVisible = ref(false)
const emergencyLoading = ref(false)
const emergencyFormRef = ref()
const emergencyForm = ref({ reason: '' })
const emergencyRules = {
  reason: [
    { required: true, message: '请输入操作原因', trigger: 'blur' },
    { min: 4, max: 200, message: '原因长度为 4-200 字', trigger: 'blur' }
  ]
}
const openEmergencyDialog = async () => {
  if (editing.value) {
    const confirmed = await ElMessageBox.confirm(
      '紧急操作不会夹带当前普通配置。继续后将放弃未保存修改。',
      '放弃修改并执行独立紧急操作？',
      {
        confirmButtonText: '继续',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(() => true).catch(() => false)
    if (!confirmed) return
    discardChanges()
  }
  emergencyForm.value = { reason: '' }
  emergencyVisible.value = true
}
const submitEmergencyStatus = async () => {
  const valid = await emergencyFormRef.value?.validate().catch(() => false)
  if (!valid || !config.value) return
  emergencyLoading.value = true
  const nextDisabled = !config.value.emergencyDisabled
  try {
    unwrapOrderFoodResponse(
      await updatePlatformCapabilityEmergencyStatus({
        emergencyDisabled: nextDisabled,
        reason: emergencyForm.value.reason,
        expectedVersion: config.value.policyVersion
      }),
      nextDisabled ? '紧急停用失败' : '解除紧急停用失败'
    )
    emergencyVisible.value = false
    ElMessage.success(nextDisabled ? '全平台 AI 已紧急停用' : '已解除全平台 AI 紧急停用')
    await loadWorkspace()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('配置已发生变化，已重新加载最新值')
      emergencyVisible.value = false
      await loadWorkspace()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '紧急状态切换失败'))
    }
  } finally {
    emergencyLoading.value = false
  }
}

loadWorkspace()
</script>

<style scoped>
.page-header,
.section-header,
.emergency-card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.page-title {
  margin: 0;
  color: #1f2937;
  font-size: 22px;
}

.page-subtitle,
.section-note,
.metric-note,
.emergency-card p {
  color: #6b7280;
  font-size: 13px;
}

.page-subtitle {
  margin: 6px 0 0;
}

.section-note {
  margin-left: 10px;
}

.section-card,
.metric-grid {
  margin-top: 18px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.metric-label {
  margin-bottom: 12px;
  color: #6b7280;
}

.metric-value {
  display: block;
  color: #15803d;
  font-size: 28px;
}

.metric-note {
  display: block;
  margin-top: 6px;
}

.emergency-card {
  align-items: center;
  border-color: #fecaca;
}

.emergency-card.active {
  border-color: #fca5a5;
  background: #fff7f7;
}

.emergency-card p {
  margin: 6px 0 0;
}

.feature-list {
  display: grid;
  gap: 14px;
  margin-bottom: 20px;
}

.current-summary {
  margin-bottom: 18px;
}

.impact-alert {
  margin-bottom: 18px;
}

.feature-item {
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.feature-heading {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.feature-heading code {
  color: #6b7280;
}

.feature-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 14px;
}

.feature-fields .wide {
  grid-column: span 2;
}

.billing-hint {
  color: #475569;
  line-height: 1.6;
}

.dialog-alert {
  margin-bottom: 16px;
}

@media (max-width: 1100px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
