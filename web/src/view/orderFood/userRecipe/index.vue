<template>
  <div class="order-food-list-page">
    <AdvancedSearchPanel>
      <el-alert
        v-if="lockedUserId"
        :title="`当前仅查看用户 ${lockedUserId} 的菜谱`"
        type="info"
        show-icon
        :closable="false"
        class="mb-4"
      />
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="关键词">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            placeholder="菜谱名称"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="用户 ID">
          <el-input
            v-model.trim="searchInfo.userId"
            clearable
            :disabled="Boolean(lockedUserId)"
            placeholder="用户 ID"
          />
        </el-form-item>
        <el-form-item label="用户昵称">
          <el-input
            v-model.trim="searchInfo.userKeyword"
            clearable
            placeholder="用户 ID 或昵称"
          />
        </el-form-item>
        <el-form-item label="菜品 ID">
          <el-input
            v-model.trim="searchInfo.dishId"
            clearable
            :disabled="Boolean(lockedDishId)"
            placeholder="所含菜品 ID"
          />
        </el-form-item>
        <el-form-item label="菜品数量">
          <div class="flex items-center gap-2">
            <el-input-number
              v-model="searchInfo.minDishCount"
              :min="0"
              :controls="false"
              placeholder="最少"
              class="w-24"
            />
            <span>至</span>
            <el-input-number
              v-model="searchInfo.maxDishCount"
              :min="0"
              :controls="false"
              placeholder="最多"
              class="w-24"
            />
          </div>
        </el-form-item>
        <el-form-item label="删除状态">
          <el-select v-model="searchInfo.deletionStatus" class="w-32">
            <el-option label="未删除" value="active" />
            <el-option label="已删除" value="deleted" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容状态">
          <el-select
            v-model="searchInfo.contentState"
            clearable
            placeholder="全部"
            class="w-40"
          >
            <el-option label="内容完整" value="ready" />
            <el-option label="空菜谱" value="empty" />
            <el-option label="含不可用菜品" value="contains_unavailable" />
          </el-select>
        </el-form-item>
        <el-form-item label="包含备注">
          <el-select
            v-model="searchInfo.hasNote"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="是" :value="true" />
            <el-option label="否" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="创建时间">
          <el-date-picker
            v-model="searchInfo.createdRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item label="更新时间">
          <el-date-picker
            v-model="searchInfo.updatedRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item label="排序">
          <div class="flex items-center gap-2">
            <el-select v-model="searchInfo.sortBy" class="w-32">
              <el-option label="创建时间" value="createdAt" />
              <el-option label="更新时间" value="updatedAt" />
              <el-option label="删除时间" value="deletedAt" />
              <el-option label="菜谱名称" value="name" />
              <el-option label="菜品数量" value="dishCount" />
            </el-select>
            <el-select v-model="searchInfo.sortOrder" class="w-24">
              <el-option label="倒序" value="desc" />
              <el-option label="正序" value="asc" />
            </el-select>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </AdvancedSearchPanel>

    <div class="gva-table-box">
      <div class="mb-4 text-sm text-gray-500">
        列表只展示菜谱摘要，不返回备注全文或有序菜品清单；本页不提供新增、编辑、恢复、调序、公开、导出或批量操作。
      </div>
      <el-alert
        v-if="listError"
        :title="listError"
        type="error"
        show-icon
        class="mb-4"
      >
        <template #default>
          <el-button link type="primary" @click="getTableData">重新加载</el-button>
        </template>
      </el-alert>

      <el-table
        v-loading="listLoading"
        :data="tableData"
        row-key="id"
        :empty-text="listError ? '加载失败' : '暂无用户菜谱'"
      >
        <el-table-column label="菜谱" min-width="240" fixed="left">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <el-image
                v-if="row.coverUrl"
                :src="row.coverUrl"
                fit="cover"
                class="h-12 w-12 shrink-0 rounded"
                :preview-src-list="[row.coverUrl]"
                preview-teleported
              />
              <div
                v-else
                class="flex h-12 w-12 shrink-0 items-center justify-center rounded bg-gray-100 text-xs text-gray-400"
              >
                受限
              </div>
              <div class="min-w-0">
                <div class="truncate font-medium">{{ row.name }}</div>
                <div class="truncate text-xs text-gray-400">{{ row.id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="所属用户" min-width="170">
          <template #default="{ row }">
            <div>{{ row.owner?.nickname || '未知用户' }}</div>
            <div class="text-xs text-gray-400">{{ row.owner?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="菜品数" width="95">
          <template #default="{ row }">
            <div>{{ row.dishCount }}</div>
            <div v-if="row.unavailableDishCount" class="text-xs text-red-500">
              不可用 {{ row.unavailableDishCount }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="内容状态" width="130">
          <template #default="{ row }">
            <el-tag :type="contentStateType(row.contentState)" size="small">
              {{ contentStateLabel(row.contentState) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="备注" width="90">
          <template #default="{ row }">{{ row.hasNote ? '有' : '无' }}</template>
        </el-table-column>
        <el-table-column label="删除状态" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.deletionStatus === 'deleted' ? 'danger' : 'success'"
              size="small"
            >
              {{ row.deletionStatus === 'deleted' ? '已软删除' : '未删除' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="165">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="165">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="canPrivateRead"
              link
              type="primary"
              @click="openPrivateDetail(row)"
            >
              私有详情
            </el-button>
            <el-button
              v-if="canReadUsers"
              link
              type="primary"
              @click="openUser(row.owner?.id)"
            >
              查看用户
            </el-button>
            <el-button
              v-if="canGovernanceExecute && row.deletionStatus !== 'deleted'"
              link
              type="danger"
              @click="openGovernanceDialog(row)"
            >
              违规处理
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="flex justify-end pt-4">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>

    <el-drawer
      v-model="detailVisible"
      size="820px"
      destroy-on-close
      @closed="resetDetail"
    >
      <template #header>
        <div>
          <div class="text-lg font-medium">
            {{ detail?.name || activeRecipe?.name || '菜谱详情' }}
          </div>
          <div class="mt-1 text-xs text-gray-400">{{ activeRecipe?.id }}</div>
        </div>
      </template>

      <div v-loading="detailLoading" class="min-h-64">
        <el-alert
          v-if="detailError"
          :title="detailError"
          type="error"
          show-icon
          :closable="false"
        >
          <template #default>
            <el-button link type="primary" @click="loadPrivateDetail">重新加载</el-button>
          </template>
        </el-alert>

        <template v-else-if="detail">
          <el-tabs>
            <el-tab-pane label="私有详情">
              <el-alert
            title="此处包含用户私有备注和有序菜品清单，访问行为已记录审计。页面只读，不能代用户编辑或调序。"
            type="warning"
            show-icon
            :closable="false"
            class="mb-4"
          />

          <div class="mb-4 flex gap-4">
            <el-image
              v-if="detail.coverUrl"
              :src="detail.coverUrl"
              fit="cover"
              class="h-32 w-32 shrink-0 rounded"
              :preview-src-list="[detail.coverUrl]"
              preview-teleported
            />
            <el-descriptions :column="2" border class="min-w-0 flex-1">
              <el-descriptions-item label="所属用户">
                {{ detail.owner?.nickname }}（{{ detail.owner?.id }}）
              </el-descriptions-item>
              <el-descriptions-item label="菜品数量">
                {{ detail.dishCount }}
              </el-descriptions-item>
              <el-descriptions-item label="内容状态">
                {{ contentStateLabel(detail.contentState) }}
              </el-descriptions-item>
              <el-descriptions-item label="删除状态">
                {{ detail.deletionStatus === 'deleted' ? '已软删除' : '未删除' }}
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">
                {{ formatDateTime(detail.createdAt) }}
              </el-descriptions-item>
              <el-descriptions-item label="更新时间">
                {{ formatDateTime(detail.updatedAt) }}
              </el-descriptions-item>
            </el-descriptions>
          </div>
          <div class="mb-4 flex flex-wrap gap-2">
            <el-button
              v-if="canReadUsers && detail.owner?.id"
              @click="openUser(detail.owner.id)"
            >
              查看用户
            </el-button>
            <el-button
              v-if="canGovernanceExecute && detail.deletionStatus !== 'deleted'"
              type="danger"
              @click="openGovernanceDialog(detail)"
            >
              违规处理
            </el-button>
            <el-button
              v-if="canReadUserDishes"
              @click="openRecipeDishes(detail.id)"
            >
              查看所含菜品
            </el-button>
          </div>

          <el-card shadow="never">
            <template #header>私有备注</template>
            <div class="whitespace-pre-wrap text-sm leading-6">
              {{ detail.note || '暂无备注' }}
            </div>
          </el-card>

          <el-card class="mt-4" shadow="never">
            <template #header>
              <div class="flex items-center justify-between">
                <span>有序菜品清单</span>
                <span class="text-xs font-normal text-gray-400">
                  仅展示，不支持编辑或调序
                </span>
              </div>
            </template>
            <el-empty
              v-if="!orderedDishes.length"
              description="该菜谱暂无菜品"
              :image-size="64"
            />
            <div v-else class="space-y-3">
              <div
                v-for="item in orderedDishes"
                :key="item.relationId"
                class="flex items-center gap-3 rounded border border-gray-200 p-3"
              >
                <el-tag round>{{ item.sortOrder }}</el-tag>
                <el-image
                  v-if="item.coverUrl"
                  :src="item.coverUrl"
                  fit="cover"
                  class="h-14 w-14 shrink-0 rounded"
                  :preview-src-list="[item.coverUrl]"
                  preview-teleported
                />
                <div
                  v-else
                  class="flex h-14 w-14 shrink-0 items-center justify-center rounded bg-gray-100 text-xs text-gray-400"
                >
                  无封面
                </div>
                <div class="min-w-0 flex-1">
                  <div class="truncate font-medium">{{ item.dishName }}</div>
                  <div class="mt-1 text-xs text-gray-400">
                    {{ sourceTypeLabel(item.sourceType) }} ·
                    {{ dishItemStatusLabel(item.dishStatus) }}
                  </div>
                  <div class="mt-1 text-xs text-gray-400">
                    关系 ID：{{ item.relationId }} · 加入时间：
                    {{ formatDateTime(item.addedAt) }}
                  </div>
                  <div v-if="item.unavailableReason" class="mt-1 text-xs text-red-500">
                    {{ item.unavailableReason }}
                  </div>
                </div>
                <el-button
                  v-if="canReadUserDishes"
                  link
                  type="primary"
                  @click="openDish(item.dishId)"
                >
                  查看菜品
                </el-button>
              </div>
            </div>
          </el-card>

              <el-card
                v-if="detail.deletionStatus === 'deleted'"
                class="mt-4"
                shadow="never"
              >
                <template #header>删除信息</template>
                <el-descriptions :column="2" border>
                  <el-descriptions-item label="删除时间">
                    {{ formatDateTime(detail.deletedAt) }}
                  </el-descriptions-item>
                  <el-descriptions-item label="删除人">
                    {{
                      detail.deletedBy?.nickname ||
                        detail.deletedBy?.username ||
                        detail.deletedBy?.id ||
                        '未记录'
                    }}
                  </el-descriptions-item>
                  <el-descriptions-item label="违规类型">
                    {{ detail.deletedViolationType || '未记录' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="删除原因">
                    {{ detail.deletedReason || '未记录' }}
                  </el-descriptions-item>
                </el-descriptions>
              </el-card>
              <div class="mt-3 flex flex-wrap items-center gap-3 text-xs text-gray-400">
                <span>违规处理记录 {{ detail.governanceRecordCount }} 条</span>
                <el-button
                  v-if="canReadGovernance && detail.governanceRecordCount > 0"
                  link
                  type="primary"
                  @click="openGovernanceRecords(detail.id)"
                >
                  查看违规处理记录
                </el-button>
                <span>操作审计 {{ detail.auditLogCount }} 条</span>
                <span>本次访问审计 ID：{{ detail.accessAuditId || '—' }}</span>
              </div>
            </el-tab-pane>
            <el-tab-pane v-if="canReadAudit" label="操作审计" lazy>
              <AuditPanel target-type="recipe" :target-id="detail.id" />
            </el-tab-pane>
          </el-tabs>
        </template>
      </div>
    </el-drawer>

    <el-dialog
      v-model="governanceVisible"
      title="菜谱违规处理"
      width="660px"
      destroy-on-close
      @closed="resetGovernance"
    >
      <el-alert
        title="该操作只软删除当前菜谱，不删除其中菜品，也不改变历史饭局或采购快照。执行前必须预览影响。"
        type="warning"
        show-icon
        :closable="false"
        class="mb-4"
      />
      <el-form
        ref="governanceFormRef"
        :model="governanceForm"
        :rules="governanceRules"
        label-width="94px"
      >
        <el-form-item label="目标菜谱">
          <span>{{ governanceTarget?.name }}（{{ governanceTarget?.id }}）</span>
        </el-form-item>
        <el-form-item label="处理动作">
          <el-tag type="danger">软删除菜谱</el-tag>
        </el-form-item>
        <el-form-item label="违规类型" prop="violationType">
          <el-input
            v-model.trim="governanceForm.violationType"
            maxlength="40"
            placeholder="例如：违规内容、侵权、垃圾信息"
            @input="clearGovernancePreview"
          />
        </el-form-item>
        <el-form-item label="严重程度" prop="severity">
          <el-radio-group
            v-model="governanceForm.severity"
            @change="clearGovernancePreview"
          >
            <el-radio-button value="normal">一般</el-radio-button>
            <el-radio-button value="serious">严重</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理原因" prop="reason">
          <el-input
            v-model.trim="governanceForm.reason"
            type="textarea"
            :rows="4"
            maxlength="300"
            show-word-limit
            placeholder="请输入 4-300 字处理原因"
            @input="clearGovernancePreview"
          />
        </el-form-item>
      </el-form>

      <el-card v-if="governancePreview" shadow="never">
        <template #header>影响预览</template>
        <el-descriptions :column="2">
          <el-descriptions-item label="关联推荐">
            {{ governancePreview.recommendationCount }}
          </el-descriptions-item>
          <el-descriptions-item label="影响用户">
            {{ governancePreview.affectedUserCount }}
          </el-descriptions-item>
          <el-descriptions-item label="通知数量">
            {{ governancePreview.notificationCount }}
          </el-descriptions-item>
          <el-descriptions-item label="执行方式">
            {{ governancePreview.asynchronous ? '异步任务' : '同步事务' }}
          </el-descriptions-item>
          <el-descriptions-item label="预览有效期" :span="2">
            {{ formatDateTime(governancePreview.expiresAt) }}
          </el-descriptions-item>
        </el-descriptions>
        <el-alert
          v-for="warning in governancePreview.warnings || []"
          :key="warning"
          :title="warning"
          type="warning"
          show-icon
          :closable="false"
          class="mt-2"
        />
      </el-card>

      <template #footer>
        <el-button @click="governanceVisible = false">取消</el-button>
        <el-button :loading="governanceLoading" @click="previewGovernance">
          {{ governancePreview ? '重新预览' : '预览影响' }}
        </el-button>
        <el-button
          v-if="governancePreview"
          type="danger"
          :loading="governanceLoading"
          @click="executeGovernance"
        >
          确认软删除
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import AuditPanel from '@/view/orderFood/components/AuditPanel.vue'
import {
  getOrderFoodUserRecipeDetail,
  getOrderFoodUserRecipeList
} from '@/api/orderfood/user-recipe'
import {
  executeOrderFoodGovernanceAction,
  previewOrderFoodGovernanceAction
} from '@/api/orderfood/governance'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodUserRecipes'
})

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const routeQueryValue = (value) => (Array.isArray(value) ? value[0] : value || '')
const lockedUserId = routeQueryValue(route.query.userId)
const lockedDishId = routeQueryValue(route.query.dishId)

const hasBtnPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))
const canReadGovernance = computed(() =>
  Boolean(btnAuth['orderfood:governance:read'])
)
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canPrivateRead = computed(() =>
  hasBtnPermission(
    'orderfood:user-recipe:private-read',
    'userRecipePrivateRead',
    'privateRead'
  )
)
const canGovernanceExecute = computed(() =>
  hasBtnPermission(
    'orderfood:governance:execute',
    'governanceExecute',
    'executeGovernance'
  )
)
const canReadUserDishes = computed(() =>
  hasBtnPermission(
    'orderfood:user-dish:read',
    'userDishRead',
    'viewUserDishes'
  )
)

const createDefaultSearch = () => ({
  keyword: '',
  userId: lockedUserId,
  userKeyword: '',
  dishId: lockedDishId,
  minDishCount: undefined,
  maxDishCount: undefined,
  deletionStatus: 'active',
  contentState: '',
  hasNote: undefined,
  createdRange: [],
  updatedRange: [],
  sortBy: 'createdAt',
  sortOrder: 'desc'
})

const searchInfo = ref(createDefaultSearch())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const tableData = ref([])
const listLoading = ref(false)
const listError = ref('')

const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(([, value]) => {
      if (value === '' || value === null || typeof value === 'undefined') return false
      return !Array.isArray(value) || value.length > 0
    })
  )

const buildListParams = () => {
  const { createdRange, updatedRange, ...filters } = searchInfo.value
  return compactParams({
    page: page.value,
    pageSize: pageSize.value,
    ...filters,
    createdFrom: toShanghaiRFC3339(createdRange?.[0]),
    createdTo: toShanghaiRFC3339(createdRange?.[1]),
    updatedFrom: toShanghaiRFC3339(updatedRange?.[0]),
    updatedTo: toShanghaiRFC3339(updatedRange?.[1])
  })
}

const getTableData = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodUserRecipeList(buildListParams()),
      '用户菜谱列表加载失败'
    )
    tableData.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    tableData.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '用户菜谱列表加载失败')
  } finally {
    listLoading.value = false
  }
}

const onSubmit = () => {
  page.value = 1
  getTableData()
}

const onReset = () => {
  searchInfo.value = createDefaultSearch()
  page.value = 1
  getTableData()
}

const handleSizeChange = () => {
  page.value = 1
  getTableData()
}

const handleCurrentChange = () => {
  getTableData()
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const activeRecipe = ref(null)
const detail = ref(null)
const orderedDishes = computed(() =>
  [...(detail.value?.dishes || [])].sort(
    (left, right) => Number(left.sortOrder) - Number(right.sortOrder)
  )
)

const resetDetail = () => {
  activeRecipe.value = null
  detail.value = null
  detailError.value = ''
}

const loadPrivateDetail = async () => {
  if (!canPrivateRead.value || !activeRecipe.value?.id) return
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getOrderFoodUserRecipeDetail(activeRecipe.value.id),
      '菜谱私有详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '菜谱私有详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const openPrivateDetail = async (row) => {
  if (!canPrivateRead.value) return
  activeRecipe.value = row
  detailVisible.value = true
  await loadPrivateDetail()
}
const openUser = (userId) => {
  if (!userId) return
  router.push({ name: 'OrderFoodUsers', query: { userId } })
}

const openDish = (dishId) => {
  router.push({
    name: 'OrderFoodUserDishes',
    query: { keyword: dishId }
  })
}
const openRecipeDishes = (recipeId) => {
  router.push({
    name: 'OrderFoodUserDishes',
    query: { recipeId }
  })
}
const openGovernanceRecords = (recipeId) => {
  router.push({
    name: 'OrderFoodGovernanceRecords',
    query: { targetType: 'recipe', targetId: recipeId }
  })
}

const governanceVisible = ref(false)
const governanceLoading = ref(false)
const governanceTarget = ref(null)
const governancePreview = ref(null)
const governanceFormRef = ref(null)
const governanceForm = ref({
  violationType: '',
  severity: 'normal',
  reason: ''
})
const governanceRules = {
  violationType: [
    { required: true, message: '请输入违规类型', trigger: 'blur' },
    { max: 40, message: '违规类型最多 40 字', trigger: 'blur' }
  ],
  severity: [{ required: true, message: '请选择严重程度', trigger: 'change' }],
  reason: [
    { required: true, message: '请输入处理原因', trigger: 'blur' },
    { min: 4, max: 300, message: '原因长度应为 4-300 字', trigger: 'blur' }
  ]
}

const clearGovernancePreview = () => {
  governancePreview.value = null
}

const openGovernanceDialog = (row) => {
  governanceTarget.value = row
  governanceForm.value = {
    violationType: '',
    severity: 'normal',
    reason: ''
  }
  governancePreview.value = null
  governanceVisible.value = true
}

const resetGovernance = () => {
  governanceTarget.value = null
  governancePreview.value = null
  governanceForm.value = {
    violationType: '',
    severity: 'normal',
    reason: ''
  }
}

const governancePayload = () => ({
  targetType: 'recipe',
  targetId: governanceTarget.value.id,
  actions: ['soft_delete_recipe'],
  violationType: governanceForm.value.violationType,
  severity: governanceForm.value.severity,
  reason: governanceForm.value.reason,
  expectedVersion: governanceTarget.value.version
})

const previewGovernance = async () => {
  const valid = await governanceFormRef.value?.validate().catch(() => false)
  if (!valid || !governanceTarget.value) return
  governanceLoading.value = true
  try {
    governancePreview.value = unwrapOrderFoodResponse(
      await previewOrderFoodGovernanceAction(governancePayload()),
      '违规处理影响预览失败'
    )
  } catch (error) {
    governancePreview.value = null
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('菜谱版本已变化，请刷新后重新预览。')
      await getTableData()
    } else if (!error?.orderFoodResponse) {
      ElMessage.error(getOrderFoodErrorMessage(error))
    }
  } finally {
    governanceLoading.value = false
  }
}

const executeGovernance = async () => {
  if (!governancePreview.value || !governanceTarget.value) return
  const targetId = governanceTarget.value.id
  governanceLoading.value = true
  try {
    const result = unwrapOrderFoodResponse(
      await executeOrderFoodGovernanceAction({
        ...governancePayload(),
        previewToken: governancePreview.value.previewToken
      }),
      '违规处理执行失败'
    )
    ElMessage.success(
      result?.status === 'pending'
        ? `异步处理任务已创建：${result.jobId || '等待调度'}`
        : `菜谱已软删除，共影响 ${result?.affectedCount || 0} 条记录`
    )
    governanceVisible.value = false
    if (detailVisible.value && activeRecipe.value?.id === targetId) {
      detailVisible.value = false
    }
    await getTableData()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('菜谱或预览已失效，请刷新后重新确认。')
      clearGovernancePreview()
      await getTableData()
    } else if (!error?.orderFoodResponse) {
      ElMessage.error(getOrderFoodErrorMessage(error))
    }
  } finally {
    governanceLoading.value = false
  }
}

const formatDateTime = (value) => (value ? formatDate(value) : '—')
const contentStateLabel = (value) =>
  ({
    ready: '内容完整',
    empty: '空菜谱',
    contains_unavailable: '含不可用菜品'
  })[value] || value
const contentStateType = (value) =>
  ({ ready: 'success', empty: 'info', contains_unavailable: 'warning' })[value] ||
  'info'
const sourceTypeLabel = (value) =>
  ({
    manual: '用户手动创建',
    creator_copy: '创作者菜品复制',
    official_copy: '官方菜品复制',
    meal_suggestion_copy: '饭局建议复制'
  })[value] || value
const dishItemStatusLabel = (value) =>
  ({ draft: '草稿', usable: '可用', deleted: '已删除' })[value] || value

getTableData().then(() => {
  if (route.query.openRecipeId) {
    openPrivateDetail({ id: routeQueryValue(route.query.openRecipeId) }).then(
      () => {
        if (
          route.query.openGovernance === '1' &&
          canGovernanceExecute.value &&
          detail.value
        ) {
          openGovernanceDialog(detail.value)
        }
      }
    )
  }
})
</script>
