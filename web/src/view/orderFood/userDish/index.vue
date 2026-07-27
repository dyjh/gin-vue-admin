<template>
  <div class="order-food-list-page">
    <AdvancedSearchPanel>
      <el-alert
        v-if="lockedUserId"
        :title="`当前仅查看用户 ${lockedUserId} 的菜品`"
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
            placeholder="菜品名称"
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
        <el-form-item label="菜谱 ID">
          <el-input
            v-model.trim="searchInfo.recipeId"
            clearable
            :disabled="Boolean(lockedRecipeId)"
            placeholder="引用菜谱 ID"
          />
        </el-form-item>
        <el-form-item label="可用状态">
          <el-select v-model="searchInfo.status" clearable placeholder="全部" class="w-32">
            <el-option label="草稿" value="draft" />
            <el-option label="可用" value="usable" />
          </el-select>
        </el-form-item>
        <el-form-item label="删除状态">
          <el-select v-model="searchInfo.deletionStatus" class="w-32">
            <el-option label="未删除" value="active" />
            <el-option label="已删除" value="deleted" />
          </el-select>
        </el-form-item>
        <el-form-item label="允许发现">
          <el-select
            v-model="searchInfo.discoverable"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="是" :value="true" />
            <el-option label="否" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源类型">
          <el-select
            v-model="searchInfo.sourceType"
            clearable
            placeholder="全部"
            class="w-40"
          >
            <el-option label="用户手动创建" value="manual" />
            <el-option label="创作者菜品复制" value="creator_copy" />
            <el-option label="官方菜品复制" value="official_copy" />
            <el-option label="饭局建议复制" value="meal_suggestion_copy" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源锁定">
          <el-select
            v-model="searchInfo.sourceLocked"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="已锁定" :value="true" />
            <el-option label="未锁定" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类 ID">
          <el-input
            v-model.trim="searchInfo.categoryId"
            clearable
            placeholder="分类 ID"
          />
        </el-form-item>
        <el-form-item label="标签 ID">
          <el-input
            v-model.trim="searchInfo.tagIdsText"
            clearable
            placeholder="多个 ID 用逗号分隔"
          />
        </el-form-item>
        <el-form-item label="图片审核">
          <el-select
            v-model="searchInfo.mediaReviewStatus"
            clearable
            placeholder="全部"
            class="w-36"
          >
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="passed" />
            <el-option label="已拒绝" value="rejected" />
            <el-option label="审核失败" value="failed" />
            <el-option label="无需审核" value="not_required" />
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
        <el-form-item label="删除时间">
          <el-date-picker
            v-model="searchInfo.deletedRange"
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
              <el-option label="菜品名称" value="name" />
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
        本页只展示菜品摘要。私有简介、配料、步骤、来源链和引用位置仅在独立授权后加载。
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
        :empty-text="listError ? '加载失败' : '暂无用户菜品'"
      >
        <el-table-column label="菜品" min-width="230" fixed="left">
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
        <el-table-column label="所属用户" min-width="160">
          <template #default="{ row }">
            <div>{{ row.owner?.nickname || '未知用户' }}</div>
            <div class="text-xs text-gray-400">{{ row.owner?.id || '—' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="分类 / 标签" min-width="180">
          <template #default="{ row }">
            <div>{{ row.category?.name || '未分类' }}</div>
            <div class="mt-1 flex flex-wrap gap-1">
              <el-tag
                v-for="tag in row.tags || []"
                :key="tag.id"
                size="small"
                effect="plain"
              >
                {{ tag.name }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="内容状态" min-width="130">
          <template #default="{ row }">
            <div class="flex flex-col items-start gap-1">
              <el-tag :type="row.status === 'usable' ? 'success' : 'info'" size="small">
                {{ dishStatusLabel(row.status) }}
              </el-tag>
              <el-tag
                v-if="row.deletionStatus === 'deleted'"
                type="danger"
                size="small"
              >
                已软删除
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="发现状态" width="105">
          <template #default="{ row }">
            <el-tag :type="row.discoverable ? 'success' : 'info'" size="small">
              {{ row.discoverable ? '允许发现' : '不可发现' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="来源" min-width="160">
          <template #default="{ row }">
            <div>{{ sourceTypeLabel(row.sourceType) }}</div>
            <div class="text-xs text-gray-400">
              {{ row.sourceLocked ? '来源已锁定' : '来源未锁定' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="引用摘要" min-width="150">
          <template #default="{ row }">
            <div>菜谱 {{ row.recipeReferenceCount }}</div>
            <div class="text-xs text-gray-400">
              未确认饭局 {{ row.unconfirmedMealReferenceCount }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="图片审核" width="110">
          <template #default="{ row }">
            <el-tag :type="reviewStatusType(row.mediaReviewStatus)" size="small">
              {{ reviewStatusLabel(row.mediaReviewStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="推荐次数" prop="recommendationCount" width="95" />
        <el-table-column label="创建时间" min-width="165">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="165">
          <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="360" fixed="right">
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
              v-if="canReadDiscoverable && row.discoverable && row.deletionStatus !== 'deleted'"
              link
              type="primary"
              @click="openCandidate(row.id)"
            >
              转到候选池
            </el-button>
            <el-button
              v-if="canReadRecommendations && row.recommendationCount > 0"
              link
              type="primary"
              @click="openRecommendations(row.id)"
            >
              查看推荐
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
      size="860px"
      destroy-on-close
      @closed="resetDetail"
    >
      <template #header>
        <div>
          <div class="text-lg font-medium">{{ detail?.name || activeDish?.name || '菜品详情' }}</div>
          <div class="mt-1 text-xs text-gray-400">{{ activeDish?.id }}</div>
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

        <el-tabs
          v-else-if="detail"
          v-model="detailTab"
          @tab-change="handleDetailTabChange"
        >
          <el-tab-pane label="私有详情" name="detail">
            <el-alert
              title="此处包含用户私有内容，访问行为已记录审计。页面仅供查看，不提供代用户编辑、恢复、公开或调序。"
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
                class="h-36 w-36 shrink-0 rounded"
                :preview-src-list="[detail.coverUrl]"
                preview-teleported
              />
              <el-descriptions :column="2" border class="min-w-0 flex-1">
                <el-descriptions-item label="所属用户">
                  {{ detail.owner?.nickname }}（{{ detail.owner?.id }}）
                </el-descriptions-item>
                <el-descriptions-item label="份数">{{ detail.serving }}</el-descriptions-item>
                <el-descriptions-item label="状态">
                  {{ dishStatusLabel(detail.status) }} /
                  {{ detail.deletionStatus === 'deleted' ? '已删除' : '未删除' }}
                </el-descriptions-item>
                <el-descriptions-item label="发现状态">
                  {{ detail.discoverable ? '允许发现' : '不可发现' }}
                </el-descriptions-item>
                <el-descriptions-item label="分类">
                  {{ detail.category?.name || '未分类' }}
                </el-descriptions-item>
                <el-descriptions-item label="来源">
                  {{ sourceTypeLabel(detail.sourceType) }}
                </el-descriptions-item>
                <el-descriptions-item label="来源锁定">
                  {{ detail.sourceLocked ? '已锁定，不可改写来源' : '未锁定' }}
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
                v-if="canReadDiscoverable && detail.discoverable && detail.deletionStatus !== 'deleted'"
                @click="openCandidate(detail.id)"
              >
                转到候选池
              </el-button>
              <el-button
                v-if="canReadRecommendations && detail.recommendationCount > 0"
                @click="openRecommendations(detail.id)"
              >
                查看推荐
              </el-button>
              <el-button
                v-if="canGovernanceExecute && detail.deletionStatus !== 'deleted'"
                type="danger"
                @click="openGovernanceDialog(detail)"
              >
                违规处理
              </el-button>
            </div>

            <el-card shadow="never">
              <template #header>简介</template>
              <div class="whitespace-pre-wrap text-sm leading-6">
                {{ detail.description || '暂无简介' }}
              </div>
            </el-card>

            <el-card class="mt-4" shadow="never">
              <template #header>配料</template>
              <el-table :data="detail.ingredients || []" size="small" empty-text="暂无配料">
                <el-table-column label="序号" prop="sortOrder" width="70" />
                <el-table-column label="配料" prop="name" min-width="150" />
                <el-table-column label="用量" width="150">
                  <template #default="{ row }">
                    {{ row.quantity || '适量' }} {{ row.unit?.name || '' }}
                  </template>
                </el-table-column>
                <el-table-column label="备注" prop="note" min-width="180" />
              </el-table>
            </el-card>

            <el-card class="mt-4" shadow="never">
              <template #header>步骤</template>
              <el-empty
                v-if="!detail.steps?.length"
                description="暂无步骤"
                :image-size="64"
              />
              <div v-else class="space-y-3">
                <div
                  v-for="step in detail.steps"
                  :key="step.sortOrder"
                  class="rounded border border-gray-200 p-3"
                >
                  <div class="flex gap-3">
                    <el-tag round>{{ step.sortOrder }}</el-tag>
                    <div class="min-w-0 flex-1 whitespace-pre-wrap text-sm leading-6">
                      {{ step.description }}
                    </div>
                    <el-image
                      v-if="step.imageUrl"
                      :src="step.imageUrl"
                      fit="cover"
                      class="h-20 w-20 shrink-0 rounded"
                      :preview-src-list="[step.imageUrl]"
                      preview-teleported
                    />
                  </div>
                </div>
              </div>
            </el-card>

            <div class="mt-4 grid gap-4 md:grid-cols-2">
              <el-card shadow="never">
                <template #header>来源链摘要</template>
                <el-descriptions :column="1">
                  <el-descriptions-item label="直接来源">
                    <el-button
                      v-if="canOpenSource(detail.sourceOverview?.directSource)"
                      link
                      type="primary"
                      @click="openSource(detail.sourceOverview.directSource)"
                    >
                      {{ detail.sourceOverview.directSource.name }}
                    </el-button>
                    <span v-else>{{ detail.sourceOverview?.directSource?.name || '无' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="根来源">
                    <el-button
                      v-if="canOpenSource(detail.sourceOverview?.rootSource)"
                      link
                      type="primary"
                      @click="openSource(detail.sourceOverview.rootSource)"
                    >
                      {{ detail.sourceOverview.rootSource.name }}
                    </el-button>
                    <span v-else>{{ detail.sourceOverview?.rootSource?.name || '无' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="原作者">
                    <el-button
                      v-if="canReadUsers && detail.sourceOverview?.originalAuthor?.id"
                      link
                      type="primary"
                      @click="openUser(detail.sourceOverview.originalAuthor.id)"
                    >
                      {{ detail.sourceOverview.originalAuthor.nickname }}
                    </el-button>
                    <span v-else>{{ detail.sourceOverview?.originalAuthor?.nickname || '无' }}</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="复制链深度">
                    {{ detail.sourceOverview?.chainDepth ?? 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="直接复制数">
                    {{ detail.sourceOverview?.directCopyCount ?? 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="后代复制数">
                    {{ detail.sourceOverview?.descendantCopyCount ?? 0 }}
                  </el-descriptions-item>
                </el-descriptions>
              </el-card>
              <el-card shadow="never">
                <template #header>引用摘要</template>
                <el-descriptions :column="1">
                  <el-descriptions-item label="活跃菜谱">
                    {{ detail.referenceOverview?.activeRecipeCount ?? 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="未确认饭局">
                    {{ detail.referenceOverview?.unconfirmedMealCount ?? 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="已确认饭局快照">
                    {{ detail.referenceOverview?.confirmedMealSnapshotCount ?? 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="采购快照">
                    {{ detail.referenceOverview?.shoppingSnapshotCount ?? 0 }}
                  </el-descriptions-item>
                </el-descriptions>
              </el-card>
            </div>

            <el-card v-if="detail.moderationSummary" class="mt-4" shadow="never">
              <template #header>图片审核摘要</template>
              <el-descriptions :column="2">
                <el-descriptions-item label="状态">
                  {{ reviewStatusLabel(detail.moderationSummary.status) }}
                </el-descriptions-item>
                <el-descriptions-item label="风险等级">
                  {{ detail.moderationSummary.riskLevel || '无' }}
                </el-descriptions-item>
                <el-descriptions-item label="风险标签" :span="2">
                  {{ detail.moderationSummary.riskLabels?.join('、') || '无' }}
                </el-descriptions-item>
                <el-descriptions-item label="请求 ID" :span="2">
                  {{ detail.moderationSummary.requestId }}
                </el-descriptions-item>
              </el-descriptions>
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
              <span>
                违规处理记录 {{ detail.governanceRecordCount }} 条，操作审计
                {{ detail.auditLogCount }} 条
              </span>
              <el-button
                v-if="canReadGovernance && detail.governanceRecordCount > 0"
                link
                type="primary"
                @click="openGovernanceRecords(detail.id)"
              >
                查看违规处理记录
              </el-button>
              <span>本次访问审计 ID：{{ detail.accessAuditId || '—' }}</span>
            </div>
          </el-tab-pane>

          <el-tab-pane label="引用位置" name="references" lazy>
            <div v-loading="referenceLoading" class="min-h-48">
              <div class="mb-4 flex flex-wrap items-center gap-3">
                <el-select
                  v-model="referenceSearch.referenceType"
                  clearable
                  placeholder="全部引用类型"
                  class="w-48"
                  @change="reloadReferences"
                >
                  <el-option label="菜谱" value="recipe" />
                  <el-option label="未确认饭局" value="unconfirmed_meal" />
                  <el-option label="已确认饭局快照" value="confirmed_meal_snapshot" />
                  <el-option label="采购快照" value="shopping_snapshot" />
                </el-select>
                <el-select
                  v-model="referenceSearch.sortOrder"
                  class="w-28"
                  @change="reloadReferences"
                >
                  <el-option label="时间倒序" value="desc" />
                  <el-option label="时间正序" value="asc" />
                </el-select>
              </div>
              <el-alert
                v-if="referenceError"
                :title="referenceError"
                type="error"
                show-icon
                class="mb-4"
                :closable="false"
              >
                <template #default>
                  <el-button link type="primary" @click="loadReferences">
                    重新加载
                  </el-button>
                </template>
              </el-alert>
              <el-table
                :data="referenceData"
                size="small"
                :empty-text="referenceError ? '加载失败' : '暂无引用'"
              >
                <el-table-column label="引用类型" min-width="150">
                  <template #default="{ row }">
                    {{ referenceTypeLabel(row.referenceType) }}
                  </template>
                </el-table-column>
                <el-table-column label="对象" min-width="220">
                  <template #default="{ row }">
                    <div>{{ row.objectLabel }}</div>
                    <div class="text-xs text-gray-400">{{ row.objectId }}</div>
                  </template>
                </el-table-column>
                <el-table-column label="对象状态" prop="objectStatus" width="120" />
                <el-table-column label="历史快照" width="100">
                  <template #default="{ row }">
                    {{ row.historicalSnapshot ? '是' : '否' }}
                  </template>
                </el-table-column>
                <el-table-column label="发生时间" min-width="165">
                  <template #default="{ row }">
                    {{ formatDateTime(row.occurredAt) }}
                  </template>
                </el-table-column>
              </el-table>
              <div class="flex justify-end pt-4">
                <el-pagination
                  v-model:current-page="referencePage"
                  v-model:page-size="referencePageSize"
                  :page-sizes="[20, 50, 100]"
                  :total="referenceTotal"
                  layout="total, sizes, prev, pager, next"
                  @size-change="handleReferenceSizeChange"
                  @current-change="loadReferences"
                />
              </div>
              <div
                v-if="referenceAccessAuditId"
                class="pt-3 text-xs text-gray-400"
              >
                本次引用访问审计 ID：{{ referenceAccessAuditId }}
              </div>
            </div>
          </el-tab-pane>
          <el-tab-pane v-if="canReadAudit" label="操作审计" name="audit" lazy>
            <AuditPanel target-type="dish" :target-id="detail.id" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-drawer>

    <el-dialog
      v-model="governanceVisible"
      title="菜品违规处理"
      width="680px"
      destroy-on-close
      @closed="resetGovernance"
    >
      <el-alert
        title="管理员不能代用户编辑、恢复或公开菜品。违规处理必须先预览影响，再确认执行。"
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
        <el-form-item label="目标菜品">
          <span>{{ governanceTarget?.name }}（{{ governanceTarget?.id }}）</span>
        </el-form-item>
        <el-form-item label="处理动作" prop="actions">
          <el-checkbox-group
            v-model="governanceForm.actions"
            class="flex flex-col items-start gap-2"
            @change="handleGovernanceActionChange"
          >
            <el-checkbox
              v-for="option in governanceActionOptions"
              :key="option.value"
              :value="option.value"
              :disabled="option.disabled"
            >
              {{ option.label }}
            </el-checkbox>
          </el-checkbox-group>
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
            <el-radio-button value="normal" :disabled="isCascadeGovernance">
              一般
            </el-radio-button>
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
        <el-form-item
          v-if="isCascadeGovernance"
          label="确认词"
        >
          <div class="w-full">
            <el-input
              v-model="governanceForm.confirmText"
              maxlength="20"
              placeholder="请输入：确认处理复制链"
            />
            <div class="mt-1 text-xs text-red-500">
              此操作会后台删除源菜品及其复制链，必须完整输入“确认处理复制链”。
            </div>
          </div>
        </el-form-item>
      </el-form>

      <el-card v-if="governancePreview" shadow="never">
        <template #header>影响预览</template>
        <el-descriptions :column="2">
          <el-descriptions-item label="推荐记录">
            {{ governancePreview.recommendationCount }}
          </el-descriptions-item>
          <el-descriptions-item label="来源菜品">
            {{ governancePreview.sourceDishCount }}
          </el-descriptions-item>
          <el-descriptions-item label="复制菜品">
            {{ governancePreview.copiedDishCount }}
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
        <el-button
          :loading="governanceLoading"
          @click="previewGovernance"
        >
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
  getOrderFoodUserDishDetail,
  getOrderFoodUserDishList,
  getOrderFoodUserDishReferences
} from '@/api/orderfood/user-dish'
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
  name: 'OrderFoodUserDishes'
})

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()
const routeQueryValue = (value) => (Array.isArray(value) ? value[0] : value || '')
const presetKeyword = routeQueryValue(route.query.keyword)
const lockedUserId = routeQueryValue(route.query.userId)
const lockedRecipeId = routeQueryValue(route.query.recipeId)
const presetStatus = routeQueryValue(route.query.status)

const hasBtnPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))
const canReadUsers = computed(() => Boolean(btnAuth['orderfood:user:read']))
const canReadDiscoverable = computed(() =>
  Boolean(btnAuth['orderfood:discoverable-dish:read'])
)
const canReadRecommendations = computed(() =>
  Boolean(btnAuth['orderfood:recommendation:read'])
)
const canReadOfficialDishes = computed(() =>
  Boolean(btnAuth['orderfood:official-dish:read'])
)
const canReadGovernance = computed(() =>
  Boolean(btnAuth['orderfood:governance:read'])
)
const canPrivateRead = computed(() =>
  hasBtnPermission(
    'orderfood:user-dish:private-read',
    'userDishPrivateRead',
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
const canGovernanceCascade = computed(() =>
  hasBtnPermission(
    'orderfood:governance:cascade',
    'governanceCascade',
    'cascadeGovernance'
  )
)
const createDefaultSearch = () => ({
  keyword: presetKeyword,
  userId: lockedUserId,
  userKeyword: '',
  recipeId: lockedRecipeId,
  status: presetStatus,
  deletionStatus: 'active',
  discoverable: undefined,
  sourceType: '',
  sourceLocked: undefined,
  categoryId: '',
  tagIdsText: '',
  mediaReviewStatus: '',
  createdRange: [],
  updatedRange: [],
  deletedRange: [],
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
  const {
    tagIdsText,
    createdRange,
    updatedRange,
    deletedRange,
    ...filters
  } = searchInfo.value
  return compactParams({
    page: page.value,
    pageSize: pageSize.value,
    ...filters,
    tagIds: tagIdsText
      ? tagIdsText
          .split(/[,，]/)
          .map((item) => item.trim())
          .filter(Boolean)
      : undefined,
    createdFrom: toShanghaiRFC3339(createdRange?.[0]),
    createdTo: toShanghaiRFC3339(createdRange?.[1]),
    updatedFrom: toShanghaiRFC3339(updatedRange?.[0]),
    updatedTo: toShanghaiRFC3339(updatedRange?.[1]),
    deletedFrom: toShanghaiRFC3339(deletedRange?.[0]),
    deletedTo: toShanghaiRFC3339(deletedRange?.[1])
  })
}

const getTableData = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodUserDishList(buildListParams()),
      '用户菜品列表加载失败'
    )
    tableData.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    tableData.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '用户菜品列表加载失败')
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
const detailTab = ref('detail')
const activeDish = ref(null)
const detail = ref(null)

const resetDetail = () => {
  detailTab.value = 'detail'
  activeDish.value = null
  detail.value = null
  detailError.value = ''
  resetReferences()
}

const loadPrivateDetail = async () => {
  if (!canPrivateRead.value || !activeDish.value?.id) return
  detailLoading.value = true
  detailError.value = ''
  try {
    detail.value = unwrapOrderFoodResponse(
      await getOrderFoodUserDishDetail(activeDish.value.id),
      '菜品私有详情加载失败'
    )
  } catch (error) {
    detail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '菜品私有详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const openPrivateDetail = async (row) => {
  if (!canPrivateRead.value) return
  activeDish.value = row
  detailVisible.value = true
  await loadPrivateDetail()
}
const openUser = (userId) => {
  if (!userId) return
  router.push({ name: 'OrderFoodUsers', query: { userId } })
}
const openCandidate = (dishId) =>
  router.push({
    name: 'OrderFoodDiscoverableDishes',
    query: { dishId, openDishId: dishId }
  })
const openRecommendations = (dishId) =>
  router.push({
    name: 'OrderFoodRecommendations',
    query: { sourceType: 'creator', sourceDishId: dishId }
  })
const openGovernanceRecords = (dishId) =>
  router.push({
    name: 'OrderFoodGovernanceRecords',
    query: { targetType: 'dish', targetId: dishId }
  })
const canOpenSource = (source) => {
  if (!source?.dishId || !source.accessible) return false
  if (source.dishType === 'user') return canPrivateRead.value
  if (source.dishType === 'official') return canReadOfficialDishes.value
  return false
}
const openSource = (source) => {
  if (!canOpenSource(source)) return
  if (source.dishType === 'user') {
    openPrivateDetail({ id: source.dishId })
    return
  }
  router.push({
    name: 'OrderFoodOfficialDishes',
    query: { officialDishId: source.dishId }
  })
}

const referenceSearch = ref({
  referenceType: '',
  sortOrder: 'desc'
})
const referencePage = ref(1)
const referencePageSize = ref(20)
const referenceTotal = ref(0)
const referenceData = ref([])
const referenceLoading = ref(false)
const referenceError = ref('')
const referenceAccessAuditId = ref('')
const referencesLoaded = ref(false)

const resetReferences = () => {
  referenceSearch.value = { referenceType: '', sortOrder: 'desc' }
  referencePage.value = 1
  referencePageSize.value = 20
  referenceTotal.value = 0
  referenceData.value = []
  referenceError.value = ''
  referenceAccessAuditId.value = ''
  referencesLoaded.value = false
}

const loadReferences = async () => {
  if (!canPrivateRead.value || !activeDish.value?.id) return
  referenceLoading.value = true
  referenceError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodUserDishReferences(
        activeDish.value.id,
        compactParams({
          page: referencePage.value,
          pageSize: referencePageSize.value,
          referenceType: referenceSearch.value.referenceType,
          sortBy: 'occurredAt',
          sortOrder: referenceSearch.value.sortOrder
        })
      ),
      '引用位置加载失败'
    )
    referenceData.value = data?.list || []
    referenceTotal.value = Number(data?.total || 0)
    referencePage.value = Number(data?.page || referencePage.value)
    referencePageSize.value = Number(data?.pageSize || referencePageSize.value)
    referenceAccessAuditId.value = data?.accessAuditId || ''
    referencesLoaded.value = true
  } catch (error) {
    referenceData.value = []
    referenceTotal.value = 0
    referenceAccessAuditId.value = ''
    referenceError.value = getOrderFoodErrorMessage(error, '引用位置加载失败')
  } finally {
    referenceLoading.value = false
  }
}

const handleDetailTabChange = (name) => {
  if (name === 'references' && !referencesLoaded.value) {
    loadReferences()
  }
}

const reloadReferences = () => {
  referencePage.value = 1
  loadReferences()
}

const handleReferenceSizeChange = () => {
  referencePage.value = 1
  loadReferences()
}

const governanceVisible = ref(false)
const governanceLoading = ref(false)
const governanceTarget = ref(null)
const governancePreview = ref(null)
const governanceFormRef = ref(null)
const governanceForm = ref({
  actions: [],
  violationType: '',
  severity: 'normal',
  reason: '',
  confirmText: ''
})
const governanceRules = {
  actions: [
    {
      type: 'array',
      required: true,
      min: 1,
      message: '请至少选择一个处理动作',
      trigger: 'change'
    }
  ],
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

const governanceActionOptions = computed(() => {
  const row = governanceTarget.value || {}
  const options = [
    {
      value: 'disable_discoverability',
      label: '关闭发现能力',
      disabled: !row.discoverable || isCascadeGovernance.value
    },
    {
      value: 'soft_delete_dish',
      label: '软删除该菜品',
      disabled: row.deletionStatus === 'deleted' || isCascadeGovernance.value
    }
  ]
  if (canGovernanceCascade.value) {
    options.push({
      value: 'delete_copy_chain',
      label: '严重违规：删除源菜品及全部复制链',
      disabled:
        row.deletionStatus === 'deleted' ||
        governanceForm.value.actions.some(
          (action) => action !== 'delete_copy_chain'
        )
    })
  }
  return options
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

const handleGovernanceActionChange = (actions) => {
  if (actions.includes('delete_copy_chain')) {
    governanceForm.value.actions = ['delete_copy_chain']
    governanceForm.value.severity = 'serious'
  }
  if (!governanceForm.value.actions.includes('delete_copy_chain')) {
    governanceForm.value.confirmText = ''
  }
  clearGovernancePreview()
}

const openGovernanceDialog = (row) => {
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
  governanceForm.value = {
    actions: [],
    violationType: '',
    severity: 'normal',
    reason: '',
    confirmText: ''
  }
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
      ElMessage.warning('菜品版本已变化，请刷新后重新预览。')
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
  if (isCascadeGovernance.value && !isCascadeConfirmValid.value) {
    ElMessage.warning('请输入完整确认词“确认处理复制链”。')
    return
  }
  governanceLoading.value = true
  try {
    const result = unwrapOrderFoodResponse(
      await executeOrderFoodGovernanceAction({
        ...governancePayload(),
        previewToken: governancePreview.value.previewToken,
        confirmText: isCascadeGovernance.value
          ? governanceForm.value.confirmText.trim()
          : undefined
      }),
      '违规处理执行失败'
    )
    ElMessage.success(
      result?.status === 'pending'
        ? `异步处理任务已创建：${result.jobId || '等待调度'}`
        : `处理完成，共影响 ${result?.affectedCount || 0} 条记录`
    )
    governanceVisible.value = false
    await getTableData()
    if (detailVisible.value && activeDish.value?.id === governanceTarget.value?.id) {
      detailVisible.value = false
    }
  } catch (error) {
    if (isOrderFoodConflict(error)) {
      ElMessage.warning('菜品或预览已失效，请刷新后重新确认。')
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
const dishStatusLabel = (value) => ({ draft: '草稿', usable: '可用' })[value] || value
const sourceTypeLabel = (value) =>
  ({
    manual: '用户手动创建',
    creator_copy: '创作者菜品复制',
    official_copy: '官方菜品复制',
    meal_suggestion_copy: '饭局建议复制'
  })[value] || value
const reviewStatusLabel = (value) =>
  ({
    pending: '待审核',
    passed: '已通过',
    rejected: '已拒绝',
    failed: '审核失败',
    not_required: '无需审核'
  })[value] || value
const reviewStatusType = (value) =>
  ({
    pending: 'warning',
    passed: 'success',
    rejected: 'danger',
    failed: 'danger',
    not_required: 'info'
  })[value] || 'info'
const referenceTypeLabel = (value) =>
  ({
    recipe: '菜谱',
    unconfirmed_meal: '未确认饭局',
    confirmed_meal_snapshot: '已确认饭局快照',
    shopping_snapshot: '采购快照'
  })[value] || value

getTableData().then(() => {
  if (route.query.openDishId) {
    openPrivateDetail({ id: routeQueryValue(route.query.openDishId) }).then(() => {
      if (
        route.query.openGovernance === '1' &&
        canGovernanceExecute.value &&
        detail.value
      ) {
        openGovernanceDialog(detail.value)
      }
    })
  }
})
</script>
