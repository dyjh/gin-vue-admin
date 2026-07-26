let idempotencySequence = 0;
let lastToastMessage = "";
let lastToastAt = 0;

function createIdempotencyKey(prefix = "miniapp") {
  idempotencySequence = (idempotencySequence + 1) % 1679616;
  const timestamp = Date.now().toString(36);
  const sequence = idempotencySequence.toString(36).padStart(4, "0");
  const random = Math.random().toString(36).slice(2, 12);
  return `${prefix}-${timestamp}-${sequence}-${random}`;
}

function findHeader(headers, name) {
  if (!headers || typeof headers !== "object") return "";
  const target = name.toLowerCase();
  const key = Object.keys(headers).find((item) => item.toLowerCase() === target);
  return key ? headers[key] : "";
}

function createResponseError(response, fallbackMessage) {
  const body = response && response.data && typeof response.data === "object"
    ? response.data
    : {};
  const error = new Error(body.msg || fallbackMessage);
  error.code = body.code;
  error.data = body.data;
  error.httpStatus = response ? response.statusCode : 0;
  error.requestId = findHeader(response && response.header, "x-request-id") || "";
  return error;
}

function createNetworkError(originalError, fallbackMessage) {
  const error = new Error(fallbackMessage);
  error.code = "NETWORK_ERROR";
  error.isNetworkError = true;
  error.cause = originalError;
  return error;
}

function showErrorToast(error, fallbackMessage) {
  if (typeof wx === "undefined" || typeof wx.showToast !== "function") return;
  const message = (error && error.message) || fallbackMessage;
  const now = Date.now();
  if (message === lastToastMessage && now - lastToastAt < 1500) return;
  lastToastMessage = message;
  lastToastAt = now;
  wx.showToast({ title: message, icon: "none" });
}

function isAuthExpiredError(error) {
  return Number(error && error.code) === 40101;
}

function isMutationMethod(method) {
  return !["GET", "HEAD", "OPTIONS"].includes(String(method || "GET").toUpperCase());
}

function canReplayAfterAuthentication(method, idempotencyKey) {
  const normalizedMethod = String(method || "GET").toUpperCase();
  if (["GET", "HEAD"].includes(normalizedMethod)) return true;
  if (["PUT", "DELETE", "POST", "PATCH"].includes(normalizedMethod)) {
    return Boolean(idempotencyKey);
  }
  return false;
}

module.exports = {
  canReplayAfterAuthentication,
  createIdempotencyKey,
  createNetworkError,
  createResponseError,
  isAuthExpiredError,
  isMutationMethod,
  showErrorToast,
};
