const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    nav: {},
    profile: null,
    canUseWechatProfile: false,
    meal: null,
    unreadCount: 0,
    dishes: [],
    showEditor: false,
    showAiUsage: false,
    dishTotal: 0,
    aiUsage: [
      { id: "ai-1", name: "不知道吃什么", cost: "8 积分", createdAt: "今天 12:20" },
      { id: "ai-2", name: "AI 备菜提醒", cost: "6 积分", createdAt: "07-16 17:40" }
    ],
  },

  onLoad() {
    const accountInfo = typeof wx.getAccountInfoSync === "function" ? wx.getAccountInfoSync() : {};
    const appId = accountInfo.miniProgram ? accountInfo.miniProgram.appId : "";
    const supportsChooseAvatar = typeof wx.canIUse !== "function" || wx.canIUse("button.open-type.chooseAvatar");
    this.setData({
      nav: getApp().globalData.nav,
      canUseWechatProfile: Boolean(appId && appId !== "touristappid" && supportsChooseAvatar),
    });
  },

  onShow() {
    this.load();
  },

  async load() {
    const [data, dishes] = await Promise.all([
      api.bootstrap(),
      api.listDishes({ page: 1, pageSize: 1 }),
    ]);
    this.setData({
      profile: data.profile,
      meal: data.meal,
      unreadCount: data.unreadCount,
      dishes: data.dishes,
      dishTotal: dishes.total,
    });
  },

  editProfile() {
    this.setData({ showEditor: true });
  },

  closeEditor() {
    this.setData({ showEditor: false });
  },

  chooseAvatar(event) {
    const avatarUrl = event.detail ? event.detail.avatarUrl : "";
    if (!avatarUrl) return;
    this.setData({ "profile.avatarUrl": avatarUrl });
  },

  chooseLocalAvatar() {
    const useAvatar = (avatarUrl) => {
      if (avatarUrl) this.setData({ "profile.avatarUrl": avatarUrl });
    };
    const handleFailure = (error) => {
      if (error && String(error.errMsg || "").includes("cancel")) return;
      wx.showToast({ title: "头像选择失败", icon: "none" });
    };

    if (typeof wx.chooseMedia === "function") {
      wx.chooseMedia({
        count: 1,
        mediaType: ["image"],
        sourceType: ["album", "camera"],
        sizeType: ["compressed"],
        success: (result) => useAvatar(result.tempFiles && result.tempFiles[0] ? result.tempFiles[0].tempFilePath : ""),
        fail: handleFailure,
      });
      return;
    }

    wx.chooseImage({
      count: 1,
      sourceType: ["album", "camera"],
      sizeType: ["compressed"],
      success: (result) => useAvatar(result.tempFilePaths ? result.tempFilePaths[0] : ""),
      fail: handleFailure,
    });
  },

  nicknameInput(event) {
    this.setData({ "profile.nickname": event.detail.value });
  },

  async saveProfile() {
    const profile = await api.updateProfile(this.data.profile);
    this.setData({ profile, showEditor: false });
    wx.showToast({ title: "资料已更新", icon: "success" });
  },

  openPoints() { go("/pages/points/index"); },
  openCheckin() { go("/pages/checkin/index"); },
  createMeal() { go("/pages/meal/create"); },
  joinMeal() { go("/pages/meal/vote", { state: "join" }); },
  openMeal() { go("/pages/meal/invite"); },
  openShopping() { go("/pages/shopping/list"); },
  openNotifications() { go("/pages/notifications/index"); },
  openAllDishes() { go("/pages/dish/list"); },
  openAiUsage() { this.setData({ showAiUsage: true }); },
  closeAiUsage() { this.setData({ showAiUsage: false }); },
  noop() {},
});
