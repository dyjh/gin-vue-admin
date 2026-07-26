<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="菜品 ID">
          <el-input v-model.trim="searchInfo.dishId" clearable placeholder="输入菜品 ID" />
        </el-form-item>
        <el-form-item label="菜品">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            maxlength="60"
            placeholder="输入菜品名称"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="作者 ID">
          <el-input
            v-model.trim="searchInfo.authorId"
            clearable
            placeholder="输入用户 ID"
          />
        </el-form-item>
        <el-form-item label="作者昵称">
          <el-input
            v-model.trim="searchInfo.authorKeyword"
            clearable
            maxlength="60"
            placeholder="输入用户 ID 或昵称"
          />
        </el-form-item>
        <el-form-item label="分类 ID">
          <el-input
            v-model.trim="searchInfo.categoryId"
            clearable
            placeholder="输入分类 ID"
          />
        </el-form-item>
        <el-form-item label="标签 ID">
          <el-input
            v-model.trim="searchInfo.tagInput"
            clearable
            placeholder="多个标签用逗号分隔"
          />
        </el-form-item>
        <el-form-item label="推荐状态">
          <el-select
            v-model="searchInfo.selected"
            clearable
            placeholder="全部"
            class="w-36"
          >
            <el-option label="已加入推荐" :value="true" />
            <el-option label="未加入推荐" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-select v-model="searchInfo.sortBy" class="w-36">
            <el-option label="允许发现时间" value="discoverableAt" />
            <el-option label="点选次数" value="voteCount" />
            <el-option label="复制次数" value="copyCount" />
            <el-option label="创建时间" value="createdAt" />
          </el-select>
        </el-form-item>
        <el-form-item label="允许发现时间">
          <el-date-picker
            v-model="searchInfo.discoverableRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="table-header">
        <div>
          <div class="text-lg font-medium">可发现菜品</div>
          <div class="mt-1 text-sm text-gray-500">
            这里只展示用户主动允许被发现的可用菜品；管理端只能查看或加入推荐，不能编辑用户内容。
          </div>
        </div>
      </div>

      <el-alert
        v-if="listError"
        :title="listError"
        type="error"
        show-icon
        class="mb-4"
      >
        <template #default>
          <el-button link type="primary" @click="loadList">重新加载</el-button>
        </template>
      </el-alert>

      <el-table
        v-loading="listLoading"
        :data="items"
        row-key="id"
        :empty-text="listError ? '加载失败' : '暂无可发现菜品'"
      >
        <el-table-column label="菜品" min-width="260" fixed="left">
          <template #default="{ row }">
            <div class="dish-cell">
              <el-image
                :src="row.coverUrl"
                fit="cover"
                class="dish-cover"
                :preview-src-list="row.coverUrl ? [row.coverUrl] : []"
                preview-teleported
              >
                <template #error>
                  <div class="cover-fallback">无图</div>
                </template>
              </el-image>
              <div class="min-w-0">
                <div class="font-medium">{{ row.name }}</div>
                <div class="mt-1 text-xs text-gray-400">{{ row.id }}</div>
                <el-tag
                  v-if="row.sourceLocked"
                  class="mt-1"
                  type="warning"
                  size="small"
                >
                  来源锁定，不能精选
                </el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="作者" min-width="180">
          <template #default="{ row }">
            <div>{{ row.author?.nickname || '未命名用户' }}</div>
            <div class="text-xs text-gray-400">{{ row.author?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="分类/标签" min-width="220">
          <template #default="{ row }">
            <el-tag effect="plain">{{ row.category?.name || '未分类' }}</el-tag>
            <div class="mt-1 flex flex-wrap gap-1">
              <el-tag
                v-for="tag in row.tags || []"
                :key="tag.id"
                type="info"
                effect="plain"
                size="small"
              >
                {{ tag.name }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="点选次数" prop="voteCount" width="100" align="right" />
        <el-table-column label="复制次数" prop="copyCount" width="100" align="right" />
        <el-table-column label="推荐" width="130">
          <template #default="{ row }">
            <el-tag :type="row.selected ? 'success' : 'info'">
              {{ row.selected ? '已加入' : '未加入' }}
            </el-tag>
            <div v-if="row.recommendationPosition" class="mt-1 text-xs text-gray-400">
              首页精选
            </div>
          </template>
        </el-table-column>
        <el-table-column label="允许发现时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.discoverableAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">查看详情</el-button>
            <el-button
              v-if="canCreateRecommendation && !row.selected && !row.sourceLocked"
              link
              type="primary"
              @click="openRecommendation(row)"
            >
              加入推荐
            </el-button>
            <el-button
              v-if="canReadRecommendations && row.selected"
              link
              type="primary"
              @click="openRecommendationPage(row.id)"
            >
              查看推荐
            </el-button>
            <el-button
              v-if="canGovern"
              link
              type="danger"
              @click="openGovernance(row.id)"
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
          @current-change="loadList"
        />
      </div>
    </div>

    <el-drawer v-model="detailVisible" title="可发现菜品详情" size="760px">
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon />
        <template v-if="detail">
          <div class="dish-cell mb-4">
            <el-image
              :src="detail.coverUrl"
              fit="cover"
              class="detail-cover"
              :preview-src-list="detail.coverUrl ? [detail.coverUrl] : []"
              preview-teleported
            />
            <div>
              <div class="text-xl font-medium">{{ detail.name }}</div>
              <div class="mt-1 text-sm text-gray-500">
                {{ detail.author?.nickname || '未命名用户' }} ·
                {{ detail.category?.name || '未分类' }}
              </div>
              <div class="mt-2 flex gap-2">
                <el-tag :type="detail.sourceLocked ? 'warning' : 'success'">
                  {{ detail.sourceLocked ? '来源锁定副本' : '可作为推荐来源' }}
                </el-tag>
                <el-tag effect="plain">默认 {{ detail.serving }} 人份</el-tag>
              </div>
            </div>
          </div>
          <div class="mb-4 flex flex-wrap gap-2">
            <el-button
              v-if="canReadUsers && detail.author?.id"
              @click="openAuthor(detail.author.id)"
            >
              查看作者
            </el-button>
            <el-button
              v-if="
                canCreateRecommendation &&
                  !detail.selected &&
                  !detail.sourceLocked
              "
              type="primary"
              @click="openRecommendation(detail)"
            >
              精选到推荐
            </el-button>
            <el-button
              v-if="canReadRecommendations && detail.selected"
              @click="openRecommendationPage(detail.id)"
            >
              查看推荐
            </el-button>
            <el-button
              v-if="canGovern"
              type="danger"
              @click="openGovernance(detail.id)"
            >
              违规处理
            </el-button>
          </div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="说明">
              {{ detail.description || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="图片状态">
              {{ mediaStatusLabel(detail.mediaReviewStatus) }}
            </el-descriptions-item>
            <el-descriptions-item label="数据">
              点选 {{ detail.voteCount }} 次，复制 {{ detail.copyCount }} 次
            </el-descriptions-item>
          </el-descriptions>
          <el-card v-if="detail.recommendation" class="mt-4" shadow="never">
            <template #header>当前推荐信息</template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="推荐 ID">
                {{ detail.recommendation.id }}
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                {{ recommendationStatusLabel(detail.recommendation.status) }}
              </el-descriptions-item>
              <el-descriptions-item label="推荐位置">
                {{ recommendationPositionLabel(detail.recommendation.position) }}
              </el-descriptions-item>
              <el-descriptions-item label="排序值">
                {{ detail.recommendation.sortOrder }}
              </el-descriptions-item>
            </el-descriptions>
          </el-card>
          <el-card v-if="canReadModeration" class="mt-4" shadow="never">
            <template #header>图片审核摘要</template>
            <el-empty
              v-if="!detail.moderationSummary"
              description="暂无图片审核记录"
              :image-size="56"
            />
            <el-descriptions v-else :column="2" border>
              <el-descriptions-item label="审核状态">
                {{ mediaStatusLabel(detail.moderationSummary.status) }}
              </el-descriptions-item>
              <el-descriptions-item label="请求 ID">
                {{ detail.moderationSummary.requestId || '—' }}
              </el-descriptions-item>
              <el-descriptions-item label="风险等级">
                {{ detail.moderationSummary.riskLevel || '—' }}
              </el-descriptions-item>
              <el-descriptions-item label="风险标签">
                {{ (detail.moderationSummary.riskLabels || []).join('、') || '无' }}
              </el-descriptions-item>
              <el-descriptions-item label="审核时间" :span="2">
                {{ formatDateTime(detail.moderationSummary.createdAt) }}
              </el-descriptions-item>
            </el-descriptions>
          </el-card>
          <div class="section-title">食材</div>
          <el-table :data="detail.ingredients || []" size="small" border>
            <el-table-column prop="name" label="食材" />
            <el-table-column label="用量">
              <template #default="{ row }">
                {{ row.quantity || '适量' }}{{ row.unit?.name || '' }}
              </template>
            </el-table-column>
            <el-table-column prop="note" label="备注">
              <template #default="{ row }">{{ row.note || '—' }}</template>
            </el-table-column>
          </el-table>
          <div class="section-title">步骤</div>
          <div
            v-for="(step, index) in detail.steps || []"
            :key="`${step.sortOrder}-${index}`"
            class="step-row"
          >
            <el-tag round>{{ index + 1 }}</el-tag>
            <span>{{ step.description }}</span>
          </div>
        </template>
      </div>
    </el-drawer>

    <el-dialog
      v-model="governanceVisible"
      title="候选菜品违规处理"
      width="620px"
      destroy-on-close
      @closed="resetGovernance"
    >
      <el-alert
        title="操作必须先生成最新影响预览，再确认执行。"
        type="warning"
        show-icon
        :closable="false"
        class="mb-4"
      />
      <el-form label-width="100px">
        <el-form-item label="目标菜品">
          {{ governanceTarget?.name }} · {{ governanceTarget?.id }}
        </el-form-item>
        <el-form-item label="处理动作">
          <el-checkbox-group
            v-model="governanceForm.actions"
            @change="handleGovernanceActionChange"
          >
            <el-checkbox value="disable_discoverability">关闭允许被发现</el-checkbox>
            <el-checkbox value="soft_delete_dish">软删除菜品</el-checkbox>
            <el-checkbox v-if="canGovernanceCascade" value="delete_copy_chain">
              严重违规：处理全部复制链
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="违规类型">
          <el-input v-model.trim="governanceForm.violationType" maxlength="40" />
        </el-form-item>
        <el-form-item label="严重程度">
          <el-radio-group v-model="governanceForm.severity" @change="clearGovernancePreview">
            <el-radio value="normal">一般</el-radio>
            <el-radio value="serious">严重</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理原因">
          <el-input
            v-model.trim="governanceForm.reason"
            type="textarea"
            maxlength="300"
            show-word-limit
            @input="clearGovernancePreview"
          />
        </el-form-item>
        <el-form-item v-if="isCascadeGovernance" label="确认词">
          <el-input
            v-model.trim="governanceForm.confirmText"
            placeholder="请输入：确认处理复制链"
          />
        </el-form-item>
      </el-form>
      <el-card v-if="governancePreview" shadow="never">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="相关推荐">
            {{ governancePreview.recommendationCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="下游复制">
            {{ governancePreview.copiedDishCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="影响用户">
            {{ governancePreview.affectedUserCount || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="通知">
            {{ governancePreview.notificationCount || 0 }}
          </el-descriptions-item>
        </el-descriptions>
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
          :disabled="isCascadeGovernance && !isCascadeConfirmValid"
          @click="executeGovernance"
        >
          确认执行
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="recommendationVisible"
      title="加入首页精选"
      width="520px"
      destroy-on-close
    >
      <el-form
        ref="recommendationFormRef"
        :model="recommendationForm"
        :rules="recommendationRules"
        label-width="90px"
      >
        <el-form-item label="来源菜品">
          <el-input :model-value="recommendationForm.dishName" disabled />
        </el-form-item>
        <el-form-item label="推荐位置">
          <el-input model-value="首页精选" disabled />
        </el-form-item>
        <el-form-item label="排序值" prop="sortOrder">
          <el-input-number
            v-model="recommendationForm.sortOrder"
            :min="1"
            :max="999999"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item label="展示说明" prop="displayNote">
          <el-input
            v-model.trim="recommendationForm.displayNote"
            type="textarea"
            :rows="3"
            maxlength="160"
            show-word-limit
            placeholder="可选，小程序详情页展示"
          />
        </el-form-item>
      </el-form>
      <el-alert
        title="创建后先进入草稿，不会立即展示给小程序用户。"
        type="info"
        show-icon
        :closable="false"
      />
      <template #footer>
        <el-button @click="recommendationVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="recommendationSaving"
          @click="submitRecommendation"
        >
          创建草稿
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
import {
  createRecommendation,
  getDiscoverableDishDetail,
  getDiscoverableDishList
} from '@/api/orderfood/recommendation'
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
  name: 'OrderFoodDiscoverableDishes'
})

const router = useRouter()
const route = useRoute()
const btnAuth = useBtnAuth()
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canReadRecommendations = computed(() =>
  Boolean(btnAuth['orderfood:recommendation:read'])
)
const canReadModeration = computed(() =>
  Boolean(btnAuth['orderfood:moderation:read'])
)
const canGovern = computed(
  () => Boolean(btnAuth['orderfood:governance:execute'])
)
const canGovernanceCascade = computed(() =>
  Boolean(btnAuth['orderfood:governance:cascade'])
)
const canCreateRecommendation = computed(() =>
  Boolean(btnAuth['orderfood:recommendation:create'])
)
const createDefaultSearch = () => ({
  dishId: String(route.query.dishId || ''),
  keyword: '',
  authorId: '',
  authorKeyword: '',
  categoryId: '',
  tagInput: '',
  selected: undefined,
  discoverableRange: [],
  sortBy: 'discoverableAt'
})
const searchInfo = ref(createDefaultSearch())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const items = ref([])
const listLoading = ref(false)
const listError = ref('')

const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(([, value]) => {
      if (value === '' || value === null || typeof value === 'undefined') return false
      return !Array.isArray(value) || value.length > 0
    })
  )
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const mediaStatusLabel = (status) =>
  ({
    pending: '审核中',
    passed: '已通过',
    rejected: '未通过',
    failed: '审核失败',
    not_required: '无需审核'
  })[status] || status || '—'
const recommendationStatusLabel = (status) =>
  ({ draft: '草稿', published: '已发布', offline: '已下线' })[status] ||
  status ||
  '—'
const recommendationPositionLabel = (position) =>
  ({ home_featured: '首页精选' })[position] || position || '—'

const loadList = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const { tagInput, discoverableRange, ...filters } = searchInfo.value
    const tagIds = String(tagInput || '')
      .split(/[,，]/)
      .map((value) => value.trim())
      .filter(Boolean)
    const data = unwrapOrderFoodResponse(
      await getDiscoverableDishList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...filters,
          tagIds,
          discoverableFrom: toShanghaiRFC3339(discoverableRange?.[0]),
          discoverableTo: toShanghaiRFC3339(discoverableRange?.[1]),
          sortOrder: 'desc'
        })
      ),
      '可发现菜品加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    items.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '可发现菜品加载失败')
  } finally {
    listLoading.value = false
  }
}
const onSubmit = () => {
  page.value = 1
  loadList()
}
const onReset = () => {
  searchInfo.value = createDefaultSearch()
  page.value = 1
  loadList()
}
const handleSizeChange = () => {
  page.value = 1
  loadList()
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref(null)
const openDetail = async (row) => {
  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getDiscoverableDishDetail(row.id, {
        includeModerationSummary: canReadModeration.value
      }),
      '菜品详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '菜品详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const recommendationVisible = ref(false)
const recommendationSaving = ref(false)
const recommendationFormRef = ref(null)
const recommendationForm = ref({
  sourceDishId: '',
  dishName: '',
  sortOrder: 1,
  displayNote: ''
})
const recommendationRules = {
  sortOrder: [
    { required: true, message: '请输入排序值', trigger: 'change' },
    { type: 'number', min: 1, message: '排序值必须大于 0', trigger: 'change' }
  ]
}
const openRecommendation = (row) => {
  recommendationForm.value = {
    sourceDishId: row.id,
    dishName: row.name,
    sortOrder: 1,
    displayNote: ''
  }
  recommendationVisible.value = true
}
const submitRecommendation = async () => {
  const valid = await recommendationFormRef.value?.validate().catch(() => false)
  if (!valid) return
  recommendationSaving.value = true
  try {
    unwrapOrderFoodResponse(
      await createRecommendation({
        sourceType: 'creator',
        sourceDishId: recommendationForm.value.sourceDishId,
        position: 'home_featured',
        sortOrder: recommendationForm.value.sortOrder,
        displayNote: recommendationForm.value.displayNote || null
      }),
      '推荐草稿创建失败'
    )
    ElMessage.success('推荐草稿已创建')
    recommendationVisible.value = false
    await loadList()
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '推荐草稿创建失败'))
  } finally {
    recommendationSaving.value = false
  }
}
const openRecommendationPage = (sourceDishId) => {
  router.push({
    name: 'OrderFoodRecommendations',
    query: { sourceType: 'creator', sourceDishId }
  })
}
const openAuthor = (userId) =>
  router.push({ name: 'OrderFoodUsers', query: { userId } })
const governanceVisible = ref(false)
const governanceLoading = ref(false)
const governanceTarget = ref(null)
const governancePreview = ref(null)
const governanceForm = ref({
  actions: [],
  violationType: '',
  severity: 'normal',
  reason: '',
  confirmText: ''
})
const isCascadeGovernance = computed(() =>
  governanceForm.value.actions.includes('delete_copy_chain')
)
const isCascadeConfirmValid = computed(
  () => governanceForm.value.confirmText.trim() === '确认处理复制链'
)
const clearGovernancePreview = () => {
  governancePreview.value = null
}
const openGovernance = (dishId) => {
  const row =
    detail.value?.id === dishId
      ? detail.value
      : items.value.find((item) => item.id === dishId)
  if (!row) return
  governanceTarget.value = row
  governanceForm.value = {
    actions: [],
    violationType: '',
    severity: 'normal',
    reason: '',
    confirmText: ''
  }
  governancePreview.value = null
  governanceVisible.value = true
}
const resetGovernance = () => {
  governanceTarget.value = null
  governancePreview.value = null
}
const handleGovernanceActionChange = (actions) => {
  if (actions.includes('delete_copy_chain')) {
    governanceForm.value.actions = ['delete_copy_chain']
    governanceForm.value.severity = 'serious'
  } else {
    governanceForm.value.confirmText = ''
  }
  clearGovernancePreview()
}
const governancePayload = () => ({
  targetType: 'dish',
  targetId: governanceTarget.value.id,
  actions: governanceForm.value.actions,
  violationType: governanceForm.value.violationType,
  severity: governanceForm.value.severity,
  reason: governanceForm.value.reason,
  expectedVersion: governanceTarget.value.version
})
const validateGovernance = (requireConfirm = false) => {
  if (!governanceForm.value.actions.length) return '请选择处理动作'
  if (!governanceForm.value.violationType.trim()) return '请输入违规类型'
  if (governanceForm.value.reason.trim().length < 4) return '处理原因至少 4 个字符'
  if (requireConfirm && isCascadeGovernance.value && !isCascadeConfirmValid.value) {
    return '请输入完整确认词“确认处理复制链”'
  }
  return ''
}
const previewGovernance = async () => {
  const validationMessage = validateGovernance()
  if (validationMessage) {
    ElMessage.warning(validationMessage)
    return
  }
  governanceLoading.value = true
  try {
    governancePreview.value = unwrapOrderFoodResponse(
      await previewOrderFoodGovernanceAction(governancePayload()),
      '违规处理影响预览失败'
    )
  } catch (error) {
    governancePreview.value = null
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('菜品版本已变化，请刷新后重新预览')
      await loadList()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '违规处理影响预览失败'))
    }
  } finally {
    governanceLoading.value = false
  }
}
const executeGovernance = async () => {
  if (!governancePreview.value) return
  const validationMessage = validateGovernance(true)
  if (validationMessage) {
    ElMessage.warning(validationMessage)
    return
  }
  governanceLoading.value = true
  try {
    unwrapOrderFoodResponse(
      await executeOrderFoodGovernanceAction({
        ...governancePayload(),
        previewToken: governancePreview.value.previewToken,
        confirmText: isCascadeGovernance.value
          ? governanceForm.value.confirmText.trim()
          : undefined
      }),
      '违规处理执行失败'
    )
    ElMessage.success('违规处理已执行')
    governanceVisible.value = false
    detailVisible.value = false
    await loadList()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('菜品或预览已失效，请刷新后重新确认')
      clearGovernancePreview()
      await loadList()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '违规处理执行失败'))
    }
  } finally {
    governanceLoading.value = false
  }
}

loadList().then(() => {
  if (route.query.openDishId) {
    openDetail({ id: String(route.query.openDishId) })
  }
})
</script>

<style scoped>
.table-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.dish-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dish-cover {
  width: 56px;
  height: 56px;
  flex: none;
  border-radius: 8px;
  background: #f2f3f5;
}

.detail-cover {
  width: 120px;
  height: 90px;
  flex: none;
  border-radius: 10px;
  background: #f2f3f5;
}

.cover-fallback {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  color: #a8abb2;
  font-size: 12px;
}

.section-title {
  margin: 20px 0 10px;
  font-weight: 600;
}

.step-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-bottom: 10px;
}
</style>
