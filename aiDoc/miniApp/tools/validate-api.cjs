const fs = require("fs");
const path = require("path");

const file = path.resolve(__dirname, "..", "api.json");
const miniAppRoot = path.resolve(__dirname, "..", "..", "..", "miniApp");
const api = JSON.parse(fs.readFileSync(file, "utf8"));
const app = JSON.parse(fs.readFileSync(path.join(miniAppRoot, "app.json"), "utf8"));
const failures = [];
const interfaceIds = new Set();
const signatures = new Set();
const knownPages = new Set(["app.js", ...(app.pages || [])]);
const knownErrors = new Set(Object.keys(api.commonErrorCodes || {}).map(Number));
const typeNames = new Set([
  ...Object.keys(api.models || {}),
  ...Object.keys(api.enums || {}),
  "Page",
]);

function fail(message) {
  failures.push(message);
}

if (api.basePath !== "/api/miniapp/v1") fail("basePath must be /api/miniapp/v1");
if ((api.interfaces || []).length !== 67) fail("mini-program contract must contain exactly 67 interfaces");
if (api.errorRegistry?.source !== "server/errors/error.go") {
  fail("errorRegistry.source must be server/errors/error.go");
}
if (api.errorRegistry?.range !== "40000-49999") {
  fail("mini-program error code range must be 40000-49999");
}
if (api.requestValidation?.source !== "server/utils/front_validator.go") {
  fail("requestValidation.source must be server/utils/front_validator.go");
}
if (api.requestValidation?.entry !== "utils.VerifyAll") {
  fail("requestValidation.entry must be utils.VerifyAll");
}

for (const code of knownErrors) {
  if (!Number.isInteger(code) || code < 40000 || code > 49999) {
    fail(`front error code is outside 40000-49999: ${code}`);
  }
}

for (const item of api.interfaces || []) {
  if (!item.id) fail("interface id is required");
  if (interfaceIds.has(item.id)) fail(`duplicate interface id: ${item.id}`);
  interfaceIds.add(item.id);

  const signature = `${item.method} ${item.path}`;
  if (signatures.has(signature)) fail(`duplicate interface signature: ${signature}`);
  signatures.add(signature);

  for (const pageId of item.usedBy || []) {
    if (!knownPages.has(pageId)) fail(`${item.id} references unknown page ${pageId}`);
  }

  const mutation = !["GET", "HEAD", "OPTIONS"].includes(item.method);
  if (mutation && item.idempotency?.required !== true) {
    fail(`${item.id} mutation must require X-Idempotency-Key`);
  }
  if (!mutation && item.idempotency) {
    fail(`${item.id} read endpoint must not declare idempotency`);
  }
  if (!Array.isArray(item.errors) || !item.errors.length) {
    fail(`${item.id} must declare errors`);
  }
  for (const code of item.errors || []) {
    if (!knownErrors.has(code)) fail(`${item.id} references unknown error code ${code}`);
  }
}

const publicSignatures = (api.interfaces || [])
  .filter((item) => item.auth === false)
  .map((item) => `${item.method} ${item.path}`)
  .sort();
const expectedPublicSignatures = [
  "GET /public/shopping-lists/{shareToken}",
  "POST /auth/wx-login",
].sort();
if (JSON.stringify(publicSignatures) !== JSON.stringify(expectedPublicSignatures)) {
  fail("only wx-login and public shopping-list share may bypass authentication");
}

function validateExplicitTypeReferences(value, location) {
  if (Array.isArray(value)) {
    value.forEach((item, index) => validateExplicitTypeReferences(item, `${location}[${index}]`));
    return;
  }
  if (!value || typeof value !== "object") return;
  if (typeof value.$ref === "string") {
    const referenced = value.$ref.replace(/^Page<(.+)>$/, "$1");
    if (!typeNames.has(referenced)) fail(`${location} references unknown type ${value.$ref}`);
  }
  for (const referenced of value.allOf || []) {
    if (!typeNames.has(referenced)) fail(`${location}.allOf references unknown type ${referenced}`);
  }
  for (const [key, child] of Object.entries(value)) {
    validateExplicitTypeReferences(child, `${location}.${key}`);
  }
}

validateExplicitTypeReferences(api.models, "models");
validateExplicitTypeReferences(api.interfaces, "interfaces");

if (api.clientAuthentication?.tokenStorageKey !== "access_token") {
  fail("clientAuthentication.tokenStorageKey must be access_token");
}
if (api.clientAuthentication?.oneTimeCodePersistence !== "forbidden") {
  fail("wx.login one-time code persistence must be forbidden");
}
if (api.clientAuthentication?.replayAfter40101?.upload !== "never") {
  fail("image upload must never be replayed automatically");
}
if (api.clientAuthentication?.replayAfter40101?.wxLoginCodeExchange !== "never") {
  fail("wx login code exchange must never be replayed automatically");
}

const lifecycle = api.mealLifecycle || {};
if (JSON.stringify(lifecycle.activeStatuses) !== JSON.stringify(["collecting", "closed", "confirmed"])) {
  fail("mealLifecycle.activeStatuses must keep confirmed meals active");
}
if (JSON.stringify(lifecycle.terminalStatuses) !== JSON.stringify(["completed", "cancelled"])) {
  fail("mealLifecycle.terminalStatuses must be completed and cancelled");
}
if (!String(lifecycle.creationLock || "").includes("collecting、closed 或 confirmed")) {
  fail("mealLifecycle must freeze the one-active-meal creation lock");
}
if (!String(lifecycle.semantics || "").includes("关闭点餐与结束饭局是两个动作")) {
  fail("mealLifecycle must distinguish closing orders from ending the meal");
}
if (!String(lifecycle.deadlineEnforcement?.serviceGuard || "").includes("截止守卫")) {
  fail("mealLifecycle must freeze the service-layer deadline guard");
}
if (!interfaceIds.has("meal_final_result_subscription")) {
  fail("mini-program contract must include the one-time meal final-result subscription endpoint");
}
if (!(api.enums?.FeatureBillingStatus || []).includes("refund_pending")) {
  fail("FeatureBillingStatus must include refund_pending");
}
if (api.commonErrorCodes?.["47006"] !== "处理失败，积分退还处理中") {
  fail("47006 must expose the frozen refund-pending user message");
}
if (api.models?.RuntimeConfig?.mealFinalResultSubscription !== "MealFinalResultSubscriptionConfig") {
  fail("RuntimeConfig must expose the meal final-result subscription template configuration");
}

const userDetail = api.models?.Profile || {};
for (const forbidden of ["openid", "openId", "unionid", "unionId", "sessionKey", "session_key"]) {
  if (Object.prototype.hasOwnProperty.call(userDetail, forbidden)) {
    fail(`Profile must not expose ${forbidden}`);
  }
}
if (api.models?.MealSummary?.coverUrl !== "string|null") {
  fail("MealSummary must expose the frozen history coverUrl field");
}
if (api.models?.Meal?.finalDishes !== "MealFinalDishSnapshot[]") {
  fail("Meal must expose final dish snapshots for history detail");
}

if (failures.length) {
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log(JSON.stringify({
  file,
  interfaces: api.interfaces.length,
  modules: new Set(api.interfaces.map((item) => item.module)).size,
  models: Object.keys(api.models).length,
  pages: knownPages.size - 1,
}, null, 2));
