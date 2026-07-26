<template>
  <div>
    <div class="gva-table-box">
      <div class="mb-5">
        <div class="text-lg font-medium">积分调整</div>
        <div class="mt-1 text-sm text-gray-500">
          人工增加或扣减积分需要先生成余额预览，再进行二次确认。
        </div>
      </div>

      <el-alert
        title="预览令牌有效 10 分钟。预览后修改用户、方向、数量或原因，必须重新预览。"
        type="warning"
        show-icon
        class="mb-5"
      />
      <el-alert
        v-if="!canReadUsers"
        title="当前角色没有用户读取权限，无法搜索并绑定调整对象。"
        type="error"
        show-icon
        :closable="false"
        class="mb-5"
      />

      <div class="adjustment-layout">
        <section class="adjustment-panel">
          <div class="panel-title">调整信息</div>
          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-width="110px"
          >
            <el-form-item label="选择用户" prop="userId">
              <el-select
                v-model="form.userId"
                filterable
                remote
                reserve-keyword
                clearable
                :remote-method="searchUsers"
                :loading="userLoading"
                :disabled="!canReadUsers"
                placeholder="输入用户 ID 或昵称搜索"
                class="w-full"
                @change="handleUserChange"
              >
                <el-option
                  v-for="user in userOptions"
                  :key="user.id"
                  :label="`${user.nickname || '未命名用户'} · ${user.id}`"
                  :value="user.id"
                >
                  <div class="flex items-center justify-between gap-4">
                    <span>{{ user.nickname || '未命名用户' }}</span>
                    <span class="text-xs text-gray-400">
                      {{ user.id }} · {{ user.points }} 积分
                    </span>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="调整方向" prop="direction">
              <el-radio-group v-model="form.direction">
                <el-radio-button value="credit">增加积分</el-radio-button>
                <el-radio-button value="debit">扣减积分</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="积分数" prop="amount">
              <el-input-number
                v-model="form.amount"
                :min="1"
                :max="100000"
                :step="1"
                class="w-full"
              />
            </el-form-item>
            <el-form-item label="调整原因" prop="reason">
              <el-input
                v-model.trim="form.reason"
                type="textarea"
                :rows="4"
                maxlength="200"
                show-word-limit
                placeholder="请输入 4-200 字原因，内容将写入流水和审计"
              />
            </el-form-item>
          </el-form>
          <el-alert
            v-if="debitExceedsBalance"
            :title="`当前最多可扣减 ${selectedUser.points} 积分，请降低扣减数量后重新预览。`"
            type="error"
            show-icon
            :closable="false"
            class="mb-4"
          />
          <div class="flex justify-end gap-2">
            <el-button @click="clearForm">清空重填</el-button>
            <el-button
              type="primary"
              :loading="previewLoading"
              :disabled="!canAdjust || !selectedUser || debitExceedsBalance"
              @click="previewAdjustment"
            >
              预览调整
            </el-button>
          </div>
        </section>

        <section class="adjustment-panel">
          <div class="panel-title">用户与余额</div>
          <el-empty v-if="!selectedUser" description="请先选择唯一用户" :image-size="80" />
          <template v-else>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="用户">
                {{ selectedUser.nickname || '未命名用户' }}
              </el-descriptions-item>
              <el-descriptions-item label="用户 ID">
                {{ selectedUser.id }}
              </el-descriptions-item>
              <el-descriptions-item label="账号状态">
                <el-tag :type="selectedUser.status === 'normal' ? 'success' : 'danger'">
                  {{ selectedUser.status === 'normal' ? '正常' : '已禁用' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="当前积分">
                {{ selectedUser.points }}
              </el-descriptions-item>
              <el-descriptions-item label="用户版本">
                {{ selectedUser.version }}
              </el-descriptions-item>
            </el-descriptions>
            <div class="mt-4 flex gap-2">
              <el-button v-if="canReadUsers" type="primary" plain @click="openUser">
                查看用户
              </el-button>
              <el-button v-if="canReadPoints" plain @click="openPointEntries">
                查看全部流水
              </el-button>
            </div>
          </template>
        </section>
      </div>

      <section v-if="preview" class="preview-panel">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="panel-title">调整预览</div>
            <div class="text-sm text-gray-500">
              有效期至 {{ formatDateTime(preview.expiresAt) }}
            </div>
          </div>
          <el-tag type="warning" effect="plain">等待确认</el-tag>
        </div>
        <div class="balance-flow">
          <div class="balance-card">
            <span>调整前</span>
            <strong>{{ preview.balanceBefore }}</strong>
          </div>
          <div class="balance-change">
            <span>{{ preview.direction === 'credit' ? '+' : '-' }}{{ preview.amount }}</span>
            <small>{{ preview.direction === 'credit' ? '人工增加' : '人工扣减' }}</small>
          </div>
          <div class="balance-card result">
            <span>调整后</span>
            <strong>{{ preview.balanceAfter }}</strong>
          </div>
        </div>
        <el-alert
          v-if="preview.direction === 'debit'"
          :title="`本次最多可扣减 ${preview.balanceBefore} 积分，确认后余额为 ${preview.balanceAfter}。`"
          type="warning"
          show-icon
          class="mt-4"
        />
        <div class="mt-5 flex justify-end">
          <el-button
            type="danger"
            :loading="submitLoading"
            :disabled="!canAdjust"
            @click="confirmAdjustment"
          >
            {{
              preview.direction === 'credit'
                ? `确认增加 ${preview.amount} 积分`
                : `确认扣减 ${preview.amount} 积分`
            }}
          </el-button>
        </div>
      </section>

      <section v-if="lastResult" class="result-panel">
        <div>
          <div class="font-medium text-green-600">积分调整已完成</div>
          <div class="mt-1 text-sm text-gray-500">
            余额从 {{ lastResult.balanceBefore }} 变为 {{ lastResult.balanceAfter }}，
            流水 ID：{{ lastResult.pointEntry?.id }}
          </div>
        </div>
        <el-button v-if="canReadPoints" type="primary" @click="openPointEntries">
          查看积分流水
        </el-button>
      </section>

      <section v-if="selectedUser && canReadPoints" class="mt-6">
        <div class="panel-title">最近 5 条积分流水</div>
        <el-table v-loading="recentLoading" :data="recentEntries" empty-text="暂无积分流水">
          <el-table-column label="标题" prop="title" min-width="180" />
          <el-table-column label="类型" width="100">
            <template #default="{ row }">{{ typeLabel(row.type) }}</template>
          </el-table-column>
          <el-table-column label="积分变化" width="110" align="right">
            <template #default="{ row }">
              {{ row.amount >= 0 ? '+' : '' }}{{ row.amount }}
            </template>
          </el-table-column>
          <el-table-column label="变化后余额" prop="balanceAfter" width="110" />
          <el-table-column label="场景" prop="scene" min-width="150" />
          <el-table-column label="创建时间" min-width="170">
            <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
          </el-table-column>
        </el-table>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import {
  createPointAdjustment,
  getPointEntryList,
  previewPointAdjustment
} from '@/api/orderfood/points'
import {
  getOrderFoodUserDetail,
  getOrderFoodUserList
} from '@/api/orderfood/user'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodPointAdjustments'
})

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canReadPoints = computed(() => Boolean(btnAuth['orderfood:points:read']))
const canAdjust = computed(() =>
  Boolean(btnAuth['orderfood:points:adjust'] || btnAuth.pointsAdjust)
)
const formRef = ref(null)
const createDefaultForm = () => ({
  userId: '',
  direction: 'credit',
  amount: 1,
  reason: ''
})
const form = ref(createDefaultForm())
const rules = {
  userId: [{ required: true, message: '请选择用户', trigger: 'change' }],
  direction: [{ required: true, message: '请选择调整方向', trigger: 'change' }],
  amount: [
    { required: true, message: '请输入积分数', trigger: 'change' },
    { type: 'number', min: 1, max: 100000, message: '积分数应为 1-100000' }
  ],
  reason: [
    { required: true, message: '请输入调整原因', trigger: 'blur' },
    { min: 4, max: 200, message: '原因长度应为 4-200 字', trigger: 'blur' }
  ]
}

const userOptions = ref([])
const userLoading = ref(false)
const selectedUser = ref(null)
const searchUsers = async (keyword) => {
  if (!canReadUsers.value) return
  userLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodUserList({
        page: 1,
        pageSize: 20,
        keyword: keyword?.trim() || undefined,
        sortBy: 'createdAt',
        sortOrder: 'desc'
      }),
      '用户搜索失败'
    )
    userOptions.value = data?.list || []
  } catch (error) {
    userOptions.value = []
    ElMessage.error(getOrderFoodErrorMessage(error, '用户搜索失败'))
  } finally {
    userLoading.value = false
  }
}
const loadSelectedUser = async (userId) => {
  if (!userId) {
    selectedUser.value = null
    recentEntries.value = []
    return
  }
  if (!canReadUsers.value) return
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodUserDetail(userId),
      '用户详情加载失败'
    )
    selectedUser.value = data
    if (!userOptions.value.some((user) => user.id === data.id)) {
      userOptions.value.unshift(data)
    }
    await loadRecentEntries()
  } catch (error) {
    selectedUser.value = null
    ElMessage.error(getOrderFoodErrorMessage(error, '用户详情加载失败'))
  }
}
const handleUserChange = (userId) => {
  selectedUser.value = null
  recentEntries.value = []
  loadSelectedUser(userId)
}

const recentEntries = ref([])
const recentLoading = ref(false)
const debitExceedsBalance = computed(
  () =>
    Boolean(selectedUser.value) &&
    form.value.direction === 'debit' &&
    Number(form.value.amount || 0) > Number(selectedUser.value.points || 0)
)
const loadRecentEntries = async () => {
  if (!canReadPoints.value || !form.value.userId) return
  recentLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getPointEntryList({
        page: 1,
        pageSize: 20,
        userId: form.value.userId,
        sortBy: 'createdAt',
        sortOrder: 'desc'
      }),
      '最近积分流水加载失败'
    )
    recentEntries.value = (data?.list || []).slice(0, 5)
  } catch (error) {
    recentEntries.value = []
    ElMessage.warning(getOrderFoodErrorMessage(error, '最近积分流水加载失败'))
  } finally {
    recentLoading.value = false
  }
}

const preview = ref(null)
const lastResult = ref(null)
const previewLoading = ref(false)
const submitLoading = ref(false)
watch(
  () => [form.value.userId, form.value.direction, form.value.amount, form.value.reason],
  () => {
    preview.value = null
    lastResult.value = null
  }
)

const previewAdjustment = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!selectedUser.value || selectedUser.value.id !== form.value.userId) {
    ElMessage.warning('请等待用户信息加载完成后再预览')
    return
  }
  if (debitExceedsBalance.value) return
  previewLoading.value = true
  lastResult.value = null
  try {
    preview.value = unwrapOrderFoodResponse(
      await previewPointAdjustment({
        userId: form.value.userId,
        direction: form.value.direction,
        amount: form.value.amount,
        reason: form.value.reason
      }),
      '积分调整预览失败'
    )
  } catch (error) {
    preview.value = null
    ElMessage.error(getOrderFoodErrorMessage(error, '积分调整预览失败'))
  } finally {
    previewLoading.value = false
  }
}

const confirmAdjustment = async () => {
  if (!preview.value) return
  const action = preview.value.direction === 'credit' ? '增加' : '扣减'
  try {
    await ElMessageBox.confirm(
      `确认给“${preview.value.user?.nickname || preview.value.user?.id}”${action} ${
        preview.value.amount
      } 积分吗？余额将从 ${preview.value.balanceBefore} 变为 ${
        preview.value.balanceAfter
      }。`,
      '二次确认积分调整',
      {
        type: 'warning',
        confirmButtonText: `确认${action} ${preview.value.amount} 积分`,
        cancelButtonText: '取消'
      }
    )
    submitLoading.value = true
    const result = unwrapOrderFoodResponse(
      await createPointAdjustment({
        previewToken: preview.value.previewToken,
        userId: form.value.userId,
        direction: form.value.direction,
        amount: form.value.amount,
        reason: form.value.reason,
        expectedUserVersion: preview.value.expectedUserVersion
      }),
      '积分调整失败'
    )
    ElMessage.success(`积分调整完成，当前余额 ${result.balanceAfter}`)
    lastResult.value = result
    preview.value = null
    await loadSelectedUser(form.value.userId)
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('用户余额或版本已变化，请重新预览')
      preview.value = null
      await loadSelectedUser(form.value.userId)
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '积分调整失败'))
    }
  } finally {
    submitLoading.value = false
  }
}

const clearForm = () => {
  form.value = createDefaultForm()
  selectedUser.value = null
  recentEntries.value = []
  preview.value = null
  lastResult.value = null
  formRef.value?.clearValidate()
}
const openUser = () => {
  router.push({ name: 'OrderFoodUsers', query: { userId: form.value.userId } })
}
const openPointEntries = () => {
  router.push({
    name: 'OrderFoodPointEntries',
    query: { userId: form.value.userId }
  })
}
const typeLabel = (type) =>
  ({ earned: '获得', spent: '消耗', refund: '退款', adjustment: '人工调整' })[
    type
  ] || type
const formatDateTime = (value) => (value ? formatDate(value) : '—')

if (canReadUsers.value) {
  searchUsers('')
}
if (canReadUsers.value && route.query.userId) {
  form.value.userId = String(route.query.userId)
  loadSelectedUser(form.value.userId)
}
</script>

<style scoped>
.adjustment-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(300px, 0.8fr);
  gap: 20px;
}

.adjustment-panel,
.preview-panel,
.result-panel {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 20px;
}

.preview-panel {
  margin-top: 24px;
}

.result-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 24px;
  border-color: var(--el-color-success-light-5);
  background: var(--el-color-success-light-9);
}

.panel-title {
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
}

.balance-flow {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 20px;
  margin-top: 24px;
}

.balance-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  border-radius: 8px;
  background: var(--el-fill-color-light);
  padding: 18px;
  text-align: center;
}

.balance-card strong {
  font-size: 28px;
}

.balance-card.result {
  color: var(--el-color-primary);
}

.balance-change {
  display: flex;
  min-width: 110px;
  flex-direction: column;
  gap: 4px;
  text-align: center;
}

.balance-change span {
  font-size: 20px;
  font-weight: 600;
}

.balance-change small {
  color: var(--el-text-color-secondary);
}

@media (max-width: 900px) {
  .adjustment-layout {
    grid-template-columns: 1fr;
  }
}
</style>
