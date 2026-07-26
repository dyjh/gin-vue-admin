const env = require("../config/env");
const { resolveAssetTree } = require("../utils/assets");
const auth = require("./auth");
const {
  canReplayAfterAuthentication,
  createIdempotencyKey,
  createNetworkError,
  createResponseError,
  isAuthExpiredError,
  isMutationMethod,
  showErrorToast,
} = require("./request-helpers");

function normalizeConfig(options = {}) {
  const method = String(options.method || "GET").toUpperCase();
  const needsIdempotency = isMutationMethod(method) && options.idempotency !== false;
  return {
    method,
    data: {},
    showError: true,
    auth: true,
    autoReplay: true,
    headers: {},
    ...options,
    method,
    idempotencyKey: needsIdempotency
      ? options.idempotencyKey || createIdempotencyKey()
      : options.idempotencyKey || "",
  };
}

function buildHeaders(config) {
  const headers = {
    "content-type": "application/json",
    ...config.headers,
  };
  if (config.auth) {
    const token = auth.getAccessToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  } else {
    delete headers.Authorization;
    delete headers.authorization;
  }
  if (config.idempotencyKey) {
    headers["X-Idempotency-Key"] = config.idempotencyKey;
  }
  return headers;
}

function sendWxRequest(config) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${env.baseUrl}${config.url}`,
      method: config.method,
      data: config.data,
      timeout: env.timeout,
      header: buildHeaders(config),
      success(response) {
        const body = response.data || {};
        if (response.statusCode >= 200 && response.statusCode < 300 && body.code === 0) {
          resolve(resolveAssetTree(body.data));
          return;
        }
        reject(createResponseError(response, "请求失败，请稍后再试"));
      },
      fail(originalError) {
        reject(createNetworkError(originalError, "网络暂时不可用"));
      },
    });
  });
}

async function sendOnce(config) {
  return sendWxRequest(config);
}

async function execute(config) {
  if (config.auth) await auth.ensureAuthenticated();

  try {
    return await sendOnce(config);
  } catch (error) {
    if (
      config.auth &&
      config.autoReplay &&
      !config.authReplayAttempted &&
      isAuthExpiredError(error)
    ) {
      await auth.reauthenticate();
      if (canReplayAfterAuthentication(config.method, config.idempotencyKey)) {
        return sendOnce({ ...config, authReplayAttempted: true });
      }
      error.reauthenticated = true;
      error.retryRequired = true;
      error.message = "登录已恢复，请重新提交本次操作";
    }
    throw error;
  }
}

async function request(options) {
  const config = normalizeConfig(options);
  try {
    return await execute(config);
  } catch (error) {
    if (config.showError) showErrorToast(error, "请求失败，请稍后再试");
    throw error;
  }
}

request.createIdempotencyKey = createIdempotencyKey;
request.canReplayAfterAuthentication = canReplayAfterAuthentication;

module.exports = request;
