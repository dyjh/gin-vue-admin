const env = require("../config/env");
const remoteAssetPaths = require("../config/remote-assets");

const LOCAL_ASSET_PREFIX = "/assets/";
const remoteAssets = new Set(remoteAssetPaths);

function getAssetBaseUrl() {
  return String(env.assetBaseUrl || "").replace(/\/+$/, "");
}

function isRemoteAsset(source) {
  return typeof source === "string" && remoteAssets.has(source);
}

function resolveAssetUrl(source) {
  if (typeof source !== "string" || !source.startsWith(LOCAL_ASSET_PREFIX)) return source;
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
  isRemoteAsset,
  resolveAssetTree,
  resolveAssetUrl,
};
