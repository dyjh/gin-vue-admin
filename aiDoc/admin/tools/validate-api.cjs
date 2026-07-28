const fs = require("fs");
const path = require("path");

const file = path.resolve(__dirname, "..", "api.json");
const api = JSON.parse(fs.readFileSync(file, "utf8"));
const failures = [];
const ids = new Set();
const signatures = new Set();
const permissions = new Set(api.permissions || []);
const typeNames = new Set([
  ...Object.keys(api.models || {}),
  ...Object.keys(api.enums || {}),
  "Page",
  "CNY",
]);
const expectedPageOrder = [
  "dashboard",
  "users",
  "pointEntries",
  "pointRules",
  "userDishes",
  "userRecipes",
  "suggestionCatalog",
  "discoverableDishes",
  "recommendations",
  "officialDishes",
  "categories",
  "tags",
  "units",
  "governanceRecords",
  "media",
  "moderationRecords",
  "moderationConfig",
  "meals",
  "shoppingLists",
  "platformPolicy",
  "aiCapabilities",
  "aiUsages",
  "aiProviders",
  "aiModels",
  "notifications",
  "subscribeLogs",
  "subscribeScenes",
  "wechatConfig",
];
const expectedGroupOrder = [
  "usersAndPoints",
  "dishOperations",
  "contentSafety",
  "mealManagement",
  "aiCapabilities",
  "messageCenter",
];
const expectedGovernanceActionMatrix = {
  dish: ["disable_discoverability", "soft_delete_dish", "delete_copy_chain"],
  official_dish: ["soft_delete_official_dish", "delete_copy_chain"],
  recipe: ["soft_delete_recipe"],
  checkin: ["soft_delete_checkin"],
};

function fail(message) {
  failures.push(message);
}

if (api.basePath !== "/api/orderfood") fail("basePath must be /api/orderfood");
if (Object.keys(api.pages || {}).length !== 28) fail("admin contract must cover 28 pages");
if (JSON.stringify(Object.keys(api.pages || {})) !== JSON.stringify(expectedPageOrder)) {
  fail("admin pages must follow the frozen 28-page product order");
}
if (api.menu?.root) {
  fail("admin menu must not define a 来干饭 business root");
}
if (
  api.menu?.directPages?.dashboard?.routeName !== "OrderFoodDashboard" ||
  api.menu?.directPages?.wechatConfig?.routeName !== "OrderFoodWeChatConfig"
) {
  fail("admin menu must expose data overview and WeChat configuration as direct top-level pages");
}
if (JSON.stringify(Object.keys(api.menu?.groups || {})) !== JSON.stringify(expectedGroupOrder)) {
  fail("admin menu must define the frozen six groups in product order");
}
if (!(api.source || []).includes("aiDoc/prd/order-food-admin-menu-page-review.md")) {
  fail("source must include order-food-admin-menu-page-review.md");
}
if (JSON.stringify(api.governanceActionMatrix) !== JSON.stringify(expectedGovernanceActionMatrix)) {
  fail("governanceActionMatrix must match the frozen target/action rules");
}
if (api.models?.RecommendationCreateInput?.position !== "home_featured|required") {
  fail("recommendation position must be frozen to home_featured");
}
if (api.errorRegistry?.source !== "server/errors/error.go") {
  fail("errorRegistry.source must be server/errors/error.go");
}
if (api.errorRegistry?.range !== "20000-29999") {
  fail("admin error code range must be 20000-29999");
}

for (const code of Object.keys(api.commonErrorCodes || {})) {
  const numeric = Number(code);
  if (!Number.isInteger(numeric) || numeric < 20000 || numeric > 29999) {
    fail(`admin error code is outside 20000-29999: ${code}`);
  }
}

for (const item of api.interfaces || []) {
  if (!item.id) fail("interface id is required");
  if (ids.has(item.id)) fail(`duplicate interface id: ${item.id}`);
  ids.add(item.id);

  const signature = `${item.method} ${item.path}`;
  if (signatures.has(signature)) fail(`duplicate interface signature: ${signature}`);
  signatures.add(signature);

  if (item.auth !== "gva_admin") fail(`${item.id} must use gva_admin auth`);
  if (!item.permission) fail(`${item.id} permission is required`);
  if (!permissions.has(item.permission)) fail(`${item.id} permission is missing from permissions inventory`);
  for (const conditional of item.conditionalPermissions || []) {
    if (!permissions.has(conditional.permission)) {
      fail(`${item.id} conditional permission is missing: ${conditional.permission}`);
    }
  }
  for (const pageId of item.usedBy || []) {
    if (!Object.prototype.hasOwnProperty.call(api.pages || {}, pageId)) {
      fail(`${item.id} references unknown usedBy page ${pageId}`);
    } else if (!(api.pages[pageId].interfaceIds || []).includes(item.id)) {
      fail(`${item.id} usedBy ${pageId} is missing from that page's interfaceIds`);
    }
  }

  if (item.mutation) {
    if (!item.idempotency || item.idempotency.required !== true) {
      fail(`${item.id} mutation must require idempotency`);
    }
  } else if (item.idempotency) {
    fail(`${item.id} non-mutation must not declare idempotency`);
  }

  if (!Array.isArray(item.errors) || item.errors.length === 0) {
    fail(`${item.id} must declare errors`);
  }
  for (const code of item.errors || []) {
    if (!Object.prototype.hasOwnProperty.call(api.commonErrorCodes, String(code))) {
      fail(`${item.id} references unknown error code ${code}`);
    }
  }
}

if (ids.has("recommendation_publish_preview")) {
  fail("recommendation_publish_preview must not remain in the frozen contract");
}

// 配置资源只允许读取当前值、直接更新和诊断测试，不允许重新引入配置工作流。
const directConfigurationRoots = [
  "/point-rules",
  "/suggestion-catalog/policy",
  "/moderation-config",
  "/platform-capability-policy",
  "/ai-capabilities",
];
for (const item of api.interfaces || []) {
  const isConfigurationInterface = directConfigurationRoots.some(
    (root) => item.path === root || item.path.startsWith(`${root}/`),
  );
  if (!isConfigurationInterface) continue;
  if (/\/(draft|publish|publish-preview|versions)(?:\/|$)/.test(item.path)) {
    fail(`${item.id} must use direct configuration update semantics`);
  }
  if (/^orderfood:(point-rule|suggestion-catalog|moderation-config|platform-policy|capability|prompt):publish$/.test(item.permission)) {
    fail(`${item.id} must not require a configuration publish permission`);
  }
}
for (const permission of permissions) {
  if (/^orderfood:(point-rule|suggestion-catalog|moderation-config|platform-policy|capability|prompt):publish$/.test(permission)) {
    fail(`${permission} must not remain in the direct configuration contract`);
  }
}
for (const requiredInterfaceId of [
  "official_dish_cover_list",
  "official_dish_cover_upload",
  "ai_capability_prompt_get",
  "ai_capability_prompt_update",
  "ai_capability_prompt_validation",
  "ai_capability_prompt_render_preview",
  "ai_capability_prompt_test",
]) {
  if (!ids.has(requiredInterfaceId)) fail(`${requiredInterfaceId} is required by the frozen product contract`);
}
if (JSON.stringify(api.enums?.AiProviderType) !== JSON.stringify(["bailian", "deepseek", "openai"])) {
  fail("AiProviderType must be bailian, deepseek and openai");
}
if (!(api.enums?.FeatureBillingStatus || []).includes("refund_pending")) {
  fail("FeatureBillingStatus must include refund_pending");
}
if (Object.prototype.hasOwnProperty.call(api.models?.UserSummary || {}, "capabilityOverride")) {
  fail("UserSummary must not expose a per-user capability override");
}
if (Object.prototype.hasOwnProperty.call(api.models?.AiCapabilityUpdateInput || {}, "fallbackModelId")) {
  fail("AiCapabilityUpdateInput must not expose a fallback model");
}
if (Object.prototype.hasOwnProperty.call(api.models?.AiCapabilityWorkspace || {}, "prompt")) {
  fail("AiCapabilityWorkspace must not bypass the independent prompt:read permission");
}
for (const modelName of ["AiProviderCreateInput", "AiModelCreateInput", "SubscribeSceneConfigureInput"]) {
  if (Object.prototype.hasOwnProperty.call(api.models?.[modelName] || {}, "enabled")) {
    fail(`${modelName} must not expose enabled`);
  }
}
if (
  api.models?.SubscribeSceneConfigureInput?.fieldMappings !==
  "object|required|minProperties:3|maxProperties:3|requiredKeys:mealName,result,resultAt|maxKeyLength:64|maxValueLength:40"
) {
  fail("the meal_status scene must expose exactly three fixed field mappings");
}
for (const forbiddenId of [
  "subscribe_template_create",
  "subscribe_template_update",
  "subscribe_template_delete",
]) {
  if (ids.has(forbiddenId)) fail(`${forbiddenId} must not remain in the fixed scene contract`);
}
for (const requiredId of [
  "subscribe_scene_list",
  "subscribe_scene_detail",
  "subscribe_scene_configure",
  "subscribe_scene_status_update",
]) {
  if (!ids.has(requiredId)) fail(`${requiredId} is required by the fixed scene contract`);
}
for (const item of api.interfaces || []) {
  if (item.path.startsWith("/subscribe-templates")) {
    fail(`${item.id} must use the fixed /subscribe-scenes resource`);
  }
}
if (
  api.models?.AiUsageSummary?.promptMode !== "default|custom" ||
  api.models?.AiUsageSummary?.promptHash !== "string"
) {
  fail("AiUsageSummary must retain prompt mode and hash without prompt body snapshots");
}
if (api.roleBootstrap?.firstInjectedAdministratorReceivesAllPermissions !== true) {
  fail("the first injected administrator must receive all orderfood permissions");
}

function validateExplicitTypeReferences(value, location) {
  if (Array.isArray(value)) {
    value.forEach((item, index) => validateExplicitTypeReferences(item, `${location}[${index}]`));
    return;
  }
  if (value && typeof value === "object") {
    if (typeof value.$ref === "string") {
      const referenced = value.$ref
        .replace(/^Page<(.+)>$/, "$1")
        .replace(/\[\]$/, "");
      if (!typeNames.has(referenced)) fail(`${location} references unknown type ${value.$ref}`);
    }
    if (Array.isArray(value.allOf)) {
      for (const referenced of value.allOf) {
        if (!typeNames.has(referenced)) fail(`${location}.allOf references unknown type ${referenced}`);
      }
    }
    for (const [key, child] of Object.entries(value)) {
      validateExplicitTypeReferences(child, `${location}.${key}`);
    }
  }
}

validateExplicitTypeReferences(api.models, "models");
validateExplicitTypeReferences(api.interfaces, "interfaces");

const seenPageNumbers = new Set();
const seenRouteNames = new Set();
const seenPagePositions = new Set();
for (const [pageId, page] of Object.entries(api.pages || {})) {
  if (!page.component.startsWith("view/orderFood/")) {
    fail(`${pageId} component must be under view/orderFood`);
  }
  if (!/^P\d{2}$/.test(page.pageNumber || "")) {
    fail(`${pageId} pageNumber must use the Pxx format`);
  } else if (seenPageNumbers.has(page.pageNumber)) {
    fail(`${pageId} duplicates pageNumber ${page.pageNumber}`);
  }
  seenPageNumbers.add(page.pageNumber);
  const isDirectPage = page.level === "top";
  if (isDirectPage) {
    if (!Object.prototype.hasOwnProperty.call(api.menu?.directPages || {}, pageId)) {
      fail(`${pageId} is top-level but missing from menu.directPages`);
    }
  } else if (!Object.prototype.hasOwnProperty.call(api.menu?.groups || {}, page.group)) {
    fail(`${pageId} references unknown menu group ${page.group}`);
  }
  if (!page.path || !page.routeName || !Number.isInteger(page.sortOrder) || page.sortOrder < 1) {
    fail(`${pageId} must define path, routeName and positive sortOrder`);
  }
  if (seenRouteNames.has(page.routeName)) fail(`${pageId} duplicates routeName ${page.routeName}`);
  seenRouteNames.add(page.routeName);
  const pagePosition = `${isDirectPage ? "top" : page.group}:${page.sortOrder}`;
  if (seenPagePositions.has(pagePosition)) {
    fail(`${pageId} duplicates menu position ${pagePosition}`);
  }
  seenPagePositions.add(pagePosition);
  if (!permissions.has(page.permission)) {
    fail(`${pageId} permission is missing: ${page.permission}`);
  }
  for (const interfaceId of page.interfaceIds || []) {
    if (!ids.has(interfaceId)) {
      fail(`${pageId} references unknown interface ${interfaceId}`);
      continue;
    }
    const referencedInterface = (api.interfaces || []).find((item) => item.id === interfaceId);
    if (!(referencedInterface.usedBy || []).includes(pageId)) {
      fail(`${pageId} references ${interfaceId}, but the interface usedBy omits the page`);
    }
  }
}

const governanceRole = api.roles?.orderfood_governance;
for (const required of [
  "orderfood:user-dish:read",
  "orderfood:user-dish:private-read",
  "orderfood:user-recipe:read",
  "orderfood:user-recipe:private-read",
  "orderfood:moderation-config:read",
]) {
  if (!governanceRole?.permissions?.includes(required)) {
    fail(`orderfood_governance must include ${required}`);
  }
}
for (const forbidden of [
  "orderfood:user:preference:read",
  "orderfood:moderation-config:update",
  "orderfood:moderation-config:credential-write",
  "orderfood:moderation-config:test",
  "orderfood:governance:cascade",
]) {
  if (governanceRole?.permissions?.includes(forbidden)) {
    fail(`orderfood_governance must not include ${forbidden}`);
  }
}

for (const [roleCode, role] of Object.entries(api.roles || {})) {
  const seen = new Set();
  for (const permission of role.permissions || []) {
    if (!permissions.has(permission)) fail(`${roleCode} references unknown permission ${permission}`);
    if (seen.has(permission)) fail(`${roleCode} has duplicate permission ${permission}`);
    seen.add(permission);
  }
}

if (failures.length) {
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log(
  JSON.stringify(
    {
      file,
      pages: Object.keys(api.pages).length,
      interfaces: api.interfaces.length,
      models: Object.keys(api.models).length,
      permissions: api.permissions.length,
      roles: Object.keys(api.roles).length,
    },
    null,
    2
  )
);
