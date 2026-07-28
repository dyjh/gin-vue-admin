<template>
  <div class="order-food-list-page">
    <AdvancedSearchPanel>
      <el-form :inline="true" :model="searchInfo" label-position="left">
        <el-form-item label="用户">
          <el-input
            v-model.trim="searchInfo.keyword"
            clearable
            placeholder="用户 ID 或昵称"
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-form-item label="账号状态">
          <el-select
            v-model="searchInfo.status"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="正常" value="normal" />
            <el-option label="已禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="能力状态">
          <el-select
            v-model="searchInfo.capabilityEffective"
            clearable
            placeholder="全部"
            class="w-32"
          >
            <el-option label="有效开启" value="enabled" />
            <el-option label="有效关闭" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="积分">
          <div class="flex items-center gap-2">
            <el-input-number
              v-model="searchInfo.minPoints"
              :min="0"
              :controls="false"
              placeholder="最低"
              class="w-28"
            />
            <span>至</span>
            <el-input-number
              v-model="searchInfo.maxPoints"
              :min="0"
              :controls="false"
              placeholder="最高"
              class="w-28"
            />
          </div>
        </el-form-item>
        <el-form-item label="注册时间">
          <el-date-picker
            v-model="searchInfo.registeredRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item label="最近登录">
          <el-date-picker
            v-model="searchInfo.lastLoginRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            range-separator="至"
          />
        </el-form-item>
        <el-form-item label="活跃日期">
          <el-date-picker
            v-model="searchInfo.activeRange"
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
              <el-option label="注册时间" value="createdAt" />
              <el-option label="最近登录" value="lastLoginAt" />
              <el-option label="积分" value="points" />
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
        :empty-text="listError ? '加载失败' : '暂无用户'"
      >
        <el-table-column label="用户" min-width="280" fixed="left">
          <template #default="{ row }">
            <div class="flex min-w-0 items-center gap-2">
              <el-avatar
                :size="36"
                :src="row.avatarUrl || undefined"
                class="shrink-0"
              >
                {{ row.nickname?.slice(0, 1) || '用' }}
              </el-avatar>
              <div class="min-w-0 flex-1 overflow-hidden leading-5">
                <div class="truncate font-medium">{{ row.nickname || '未命名用户' }}</div>
                <div class="truncate text-xs text-gray-400">{{ row.id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="积分" prop="points" width="90" align="right" />
        <el-table-column label="整体能力" min-width="180">
          <template #default="{ row }">
            <div class="flex items-center gap-2 whitespace-nowrap">
              <el-tag
                :type="row.capabilityEffective === 'enabled' ? 'success' : 'warning'"
              >
                {{ capabilityEffectiveLabel(row.capabilityEffective) }}
              </el-tag>
              <span class="text-xs text-gray-400">
                {{ capabilitySourceLabel(row.capabilitySource) }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="打卡天数" prop="checkinDayCount" width="100" />
        <el-table-column label="菜品数" prop="dishCount" width="90" />
        <el-table-column label="饭局数" prop="mealCount" width="90" />
        <el-table-column label="账号状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'normal' ? 'success' : 'danger'">
              {{ userStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="注册时间" min-width="165">
          <template #default="{ row }">{{ formatDateTime(row.registeredAt) }}</template>
        </el-table-column>
        <el-table-column label="最近登录" min-width="165">
          <template #default="{ row }">{{ formatDateTime(row.lastLoginAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="360" fixed="right">
          <template #default="{ row }">
            <div class="table-row-actions">
              <el-button link type="primary" @click="openDetail(row)">
                查看详情
              </el-button>
              <el-button
                v-if="canReadPreference"
                link
                type="primary"
                @click="openDetail(row, 'preference')"
              >
                偏好画像
              </el-button>
              <el-button
                v-if="canAdjustPoints"
                link
                type="primary"
                @click="openPointAdjustmentFor(row)"
              >
                调整积分
              </el-button>
              <el-dropdown
                v-if="canDisableUser"
                trigger="click"
              >
                <el-button link type="primary">更多</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-if="canDisableUser"
                      @click="openStatusDialog(row)"
                    >
                      {{ row.status === 'normal' ? '禁用用户' : '恢复用户' }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
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
      size="min(860px, 92vw)"
      class="user-detail-drawer"
      destroy-on-close
      @closed="resetDetail"
    >
      <template #header>
        <div class="user-detail-header">
          <div class="user-detail-identity">
            <el-avatar
              :size="46"
              :src="userDetail?.avatarUrl || selectedUser?.avatarUrl || undefined"
              class="user-detail-avatar"
            >
              {{ detailTitle.slice(0, 1) }}
            </el-avatar>
            <div class="user-detail-name">
              <div class="user-detail-title">{{ detailTitle }}</div>
              <div class="user-detail-meta">
                <span>{{ activeUserId }}</span>
                <el-tag
                  v-if="userDetail"
                  size="small"
                  effect="plain"
                  :type="userDetail.status === 'normal' ? 'success' : 'danger'"
                >
                  {{ userStatusLabel(userDetail.status) }}
                </el-tag>
              </div>
            </div>
          </div>
          <div v-if="userDetail" class="user-detail-actions">
            <el-button v-if="canAdjustPoints" type="primary" plain @click="openPointAdjustment">
              调整积分
            </el-button>
            <el-button
              v-if="canDisableUser"
              plain
              :type="userDetail.status === 'normal' ? 'danger' : 'primary'"
              @click="openStatusDialog(userDetail)"
            >
              {{ userDetail.status === 'normal' ? '禁用用户' : '恢复用户' }}
            </el-button>
          </div>
        </div>
      </template>

      <div v-loading="detailLoading" class="user-detail-body">
        <el-alert
          v-if="detailError"
          :title="detailError"
          type="error"
          show-icon
          :closable="false"
        >
          <template #default>
            <el-button link type="primary" @click="loadUserDetail">重新加载</el-button>
          </template>
        </el-alert>

        <el-tabs
          v-else-if="userDetail"
          v-model="detailTab"
          class="user-detail-tabs"
          @tab-change="handleDetailTabChange"
        >
          <el-tab-pane label="概览" name="overview">
            <div class="detail-tab-panel">
              <section class="overview-metric-grid">
                <article class="overview-metric-card is-primary">
                  <span class="overview-metric-label">当前积分</span>
                  <strong>{{ userDetail.points }}</strong>
                  <div class="overview-metric-actions">
                    <el-button link type="primary" @click="openPointEntries">查看流水</el-button>
                    <el-button v-if="canAdjustPoints" link type="primary" @click="openPointAdjustment">
                      调整积分
                    </el-button>
                  </div>
                </article>
                <article class="overview-metric-card">
                  <span class="overview-metric-label">菜品</span>
                  <strong>{{ userDetail.dishCount }}</strong>
                  <el-button v-if="canReadUserDishes" link type="primary" @click="openUserDishes">
                    查看全部
                  </el-button>
                </article>
                <article class="overview-metric-card">
                  <span class="overview-metric-label">菜谱</span>
                  <strong>{{ userDetail.recipeCount }}</strong>
                  <el-button v-if="canReadUserRecipes" link type="primary" @click="openUserRecipes">
                    查看全部
                  </el-button>
                </article>
                <article class="overview-metric-card">
                  <span class="overview-metric-label">饭局</span>
                  <strong>{{ userDetail.mealCount }}</strong>
                  <span class="overview-metric-hint">累计创建</span>
                </article>
              </section>

              <section class="detail-card">
                <header class="detail-card-header">
                  <div><h3>账户资料</h3><p>基础身份、活跃情况与能力配置</p></div>
                </header>
                <dl class="user-info-grid">
                  <div class="user-info-item">
                    <dt>账号状态</dt>
                    <dd>
                      <el-tag size="small" effect="plain" :type="userDetail.status === 'normal' ? 'success' : 'danger'">
                        {{ userStatusLabel(userDetail.status) }}
                      </el-tag>
                    </dd>
                  </div>
                  <div class="user-info-item">
                    <dt>微信身份</dt>
                    <dd class="user-info-identity">{{ userDetail.wechatIdentityMasked || '未提供' }}</dd>
                  </div>
                  <div class="user-info-item">
                    <dt>整体能力</dt>
                    <dd>
                      {{ capabilityEffectiveLabel(userDetail.capabilityEffective) }}
                      <span class="user-info-note">{{ capabilitySourceLabel(userDetail.capabilitySource) }}</span>
                    </dd>
                  </div>
                  <div class="user-info-item"><dt>打卡记录</dt><dd>{{ userDetail.checkinCount }} 次 / {{ userDetail.checkinDayCount }} 天</dd></div>
                  <div class="user-info-item"><dt>注册时间</dt><dd>{{ formatDateTime(userDetail.registeredAt) }}</dd></div>
                  <div class="user-info-item"><dt>最近登录</dt><dd>{{ formatDateTime(userDetail.lastLoginAt) }}</dd></div>
                  <div v-if="userDetail.disabledReason" class="user-info-item is-full"><dt>禁用原因</dt><dd>{{ userDetail.disabledReason }}</dd></div>
                </dl>
              </section>

              <section class="detail-card">
                <header class="detail-card-header">
                  <div><h3>整体能力实际效果</h3><p>当前配置在小程序端的最终生效结果</p></div>
                </header>
                <div class="capability-effect-grid">
                  <div class="capability-effect-item" :class="{ 'is-enabled': userDetail.capabilityEffects?.pointsVisible }">
                    <span class="capability-effect-dot"></span><div><strong>积分入口</strong><span>{{ booleanLabel(userDetail.capabilityEffects?.pointsVisible) }}</span></div>
                  </div>
                  <div class="capability-effect-item" :class="{ 'is-enabled': userDetail.capabilityEffects?.pointEntriesVisible }">
                    <span class="capability-effect-dot"></span><div><strong>积分流水</strong><span>{{ booleanLabel(userDetail.capabilityEffects?.pointEntriesVisible) }}</span></div>
                  </div>
                  <div class="capability-effect-item" :class="{ 'is-enabled': userDetail.capabilityEffects?.checkinRewardEnabled }">
                    <span class="capability-effect-dot"></span><div><strong>打卡奖励</strong><span>{{ booleanLabel(userDetail.capabilityEffects?.checkinRewardEnabled) }}</span></div>
                  </div>
                  <div class="capability-effect-item" :class="{ 'is-enabled': userDetail.capabilityEffects?.preferenceAnalysisEnabled }">
                    <span class="capability-effect-dot"></span><div><strong>偏好分析</strong><span>{{ booleanLabel(userDetail.capabilityEffects?.preferenceAnalysisEnabled) }}</span></div>
                  </div>
                </div>
              </section>
            </div>
          </el-tab-pane>
          <el-tab-pane
            v-if="canReadPreference"
            label="偏好画像"
            name="preference"
            lazy
          >
            <div v-loading="preferenceLoading" class="detail-tab-panel preference-panel">
              <el-alert
                v-if="preferenceError"
                :title="preferenceError"
                type="error"
                show-icon
                :closable="false"
              >
                <template #default>
                  <el-button link type="primary" @click="loadPreference">重新加载</el-button>
                </template>
              </el-alert>

              <template v-else-if="preference">
                <section class="preference-status-card">
                  <div class="preference-status-copy">
                    <el-tag size="small" effect="plain" :type="preferenceStateType(preference.updateState)">
                      {{ preferenceStateLabel(preference.updateState) }}
                    </el-tag>
                    <div>
                      <h3>画像更新状态</h3>
                      <p>根据用户打卡与使用记录异步归纳，不展示原始隐私内容。</p>
                    </div>
                  </div>
                  <div class="preference-meta-grid">
                    <div><span>累计打卡</span><strong>{{ preference.checkinDayCount }} 天</strong></div>
                    <div><span>最近聚合</span><strong>{{ formatDateTime(preference.lastAggregatedAt) }}</strong></div>
                  </div>
                </section>

                <el-alert
                  v-if="preference.updateState === 'paused'"
                  :title="preference.pausedReasonSummary || '当前已暂停更新'"
                  type="warning"
                  show-icon
                  class="preference-alert"
                  :closable="false"
                />
                <el-alert
                  v-else-if="preference.updateState === 'pending'"
                  title="画像正在等待下一次异步聚合，已有结果仍可继续查看。"
                  type="info"
                  show-icon
                  class="preference-alert"
                  :closable="false"
                />
                <el-alert
                  v-else-if="preference.updateState === 'failed'"
                  :title="preference.latestFailure?.safeMessage || '最近一次画像更新失败'"
                  type="error"
                  show-icon
                  class="preference-alert"
                  :closable="false"
                >
                  <template v-if="preference.latestFailure" #default>
                    <div class="text-xs">
                      失败时间：{{ formatDateTime(preference.latestFailure.failedAt) }}；
                      错误码：{{ preference.latestFailure.stableErrorCode }}；
                      请求 ID：{{ preference.latestFailure.requestId }}
                    </div>
                  </template>
                </el-alert>

                <section v-if="!preference.hasProfile" class="preference-empty-card">
                  <el-empty :image-size="86" description="暂无可用画像">
                    <p class="preference-empty-note">符合条件的打卡和其他使用记录会异步参与归纳</p>
                  </el-empty>
                </section>

                <template v-else>
                  <section class="detail-card">
                    <header class="detail-card-header">
                      <div><h3>画像标签</h3><p>由明确设置与使用记录共同归纳</p></div>
                    </header>
                    <div class="detail-card-content">
                      <preference-section title="常用菜品标签" :items="preference.profile?.commonDishTags" />
                      <preference-section title="常点菜品标签" :items="preference.profile?.frequentlyOrderedDishTags" />
                      <preference-section title="常用配料" :items="preference.profile?.commonIngredients" />
                    </div>
                  </section>

                  <div class="preference-summary-grid">
                    <el-card class="preference-content-card" shadow="never">
                      <template #header>口味偏好摘要</template>
                      <preference-text :value="preference.profile?.tastePreferenceSummary" />
                    </el-card>
                    <el-card class="preference-content-card" shadow="never">
                      <template #header>忌口或偏好备注摘要</template>
                      <preference-text :value="preference.profile?.avoidanceOrPreferenceSummary" />
                    </el-card>
                  </div>

                  <el-card class="preference-content-card" shadow="never">
                    <template #header>用户明确设置</template>
                    <el-empty
                      v-if="!preference.profile?.userSettings?.length"
                      description="暂无用户主动设置"
                      :image-size="64"
                    />
                    <el-table v-else :data="preference.profile.userSettings" size="small">
                      <el-table-column label="类型" width="130">
                        <template #default="{ row }">{{ preferenceSettingTypeLabel(row.type) }}</template>
                      </el-table-column>
                      <el-table-column label="内容" prop="value" min-width="220" />
                      <el-table-column label="更新时间" min-width="160">
                        <template #default="{ row }">{{ formatDateTime(row.updatedAt) }}</template>
                      </el-table-column>
                    </el-table>
                  </el-card>
                </template>

                <section class="detail-card evidence-card">
                  <header class="detail-card-header">
                    <div><h3>证据来源</h3><p>仅展示结构化计数，便于判断画像依据是否充分</p></div>
                  </header>
                  <el-table
                    :data="preference.evidenceSources || []"
                    size="small"
                    class="preference-table"
                    empty-text="暂无结构化证据"
                  >
                    <el-table-column label="来源" min-width="160">
                      <template #default="{ row }">{{ evidenceSourceLabel(row.sourceType) }}</template>
                    </el-table-column>
                    <el-table-column label="总数" prop="totalCount" width="80" />
                    <el-table-column label="已聚合" prop="aggregatedCount" width="90" />
                    <el-table-column label="待聚合" prop="pendingCount" width="90" />
                    <el-table-column label="最近证据" min-width="160">
                      <template #default="{ row }">{{ formatDateTime(row.latestOccurredAt) }}</template>
                    </el-table-column>
                  </el-table>
                </section>

                <div class="preference-privacy-note">
                  <span class="preference-privacy-icon">i</span>
                  <span>
                    本页只展示聚合结果和结构化计数，不展示打卡原图、完整证据、
                    提示词或模型推理。访问审计 ID：{{ preference.accessAuditId || '—' }}
                  </span>
                </div>
              </template>
            </div>
          </el-tab-pane>
          <el-tab-pane
            v-if="canReadPoints"
            label="积分流水"
            name="points"
            lazy
          >
            <div v-loading="pointsLoading" class="min-h-48">
              <div class="mb-3 flex items-center justify-between">
                <span class="text-sm text-gray-500">
                  最近 {{ pointEntries.length }} 条，共 {{ pointEntryTotal }} 条
                </span>
                <el-button link type="primary" @click="openPointEntries">
                  查看全部
                </el-button>
              </div>
              <el-alert
                v-if="pointsError"
                :title="pointsError"
                type="error"
                show-icon
                :closable="false"
                class="mb-3"
              >
                <template #default>
                  <el-button link type="primary" @click="loadPointEntries">
                    重新加载
                  </el-button>
                </template>
              </el-alert>
              <el-table
                v-else
                :data="pointEntries"
                size="small"
                empty-text="暂无积分流水"
              >
                <el-table-column label="说明" prop="title" min-width="180" />
                <el-table-column label="类型" width="90">
                  <template #default="{ row }">
                    {{ pointTypeLabel(row.type) }}
                  </template>
                </el-table-column>
                <el-table-column label="积分变化" width="100" align="right">
                  <template #default="{ row }">
                    <span :class="row.amount >= 0 ? 'text-green-600' : 'text-red-500'">
                      {{ row.amount >= 0 ? '+' : '' }}{{ row.amount }}
                    </span>
                  </template>
                </el-table-column>
                <el-table-column
                  label="变化后余额"
                  prop="balanceAfter"
                  width="105"
                  align="right"
                />
                <el-table-column label="场景" prop="scene" min-width="130" />
                <el-table-column label="时间" min-width="165">
                  <template #default="{ row }">
                    {{ formatDateTime(row.createdAt) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>
          <el-tab-pane
            v-if="canReadAIUsage"
            label="AI 调用"
            name="aiUsage"
            lazy
          >
            <div v-loading="aiUsageLoading" class="min-h-48">
              <div class="mb-3 flex items-center justify-between">
                <span class="text-sm text-gray-500">
                  最近 {{ aiUsageItems.length }} 条，共 {{ aiUsageTotal }} 条
                </span>
                <el-button link type="primary" @click="openAIUsages">
                  查看全部
                </el-button>
              </div>
              <el-alert
                v-if="aiUsageError"
                :title="aiUsageError"
                type="error"
                show-icon
                :closable="false"
                class="mb-3"
              >
                <template #default>
                  <el-button link type="primary" @click="loadAIUsages">
                    重新加载
                  </el-button>
                </template>
              </el-alert>
              <el-table
                v-else
                :data="aiUsageItems"
                size="small"
                empty-text="暂无 AI 调用记录"
              >
                <el-table-column label="能力" min-width="150">
                  <template #default="{ row }">
                    {{ aiCapabilityLabel(row.capabilityCode) }}
                  </template>
                </el-table-column>
                <el-table-column label="供应商 / 模型" min-width="170">
                  <template #default="{ row }">
                    {{ row.providerName || '—' }} / {{ row.modelName || '—' }}
                  </template>
                </el-table-column>
                <el-table-column label="执行" width="90">
                  <template #default="{ row }">
                    {{ aiExecutionLabel(row.executionStatus) }}
                  </template>
                </el-table-column>
                <el-table-column label="计费" width="100">
                  <template #default="{ row }">
                    {{ aiBillingLabel(row.billingStatus) }}
                  </template>
                </el-table-column>
                <el-table-column label="积分" prop="pointCost" width="80" align="right" />
                <el-table-column label="时间" min-width="165">
                  <template #default="{ row }">
                    {{ formatDateTime(row.createdAt) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>
          <el-tab-pane
            v-if="canReadNotifications"
            label="站内通知"
            name="notifications"
            lazy
          >
            <div v-loading="notificationsLoading" class="min-h-48">
              <div class="mb-3 flex items-center justify-between">
                <span class="text-sm text-gray-500">
                  最近 {{ notificationItems.length }} 条，共 {{ notificationTotal }} 条
                </span>
                <el-button link type="primary" @click="openNotifications">
                  查看全部
                </el-button>
              </div>
              <el-alert
                v-if="notificationsError"
                :title="notificationsError"
                type="error"
                show-icon
                :closable="false"
                class="mb-3"
              >
                <template #default>
                  <el-button link type="primary" @click="loadNotifications">
                    重新加载
                  </el-button>
                </template>
              </el-alert>
              <el-table
                v-else
                :data="notificationItems"
                size="small"
                empty-text="暂无站内通知"
              >
                <el-table-column label="标题" prop="title" min-width="210" />
                <el-table-column label="类型" width="110">
                  <template #default="{ row }">
                    {{ notificationTypeLabel(row.type) }}
                  </template>
                </el-table-column>
                <el-table-column label="状态" width="80">
                  <template #default="{ row }">
                    {{ row.read ? '已读' : '未读' }}
                  </template>
                </el-table-column>
                <el-table-column label="目标" min-width="150">
                  <template #default="{ row }">
                    {{ row.targetType || '—' }}
                  </template>
                </el-table-column>
                <el-table-column label="时间" min-width="165">
                  <template #default="{ row }">
                    {{ formatDateTime(row.createdAt) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>
          <el-tab-pane v-if="canReadAudit" label="操作审计" name="audit" lazy>
            <AuditPanel target-type="user" :target-id="activeUserId" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-drawer>

    <PointAdjustmentDialog
      v-model="pointAdjustmentVisible"
      :user="pointAdjustmentUser"
      @success="handlePointAdjustmentSuccess"
    />

    <el-dialog
      v-model="statusDialogVisible"
      :title="statusForm.status === 'disabled' ? '禁用用户' : '恢复用户'"
      width="520px"
      destroy-on-close
    >
      <el-alert
        :title="
          statusForm.status === 'disabled'
            ? '禁用后将阻止该用户登录；既有内容不删除、不下线，但该用户发起的进行中饭局会立即取消。'
            : '恢复账号不会自动恢复此前已处理的违规内容。'
        "
        :type="statusForm.status === 'disabled' ? 'warning' : 'info'"
        show-icon
        :closable="false"
        class="mb-4"
      />
      <div
        v-if="statusForm.status === 'disabled'"
        v-loading="statusImpactLoading"
        class="mb-4"
      >
        <el-alert
          v-if="statusImpactError"
          type="error"
          :closable="false"
          :title="statusImpactError"
        />
        <el-descriptions v-else-if="statusImpact" :column="1" border>
          <el-descriptions-item label="受影响饭局">
            {{ statusImpact.mealId }}（{{ mealStatusLabel(statusImpact.status) }}）
          </el-descriptions-item>
          <el-descriptions-item label="参与者">
            {{ statusImpact.participantCount }} 人，除被禁用发起人外均写入站内通知
          </el-descriptions-item>
          <el-descriptions-item label="保留数据">
            最终菜单快照 {{ statusImpact.finalSnapshotCount }} 条，采购清单
            {{ statusImpact.shoppingListCount }} 份
          </el-descriptions-item>
        </el-descriptions>
        <el-empty
          v-else
          :image-size="56"
          description="该用户没有发起中的饭局"
        />
      </div>
      <el-form
        ref="statusFormRef"
        :model="statusForm"
        :rules="reasonRules"
        label-width="84px"
      >
        <el-form-item label="目标用户">
          <span>{{ selectedUser?.nickname }}（{{ selectedUser?.id }}）</span>
        </el-form-item>
        <el-form-item label="原因" prop="reason">
          <el-input
            v-model.trim="statusForm.reason"
            type="textarea"
            :rows="4"
            maxlength="200"
            show-word-limit
            placeholder="请输入 4-200 字操作原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusDialogVisible = false">取消</el-button>
        <el-button
          :type="statusForm.status === 'disabled' ? 'danger' : 'primary'"
          :loading="mutationLoading"
          :disabled="
            statusForm.status === 'disabled' &&
            (statusImpactLoading || Boolean(statusImpactError))
          "
          @click="submitStatus"
        >
          确认{{ statusForm.status === 'disabled' ? '禁用' : '恢复' }}
        </el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { computed, defineComponent, h, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElTag } from 'element-plus'
import { useBtnAuth } from '@/utils/btnAuth'
import {
  formatOrderFoodDateTime as formatDate,
  getShanghaiPresetRange,
  toShanghaiRFC3339
} from '@/view/orderFood/utils/time'
import AuditPanel from '@/view/orderFood/components/AuditPanel.vue'
import PointAdjustmentDialog from '@/view/orderFood/components/PointAdjustmentDialog.vue'
import {
  getOrderFoodUserDetail,
  getOrderFoodUserList,
  getOrderFoodUserPreferenceProfile,
  updateOrderFoodUserStatus
} from '@/api/orderfood/user'
import { getPointEntryList } from '@/api/orderfood/points'
import { getAIUsageList } from '@/api/orderfood/operations'
import { getOrderFoodNotificationList } from '@/api/orderfood/message'
import {
  getOrderFoodErrorMessage,
  isOrderFoodConflict,
  unwrapOrderFoodResponse
} from '@/api/orderfood/request'

defineOptions({
  name: 'OrderFoodUsers'
})

const PreferenceSection = defineComponent({
  name: 'PreferenceSection',
  props: {
    title: {
      type: String,
      required: true
    },
    items: {
      type: Array,
      default: () => []
    }
  },
  setup(props) {
    return () =>
      h(
        'div',
        { class: 'preference-tag-section' },
        [
          h('div', { class: 'preference-section-title' }, props.title),
          props.items.length
            ? h(
                'div',
                { class: 'preference-tag-list' },
                props.items.map((item, index) =>
                  h(
                    ElTag,
                    {
                      key: `${item.id || item.name}-${index}`,
                      type: item.origin === 'user_setting' ? 'success' : 'info',
                      effect: 'plain'
                    },
                    {
                      default: () =>
                        `${item.name} · ${
                          item.origin === 'user_setting'
                            ? '用户设置'
                            : '根据使用记录归纳'
                        }`
                    }
                  )
                )
              )
            : h('div', { class: 'preference-section-empty' }, '暂无')
        ]
      )
  }
})

const PreferenceText = defineComponent({
  name: 'PreferenceText',
  props: {
    value: {
      type: Object,
      default: null
    }
  },
  setup(props) {
    return () =>
      props.value
        ? h('div', { class: 'preference-summary' }, [
            h('div', { class: 'preference-summary-text' }, props.value.text),
            h(
              ElTag,
              {
                size: 'small',
                type: props.value.origin === 'user_setting' ? 'success' : 'info'
              },
              {
                default: () =>
                  props.value.origin === 'user_setting'
                    ? '用户设置'
                    : '根据使用记录归纳'
              }
            )
          ])
        : h('div', { class: 'preference-section-empty' }, '暂无')
  }
})

const route = useRoute()
const router = useRouter()
const btnAuth = useBtnAuth()

const hasBtnPermission = (...keys) => keys.some((key) => Boolean(btnAuth[key]))

const canReadPreference = computed(() =>
  hasBtnPermission(
    'orderfood:user:preference:read',
    'preferenceRead',
    'viewPreference'
  )
)
const canReadAudit = computed(() => Boolean(btnAuth['orderfood:audit:read']))
const canReadPoints = computed(() => Boolean(btnAuth['orderfood:points:read']))
const canReadAIUsage = computed(() => Boolean(btnAuth['orderfood:ai-usage:read']))
const canReadNotifications = computed(() =>
  Boolean(btnAuth['orderfood:notification:read'])
)
const canDisableUser = computed(() =>
  hasBtnPermission('orderfood:user:disable', 'userDisable', 'disableUser')
)
const canAdjustPoints = computed(() =>
  hasBtnPermission('orderfood:points:adjust', 'pointsAdjust')
)
const canReadUserDishes = computed(() =>
  hasBtnPermission(
    'orderfood:user-dish:read',
    'userDishRead',
    'viewUserDishes'
  )
)
const canReadUserRecipes = computed(() =>
  hasBtnPermission(
    'orderfood:user-recipe:read',
    'userRecipeRead',
    'viewUserRecipes'
  )
)

const emptySearch = () => ({
  keyword: '',
  status: '',
  capabilityEffective: '',
  minPoints: undefined,
  maxPoints: undefined,
  registeredRange: [],
  lastLoginRange: [],
  activeRange: [],
  sortBy: 'createdAt',
  sortOrder: 'desc'
})
const createDefaultSearch = () => ({
  ...emptySearch(),
  keyword: String(route.query.userId || ''),
  capabilityEffective: String(route.query.capabilityEffective || ''),
  registeredRange: getShanghaiPresetRange(String(route.query.registeredRange || '')),
  lastLoginRange: getShanghaiPresetRange(String(route.query.lastLoginRange || '')),
  activeRange: getShanghaiPresetRange(String(route.query.activeRange || ''))
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
  const { registeredRange, lastLoginRange, activeRange, ...filters } = searchInfo.value
  return compactParams({
    page: page.value,
    pageSize: pageSize.value,
    ...filters,
    registeredFrom: toShanghaiRFC3339(registeredRange?.[0]),
    registeredTo: toShanghaiRFC3339(registeredRange?.[1]),
    lastLoginFrom: toShanghaiRFC3339(lastLoginRange?.[0]),
    lastLoginTo: toShanghaiRFC3339(lastLoginRange?.[1]),
    activeFrom: toShanghaiRFC3339(activeRange?.[0]),
    activeTo: toShanghaiRFC3339(activeRange?.[1])
  })
}

const getTableData = async () => {
  listLoading.value = true
  listError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodUserList(buildListParams()),
      '用户列表加载失败'
    )
    tableData.value = data?.list || []
    total.value = Number(data?.total || 0)
    page.value = Number(data?.page || page.value)
    pageSize.value = Number(data?.pageSize || pageSize.value)
  } catch (error) {
    tableData.value = []
    total.value = 0
    listError.value = getOrderFoodErrorMessage(error, '用户列表加载失败')
  } finally {
    listLoading.value = false
  }
}

const onSubmit = () => {
  page.value = 1
  getTableData()
}

const onReset = () => {
  searchInfo.value = emptySearch()
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
const detailTab = ref('overview')
const activeUserId = ref('')
const userDetail = ref(null)
const preference = ref(null)
const preferenceLoading = ref(false)
const preferenceError = ref('')
const preferenceLoadedUserId = ref('')
const pointEntries = ref([])
const pointEntryTotal = ref(0)
const pointsLoading = ref(false)
const pointsError = ref('')
const pointsLoadedUserId = ref('')
const aiUsageItems = ref([])
const aiUsageTotal = ref(0)
const aiUsageLoading = ref(false)
const aiUsageError = ref('')
const aiUsageLoadedUserId = ref('')
const notificationItems = ref([])
const notificationTotal = ref(0)
const notificationsLoading = ref(false)
const notificationsError = ref('')
const notificationsLoadedUserId = ref('')

const detailTitle = computed(
  () => userDetail.value?.nickname || selectedUser.value?.nickname || '用户详情'
)

const resetDetail = () => {
  detailTab.value = 'overview'
  activeUserId.value = ''
  userDetail.value = null
  detailError.value = ''
  preference.value = null
  preferenceError.value = ''
  preferenceLoadedUserId.value = ''
  pointEntries.value = []
  pointEntryTotal.value = 0
  pointsError.value = ''
  pointsLoadedUserId.value = ''
  aiUsageItems.value = []
  aiUsageTotal.value = 0
  aiUsageError.value = ''
  aiUsageLoadedUserId.value = ''
  notificationItems.value = []
  notificationTotal.value = 0
  notificationsError.value = ''
  notificationsLoadedUserId.value = ''
}

const loadUserDetail = async () => {
  if (!activeUserId.value) return
  detailLoading.value = true
  detailError.value = ''
  try {
    userDetail.value = unwrapOrderFoodResponse(
      await getOrderFoodUserDetail(activeUserId.value),
      '用户详情加载失败'
    )
  } catch (error) {
    userDetail.value = null
    detailError.value = getOrderFoodErrorMessage(error, '用户详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const loadPreference = async () => {
  if (
    !canReadPreference.value ||
    !activeUserId.value ||
    preferenceLoading.value
  ) {
    return
  }
  if (
    preferenceLoadedUserId.value === activeUserId.value &&
    preference.value &&
    !preferenceError.value
  ) {
    return
  }

  preferenceLoading.value = true
  preferenceError.value = ''
  try {
    preference.value = unwrapOrderFoodResponse(
      await getOrderFoodUserPreferenceProfile(activeUserId.value),
      '偏好画像加载失败'
    )
    preferenceLoadedUserId.value = activeUserId.value
  } catch (error) {
    preference.value = null
    preferenceError.value = getOrderFoodErrorMessage(error, '偏好画像加载失败')
  } finally {
    preferenceLoading.value = false
  }
}

const loadPointEntries = async () => {
  if (!canReadPoints.value || !activeUserId.value || pointsLoading.value) return
  if (
    pointsLoadedUserId.value === activeUserId.value &&
    !pointsError.value
  ) {
    return
  }
  pointsLoading.value = true
  pointsError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getPointEntryList({
        page: 1,
        pageSize: 5,
        userId: activeUserId.value,
        sortBy: 'createdAt',
        sortOrder: 'desc'
      }),
      '积分流水加载失败'
    )
    pointEntries.value = data?.list || []
    pointEntryTotal.value = Number(data?.total || 0)
    pointsLoadedUserId.value = activeUserId.value
  } catch (error) {
    pointEntries.value = []
    pointEntryTotal.value = 0
    pointsError.value = getOrderFoodErrorMessage(error, '积分流水加载失败')
  } finally {
    pointsLoading.value = false
  }
}

const loadAIUsages = async () => {
  if (!canReadAIUsage.value || !activeUserId.value || aiUsageLoading.value) return
  if (
    aiUsageLoadedUserId.value === activeUserId.value &&
    !aiUsageError.value
  ) {
    return
  }
  aiUsageLoading.value = true
  aiUsageError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getAIUsageList({
        page: 1,
        pageSize: 20,
        userId: activeUserId.value,
        sortBy: 'createdAt',
        sortOrder: 'desc'
      }),
      'AI 调用记录加载失败'
    )
    aiUsageItems.value = data?.list || []
    aiUsageTotal.value = Number(data?.total || 0)
    aiUsageLoadedUserId.value = activeUserId.value
  } catch (error) {
    aiUsageItems.value = []
    aiUsageTotal.value = 0
    aiUsageError.value = getOrderFoodErrorMessage(error, 'AI 调用记录加载失败')
  } finally {
    aiUsageLoading.value = false
  }
}

const loadNotifications = async () => {
  if (
    !canReadNotifications.value ||
    !activeUserId.value ||
    notificationsLoading.value
  ) {
    return
  }
  if (
    notificationsLoadedUserId.value === activeUserId.value &&
    !notificationsError.value
  ) {
    return
  }
  notificationsLoading.value = true
  notificationsError.value = ''
  try {
    const data = unwrapOrderFoodResponse(
      await getOrderFoodNotificationList({
        page: 1,
        pageSize: 20,
        userId: activeUserId.value,
        sortBy: 'createdAt',
        sortOrder: 'desc'
      }),
      '站内通知加载失败'
    )
    notificationItems.value = data?.list || []
    notificationTotal.value = Number(data?.total || 0)
    notificationsLoadedUserId.value = activeUserId.value
  } catch (error) {
    notificationItems.value = []
    notificationTotal.value = 0
    notificationsError.value = getOrderFoodErrorMessage(error, '站内通知加载失败')
  } finally {
    notificationsLoading.value = false
  }
}

const openDetail = async (row, tab = 'overview') => {
  selectedUser.value = row
  activeUserId.value = row.id
  detailTab.value =
    tab === 'preference' && canReadPreference.value ? 'preference' : 'overview'
  detailVisible.value = true
  await loadUserDetail()
  if (detailTab.value === 'preference') {
    await loadPreference()
  }
}

const handleDetailTabChange = (name) => {
  if (name === 'preference') {
    loadPreference()
  } else if (name === 'points') {
    loadPointEntries()
  } else if (name === 'aiUsage') {
    loadAIUsages()
  } else if (name === 'notifications') {
    loadNotifications()
  }
}

const openUserDishes = () => {
  router.push({
    name: 'OrderFoodUserDishes',
    query: { userId: activeUserId.value }
  })
}

const openUserRecipes = () => {
  router.push({
    name: 'OrderFoodUserRecipes',
    query: { userId: activeUserId.value }
  })
}

const openPointEntries = () => {
  router.push({
    name: 'OrderFoodPointEntries',
    query: { userId: activeUserId.value }
  })
}

const pointAdjustmentVisible = ref(false)
const pointAdjustmentUser = ref(null)

const openPointAdjustmentFor = (user) => {
  if (!user) return
  pointAdjustmentUser.value =
    typeof user === 'string' ? { id: user } : user
  pointAdjustmentVisible.value = true
}

const openPointAdjustment = () =>
  openPointAdjustmentFor(
    userDetail.value || selectedUser.value || { id: activeUserId.value }
  )

const handlePointAdjustmentSuccess = async () => {
  pointsLoadedUserId.value = ''
  await getTableData()
  if (detailVisible.value && activeUserId.value) {
    await loadUserDetail()
    if (detailTab.value === 'points') {
      await loadPointEntries()
    }
  }
}

const openAIUsages = () => {
  router.push({
    name: 'OrderFoodAIUsages',
    query: { userId: activeUserId.value }
  })
}

const openNotifications = () => {
  router.push({
    name: 'OrderFoodNotifications',
    query: { userId: activeUserId.value }
  })
}

const selectedUser = ref(null)
const mutationLoading = ref(false)
const reasonRules = {
  reason: [
    { required: true, message: '请输入操作原因', trigger: 'blur' },
    { min: 4, max: 200, message: '原因长度应为 4-200 字', trigger: 'blur' }
  ]
}

const statusDialogVisible = ref(false)
const statusFormRef = ref(null)
const statusForm = ref({
  status: 'disabled',
  reason: ''
})
const statusImpact = ref(null)
const statusImpactLoading = ref(false)
const statusImpactError = ref('')

const openStatusDialog = async (row) => {
  selectedUser.value = row
  statusForm.value = {
    status: row.status === 'normal' ? 'disabled' : 'normal',
    reason: ''
  }
  statusImpact.value = null
  statusImpactError.value = ''
  statusDialogVisible.value = true
  if (statusForm.value.status !== 'disabled') return
  statusImpactLoading.value = true
  try {
    const detail = unwrapOrderFoodResponse(
      await getOrderFoodUserDetail(row.id),
      '用户禁用影响加载失败'
    )
    statusImpact.value = detail.activeCreatedMealImpact || null
  } catch (error) {
    statusImpactError.value = getOrderFoodErrorMessage(
      error,
      '用户禁用影响加载失败，请重试'
    )
  } finally {
    statusImpactLoading.value = false
  }
}

const handleMutationError = async (error) => {
  if (isOrderFoodConflict(error)) {
    ElMessage.warning('数据已被其他管理员更新，已为你刷新最新状态，请重新确认。')
    await getTableData()
    if (detailVisible.value && activeUserId.value) {
      await loadUserDetail()
    }
    return
  }
  if (!error?.orderFoodResponse) {
    ElMessage.error(getOrderFoodErrorMessage(error))
  }
}

const submitStatus = async () => {
  const valid = await statusFormRef.value?.validate().catch(() => false)
  if (!valid || !selectedUser.value) return

  mutationLoading.value = true
  try {
    const result = unwrapOrderFoodResponse(
      await updateOrderFoodUserStatus(selectedUser.value.id, {
        status: statusForm.value.status,
        reason: statusForm.value.reason,
        expectedVersion: selectedUser.value.version
      }),
      '用户状态更新失败'
    )
    ElMessage.success(
      statusForm.value.status === 'disabled'
        ? `用户已禁用，已取消 ${result.cancelledMealCount || 0} 个饭局并通知 ${result.notifiedParticipantCount || 0} 位参与者`
        : '用户已恢复'
    )
    statusDialogVisible.value = false
    await getTableData()
    if (detailVisible.value && activeUserId.value === selectedUser.value.id) {
      await loadUserDetail()
    }
  } catch (error) {
    await handleMutationError(error)
  } finally {
    mutationLoading.value = false
  }
}

const formatDateTime = (value) => (value ? formatDate(value) : '—')
const booleanLabel = (value) => (value ? '开启' : '关闭')
const pointTypeLabel = (value) =>
  ({ earned: '获得', spent: '消耗', refund: '退款', adjustment: '人工调整' })[
    value
  ] || value || '—'
const aiCapabilityLabel = (value) =>
  ({
    dish_text_extract: '菜品文本解析',
    recipe_image_extract: '菜谱长截图解析',
    dish_cover_create: '菜品封面生成',
    checkin_image_analyze: '打卡图片分析',
    meal_suggest: '饭局菜品建议',
    prep_sequence: '备菜顺序生成'
  })[value] || value || '—'
const aiExecutionLabel = (value) =>
  ({ pending: '等待中', processing: '处理中', succeeded: '成功', failed: '失败' })[
    value
  ] || value || '—'
const aiBillingLabel = (value) =>
  ({
    not_charged: '未扣积分',
    charged: '已扣积分',
    refund_pending: '待退积分',
    refunded: '已退积分'
  })[value] || value || '—'
const notificationTypeLabel = (value) =>
  ({
    governance: '内容处理',
    discoverability: '公开状态',
    points: '积分',
    feature_refund: '功能退积分',
    meal: '饭局'
  })[value] || value || '—'
const userStatusLabel = (value) => ({ normal: '正常', disabled: '已禁用' })[value] || '未知'
const mealStatusLabel = (value) =>
  ({
    collecting: '收集中',
    closed: '已关闭点餐',
    confirmed: '采购进行中'
  })[value] || value || '未知'
const capabilityEffectiveLabel = (value) =>
  ({ enabled: '有效开启', disabled: '有效关闭' })[value] || '未知'
const capabilitySourceLabel = (value) =>
  ({
    emergency: '平台紧急停用',
    platform_default: '平台默认'
  })[value] || '来源未知'
const preferenceStateLabel = (value) =>
  ({ active: '持续更新', pending: '等待聚合', paused: '暂停更新', failed: '最近更新失败' })[
    value
  ] || '状态未知'
const preferenceStateType = (value) =>
  ({ active: 'success', pending: 'info', paused: 'warning', failed: 'danger' })[
    value
  ] || 'info'
const evidenceSourceLabel = (value) =>
  ({
    checkin_image: '打卡图片',
    recommendation_adopted: '采用推荐',
    dish_saved: '保存菜品',
    dish_ordered: '点菜',
    reshuffle: '换一道',
    skip_for_now: '先不考虑',
    explicit_setting: '用户明确设置'
  })[value] || value
const preferenceSettingTypeLabel = (value) =>
  ({
    allergy: '过敏',
    avoidance: '忌口',
    dietary_restriction: '饮食限制',
    preference: '主动偏好'
  })[value] || value

getTableData().then(() => {
  if (route.query.userId) {
    openDetail({ id: String(route.query.userId) })
  }
})
</script>

<style scoped>
.user-detail-drawer {
  overflow: hidden;
  border-radius: 14px 0 0 14px;
  background: var(--el-fill-color-extra-light);
}

.user-detail-drawer :deep(.el-drawer__header) {
  margin: 0;
  padding: 17px 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}

.user-detail-drawer :deep(.el-drawer__body) {
  padding: 0;
  background: var(--el-fill-color-extra-light);
}

.user-detail-header {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding-right: 8px;
}

.user-detail-identity {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
}

.user-detail-avatar {
  flex: none;
  border: 2px solid var(--el-bg-color);
  box-shadow: 0 0 0 1px var(--el-border-color-light);
}

.user-detail-name { min-width: 0; }
.user-detail-title {
  overflow: hidden;
  color: var(--el-text-color-primary);
  font-size: 17px;
  font-weight: 650;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-detail-meta {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.user-detail-meta > span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-detail-actions {
  display: flex;
  flex: none;
  align-items: center;
  gap: 8px;
}

.user-detail-actions .el-button + .el-button,
.overview-metric-actions .el-button + .el-button { margin-left: 0; }
.user-detail-body { min-height: 280px; }

.user-detail-tabs :deep(.el-tabs__header) {
  position: sticky;
  z-index: 3;
  top: 0;
  margin: 0;
  padding: 0 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}

.user-detail-tabs :deep(.el-tabs__nav-wrap::after) { display: none; }
.user-detail-tabs :deep(.el-tabs__item) { height: 46px; padding: 0 17px; }
.user-detail-tabs :deep(.el-tabs__content) { overflow: visible; }
.detail-tab-panel { min-height: 260px; padding: 18px 20px 24px; }

.overview-metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.overview-metric-card {
  display: flex;
  min-height: 116px;
  flex-direction: column;
  align-items: flex-start;
  padding: 15px 16px 13px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-bg-color);
  box-shadow: 0 2px 8px rgb(31 42 61 / 3%);
}

.overview-metric-card.is-primary {
  border-color: var(--el-color-primary-light-8);
  background: linear-gradient(145deg, var(--el-bg-color), var(--el-color-primary-light-9));
}

.overview-metric-label { color: var(--el-text-color-secondary); font-size: 12px; }
.overview-metric-card strong {
  margin-top: 8px;
  color: var(--el-text-color-primary);
  font-size: 25px;
  font-weight: 650;
  line-height: 1;
}

.overview-metric-card.is-primary strong { color: var(--el-color-primary); }
.overview-metric-card > .el-button,
.overview-metric-actions,
.overview-metric-hint { margin-top: auto; }
.overview-metric-actions { display: flex; align-items: center; gap: 10px; }
.overview-metric-hint { color: var(--el-text-color-placeholder); font-size: 12px; line-height: 22px; }

.detail-card {
  margin-top: 14px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-bg-color);
  box-shadow: 0 2px 8px rgb(31 42 61 / 3%);
}

.detail-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.detail-card-header h3 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 650;
  line-height: 1.45;
}

.detail-card-header p {
  margin: 3px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.detail-card-content { padding: 2px 16px 16px; }
.user-info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 28px;
  margin: 0;
  padding: 4px 16px;
}

.user-info-item {
  display: grid;
  min-width: 0;
  grid-template-columns: 82px minmax(0, 1fr);
  align-items: center;
  min-height: 52px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.user-info-item:nth-last-child(-n + 2):not(.is-full),
.user-info-item.is-full { border-bottom: 0; }
.user-info-item.is-full { grid-column: 1 / -1; }
.user-info-item dt { color: var(--el-text-color-secondary); font-size: 13px; }
.user-info-item dd { min-width: 0; margin: 0; color: var(--el-text-color-primary); font-size: 13px; line-height: 1.55; }
.user-info-identity { overflow: hidden; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; text-overflow: ellipsis; white-space: nowrap; }
.user-info-note { margin-left: 6px; color: var(--el-text-color-placeholder); font-size: 12px; }

.capability-effect-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  padding: 14px 16px 16px;
}

.capability-effect-item {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 11px;
  padding: 11px 12px;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  background: var(--el-fill-color-extra-light);
}

.capability-effect-dot {
  width: 7px;
  height: 7px;
  flex: none;
  border-radius: 50%;
  background: var(--el-text-color-placeholder);
  box-shadow: 0 0 0 4px var(--el-fill-color);
}

.capability-effect-item.is-enabled .capability-effect-dot {
  background: var(--el-color-success);
  box-shadow: 0 0 0 4px var(--el-color-success-light-9);
}

.capability-effect-item > div {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.capability-effect-item strong { color: var(--el-text-color-primary); font-size: 13px; font-weight: 500; }
.capability-effect-item span:last-child { color: var(--el-text-color-secondary); font-size: 12px; }
.preference-panel { display: flex; flex-direction: column; gap: 14px; }

.preference-status-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 15px 16px;
  border: 1px solid var(--el-color-primary-light-8);
  border-radius: 10px;
  background: linear-gradient(145deg, var(--el-bg-color), var(--el-color-primary-light-9));
}

.preference-status-copy { display: flex; min-width: 0; align-items: flex-start; gap: 11px; }
.preference-status-copy h3 { margin: 0; color: var(--el-text-color-primary); font-size: 14px; font-weight: 650; line-height: 1.45; }
.preference-status-copy p { margin: 4px 0 0; color: var(--el-text-color-secondary); font-size: 12px; line-height: 1.55; }
.preference-meta-grid { display: grid; flex: none; grid-template-columns: 92px minmax(140px, auto); }
.preference-meta-grid > div { padding: 0 15px; border-left: 1px solid var(--el-border-color-light); }
.preference-meta-grid span,
.preference-meta-grid strong { display: block; }
.preference-meta-grid span { color: var(--el-text-color-secondary); font-size: 11px; }
.preference-meta-grid strong { margin-top: 5px; color: var(--el-text-color-primary); font-size: 12px; font-weight: 600; white-space: nowrap; }
.preference-alert { margin: 0; border-radius: 9px; }

.preference-empty-card {
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-bg-color);
}

.preference-empty-card :deep(.el-empty) { padding: 28px 0 24px; }
.preference-empty-card :deep(.el-empty__description) { margin-top: 10px; }
.preference-empty-card :deep(.el-empty__description p) { color: var(--el-text-color-regular); font-size: 14px; }
.preference-empty-note { margin: -3px 0 0; color: var(--el-text-color-secondary); font-size: 12px; }
.preference-tag-section { margin-top: 14px; }
.preference-tag-section:first-child { margin-top: 12px; }
.preference-section-title { margin-bottom: 8px; color: var(--el-text-color-regular); font-size: 13px; font-weight: 600; }
.preference-tag-list { display: flex; flex-wrap: wrap; gap: 8px; }
.preference-section-empty { color: var(--el-text-color-placeholder); font-size: 13px; }
.preference-summary-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.preference-content-card { border-color: var(--el-border-color-lighter); border-radius: 10px; }
.preference-content-card :deep(.el-card__header) { padding: 14px 16px; color: var(--el-text-color-primary); font-size: 14px; font-weight: 650; }
.preference-content-card :deep(.el-card__body) { padding: 15px 16px; }
.preference-summary { display: flex; flex-direction: column; gap: 10px; }
.preference-summary-text { color: var(--el-text-color-regular); font-size: 13px; line-height: 1.7; }
.evidence-card { margin-top: 0; }
.preference-table { width: 100%; }
.user-detail-tabs :deep(.el-table th.el-table__cell) { background: var(--el-fill-color-extra-light); color: var(--el-text-color-secondary); font-weight: 600; }

.preference-privacy-note {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  padding: 10px 12px;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  background: var(--el-fill-color-extra-light);
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.6;
}

.preference-privacy-icon {
  display: inline-flex;
  width: 17px;
  height: 17px;
  flex: none;
  align-items: center;
  justify-content: center;
  margin-top: 1px;
  border-radius: 50%;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 11px;
  font-weight: 700;
}

@media (max-width: 720px) {
  .user-detail-header { align-items: flex-start; flex-direction: column; }
  .overview-metric-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .user-info-grid,
  .capability-effect-grid,
  .preference-summary-grid { grid-template-columns: 1fr; }
  .user-info-item:nth-last-child(-n + 2):not(.is-full) { border-bottom: 1px solid var(--el-border-color-extra-light); }
  .user-info-item:last-child { border-bottom: 0; }
  .preference-status-card { align-items: flex-start; flex-direction: column; }
  .preference-meta-grid { width: 100%; grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .preference-meta-grid > div:first-child { padding-left: 0; border-left: 0; }
}
</style>
