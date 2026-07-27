#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");

const root = path.resolve(__dirname, "../../..");
const parser = require(path.join(root, "web/node_modules/@babel/parser"));
const contract = require(path.join(root, "aiDoc/admin/api.json"));
const swagger = require(path.join(root, "server/docs/swagger.json"));

const failures = [];
const fail = (message) => failures.push(message);
const normalizedPath = (value) => value.replace(/\{[^}]+\}/g, "{}");
const endpointKey = (method, endpointPath) =>
  `${String(method).toUpperCase()} ${normalizedPath(endpointPath)}`;
const escapeRegExp = (value) =>
  value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

// 校验管理端契约中的每个页面均有真实 Vue 组件和菜单种子。
const menuSeedSource = fs.readFileSync(
  path.join(root, "server/initialize/orderfood_admin_seed.go"),
  "utf8",
);
for (const [pageId, page] of Object.entries(contract.pages || {})) {
  const componentFile = path.join(root, "web/src", page.component);
  if (!fs.existsSync(componentFile)) {
    fail(`${pageId} component is missing: ${page.component}`);
  }
  if (
    !new RegExp(`Name:\\s*"${escapeRegExp(page.routeName)}"`).test(
      menuSeedSource,
    )
  ) {
    fail(`${pageId} route is missing from the GVA menu seed: ${page.routeName}`);
  }
  if (
    !new RegExp(`Component:\\s*"${escapeRegExp(page.component)}"`).test(
      menuSeedSource,
    )
  ) {
    fail(`${pageId} component is missing from the GVA menu seed: ${page.component}`);
  }
}

// 校验 Swagger 与冻结契约的 method + path 集合完全一致，禁止漏接口和游离接口。
const contractEndpoints = new Map();
for (const item of contract.interfaces || []) {
  contractEndpoints.set(
    endpointKey(item.method, `/orderfood${item.path}`),
    item,
  );
}
const swaggerEndpoints = new Set();
for (const [endpointPath, operations] of Object.entries(swagger.paths || {})) {
  if (!endpointPath.startsWith("/orderfood/")) continue;
  for (const method of Object.keys(operations)) {
    if (!["get", "post", "put", "delete", "patch"].includes(method)) continue;
    swaggerEndpoints.add(endpointKey(method, endpointPath));
  }
}
for (const [key, item] of contractEndpoints) {
  if (!swaggerEndpoints.has(key)) {
    fail(`${item.id} is missing from Swagger: ${key}`);
  }
}
for (const key of swaggerEndpoints) {
  if (!contractEndpoints.has(key)) {
    fail(`Swagger contains an interface outside the frozen contract: ${key}`);
  }
}

// expressionPath 将 API 封装中的静态路径和模板路径转换为可比较表达式。
const expressionPath = (node) => {
  if (!node) return null;
  if (node.type === "StringLiteral") return node.value;
  if (node.type === "TemplateLiteral") {
    let result = "";
    for (let index = 0; index < node.quasis.length; index += 1) {
      result += node.quasis[index].value.cooked ?? node.quasis[index].value.raw;
      if (index >= node.expressions.length) continue;
      const expression = node.expressions[index];
      if (
        expression.type === "CallExpression" &&
        expression.callee.type === "Identifier" &&
        expression.callee.name === "getResourcePath"
      ) {
        result += "{resource}";
      } else {
        result += "{}";
      }
    }
    return result;
  }
  if (node.type === "BinaryExpression" && node.operator === "+") {
    const left = expressionPath(node.left);
    const right = expressionPath(node.right);
    return left === null || right === null ? null : left + right;
  }
  return null;
};

// objectProperty 读取对象字面量中的普通属性。
const objectProperty = (objectNode, propertyName) => {
  if (!objectNode || objectNode.type !== "ObjectExpression") return null;
  return (
    objectNode.properties.find((property) => {
      if (property.type !== "ObjectProperty") return false;
      if (property.key.type === "Identifier") {
        return property.key.name === propertyName;
      }
      return property.key.type === "StringLiteral" &&
        property.key.value === propertyName;
    }) || null
  );
};

const webCalls = [];
const webExports = [];
const apiDirectory = path.join(root, "web/src/api/orderfood");
for (const fileName of fs.readdirSync(apiDirectory)) {
  if (!fileName.endsWith(".js") || fileName === "request.js") continue;
  const filePath = path.join(apiDirectory, fileName);
  const source = fs.readFileSync(filePath, "utf8");
  for (const match of source.matchAll(/export const (\w+)\s*=/g)) {
    webExports.push({ fileName, name: match[1] });
  }
  const ast = parser.parse(source, {
    sourceType: "module",
    plugins: ["optionalChaining"],
  });

  // walk 递归遍历 Babel AST，定位全部 orderFoodRequest 调用。
  const walk = (node) => {
    if (!node || typeof node !== "object") return;
    if (
      node.type === "CallExpression" &&
      node.callee?.type === "Identifier" &&
      node.callee.name === "orderFoodRequest"
    ) {
      const input = node.arguments[0];
      const pathProperty = objectProperty(input, "path");
      const methodProperty = objectProperty(input, "method");
      const mutationProperty = objectProperty(input, "mutation");
      const requestPath = expressionPath(pathProperty?.value);
      const method =
        methodProperty?.value?.type === "StringLiteral"
          ? methodProperty.value.value
          : null;
      const mutation =
        mutationProperty?.value?.type === "BooleanLiteral"
          ? mutationProperty.value.value
          : false;
      if (!requestPath || !method) {
        fail(`${fileName} contains an unreadable orderFoodRequest call`);
      } else {
        webCalls.push({ fileName, method, path: requestPath, mutation });
      }
    }
    for (const value of Object.values(node)) {
      if (Array.isArray(value)) {
        for (const child of value) walk(child);
      } else if (value && typeof value === "object" && value.type) {
        walk(value);
      }
    }
  };
  walk(ast);
}

const resourcePaths = ["categories", "tags", "units"];
const expandedWebCalls = webCalls.flatMap((item) => {
  if (!item.path.includes("{resource}")) return [item];
  return resourcePaths.map((resource) => ({
    ...item,
    path: item.path.replace("{resource}", resource),
  }));
});
const webEndpoints = new Map();
for (const item of expandedWebCalls) {
  const key = endpointKey(item.method, `/orderfood${item.path}`);
  if (webEndpoints.has(key)) {
    fail(
      `Web API endpoint is wrapped more than once: ${key} in ${webEndpoints.get(key).fileName} and ${item.fileName}`,
    );
    continue;
  }
  webEndpoints.set(key, item);
}
for (const [key, item] of contractEndpoints) {
  const webCall = webEndpoints.get(key);
  if (!webCall) {
    fail(`${item.id} is missing from web/src/api/orderfood: ${key}`);
    continue;
  }
  if (Boolean(item.mutation) !== webCall.mutation) {
    fail(
      `${item.id} mutation flag mismatch: contract=${Boolean(item.mutation)} web=${webCall.mutation}`,
    );
  }
}
for (const [key, item] of webEndpoints) {
  if (!contractEndpoints.has(key)) {
    fail(`Web API wrapper is outside the frozen contract: ${key} (${item.fileName})`);
  }
}

// 校验每个 API 封装都由管理端页面或共享业务组件实际使用。
const orderFoodViewSources = [];
const collectVueSources = (directory) => {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const entryPath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      collectVueSources(entryPath);
    } else if (entry.isFile() && entry.name.endsWith(".vue")) {
      orderFoodViewSources.push(fs.readFileSync(entryPath, "utf8"));
    }
  }
};
collectVueSources(path.join(root, "web/src/view/orderFood"));
const allOrderFoodViews = orderFoodViewSources.join("\n");
for (const item of webExports) {
  const usagePattern = new RegExp(`\\b${escapeRegExp(item.name)}\\b`);
  if (!usagePattern.test(allOrderFoodViews)) {
    fail(`Web API wrapper is not used by any orderfood page: ${item.name} (${item.fileName})`);
  }
}

// 校验点餐管理端不再复用浏览器本地时区格式化，并固定使用上海业务时区。
const orderFoodTimeSource = fs.readFileSync(
  path.join(root, "web/src/view/orderFood/utils/time.js"),
  "utf8",
);
if (!orderFoodTimeSource.includes("Asia/Shanghai")) {
  fail("OrderFood time utility must use Asia/Shanghai");
}
if (allOrderFoodViews.includes("from '@/utils/format'")) {
  fail("OrderFood pages must not use the browser-local formatDate utility");
}
for (const forbiddenPattern of [
  /value-format="YYYY-MM-DDTHH:mm:ssZ"/,
  /\.toISOString\(\)/,
]) {
  if (forbiddenPattern.test(allOrderFoodViews)) {
    fail(`OrderFood pages contain a browser-timezone-dependent date filter: ${forbiddenPattern}`);
  }
}

// 校验冻结为详情标签的对象审计入口都已接入，避免只有服务端写日志但页面不可核对。
const auditPanelPages = {
  "user/index.vue": 'target-type="user"',
  "userDish/index.vue": 'target-type="dish"',
  "userRecipe/index.vue": 'target-type="recipe"',
  "moderationConfig/index.vue": 'target-type="moderation_config"',
  "platformPolicy/index.vue": 'target-type="platform_policy"',
  "aiCapability/index.vue": 'target-type="ai_capability"',
  "aiProvider/index.vue": 'target-type="ai_provider"',
  "aiModel/index.vue": 'target-type="ai_model"',
};
for (const [relativePath, targetFragment] of Object.entries(auditPanelPages)) {
  const source = fs.readFileSync(
    path.join(root, "web/src/view/orderFood", relativePath),
    "utf8",
  );
  if (!source.includes("<AuditPanel") || !source.includes(targetFragment)) {
    fail(`${relativePath} is missing its object audit panel`);
  }
  if (
    !source.includes("orderfood:audit:read") ||
    !source.includes("canReadAudit")
  ) {
    fail(`${relativePath} must hide its audit panel without audit:read`);
  }
}

// 校验用户详情已闭环展示冻结契约要求的用户维度数据，并在完整列表页保留用户筛选。
const userViewSource = fs.readFileSync(
  path.join(root, "web/src/view/orderFood/user/index.vue"),
  "utf8",
);
for (const requiredFragment of [
  'name="points"',
  'name="aiUsage"',
  'name="notifications"',
  "getPointEntryList",
  "getAIUsageList",
  "getOrderFoodNotificationList",
  "orderfood:points:read",
  "orderfood:ai-usage:read",
  "orderfood:notification:read",
  "openPointAdjustmentFor(row)",
  "openStatusDialog(userDetail)",
  "userDetail.mealCount",
  "pageSize: 5",
]) {
  if (!userViewSource.includes(requiredFragment)) {
    fail(`user/index.vue is missing the user detail closure: ${requiredFragment}`);
  }
}
const notificationViewSource = fs.readFileSync(
  path.join(root, "web/src/view/orderFood/notification/index.vue"),
  "utf8",
);
if (!notificationViewSource.includes("route.query.userId")) {
  fail("notification/index.vue must preserve the userId drill-down filter");
}

// 校验推荐草稿只能从来源页面创建，推荐精选页本身不提供统一新建入口。
const recommendationViewSource = fs.readFileSync(
  path.join(root, "web/src/view/orderFood/recommendation/index.vue"),
  "utf8",
);
for (const forbiddenFragment of [
  "createRecommendation",
  "canCreate",
  "openCreate",
  "创建推荐草稿",
]) {
  if (recommendationViewSource.includes(forbiddenFragment)) {
    fail(`recommendation/index.vue violates source-bound creation: ${forbiddenFragment}`);
  }
}
for (const requiredFragment of [
  "违规处理",
  "entryRecommendationId",
  "soft_delete_official_dish",
  "previewOrderFoodGovernanceAction",
  "executeOrderFoodGovernanceAction",
]) {
  if (!recommendationViewSource.includes(requiredFragment)) {
    fail(`recommendation/index.vue is missing source-linked governance: ${requiredFragment}`);
  }
}

// 校验AI排障链路能按能力、供应商和模型精确下钻，且积分流水保留关联对象筛选。
const aiUsageViewSource = fs.readFileSync(
  path.join(root, "web/src/view/orderFood/aiUsage/index.vue"),
  "utf8",
);
for (const requiredFragment of [
  "searchInfo.providerId",
  "searchInfo.modelId",
  "searchInfo.idempotencyKey",
  "openUser",
  "openPointEntries",
  "openCapability",
  "openProvider",
  "openModel",
  "copyRequestID",
]) {
  if (!aiUsageViewSource.includes(requiredFragment)) {
    fail(`aiUsage/index.vue is missing the diagnostic drill-down: ${requiredFragment}`);
  }
}
const aiDrillDownPages = {
  "aiCapability/index.vue": ["openCapabilityUsages", "route.query.capabilityCode"],
  "aiProvider/index.vue": [
    "openProviderModels",
    "openProviderUsages",
    "route.query.providerId",
  ],
  "aiModel/index.vue": [
    "openModelProvider",
    "openModelUsages",
    "route.query.providerId",
    "route.query.modelId",
  ],
  "points/index.vue": [
    "route.query.relatedObjectType",
    "route.query.relatedObjectId",
    "canReadUsers",
    "canReadAIUsage",
    "isAIUsageEntry",
    "openRelatedRecord",
    "OrderFoodAIUsages",
  ],
};
for (const [relativePath, requiredFragments] of Object.entries(aiDrillDownPages)) {
  const source = fs.readFileSync(
    path.join(root, "web/src/view/orderFood", relativePath),
    "utf8",
  );
  for (const requiredFragment of requiredFragments) {
    if (!source.includes(requiredFragment)) {
      fail(`${relativePath} is missing the diagnostic drill-down: ${requiredFragment}`);
    }
  }
}
const aiUsageModel = contract.models?.AiUsageSummary || {};
if (!aiUsageModel.providerId || !aiUsageModel.modelId) {
  fail("AiUsageSummary must expose actual providerId and modelId for exact drill-down");
}

// 校验消息详情可以按权限返回关联用户、业务目标、模板和饭局，不停留在不可操作的文本ID。
const notificationClosurePages = {
  "notification/index.vue": [
    "canReadUsers",
    "canReadSubscribeLogs",
    "canOpenTarget",
    "openTarget",
    "OrderFoodGovernanceRecords",
    "OrderFoodPointEntries",
  ],
  "subscribeLog/index.vue": [
    "canReadUsers",
    "canReadTemplates",
    "canReadMeals",
    "openMeal",
    "detail.relatedMealId",
  ],
  "governance/index.vue": [
    "route.query.targetType",
    "route.query.targetId",
  ],
  "userDish/index.vue": ["route.query.openDishId"],
  "userRecipe/index.vue": ["route.query.openRecipeId"],
  "recommendation/index.vue": ["route.query.recommendationId"],
};
for (const [relativePath, requiredFragments] of Object.entries(
  notificationClosurePages,
)) {
  const source = fs.readFileSync(
    path.join(root, "web/src/view/orderFood", relativePath),
    "utf8",
  );
  for (const requiredFragment of requiredFragments) {
    if (!source.includes(requiredFragment)) {
      fail(`${relativePath} is missing the message target closure: ${requiredFragment}`);
    }
  }
}
if (!contract.models?.SubscribeLogDetail?.relatedMealId) {
  fail("SubscribeLogDetail must expose relatedMealId for the meal drill-down");
}

// 校验积分、内容安全和饭局管理详情页提供冻结页面契约要求的对象跳转与排障操作。
const productClosurePages = {
  "points/index.vue": [
    "查看关联记录",
    "canReadAIUsage",
    "openRelatedRecord",
  ],
  "meal/index.vue": [
    "查看创建者",
    "查看采购清单",
    "复制饭局 ID",
    "canReadUsers",
    "canReadShopping",
    "searchInfo.createdRange",
    "searchInfo.deadlineRange",
    "row.finalDishCount",
    "row.confirmedAt",
    "row.completedAt",
    "cancelledFromStatus",
    "shareRevocationStatus",
    "closeSourceLabel",
    "formatSnapshot",
  ],
  "shopping/index.vue": [
    "查看饭局",
    "查看创建者",
    "复制清单 ID",
    "canReadMeals",
    "canReadUsers",
    "searchInfo.createdRange",
    "row.totalCount",
    "row.createdAt",
    "detail.shareRevokedAt",
  ],
  "media/index.vue": [
    "查看上传用户",
    "查看绑定对象",
    "canOpenUploader",
    "canOpenBoundObject",
    "soft_delete_checkin",
    "canGovernObject",
  ],
  "moderation/index.vue": [
    "查看关联对象",
    "canOpenObject",
    "row.providerRequestId",
    "row.user",
  ],
  "governance/index.vue": [
    "对同一目标追加处理",
    "canAppendProcessing",
    "openAppendProcessing",
    "previewAppendProcessing",
    "executeAppendProcessing",
    "activeJob.items",
    "searchInfo.createdRange",
  ],
  "discoverableDish/index.vue": [
    "查看作者",
    "精选到推荐",
    "违规处理",
    "searchInfo.authorKeyword",
    "searchInfo.tagInput",
    "searchInfo.discoverableRange",
    "includeModerationSummary: canReadModeration.value",
    "detail.moderationSummary",
    "detail.recommendation",
    "!row.sourceLocked",
    "previewOrderFoodGovernanceAction",
    "executeOrderFoodGovernanceAction",
  ],
  "officialDish/index.vue": [
    "searchInfo.tagIds",
    "searchInfo.createdRange",
    "row.updatedBy",
    "row.onlineRecommendationCount",
    "submitEditor('draft')",
    "submitEditor('usable')",
    "管理端上传不走图片审核",
  ],
  "userDish/index.vue": [
    "查看用户",
    "转到候选池",
    "查看推荐",
    "canOpenSource",
    "searchInfo.userKeyword",
    "detail.deletedBy",
    "referenceAccessAuditId",
    "openGovernanceRecords",
  ],
  "userRecipe/index.vue": [
    "查看用户",
    "openUser",
    "canReadUsers",
    "searchInfo.userKeyword",
    "openRecipeDishes",
    "item.relationId",
    "item.addedAt",
    "detail.deletedBy",
    "openGovernanceRecords",
  ],
  "subscribeTemplate/index.vue": ["canReadLogs", "orderfood:subscribe-log:read"],
  "components/PointAdjustmentDialog.vue": [
    "canReadUsers",
    "canAdjust",
    "debitExceedsBalance",
    "previewPointAdjustment",
    "createPointAdjustment",
  ],
  "suggestionCatalog/index.vue": [
    "administratorLabel(workspace?.policy?.updatedBy)",
    "workspace?.policy?.reason",
    "resetDishes",
    "resetIngredients",
    "workspace?.policy?.retryCount",
    "workspace?.readiness?.ready",
  ],
};
for (const [relativePath, requiredFragments] of Object.entries(
  productClosurePages,
)) {
  const source = fs.readFileSync(
    path.join(root, "web/src/view/orderFood", relativePath),
    "utf8",
  );
  for (const requiredFragment of requiredFragments) {
    if (!source.includes(requiredFragment)) {
      fail(`${relativePath} is missing the product closure: ${requiredFragment}`);
    }
  }
}
if (!contract.models?.GovernanceRecordDetail?.targetExists) {
  fail("GovernanceRecordDetail must expose targetExists");
}
if (!contract.models?.MealAdminSummary?.finalDishCount) {
  fail("MealAdminSummary must expose finalDishCount");
}
for (const field of ["boundObjectVersion", "boundObjectDeleted"]) {
  if (!contract.models?.MediaSummary?.[field]) {
    fail(`MediaSummary must expose ${field} for bound-object governance`);
  }
}
const discoverableList = contract.interfaces?.find(
  (item) => item.id === "discoverable_dish_list",
);
if (!discoverableList?.request?.query?.dishId) {
  fail("discoverable dish list must support exact dishId drill-down");
}
for (const field of [
  "user",
  "objectType",
  "objectId",
  "providerRequestId",
]) {
  if (!contract.models?.ModerationRecordSummary?.[field]) {
    fail(`ModerationRecordSummary must expose ${field}`);
  }
}

// 校验运营概览的明细入口同时具备目标权限保护和可被目标页面识别的精确筛选。
const dashboardViewSource = fs.readFileSync(
  path.join(root, "web/src/view/orderFood/dashboard/index.vue"),
  "utf8",
);
for (const requiredFragment of [
  "targetPermissions",
  "canOpenTarget",
  "orderfood:user:read",
  "orderfood:meal:read",
  "orderfood:shopping:read",
  "orderfood:media:read",
  "orderfood:governance:read",
  "orderfood:moderation:read",
  "orderfood:subscribe-log:read",
]) {
  if (!dashboardViewSource.includes(requiredFragment)) {
    fail(`dashboard/index.vue is missing the permission-safe drill-down: ${requiredFragment}`);
  }
}
const dashboardServiceSource = fs.readFileSync(
  path.join(root, "server/service/orderfood/dashboard.go"),
  "utf8",
);
for (const requiredFragment of [
  '"timeField": "created"',
  '"timeField": "joined"',
  '"timeField": "confirmed"',
  '"registeredRange": rangeCode',
  '"activeRange": rangeCode',
  '"capabilityEffective": "enabled"',
  '"jobStatus": "abnormal"',
  '"status": "rejected_or_failed"',
]) {
  if (!dashboardServiceSource.includes(requiredFragment)) {
    fail(`dashboard service is missing the exact target query: ${requiredFragment}`);
  }
}
const dashboardDrillDownPages = {
  "user/index.vue": [
    "route.query.registeredRange",
    "route.query.activeRange",
    "route.query.capabilityEffective",
    "searchInfo.activeRange",
    "activeFrom",
    "activeTo",
  ],
  "meal/index.vue": [
    "route.query.timeField",
    "route.query.range",
    "searchInfo.joinedRange",
    "searchInfo.confirmedRange",
    "joinedFrom",
    "confirmedFrom",
  ],
  "shopping/index.vue": ["route.query.range", "getShanghaiPresetRange"],
  "media/index.vue": [
    "route.query.sourceScene",
    "route.query.boundObjectType",
    "route.query.range",
  ],
  "governance/index.vue": ["route.query.jobStatus", "value: 'abnormal'"],
  "moderation/index.vue": [
    "route.query.status",
    "route.query.range",
    "rejected_or_failed",
  ],
  "subscribeLog/index.vue": [
    "route.query.status",
    "route.query.range",
    "getShanghaiPresetRange",
  ],
};
for (const [relativePath, requiredFragments] of Object.entries(
  dashboardDrillDownPages,
)) {
  const source = fs.readFileSync(
    path.join(root, "web/src/view/orderFood", relativePath),
    "utf8",
  );
  for (const requiredFragment of requiredFragments) {
    if (!source.includes(requiredFragment)) {
      fail(`${relativePath} is missing the dashboard drill-down: ${requiredFragment}`);
    }
  }
}
const requiredDashboardQueryFields = {
  user_list: ["activeFrom", "activeTo"],
  meal_list: [
    "joinedFrom",
    "joinedTo",
    "confirmedFrom",
    "confirmedTo",
    "completedFrom",
    "completedTo",
  ],
};
for (const [interfaceId, fields] of Object.entries(requiredDashboardQueryFields)) {
  const item = contract.interfaces?.find((candidate) => candidate.id === interfaceId);
  for (const field of fields) {
    if (!item?.request?.query?.[field]) {
      fail(`${interfaceId} must expose ${field} for dashboard drill-down`);
    }
  }
}

// 校验业务权限集合与 GVA 按钮权限的默认模板使用同一组编码。
const permissionSource = fs.readFileSync(
  path.join(root, "server/service/orderfood/permission.go"),
  "utf8",
);
const sourcePermissions = new Set(
  [...permissionSource.matchAll(/"((?:orderfood:)[a-z0-9:-]+)"/g)].map(
    (match) => match[1],
  ),
);
const contractPermissions = new Set(contract.permissions || []);
for (const permission of contractPermissions) {
  if (!sourcePermissions.has(permission)) {
    fail(`permission is missing from DefaultRolePermissionMatrix: ${permission}`);
  }
}
for (const permission of sourcePermissions) {
  if (!contractPermissions.has(permission)) {
    fail(`DefaultRolePermissionMatrix contains an unknown permission: ${permission}`);
  }
}
if (
  !permissionSource.includes("Model(&system.SysAuthorityBtn{})") ||
  !permissionSource.includes("JOIN sys_base_menu_btns")
) {
  fail("PermissionService must read GVA button permissions as the business truth");
}

if (failures.length > 0) {
  for (const failure of failures) {
    process.stderr.write(`- ${failure}\n`);
  }
  process.exit(1);
}

process.stdout.write(
  `${JSON.stringify(
    {
      pages: Object.keys(contract.pages || {}).length,
      contractInterfaces: contractEndpoints.size,
      swaggerInterfaces: swaggerEndpoints.size,
      webInterfaces: webEndpoints.size,
      usedWebExports: webExports.length,
      permissions: contractPermissions.size,
    },
    null,
    2,
  )}\n`,
);
