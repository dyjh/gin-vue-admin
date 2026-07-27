<template>
  <div class="order-food-list-page">
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="菜品">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            maxlength="60"
            placeholder="输入菜品名称"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="searchInfo.categoryId" clearable placeholder="全部" class="w-40">
            <el-option
              v-for="item in categories"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select
            v-model="searchInfo.tagIds"
            multiple
            clearable
            collapse-tags
            :multiple-limit="3"
            placeholder="全部"
            class="w-56"
          >
            <el-option
              v-for="item in tags"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable placeholder="全部" class="w-32">
            <el-option label="草稿" value="draft" />
            <el-option label="可用" value="usable" />
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
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </AdvancedSearchPanel>

    <div class="gva-table-box">
      <div class="table-header">
        <div>
          <div class="text-lg font-medium">官方菜品</div>
          <div class="mt-1 text-sm text-gray-500">
            平台维护的标准菜品内容。编辑或删除时会自动下线在线推荐，不影响用户已经复制的菜品和历史饭局。
          </div>
        </div>
        <el-button v-if="canCreate" type="primary" @click="openCreate">新增官方菜品</el-button>
      </div>

      <el-alert v-if="listError" :title="listError" type="error" show-icon class="mb-4">
        <template #default>
          <el-button link type="primary" @click="loadList">重新加载</el-button>
        </template>
      </el-alert>

      <el-table
        v-loading="listLoading"
        :data="items"
        row-key="id"
        :empty-text="listError ? '加载失败' : '暂无官方菜品'"
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
                <template #error><div class="cover-fallback">无图</div></template>
              </el-image>
              <div class="min-w-0">
                <div class="font-medium">{{ row.name }}</div>
                <div class="mt-1 text-xs text-gray-400">{{ row.id }}</div>
              </div>
            </div>
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
        <el-table-column label="份量" width="90">
          <template #default="{ row }">{{ row.serving }} 人份</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'usable' ? 'success' : 'info'">
              {{ row.status === 'usable' ? '可用' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="推荐" width="90" align="right">
          <template #default="{ row }">
            <div>{{ row.recommendationCount || 0 }}</div>
            <div
              v-if="row.onlineRecommendationCount"
              class="text-xs text-green-600"
            >
              在线 {{ row.onlineRecommendationCount }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="复制" width="90" align="right">
          <template #default="{ row }">{{ row.copyCount || 0 }}</template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="修改管理员" min-width="150">
          <template #default="{ row }">
            {{
              row.updatedBy?.nickname ||
                row.updatedBy?.username ||
                row.updatedBy?.id ||
                '—'
            }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">查看</el-button>
            <el-button v-if="canUpdate" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button
              v-if="canCreateRecommendation && !row.recommendationCount && row.status === 'usable'"
              link
              type="success"
              @click="createRecommendationDraft(row)"
            >
              加入推荐
            </el-button>
            <el-button
              v-else-if="canReadRecommendations && row.recommendationCount"
              link
              type="primary"
              @click="openRecommendationPage(row.id)"
            >
              查看推荐
            </el-button>
            <el-button v-if="canDelete" link type="danger" @click="removeItem(row)">
              软删除
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

    <el-drawer v-model="detailVisible" title="官方菜品详情" size="760px">
      <div v-loading="detailLoading" class="min-h-48">
        <el-alert v-if="detailError" :title="detailError" type="error" show-icon />
        <template v-if="detail">
          <div class="dish-cell mb-4">
            <el-image
              :src="detail.coverUrl"
              fit="cover"
              class="detail-cover"
              :preview-src-list="[detail.coverUrl]"
              preview-teleported
            />
            <div>
              <div class="text-xl font-medium">{{ detail.name }}</div>
              <div class="mt-1 text-sm text-gray-500">
                {{ detail.category?.name }} · 默认 {{ detail.serving }} 人份
              </div>
              <div class="mt-2 flex gap-2">
                <el-tag :type="detail.status === 'usable' ? 'success' : 'info'">
                  {{ detail.status === 'usable' ? '可用' : '草稿' }}
                </el-tag>
                <el-tag effect="plain">推荐 {{ detail.recommendationCount || 0 }} 条</el-tag>
                <el-tag v-if="detail.onlineRecommendationCount" type="success">
                  在线 {{ detail.onlineRecommendationCount }} 条
                </el-tag>
                <el-tag effect="plain">复制 {{ detail.copyCount || 0 }} 次</el-tag>
              </div>
            </div>
          </div>
          <div class="mb-4 flex flex-wrap gap-2">
            <el-button
              v-if="canCreateRecommendation && !detail.recommendationCount && detail.status === 'usable'"
              type="success"
              @click="createRecommendationDraft(detail)"
            >
              精选到推荐
            </el-button>
            <el-button
              v-if="canReadRecommendations && detail.recommendationCount"
              @click="openRecommendationPage(detail.id)"
            >
              查看推荐
            </el-button>
          </div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="说明">{{ detail.description || '—' }}</el-descriptions-item>
            <el-descriptions-item label="标签">
              <div class="flex flex-wrap gap-1">
                <el-tag
                  v-for="tag in detail.tags || []"
                  :key="tag.id"
                  effect="plain"
                  size="small"
                >
                  {{ tag.name }}
                </el-tag>
                <span v-if="!(detail.tags || []).length">—</span>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="创建人">
              {{ detail.createdBy?.nickname || detail.createdBy?.username || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="更新人">
              {{ detail.updatedBy?.nickname || detail.updatedBy?.username || '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(detail.createdAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatDateTime(detail.updatedAt) }}
            </el-descriptions-item>
          </el-descriptions>
          <div class="section-title">食材</div>
          <el-table :data="detail.ingredients || []" size="small" border>
            <el-table-column prop="name" label="食材" />
            <el-table-column label="用量">
              <template #default="{ row }">
                {{ row.quantity || '适量' }}{{ row.unit?.name || '' }}
              </template>
            </el-table-column>
            <el-table-column label="备注">
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
      v-model="editorVisible"
      :title="editorMode === 'create' ? '新增官方菜品' : '编辑官方菜品'"
      width="900px"
      destroy-on-close
      @closed="resetEditor"
    >
      <el-form
        ref="editorFormRef"
        :model="editorForm"
        :rules="editorRules"
        label-width="90px"
        v-loading="editorLoading"
      >
        <div class="form-grid">
          <el-form-item label="菜品名称" prop="name">
            <el-input v-model.trim="editorForm.name" maxlength="40" show-word-limit />
          </el-form-item>
          <el-form-item label="分类" prop="categoryId">
            <el-select v-model="editorForm.categoryId" filterable class="w-full">
              <el-option
                v-for="item in enabledCategories"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="默认份量" prop="serving">
            <el-input-number v-model="editorForm.serving" :min="1" :max="20" />
          </el-form-item>
          <el-form-item label="标签" class="form-span">
            <el-select
              v-model="editorForm.tagIds"
              multiple
              filterable
              :multiple-limit="3"
              class="w-full"
              placeholder="最多选择 3 个标签"
            >
              <el-option
                v-for="item in enabledTags"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="说明" class="form-span">
            <el-input
              v-model.trim="editorForm.description"
              type="textarea"
              :rows="2"
              maxlength="180"
              show-word-limit
            />
          </el-form-item>
          <el-form-item label="封面" prop="coverFileId" class="form-span">
            <div class="cover-editor">
              <el-image
                v-if="editorForm.coverUrl"
                :src="editorForm.coverUrl"
                fit="cover"
                class="editor-cover"
              />
              <div v-else class="editor-cover cover-fallback">未选择</div>
              <div>
                <el-button @click="openCoverLibrary">从封面库选择</el-button>
                <el-upload
                  v-if="canUploadCover"
                  class="inline-upload"
                  :show-file-list="false"
                  accept="image/jpeg,image/png,image/gif"
                  :http-request="uploadCover"
                >
                  <el-button :loading="coverUploading" type="primary">上传新封面</el-button>
                </el-upload>
                <div class="mt-2 text-xs text-gray-500">
                  JPG、PNG 或 GIF，最大 10 MB；管理端上传不走图片审核。
                </div>
              </div>
            </div>
          </el-form-item>
        </div>

        <div class="editor-section">
          <div class="section-header">
            <span>食材</span>
            <el-button link type="primary" @click="addIngredient">添加食材</el-button>
          </div>
          <div
            v-for="(item, index) in editorForm.ingredients"
            :key="`ingredient-${index}`"
            class="ingredient-row"
          >
            <el-input v-model.trim="item.name" maxlength="30" placeholder="食材名称" />
            <el-input v-model.trim="item.quantity" maxlength="20" placeholder="用量，如 200" />
            <el-select v-model="item.unitId" clearable filterable placeholder="单位">
              <el-option
                v-for="unit in enabledUnits"
                :key="unit.id"
                :label="unit.name"
                :value="unit.id"
              />
            </el-select>
            <el-input v-model.trim="item.note" maxlength="80" placeholder="备注（可选）" />
            <el-button link type="danger" @click="removeIngredient(index)">移除</el-button>
          </div>
        </div>

        <div class="editor-section">
          <div class="section-header">
            <span>步骤</span>
            <el-button link type="primary" @click="addStep">添加步骤</el-button>
          </div>
          <div
            v-for="(item, index) in editorForm.steps"
            :key="`step-${index}`"
            class="step-editor-row"
          >
            <span class="step-number">{{ index + 1 }}</span>
            <el-input
              v-model.trim="item.description"
              type="textarea"
              :rows="2"
              maxlength="500"
              show-word-limit
              placeholder="步骤说明"
            />
            <el-button link type="danger" @click="removeStep(index)">移除</el-button>
          </div>
        </div>
      </el-form>
      <el-alert
        v-if="editorMode === 'edit' && editorForm.onlineRecommendationCount > 0"
        :title="`保存后将自动下线 ${editorForm.onlineRecommendationCount} 条当前在线推荐。`"
        type="warning"
        show-icon
        :closable="false"
      />
      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button
          :loading="editorSaving"
          @click="submitEditor('draft')"
        >
          保存草稿
        </el-button>
        <el-button
          type="primary"
          :loading="editorSaving"
          @click="submitEditor('usable')"
        >
          保存并设为可用
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="coverLibraryVisible" title="选择官方菜品封面" width="780px">
      <div class="mb-3 flex gap-2">
        <el-input
          v-model.trim="coverKeyword"
          clearable
          placeholder="搜索原始文件名"
          @keyup.enter="loadCovers"
        />
        <el-button @click="loadCovers">搜索</el-button>
      </div>
      <div v-loading="coverLoading" class="cover-grid">
        <button
          v-for="cover in covers"
          :key="cover.fileId"
          type="button"
          class="cover-card"
          :class="{ selected: editorForm.coverFileId === cover.fileId }"
          @click="selectCover(cover)"
        >
          <el-image :src="cover.url" fit="cover" class="cover-card-image" />
          <span :title="cover.fileName">{{ cover.fileName }}</span>
        </button>
      </div>
      <el-empty v-if="!coverLoading && covers.length === 0" description="暂无可用封面" />
      <div class="flex justify-end pt-3">
        <el-pagination
          v-model:current-page="coverPage"
          :page-size="20"
          :total="coverTotal"
          layout="total, prev, pager, next"
          @current-change="loadCovers"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import { getCatalogItemList } from '@/api/orderfood/catalog'
import {
  createOfficialDish,
  deleteOfficialDish,
  getOfficialDishCoverList,
  getOfficialDishDetail,
  getOfficialDishList,
  updateOfficialDish,
  uploadOfficialDishCover
} from '@/api/orderfood/officialDish'
import { createRecommendation } from '@/api/orderfood/recommendation'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodOfficialDishes' })

const router = useRouter()
const route = useRoute()
const btnAuth = useBtnAuth()
const hasPermission = (permission) => Boolean(btnAuth[permission])
const canCreate = computed(() => hasPermission('orderfood:official-dish:create'))
const canUpdate = computed(() => hasPermission('orderfood:official-dish:update'))
const canDelete = computed(() => hasPermission('orderfood:official-dish:delete'))
const canUploadCover = computed(() =>
  hasPermission('orderfood:official-dish:cover-upload')
)
const canCreateRecommendation = computed(() =>
  hasPermission('orderfood:recommendation:create')
)
const canReadRecommendations = computed(() =>
  hasPermission('orderfood:recommendation:read')
)

const compactParams = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(([, value]) => {
      if (value === '' || value === null || typeof value === 'undefined') return false
      return !Array.isArray(value) || value.length > 0
    })
  )
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const createDefaultSearch = () => ({
  keyword: '',
  categoryId: '',
  tagIds: [],
  status: '',
  createdRange: []
})
const searchInfo = ref(createDefaultSearch())
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const items = ref([])
const listLoading = ref(false)
const listError = ref('')

const categories = ref([])
const tags = ref([])
const units = ref([])
const enabledCategories = computed(() => categories.value.filter((item) => item.enabled))
const enabledTags = computed(() => tags.value.filter((item) => item.enabled))
const enabledUnits = computed(() => units.value.filter((item) => item.enabled))
const loadCatalogs = async () => {
  try {
    const [categoryData, tagData, unitData] = await Promise.all(
      ['category', 'tag', 'unit'].map(async (resource) =>
        unwrapOrderFoodResponse(
          await getCatalogItemList(resource, {
            page: 1,
            pageSize: 100,
            sortBy: 'sortOrder',
            sortOrder: 'asc'
          }),
          '基础数据加载失败'
        )
      )
    )
    categories.value = categoryData?.list || []
    tags.value = tagData?.list || []
    units.value = unitData?.list || []
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '分类、标签或单位加载失败'))
  }
}

const loadList = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const { createdRange, ...filters } = searchInfo.value
    const data = unwrapOrderFoodResponse(
      await getOfficialDishList(
        compactParams({
          page: page.value,
          pageSize: pageSize.value,
          ...filters,
          createdFrom: toShanghaiRFC3339(createdRange?.[0]),
          createdTo: toShanghaiRFC3339(createdRange?.[1]),
          sortBy: 'updatedAt',
          sortOrder: 'desc'
        })
      ),
      '官方菜品加载失败'
    )
    items.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    items.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '官方菜品加载失败')
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
const loadDetail = async (id) => {
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getOfficialDishDetail(id),
      '官方菜品详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '官方菜品详情加载失败')
  } finally {
    detailLoading.value = false
  }
}
const openDetail = (row) => {
  detailVisible.value = true
  loadDetail(row.id)
}

const emptyIngredient = (sortOrder = 1) => ({
  name: '',
  quantity: '',
  unitId: null,
  note: '',
  sortOrder
})
const emptyStep = (sortOrder = 1) => ({
  description: '',
  imageUrl: null,
  sortOrder
})
const createDefaultEditor = () => ({
  id: '',
  name: '',
  categoryId: '',
  tagIds: [],
  serving: 1,
  description: '',
  coverFileId: '',
  coverUrl: '',
  status: 'draft',
  ingredients: [],
  steps: [],
  recommendationCount: 0,
  onlineRecommendationCount: 0,
  expectedVersion: undefined
})
const editorVisible = ref(false)
const editorMode = ref('create')
const editorLoading = ref(false)
const editorSaving = ref(false)
const editorFormRef = ref(null)
const editorForm = ref(createDefaultEditor())
const editorRules = {
  name: [{ required: true, message: '请输入菜品名称', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请选择分类', trigger: 'change' }],
  serving: [{ required: true, message: '请输入默认份量', trigger: 'change' }],
  coverFileId: [{ required: true, message: '请选择封面', trigger: 'change' }]
}
const openCreate = () => {
  editorMode.value = 'create'
  editorForm.value = createDefaultEditor()
  editorVisible.value = true
}
const openEdit = async (row) => {
  editorMode.value = 'edit'
  editorVisible.value = true
  editorLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getOfficialDishDetail(row.id),
      '官方菜品详情加载失败'
    )
    editorForm.value = {
      id: data.id,
      name: data.name,
      categoryId: data.category?.id || '',
      tagIds: (data.tags || []).map((item) => item.id),
      serving: data.serving,
      description: data.description || '',
      coverFileId: data.coverFileId,
      coverUrl: data.coverUrl,
      status: data.status,
      ingredients: (data.ingredients || []).map((item, index) => ({
        name: item.name,
        quantity: item.quantity || '',
        unitId: item.unit?.id || null,
        note: item.note || '',
        sortOrder: item.sortOrder || index + 1
      })),
      steps: (data.steps || []).map((item, index) => ({
        description: item.description,
        imageUrl: item.imageUrl || null,
        sortOrder: item.sortOrder || index + 1
      })),
      recommendationCount: data.recommendationCount || 0,
      onlineRecommendationCount: data.onlineRecommendationCount || 0,
      expectedVersion: data.version
    }
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '官方菜品详情加载失败'))
    editorVisible.value = false
  } finally {
    editorLoading.value = false
  }
}
const resetEditor = () => {
  editorForm.value = createDefaultEditor()
  editorFormRef.value?.clearValidate()
}
const addIngredient = () =>
  editorForm.value.ingredients.push(emptyIngredient(editorForm.value.ingredients.length + 1))
const removeIngredient = (index) => {
  editorForm.value.ingredients.splice(index, 1)
}
const addStep = () => editorForm.value.steps.push(emptyStep(editorForm.value.steps.length + 1))
const removeStep = (index) => {
  editorForm.value.steps.splice(index, 1)
}
const editorPayload = () => ({
  name: editorForm.value.name,
  categoryId: editorForm.value.categoryId,
  tagIds: editorForm.value.tagIds,
  serving: editorForm.value.serving,
  description: editorForm.value.description || null,
  coverFileId: editorForm.value.coverFileId,
  status: editorForm.value.status,
  ingredients: editorForm.value.ingredients.map((item, index) => ({
    name: item.name,
    quantity: item.quantity || null,
    unitId: item.unitId || null,
    note: item.note || null,
    sortOrder: index + 1
  })),
  steps: editorForm.value.steps.map((item, index) => ({
    description: item.description,
    imageUrl: item.imageUrl || null,
    sortOrder: index + 1
  }))
})
const submitEditor = async (targetStatus) => {
  editorForm.value.status = targetStatus
  const valid = await editorFormRef.value?.validate().catch(() => false)
  if (!valid) return
  if (
    targetStatus === 'usable' &&
    (editorForm.value.ingredients.length === 0 ||
      editorForm.value.steps.length === 0)
  ) {
    ElMessage.warning('设为可用前，至少填写一项食材和一个步骤')
    return
  }
  if (editorForm.value.ingredients.some((item) => !item.name.trim())) {
    ElMessage.warning('请填写完整的食材名称')
    return
  }
  if (editorForm.value.steps.some((item) => !item.description.trim())) {
    ElMessage.warning('请填写完整的步骤说明')
    return
  }
  if (
    editorMode.value === 'edit' &&
    editorForm.value.onlineRecommendationCount > 0
  ) {
    const confirmed = await ElMessageBox.confirm(
      `保存后将自动下线 ${editorForm.value.onlineRecommendationCount} 条当前在线推荐。是否继续？`,
      '确认保存',
      { type: 'warning', confirmButtonText: '保存并下线推荐' }
    ).then(() => true).catch(() => false)
    if (!confirmed) return
  }
  editorSaving.value = true
  try {
    if (editorMode.value === 'create') {
      unwrapOrderFoodResponse(
        await createOfficialDish(editorPayload()),
        '官方菜品创建失败'
      )
      ElMessage.success('官方菜品已创建')
    } else {
      const result = unwrapOrderFoodResponse(
        await updateOfficialDish(editorForm.value.id, {
          ...editorPayload(),
          expectedVersion: editorForm.value.expectedVersion
        }),
        '官方菜品保存失败'
      )
      const count = Number(result?.offlineRecommendationCount || 0)
      ElMessage.success(count > 0 ? `已保存并下线 ${count} 条推荐` : '官方菜品已保存')
    }
    editorVisible.value = false
    await loadList()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('菜品版本已变化，请刷新后重试')
      await loadList()
      return
    }
    ElMessage.error(getOrderFoodErrorMessage(error, '官方菜品保存失败'))
  } finally {
    editorSaving.value = false
  }
}

const removeItem = async (row) => {
  const { value: reason } = await ElMessageBox.prompt(
    `${row.onlineRecommendationCount ? `将同步下线 ${row.onlineRecommendationCount} 条当前在线推荐。` : ''}请输入删除原因（至少 4 个字符）`,
    '删除官方菜品',
    {
      type: 'warning',
      inputValidator: (value) =>
        String(value || '').trim().length >= 4 || '删除原因至少 4 个字符'
    }
  ).catch(() => ({ value: '' }))
  if (!reason) return
  try {
    const result = unwrapOrderFoodResponse(
      await deleteOfficialDish(row.id, {
        reason: reason.trim(),
        expectedVersion: row.version
      }),
      '官方菜品删除失败'
    )
    ElMessage.success(
      result?.offlineRecommendationCount
        ? `已删除并下线 ${result.offlineRecommendationCount} 条推荐`
        : '官方菜品已删除'
    )
    await loadList()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('菜品版本已变化，请刷新后重试')
      await loadList()
      return
    }
    ElMessage.error(getOrderFoodErrorMessage(error, '官方菜品删除失败'))
  }
}

const createRecommendationDraft = async (row) => {
  try {
    unwrapOrderFoodResponse(
      await createRecommendation({
        sourceType: 'official',
        sourceDishId: row.id,
        position: 'home_featured',
        sortOrder: 1,
        displayNote: null
      }),
      '推荐草稿创建失败'
    )
    ElMessage.success('推荐草稿已创建')
    await loadList()
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '推荐草稿创建失败'))
  }
}
const openRecommendationPage = (sourceDishId) =>
  router.push({
    name: 'OrderFoodRecommendations',
    query: { sourceType: 'official', sourceDishId }
  })

const coverLibraryVisible = ref(false)
const coverLoading = ref(false)
const coverUploading = ref(false)
const coverKeyword = ref('')
const coverPage = ref(1)
const coverTotal = ref(0)
const covers = ref([])
const loadCovers = async () => {
  coverLoading.value = true
  try {
    const data = unwrapOrderFoodResponse(
      await getOfficialDishCoverList(
        compactParams({
          page: coverPage.value,
          pageSize: 20,
          keyword: coverKeyword.value,
          sortBy: 'createdAt',
          sortOrder: 'desc'
        })
      ),
      '封面库加载失败'
    )
    covers.value = data?.list || []
    coverTotal.value = Number(data?.total || 0)
  } catch (error) {
    covers.value = []
    ElMessage.error(getOrderFoodErrorMessage(error, '封面库加载失败'))
  } finally {
    coverLoading.value = false
  }
}
const openCoverLibrary = () => {
  coverLibraryVisible.value = true
  coverPage.value = 1
  loadCovers()
}
const selectCover = (cover) => {
  editorForm.value.coverFileId = cover.fileId
  editorForm.value.coverUrl = cover.url
  coverLibraryVisible.value = false
  editorFormRef.value?.validateField('coverFileId')
}
const uploadCover = async ({ file }) => {
  coverUploading.value = true
  try {
    const cover = unwrapOrderFoodResponse(
      await uploadOfficialDishCover(file),
      '封面上传失败'
    )
    selectCover(cover)
    ElMessage.success('封面已上传并选中')
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '封面上传失败'))
  } finally {
    coverUploading.value = false
  }
}

Promise.all([loadCatalogs(), loadList()]).then(() => {
  if (route.query.officialDishId) {
    openDetail({ id: String(route.query.officialDishId) })
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

.dish-cell,
.cover-editor {
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
}

.editor-cover {
  width: 128px;
  height: 88px;
  flex: none;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
}

.cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #a8abb2;
  font-size: 12px;
  background: #f2f3f5;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 16px;
}

.form-span {
  grid-column: 1 / -1;
}

.inline-upload {
  display: inline-block;
  margin-left: 8px;
}

.section-title {
  margin: 20px 0 10px;
  font-weight: 600;
}

.step-row,
.section-header,
.ingredient-row,
.step-editor-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.step-row {
  margin-bottom: 10px;
}

.editor-section {
  margin-top: 18px;
}

.section-header {
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
}

.ingredient-row {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1fr 1.4fr auto;
  margin-bottom: 8px;
}

.step-editor-row {
  margin-bottom: 10px;
}

.step-number {
  width: 28px;
  padding-top: 8px;
  text-align: center;
  color: var(--el-text-color-secondary);
}

.cover-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  min-height: 160px;
}

.cover-card {
  overflow: hidden;
  padding: 0;
  border: 2px solid transparent;
  border-radius: 8px;
  color: var(--el-text-color-regular);
  background: var(--el-bg-color);
  cursor: pointer;
}

.cover-card.selected {
  border-color: var(--el-color-primary);
}

.cover-card-image {
  width: 100%;
  height: 100px;
  display: block;
}

.cover-card span {
  display: block;
  overflow: hidden;
  padding: 6px 8px;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
}
</style>
