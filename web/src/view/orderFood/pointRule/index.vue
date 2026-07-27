<template>
  <div class="order-food-page" v-loading="loading">
    <div class="page-header">
      <div class="page-heading">
        <h2 class="page-title">积分规则</h2>
        <p class="page-subtitle">
          配置每日首次做菜打卡奖励。保存后立即对新的打卡生效，既有积分流水不会重新计算。
        </p>
      </div>
      <div class="page-actions">
        <el-button class="refresh-button" @click="loadConfig">刷新</el-button>
        <el-button
          v-if="canUpdate"
          class="edit-button"
          type="primary"
          @click="openEditor"
        >
          编辑配置
        </el-button>
      </div>
    </div>

    <el-alert
      v-if="pageError"
      type="error"
      :title="pageError"
      show-icon
      :closable="false"
      class="section-card"
    />

    <section v-if="config" class="rule-panel">
      <header class="rule-panel-header">
        <div class="rule-panel-heading">
          <span class="status-mark" aria-hidden="true">
            <span />
          </span>
          <div>
            <strong>当前积分规则</strong>
            <span>规则保存后实时应用于新的打卡</span>
          </div>
        </div>
        <span class="status-pill">
          <i aria-hidden="true" />
          已生效
        </span>
      </header>

      <div class="rule-panel-body">
        <article class="reward-card">
          <span class="reward-eyebrow">每日首次做菜打卡</span>
          <div class="reward-value">
            <strong>{{ config.dailyCheckinReward }}</strong>
            <div>
              <span>积分</span>
              <small>每位用户 / 每日一次</small>
            </div>
          </div>
          <p>用户当天首次完成做菜打卡后，系统自动发放奖励。</p>
          <span class="reward-scope">新配置从下一条符合条件的打卡起生效</span>
        </article>

        <div class="config-side">
          <div class="meta-intro">
            <span>配置信息</span>
            <p>记录当前版本及最近一次修改信息。</p>
          </div>
          <dl class="meta-grid">
            <div class="meta-item">
              <dt>配置版本</dt>
              <dd>v{{ config.version }}</dd>
            </div>
            <div class="meta-item">
              <dt>更新时间</dt>
              <dd>{{ formatDateTime(config.updatedAt) }}</dd>
            </div>
            <div class="meta-item">
              <dt>更新人</dt>
              <dd>{{ administratorLabel(config.updatedBy) }}</dd>
            </div>
            <div class="meta-item">
              <dt>修改原因</dt>
              <dd>{{ config.reason || '—' }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <footer class="rule-panel-footer">
        <span class="rule-note-icon" aria-hidden="true">
          <span />
        </span>
        <div>
          <strong>生效说明</strong>
          <span>仅影响保存后的新打卡，既有积分流水不会重新计算。</span>
        </div>
      </footer>
    </section>

    <el-dialog v-model="editorVisible" title="编辑积分规则" width="520px">
      <el-alert
        title="保存成功后立即生效"
        description="其他管理员先保存时，本次提交会被拒绝并重新加载最新配置。"
        type="info"
        show-icon
        :closable="false"
        class="dialog-alert"
      />
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="110px"
      >
        <el-form-item label="奖励积分" prop="dailyCheckinReward">
          <el-input-number
            v-model="form.dailyCheckinReward"
            :min="1"
            :max="100"
            :step="1"
          />
        </el-form-item>
        <el-form-item label="修改原因" prop="reason">
          <el-input
            v-model.trim="form.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="说明本次调整原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveConfig">
          保存并立即生效
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import { getPointRule, updatePointRule } from '@/api/orderfood/points'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodPointRules' })

const btnAuth = useBtnAuth()
const canUpdate = computed(() =>
  Boolean(btnAuth['orderfood:point-rule:update'] || btnAuth.pointRuleUpdate)
)
const config = ref(null)
const loading = ref(false)
const pageError = ref('')
const editorVisible = ref(false)
const saving = ref(false)
const formRef = ref()
const form = ref({
  dailyCheckinReward: 1,
  reason: ''
})
const rules = {
  dailyCheckinReward: [
    { required: true, message: '请输入奖励积分', trigger: 'change' },
    { type: 'number', min: 1, max: 100, message: '奖励积分必须为 1-100', trigger: 'change' }
  ],
  reason: [
    { required: true, message: '请输入修改原因', trigger: 'blur' },
    { min: 2, max: 200, message: '原因长度为 2-200 字', trigger: 'blur' }
  ]
}

const formatDateTime = (value) => (value ? formatDate(value) : '—')
const administratorLabel = (administrator) =>
  administrator?.nickname || administrator?.username || administrator?.id || '—'

const loadConfig = async () => {
  loading.value = true
  pageError.value = ''
  try {
    config.value = unwrapOrderFoodResponse(
      await getPointRule(),
      '加载积分规则失败'
    )
  } catch (error) {
    pageError.value = getOrderFoodErrorMessage(error, '加载积分规则失败')
  } finally {
    loading.value = false
  }
}

const openEditor = () => {
  form.value = {
    dailyCheckinReward: Number(config.value?.dailyCheckinReward || 1),
    reason: ''
  }
  editorVisible.value = true
}

const saveConfig = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid || !config.value) return
  saving.value = true
  try {
    config.value = unwrapOrderFoodResponse(
      await updatePointRule({
        dailyCheckinReward: form.value.dailyCheckinReward,
        reason: form.value.reason,
        expectedVersion: config.value.version
      }),
      '保存积分规则失败'
    )
    editorVisible.value = false
    ElMessage.success('积分规则已保存并生效')
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('配置已被其他管理员更新，已重新加载最新值')
      editorVisible.value = false
      await loadConfig()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存积分规则失败'))
    }
  } finally {
    saving.value = false
  }
}

loadConfig()
</script>

<style scoped lang="scss">
.order-food-page {
  --rule-ink: #263548;
  --rule-muted: #748197;
  --rule-blue: #3378e5;
  --rule-blue-deep: #245fbf;
  --rule-blue-soft: #f3f7fd;
  --rule-line: #e3eaf3;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 4px 2px 0;
}

.page-heading {
  min-width: 0;
}

.page-title {
  margin: 0;
  color: var(--rule-ink);
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.page-subtitle {
  max-width: 760px;
  margin: 7px 0 0;
  color: var(--rule-muted);
  font-size: 13px;
  line-height: 1.7;
}

.page-actions {
  display: flex;
  flex-shrink: 0;
  gap: 10px;
}

.page-actions :deep(.el-button) {
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

.section-card {
  margin-top: 18px;
}

.rule-panel {
  overflow: hidden;
  margin-top: 20px;
  border: 1px solid var(--rule-line);
  border-radius: 12px;
  background: var(--el-bg-color);
  box-shadow: 0 4px 18px rgb(38 53 72 / 5%);
}

.rule-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 17px 20px;
  border-bottom: 1px solid var(--rule-line);
}

.rule-panel-heading {
  display: flex;
  align-items: center;
  gap: 11px;
}

.rule-panel-heading > div {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.rule-panel-heading strong {
  color: var(--rule-ink);
  font-size: 15px;
}

.rule-panel-heading > div > span {
  color: var(--rule-muted);
  font-size: 12px;
}

.status-mark {
  display: block;
  width: 4px;
  height: 32px;
  border-radius: 2px;
  background: var(--rule-blue);
}

.status-mark > span {
  display: none;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 9px;
  border: 1px solid rgb(51 120 229 / 18%);
  border-radius: 6px;
  background: rgb(51 120 229 / 7%);
  color: var(--rule-blue-deep);
  font-size: 12px;
  font-weight: 600;
}

.status-pill i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--rule-blue);
}

.rule-panel-body {
  display: grid;
  grid-template-columns: minmax(280px, 0.85fr) minmax(420px, 1.55fr);
  gap: 18px;
  padding: 20px;
}

.reward-card {
  min-height: 184px;
  padding: 20px;
  border: 1px solid #dce8f8;
  border-radius: 10px;
  background: color-mix(in srgb, var(--el-bg-color) 82%, #eaf2ff 18%);
}

.reward-eyebrow {
  color: var(--rule-blue-deep);
  font-size: 12px;
  font-weight: 700;
}

.reward-value {
  display: flex;
  align-items: center;
  gap: 13px;
  margin: 14px 0 10px;
}

.reward-value strong {
  color: var(--rule-blue);
  font-size: 52px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  line-height: 1;
  letter-spacing: -0.05em;
}

.reward-value > div {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.reward-value span {
  color: var(--rule-ink);
  font-size: 16px;
  font-weight: 700;
}

.reward-value small {
  color: var(--rule-muted);
  font-size: 11px;
}

.reward-card p {
  max-width: 310px;
  margin: 0 0 14px;
  color: #66758b;
  font-size: 12px;
  line-height: 1.7;
}

.reward-scope {
  display: block;
  padding-top: 11px;
  border-top: 1px solid #dce8f8;
  color: #4e6f9d;
  font-size: 11px;
  font-weight: 600;
}

.config-side {
  min-width: 0;
}

.meta-intro {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.meta-intro > span {
  color: var(--rule-ink);
  font-size: 14px;
  font-weight: 700;
}

.meta-intro p {
  margin: 0;
  color: var(--rule-muted);
  font-size: 11px;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin: 0;
}

.meta-item {
  min-width: 0;
  min-height: 76px;
  padding: 14px 15px;
  border: 1px solid var(--rule-line);
  border-radius: 8px;
  background: color-mix(in srgb, var(--el-bg-color) 96%, #eef4fb 4%);
}

.meta-item dt {
  margin-bottom: 8px;
  color: var(--rule-muted);
  font-size: 11px;
}

.meta-item dd {
  overflow: hidden;
  margin: 0;
  color: var(--rule-ink);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-panel-footer {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 13px 20px;
  border-top: 1px solid var(--rule-line);
  background: color-mix(in srgb, var(--el-bg-color) 94%, #eef3f8 6%);
}

.rule-note-icon {
  display: grid;
  flex: 0 0 auto;
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: rgb(51 120 229 / 9%);
  place-items: center;
}

.rule-note-icon > span {
  width: 8px;
  height: 8px;
  border: 2px solid #5f7fa9;
  border-radius: 50%;
  box-shadow: inset 0 0 0 1px var(--el-bg-color);
}

.rule-panel-footer > div {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 5px 10px;
}

.rule-panel-footer strong {
  color: #3f5877;
  font-size: 12px;
}

.rule-panel-footer span {
  color: var(--rule-muted);
  font-size: 11px;
}

.dialog-alert {
  margin-bottom: 18px;
}

@media (max-width: 960px) {
  .rule-panel-body {
    grid-template-columns: 1fr;
  }

  .reward-card {
    min-height: auto;
  }
}

@media (max-width: 720px) {
  .page-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .rule-panel-header,
  .rule-panel-body {
    padding-right: 16px;
    padding-left: 16px;
  }

  .meta-intro {
    align-items: flex-start;
    flex-direction: column;
    gap: 5px;
  }

  .meta-grid {
    grid-template-columns: 1fr;
  }

  .rule-panel-footer {
    padding-right: 16px;
    padding-left: 16px;
  }
}
</style>
