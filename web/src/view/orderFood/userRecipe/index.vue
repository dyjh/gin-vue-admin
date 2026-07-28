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
              @click="openRecipeDetail(row)"
            >
              详情
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
      size="min(920px, 94vw)"
      class="recipe-detail-drawer"
      destroy-on-close
      @closed="resetDetail"
    >
      <template #header>
        <div class="recipe-detail-header">
          <div class="recipe-detail-identity">
            <el-image
              v-if="detail?.coverUrl || activeRecipe?.coverUrl"
              :src="detail?.coverUrl || activeRecipe?.coverUrl"
              fit="cover"
              class="recipe-detail-header-cover"
            />
            <div v-else class="recipe-detail-header-placeholder">谱</div>
            <div class="recipe-detail-heading">
              <span class="recipe-detail-kicker">用户菜谱</span>
              <div class="recipe-detail-title">
                {{ detail?.name || activeRecipe?.name || '菜谱详情' }}
              </div>
              <div class="recipe-detail-id">
                {{ detail?.id || activeRecipe?.id || '—' }}
              </div>
            </div>
          </div>
          <div v-if="detail" class="recipe-detail-actions">
            <el-button
              v-if="canReadUsers && detail.owner?.id"
              plain
              @click="openUser(detail.owner.id)"
            >
              查看用户
            </el-button>
            <el-button
              v-if="canReadUserDishes"
              plain
              @click="openRecipeDishes(detail.id)"
            >
              查看所含菜品
            </el-button>
            <el-button
              v-if="canGovernanceExecute && detail.deletionStatus !== 'deleted'"
              type="danger"
              plain
              @click="openGovernanceDialog(detail)"
            >
              违规处理
            </el-button>
          </div>
        </div>
      </template>

      <div v-loading="detailLoading" class="recipe-detail-body">
        <el-alert
          v-if="detailError"
          :title="detailError"
          type="error"
          show-icon
          :closable="false"
          class="recipe-detail-error"
        >
          <template #default>
            <el-button link type="primary" @click="loadRecipeDetail">
              重新加载
            </el-button>
          </template>
        </el-alert>

        <el-tabs
          v-else-if="detail"
          v-model="detailTab"
          class="recipe-detail-tabs"
        >
          <el-tab-pane label="详情" name="detail">
            <div class="recipe-detail-panel">
              <section class="recipe-summary-card">
                <div class="recipe-cover-shell">
                  <el-image
                    v-if="detail.coverUrl"
                    :src="detail.coverUrl"
                    fit="cover"
                    class="recipe-cover"
                    :preview-src-list="[detail.coverUrl]"
                    preview-teleported
                  />
                  <div v-else class="recipe-cover-placeholder">
                    <span>谱</span>
                    <small>暂无封面</small>
                  </div>
                </div>

                <div class="recipe-summary-content">
                  <div class="recipe-summary-topline">
                    <div>
                      <span class="recipe-summary-label">所属用户</span>
                      <strong>{{ detail.owner?.nickname || '未知用户' }}</strong>
                      <span class="recipe-owner-id">
                        {{ detail.owner?.id || '—' }}
                      </span>
                    </div>
                    <div class="recipe-status-group">
                      <el-tag
                        size="small"
                        effect="plain"
                        :type="contentStateType(detail.contentState)"
                      >
                        {{ contentStateLabel(detail.contentState) }}
                      </el-tag>
                      <el-tag
                        size="small"
                        effect="plain"
                        :type="
                          detail.deletionStatus === 'deleted'
                            ? 'danger'
                            : 'success'
                        "
                      >
                        {{
                          detail.deletionStatus === 'deleted'
                            ? '已软删除'
                            : '正常'
                        }}
                      </el-tag>
                    </div>
                  </div>

                  <div class="recipe-metric-grid">
                    <div class="recipe-metric-item">
                      <span>菜品总数</span>
                      <strong>{{ detail.dishCount }}</strong>
                    </div>
                    <div
                      class="recipe-metric-item"
                      :class="{ 'has-warning': detail.unavailableDishCount > 0 }"
                    >
                      <span>不可用菜品</span>
                      <strong>{{ detail.unavailableDishCount }}</strong>
                    </div>
                    <div class="recipe-metric-item">
                      <span>当前版本</span>
                      <strong>V{{ detail.version }}</strong>
                    </div>
                  </div>
                </div>
              </section>

              <div class="recipe-access-note">
                <span class="recipe-access-dot"></span>
                <div>
                  <strong>只读访问已记录</strong>
                  <p>
                    本页包含用户备注和有序菜品清单，仅供管理查看，不能代用户编辑或调序。
                  </p>
                </div>
              </div>

              <div class="recipe-info-layout">
                <section class="recipe-section-card">
                  <header class="recipe-section-header">
                    <div>
                      <h3>基础信息</h3>
                      <p>菜谱状态与时间记录</p>
                    </div>
                  </header>
                  <dl class="recipe-info-grid">
                    <div class="recipe-info-item">
                      <dt>内容状态</dt>
                      <dd>{{ contentStateLabel(detail.contentState) }}</dd>
                    </div>
                    <div class="recipe-info-item">
                      <dt>删除状态</dt>
                      <dd>
                        {{
                          detail.deletionStatus === 'deleted'
                            ? '已软删除'
                            : '未删除'
                        }}
                      </dd>
                    </div>
                    <div class="recipe-info-item">
                      <dt>创建时间</dt>
                      <dd>{{ formatDateTime(detail.createdAt) }}</dd>
                    </div>
                    <div class="recipe-info-item">
                      <dt>更新时间</dt>
                      <dd>{{ formatDateTime(detail.updatedAt) }}</dd>
                    </div>
                  </dl>
                </section>

                <section class="recipe-section-card recipe-note-card">
                  <header class="recipe-section-header">
                    <div>
                      <h3>菜谱备注</h3>
                      <p>{{ detail.note ? '用户填写的菜谱说明' : '用户未填写备注' }}</p>
                    </div>
                  </header>
                  <div
                    class="recipe-note-content"
                    :class="{ 'is-empty': !detail.note }"
                  >
                    {{ detail.note || '暂无备注' }}
                  </div>
                </section>
              </div>

              <section class="recipe-section-card recipe-dishes-card">
                <header class="recipe-section-header">
                  <div>
                    <h3>菜品清单</h3>
                    <p>按用户设置的顺序展示，共 {{ orderedDishes.length }} 道菜</p>
                  </div>
                  <el-tag size="small" effect="plain">只读</el-tag>
                </header>

                <el-empty
                  v-if="!orderedDishes.length"
                  description="该菜谱暂无菜品"
                  :image-size="68"
                  class="recipe-dishes-empty"
                />
                <div v-else class="recipe-dish-list">
                  <article
                    v-for="item in orderedDishes"
                    :key="item.relationId"
                    class="recipe-dish-item"
                  >
                    <span class="recipe-dish-order">
                      {{ String(item.sortOrder).padStart(2, '0') }}
                    </span>
                    <el-image
                      v-if="item.coverUrl"
                      :src="item.coverUrl"
                      fit="cover"
                      class="recipe-dish-cover"
                      :preview-src-list="[item.coverUrl]"
                      preview-teleported
                    />
                    <div v-else class="recipe-dish-cover is-placeholder">
                      无图
                    </div>
                    <div class="recipe-dish-main">
                      <div class="recipe-dish-title-row">
                        <strong>{{ item.dishName }}</strong>
                        <el-tag
                          size="small"
                          effect="plain"
                          :type="dishItemStatusType(item.dishStatus)"
                        >
                          {{ dishItemStatusLabel(item.dishStatus) }}
                        </el-tag>
                      </div>
                      <div class="recipe-dish-meta">
                        <span>{{ sourceTypeLabel(item.sourceType) }}</span>
                        <span>加入于 {{ formatDateTime(item.addedAt) }}</span>
                      </div>
                      <p v-if="item.unavailableReason" class="recipe-dish-warning">
                        {{ item.unavailableReason }}
                      </p>
                    </div>
                    <el-button
                      v-if="canReadUserDishes"
                      link
                      type="primary"
                      @click="openDish(item.dishId)"
                    >
                      查看菜品
                    </el-button>
                  </article>
                </div>
              </section>

              <section
                v-if="detail.deletionStatus === 'deleted'"
                class="recipe-section-card recipe-deletion-card"
              >
                <header class="recipe-section-header">
                  <div>
                    <h3>删除信息</h3>
                    <p>该菜谱的下线处理记录</p>
                  </div>
                  <el-tag type="danger" size="small" effect="plain">
                    已软删除
                  </el-tag>
                </header>
                <dl class="recipe-info-grid">
                  <div class="recipe-info-item">
                    <dt>删除时间</dt>
                    <dd>{{ formatDateTime(detail.deletedAt) }}</dd>
                  </div>
                  <div class="recipe-info-item">
                    <dt>删除人</dt>
                    <dd>
                      {{
                        detail.deletedBy?.nickname ||
                          detail.deletedBy?.username ||
                          detail.deletedBy?.id ||
                          '未记录'
                      }}
                    </dd>
                  </div>
                  <div class="recipe-info-item">
                    <dt>违规类型</dt>
                    <dd>{{ detail.deletedViolationType || '未记录' }}</dd>
                  </div>
                  <div class="recipe-info-item">
                    <dt>删除原因</dt>
                    <dd>{{ detail.deletedReason || '未记录' }}</dd>
                  </div>
                </dl>
              </section>

              <footer class="recipe-audit-footer">
                <div>
                  <span>违规处理</span>
                  <strong>{{ detail.governanceRecordCount }} 条</strong>
                  <el-button
                    v-if="canReadGovernance && detail.governanceRecordCount > 0"
                    link
                    type="primary"
                    @click="openGovernanceRecords(detail.id)"
                  >
                    查看记录
                  </el-button>
                </div>
                <div>
                  <span>操作审计</span>
                  <strong>{{ detail.auditLogCount }} 条</strong>
                </div>
                <div class="recipe-audit-id">
                  <span>本次访问审计 ID</span>
                  <strong>{{ detail.accessAuditId || '—' }}</strong>
                </div>
              </footer>
            </div>
          </el-tab-pane>
          <el-tab-pane
            v-if="canReadAudit"
            label="操作审计"
            name="audit"
            lazy
          >
            <div class="recipe-detail-panel">
              <AuditPanel target-type="recipe" :target-id="detail.id" />
            </div>
          </el-tab-pane>
        </el-tabs>
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
const detailTab = ref('detail')
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
  detailTab.value = 'detail'
}

const loadRecipeDetail = async () => {
  if (!canPrivateRead.value || !activeRecipe.value?.id) return
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getOrderFoodUserRecipeDetail(activeRecipe.value.id),
      '菜谱详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '菜谱详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const openRecipeDetail = async (row) => {
  if (!canPrivateRead.value) return
  activeRecipe.value = row
  detailVisible.value = true
  await loadRecipeDetail()
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
const dishItemStatusType = (value) =>
  ({ draft: 'info', usable: 'success', deleted: 'danger' })[value] || 'info'

getTableData().then(() => {
  if (route.query.openRecipeId) {
    openRecipeDetail({ id: routeQueryValue(route.query.openRecipeId) }).then(
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

<style scoped>
.recipe-detail-drawer {
  overflow: hidden;
  border-radius: 14px 0 0 14px;
  background: var(--el-fill-color-extra-light);
}

.recipe-detail-drawer :deep(.el-drawer__header) {
  margin: 0;
  padding: 16px 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}

.recipe-detail-drawer :deep(.el-drawer__body) {
  padding: 0;
  background: var(--el-fill-color-extra-light);
}

.recipe-detail-header {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding-right: 8px;
}

.recipe-detail-identity {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
}

.recipe-detail-header-cover,
.recipe-detail-header-placeholder {
  width: 48px;
  height: 48px;
  flex: none;
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 11px;
  box-shadow: 0 3px 10px rgb(31 42 61 / 7%);
}

.recipe-detail-header-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  border-color: var(--el-color-primary-light-8);
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 20px;
  font-weight: 700;
}

.recipe-detail-heading { min-width: 0; }
.recipe-detail-kicker {
  display: block;
  margin-bottom: 2px;
  color: var(--el-color-primary);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.12em;
}

.recipe-detail-title {
  overflow: hidden;
  color: var(--el-text-color-primary);
  font-size: 17px;
  font-weight: 650;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recipe-detail-id {
  overflow: hidden;
  margin-top: 3px;
  color: var(--el-text-color-placeholder);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recipe-detail-actions {
  display: flex;
  flex: none;
  align-items: center;
  gap: 8px;
}

.recipe-detail-actions .el-button + .el-button { margin-left: 0; }
.recipe-detail-body { min-height: 300px; }
.recipe-detail-error { margin: 20px; width: auto; }

.recipe-detail-tabs :deep(.el-tabs__header) {
  position: sticky;
  z-index: 3;
  top: 0;
  margin: 0;
  padding: 0 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}

.recipe-detail-tabs :deep(.el-tabs__nav-wrap::after) { display: none; }
.recipe-detail-tabs :deep(.el-tabs__item) {
  height: 46px;
  padding: 0 18px;
}

.recipe-detail-tabs :deep(.el-tabs__content) { overflow: visible; }
.recipe-detail-panel { padding: 18px 20px 26px; }

.recipe-summary-card {
  display: grid;
  grid-template-columns: 146px minmax(0, 1fr);
  gap: 20px;
  padding: 18px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  background:
    linear-gradient(145deg, rgb(255 255 255 / 98%), var(--el-color-primary-light-9));
  box-shadow: 0 4px 16px rgb(31 42 61 / 5%);
}

.recipe-cover-shell {
  width: 146px;
  height: 146px;
  overflow: hidden;
  border: 4px solid var(--el-bg-color);
  border-radius: 12px;
  box-shadow: 0 7px 20px rgb(31 42 61 / 11%);
}

.recipe-cover { width: 100%; height: 100%; }
.recipe-cover-placeholder {
  display: flex;
  width: 100%;
  height: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 7px;
  background:
    linear-gradient(145deg, var(--el-color-primary-light-9), var(--el-fill-color-light));
  color: var(--el-color-primary);
}

.recipe-cover-placeholder span { font-size: 34px; font-weight: 700; }
.recipe-cover-placeholder small {
  color: var(--el-text-color-placeholder);
  font-size: 11px;
}

.recipe-summary-content {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: space-between;
  gap: 18px;
}

.recipe-summary-topline {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.recipe-summary-topline > div:first-child {
  display: flex;
  min-width: 0;
  flex-direction: column;
}

.recipe-summary-label {
  margin-bottom: 5px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
}

.recipe-summary-topline strong {
  overflow: hidden;
  color: var(--el-text-color-primary);
  font-size: 16px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recipe-owner-id {
  overflow: hidden;
  margin-top: 5px;
  color: var(--el-text-color-placeholder);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recipe-status-group {
  display: flex;
  flex: none;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 7px;
}

.recipe-metric-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 9px;
  background: var(--el-bg-color);
}

.recipe-metric-item {
  display: flex;
  min-height: 67px;
  flex-direction: column;
  justify-content: center;
  padding: 10px 14px;
  border-right: 1px solid var(--el-border-color-extra-light);
}

.recipe-metric-item:last-child { border-right: 0; }
.recipe-metric-item span {
  color: var(--el-text-color-secondary);
  font-size: 11px;
}

.recipe-metric-item strong {
  margin-top: 5px;
  color: var(--el-text-color-primary);
  font-size: 20px;
  font-weight: 650;
  line-height: 1;
}

.recipe-metric-item.has-warning strong { color: var(--el-color-warning); }

.recipe-access-note {
  display: flex;
  align-items: flex-start;
  gap: 11px;
  margin-top: 14px;
  padding: 11px 13px;
  border: 1px solid var(--el-color-warning-light-8);
  border-radius: 9px;
  background: var(--el-color-warning-light-9);
}

.recipe-access-dot {
  width: 7px;
  height: 7px;
  flex: none;
  margin-top: 6px;
  border-radius: 50%;
  background: var(--el-color-warning);
  box-shadow: 0 0 0 4px rgb(230 162 60 / 12%);
}

.recipe-access-note strong {
  color: var(--el-text-color-primary);
  font-size: 12px;
  font-weight: 650;
}

.recipe-access-note p {
  margin: 2px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.55;
}

.recipe-info-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.85fr);
  gap: 14px;
  margin-top: 14px;
}

.recipe-section-card {
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-bg-color);
  box-shadow: 0 2px 8px rgb(31 42 61 / 3%);
}

.recipe-section-header {
  display: flex;
  min-height: 61px;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 12px 15px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.recipe-section-header h3 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 650;
  line-height: 1.4;
}

.recipe-section-header p {
  margin: 3px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 11px;
  line-height: 1.5;
}

.recipe-info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 20px;
  margin: 0;
  padding: 2px 15px 8px;
}

.recipe-info-item {
  display: grid;
  min-width: 0;
  grid-template-columns: 68px minmax(0, 1fr);
  align-items: center;
  min-height: 50px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.recipe-info-item:nth-last-child(-n + 2) { border-bottom: 0; }
.recipe-info-item dt {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.recipe-info-item dd {
  min-width: 0;
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 12px;
  line-height: 1.55;
}

.recipe-note-card {
  display: flex;
  min-height: 172px;
  flex-direction: column;
}

.recipe-note-content {
  flex: 1;
  padding: 14px 15px 18px;
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.75;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.recipe-note-content.is-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-placeholder);
}

.recipe-dishes-card,
.recipe-deletion-card { margin-top: 14px; }
.recipe-dishes-empty :deep(.el-empty) { padding: 26px 0; }
.recipe-dish-list { padding: 0 15px; }

.recipe-dish-item {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 13px;
  padding: 13px 0;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.recipe-dish-item:last-child { border-bottom: 0; }
.recipe-dish-order {
  width: 32px;
  flex: none;
  color: var(--el-color-primary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-align: center;
}

.recipe-dish-cover {
  width: 58px;
  height: 58px;
  flex: none;
  overflow: hidden;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 9px;
}

.recipe-dish-cover.is-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-placeholder);
  font-size: 11px;
}

.recipe-dish-main { min-width: 0; flex: 1; }
.recipe-dish-title-row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}

.recipe-dish-title-row strong {
  overflow: hidden;
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recipe-dish-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 7px 16px;
  margin-top: 6px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
}

.recipe-dish-warning {
  margin: 5px 0 0;
  color: var(--el-color-danger);
  font-size: 11px;
  line-height: 1.5;
}

.recipe-audit-footer {
  display: grid;
  grid-template-columns: 150px 120px minmax(0, 1fr);
  gap: 16px;
  margin-top: 14px;
  padding: 13px 15px;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 9px;
  background: var(--el-fill-color-light);
}

.recipe-audit-footer > div {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 7px;
}

.recipe-audit-footer span {
  color: var(--el-text-color-secondary);
  font-size: 11px;
}

.recipe-audit-footer strong {
  color: var(--el-text-color-primary);
  font-size: 11px;
  font-weight: 600;
}

.recipe-audit-id { justify-content: flex-end; }
.recipe-audit-id strong {
  overflow: hidden;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 720px) {
  .recipe-detail-header { align-items: flex-start; }
  .recipe-detail-actions { display: none; }
  .recipe-detail-panel { padding: 14px; }
  .recipe-summary-card {
    grid-template-columns: 96px minmax(0, 1fr);
    gap: 14px;
    padding: 14px;
  }

  .recipe-cover-shell { width: 96px; height: 96px; }
  .recipe-summary-topline { flex-direction: column; }
  .recipe-status-group { justify-content: flex-start; }
  .recipe-metric-item { min-height: 58px; padding: 8px; }
  .recipe-metric-item strong { font-size: 17px; }
  .recipe-info-layout { grid-template-columns: 1fr; }
  .recipe-info-grid { grid-template-columns: 1fr; }
  .recipe-info-item:nth-last-child(-n + 2) { border-bottom: 1px solid var(--el-border-color-extra-light); }
  .recipe-info-item:last-child { border-bottom: 0; }
  .recipe-audit-footer { grid-template-columns: 1fr; gap: 8px; }
  .recipe-audit-id { justify-content: flex-start; }
}
</style>
