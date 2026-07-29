const fs = require("fs");
const path = require("path");

const root = path.resolve(__dirname, "..");
const remoteAssetPaths = require("../config/remote-assets");
const remoteAssets = new Set(remoteAssetPaths);
const failures = [];

function fail(message) {
  failures.push(message);
}

function walk(directory, predicate, result = []) {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    if ([".preview", "design"].includes(entry.name)) continue;
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) walk(fullPath, predicate, result);
    else if (predicate(fullPath)) result.push(fullPath);
  }
  return result;
}

function parseJson(file) {
  try {
    return JSON.parse(fs.readFileSync(file, "utf8"));
  } catch (error) {
    fail(`Invalid JSON ${path.relative(root, file)}: ${error.message}`);
    return null;
  }
}

function scanTags(source, file) {
  const stack = [];
  let index = 0;
  while (index < source.length) {
    const start = source.indexOf("<", index);
    if (start < 0) break;
    if (source.startsWith("<!--", start)) {
      const end = source.indexOf("-->", start + 4);
      index = end < 0 ? source.length : end + 3;
      continue;
    }
    let quote = "";
    let end = start + 1;
    for (; end < source.length; end += 1) {
      const char = source[end];
      if (quote) {
        if (char === quote) quote = "";
      } else if (char === '"' || char === "'") {
        quote = char;
      } else if (char === ">") {
        break;
      }
    }
    if (end >= source.length) {
      fail(`Unclosed tag in ${path.relative(root, file)}`);
      break;
    }
    const raw = source.slice(start + 1, end).trim();
    index = end + 1;
    if (!raw || raw.startsWith("!") || raw.startsWith("?")) continue;
    const closing = raw.startsWith("/");
    const selfClosing = raw.endsWith("/");
    const name = raw.replace(/^\//, "").split(/\s/)[0].replace(/\/$/, "");
    if (closing) {
      const expected = stack.pop();
      if (expected !== name) fail(`Tag mismatch in ${path.relative(root, file)}: expected </${expected}> but found </${name}>`);
    } else if (!selfClosing) {
      stack.push(name);
    }
  }
  if (stack.length) fail(`Unclosed tags in ${path.relative(root, file)}: ${stack.join(", ")}`);
}

const app = parseJson(path.join(root, "app.json"));
const pages = app ? app.pages : [];

for (const route of pages) {
  for (const extension of [".js", ".json", ".wxml", ".wxss"]) {
    const file = path.join(root, `${route}${extension}`);
    if (!fs.existsSync(file)) fail(`Missing page file: ${route}${extension}`);
  }
}

for (const file of walk(root, (name) => name.endsWith(".json"))) {
  parseJson(file);
}

for (const file of walk(root, (name) => name.endsWith(".js") || name.endsWith(".cjs") || name.endsWith(".wxs"))) {
  const source = fs.readFileSync(file, "utf8");
  try {
    new Function(source);
  } catch (error) {
    fail(`JavaScript syntax error in ${path.relative(root, file)}: ${error.message}`);
  }
}

for (const file of walk(root, (name) => name.endsWith(".wxml"))) {
  const source = fs.readFileSync(file, "utf8");
  scanTags(source, file);
  if (/\{\{[^}]*\b(?!array\.)[a-zA-Z_$][\w$]*\.indexOf\s*\(/.test(source)) fail(`Unsupported direct method call in ${path.relative(root, file)}`);
  const tags = source.match(/<[^>]+>/g) || [];
  for (const tag of tags) {
    if (tag.includes("wx:for=") && tag.includes("wx:if=")) {
      fail(`wx:for and wx:if share one element in ${path.relative(root, file)}`);
    }
  }

  const jsFile = file.replace(/\.wxml$/, ".js");
  if (fs.existsSync(jsFile)) {
    const js = fs.readFileSync(jsFile, "utf8");
    const eventPattern = /(?:bind|catch)(?::)?[a-zA-Z]+="([a-zA-Z_$][\w$]*)"/g;
    let match;
    while ((match = eventPattern.exec(source))) {
      const handler = match[1];
      if (!new RegExp(`\\b${handler}\\b`).test(js)) {
        fail(`Missing event handler ${handler} in ${path.relative(root, jsFile)}`);
      }
    }
  }

  const assetPattern = /(?:src|image)="(\/assets\/[^"]+)"/g;
  let asset;
  while ((asset = assetPattern.exec(source))) {
    const target = path.join(root, asset[1].slice(1));
    const isCloudImage = asset[1].startsWith("/assets/images/");
    if (!isCloudImage && !remoteAssets.has(asset[1]) && !fs.existsSync(target)) {
      fail(`Missing local or remote asset ${asset[1]} referenced by ${path.relative(root, file)}`);
    }
  }
}

const api = parseJson(path.resolve(root, "..", "aiDoc", "miniApp", "api.json"));
if (api) {
  if (!Array.isArray(api.interfaces) || api.interfaces.length < 50) fail("api.json must contain the complete interface inventory");
  const ids = new Set();
  const signatures = new Set();
  for (const item of api.interfaces || []) {
    if (ids.has(item.id)) fail(`Duplicate API id: ${item.id}`);
    ids.add(item.id);
    const signature = `${item.method} ${item.path}`;
    if (signatures.has(signature)) fail(`Duplicate API signature: ${signature}`);
    signatures.add(signature);
    for (const route of item.usedBy || []) {
      if (route !== "app.js" && !pages.includes(route)) fail(`API ${item.id} references unknown page ${route}`);
    }
  }
  if (api.clientAuthentication?.tokenStorageKey !== "access_token") {
    fail("api.json must freeze the mini-program access token storage key");
  }
  if (api.clientAuthentication?.replayAfter40101?.upload !== "never") {
    fail("api.json must forbid automatic image upload replay");
  }
}

const appSource = fs.readFileSync(path.join(root, "app.js"), "utf8");
const authSource = fs.readFileSync(path.join(root, "services", "auth.js"), "utf8");
const requestSource = fs.readFileSync(path.join(root, "services", "request.js"), "utf8");
const uploadSource = fs.readFileSync(path.join(root, "services", "upload.js"), "utf8");
const apiClientSource = fs.readFileSync(path.join(root, "services", "api.js"), "utf8");
const mealInviteSource = fs.readFileSync(path.join(root, "pages", "meal", "invite.js"), "utf8");
const shoppingListSource = fs.readFileSync(path.join(root, "pages", "shopping", "list.js"), "utf8");
if (!appSource.includes("auth.startAuthentication()")) {
  fail("App.onLaunch must start the shared authentication Promise");
}
if (!authSource.includes("wx.login({") || !authSource.includes("/auth/wx-login")) {
  fail("auth service must exchange wx.login code through /auth/wx-login");
}
if (!requestSource.includes("auth: true") || !requestSource.includes("auth.reauthenticate()")) {
  fail("request service must wait for authentication and handle 40101 reauthentication");
}
if (!requestSource.includes("canReplayAfterAuthentication")) {
  fail("request service must enforce the frozen safe replay policy");
}
if (!uploadSource.includes("error.retryRequired = true")) {
  fail("upload service must require explicit user retry after reauthentication");
}
if (!/getSharedShoppingList:[\s\S]*?auth:\s*false/.test(apiClientSource)) {
  fail("public shopping share request must explicitly use auth:false");
}
if (!fs.existsSync(path.join(root, "tools", "test-auth-flow.cjs"))) {
  fail("authentication flow test is required");
}
if (!mealInviteSource.includes("api.cancelMeal(") || !mealInviteSource.includes('"collecting", "closed"')) {
  fail("meal invite page must support creator cancellation before menu confirmation");
}
if (!shoppingListSource.includes("api.completeMeal(") || !shoppingListSource.includes('meal.status !== "confirmed"')) {
  fail("shopping list page must support creator-completed meal closure from confirmed");
}

if (remoteAssets.size !== remoteAssetPaths.length) fail("remote-assets.js contains duplicate paths");
for (const source of remoteAssetPaths) {
  if (!/^\/assets\/(?:images|icons)\//.test(source)) fail(`Invalid remote asset path: ${source}`);
}

const iconDir = path.join(root, "assets", "icons");
const iconNames = new Set(
  remoteAssetPaths
    .filter((source) => source.startsWith("/assets/icons/"))
    .map((source) => path.basename(source))
);
if (fs.existsSync(iconDir)) {
  fs.readdirSync(iconDir).filter((name) => /\.(png|jpe?g|webp|gif|bmp|avif)$/i.test(name)).forEach((name) => iconNames.add(name));
}
if (iconNames.size < 30) fail("Configured icon assets are incomplete");

if (failures.length) {
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log(JSON.stringify({
  pages: pages.length,
  jsonFiles: walk(root, (name) => name.endsWith(".json")).length,
  jsFiles: walk(root, (name) => name.endsWith(".js") || name.endsWith(".cjs") || name.endsWith(".wxs")).length,
  wxmlFiles: walk(root, (name) => name.endsWith(".wxml")).length,
  apiInterfaces: api.interfaces.length,
  icons: iconNames.size
}, null, 2));
