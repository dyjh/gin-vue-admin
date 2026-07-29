const env = require("../config/env");
const remoteAssetPaths = require("../config/remote-assets");

const LOCAL_ASSET_PREFIX = "/assets/";
const REMOTE_IMAGE_PREFIX = "/assets/images/";
const UPLOAD_IMAGE_PREFIX = "uploads/";
const remoteAssets = new Set(remoteAssetPaths);

function getAssetBaseUrl() {
  return String(env.assetBaseUrl || "").replace(/\/+$/, "");
}

// getImageBaseUrl 获取用户上传图片的公网访问地址。
function getImageBaseUrl() {
  return String(env.imageBaseUrl || "").replace(/\/+$/, "");
}

function isRemoteAsset(source) {
  return typeof source === "string"
    && (source.startsWith(REMOTE_IMAGE_PREFIX) || remoteAssets.has(source));
}

function resolveAssetUrl(source) {
  if (typeof source !== "string") return source;
  // 将接口返回的历史相对上传地址转换为小程序可访问的完整 HTTPS 地址。
  const normalizedSource = source.startsWith("/") ? source.slice(1) : source;
  if (normalizedSource.startsWith(UPLOAD_IMAGE_PREFIX)) {
    const imageBaseUrl = getImageBaseUrl();
    return imageBaseUrl ? `${imageBaseUrl}/${normalizedSource}` : source;
  }
  if (!source.startsWith(LOCAL_ASSET_PREFIX)) return source;
  if (!isRemoteAsset(source)) return source;
  const baseUrl = getAssetBaseUrl();
  return baseUrl ? `${baseUrl}${source}` : source;
}

function resolveAssetTree(value, seen = new WeakMap()) {
  if (typeof value === "string") return resolveAssetUrl(value);
  if (!value || typeof value !== "object") return value;
  if (seen.has(value)) return seen.get(value);

  const output = Array.isArray(value) ? [] : {};
  seen.set(value, output);
  Object.keys(value).forEach((key) => {
    output[key] = resolveAssetTree(value[key], seen);
  });
  return output;
}

module.exports = {
  getAssetBaseUrl,
  getImageBaseUrl,
  isRemoteAsset,
  resolveAssetTree,
  resolveAssetUrl,
};
