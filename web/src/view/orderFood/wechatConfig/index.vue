<template>
  <div class="order-food-page" v-loading="loading">
    <div class="page-header">
      <div>
        <h2 class="page-title">微信小程序配置</h2>
        <p class="page-subtitle">
          统一用于微信登录凭证兑换和订阅消息发送，保存后立即生效。
        </p>
      </div>
      <div class="header-actions">
        <el-button
          v-if="canUpdate && config?.configured && !editing"
          type="primary"
          @click="beginEdit"
        >
          编辑当前配置
        </el-button>
        <el-button :loading="loading" @click="refreshConfig">刷新</el-button>
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

    <el-alert
      v-if="config && !config.configured"
      title="微信小程序尚未配置"
      description="完成 AppID 和 AppSecret 配置前，小程序微信登录与订阅消息发送均不可用。"
      type="warning"
      show-icon
      :closable="false"
      class="section-card"
    />

    <el-card v-if="config" shadow="never" class="section-card">
      <template #header>
        <div class="section-header">
          <strong>当前配置</strong>
          <el-tag :type="config.configured ? 'success' : 'warning'">
            {{ config.configured ? '已配置' : '未配置' }}
          </el-tag>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="AppID">
          {{ config.appId || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="AppSecret">
          <el-tag :type="config.appSecretConfigured ? 'success' : 'warning'">
            {{ config.appSecretConfigured ? '已安全配置' : '未配置' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="配置版本">
          {{ config.version ? `v${config.version}` : '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="Secret 更新时间">
          {{ formatDateTime(config.secretUpdatedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="最近更新人">
          {{ administratorLabel(config.updatedBy) }}
        </el-descriptions-item>
        <el-descriptions-item label="最近更新时间">
          {{ formatDateTime(config.updatedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="修改原因" :span="2">
          {{ config.reason || '—' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card
      v-if="config && canUpdate && (editing || !config.configured)"
      shadow="never"
      class="section-card"
    >
      <template #header>
        <div class="section-header">
          <strong>{{ config.configured ? '编辑配置' : '首次配置' }}</strong>
          <span class="section-note">AppSecret 加密保存，接口不会回显原文</span>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="config-form"
      >
        <el-form-item label="微信小程序 AppID" prop="appId">
          <el-input
            v-model.trim="form.appId"
            maxlength="18"
            placeholder="例如 wx1234567890abcdef"
          />
        </el-form-item>
        <el-form-item label="微信小程序 AppSecret" prop="appSecret">
          <el-input
            v-model.trim="form.appSecret"
            type="password"
            show-password
            maxlength="32"
            :placeholder="
              config.appSecretConfigured
                ? '留空表示保持当前 AppSecret'
                : '请输入 32 位 AppSecret'
            "
          />
          <div class="field-tip">
            更换 AppID 时必须同时填写对应的新 AppSecret。
          </div>
        </el-form-item>
        <el-form-item label="修改原因" prop="reason">
          <el-input
            v-model.trim="form.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="请说明本次配置或变更原因"
          />
        </el-form-item>
        <div class="form-actions">
          <el-button type="primary" :loading="saving" @click="saveConfig">
            保存并立即生效
          </el-button>
          <el-button v-if="config.configured" :disabled="saving" @click="discardChanges">
            取消
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  getOrderFoodWeChatConfig,
  updateOrderFoodWeChatConfig
} from '@/api/orderfood/wechat'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'
import { formatOrderFoodDateTime } from '@/view/orderFood/utils/time'

defineOptions({ name: 'OrderFoodWeChatConfig' })

const btnAuth = useBtnAuth()
const canUpdate = computed(() =>
  Boolean(btnAuth['orderfood:wechat-config:update'])
)
const config = ref(null)
const loading = ref(false)
const saving = ref(false)
const editing = ref(false)
const pageError = ref('')
const formRef = ref()
const form = reactive({
  appId: '',
  appSecret: '',
  reason: ''
})

const validateAppID = (_rule, value, callback) => {
  if (!/^wx[0-9A-Za-z]{16}$/.test(String(value || '').trim())) {
    callback(new Error('请输入以 wx 开头的 18 位 AppID'))
    return
  }
  callback()
}

const validateAppSecret = (_rule, value, callback) => {
  const secret = String(value || '').trim()
  if (!secret && config.value?.appSecretConfigured && form.appId === config.value.appId) {
    callback()
    return
  }
  if (!/^[0-9A-Za-z]{32}$/.test(secret)) {
    callback(new Error('请输入 32 位 AppSecret'))
    return
  }
  callback()
}

const rules = {
  appId: [{ validator: validateAppID, trigger: 'blur' }],
  appSecret: [{ validator: validateAppSecret, trigger: 'blur' }],
  reason: [
    { required: true, message: '请输入修改原因', trigger: 'blur' },
    { min: 2, max: 200, message: '原因长度为 2-200 字', trigger: 'blur' }
  ]
}

const hydrateForm = () => {
  form.appId = config.value?.appId || ''
  form.appSecret = ''
  form.reason = ''
  formRef.value?.clearValidate()
}

const administratorLabel = (administrator) => {
  if (!administrator) return '—'
  return administrator.nickname || administrator.username || administrator.id || '—'
}

const formatDateTime = (value) =>
  value ? formatOrderFoodDateTime(value) : '—'

const loadConfig = async () => {
  loading.value = true
  pageError.value = ''
  try {
    config.value = unwrapOrderFoodResponse(
      await getOrderFoodWeChatConfig(),
      '加载微信配置失败'
    )
    hydrateForm()
    editing.value = !config.value.configured && canUpdate.value
  } catch (error) {
    pageError.value = getOrderFoodErrorMessage(error, '加载微信配置失败')
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

const refreshConfig = async () => {
  if (editing.value && config.value?.configured) {
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
  await loadConfig()
}

const saveConfig = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid || !config.value) return
  const confirmed = await ElMessageBox.confirm(
    '保存后将立即用于微信登录和订阅消息发送，是否继续？',
    '确认保存并立即生效',
    {
      confirmButtonText: '确认保存',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => true).catch(() => false)
  if (!confirmed) return

  saving.value = true
  try {
    const payload = {
      appId: form.appId,
      reason: form.reason,
      expectedVersion: Number(config.value.version || 0)
    }
    if (form.appSecret) payload.appSecret = form.appSecret
    config.value = unwrapOrderFoodResponse(
      await updateOrderFoodWeChatConfig(payload),
      '保存微信配置失败'
    )
    hydrateForm()
    editing.value = false
    ElMessage.success('微信小程序配置已保存并生效')
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('配置已被其他管理员更新，已重新加载最新值')
      await loadConfig()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存微信配置失败'))
    }
  } finally {
    saving.value = false
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.page-header,
.header-actions,
.section-header,
.form-actions {
  display: flex;
  align-items: center;
}

.page-header,
.section-header {
  justify-content: space-between;
  gap: 16px;
}

.page-title {
  margin: 0;
  font-size: 22px;
}

.page-subtitle,
.section-note,
.field-tip {
  color: var(--el-text-color-secondary);
}

.page-subtitle {
  margin: 6px 0 0;
}

.header-actions,
.form-actions {
  gap: 10px;
}

.section-card {
  margin-top: 16px;
}

.config-form {
  max-width: 720px;
}

.field-tip {
  margin-top: 6px;
  font-size: 12px;
}
</style>
