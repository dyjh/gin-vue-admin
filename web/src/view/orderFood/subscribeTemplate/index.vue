<template>
  <div class="subscribe-scenes-page">
    <section class="scene-hero">
      <div class="hero-copy">
        <div class="hero-eyebrow">WECHAT SUBSCRIPTION</div>
        <h1>订阅场景配置</h1>
        <p>
          业务场景、触发时机和消息字段由系统固定。这里只负责绑定微信模板，避免运营配置改变业务语义。
        </p>
      </div>
      <div class="hero-summary" aria-label="场景配置概览">
        <div>
          <strong>{{ scenes.length }}</strong>
          <span>固定场景</span>
        </div>
        <i />
        <div>
          <strong>{{ configuredCount }}</strong>
          <span>已配置</span>
        </div>
        <i />
        <div>
          <strong>{{ enabledCount }}</strong>
          <span>运行中</span>
        </div>
      </div>
    </section>

    <el-alert
      v-if="listError"
      :title="listError"
      type="error"
      show-icon
      :closable="false"
      class="scene-alert"
    >
      <template #default>
        <el-button link type="primary" @click="loadScenes">重新加载</el-button>
      </template>
    </el-alert>

    <div v-loading="listLoading" class="scene-grid">
      <article
        v-for="scene in scenes"
        :key="scene.scene"
        class="scene-card"
        :class="{
          'is-enabled': scene.enabled,
          'is-focused': focusedScene === scene.scene
        }"
      >
        <header class="scene-card__header">
          <div class="scene-identity">
            <div>
              <div class="scene-code">{{ scene.scene }}</div>
              <h2>{{ scene.name }}</h2>
            </div>
          </div>
          <el-tag :type="statusMeta(scene).type" effect="light" round>
            <span class="status-dot" />{{ statusMeta(scene).label }}
          </el-tag>
        </header>

        <p class="scene-purpose">{{ scene.purpose }}</p>

        <div class="rule-strip">
          <div class="rule-item">
            <span class="rule-index">01</span>
            <div>
              <small>触发条件</small>
              <strong>{{ scene.triggerDescription }}</strong>
            </div>
          </div>
          <div class="rule-item">
            <span class="rule-index">02</span>
            <div>
              <small>接收对象</small>
              <strong>{{ scene.recipientDescription }}</strong>
            </div>
          </div>
        </div>

        <div class="binding-panel">
          <div class="binding-panel__heading">
            <div>
              <small>微信模板绑定</small>
              <strong v-if="scene.configured">
                {{ scene.wechatTemplateIdMasked || '已配置' }}
              </strong>
              <strong v-else>尚未配置模板 ID</strong>
            </div>
            <span v-if="scene.configured" class="version-badge">v{{ scene.version }}</span>
          </div>

          <div class="field-list">
            <div
              v-for="field in scene.requiredFields"
              :key="field.key"
              class="field-chip"
            >
              <span>{{ field.label }}</span>
              <code>{{ field.key }}</code>
            </div>
          </div>

          <div v-if="scene.configured" class="mapping-summary">
            <span>当前映射</span>
            <p>{{ (scene.fieldMappingSummary || []).join(' · ') || '—' }}</p>
          </div>
          <div v-else class="empty-binding">
            首次保存后保持停用，确认微信模板字段无误再单独启用。
          </div>
        </div>

        <footer class="scene-card__footer">
          <div class="scene-metrics">
            <span><b>{{ scene.sendCount || 0 }}</b>发送</span>
            <span><b>{{ scene.failureCount || 0 }}</b>失败</span>
            <span><b>{{ formatDateTime(scene.updatedAt) }}</b>更新</span>
          </div>
          <div class="scene-actions">
            <el-button
              v-if="canReadLogs"
              plain
              @click="openLogs(scene)"
            >
              发送记录
            </el-button>
            <el-button
              v-if="canChangeStatus && scene.configured"
              :type="scene.enabled ? 'danger' : 'success'"
              plain
              @click="changeStatus(scene)"
            >
              {{ scene.enabled ? '停用场景' : '启用场景' }}
            </el-button>
            <el-button
              v-if="canUpdate"
              type="primary"
              @click="openConfigure(scene)"
            >
              {{ scene.configured ? '调整配置' : '配置模板' }}
            </el-button>
          </div>
        </footer>
      </article>

      <el-empty
        v-if="!listLoading && !listError && scenes.length === 0"
        description="暂无固定订阅场景"
      />
    </div>

    <el-dialog
      v-model="editorVisible"
      :title="editorForm.configured ? '调整场景配置' : '配置订阅场景'"
      width="720px"
      destroy-on-close
      class="scene-editor-dialog"
      @closed="resetEditor"
    >
      <div v-loading="detailLoading">
        <div class="editor-context">
          <div class="editor-context__mark">饭</div>
          <div>
            <span>{{ editorForm.scene }}</span>
            <h3>{{ editorForm.name || '饭局最终结果' }}</h3>
            <p>场景规则与业务字段由系统维护，管理员无需创建消息场景。</p>
          </div>
          <el-tag :type="editorForm.enabled ? 'success' : 'info'" round>
            {{ editorForm.enabled ? '当前已启用' : '当前未启用' }}
          </el-tag>
        </div>

        <el-form
          ref="editorFormRef"
          :model="editorForm"
          :rules="editorRules"
          label-position="top"
          class="scene-editor-form"
        >
          <el-form-item label="微信模板 ID" prop="wechatTemplateId">
            <el-input
              v-model.trim="editorForm.wechatTemplateId"
              maxlength="100"
              show-word-limit
              clearable
              placeholder="请输入微信公众平台中的模板 ID"
            />
            <div class="form-help">
              更换模板 ID 不会自动改变当前启用状态；新场景首次保存时默认停用。
            </div>
          </el-form-item>

          <el-form-item prop="mappingRows">
            <template #label>
              <div class="mapping-label">
                <span>固定字段映射</span>
                <small>右侧填写微信模板字段名，例如 thing1、phrase2、time3</small>
              </div>
            </template>
            <div class="mapping-editor">
              <div class="mapping-editor__head">
                <span>业务字段（不可修改）</span>
                <span>微信模板字段</span>
              </div>
              <div
                v-for="row in editorForm.mappingRows"
                :key="row.businessField"
                class="mapping-row"
              >
                <div class="business-field">
                  <div>
                    <strong>{{ row.label }}</strong>
                    <code>{{ row.businessField }}</code>
                  </div>
                  <small>{{ row.description }}</small>
                </div>
                <span class="mapping-arrow">→</span>
                <el-input
                  v-model.trim="row.wechatField"
                  maxlength="40"
                  :placeholder="`${row.label}对应的字段名`"
                />
              </div>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <div class="dialog-footer-note">
          <span>固定 {{ editorForm.mappingRows.length }} 个业务字段，不支持增删</span>
          <div>
            <el-button @click="editorVisible = false">取消</el-button>
            <el-button
              type="primary"
              :loading="editorLoading"
              :disabled="detailLoading"
              @click="submitEditor"
            >
              保存配置
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import {
  configureSubscribeScene,
  getSubscribeSceneDetail,
  getSubscribeSceneList,
  updateSubscribeSceneStatus
} from '@/api/orderfood/message'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodSubscribeScenes'
})

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const hasBtnPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))
const canUpdate = computed(() =>
  hasBtnPermission('orderfood:subscribe-scene:update', 'subscribeSceneUpdate')
)
const canChangeStatus = computed(() =>
  hasBtnPermission('orderfood:subscribe-scene:status', 'subscribeSceneStatus')
)
const canReadLogs = computed(() =>
  hasBtnPermission('orderfood:subscribe-log:read', 'subscribeLogRead')
)

const scenes = ref([])
const listLoading = ref(false)
const listError = ref('')
const focusedScene = ref(String(route.query.scene || ''))
const configuredCount = computed(
  () => scenes.value.filter((scene) => scene.configured).length
)
const enabledCount = computed(
  () => scenes.value.filter((scene) => scene.enabled).length
)

const loadScenes = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    scenes.value =
      unwrapOrderFoodResponse(
        await getSubscribeSceneList(),
        '订阅场景加载失败'
      ) || []
  } catch (error) {
    scenes.value = []
    listError.value = getOrderFoodErrorMessage(error, '订阅场景加载失败')
  } finally {
    listLoading.value = false
  }
}

const statusMeta = (scene) => {
  if (!scene.configured) return { label: '待配置', type: 'warning' }
  if (scene.enabled) return { label: '运行中', type: 'success' }
  return { label: '已停用', type: 'info' }
}

const createDefaultEditor = () => ({
  scene: '',
  name: '',
  configured: false,
  enabled: false,
  wechatTemplateId: '',
  expectedVersion: undefined,
  mappingRows: []
})
const editorVisible = ref(false)
const detailLoading = ref(false)
const editorLoading = ref(false)
const editorFormRef = ref(null)
const editorForm = ref(createDefaultEditor())

const mappingValidator = (_rule, rows, callback) => {
  if (!Array.isArray(rows) || rows.length === 0) {
    callback(new Error('固定业务字段定义缺失，请刷新后重试'))
    return
  }
  const values = rows.map((row) => row.wechatField?.trim())
  if (values.some((value) => !value)) {
    callback(new Error('请填写每个业务字段对应的微信模板字段'))
    return
  }
  if (new Set(values).size !== values.length) {
    callback(new Error('同一个微信模板字段不能被重复映射'))
    return
  }
  callback()
}
const editorRules = {
  wechatTemplateId: [
    { required: true, message: '请输入微信模板 ID', trigger: 'blur' },
    { max: 100, message: '微信模板 ID 不能超过 100 个字符', trigger: 'blur' }
  ],
  mappingRows: [{ validator: mappingValidator, trigger: 'change' }]
}

const resetEditor = () => {
  editorForm.value = createDefaultEditor()
  editorFormRef.value?.clearValidate()
}

const openConfigure = async (scene) => {
  focusedScene.value = scene.scene
  editorVisible.value = true
  detailLoading.value = true
  try {
    const detail = unwrapOrderFoodResponse(
      await getSubscribeSceneDetail(scene.scene),
      '场景配置加载失败'
    )
    editorForm.value = {
      scene: detail.scene,
      name: detail.name,
      configured: Boolean(detail.configured),
      enabled: Boolean(detail.enabled),
      wechatTemplateId: detail.wechatTemplateId || '',
      expectedVersion: detail.configured ? Number(detail.version) : undefined,
      mappingRows: (detail.requiredFields || []).map((field) => ({
        businessField: field.key,
        label: field.label,
        description: field.description,
        wechatField: detail.fieldMappings?.[field.key] || ''
      }))
    }
  } catch (error) {
    editorVisible.value = false
    ElMessage.error(getOrderFoodErrorMessage(error, '场景配置加载失败'))
  } finally {
    detailLoading.value = false
  }
}

const buildMappings = () =>
  Object.fromEntries(
    editorForm.value.mappingRows.map((row) => [
      row.businessField,
      row.wechatField.trim()
    ])
  )

const handleConflict = async (error, action) => {
  if (!isOrderFoodConflict(error)) return false
  ElMessage.warning(`${action}失败：场景配置已变化，请刷新后重试`)
  editorVisible.value = false
  await loadScenes()
  return true
}

const submitEditor = async () => {
  const valid = await editorFormRef.value?.validate().catch(() => false)
  if (!valid) return
  editorLoading.value = true
  try {
    const payload = {
      wechatTemplateId: editorForm.value.wechatTemplateId,
      fieldMappings: buildMappings()
    }
    if (editorForm.value.configured) {
      payload.expectedVersion = editorForm.value.expectedVersion
    }
    unwrapOrderFoodResponse(
      await configureSubscribeScene(editorForm.value.scene, payload),
      '场景配置保存失败'
    )
    ElMessage.success(
      editorForm.value.configured
        ? '场景配置已保存，启用状态未改变'
        : '场景配置已保存，当前保持停用'
    )
    editorVisible.value = false
    await loadScenes()
  } catch (error) {
    if (!(await handleConflict(error, '保存'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '场景配置保存失败'))
    }
  } finally {
    editorLoading.value = false
  }
}

const reasonValidator = (value) => {
  const length = value?.trim().length || 0
  return length >= 4 && length <= 200 ? true : '请输入 4-200 字操作原因'
}
const requestReason = async (title, message) => {
  const result = await ElMessageBox.prompt(message, title, {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputPlaceholder: '请输入 4-200 字操作原因',
    inputValidator: reasonValidator
  })
  return result.value.trim()
}

const changeStatus = async (scene) => {
  const enabled = !scene.enabled
  try {
    const reason = await requestReason(
      enabled ? '启用订阅场景' : '停用订阅场景',
      enabled
        ? `确认启用“${scene.name}”吗？启用后符合条件的业务事件将创建微信发送任务。`
        : `确认停用“${scene.name}”吗？停用后不再创建新的微信发送任务。`
    )
    unwrapOrderFoodResponse(
      await updateSubscribeSceneStatus(scene.scene, {
        enabled,
        reason,
        expectedVersion: scene.version
      }),
      enabled ? '场景启用失败' : '场景停用失败'
    )
    ElMessage.success(enabled ? '场景已启用' : '场景已停用')
    await loadScenes()
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (!(await handleConflict(error, '切换状态'))) {
      ElMessage.error(getOrderFoodErrorMessage(error, '场景状态更新失败'))
    }
  }
}

const openLogs = (scene) => {
  router.push({
    name: 'OrderFoodSubscribeLogs',
    query: {
      scene: scene.scene,
      ...(scene.templateId ? { templateId: scene.templateId } : {})
    }
  })
}
const formatDateTime = (value) => (value ? formatDate(value) : '尚未配置')

loadScenes()
</script>

<style scoped>
.subscribe-scenes-page {
  --scene-ink: #172033;
  --scene-muted: #68748a;
  --scene-line: #e3e8ef;
  --scene-soft: #f5f7fa;
  --scene-blue: #2867e8;
  --scene-green: #16866f;
  min-height: calc(100vh - 128px);
  padding-top: 16px;
}

.scene-hero {
  position: relative;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 32px;
  overflow: hidden;
  margin-bottom: 18px;
  padding: 30px 34px;
  border: 1px solid #dfe6ef;
  border-radius: 14px;
  background:
    linear-gradient(120deg, rgb(255 255 255 / 98%), rgb(246 249 253 / 96%)),
    repeating-linear-gradient(90deg, transparent 0 62px, #e9edf3 63px 64px);
  box-shadow: 0 12px 34px rgb(25 45 76 / 7%);
}

.scene-hero::after {
  position: absolute;
  right: -34px;
  bottom: -76px;
  width: 230px;
  height: 230px;
  border: 42px solid rgb(40 103 232 / 7%);
  border-radius: 50%;
  content: '';
}

.hero-copy,
.hero-summary {
  position: relative;
  z-index: 1;
}

.hero-eyebrow,
.scene-code {
  color: var(--scene-blue);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.16em;
}

.hero-copy h1 {
  margin: 8px 0 8px;
  color: var(--scene-ink);
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.hero-copy p {
  max-width: 680px;
  margin: 0;
  color: var(--scene-muted);
  font-size: 14px;
  line-height: 1.8;
}

.hero-summary {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  gap: 22px;
  padding: 15px 20px;
  border: 1px solid rgb(255 255 255 / 72%);
  border-radius: 12px;
  background: rgb(255 255 255 / 70%);
  box-shadow: 0 8px 24px rgb(41 59 88 / 8%);
  backdrop-filter: blur(10px);
}

.hero-summary div {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 58px;
}

.hero-summary strong {
  color: var(--scene-ink);
  font-size: 22px;
  line-height: 1.1;
}

.hero-summary span {
  margin-top: 5px;
  color: var(--scene-muted);
  font-size: 12px;
}

.hero-summary i {
  width: 1px;
  height: 28px;
  background: var(--scene-line);
}

.scene-alert {
  margin-bottom: 18px;
}

.scene-grid {
  min-height: 240px;
}

.scene-card {
  position: relative;
  overflow: hidden;
  padding: 28px;
  border: 1px solid var(--scene-line);
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 10px 28px rgb(35 48 70 / 6%);
  transition: border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.scene-card::before {
  position: absolute;
  top: 0;
  left: 0;
  width: 5px;
  height: 100%;
  background: #aeb8c8;
  content: '';
}

.scene-card.is-enabled::before {
  background: var(--scene-green);
}

.scene-card.is-focused {
  border-color: rgb(40 103 232 / 42%);
  box-shadow: 0 14px 34px rgb(40 103 232 / 11%);
}

.scene-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 16px 38px rgb(35 48 70 / 9%);
}

.scene-card__header,
.scene-card__footer,
.scene-identity,
.binding-panel__heading,
.dialog-footer-note,
.mapping-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.scene-identity {
  justify-content: flex-start;
}

.editor-context__mark {
  display: grid;
  width: 46px;
  height: 46px;
  flex: 0 0 46px;
  place-items: center;
  border: 1px solid #d8e2f5;
  border-radius: 11px;
  color: var(--scene-blue);
  background: #f1f5fd;
  font-family: STKaiti, KaiTi, serif;
  font-size: 22px;
  font-weight: 700;
}

.scene-identity h2 {
  margin: 4px 0 0;
  color: var(--scene-ink);
  font-size: 20px;
}

.status-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  margin-right: 6px;
  border-radius: 50%;
  background: currentcolor;
  vertical-align: 1px;
}

.scene-purpose {
  margin: 20px 0;
  color: #48556a;
  font-size: 14px;
}

.rule-strip {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.rule-item {
  display: flex;
  align-items: center;
  gap: 13px;
  padding: 14px 16px;
  border: 1px solid #e9edf2;
  border-radius: 10px;
  background: #fafbfd;
}

.rule-index {
  color: #a6b0c0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.rule-item div {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.rule-item small,
.binding-panel small,
.mapping-summary > span {
  color: #8a95a6;
  font-size: 11px;
}

.rule-item strong,
.binding-panel strong {
  color: #334056;
  font-size: 13px;
  font-weight: 600;
}

.binding-panel {
  padding: 20px;
  border: 1px solid #dfe5ee;
  border-radius: 12px;
  background: linear-gradient(135deg, #f8fafc 0%, #f3f6fa 100%);
}

.binding-panel__heading > div {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.version-badge {
  padding: 4px 8px;
  border: 1px solid #dce3ec;
  border-radius: 999px;
  color: #6c778a;
  background: #fff;
  font-size: 11px;
}

.field-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.field-chip {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 7px 9px;
  border: 1px solid #e0e6ef;
  border-radius: 7px;
  background: #fff;
}

.field-chip span {
  color: #425067;
  font-size: 12px;
}

.field-chip code,
.business-field code {
  color: var(--scene-blue);
  font-size: 11px;
}

.mapping-summary,
.empty-binding {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed #d7dee8;
}

.mapping-summary p,
.empty-binding {
  margin-bottom: 0;
  color: #6d7889;
  font-size: 12px;
  line-height: 1.7;
}

.mapping-summary p {
  margin-top: 5px;
}

.scene-card__footer {
  align-items: flex-end;
  margin-top: 22px;
}

.scene-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  color: #8a94a5;
  font-size: 11px;
}

.scene-metrics span {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.scene-metrics b {
  max-width: 150px;
  overflow: hidden;
  color: #4c586c;
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scene-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.editor-context {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) auto;
  align-items: center;
  gap: 14px;
  margin-bottom: 24px;
  padding: 16px;
  border: 1px solid #e1e7f0;
  border-radius: 11px;
  background: #f7f9fc;
}

.editor-context span {
  color: var(--scene-blue);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.12em;
}

.editor-context h3 {
  margin: 3px 0;
  color: var(--scene-ink);
  font-size: 16px;
}

.editor-context p {
  margin: 0;
  color: #7a8596;
  font-size: 12px;
}

.scene-editor-form :deep(.el-form-item) {
  margin-bottom: 24px;
}

.form-help {
  margin-top: 7px;
  color: #8a95a6;
  font-size: 12px;
  line-height: 1.6;
}

.mapping-label {
  width: 100%;
}

.mapping-label small {
  color: #8a95a6;
  font-weight: 400;
}

.mapping-editor {
  width: 100%;
  overflow: hidden;
  border: 1px solid #dfe5ed;
  border-radius: 10px;
}

.mapping-editor__head {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) minmax(210px, 1fr);
  gap: 46px;
  padding: 10px 14px;
  color: #8a95a6;
  background: #f6f8fb;
  font-size: 11px;
}

.mapping-row {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) 22px minmax(210px, 1fr);
  align-items: center;
  gap: 12px;
  padding: 13px 14px;
  border-top: 1px solid #edf0f4;
}

.business-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.business-field > div {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.business-field strong {
  color: #3c485d;
  font-size: 13px;
}

.business-field small {
  max-width: 135px;
  overflow: hidden;
  color: #9aa3b1;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mapping-arrow {
  color: #a4afbf;
  text-align: center;
}

.dialog-footer-note > span {
  color: #8b95a5;
  font-size: 12px;
}

:global(.scene-editor-dialog) {
  display: flex;
  max-height: 88vh;
  flex-direction: column;
  margin-top: 6vh;
}

:global(.scene-editor-dialog .el-dialog__body) {
  overflow-y: auto;
}

@media (max-width: 820px) {
  .scene-hero,
  .scene-card__footer {
    align-items: stretch;
    flex-direction: column;
  }

  .hero-summary {
    align-self: flex-start;
  }

  .rule-strip {
    grid-template-columns: 1fr;
  }

  .scene-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 640px) {
  .scene-hero,
  .scene-card {
    padding: 20px;
  }

  .scene-card__header,
  .dialog-footer-note,
  .mapping-label {
    align-items: flex-start;
    flex-direction: column;
  }

  .editor-context {
    grid-template-columns: 48px 1fr;
  }

  .editor-context .el-tag {
    grid-column: 1 / -1;
    justify-self: flex-start;
  }

  .mapping-editor__head {
    display: none;
  }

  .mapping-row {
    grid-template-columns: 1fr;
  }

  .mapping-arrow {
    text-align: left;
    transform: rotate(90deg);
  }
}
</style>