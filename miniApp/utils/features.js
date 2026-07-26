const auth = require("../services/auth");

function runtimeConfig() {
  const app = typeof getApp === "function" ? getApp() : null;
  return (app && app.globalData && app.globalData.runtimeConfig)
    || auth.getRuntimeConfig()
    || null;
}

function getFeature(code) {
  const config = runtimeConfig();
  if (!config || !config.enhancedFeaturesEnabled || !Array.isArray(config.features)) return null;
  return config.features.find((item) => item.code === code && item.enabled) || null;
}

function pointsEnabled() {
  const config = runtimeConfig();
  return Boolean(config && config.enhancedFeaturesEnabled && config.pointsEnabled);
}

module.exports = {
  getFeature,
  pointsEnabled,
  runtimeConfig,
};
