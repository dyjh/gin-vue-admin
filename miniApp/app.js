const { DEFAULT_SHARE_TITLE, SHARE_IMAGES } = require("./config/share");
const auth = require("./services/auth");

const registerPage = Page;

Page = function registerPageWithDefaultShare(options = {}) {
  const onShareAppMessage = options.onShareAppMessage;

  return registerPage({
    ...options,
    onShareAppMessage(...args) {
      const shareOptions =
        typeof onShareAppMessage === "function"
          ? onShareAppMessage.apply(this, args) || {}
          : {};

      return {
        title: DEFAULT_SHARE_TITLE,
        imageUrl: SHARE_IMAGES.default,
        ...shareOptions,
      };
    },
  });
};

App({
  globalData: {
    nav: {
      statusBarHeight: 20,
      menuTop: 24,
      menuHeight: 32,
      navHeight: 88,
    },
    authPromise: null,
    authError: null,
    profile: null,
    runtimeConfig: null,
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
    const authentication = auth.startAuthentication();
    this.globalData.authPromise = authentication;
    authentication.then(
      (session) => {
        this.globalData.authError = null;
        this.globalData.profile = session.profile;
        this.globalData.runtimeConfig = session.runtimeConfig;
      },
      (error) => {
        this.globalData.authError = {
          code: error.code || "AUTH_FAILED",
          message: error.message || "登录失败，请稍后重试",
        };
      }
    );
  },

  retryAuthentication() {
    const authentication = auth.retryAuthentication();
    this.globalData.authPromise = authentication;
    authentication.then(
      (session) => {
        this.globalData.authError = null;
        this.globalData.profile = session.profile;
        this.globalData.runtimeConfig = session.runtimeConfig;
      },
      (error) => {
        this.globalData.authError = {
          code: error.code || "AUTH_FAILED",
          message: error.message || "登录失败，请稍后重试",
        };
      }
    );
    return authentication;
  },
});
