const { resolveAssetUrl } = require("../utils/assets");

const SHARE_IMAGES = Object.freeze({
  default: resolveAssetUrl("/assets/images/share-global-cooking-v1.jpg"),
  shoppingList: resolveAssetUrl("/assets/images/share-shopping-list-v1.jpg"),
});

module.exports = {
  DEFAULT_SHARE_TITLE: "\u6765\u5e72\u996d",
  SHARE_IMAGES,
};
