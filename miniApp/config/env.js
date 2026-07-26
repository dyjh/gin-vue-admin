const publicDomains = require("./domains");

module.exports = {
  baseUrl: `${publicDomains.api}/api/miniapp/v1`,
  assetBaseUrl: publicDomains.assets,
  publicDomains,
  timeout: 12000,
};
