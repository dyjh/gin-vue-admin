<template>
  <div>
    <div class="gva-card mb-4">
      <div class="page-header">
        <div>
          <h2 class="page-title">标准菜品索引</h2>
          <p class="page-subtitle">
            维护生成建议的菜名、菜系和配料校验依据。这里不是完整菜谱库，也不会进入平台推荐池。
          </p>
        </div>
        <el-button :loading="loading" @click="loadPage">刷新</el-button>
      </div>
    </div>

    <el-alert
      v-if="pageError"
      :title="pageError"
      type="error"
      show-icon
      class="mb-4"
    />

    <div v-loading="loading">
      <div class="overview-grid mb-4">
        <div class="gva-card policy-card">
          <div class="section-header">
            <div>
              <div class="section-title">生成建议校验</div>
              <div class="section-note">
                控制模型生成的菜品是否必须匹配本页菜品索引和食材词库。
              </div>
            </div>
            <el-tag
              :type="workspace?.policy?.catalogValidationEnabled ? 'success' : 'info'"
            >
              当前{{ workspace?.policy?.catalogValidationEnabled ? '已开启' : '已关闭' }}
            </el-tag>
          </div>

          <div class="policy-control">
            <div>
              <div class="control-title">使用索引和词库校验生成结果</div>
              <div class="section-note">
                基础字段、分类、标签、单位和配料步骤一致性始终校验，不受此开关影响。
              </div>
            </div>
            <el-switch
              v-model="policyEnabled"
              :disabled="!canUpdate"
              inline-prompt
              active-text="开"
              inactive-text="关"
            />
          </div>

          <el-alert
            v-if="policyEnabled"
            title="开启后，首次生成或校验失败会在同一笔使用中再尝试 1 次；两次都失败才退积分，免费次数也不会被消耗。"
            type="warning"
            :closable="false"
            show-icon
            class="mb-4"
          />
          <el-alert
            v-else
            title="关闭后不匹配标准索引也可返回，但基础结构校验仍保留，且失败时不会额外重试。"
            type="info"
            :closable="false"
            show-icon
            class="mb-4"
          />

          <el-descriptions :column="2" border class="policy-meta">
            <el-descriptions-item label="配置版本">
              v{{ workspace?.policy?.version ?? '—' }}
            </el-descriptions-item>
            <el-descriptions-item label="固定额外重试">
              {{ workspace?.policy?.retryCount ?? 1 }} 次
            </el-descriptions-item>
            <el-descriptions-item label="更新人">
              {{ administratorLabel(workspace?.policy?.updatedBy) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatDateTime(workspace?.policy?.updatedAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="修改原因" :span="2">
              {{ workspace?.policy?.reason || '—' }}
            </el-descriptions-item>
          </el-descriptions>
          <el-alert
            v-if="workspace?.readiness && !workspace.readiness.ready"
            :title="`索引尚不可用于校验：${workspace.readiness.blockers?.join('；') || '请补齐菜品索引'}`"
            type="error"
            :closable="false"
            show-icon
            class="mt-4"
          />
          <el-alert
            v-else-if="workspace?.readiness?.ready"
            title="索引已满足开启条件；线上开启后，破坏可用性的目录修改会被阻止。"
            type="success"
            :closable="false"
            show-icon
            class="mt-4"
          />
          <el-input
            v-if="canUpdate"
            v-model.trim="policyReason"
            type="textarea"
            :rows="2"
            maxlength="200"
            show-word-limit
            placeholder="填写修改原因（必填）"
            class="mt-4"
          />
          <div class="policy-actions">
            <el-button
              v-if="canUpdate"
              type="primary"
              :loading="policySaving"
              :disabled="policyEnabled && !workspace?.readiness?.ready"
              @click="savePolicy"
            >
              保存并立即生效
            </el-button>
          </div>
        </div>

        <div class="metric-grid">
          <div class="gva-card metric-card">
            <span>菜品索引</span>
            <strong>{{ workspace?.enabledDishCount ?? 0 }}</strong>
            <small>已启用 / 共 {{ workspace?.dishCount ?? 0 }}</small>
          </div>
          <div class="gva-card metric-card">
            <span>食材词库</span>
            <strong>{{ workspace?.enabledIngredientCount ?? 0 }}</strong>
            <small>已启用 / 共 {{ workspace?.ingredientCount ?? 0 }}</small>
          </div>
          <div class="gva-card metric-card">
            <span>开启条件</span>
            <strong>
              {{ workspace?.readiness?.enabledDishCount ?? 0 }}
              / {{ workspace?.readiness?.requiredMinimumDishCount ?? 6 }}
            </strong>
            <small>
              无效引用 {{ workspace?.readiness?.invalidReferenceCount ?? 0 }}，
              名称冲突 {{ workspace?.readiness?.nameCollisionCount ?? 0 }}
            </small>
          </div>
        </div>
      </div>

      <div class="gva-card">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <el-tab-pane label="菜品索引" name="dishes">
            <div class="toolbar">
              <div class="filters">
                <el-input
                  v-model.trim="dishQuery.keyword"
                  clearable
                  placeholder="搜索菜名或别名"
                  @keyup.enter="searchDishes"
                  @clear="searchDishes"
                />
                <el-select v-model="dishQuery.enabled" clearable placeholder="全部状态">
                  <el-option label="已启用" :value="true" />
                  <el-option label="已停用" :value="false" />
                </el-select>
                <el-button type="primary" @click="searchDishes">查询</el-button>
                <el-button @click="resetDishes">重置</el-button>
              </div>
              <el-button v-if="canUpdate" type="primary" @click="openDishEditor()">
                新增菜品索引
              </el-button>
            </div>
            <el-table
              v-loading="dishLoading"
              :data="dishes"
              row-key="id"
              empty-text="暂无菜品索引"
            >
              <el-table-column label="标准菜名" prop="name" min-width="150" />
              <el-table-column label="别名" min-width="180">
                <template #default="{ row }">
                  {{ row.aliases?.join('、') || '—' }}
                </template>
              </el-table-column>
              <el-table-column label="菜系" prop="cuisine" width="120" />
              <el-table-column label="分类" prop="category" width="120" />
              <el-table-column label="配料范围" min-width="220">
                <template #default="{ row }">
                  <el-tag
                    v-for="item in row.ingredients"
                    :key="item.ingredientId"
                    class="mr-1"
                    :type="item.required ? 'success' : 'info'"
                    size="small"
                  >
                    {{ item.name }}{{ item.required ? '（必需）' : '' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="来源" prop="sourceName" min-width="150" />
              <el-table-column label="状态" width="90">
                <template #default="{ row }">
                  <el-tag :type="row.enabled ? 'success' : 'info'">
                    {{ row.enabled ? '启用' : '停用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="更新时间" width="180">
                <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
              </el-table-column>
              <el-table-column v-if="canUpdate" label="操作" width="90" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="openDishEditor(row)">编辑</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="pagination">
              <el-pagination
                v-model:current-page="dishQuery.page"
                v-model:page-size="dishQuery.pageSize"
                :page-sizes="[20, 50, 100]"
                :total="dishTotal"
                layout="total, sizes, prev, pager, next"
                @size-change="loadDishes"
                @current-change="loadDishes"
              />
            </div>
          </el-tab-pane>

          <el-tab-pane label="食材词库" name="ingredients">
            <div class="toolbar">
              <div class="filters">
                <el-input
                  v-model.trim="ingredientQuery.keyword"
                  clearable
                  placeholder="搜索食材名或别名"
                  @keyup.enter="searchIngredients"
                  @clear="searchIngredients"
                />
                <el-select v-model="ingredientQuery.enabled" clearable placeholder="全部状态">
                  <el-option label="已启用" :value="true" />
                  <el-option label="已停用" :value="false" />
                </el-select>
                <el-button type="primary" @click="searchIngredients">查询</el-button>
                <el-button @click="resetIngredients">重置</el-button>
              </div>
              <el-button v-if="canUpdate" type="primary" @click="openIngredientEditor()">
                新增标准食材
              </el-button>
            </div>
            <el-table
              v-loading="ingredientLoading"
              :data="ingredients"
              row-key="id"
              empty-text="暂无标准食材"
            >
              <el-table-column label="标准名称" prop="name" min-width="180" />
              <el-table-column label="别名" min-width="260">
                <template #default="{ row }">
                  {{ row.aliases?.join('、') || '—' }}
                </template>
              </el-table-column>
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.enabled ? 'success' : 'info'">
                    {{ row.enabled ? '启用' : '停用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="版本" width="90">
                <template #default="{ row }">v{{ row.version }}</template>
              </el-table-column>
              <el-table-column label="更新时间" width="180">
                <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
              </el-table-column>
              <el-table-column v-if="canUpdate" label="操作" width="90" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" @click="openIngredientEditor(row)">
                    编辑
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="pagination">
              <el-pagination
                v-model:current-page="ingredientQuery.page"
                v-model:page-size="ingredientQuery.pageSize"
                :page-sizes="[20, 50, 100]"
                :total="ingredientTotal"
                layout="total, sizes, prev, pager, next"
                @size-change="loadIngredients"
                @current-change="loadIngredients"
              />
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <el-dialog
      v-model="ingredientEditorVisible"
      :title="ingredientForm.id ? '编辑标准食材' : '新增标准食材'"
      width="560px"
      destroy-on-close
    >
      <el-form
        ref="ingredientFormRef"
        :model="ingredientForm"
        :rules="ingredientRules"
        label-width="90px"
      >
        <el-form-item label="标准名称" prop="name">
          <el-input v-model.trim="ingredientForm.name" maxlength="80" />
        </el-form-item>
        <el-form-item label="别名">
          <el-input
            v-model.trim="ingredientForm.aliasesText"
            placeholder="多个别名用逗号分隔"
            maxlength="500"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="ingredientForm.enabled" active-text="启用" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="变更原因" prop="reason">
          <el-input
            v-model.trim="ingredientForm.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ingredientEditorVisible = false">取消</el-button>
        <el-button type="primary" :loading="ingredientSaving" @click="saveIngredient">
          保存
        </el-button>
      </template>
    </el-dialog>

    <el-drawer
      v-model="dishEditorVisible"
      :title="dishForm.id ? '编辑菜品索引' : '新增菜品索引'"
      size="680px"
      destroy-on-close
    >
      <el-form ref="dishFormRef" :model="dishForm" :rules="dishRules" label-width="100px">
        <el-form-item label="标准菜名" prop="name">
          <el-input v-model.trim="dishForm.name" maxlength="120" />
        </el-form-item>
        <el-form-item label="别名">
          <el-input
            v-model.trim="dishForm.aliasesText"
            placeholder="多个别名用逗号分隔"
            maxlength="800"
          />
        </el-form-item>
        <el-form-item label="菜系" prop="cuisine">
          <el-input v-model.trim="dishForm.cuisine" maxlength="80" placeholder="如：川菜、粤菜、家常菜" />
        </el-form-item>
        <el-form-item label="分类" prop="categoryId">
          <el-select v-model="dishForm.categoryId" class="w-full">
            <el-option
              v-for="category in workspace?.categories || []"
              :key="category.id"
              :label="category.name"
              :value="category.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="依据来源" prop="sourceName">
          <el-input v-model.trim="dishForm.sourceName" maxlength="160" />
        </el-form-item>
        <el-form-item label="来源链接">
          <el-input v-model.trim="dishForm.sourceUrl" maxlength="500" />
        </el-form-item>
        <el-form-item label="配料范围" prop="ingredients">
          <div class="ingredient-editor">
            <div
              v-for="(item, index) in dishForm.ingredients"
              :key="index"
              class="ingredient-row"
            >
              <el-select
                v-model="item.ingredientId"
                filterable
                placeholder="选择食材"
                class="ingredient-select"
              >
                <el-option
                  v-for="ingredient in ingredientOptions"
                  :key="ingredient.id"
                  :label="ingredient.name"
                  :value="ingredient.id"
                />
              </el-select>
              <el-checkbox v-model="item.required">必需</el-checkbox>
              <el-button link type="danger" @click="removeDishIngredient(index)">移除</el-button>
            </div>
            <el-button @click="addDishIngredient">添加食材</el-button>
          </div>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="dishForm.enabled" active-text="启用" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="变更原因" prop="reason">
          <el-input
            v-model.trim="dishForm.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dishEditorVisible = false">取消</el-button>
        <el-button type="primary" :loading="dishSaving" @click="saveDish">保存</el-button>
      </template>
    </el-drawer>

  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import { formatOrderFoodDateTime as formatDate } from '@/view/orderFood/utils/time'
import {
  createStandardDish,
  createStandardIngredient,
  getStandardDishes,
  getStandardIngredients,
  getSuggestionCatalogWorkspace,
  updateSuggestionValidationPolicy,
  updateStandardDish,
  updateStandardIngredient
} from '@/api/orderfood/suggestion-catalog'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({ name: 'OrderFoodSuggestionCatalog' })

const btnAuth = useBtnAuth()
const hasBtnPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))
const canUpdate = computed(() =>
  hasBtnPermission('orderfood:suggestion-catalog:update', 'suggestionCatalogUpdate')
)

const loading = ref(false)
const pageError = ref('')
const workspace = ref(null)
const policyEnabled = ref(false)
const policyReason = ref('')
const policySaving = ref(false)
const activeTab = ref('dishes')
const formatDateTime = (value) => (value ? formatDate(value) : '—')
const administratorLabel = (administrator) =>
  administrator?.nickname || administrator?.username || administrator?.id || '—'

const dishes = ref([])
const dishTotal = ref(0)
const dishLoading = ref(false)
const dishQuery = ref({ page: 1, pageSize: 20, keyword: '', enabled: null, sortOrder: 'asc' })
const ingredients = ref([])
const ingredientOptions = ref([])
const ingredientTotal = ref(0)
const ingredientLoading = ref(false)
const ingredientQuery = ref({ page: 1, pageSize: 20, keyword: '', enabled: null, sortOrder: 'asc' })

const splitAliases = (value) =>
  [...new Set(String(value || '').split(/[,，]/).map((item) => item.trim()).filter(Boolean))]

const loadWorkspace = async () => {
  const response = await getSuggestionCatalogWorkspace()
  workspace.value = unwrapOrderFoodResponse(response, '加载标准菜品索引失败')
  policyEnabled.value = workspace.value?.policy?.catalogValidationEnabled ?? false
  policyReason.value = ''
}

const loadDishes = async () => {
  dishLoading.value = true
  try {
    const response = await getStandardDishes(dishQuery.value)
    const data = unwrapOrderFoodResponse(response, '加载菜品索引失败')
    dishes.value = data.list || []
    dishTotal.value = Number(data.total || 0)
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '加载菜品索引失败'))
  } finally {
    dishLoading.value = false
  }
}

const loadIngredients = async () => {
  ingredientLoading.value = true
  try {
    const response = await getStandardIngredients(ingredientQuery.value)
    const data = unwrapOrderFoodResponse(response, '加载食材词库失败')
    ingredients.value = data.list || []
    ingredientTotal.value = Number(data.total || 0)
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '加载食材词库失败'))
  } finally {
    ingredientLoading.value = false
  }
}

const loadIngredientOptions = async () => {
  const response = await getStandardIngredients({
    page: 1,
    pageSize: 100,
    enabled: true,
    sortOrder: 'asc'
  })
  const data = unwrapOrderFoodResponse(response, '加载可用食材失败')
  ingredientOptions.value = data.list || []
}

const loadPage = async () => {
  loading.value = true
  pageError.value = ''
  try {
    await Promise.all([loadWorkspace(), loadDishes(), loadIngredients()])
  } catch (error) {
    pageError.value = getOrderFoodErrorMessage(error, '加载标准菜品索引失败')
  } finally {
    loading.value = false
  }
}

const searchDishes = () => {
  dishQuery.value.page = 1
  loadDishes()
}
const resetDishes = () => {
  dishQuery.value = {
    page: 1,
    pageSize: dishQuery.value.pageSize,
    keyword: '',
    enabled: null,
    sortOrder: 'asc'
  }
  loadDishes()
}
const searchIngredients = () => {
  ingredientQuery.value.page = 1
  loadIngredients()
}
const resetIngredients = () => {
  ingredientQuery.value = {
    page: 1,
    pageSize: ingredientQuery.value.pageSize,
    keyword: '',
    enabled: null,
    sortOrder: 'asc'
  }
  loadIngredients()
}
const handleTabChange = (name) => {
  if (name === 'dishes') loadDishes()
  if (name === 'ingredients') loadIngredients()
}

const savePolicy = async () => {
  if (policyEnabled.value && !workspace.value?.readiness?.ready) {
    ElMessage.warning('请先处理索引开启条件中的阻塞项')
    return
  }
  if (policyReason.value.trim().length < 2) {
    ElMessage.warning('请填写至少 2 个字的修改原因')
    return
  }
  policySaving.value = true
  try {
    const response = await updateSuggestionValidationPolicy({
      catalogValidationEnabled: policyEnabled.value,
      reason: policyReason.value,
      expectedVersion: workspace.value?.policy?.version
    })
    unwrapOrderFoodResponse(response, '保存校验策略失败')
    ElMessage.success('校验策略已保存并生效')
    await loadWorkspace()
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('配置已被其他管理员修改，页面将重新加载')
      await loadWorkspace()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存校验策略失败'))
    }
  } finally {
    policySaving.value = false
  }
}

const ingredientEditorVisible = ref(false)
const ingredientSaving = ref(false)
const ingredientFormRef = ref()
const ingredientForm = ref({})
const ingredientRules = {
  name: [{ required: true, message: '请输入标准名称', trigger: 'blur' }],
  reason: [
    { required: true, message: '请输入变更原因', trigger: 'blur' },
    { min: 4, max: 200, message: '原因长度为 4-200 字', trigger: 'blur' }
  ]
}
const openIngredientEditor = (row) => {
  ingredientForm.value = {
    id: row?.id || '',
    name: row?.name || '',
    aliasesText: row?.aliases?.join('，') || '',
    enabled: row?.enabled ?? true,
    expectedVersion: row?.version || null,
    reason: ''
  }
  ingredientEditorVisible.value = true
}
const saveIngredient = async () => {
  const valid = await ingredientFormRef.value?.validate().catch(() => false)
  if (!valid) return
  ingredientSaving.value = true
  const data = {
    name: ingredientForm.value.name,
    aliases: splitAliases(ingredientForm.value.aliasesText),
    enabled: ingredientForm.value.enabled,
    expectedVersion: ingredientForm.value.expectedVersion,
    reason: ingredientForm.value.reason
  }
  try {
    const response = ingredientForm.value.id
      ? await updateStandardIngredient(ingredientForm.value.id, data)
      : await createStandardIngredient(data)
    unwrapOrderFoodResponse(response, '保存标准食材失败')
    ElMessage.success('标准食材已保存')
    ingredientEditorVisible.value = false
    await Promise.all([loadWorkspace(), loadIngredients()])
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('数据已变化，请刷新后重试')
      await loadIngredients()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存标准食材失败'))
    }
  } finally {
    ingredientSaving.value = false
  }
}

const dishEditorVisible = ref(false)
const dishSaving = ref(false)
const dishFormRef = ref()
const dishForm = ref({})
const dishRules = {
  name: [{ required: true, message: '请输入标准菜名', trigger: 'blur' }],
  cuisine: [{ required: true, message: '请输入菜系', trigger: 'blur' }],
  categoryId: [{ required: true, message: '请选择分类', trigger: 'change' }],
  sourceName: [{ required: true, message: '请输入依据来源', trigger: 'blur' }],
  ingredients: [{
    type: 'array',
    required: true,
    min: 1,
    message: '至少配置一种食材',
    trigger: 'change'
  }],
  reason: [
    { required: true, message: '请输入变更原因', trigger: 'blur' },
    { min: 4, max: 200, message: '原因长度为 4-200 字', trigger: 'blur' }
  ]
}
const openDishEditor = async (row) => {
  try {
    await loadIngredientOptions()
  } catch (error) {
    ElMessage.error(getOrderFoodErrorMessage(error, '加载可用食材失败'))
    return
  }
  dishForm.value = {
    id: row?.id || '',
    name: row?.name || '',
    aliasesText: row?.aliases?.join('，') || '',
    cuisine: row?.cuisine || '',
    categoryId: row?.categoryId || '',
    sourceName: row?.sourceName || '',
    sourceUrl: row?.sourceUrl || '',
    enabled: row?.enabled ?? true,
    ingredients: (row?.ingredients || []).map((item) => ({
      ingredientId: item.ingredientId,
      required: item.required
    })),
    expectedVersion: row?.version || null,
    reason: ''
  }
  dishEditorVisible.value = true
}
const addDishIngredient = () => {
  dishForm.value.ingredients.push({ ingredientId: '', required: false })
}
const removeDishIngredient = (index) => {
  dishForm.value.ingredients.splice(index, 1)
}
const saveDish = async () => {
  const valid = await dishFormRef.value?.validate().catch(() => false)
  if (!valid) return
  const ingredientIds = dishForm.value.ingredients.map((item) => item.ingredientId)
  if (ingredientIds.some((id) => !id) || new Set(ingredientIds).size !== ingredientIds.length) {
    ElMessage.warning('请选择食材，并避免重复')
    return
  }
  dishSaving.value = true
  const data = {
    name: dishForm.value.name,
    aliases: splitAliases(dishForm.value.aliasesText),
    cuisine: dishForm.value.cuisine,
    categoryId: dishForm.value.categoryId,
    sourceName: dishForm.value.sourceName,
    sourceUrl: dishForm.value.sourceUrl || null,
    enabled: dishForm.value.enabled,
    ingredients: dishForm.value.ingredients.map((item, index) => ({
      ingredientId: item.ingredientId,
      required: item.required,
      sortOrder: index + 1
    })),
    expectedVersion: dishForm.value.expectedVersion,
    reason: dishForm.value.reason
  }
  try {
    const response = dishForm.value.id
      ? await updateStandardDish(dishForm.value.id, data)
      : await createStandardDish(data)
    unwrapOrderFoodResponse(response, '保存菜品索引失败')
    ElMessage.success('菜品索引已保存')
    dishEditorVisible.value = false
    await Promise.all([loadWorkspace(), loadDishes()])
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('数据已变化，请刷新后重试')
      await loadDishes()
    } else {
      ElMessage.error(getOrderFoodErrorMessage(error, '保存菜品索引失败'))
    }
  } finally {
    dishSaving.value = false
  }
}

loadPage()
</script>

<style scoped lang="scss">
.page-header,
.section-header,
.toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.page-title {
  margin: 0;
  color: #1f2937;
  font-size: 22px;
  font-weight: 650;
}

.page-subtitle,
.section-note {
  margin: 6px 0 0;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.6;
}

.overview-grid {
  display: grid;
  grid-template-columns: minmax(520px, 1fr) 300px;
  gap: 16px;
}

.section-title,
.control-title {
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

.policy-control {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin: 22px 0 18px;
  padding: 18px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #f9fafb;
}

.policy-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 20px;
  color: #6b7280;
  font-size: 13px;
}

.policy-meta {
  margin-top: 4px;
}

.policy-actions {
  margin-top: 18px;
}

.metric-grid {
  display: grid;
  gap: 16px;
}

.metric-card {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.metric-card span,
.metric-card small {
  color: #6b7280;
}

.metric-card strong {
  margin: 8px 0;
  color: #15803d;
  font-size: 34px;
}

.toolbar {
  margin: 8px 0 18px;
}

.filters {
  display: flex;
  gap: 10px;
}

.filters .el-input {
  width: 240px;
}

.filters .el-select {
  width: 130px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 18px;
}

.ingredient-editor {
  width: 100%;
}

.ingredient-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.ingredient-select {
  flex: 1;
}

@media (max-width: 980px) {
  .overview-grid {
    grid-template-columns: 1fr;
  }

  .page-header,
  .toolbar,
  .filters {
    flex-direction: column;
  }
}
</style>
