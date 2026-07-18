const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    nav: {},
    profile: null,
    meal: null,
    unreadCount: 0,
    dishes: [],
    showEditor: false,
    showAllDishes: false,
    showAiUsage: false,
    allDishes: [],
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
    const [data, dishes] = await Promise.all([
      api.bootstrap(),
      api.listDishes({ page: 1, pageSize: 100 }),
    ]);
    this.setData({
      profile: data.profile,
      meal: data.meal,
      unreadCount: data.unreadCount,
      dishes: data.dishes,
      allDishes: dishes.list,
    });
  },

  editProfile() {
    this.setData({ showEditor: true });
  },

  closeEditor() {
    this.setData({ showEditor: false });
  },

  chooseAvatar(event) {
    this.setData({ "profile.avatarUrl": event.detail.avatarUrl });
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
  openAllDishes() { this.setData({ showAllDishes: true }); },
  closeAllDishes() { this.setData({ showAllDishes: false }); },
  openAiUsage() { this.setData({ showAiUsage: true }); },
  closeAiUsage() { this.setData({ showAiUsage: false }); },
  openDish(event) { go("/pages/dish/detail", { id: event.detail.id }); },
  noop() {},
});
