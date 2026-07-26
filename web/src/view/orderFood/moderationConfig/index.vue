<template>
  <div class="order-food-page" v-loading="loading">
    <div class="page-header">
      <div>
        <h2 class="page-title">图片审核配置</h2>
        <p class="page-subtitle">
          当前接入阿里云图片审核，此配置只作用于小程序用户上传图片；AI 生成图和管理端上传图不经过自动审核。
        </p>
      </div>
      <el-button @click="loadConfig">刷新</el-button>
    </div>

    <el-alert
      v-if="pageError"
      :title="pageError"
      type="error"
      show-icon
      :closable="false"
      class="section-card"
    />

    <el-alert
      v-if="healthAlert"
      :title="healthAlert"
      :type="healthAlertType"
      show-icon
      :closable="false"
      class="section-card"
    >
      <el-button
        v-if="config?.health?.lastFailureAt"
        type="primary"
        link
        @click="openFailedRecords"
      >
        查看审核失败记录
      </el-button>
    </el-alert>

    <div v-if="config" class="layout">
      <el-card shadow="never">
        <template #header>
          <div class="section-header">
            <div>
              <strong>当前配置</strong>
              <span class="section-note">保存成功后立即生效</span>
            </div>
            <div class="header-actions">
              <el-button
                v-if="canUpdate && !editing"
                type="primary"
                plain
                @click="startEditing"
              >
                编辑当前配置
              </el-button>
              <el-tag :type="config.enabled ? 'success' : 'info'">
                {{ config.enabled ? '审核已启用' : '审核已停用' }}
              </el-tag>
            </div>
          </div>
        </template>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="130px"
          :disabled="!editing"
        >
          <el-form-item label="启用图片审核">
            <el-switch v-model="form.enabled" />
          </el-form-item>
          <el-form-item label="接口供应商">
            <el-input value="阿里云" disabled />
          </el-form-item>
          <el-form-item label="地域" prop="region">
            <el-input v-model.trim="form.region" maxlength="64" />
          </el-form-item>
          <el-form-item label="接口地址" prop="endpoint">
            <el-input v-model.trim="form.endpoint" maxlength="300" />
          </el-form-item>
          <el-form-item label="服务编码" prop="serviceCode">
            <el-input v-model.trim="form.serviceCode" maxlength="64" />
          </el-form-item>
          <el-form-item label="超时时间" prop="timeoutMs">
            <el-input-number v-model="form.timeoutMs" :min="1000" :max="10000" :step="500" />
            <span class="unit">毫秒</span>
          </el-form-item>
          <el-form-item label="重试次数" prop="retryCount">
            <el-input-number v-model="form.retryCount" :min="0" :max="3" />
          </el-form-item>
          <el-form-item label="重试间隔" prop="retryBackoffMs">
            <el-input-number v-model="form.retryBackoffMs" :min="0" :max="2000" :step="100" />
            <span class="unit">毫秒</span>
          </el-form-item>
          <el-form-item label="凭据引用" prop="credentialRef">
            <div class="credential-editor">
              <div v-if="!credentialEditing" class="credential-current">
                <span>
                  {{ config.credential?.configured ? '已配置' : '未配置' }}
                  <template v-if="config.credential?.updatedAt">
                    · {{ formatDateTime(config.credential.updatedAt) }} 更新
                  </template>
                </span>
                <el-button
                  v-if="editing && canWriteCredential"
                  type="primary"
                  link
                  @click="credentialEditing = true"
                >
                  {{ config.credential?.configured ? '更换凭据' : '配置凭据' }}
                </el-button>
              </div>
              <template v-else>
                <el-input
                  v-model.trim="form.credentialRef"
                  type="password"
                  show-password
                  maxlength="300"
                  :placeholder="credentialPlaceholder"
                />
                <div class="field-note">
                  填写格式：env://环境变量名。环境变量内容为
                  {"accessKeyId":"...","accessKeySecret":"..."}；服务端只保存环境变量引用，不回显凭据内容。
                </div>
              </template>
            </div>
          </el-form-item>
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
          <el-form-item v-if="editing">
            <el-button type="primary" :loading="saving" @click="saveConfig">
              保存并立即生效
            </el-button>
            <el-button @click="cancelEditing">放弃未保存修改</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <div class="side-column">
        <el-card shadow="never">
          <template #header><strong>运行状态</strong></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="接口供应商">
              阿里云图片审核
            </el-descriptions-item>
            <el-descriptions-item label="配置版本">
              v{{ config.configVersion }}
            </el-descriptions-item>
            <el-descriptions-item label="健康状态">
              <el-tag :type="healthTagType">{{ healthLabel }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="凭据状态">
              {{ config.credential?.configured ? '已配置' : '未配置' }}
            </el-descriptions-item>
            <el-descriptions-item label="凭据更新时间">
              {{ formatDateTime(config.credential?.updatedAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="最坏等待">
              {{ config.worstCaseWaitMs }} 毫秒
            </el-descriptions-item>
            <el-descriptions-item label="最近修改">
              {{ formatDateTime(config.updatedAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="最近连接测试">
              <template v-if="config.health?.lastConnectionTest">
                <el-tag
                  size="small"
                  :type="config.health.lastConnectionTest.success ? 'success' : 'danger'"
                >
                  {{ config.health.lastConnectionTest.success ? '成功' : '失败' }}
                </el-tag>
                {{ formatDateTime(config.health.lastConnectionTest.testedAt) }}
              </template>
              <template v-else>—</template>
            </el-descriptions-item>
            <el-descriptions-item label="最近审核成功">
              {{ formatDateTime(config.health?.lastModerationSucceededAt) }}
            </el-descriptions-item>
            <el-descriptions-item
              v-if="config.health?.consecutiveFailureCount"
              label="连续失败"
            >
              {{ config.health.consecutiveFailureCount }} 次
              <template v-if="config.health.lastFailureSafeSummary">
                · {{ config.health.lastFailureSafeSummary }}
              </template>
            </el-descriptions-item>
            <el-descriptions-item label="修改原因">
              {{ config.reason || '—' }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never">
          <template #header>
            <div class="section-header">
              <strong>配置测试</strong>
              <span class="section-note">测试当前已保存配置</span>
            </div>
          </template>
          <el-button
            v-if="canTest"
            type="primary"
            plain
            :disabled="!connectionReady || editing"
            :loading="connectionTesting"
            @click="testConnection"
          >
            测试连接
          </el-button>
          <el-upload
            v-if="canTest"
            class="test-upload"
            :auto-upload="false"
            :show-file-list="true"
            :limit="1"
            accept="image/jpeg,image/png,image/webp"
            :on-change="selectTestImage"
            :on-remove="removeTestImage"
          >
            <el-button>选择测试图片</el-button>
          </el-upload>
          <el-button
            v-if="canTest"
            :disabled="!testImage || !connectionReady || editing"
            :loading="imageTesting"
            @click="testImageModeration"
          >
            测试图片审核
          </el-button>
          <div v-if="canTest && !connectionReady" class="field-note">
            请先保存完整的地域、接口地址、服务编码和凭据配置。
          </div>
          <div v-else-if="canTest && editing" class="field-note">
            当前有未保存修改；配置测试只使用已经生效的配置，请先保存或放弃修改。
          </div>
          <el-alert
            v-if="testResult"
            :type="testResultAlertType"
            :title="testResult.safeMessage || testResult.category || '测试完成'"
            show-icon
            :closable="false"
            class="test-result"
          />
          <el-descriptions v-if="testResult" :column="1" border class="test-details">
            <el-descriptions-item label="测试类型">
              {{ testResultType === 'connection' ? '连接测试' : '图片审核测试' }}
            </el-descriptions-item>
            <el-descriptions-item label="配置版本">
              v{{ testResult.configVersion }}
            </el-descriptions-item>
            <el-descriptions-item
              v-if="testResultType === 'connection'"
              label="连接结果"
            >
              {{ testResult.success ? '成功' : '失败' }}
            </el-descriptions-item>
            <el-descriptions-item v-else label="映射结果">
              {{ moderationStatusLabel(testResult.mappedStatus) }}
            </el-descriptions-item>
            <el-descriptions-item label="结果分类">
              {{ moderationCategoryLabel(testResult.category) }}
            </el-descriptions-item>
            <el-descriptions-item
              v-if="testResultType === 'image'"
              label="风险标签"
            >
              <template v-if="testResult.riskLabels?.length">
                <el-tag
                  v-for="label in testResult.riskLabels"
                  :key="label"
                  size="small"
                  type="danger"
                  class="risk-tag"
                >
                  {{ label }}
                </el-tag>
              </template>
              <template v-else>无</template>
            </el-descriptions-item>
            <el-descriptions-item
              v-if="testResultType === 'image'"
              label="风险等级"
            >
              {{ testResult.riskLevel || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="请求 ID">
              {{ testResult.requestId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item
              v-if="testResultType === 'image'"
              label="供应商请求 ID"
            >
              {{ testResult.providerRequestId || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="耗时">
              {{ testResult.durationMs }} 毫秒
            </el-descriptions-item>
            <el-descriptions-item label="测试时间">
              {{ formatDateTime(testResult.testedAt) }}
            </el-descriptions-item>
            <el-descriptions-item
              v-if="testResultType === 'image'"
              label="临时文件清理"
            >
              {{ testResult.temporaryFileCleaned ? '已完成' : '未完成' }}
            </el-descriptions-item>
          </el-descriptions>
          <div
            v-if="testResultType === 'image' && hasProviderSummary"
            class="provider-summary"
          >
            <strong>供应商脱敏摘要</strong>
            <pre>{{ providerSummaryText }}</pre>
          </div>
        </el-card>

        <el-card shadow="never">
          <template #header><strong>固定审核规则</strong></template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="执行方式">
              {{ safetyPolicyLabel(config.safetyPolicy?.executionMode) }}
            </el-descriptions-item>
            <el-descriptions-item label="失败处理">
              {{ safetyPolicyLabel(config.safetyPolicy?.failurePolicy) }}
            </el-descriptions-item>
            <el-descriptions-item label="强制场景">
              {{ moderationSceneLabels(config.safetyPolicy?.mandatoryScenes) }}
            </el-descriptions-item>
            <el-descriptions-item label="排除场景">
              {{ moderationSceneLabels(config.safetyPolicy?.excludedScenes) }}
            </el-descriptions-item>
            <el-descriptions-item label="供应商明确通过">
              映射为“审核通过”
            </el-descriptions-item>
            <el-descriptions-item label="疑似风险或违规">
              映射为“审核不通过”
            </el-descriptions-item>
            <el-descriptions-item label="调用异常">
              映射为“审核失败”，拒绝上传
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </div>
    </div>

    <el-card v-if="config && canReadAudit" shadow="never" class="section-card">
      <AuditPanel target-type="moderation_config" />
    </el-card>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import AuditPanel from '@/view/orderFood/components/AuditPanel.vue'
import {
  getOrderFoodModerationWorkspace,
  testOrderFoodModerationConnection,
  testOrderFoodModerationImage,
  updateOrderFoodModerationConfig
} from '@/api/orderfood/moderation'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodModerationConfig' })

const router = useRouter()
const btnAuth = useBtnAuth()
const hasPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))
const canUpdate = computed(() =>
  hasPermission('orderfood:moderation-config:update', 'moderationConfigUpdate')
)
const canWriteCredential = computed(() =>
  hasPermission(
    'orderfood:moderation-config:credential-write',
    'moderationConfigCredentialWrite'
  )
)
const canTest = computed(() =>
  hasPermission('orderfood:moderation-config:test', 'moderationConfigTest')
)
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))

const config = ref(null)
const loading = ref(false)
const pageError = ref('')
const saving = ref(false)
const editing = ref(false)
const credentialEditing = ref(false)
const connectionTesting = ref(false)
const imageTesting = ref(false)
const testImage = ref(null)
const testResult = ref(null)
const testResultType = ref('')
const formRef = ref()
const form = ref({})
// validateEndpoint 校验阿里云接口地址必须是无路径和查询参数的 HTTPS 地址。
const validateEndpoint = (_rule, value, callback) => {
  try {
    const endpoint = new URL(String(value || '').trim())
    if (
      endpoint.protocol !== 'https:' ||
      !endpoint.hostname ||
      endpoint.username ||
      endpoint.password ||
      (endpoint.pathname && endpoint.pathname !== '/') ||
      endpoint.search ||
      endpoint.hash
    ) {
      callback(new Error('接口地址必须是无路径、查询参数和用户信息的 HTTPS 地址'))
      return
    }
    callback()
  } catch (_error) {
    callback(new Error('请输入有效的 HTTPS 接口地址'))
  }
}
// validateCredentialRef 校验当前生产环境实际支持的环境变量凭据引用。
const validateCredentialRef = (_rule, value, callback) => {
  if (!value || /^env:\/\/[A-Za-z_][A-Za-z0-9_]*$/.test(value)) {
    callback()
    return
  }
  callback(new Error('凭据引用格式应为 env://环境变量名'))
}
const rules = {
  region: [{ required: true, message: '请输入地域', trigger: 'blur' }],
  endpoint: [
    { required: true, message: '请输入接口地址', trigger: 'blur' },
    { validator: validateEndpoint, trigger: 'blur' }
  ],
  serviceCode: [{ required: true, message: '请输入服务编码', trigger: 'blur' }],
  credentialRef: [{ validator: validateCredentialRef, trigger: 'blur' }],
  reason: [
    { required: true, message: '请输入修改原因', trigger: 'blur' },
    { min: 2, max: 200, message: '原因长度为 2-200 字', trigger: 'blur' }
  ]
}

const credentialPlaceholder = computed(() =>
  config.value?.credential?.configured
    ? '已配置，留空保持不变'
    : '例如 env://ORDERFOOD_ALIYUN_MODERATION'
)
const healthLabel = computed(() => {
  const labels = {
    unconfigured: '未配置',
    disabled: '已停用',
    healthy: '正常',
    degraded: '降级',
    unavailable: '不可用'
  }
  return labels[config.value?.health?.status] || config.value?.health?.status || '未知'
})
const healthTagType = computed(() => {
  const status = config.value?.health?.status
  if (status === 'healthy') return 'success'
  if (status === 'degraded') return 'warning'
  if (status === 'unavailable') return 'danger'
  return 'info'
})
const healthAlert = computed(() => {
  const health = config.value?.health
  if (!health) return ''
  if (health.status === 'unavailable') {
    return health.lastFailureSafeSummary
      ? `图片审核当前不可用：${health.lastFailureSafeSummary}`
      : '图片审核当前不可用，需要审核的用户图片上传会被拒绝。'
  }
  if (health.status === 'degraded') {
    return '当前接入配置尚未通过连接测试；在确认可用前，健康状态保持为降级。'
  }
  return ''
})
const healthAlertType = computed(() =>
  config.value?.health?.status === 'unavailable' ? 'error' : 'warning'
)
const connectionReady = computed(
  () =>
    Boolean(config.value?.region?.trim()) &&
    Boolean(config.value?.endpoint?.trim()) &&
    Boolean(config.value?.serviceCode?.trim()) &&
    Boolean(config.value?.credential?.configured)
)
const testResultAlertType = computed(() => {
  if (testResultType.value === 'connection') {
    return testResult.value?.success ? 'success' : 'error'
  }
  if (testResult.value?.mappedStatus === 'passed') return 'success'
  if (testResult.value?.mappedStatus === 'rejected') return 'warning'
  return 'error'
})
const hasProviderSummary = computed(
  () =>
    testResult.value?.providerResponseSummary &&
    Object.keys(testResult.value.providerResponseSummary).length > 0
)
const providerSummaryText = computed(() =>
  hasProviderSummary.value
    ? JSON.stringify(testResult.value.providerResponseSummary, null, 2)
    : ''
)
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const moderationStatusLabel = (value) =>
  ({ passed: '审核通过', rejected: '审核不通过', failed: '审核失败' })[value] ||
  value ||
  '—'
const moderationCategoryLabel = (value) =>
  ({
    ok: '连接正常',
    passed: '审核通过',
    content_risk: '图片存在风险',
    authentication_failed: '凭据校验失败',
    rate_limited: '调用频率受限',
    timeout: '调用超时',
    unavailable: '服务不可用',
    invalid_response: '返回结果无效'
  })[value] ||
  value ||
  '—'
const safetyPolicyLabel = (value) =>
  ({
    synchronous_before_accept: '保存前同步审核（固定）',
    fail_closed: '审核异常时拒绝上传（固定）'
  })[value] ||
  value ||
  '—'
const moderationSceneLabels = (values) => {
  const labels = {
    profile_avatar: '用户头像',
    dish_cover: '菜品封面',
    dish_step: '菜品步骤图',
    dish_extract: '菜品长截图解析图片',
    checkin: '打卡照片',
    generated_cover: '生成的菜品图',
    official_dish_cover: '管理端官方菜品封面'
  }
  return values?.map((value) => labels[value] || value).join('、') || '—'
}

const hydrateForm = () => {
  form.value = {
    enabled: Boolean(config.value?.enabled),
    region: config.value?.region || '',
    endpoint: config.value?.endpoint || '',
    serviceCode: config.value?.serviceCode || '',
    timeoutMs: Number(config.value?.timeoutMs || 3000),
    retryCount: Number(config.value?.retryCount || 0),
    retryBackoffMs: Number(config.value?.retryBackoffMs || 0),
    credentialRef: '',
    reason: ''
  }
  credentialEditing.value = false
  formRef.value?.clearValidate()
}

// startEditing 进入当前配置编辑状态，不创建任何草稿版本。
const startEditing = () => {
  hydrateForm()
  editing.value = true
}

// cancelEditing 放弃本地未保存内容并恢复当前已生效配置。
const cancelEditing = () => {
  hydrateForm()
  editing.value = false
}

const loadConfig = async () => {
  loading.value = true
  pageError.value = ''
  try {
    config.value = unwrapOrderFoodResponse(
      await getOrderFoodModerationWorkspace(),
      '加载图片审核配置失败'
    )
    hydrateForm()
    editing.value = false
  } catch (error) {
    pageError.value = getOrderFoodErrorMessage(error, '加载图片审核配置失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid || !config.value) return
  if (
    form.value.enabled &&
    !config.value.credential?.configured &&
    !form.value.credentialRef
  ) {
    credentialEditing.value = true
    ElMessage.error('启用图片审核前必须配置凭据引用')
    return
  }
  if (config.value.enabled && !form.value.enabled) {
    try {
      await ElMessageBox.confirm(
        '停用后，需要审核的小程序用户图片上传会被直接拒绝，不会跳过审核。确认停用吗？',
        '确认停用图片审核',
        {
          confirmButtonText: '确认停用',
          cancelButtonText: '取消',
          type: 'error'
        }
      )
    } catch (_error) {
      return
    }
  }
  saving.value = true
  try {
    const payload = {
      enabled: form.value.enabled,
      region: form.value.region,
      endpoint: form.value.endpoint,
      serviceCode: form.value.serviceCode,
      timeoutMs: form.value.timeoutMs,
      retryCount: form.value.retryCount,
      retryBackoffMs: form.value.retryBackoffMs,
      reason: form.value.reason,
      expectedVersion: config.value.configVersion
    }
    if (form.value.credentialRef) payload.credentialRef = form.value.credentialRef
    config.value = unwrapOrderFoodResponse(
      await updateOrderFoodModerationConfig(payload),
      '保存图片审核配置失败'
    )
    hydrateForm()
    editing.value = false
    ElMessage.success('图片审核配置已保存并生效')
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('配置已被其他管理员更新，已重新加载最新值')
      await loadConfig()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存图片审核配置失败'))
    }
  } finally {
    saving.value = false
  }
}

const testConnection = async () => {
  if (!connectionReady.value) return
  connectionTesting.value = true
  try {
    testResult.value = unwrapOrderFoodResponse(
      await testOrderFoodModerationConnection({
        expectedVersion: config.value?.configVersion
      }),
      '连接测试失败'
    )
    testResultType.value = 'connection'
    if (testResult.value.success) {
      ElMessage.success(testResult.value.safeMessage || '连接测试通过')
    } else {
      ElMessage.error(testResult.value.safeMessage || '连接测试失败')
    }
    await loadConfig()
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '连接测试失败'))
  } finally {
    connectionTesting.value = false
  }
}

const selectTestImage = (uploadFile) => {
  testImage.value = uploadFile.raw
}
const removeTestImage = () => {
  testImage.value = null
}
const testImageModeration = async () => {
  if (!testImage.value || !connectionReady.value) return
  imageTesting.value = true
  try {
    testResult.value = unwrapOrderFoodResponse(
      await testOrderFoodModerationImage({
        expectedVersion: config.value?.configVersion,
        file: testImage.value
      }),
      '图片审核测试失败'
    )
    testResultType.value = 'image'
    if (testResult.value.mappedStatus === 'passed') {
      ElMessage.success(testResult.value.safeMessage || '图片审核测试通过')
    } else if (testResult.value.mappedStatus === 'rejected') {
      ElMessage.warning(testResult.value.safeMessage || '测试图片审核不通过')
    } else {
      ElMessage.error(testResult.value.safeMessage || '图片审核测试失败')
    }
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '图片审核测试失败'))
  } finally {
    imageTesting.value = false
  }
}

// openFailedRecords 打开图片审核记录并直接筛选审核失败状态。
const openFailedRecords = () => {
  router.push({ name: 'OrderFoodModerationRecords', query: { status: 'failed' } })
}

loadConfig()
</script>

<style scoped lang="scss">
.page-header,
.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.header-actions,
.credential-current {
  display: flex;
  align-items: center;
  gap: 10px;
}

.credential-editor {
  width: 100%;
}

.credential-current {
  justify-content: space-between;
  min-height: 32px;
}

.page-title {
  margin: 0;
  color: #1f2937;
  font-size: 22px;
}

.page-subtitle,
.section-note,
.field-note {
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
.layout {
  margin-top: 18px;
}

.layout {
  display: grid;
  grid-template-columns: minmax(620px, 1fr) minmax(320px, 420px);
  gap: 18px;
}

.side-column {
  display: grid;
  align-content: start;
  gap: 18px;
}

.unit {
  margin-left: 8px;
  color: #6b7280;
}

.field-note {
  width: 100%;
  margin-top: 5px;
}

.test-upload {
  display: inline-block;
  margin: 0 8px;
}

.test-result {
  margin-top: 16px;
}

.test-details,
.provider-summary {
  margin-top: 14px;
}

.risk-tag {
  margin: 0 6px 4px 0;
}

.provider-summary pre {
  max-height: 240px;
  margin: 10px 0 0;
  padding: 12px;
  overflow: auto;
  color: #374151;
  background: #f3f4f6;
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-all;
}

@media (max-width: 1100px) {
  .layout {
    grid-template-columns: 1fr;
  }
}
</style>
