const assert = require("assert");

const storage = new Map();
const requestLog = [];
const uploadLog = [];
let loginCount = 0;
let loginExchangeCount = 0;
let issuedTokenCount = 0;
let appDefinition;
let expiredToken = "";
let uploadShouldExpire = false;
let runtimeResponseVersion = 10;

function respond(callback) {
  setTimeout(callback, 0);
}

global.Page = function registerPage() {};
global.App = function registerApp(options) {
  appDefinition = options;
};
global.getApp = function getRegisteredApp() {
  return appDefinition;
};
global.wx = {
  getStorageSync(key) {
    return storage.get(key);
  },
  setStorageSync(key, value) {
    storage.set(key, value);
  },
  removeStorageSync(key) {
    storage.delete(key);
  },
  getWindowInfo() {
    return { statusBarHeight: 20 };
  },
  getMenuButtonBoundingClientRect() {
    return { top: 24, height: 32, bottom: 56 };
  },
  showToast() {},
  login(options) {
    loginCount += 1;
    const code = `one-time-code-${loginCount}`;
    respond(() => options.success({ code }));
  },
  request(options) {
    requestLog.push({
      url: options.url,
      method: options.method,
      data: options.data,
      header: { ...(options.header || {}) },
    });
    if (options.url.endsWith("/auth/wx-login")) {
      loginExchangeCount += 1;
      issuedTokenCount += 1;
      const token = `token-${issuedTokenCount}`;
      respond(() => options.success({
        statusCode: 200,
        header: { "X-Request-Id": `auth-${issuedTokenCount}` },
        data: {
          code: 0,
          data: {
            accessToken: token,
            expiresIn: 3600,
            isNewUser: issuedTokenCount === 1,
            profile: { id: "user-1", nickname: "认证测试" },
            runtimeConfig: {
              policyVersion: issuedTokenCount,
              enhancedFeaturesEnabled: false,
              pointsEnabled: false,
              features: [],
            },
          },
          msg: "ok",
        },
      }));
      return;
    }

    if (options.url.endsWith("/runtime-config")) {
      respond(() => options.success({
        statusCode: 200,
        header: { "X-Request-Id": "runtime-config" },
        data: {
          code: 0,
          data: {
            policyVersion: runtimeResponseVersion,
            enhancedFeaturesEnabled: true,
            pointsEnabled: true,
            features: [],
          },
          msg: "ok",
        },
      }));
      return;
    }

    if (options.url.endsWith("/bootstrap")) {
      respond(() => options.success({
        statusCode: 200,
        header: { "X-Request-Id": "bootstrap" },
        data: {
          code: 0,
          data: {
            profile: { id: "user-1", nickname: "认证测试" },
            runtimeConfig: {
              policyVersion: runtimeResponseVersion,
              enhancedFeaturesEnabled: true,
              pointsEnabled: true,
              features: [],
            },
            dishes: [],
            recommendations: [],
            activeMeal: null,
            unreadCount: 0,
          },
          msg: "ok",
        },
      }));
      return;
    }

    const authorization = options.header && (options.header.Authorization || options.header.authorization);
    if (authorization === `Bearer ${expiredToken}`) {
      respond(() => options.success({
        statusCode: 401,
        header: { "X-Request-Id": "expired-request" },
        data: { code: 40101, data: null, msg: "登录已过期" },
      }));
      return;
    }

    respond(() => options.success({
      statusCode: 200,
      header: { "X-Request-Id": "business-request" },
      data: {
        code: 0,
        data: {
          path: options.url,
          authorization: authorization || "",
        },
        msg: "ok",
      },
    }));
  },
  uploadFile(options) {
    uploadLog.push({
      url: options.url,
      filePath: options.filePath,
      header: { ...(options.header || {}) },
    });
    if (uploadShouldExpire) {
      uploadShouldExpire = false;
      respond(() => options.success({
        statusCode: 401,
        header: { "X-Request-Id": "upload-expired" },
        data: JSON.stringify({ code: 40101, data: null, msg: "登录已过期" }),
      }));
      return;
    }
    respond(() => options.success({
      statusCode: 200,
      header: { "X-Request-Id": "upload-success" },
      data: JSON.stringify({
        code: 0,
        data: { fileId: "file-1", url: "https://example.test/file-1.jpg", reviewStatus: "passed" },
        msg: "ok",
      }),
    }));
  },
};

const env = require("../config/env");
env.baseUrl = "https://api.test/api/miniapp/v1";

const auth = require("../services/auth");
const api = require("../services/api");
const request = require("../services/request");
const { uploadImage } = require("../services/upload");

async function main() {
  require("../app");
  assert.ok(appDefinition, "App definition should be registered");

  appDefinition.onLaunch.call(appDefinition);
  const coldStartRequests = await Promise.all([
    request({ url: "/concurrent-a" }),
    request({ url: "/concurrent-b" }),
  ]);
  await appDefinition.globalData.authPromise;

  assert.strictEqual(loginCount, 1, "cold start and concurrent page requests must share one wx.login");
  assert.strictEqual(loginExchangeCount, 1, "the one-time code must be exchanged once");
  assert.ok(coldStartRequests.every((item) => item.authorization === "Bearer token-1"));
  assert.strictEqual(appDefinition.globalData.runtimeConfig.policyVersion, 1);

  await api.getRuntimeConfig();
  assert.strictEqual(
    appDefinition.globalData.runtimeConfig.policyVersion,
    runtimeResponseVersion,
    "runtime endpoint must replace the in-memory configuration"
  );
  runtimeResponseVersion += 1;
  await api.bootstrap();
  assert.strictEqual(
    auth.getRuntimeConfig().policyVersion,
    runtimeResponseVersion,
    "bootstrap must replace the persisted configuration"
  );

  expiredToken = "token-1";
  const reloginResults = await Promise.all([
    request({ url: "/reauth-a" }),
    request({ url: "/reauth-b" }),
  ]);
  assert.strictEqual(loginCount, 2, "concurrent 40101 responses must share one relogin");
  assert.strictEqual(loginExchangeCount, 2);
  assert.ok(reloginResults.every((item) => item.authorization === "Bearer token-2"));

  expiredToken = "token-2";
  const postLogStart = requestLog.length;
  await request({ url: "/post-retry", method: "POST", data: { value: 1 } });
  const postAttempts = requestLog.slice(postLogStart).filter((item) => item.url.endsWith("/post-retry"));
  assert.strictEqual(postAttempts.length, 2, "idempotent POST should replay once after relogin");
  assert.ok(postAttempts[0].header["X-Idempotency-Key"]);
  assert.strictEqual(
    postAttempts[0].header["X-Idempotency-Key"],
    postAttempts[1].header["X-Idempotency-Key"],
    "replayed mutation must reuse the original idempotency key"
  );
  assert.strictEqual(loginCount, 3);

  auth.clearSession();
  expiredToken = "";
  const loginCountBeforePublicRequest = loginCount;
  const publicResult = await request({ url: "/public/resource", auth: false });
  assert.strictEqual(loginCount, loginCountBeforePublicRequest, "public request must not wait for or trigger login");
  assert.strictEqual(publicResult.authorization, "");

  await auth.startAuthentication();
  uploadShouldExpire = true;
  const uploadCountBefore = uploadLog.length;
  await assert.rejects(
    () => uploadImage("/tmp/auth-test.jpg", "dish_cover"),
    (error) => error.retryRequired === true && error.reauthenticated === true
  );
  assert.strictEqual(
    uploadLog.length - uploadCountBefore,
    1,
    "image upload must never be replayed automatically"
  );

  const storedText = JSON.stringify(Array.from(storage.entries()));
  assert.strictEqual(storedText.includes("one-time-code-"), false, "wx.login code must never be persisted");

  console.log(JSON.stringify({
    coldStartWxLoginCount: 1,
    concurrentReloginSingleFlight: true,
    mutationIdempotencyKeyReused: true,
    publicRequestSkippedAuthentication: true,
    runtimeConfigRefreshApplied: true,
    uploadAutoReplayCount: 0,
    totalWxLoginCount: loginCount,
  }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
