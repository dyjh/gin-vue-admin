let registered = false;

function applyUpdate(updateManager) {
  if (typeof updateManager.applyUpdate === "function") {
    updateManager.applyUpdate();
  }
}

function showUpdateReady(wxApi, updateManager) {
  if (typeof wxApi.showModal !== "function") {
    applyUpdate(updateManager);
    return;
  }

  wxApi.showModal({
    title: "发现新版本",
    content: "新版本已准备好，需要重启小程序完成更新。",
    confirmText: "重启更新",
    showCancel: false,
    success(result) {
      if (result.confirm) {
        applyUpdate(updateManager);
      }
    },
    fail() {
      applyUpdate(updateManager);
    },
  });
}

function showUpdateFailed(wxApi) {
  const content = "新版本下载失败，请检查网络后关闭并重新打开小程序。";

  if (typeof wxApi.showModal === "function") {
    wxApi.showModal({
      title: "更新失败",
      content,
      confirmText: "知道了",
      showCancel: false,
    });
    return;
  }

  if (typeof wxApi.showToast === "function") {
    wxApi.showToast({
      title: content,
      icon: "none",
      duration: 3000,
    });
  }
}

function registerUpdateManager(wxApi = typeof wx === "undefined" ? null : wx) {
  if (
    registered ||
    !wxApi ||
    typeof wxApi.getUpdateManager !== "function"
  ) {
    return null;
  }

  const updateManager = wxApi.getUpdateManager();
  if (!updateManager) {
    return null;
  }

  registered = true;
  let updateReadyHandled = false;
  let updateFailedHandled = false;

  if (typeof updateManager.onUpdateReady === "function") {
    updateManager.onUpdateReady(() => {
      if (updateReadyHandled) return;
      updateReadyHandled = true;
      showUpdateReady(wxApi, updateManager);
    });
  }

  if (typeof updateManager.onUpdateFailed === "function") {
    updateManager.onUpdateFailed(() => {
      if (updateFailedHandled || updateReadyHandled) return;
      updateFailedHandled = true;
      showUpdateFailed(wxApi);
    });
  }

  return updateManager;
}

module.exports = {
  registerUpdateManager,
};
