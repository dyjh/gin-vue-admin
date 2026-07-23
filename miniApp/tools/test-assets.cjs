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

  const remoteImage = "/assets/images/ai-kung-pao-chicken-step.jpg";
  const remoteIcon = "/assets/icons/home-green.png";
  const localDraft = "/assets/images/new-local-draft.jpg";

  assert.strictEqual(isRemoteAsset(remoteImage), true);
  assert.strictEqual(isRemoteAsset(remoteIcon), true);
  assert.strictEqual(isRemoteAsset(localDraft), false);
  assert.strictEqual(resolveAssetUrl(remoteImage), "https://cache.ljdyjh.cn" + remoteImage);
  assert.strictEqual(resolveAssetUrl(localDraft), localDraft);
  assert.strictEqual(resolveAssetUrl("https://example.com/image.jpg"), "https://example.com/image.jpg");

  const mixedTree = resolveAssetTree({
    image: remoteImage,
    draft: localDraft,
    nested: [{ icon: remoteIcon }],
  });
  assert.strictEqual(mixedTree.image, "https://cache.ljdyjh.cn" + remoteImage);
  assert.strictEqual(mixedTree.draft, localDraft);
  assert.strictEqual(mixedTree.nested[0].icon, "https://cache.ljdyjh.cn" + remoteIcon);

  console.log(JSON.stringify({ remoteAssets: remoteAssetPaths.length, remote: true, localDraft: true }));
} finally {
  env.assetBaseUrl = originalBaseUrl;
}
