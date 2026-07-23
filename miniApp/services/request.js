const env = require("../config/env");
const mock = require("../mock/router");
const { resolveAssetTree } = require("../utils/assets");

function request(options) {
  const config = {
    method: "GET",
    data: {},
    showError: true,
    ...options,
  };

  if (env.useMock) {
    return mock.handle(config).then(resolveAssetTree);
  }

  return new Promise((resolve, reject) => {
    wx.request({
      url: `${env.baseUrl}${config.url}`,
      method: config.method,
      data: config.data,
      timeout: env.timeout,
      header: {
        "content-type": "application/json",
        authorization: wx.getStorageSync("access_token") ? `Bearer ${wx.getStorageSync("access_token")}` : "",
      },
      success(response) {
        const body = response.data || {};
        if (response.statusCode >= 200 && response.statusCode < 300 && body.code === 0) {
          resolve(resolveAssetTree(body.data));
          return;
        }
        const error = new Error(body.msg || "请求失败，请稍后再试");
        error.code = body.code;
        if (config.showError) wx.showToast({ title: error.message, icon: "none" });
        reject(error);
      },
      fail(error) {
        if (config.showError) wx.showToast({ title: "网络暂时不可用", icon: "none" });
        reject(error);
      },
    });
  });
}

module.exports = request;
