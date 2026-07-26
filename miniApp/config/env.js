const publicDomains = require("./domains");

module.exports = {
  baseUrl: `${publicDomains.api}/api/miniapp/v1`,
  assetBaseUrl: publicDomains.assets,
  imageBaseUrl: publicDomains.imageUrl,
  publicDomains,
  timeout: 12000,
};
