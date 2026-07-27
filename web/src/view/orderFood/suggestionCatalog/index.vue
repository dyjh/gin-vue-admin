<template>
  <div class="suggestion-page">
    <div v-loading="loading" class="gva-table-box workspace-card mb-4">
      <div class="page-hero">
        <div class="page-header">
          <div class="page-heading">
            <h2 class="page-title">标准菜品索引</h2>
            <p class="page-subtitle">
              维护生成建议的菜名、菜系和配料校验依据。这里不是完整菜谱库，也不会进入平台推荐池。
            </p>
          </div>
          <div class="header-actions">
            <span
              class="readiness-pill"
              :class="workspace?.readiness?.ready ? 'is-ready' : 'is-blocked'"
            >
              <i />
              {{ workspace?.readiness?.ready ? '索引可用' : '索引待完善' }}
            </span>
            <el-button :loading="loading" @click="loadPage">刷新</el-button>
          </div>
        </div>
      </div>

      <el-alert
        v-if="pageError"
        :title="pageError"
        type="error"
        show-icon
        class="workspace-error"
      />

      <div class="overview-panel">
        <div class="policy-card">
          <div class="section-header">
            <div>
              <div class="section-title">生成建议校验</div>
              <div class="section-note">
                控制生成结果是否必须匹配标准菜品索引和食材词库。
              </div>
            </div>
            <span class="policy-version">
              v{{ workspace?.policy?.version ?? '—' }}
            </span>
          </div>

          <div class="policy-control">
            <div>
              <div class="control-title">生成结果必须匹配标准索引</div>
              <div class="section-note">
                <template v-if="policyEnabled">
                  校验失败时同一笔使用会再尝试 1 次，两次均失败才退还积分。
                </template>
                <template v-else>
                  基础结构校验继续生效，但不会因索引不匹配而拦截结果。
                </template>
              </div>
            </div>
            <div class="control-switch">
              <span>{{ policyEnabled ? '已开启' : '已关闭' }}</span>
              <el-switch v-model="policyEnabled" :disabled="!canUpdate" />
            </div>
          </div>

          <div
            v-if="workspace?.readiness"
            class="readiness-note"
            :class="workspace.readiness.ready ? 'is-ready' : 'is-blocked'"
          >
            <span class="status-mark" />
            <div>
              <strong>{{ workspace.readiness.ready ? '开启条件已满足' : '暂不可开启校验' }}</strong>
              <span v-if="workspace.readiness.ready">
                开启后，破坏索引可用性的目录修改会被阻止。
              </span>
              <span v-else>
                {{ workspace.readiness.blockers?.join('；') || '请补齐菜品索引' }}
              </span>
            </div>
          </div>

          <div v-if="canUpdate && policyChanged" class="policy-editor">
            <el-input
              v-model.trim="policyReason"
              type="textarea"
              :rows="2"
              maxlength="200"
              show-word-limit
              placeholder="填写本次修改原因（必填）"
            />
            <el-button
              type="primary"
              :loading="policySaving"
              :disabled="policyEnabled && !workspace?.readiness?.ready"
              @click="savePolicy"
            >
              保存并立即生效
            </el-button>
          </div>

          <details class="policy-details">
            <summary>配置详情</summary>
            <div class="policy-meta">
              <div class="meta-item">
                <span>固定额外重试</span>
                <strong>{{ workspace?.policy?.retryCount ?? 1 }} 次</strong>
              </div>
              <div class="meta-item">
                <span>更新人</span>
                <strong>{{ administratorLabel(workspace?.policy?.updatedBy) }}</strong>
              </div>
              <div class="meta-item">
                <span>更新时间</span>
                <strong>{{ formatDateTime(workspace?.policy?.updatedAt) }}</strong>
              </div>
              <div class="meta-item meta-reason">
                <span>最近修改原因</span>
                <strong>{{ workspace?.policy?.reason || '—' }}</strong>
              </div>
            </div>
          </details>
        </div>

        <div class="metric-panel">
          <div class="metric-heading">
            <strong>索引概况</strong>
            <span>生成校验的基础数据</span>
          </div>
          <div class="metric-row">
            <div>
              <span>菜品索引</span>
              <small>已启用 / 总数</small>
            </div>
            <strong>
              {{ workspace?.enabledDishCount ?? 0 }}
              <em>/ {{ workspace?.dishCount ?? 0 }}</em>
            </strong>
          </div>
          <div class="metric-row">
            <div>
              <span>食材词库</span>
              <small>已启用 / 总数</small>
            </div>
            <strong>
              {{ workspace?.enabledIngredientCount ?? 0 }}
              <em>/ {{ workspace?.ingredientCount ?? 0 }}</em>
            </strong>
          </div>
          <div class="metric-row">
            <div>
              <span>最低菜品数</span>
              <small>
                无效引用 {{ workspace?.readiness?.invalidReferenceCount ?? 0 }} ·
                名称冲突 {{ workspace?.readiness?.nameCollisionCount ?? 0 }}
              </small>
            </div>
            <strong>
              {{ workspace?.readiness?.enabledDishCount ?? 0 }}
              <em>/ {{ workspace?.readiness?.requiredMinimumDishCount ?? 6 }}</em>
            </strong>
          </div>
        </div>
      </div>
    </div>

    <div v-loading="loading">
      <div class="gva-table-box catalog-card">
        <el-tabs v-model="activeTab" class="catalog-tabs" @tab-change="handleTabChange">
          <el-tab-pane name="dishes">
            <template #label>
              <span class="tab-label">菜品索引 <em>{{ dishTotal }}</em></span>
            </template>
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

          <el-tab-pane name="ingredients">
            <template #label>
              <span class="tab-label">食材词库 <em>{{ ingredientTotal }}</em></span>
            </template>
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
const policyChanged = computed(() =>
  policyEnabled.value !== Boolean(workspace.value?.policy?.catalogValidationEnabled)
)
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
.suggestion-page {
  --catalog-blue: #2563eb;
  --catalog-border: #e4ebf4;
  --catalog-text: #243247;
  --catalog-muted: #6d7a8c;
}

.workspace-card {
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--catalog-border);
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(31, 65, 114, 0.04);
}

.page-hero {
  padding: 18px 20px;
  border-bottom: 1px solid #e8edf4;
}

.workspace-error {
  margin: 16px 20px 0;
}

.page-header,
.section-header,
.toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.page-heading {
  max-width: 760px;
}

.page-title {
  margin: 0;
  color: var(--catalog-text);
  font-size: 22px;
  font-weight: 650;
  line-height: 1.35;
}

.page-subtitle,
.section-note {
  margin: 6px 0 0;
  color: var(--catalog-muted);
  font-size: 13px;
  line-height: 1.65;
}

.header-actions,
.control-switch {
  display: flex;
  align-items: center;
  gap: 12px;
}

.readiness-pill {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  min-height: 30px;
  padding: 0 11px;
  border: 1px solid var(--catalog-border);
  border-radius: 6px;
  color: #657287;
  background: #fff;
  font-size: 12px;
  white-space: nowrap;
}

.readiness-pill i,
.status-mark {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #94a3b8;
}

.readiness-pill.is-ready {
  color: #1d5db8;
  border-color: #cfe0f8;
  background: #f4f8fe;
}

.readiness-pill.is-ready i {
  background: var(--catalog-blue);
}

.readiness-pill.is-blocked {
  color: #9a6415;
  border-color: #f0dfbf;
  background: #fffaf1;
}

.readiness-pill.is-blocked i {
  background: #d79a32;
}

.overview-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
}

.policy-card {
  min-width: 0;
  padding: 22px 20px 20px;
}

.section-header {
  align-items: center;
  padding-bottom: 18px;
  border-bottom: 1px solid #edf1f6;
}

.section-title,
.control-title {
  color: var(--catalog-text);
  font-size: 16px;
  font-weight: 600;
}

.policy-version {
  padding: 5px 9px;
  border-radius: 5px;
  color: #627086;
  background: #f4f6f9;
  font-size: 12px;
  white-space: nowrap;
}

.policy-control {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin: 0;
  padding: 17px 0 15px;
  border-bottom: 1px solid #edf1f6;
}

.control-switch > span {
  color: #526176;
  font-size: 13px;
  white-space: nowrap;
}

.readiness-note {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-top: 14px;
  padding: 8px 10px;
  border-left: 3px solid var(--catalog-blue);
  border-radius: 0 4px 4px 0;
  color: #5e6e83;
  background: #f6f9fd;
  font-size: 13px;
  line-height: 1.5;
}

.status-mark {
  flex: none;
  margin-top: 6px;
  background: var(--catalog-blue);
}

.readiness-note > div {
  min-width: 0;
}

.readiness-note strong,
.readiness-note > div > span {
  display: block;
}

.readiness-note strong {
  color: #345f95;
  font-weight: 600;
}

.readiness-note > div > span {
  margin-top: 2px;
}

.readiness-note.is-blocked {
  color: #7d684a;
  border-left-color: #d79a32;
  background: #fffaf3;
}

.readiness-note.is-blocked .status-mark {
  background: #d79a32;
}

.readiness-note.is-blocked strong {
  color: #9b6516;
}

.policy-details {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid #edf1f6;
}

.policy-details summary {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #617087;
  font-size: 13px;
  cursor: pointer;
  list-style: none;
  user-select: none;
}

.policy-details summary::-webkit-details-marker {
  display: none;
}

.policy-details summary::before {
  content: '+';
  width: 16px;
  height: 16px;
  border: 1px solid #d8e0eb;
  border-radius: 4px;
  color: #718096;
  line-height: 14px;
  text-align: center;
}

.policy-details[open] summary::before {
  content: '−';
}

.policy-meta {
  display: grid;
  grid-template-columns: 0.7fr 1fr 1.3fr;
  gap: 16px;
  padding-top: 16px;
}

.meta-item {
  min-width: 0;
}

.meta-item span,
.metric-row small,
.metric-heading span {
  display: block;
  color: #8490a1;
  font-size: 12px;
}

.meta-item strong {
  display: block;
  margin-top: 5px;
  overflow: hidden;
  color: #435168;
  font-size: 13px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta-reason {
  grid-column: 1 / -1;
}

.meta-reason strong {
  white-space: normal;
}

.policy-editor {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: end;
  gap: 12px;
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid #edf1f6;
}

.policy-editor .el-button {
  min-width: 132px;
}

.metric-panel {
  padding: 22px 20px;
  border-left: 1px solid var(--catalog-border);
  background: #f8fbff;
}

.metric-heading {
  margin-bottom: 12px;
}

.metric-heading strong {
  display: block;
  margin-bottom: 4px;
  color: var(--catalog-text);
  font-size: 15px;
  font-weight: 600;
}

.metric-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 82px;
  border-bottom: 1px solid #e5edf7;
}

.metric-row:last-child {
  border-bottom: 0;
}

.metric-row span {
  color: #46566d;
  font-size: 13px;
  font-weight: 500;
}

.metric-row small {
  margin-top: 6px;
  line-height: 1.45;
}

.metric-row > strong {
  flex: none;
  color: var(--catalog-blue);
  font-size: 28px;
  font-weight: 650;
  letter-spacing: -0.02em;
}

.metric-row em {
  color: #8290a3;
  font-size: 14px;
  font-style: normal;
  font-weight: 500;
}

.catalog-card {
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--catalog-border);
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(31, 65, 114, 0.04);
}

.catalog-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 18px;
}

.catalog-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background: #e8edf4;
}

.catalog-tabs :deep(.el-tabs__item) {
  height: 50px;
  color: #68768a;
  font-weight: 500;
}

.catalog-tabs :deep(.el-tabs__item.is-active) {
  color: var(--catalog-blue);
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}

.tab-label em {
  min-width: 22px;
  padding: 1px 6px;
  border-radius: 10px;
  color: #748297;
  background: #eef1f5;
  font-size: 11px;
  font-style: normal;
  line-height: 18px;
  text-align: center;
}

.catalog-tabs :deep(.el-tabs__item.is-active) .tab-label em {
  color: #245da8;
  background: #eaf2fd;
}

.toolbar {
  align-items: center;
  margin: 0;
  padding: 14px 18px;
  border-bottom: 1px solid #e8edf4;
  background: #fff;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.filters .el-input {
  width: 240px;
}

.filters .el-select {
  width: 130px;
}

.filters .el-button {
  margin-left: 0;
}

.catalog-card :deep(.el-table) {
  margin: 0;
  border: 0;
  border-radius: 0;
}

.catalog-card :deep(.el-table__header th.el-table__cell) {
  height: 44px;
  color: #536176;
  background: #f6f8fb;
  font-weight: 600;
}

.catalog-card :deep(.el-table__body td.el-table__cell) {
  padding: 12px 0;
}

.catalog-card :deep(.el-table__inner-wrapper::before) {
  display: none;
}

.catalog-card :deep(.el-tag) {
  border-radius: 4px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 16px 18px 18px;
  border-top: 1px solid #edf1f6;
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

@media (max-width: 820px) {
  .overview-panel {
    grid-template-columns: 1fr;
  }

  .metric-panel {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 18px;
    border-top: 1px solid var(--catalog-border);
    border-left: 0;
  }

  .metric-heading {
    grid-column: 1 / -1;
    margin-bottom: 0;
  }

  .metric-row {
    display: block;
    min-height: 0;
    padding: 0 0 6px;
    border-bottom: 0;
  }

  .metric-row > strong {
    display: block;
    margin-top: 10px;
  }
}

@media (max-width: 760px) {
  .page-header,
  .section-header,
  .policy-control,
  .toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .header-actions {
    justify-content: space-between;
  }

  .overview-panel,
  .policy-meta,
  .policy-editor {
    grid-template-columns: 1fr;
  }

  .policy-card {
    padding: 20px;
  }

  .control-switch {
    justify-content: space-between;
  }

  .metric-panel {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 10px;
    padding: 16px 20px;
  }

  .metric-heading {
    display: none;
  }

  .metric-row {
    padding: 0;
  }

  .metric-row small {
    display: none;
  }

  .metric-row > strong {
    margin-top: 6px;
    font-size: 22px;
  }

  .meta-reason {
    grid-column: 1;
  }

  .filters {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 116px;
    width: 100%;
  }

  .filters .el-input,
  .filters .el-select {
    width: 100%;
  }

  .toolbar > .el-button {
    width: 100%;
    margin-left: 0;
  }
}
</style>
