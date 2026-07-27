<template>
  <div class="order-food-page" v-loading="loading">
    <div class="page-header">
      <div class="page-heading">
        <h2 class="page-title">微信小程序配置</h2>
        <p class="page-subtitle">
          统一管理微信登录凭证与订阅消息通道，配置保存后立即生效。
        </p>
      </div>
      <div class="page-actions">
        <el-button
          v-if="canUpdate && config?.configured && !editing"
          class="edit-button"
          type="primary"
          @click="beginEdit"
        >
          编辑当前配置
        </el-button>
        <el-button class="refresh-button" :loading="loading" @click="refreshConfig">
          刷新
        </el-button>
      </div>
    </div>

    <el-alert
      v-if="pageError"
      :title="pageError"
      type="error"
      show-icon
      :closable="false"
      class="page-alert"
    />

    <el-alert
      v-if="config && !config.configured"
      title="微信小程序尚未配置"
      description="完成 AppID 和 AppSecret 配置前，小程序微信登录与订阅消息发送均不可用。"
      type="warning"
      show-icon
      :closable="false"
      class="page-alert"
    />

    <section v-if="config" class="wechat-config-panel">
      <header class="panel-header">
        <div class="panel-heading">
          <span class="status-mark" aria-hidden="true" />
          <div>
            <strong>当前微信连接</strong>
            <span>供小程序登录与订阅消息发送统一使用</span>
          </div>
        </div>
        <span
          class="connection-status"
          :class="{ 'is-ready': config.configured }"
        >
          <i aria-hidden="true" />
          {{ config.configured ? '已配置' : '待配置' }}
        </span>
      </header>

      <div class="panel-body">
        <div class="overview-grid">
          <article class="identity-card">
            <div class="identity-card-header">
              <span>小程序身份</span>
              <span class="version-badge">
                {{ config.version ? `v${config.version}` : '未建档' }}
              </span>
            </div>

            <div class="app-id-block">
              <span>AppID</span>
              <strong>{{ config.appId || '尚未配置' }}</strong>
              <p>微信小程序的唯一身份标识，修改后将影响后续登录凭证兑换。</p>
            </div>

            <div class="secret-summary">
              <span class="secret-mark" aria-hidden="true" />
              <div>
                <strong>AppSecret</strong>
                <span>
                  {{ config.appSecretConfigured ? '密钥已保存且不会在页面回显' : '尚未保存密钥' }}
                </span>
              </div>
              <span
                class="secret-state"
                :class="{ 'is-ready': config.appSecretConfigured }"
              >
                {{ config.appSecretConfigured ? '已安全配置' : '未配置' }}
              </span>
            </div>
          </article>

          <div class="metadata-panel">
            <div class="metadata-heading">
              <div>
                <strong>配置记录</strong>
                <span>最近一次生效配置的版本与维护信息</span>
              </div>
            </div>

            <dl class="metadata-grid">
              <div class="metadata-item">
                <dt>配置版本</dt>
                <dd>{{ config.version ? `v${config.version}` : '—' }}</dd>
              </div>
              <div class="metadata-item">
                <dt>Secret 更新时间</dt>
                <dd>{{ formatDateTime(config.secretUpdatedAt) }}</dd>
              </div>
              <div class="metadata-item">
                <dt>最近更新人</dt>
                <dd>{{ administratorLabel(config.updatedBy) }}</dd>
              </div>
              <div class="metadata-item">
                <dt>最近更新时间</dt>
                <dd>{{ formatDateTime(config.updatedAt) }}</dd>
              </div>
            </dl>

            <div class="reason-row">
              <span>最近修改原因</span>
              <strong>{{ config.reason || '未填写' }}</strong>
            </div>
          </div>
        </div>
      </div>

      <footer class="panel-footer">
        <span class="footer-mark" aria-hidden="true" />
        <div>
          <strong>生效范围</strong>
          <span>保存后立即用于新的微信登录请求与订阅消息发送，不影响既有用户数据。</span>
        </div>
      </footer>

      <section
        v-if="canUpdate && (editing || !config.configured)"
        class="config-editor"
      >
        <div class="editor-heading">
          <div>
            <strong>{{ config.configured ? '编辑微信配置' : '完成首次配置' }}</strong>
            <span>AppSecret 加密保存，接口和页面均不会返回原文。</span>
          </div>
          <span class="editor-note">保存后立即生效</span>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          class="config-form"
        >
          <div class="form-grid">
            <el-form-item label="微信小程序 AppID" prop="appId">
              <el-input
                v-model.trim="form.appId"
                maxlength="18"
                placeholder="例如 wx1234567890abcdef"
              />
              <div class="field-tip">以 wx 开头的 18 位小程序身份标识。</div>
            </el-form-item>
            <el-form-item label="微信小程序 AppSecret" prop="appSecret">
              <el-input
                v-model.trim="form.appSecret"
                type="password"
                show-password
                autocomplete="new-password"
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
            <el-form-item class="reason-field" label="修改原因" prop="reason">
              <el-input
                v-model.trim="form.reason"
                type="textarea"
                :rows="3"
                maxlength="200"
                show-word-limit
                placeholder="请说明本次配置或变更原因"
              />
            </el-form-item>
          </div>
          <div class="form-actions">
            <el-button
              v-if="config.configured"
              :disabled="saving"
              @click="discardChanges"
            >
              取消
            </el-button>
            <el-button type="primary" :loading="saving" @click="saveConfig">
              保存并立即生效
            </el-button>
          </div>
        </el-form>
      </section>
    </section>
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
.order-food-page {
  --wechat-ink: #26364d;
  --wechat-muted: #748197;
  --wechat-blue: #3378e5;
  --wechat-blue-deep: #245fbf;
  --wechat-blue-soft: #f3f7fd;
  --wechat-line: #e2eaf4;
  overflow: hidden;
  margin-top: 8px;
  padding: 0;
  border: 1px solid var(--wechat-line);
  border-radius: 10px;
  background: var(--el-bg-color);
  box-shadow: 0 8px 24px rgb(31 65 114 / 4%);
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  padding: 18px 20px;
  border-bottom: 1px solid #e8edf4;
}

.page-heading {
  min-width: 0;
}

.page-title {
  margin: 0;
  color: var(--wechat-ink);
  font-size: 22px;
  font-weight: 650;
  line-height: 1.35;
}

.page-subtitle {
  max-width: 760px;
  margin: 6px 0 0;
  color: var(--wechat-muted);
  font-size: 13px;
  line-height: 1.65;
}

.page-actions {
  display: flex;
  flex-shrink: 0;
  gap: 10px;
}

.page-actions :deep(.el-button),
.form-actions :deep(.el-button) {
  min-height: 34px;
  margin-left: 0;
  border-radius: 6px;
  padding-inline: 17px;
}

.refresh-button {
  background: var(--el-bg-color);
}

.edit-button {
  box-shadow: none;
}

.page-alert {
  width: auto;
  margin: 16px 20px 0;
}

.wechat-config-panel {
  overflow: hidden;
  margin-top: 0;
  background: var(--el-bg-color);
}

.page-alert + .wechat-config-panel {
  margin-top: 14px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 17px 20px;
  border-bottom: 1px solid var(--wechat-line);
}

.panel-heading {
  display: flex;
  align-items: center;
  gap: 11px;
}

.panel-heading > div,
.editor-heading > div,
.metadata-heading > div,
.panel-footer > div,
.secret-summary > div {
  display: flex;
  flex-direction: column;
}

.panel-heading > div {
  gap: 3px;
}

.panel-heading strong,
.metadata-heading strong,
.editor-heading strong,
.panel-footer strong {
  color: var(--wechat-ink);
}

.panel-heading strong {
  font-size: 15px;
}

.panel-heading span,
.metadata-heading span,
.editor-heading span,
.panel-footer span {
  color: var(--wechat-muted);
  font-size: 12px;
}

.status-mark {
  display: block;
  width: 4px;
  height: 32px;
  border-radius: 2px;
  background: var(--wechat-blue);
}

.connection-status,
.secret-state,
.version-badge,
.editor-note {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  white-space: nowrap;
}

.connection-status {
  gap: 6px;
  padding: 5px 9px;
  border: 1px solid #e1e7ef;
  border-radius: 6px;
  background: #f7f9fc;
  color: var(--wechat-muted);
  font-size: 12px;
  font-weight: 600;
}

.connection-status i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #a7b1c0;
}

.connection-status.is-ready {
  border-color: rgb(51 120 229 / 18%);
  background: rgb(51 120 229 / 7%);
  color: var(--wechat-blue-deep);
}

.connection-status.is-ready i {
  background: var(--wechat-blue);
}

.panel-body {
  padding: 20px;
}

.overview-grid {
  display: grid;
  grid-template-columns: minmax(310px, 0.9fr) minmax(430px, 1.35fr);
  gap: 18px;
}

.identity-card {
  min-width: 0;
  padding: 20px;
  border: 1px solid #dce8f8;
  border-radius: 10px;
  background: color-mix(in srgb, var(--el-bg-color) 82%, #eaf2ff 18%);
}

.identity-card-header,
.secret-summary,
.metadata-heading,
.editor-heading,
.form-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.identity-card-header > span:first-child {
  color: var(--wechat-blue-deep);
  font-size: 12px;
  font-weight: 700;
}

.version-badge {
  padding: 4px 8px;
  border: 1px solid #dbe5f2;
  border-radius: 5px;
  background: rgb(255 255 255 / 70%);
  color: #6c7b91;
  font-size: 11px;
  font-weight: 600;
}

.app-id-block {
  margin-top: 20px;
}

.app-id-block > span {
  display: block;
  margin-bottom: 7px;
  color: var(--wechat-muted);
  font-size: 11px;
}

.app-id-block strong {
  display: block;
  overflow-wrap: anywhere;
  color: var(--wechat-ink);
  font-family: Consolas, "Courier New", monospace;
  font-size: 19px;
  font-weight: 700;
  letter-spacing: 0.015em;
  line-height: 1.35;
}

.app-id-block p {
  margin: 9px 0 0;
  color: #68778d;
  font-size: 12px;
  line-height: 1.65;
}

.secret-summary {
  gap: 10px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #dce8f8;
}

.secret-mark {
  display: block;
  flex: 0 0 auto;
  width: 3px;
  height: 30px;
  border-radius: 2px;
  background: #87aee8;
}

.secret-summary > div {
  flex: 1;
  min-width: 0;
  gap: 3px;
}

.secret-summary strong {
  color: var(--wechat-ink);
  font-size: 13px;
}

.secret-summary > div span {
  overflow: hidden;
  color: var(--wechat-muted);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.secret-state {
  padding: 4px 8px;
  border: 1px solid #e0e6ed;
  border-radius: 5px;
  background: #f6f8fa;
  color: #7c8797;
  font-size: 11px;
  font-weight: 600;
}

.secret-state.is-ready {
  border-color: #d6e5fa;
  background: #edf4fe;
  color: var(--wechat-blue-deep);
}

.metadata-panel {
  min-width: 0;
}

.metadata-heading {
  align-items: baseline;
  margin-bottom: 12px;
}

.metadata-heading > div {
  gap: 4px;
}

.metadata-heading strong {
  font-size: 14px;
}

.metadata-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin: 0;
}

.metadata-item {
  min-width: 0;
  min-height: 76px;
  padding: 14px 15px;
  border: 1px solid var(--wechat-line);
  border-radius: 8px;
  background: color-mix(in srgb, var(--el-bg-color) 96%, #eef4fb 4%);
}

.metadata-item dt {
  margin-bottom: 8px;
  color: var(--wechat-muted);
  font-size: 11px;
}

.metadata-item dd {
  overflow: hidden;
  margin: 0;
  color: var(--wechat-ink);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reason-row {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 12px;
  margin-top: 10px;
  padding: 14px 15px;
  border: 1px solid var(--wechat-line);
  border-radius: 8px;
  background: color-mix(in srgb, var(--el-bg-color) 96%, #eef4fb 4%);
}

.reason-row span {
  color: var(--wechat-muted);
  font-size: 11px;
}

.reason-row strong {
  overflow-wrap: anywhere;
  color: var(--wechat-ink);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
}

.panel-footer {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 13px 20px;
  border-top: 1px solid var(--wechat-line);
  background: color-mix(in srgb, var(--el-bg-color) 94%, #eef3f8 6%);
}

.panel-footer > div {
  gap: 2px;
}

.panel-footer strong {
  font-size: 12px;
}

.footer-mark {
  display: block;
  flex: 0 0 auto;
  width: 7px;
  height: 7px;
  border: 2px solid #7da7e3;
  border-radius: 50%;
}

.config-editor {
  padding: 20px;
  border-top: 1px solid var(--wechat-line);
  background: color-mix(in srgb, var(--el-bg-color) 97%, #edf3fb 3%);
}

.editor-heading {
  gap: 16px;
  margin-bottom: 18px;
}

.editor-heading > div {
  gap: 4px;
}

.editor-heading strong {
  font-size: 15px;
}

.editor-note {
  padding: 4px 8px;
  border: 1px solid #dbe6f6;
  border-radius: 5px;
  background: #f4f8fd;
  color: #5f7da8 !important;
  font-size: 11px !important;
  font-weight: 600;
}

.config-form {
  width: 100%;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 2px 18px;
}

.reason-field {
  grid-column: 1 / -1;
}

.config-form :deep(.el-form-item__label) {
  color: #44536a;
  font-size: 13px;
  font-weight: 600;
}

.config-form :deep(.el-input__wrapper),
.config-form :deep(.el-textarea__inner) {
  border-radius: 6px;
  box-shadow: 0 0 0 1px var(--wechat-line) inset;
}

.config-form :deep(.el-input__wrapper.is-focus),
.config-form :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 1px var(--wechat-blue) inset;
}

.field-tip {
  margin-top: 6px;
  color: var(--wechat-muted);
  font-size: 11px;
  line-height: 1.5;
}

.form-actions {
  justify-content: flex-end;
  gap: 10px;
  padding-top: 4px;
}

@media (max-width: 880px) {
  .overview-grid,
  .form-grid {
    grid-template-columns: 1fr;
  }

  .reason-field {
    grid-column: auto;
  }
}

@media (max-width: 640px) {
  .page-header,
  .panel-header,
  .editor-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .page-actions {
    width: 100%;
  }

  .page-actions :deep(.el-button) {
    flex: 1;
  }

  .panel-body,
  .config-editor {
    padding: 16px;
  }

  .metadata-grid {
    grid-template-columns: 1fr;
  }

  .reason-row {
    grid-template-columns: 1fr;
    gap: 6px;
  }

  .secret-summary {
    align-items: flex-start;
    flex-wrap: wrap;
  }
}
</style>