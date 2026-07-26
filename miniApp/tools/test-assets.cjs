const assert = require("assert");
const env = require("../config/env");
const remoteAssetPaths = require("../config/remote-assets");
const { getAssetBaseUrl, isRemoteAsset, resolveAssetTree, resolveAssetUrl } = require("../utils/assets");

const originalBaseUrl = env.assetBaseUrl;

try {
  env.assetBaseUrl = "https://cache.ljdyjh.cn/";
  assert.strictEqual(getAssetBaseUrl(), "https://cache.ljdyjh.cn");
  assert.strictEqual(remoteAssetPaths.length, new Set(remoteAssetPaths).size);
  assert(remoteAssetPaths.every((source) => /^\/assets\/(?:images|icons)\//.test(source)));

  const remoteImage = "/assets/images/home-approved-header-v1.jpg";
  const remoteIcon = "/assets/icons/home-green.png";
  const localOnly = "/assets/images/new-local-only.jpg";

  assert.strictEqual(isRemoteAsset(remoteImage), true);
  assert.strictEqual(isRemoteAsset(remoteIcon), true);
  assert.strictEqual(isRemoteAsset(localOnly), false);
  assert.strictEqual(resolveAssetUrl(remoteImage), "https://cache.ljdyjh.cn" + remoteImage);
  assert.strictEqual(resolveAssetUrl(localOnly), localOnly);
  assert.strictEqual(resolveAssetUrl("https://example.com/image.jpg"), "https://example.com/image.jpg");

  const mixedTree = resolveAssetTree({
    image: remoteImage,
    local: localOnly,
    nested: [{ icon: remoteIcon }],
  });
  assert.strictEqual(mixedTree.image, "https://cache.ljdyjh.cn" + remoteImage);
  assert.strictEqual(mixedTree.local, localOnly);
  assert.strictEqual(mixedTree.nested[0].icon, "https://cache.ljdyjh.cn" + remoteIcon);

  console.log(JSON.stringify({ remoteAssets: remoteAssetPaths.length, remote: true, localOnly: true }));
} finally {
  env.assetBaseUrl = originalBaseUrl;
}
