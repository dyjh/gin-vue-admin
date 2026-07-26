<template>
  <div class="order-food-page" v-loading="loading">
    <div class="page-header">
      <div>
        <h2 class="page-title">积分规则</h2>
        <p class="page-subtitle">
          配置每日首次做菜打卡奖励。保存后立即对新的打卡生效，既有积分流水不会重新计算。
        </p>
      </div>
      <div class="page-actions">
        <el-button @click="loadConfig">刷新</el-button>
        <el-button v-if="canUpdate" type="primary" @click="openEditor">
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

    <el-card v-if="config" shadow="never" class="section-card">
      <template #header>
        <div class="section-header">
          <strong>当前配置</strong>
          <el-tag type="success">已生效</el-tag>
        </div>
      </template>
      <div class="reward-value">
        <strong>{{ config.dailyCheckinReward }}</strong>
        <span>积分 / 每日首次打卡</span>
      </div>
      <el-descriptions :column="2" border class="config-meta">
        <el-descriptions-item label="配置版本">
          v{{ config.version }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatDateTime(config.updatedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="更新人">
          {{ administratorLabel(config.updatedBy) }}
        </el-descriptions-item>
        <el-descriptions-item label="修改原因">
          {{ config.reason || '—' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

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
.page-header,
.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.page-title {
  margin: 0;
  color: #1f2937;
  font-size: 22px;
}

.page-subtitle {
  margin: 6px 0 0;
  color: #6b7280;
  font-size: 13px;
}

.page-actions {
  display: flex;
  gap: 8px;
}

.section-card {
  margin-top: 18px;
}

.reward-value {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin: 12px 0 24px;
  color: #4b5563;
}

.reward-value strong {
  color: #f59e0b;
  font-size: 44px;
}

.config-meta {
  max-width: 900px;
}

.dialog-alert {
  margin-bottom: 18px;
}
</style>
