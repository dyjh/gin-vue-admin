const publicDomains = require("./domains");

module.exports = {
  useMock: true,
  baseUrl: `${publicDomains.api}/api/miniapp/v1`,
  assetBaseUrl: publicDomains.assets,
  publicDomains,
  timeout: 12000,
};
