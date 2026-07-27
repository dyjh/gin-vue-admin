<template>
  <div v-loading="loading" class="order-food-page platform-policy-page">
    <div class="gva-table-box policy-workspace">
      <div class="page-hero">
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
      </div>

      <el-alert
        v-if="pageError"
        :title="pageError"
        type="error"
        show-icon
        :closable="false"
        class="workspace-alert"
      />

      <template v-if="workspace">
        <div class="metric-grid">
          <div class="metric-item">
            <span class="metric-index">01</span>
            <div>
              <div class="metric-label">平台 AI 总开关</div>
              <strong class="metric-status" :class="{ muted: !config.platformDefaultEnabled }">
                {{ config.platformDefaultEnabled ? '全部开放' : '全部关闭' }}
              </strong>
            </div>
          </div>
          <div class="metric-item">
            <span class="metric-index">02</span>
            <div>
              <div class="metric-label">紧急状态</div>
              <strong class="metric-status" :class="{ danger: config.emergencyDisabled }">
                {{ config.emergencyDisabled ? '已紧急停用' : '运行正常' }}
              </strong>
            </div>
          </div>
          <div class="metric-item">
            <span class="metric-index">03</span>
            <div>
              <div class="metric-label">当前生效用户</div>
              <strong class="metric-value">
                {{ workspace.userCounts?.effectiveEnabledUserCount ?? 0 }}
              </strong>
              <span class="metric-note">
                正常用户 {{ workspace.userCounts?.normalUserCount ?? 0 }} 人
              </span>
            </div>
          </div>
          <div class="metric-item">
            <span class="metric-index">04</span>
            <div>
              <div class="metric-label">能力就绪</div>
              <strong class="metric-value">
                {{ workspace.readiness?.readyCount ?? 0 }}/{{ workspace.readiness?.totalCount ?? 0 }}
              </strong>
              <span class="metric-note">
                {{ workspace.readiness?.allReady ? '全部可用' : '仍有能力未完成配置' }}
              </span>
            </div>
          </div>
        </div>

        <el-alert
          v-if="!workspace.readiness?.allReady"
          type="warning"
          :title="`以下能力尚未就绪：${unreadyCapabilityLabels.join('、') || '未知'}`"
          description="需要先完成供应商、模型、能力和提示词配置，才能打开平台总开关。"
          show-icon
          :closable="false"
          class="workspace-alert readiness-alert"
        />

        <div
          v-if="canUpdate"
          class="emergency-strip"
          :class="{ active: config.emergencyDisabled }"
        >
          <div class="emergency-copy">
            <span class="emergency-mark">!</span>
            <div>
              <strong>{{ config.emergencyDisabled ? '所有 AI 功能已紧急停用' : '全平台 AI 紧急停用' }}</strong>
              <p>独立于常规配置的应急操作，切换后立即生效，不修改总开关与入口文案。</p>
            </div>
          </div>
          <el-button
            :type="config.emergencyDisabled ? 'success' : 'danger'"
            plain
            @click="openEmergencyDialog"
          >
            {{ config.emergencyDisabled ? '解除紧急停用' : '立即紧急停用' }}
          </el-button>
        </div>
      </template>
    </div>

    <template v-if="workspace">
      <div class="gva-table-box configuration-card">
        <div class="configuration-header">
          <div>
            <strong>当前配置</strong>
            <span>v{{ config.policyVersion }} · {{ formatDateTime(config.updatedAt) }}</span>
          </div>
          <span class="effective-status"><i /> 已生效</span>
        </div>

        <div class="config-meta">
          <div>
            <span>更新人</span>
            <strong>{{ administratorLabel(config.updatedBy) }}</strong>
          </div>
          <div>
            <span>更新时间</span>
            <strong>{{ formatDateTime(config.updatedAt) }}</strong>
          </div>
          <div class="reason">
            <span>修改原因</span>
            <strong>{{ config.reason || '—' }}</strong>
          </div>
        </div>

        <el-alert
          v-if="editing"
          :title="impactTitle"
          :description="impactDescription"
          type="info"
          show-icon
          :closable="false"
          class="impact-alert"
        />

        <template v-if="!editing">
          <div class="switch-summary">
            <div>
              <strong>平台 AI 总开关</strong>
              <span>统一决定正常用户是否能够使用下列 AI 入口。</span>
            </div>
            <span
              class="switch-state"
              :class="{ enabled: config.platformDefaultEnabled }"
            >
              {{ config.platformDefaultEnabled ? '全部 AI 功能开放' : '全部 AI 功能关闭' }}
            </span>
          </div>

          <div class="feature-summary-list">
            <div
              v-for="feature in form.featureLabels"
              :key="feature.code"
              class="feature-summary-item"
            >
              <span class="feature-order">{{ feature.sortOrder }}</span>
              <div class="feature-summary-copy">
                <div>
                  <strong>{{ featureCodeLabel(feature.code) }}</strong>
                  <code>{{ feature.code }}</code>
                </div>
                <p>{{ feature.description }}</p>
              </div>
              <div class="feature-summary-field">
                <span>入口标题</span>
                <strong>{{ feature.title }}</strong>
              </div>
              <div class="feature-summary-field">
                <span>按钮文案</span>
                <strong>{{ feature.actionLabel }}</strong>
              </div>
              <div class="feature-summary-billing">
                <span>积分：{{ feature.costHint || '尚未配置' }}</span>
                <span>免费：{{ feature.freeQuotaHint || '尚未配置' }}</span>
              </div>
            </div>
          </div>
        </template>

        <el-form
          v-else
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          :disabled="!canUpdate"
          class="policy-form"
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

          <el-form-item label="修改原因" prop="reason">
            <el-input
              v-model.trim="form.reason"
              type="textarea"
              :rows="3"
              maxlength="200"
              show-word-limit
              placeholder="说明本次配置修改原因"
            />
          </el-form-item>
          <el-form-item>
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
      </div>

      <div v-if="canReadAudit" class="gva-table-box audit-card">
        <AuditPanel target-type="platform_policy" target-id="platform" />
      </div>
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
const capabilityCodeLabel = (code) =>
  ({
    dish_text_extract: '菜品文本解析',
    recipe_image_extract: '菜谱长截图解析',
    dish_cover_create: '菜品封面生成',
    checkin_image_analyze: '打卡图片分析',
    meal_suggest: '饭局菜品建议',
    prep_sequence: '备菜顺序生成'
  })[code] || code
const unreadyCapabilityLabels = computed(() =>
  (workspace.value?.readiness?.unreadyCapabilities || []).map(capabilityCodeLabel)
)
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
.platform-policy-page {
  --policy-blue: #2563eb;
  --policy-border: #e4ebf4;
  --policy-text: #243247;
  --policy-muted: #6d7a8c;
}

.policy-workspace,
.configuration-card,
.audit-card {
  overflow: hidden;
  padding: 0;
  border: 1px solid var(--policy-border);
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(31, 65, 114, 0.04);
}

.configuration-card,
.audit-card {
  margin-top: 14px;
}

.page-hero {
  padding: 18px 20px;
  border-bottom: 1px solid #e8edf4;
}

.page-header,
.configuration-header,
.emergency-strip,
.emergency-copy,
.switch-summary,
.feature-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.page-header {
  align-items: flex-start;
}

.header-actions {
  display: flex;
  flex: none;
  gap: 10px;
}

.page-title {
  margin: 0;
  color: var(--policy-text);
  font-size: 22px;
  font-weight: 650;
  line-height: 1.35;
}

.page-subtitle {
  margin: 6px 0 0;
  color: var(--policy-muted);
  font-size: 13px;
  line-height: 1.65;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.metric-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
  padding: 18px 20px;
  border-right: 1px solid #e8edf4;
}

.metric-item:last-child {
  border-right: 0;
}

.metric-index {
  display: inline-flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  color: #4771ae;
  background: #edf4fd;
  font-size: 11px;
  font-weight: 600;
}

.metric-label {
  margin-bottom: 7px;
  color: #647288;
  font-size: 13px;
}

.metric-status,
.metric-value {
  display: block;
  color: var(--policy-blue);
  font-size: 21px;
  font-weight: 650;
  line-height: 1.25;
}

.metric-status.muted {
  color: #64748b;
}

.metric-status.danger {
  color: #c24146;
}

.metric-note {
  display: block;
  margin-top: 5px;
  color: #8a95a6;
  font-size: 12px;
  line-height: 1.45;
}

.workspace-alert {
  width: auto;
  margin: 16px 20px 0;
}

.readiness-alert {
  border: 1px solid #f0dfbf;
  background: #fffaf2;
}

.emergency-strip {
  margin: 16px 20px 18px;
  padding: 13px 15px;
  border: 1px solid #f2d4d6;
  border-radius: 8px;
  background: #fffafb;
}

.emergency-strip.active {
  border-color: #efb9bd;
  background: #fff5f6;
}

.emergency-copy {
  justify-content: flex-start;
  min-width: 0;
  gap: 11px;
}

.emergency-mark {
  display: inline-flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  color: #b94a50;
  background: #fbeaec;
  font-size: 14px;
  font-weight: 700;
}

.emergency-copy strong {
  color: #553236;
  font-size: 14px;
  font-weight: 600;
}

.emergency-copy p {
  margin: 4px 0 0;
  color: #80666a;
  font-size: 12px;
  line-height: 1.5;
}

.configuration-header {
  min-height: 56px;
  padding: 0 20px;
  border-bottom: 1px solid #e8edf4;
}

.configuration-header > div {
  display: flex;
  align-items: center;
  gap: 10px;
}

.configuration-header strong {
  color: var(--policy-text);
  font-size: 16px;
  font-weight: 600;
}

.configuration-header > div > span {
  color: #8290a3;
  font-size: 12px;
}

.effective-status,
.switch-state {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 4px 9px;
  border: 1px solid #cfe0f8;
  border-radius: 6px;
  color: #2a64aa;
  background: #f4f8fe;
  font-size: 12px;
  white-space: nowrap;
}

.effective-status i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--policy-blue);
}

.config-meta {
  display: grid;
  grid-template-columns: 0.65fr 1fr 1.8fr;
  gap: 20px;
  padding: 13px 20px;
  border-bottom: 1px solid #e8edf4;
  background: #f8fafc;
}

.config-meta > div {
  min-width: 0;
}

.config-meta span,
.feature-summary-field span,
.feature-summary-billing span {
  display: block;
  color: #8793a5;
  font-size: 12px;
}

.config-meta strong {
  display: block;
  margin-top: 4px;
  overflow: hidden;
  color: #46556b;
  font-size: 13px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.impact-alert {
  margin: 16px 20px 0;
}

.switch-summary {
  min-height: 66px;
  padding: 0 20px;
  border-bottom: 1px solid #e8edf4;
}

.switch-summary > div strong,
.switch-summary > div span {
  display: block;
}

.switch-summary > div strong {
  color: var(--policy-text);
  font-size: 14px;
  font-weight: 600;
}

.switch-summary > div span {
  margin-top: 4px;
  color: var(--policy-muted);
  font-size: 12px;
}

.switch-state {
  color: #66758a;
  border-color: #dde4ec;
  background: #f7f9fb;
}

.switch-state.enabled {
  color: #1d6d4a;
  border-color: #cce9dc;
  background: #f1faf6;
}

.feature-summary-item {
  display: grid;
  grid-template-columns: 32px minmax(220px, 1.35fr) minmax(120px, 0.55fr) minmax(110px, 0.5fr) minmax(190px, 0.8fr);
  align-items: center;
  gap: 14px;
  min-height: 76px;
  padding: 10px 20px;
  border-bottom: 1px solid #edf1f6;
}

.feature-summary-item:last-child {
  border-bottom: 0;
}

.feature-summary-item:hover {
  background: #fbfcfe;
}

.feature-order {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid #d8e5f5;
  border-radius: 50%;
  color: #3869a8;
  background: #f5f9fe;
  font-size: 12px;
  font-weight: 600;
}

.feature-summary-copy {
  min-width: 0;
}

.feature-summary-copy > div {
  display: flex;
  align-items: center;
  gap: 8px;
}

.feature-summary-copy strong {
  color: #334258;
  font-size: 14px;
  font-weight: 600;
}

.feature-summary-copy code,
.feature-heading code {
  padding: 2px 6px;
  border-radius: 4px;
  color: #708097;
  background: #f1f4f8;
  font-size: 11px;
}

.feature-summary-copy p {
  margin: 5px 0 0;
  overflow: hidden;
  color: #788598;
  font-size: 12px;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feature-summary-field {
  min-width: 0;
}

.feature-summary-field strong {
  display: block;
  margin-top: 5px;
  overflow: hidden;
  color: #46556b;
  font-size: 13px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feature-summary-billing {
  min-width: 0;
  padding-left: 14px;
  border-left: 1px solid #e7ecf2;
}

.feature-summary-billing span {
  overflow: hidden;
  color: #6f7d91;
  line-height: 1.7;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.policy-form {
  padding: 18px 20px 2px;
}

.feature-list {
  display: grid;
  gap: 12px;
  margin-bottom: 20px;
}

.feature-item {
  padding: 16px;
  border: 1px solid #e1e8f1;
  border-radius: 8px;
  background: #fbfcfe;
}

.feature-heading {
  margin-bottom: 12px;
}

.feature-heading strong {
  color: #334258;
  font-size: 14px;
  font-weight: 600;
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
  color: #536176;
  line-height: 1.6;
}

.audit-card {
  padding: 18px 20px;
}

.dialog-alert {
  margin-bottom: 16px;
}

@media (max-width: 1100px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .metric-item:nth-child(2) {
    border-right: 0;
  }

  .metric-item:nth-child(n + 3) {
    border-top: 1px solid #e8edf4;
  }

  .feature-summary-item {
    grid-template-columns: 32px minmax(0, 1fr) 130px 120px;
  }

  .feature-summary-billing {
    grid-column: 2 / -1;
    padding: 0;
    border-left: 0;
  }

  .feature-summary-billing span {
    display: inline-block;
    margin-right: 18px;
  }
}

@media (max-width: 760px) {
  .page-header,
  .emergency-strip,
  .switch-summary {
    align-items: stretch;
    flex-direction: column;
  }

  .header-actions {
    width: 100%;
  }

  .header-actions .el-button {
    flex: 1;
    margin-left: 0;
  }

  .metric-grid,
  .config-meta,
  .feature-fields {
    grid-template-columns: 1fr;
  }

  .metric-item {
    border-top: 1px solid #e8edf4;
    border-right: 0;
  }

  .metric-item:first-child {
    border-top: 0;
  }

  .emergency-strip .el-button {
    width: 100%;
    margin-left: 0;
  }

  .configuration-header {
    align-items: flex-start;
    min-height: 0;
    padding: 14px 16px;
  }

  .configuration-header > div {
    align-items: flex-start;
    flex-direction: column;
    gap: 3px;
  }

  .feature-summary-item {
    grid-template-columns: 32px minmax(0, 1fr);
    padding: 14px 16px;
  }

  .feature-summary-field,
  .feature-summary-billing {
    grid-column: 2;
  }

  .feature-summary-billing span {
    display: block;
    margin-right: 0;
  }

  .feature-fields .wide {
    grid-column: 1;
  }
}
</style>
