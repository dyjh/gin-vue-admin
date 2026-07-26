const env = require("../config/env");
const { resolveAssetTree } = require("../utils/assets");
const {
  createIdempotencyKey,
  createNetworkError,
  createResponseError,
} = require("./request-helpers");

const STORAGE_KEYS = Object.freeze({
  accessToken: "access_token",
  expiresAt: "access_token_expires_at",
  profile: "auth_profile",
  runtimeConfig: "runtime_config",
});
const EXPIRY_SKEW_MS = 30000;

let activeAuthenticationPromise = null;

function getStoredValue(key) {
  try {
    return wx.getStorageSync(key);
  } catch (error) {
    return undefined;
  }
}

function setStoredValue(key, value) {
  try {
    wx.setStorageSync(key, value);
  } catch (error) {
    // Storage failure must not expose the one-time login code.
  }
}

function removeStoredValue(key) {
  try {
    if (typeof wx.removeStorageSync === "function") wx.removeStorageSync(key);
    else wx.setStorageSync(key, "");
  } catch (error) {
    // A following authenticated request still relies on the in-memory flow.
  }
}

function getAccessToken() {
  return getStoredValue(STORAGE_KEYS.accessToken) || "";
}

function getRuntimeConfig() {
  return getStoredValue(STORAGE_KEYS.runtimeConfig) || null;
}

function getProfile() {
  return getStoredValue(STORAGE_KEYS.profile) || null;
}

function updateRuntimeConfig(runtimeConfig) {
  const value = runtimeConfig || null;
  setStoredValue(STORAGE_KEYS.runtimeConfig, value);
  const app = typeof getApp === "function" ? getApp() : null;
  if (app && app.globalData) app.globalData.runtimeConfig = value;
  return value;
}

function hasUsableStoredSession() {
  const token = getAccessToken();
  const expiresAt = Number(getStoredValue(STORAGE_KEYS.expiresAt) || 0);
  return Boolean(token) && (!expiresAt || expiresAt - EXPIRY_SKEW_MS > Date.now());
}

function currentSession() {
  return {
    accessToken: getAccessToken(),
    expiresAt: Number(getStoredValue(STORAGE_KEYS.expiresAt) || 0),
    profile: getProfile(),
    runtimeConfig: getRuntimeConfig(),
  };
}

function clearAccessToken() {
  removeStoredValue(STORAGE_KEYS.accessToken);
  removeStoredValue(STORAGE_KEYS.expiresAt);
}

function clearSession() {
  activeAuthenticationPromise = null;
  Object.values(STORAGE_KEYS).forEach(removeStoredValue);
}

function persistSession(result) {
  if (!result || typeof result.accessToken !== "string" || !result.accessToken) {
    const error = new Error("登录响应缺少访问令牌");
    error.code = "INVALID_AUTH_RESPONSE";
    throw error;
  }
  const expiresIn = Math.max(0, Number(result.expiresIn) || 0);
  const expiresAt = expiresIn ? Date.now() + expiresIn * 1000 : 0;
  setStoredValue(STORAGE_KEYS.accessToken, result.accessToken);
  setStoredValue(STORAGE_KEYS.expiresAt, expiresAt);
  setStoredValue(STORAGE_KEYS.profile, result.profile || null);
  setStoredValue(STORAGE_KEYS.runtimeConfig, result.runtimeConfig || null);
  return {
    accessToken: result.accessToken,
    expiresAt,
    isNewUser: Boolean(result.isNewUser),
    profile: result.profile || null,
    runtimeConfig: result.runtimeConfig || null,
  };
}

function obtainWechatCode() {
  return new Promise((resolve, reject) => {
    if (typeof wx === "undefined" || typeof wx.login !== "function") {
      const error = new Error("当前环境无法完成微信登录");
      error.code = "WX_LOGIN_UNAVAILABLE";
      reject(error);
      return;
    }
    wx.login({
      success(result) {
        if (result && result.code) {
          resolve(result.code);
          return;
        }
        const error = new Error("微信登录失败，请稍后重试");
        error.code = "WX_LOGIN_EMPTY_CODE";
        reject(error);
      },
      fail(originalError) {
        reject(createNetworkError(originalError, "微信登录失败，请检查网络后重试"));
      },
    });
  });
}

function exchangeWechatCode(code) {
  const idempotencyKey = createIdempotencyKey("wx-login");
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${env.baseUrl}/auth/wx-login`,
      method: "POST",
      data: { code },
      timeout: env.timeout,
      header: {
        "content-type": "application/json",
        "X-Idempotency-Key": idempotencyKey,
      },
      success(response) {
        const body = response.data || {};
        if (response.statusCode >= 200 && response.statusCode < 300 && body.code === 0) {
          // 微信登录独立于通用请求封装，需要在持久化会话前补全头像等资源地址。
          resolve(resolveAssetTree(body.data || {}));
          return;
        }
        reject(createResponseError(response, "登录失败，请稍后重试"));
      },
      fail(originalError) {
        reject(createNetworkError(originalError, "登录失败，请检查网络后重试"));
      },
    });
  });
}

async function performAuthentication() {
  const code = await obtainWechatCode();
  const result = await exchangeWechatCode(code);
  return persistSession(result);
}

function beginAuthentication(options = {}) {
  if (activeAuthenticationPromise) return activeAuthenticationPromise;
  if (!options.force && hasUsableStoredSession()) return Promise.resolve(currentSession());
  if (options.force) clearAccessToken();

  activeAuthenticationPromise = performAuthentication().finally(() => {
    activeAuthenticationPromise = null;
  });
  return activeAuthenticationPromise;
}

function startAuthentication() {
  return beginAuthentication({ force: true });
}

function ensureAuthenticated() {
  return beginAuthentication({ force: false });
}

function reauthenticate() {
  return beginAuthentication({ force: true });
}

function retryAuthentication() {
  return beginAuthentication({ force: true });
}

function getAuthenticationPromise() {
  return activeAuthenticationPromise;
}

module.exports = {
  STORAGE_KEYS,
  clearSession,
  ensureAuthenticated,
  getAccessToken,
  getAuthenticationPromise,
  getProfile,
  getRuntimeConfig,
  updateRuntimeConfig,
  reauthenticate,
  retryAuthentication,
  startAuthentication,
};
