<template>
  <el-dialog
    v-model="visible"
    title="调整积分"
    width="520px"
    destroy-on-close
    :close-on-click-modal="!actionLoading"
    :close-on-press-escape="!actionLoading"
    @closed="resetState"
  >
    <div v-loading="userLoading" class="min-h-56">
      <el-alert
        v-if="!canReadUsers"
        title="当前角色没有用户读取权限，无法加载调整对象。"
        type="error"
        show-icon
        :closable="false"
      />
      <el-alert
        v-else-if="userError"
        :title="userError"
        type="error"
        show-icon
        :closable="false"
      >
        <template #default>
          <el-button link type="primary" @click="loadSelectedUser(form.userId)">
            重新加载
          </el-button>
        </template>
      </el-alert>

      <el-form
        v-else
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="84px"
      >
        <el-form-item label="目标用户" prop="userId">
          <div
            v-if="hasPresetUser && selectedUser"
            class="flex min-w-0 flex-1 items-center gap-3 rounded-md border border-gray-200 px-3 py-2"
          >
            <el-avatar
              :size="34"
              :src="selectedUser.avatarUrl || undefined"
              class="shrink-0"
            >
              {{ selectedUser.nickname?.slice(0, 1) || '用' }}
            </el-avatar>
            <div class="min-w-0 flex-1">
              <div class="truncate font-medium">
                {{ selectedUser.nickname || '未命名用户' }}
              </div>
              <div class="truncate text-xs text-gray-400">
                {{ selectedUser.id }}
              </div>
            </div>
            <el-tag effect="plain">{{ selectedUser.points }} 积分</el-tag>
          </div>
          <el-select
            v-else
            v-model="form.userId"
            filterable
            remote
            reserve-keyword
            clearable
            :remote-method="searchUsers"
            :loading="userSearchLoading"
            placeholder="输入用户 ID 或昵称搜索"
            class="w-full"
            @change="handleUserChange"
          >
            <el-option
              v-for="userOption in userOptions"
              :key="userOption.id"
              :label="`${userOption.nickname || '未命名用户'} · ${userOption.id}`"
              :value="userOption.id"
            >
              <div class="flex items-center justify-between gap-4">
                <span>{{ userOption.nickname || '未命名用户' }}</span>
                <span class="text-xs text-gray-400">
                  {{ userOption.id }} · {{ userOption.points }} 积分
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
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="请输入 4-200 字原因，内容将写入流水和审计"
          />
        </el-form-item>

        <div
          v-if="selectedUser"
          class="mb-4 ml-[84px] flex items-center justify-between rounded-md bg-gray-50 px-4 py-3 text-sm"
        >
          <span class="text-gray-500">余额预估</span>
          <div class="flex items-center gap-3">
            <span>{{ selectedUser.points }}</span>
            <span class="text-gray-400">→</span>
            <strong :class="estimatedBalance < 0 ? 'text-red-500' : 'text-blue-600'">
              {{ estimatedBalance }}
            </strong>
          </div>
        </div>

        <el-alert
          v-if="debitExceedsBalance"
          :title="`当前最多可扣减 ${selectedUser.points} 积分`"
          type="error"
          show-icon
          :closable="false"
        />
      </el-form>
    </div>

    <template #footer>
      <el-button :disabled="actionLoading" @click="visible = false">
        取消
      </el-button>
      <el-button
        :type="form.direction === 'debit' ? 'danger' : 'primary'"
        :loading="actionLoading"
        :disabled="
          !canAdjust ||
          !selectedUser ||
          Boolean(userError) ||
          debitExceedsBalance
        "
        @click="submitAdjustment"
      >
        {{ form.direction === 'credit' ? '确认增加积分' : '确认扣减积分' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  createPointAdjustment,
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
  name: 'PointAdjustmentDialog'
})

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  user: {
    type: Object,
    default: null
  },
  userId: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'success'])
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const btnAuth = useBtnAuth()
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canAdjust = computed(() =>
  Boolean(btnAuth['orderfood:points:adjust'] || btnAuth.pointsAdjust)
)
const presetUserId = computed(() =>
  String(props.user?.id || props.userId || '')
)
const hasPresetUser = computed(() => Boolean(presetUserId.value))

const createDefaultForm = () => ({
  userId: '',
  direction: 'credit',
  amount: 1,
  reason: ''
})
const formRef = ref(null)
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

const selectedUser = ref(null)
const userOptions = ref([])
const userLoading = ref(false)
const userSearchLoading = ref(false)
const userError = ref('')
const actionLoading = ref(false)

const estimatedBalance = computed(() => {
  if (!selectedUser.value) return 0
  const amount = Number(form.value.amount || 0)
  return (
    Number(selectedUser.value.points || 0) +
    (form.value.direction === 'credit' ? amount : -amount)
  )
})
const debitExceedsBalance = computed(
  () =>
    Boolean(selectedUser.value) &&
    form.value.direction === 'debit' &&
    estimatedBalance.value < 0
)

const searchUsers = async (keyword = '') => {
  if (!canReadUsers.value) return
  userSearchLoading.value = true
  try {
    const normalizedKeyword = String(keyword || '').trim()
    const data = unwrapOrderFoodResponse(
      await getOrderFoodUserList({
        page: 1,
        pageSize: 20,
        keyword: normalizedKeyword || undefined,
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
    userSearchLoading.value = false
  }
}

const loadSelectedUser = async (userId) => {
  if (!userId || !canReadUsers.value) {
    selectedUser.value = null
    return
  }
  userLoading.value = true
  userError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodUserDetail(userId),
      '用户信息加载失败'
    )
    selectedUser.value = data
    if (!userOptions.value.some((item) => item.id === data.id)) {
      userOptions.value.unshift(data)
    }
  } catch (error) {
    selectedUser.value = null
    userError.value = getOrderFoodErrorMessage(error, '用户信息加载失败')
  } finally {
    userLoading.value = false
  }
}

const handleUserChange = (userId) => {
  userError.value = ''
  selectedUser.value = null
  if (userId) loadSelectedUser(userId)
}

const resetState = () => {
  form.value = createDefaultForm()
  selectedUser.value = null
  userOptions.value = []
  userError.value = ''
  userLoading.value = false
  userSearchLoading.value = false
  actionLoading.value = false
  formRef.value?.clearValidate()
}

const initialize = async () => {
  resetState()
  if (!canReadUsers.value) return
  if (presetUserId.value) {
    form.value.userId = presetUserId.value
    selectedUser.value = props.user ? { ...props.user } : null
    await loadSelectedUser(presetUserId.value)
    return
  }
  await searchUsers()
}

const submitAdjustment = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid || !selectedUser.value || debitExceedsBalance.value) return

  actionLoading.value = true
  try {
    const preview = unwrapOrderFoodResponse(
      await previewPointAdjustment({
        userId: form.value.userId,
        direction: form.value.direction,
        amount: form.value.amount,
        reason: form.value.reason
      }),
      '积分调整预览失败'
    )
    const action = preview.direction === 'credit' ? '增加' : '扣减'
    await ElMessageBox.confirm(
      `确认给“${preview.user?.nickname || preview.user?.id}”${action} ${
        preview.amount
      } 积分吗？余额将从 ${preview.balanceBefore} 变为 ${
        preview.balanceAfter
      }。`,
      '确认积分调整',
      {
        type: 'warning',
        confirmButtonText: `确认${action}`,
        cancelButtonText: '取消'
      }
    )
    const result = unwrapOrderFoodResponse(
      await createPointAdjustment({
        previewToken: preview.previewToken,
        userId: form.value.userId,
        direction: form.value.direction,
        amount: form.value.amount,
        reason: form.value.reason,
        expectedUserVersion: preview.expectedUserVersion
      }),
      '积分调整失败'
    )
    ElMessage.success(`积分调整完成，当前余额 ${result.balanceAfter}`)
    emit('success', {
      ...result,
      userId: form.value.userId
    })
    visible.value = false
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('用户积分已变化，请确认最新余额后重试')
      await loadSelectedUser(form.value.userId)
      return
    }
    ElMessage.error(getOrderFoodErrorMessage(error, '积分调整失败'))
  } finally {
    actionLoading.value = false
  }
}

watch(
  () => props.modelValue,
  (value) => {
    if (value) initialize()
  }
)
</script>
