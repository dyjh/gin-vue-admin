const { ensureSeeded } = require("./mock/store");

App({
  globalData: {
    nav: {
      statusBarHeight: 20,
      menuTop: 24,
      menuHeight: 32,
      navHeight: 88,
    },
  },

  onLaunch() {
    const windowInfo = wx.getWindowInfo ? wx.getWindowInfo() : wx.getSystemInfoSync();
    const menu = wx.getMenuButtonBoundingClientRect
      ? wx.getMenuButtonBoundingClientRect()
      : { top: windowInfo.statusBarHeight + 4, height: 32, bottom: windowInfo.statusBarHeight + 36 };

    this.globalData.nav = {
      statusBarHeight: windowInfo.statusBarHeight || 20,
      menuTop: menu.top,
      menuHeight: menu.height,
      navHeight: menu.bottom + Math.max(8, menu.top - (windowInfo.statusBarHeight || 20)),
    };
    ensureSeeded();
  },
});
