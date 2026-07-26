const assert = require("assert");
const env = require("../config/env");
const remoteAssetPaths = require("../config/remote-assets");
const {
  getAssetBaseUrl,
  getImageBaseUrl,
  isRemoteAsset,
  resolveAssetTree,
  resolveAssetUrl,
} = require("../utils/assets");

const originalBaseUrl = env.assetBaseUrl;
const originalImageBaseUrl = env.imageBaseUrl;

try {
  env.assetBaseUrl = "https://cache.ljdyjh.cn/";
  env.imageBaseUrl = "https://order-food.ljdyjh.cn/";
  assert.strictEqual(getAssetBaseUrl(), "https://cache.ljdyjh.cn");
  assert.strictEqual(getImageBaseUrl(), "https://order-food.ljdyjh.cn");
  assert.strictEqual(remoteAssetPaths.length, new Set(remoteAssetPaths).size);
  assert(remoteAssetPaths.every((source) => /^\/assets\/(?:images|icons)\//.test(source)));

  const remoteImage = "/assets/images/home-approved-header-v1.jpg";
  const remoteIcon = "/assets/icons/home-green.png";
  const localOnly = "/assets/images/new-local-only.jpg";
  const uploadedImage = "/uploads/file/profile-avatar.jpg";
  const legacyUploadedImage = "uploads/file/legacy-profile-avatar.jpg";

  assert.strictEqual(isRemoteAsset(remoteImage), true);
  assert.strictEqual(isRemoteAsset(remoteIcon), true);
  assert.strictEqual(isRemoteAsset(localOnly), false);
  assert.strictEqual(resolveAssetUrl(remoteImage), "https://cache.ljdyjh.cn" + remoteImage);
  assert.strictEqual(resolveAssetUrl(localOnly), localOnly);
  assert.strictEqual(
    resolveAssetUrl(uploadedImage),
    "https://order-food.ljdyjh.cn" + uploadedImage
  );
  assert.strictEqual(
    resolveAssetUrl(legacyUploadedImage),
    "https://order-food.ljdyjh.cn/" + legacyUploadedImage
  );
  assert.strictEqual(resolveAssetUrl("https://example.com/image.jpg"), "https://example.com/image.jpg");

  const { SHARE_IMAGES } = require("../config/share");
  assert.strictEqual(
    SHARE_IMAGES.default,
    "https://cache.ljdyjh.cn/assets/images/share-global-cooking-v1.jpg"
  );
  assert.strictEqual(
    SHARE_IMAGES.shoppingList,
    "https://cache.ljdyjh.cn/assets/images/share-shopping-list-v1.jpg"
  );

  const mixedTree = resolveAssetTree({
    image: remoteImage,
    local: localOnly,
    nested: [{ icon: remoteIcon, uploadedImage }],
  });
  assert.strictEqual(mixedTree.image, "https://cache.ljdyjh.cn" + remoteImage);
  assert.strictEqual(mixedTree.local, localOnly);
  assert.strictEqual(mixedTree.nested[0].icon, "https://cache.ljdyjh.cn" + remoteIcon);
  assert.strictEqual(
    mixedTree.nested[0].uploadedImage,
    "https://order-food.ljdyjh.cn" + uploadedImage
  );

  console.log(JSON.stringify({ remoteAssets: remoteAssetPaths.length, remote: true, localOnly: true }));
} finally {
  env.assetBaseUrl = originalBaseUrl;
  env.imageBaseUrl = originalImageBaseUrl;
}
