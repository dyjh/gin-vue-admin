const api = require("../../services/api");
const { go } = require("../../utils/navigation");
const { resolveAssetUrl } = require("../../utils/assets");

Page({
  data: {
    heroImage: resolveAssetUrl("/assets/images/profile-cat-organizing-v5.jpg"),
    nav: {},
    profile: null,
    meal: null,
    unreadCount: 0,
    dishes: [],
    showEditor: false,
    showAiUsage: false,
    dishTotal: 0,
    mealHistoryTotal: 0,
    aiUsage: [
      { id: "ai-1", name: "不知道吃什么", cost: "8 积分", createdAt: "今天 12:20" },
      { id: "ai-2", name: "AI 备菜提醒", cost: "6 积分", createdAt: "07-16 17:40" }
    ],
  },

  onLoad() {
    this.setData({ nav: getApp().globalData.nav });
  },

  onShow() {
    this.load();
  },

  async load() {
    const [data, dishes, mealHistories] = await Promise.all([
      api.bootstrap(),
      api.listDishes({ page: 1, pageSize: 1 }),
      api.listMeals({ page: 1, pageSize: 1, scope: "history" }),
    ]);
    this.setData({
      profile: data.profile,
      meal: data.meal,
      unreadCount: data.unreadCount,
      dishes: data.dishes,
      dishTotal: dishes.total,
      mealHistoryTotal: mealHistories.total,
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
  openMealHistory() { go("/pages/meal/history"); },
  openAiUsage() { this.setData({ showAiUsage: true }); },
  closeAiUsage() { this.setData({ showAiUsage: false }); },
  noop() {},
});
