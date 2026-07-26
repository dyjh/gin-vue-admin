const api = require("../../services/api");
const { go } = require("../../utils/navigation");
const { resolveAssetUrl } = require("../../utils/assets");
const { pointsEnabled, runtimeConfig } = require("../../utils/features");

function mealStatusLabel(status) {
  return ({
    collecting: "收集中",
    closed: "待确认菜单",
    confirmed: "采购进行中",
  })[status] || status;
}

Page({
  data: {
    heroImage: resolveAssetUrl("/assets/images/profile-cat-organizing-v5.jpg"),
    nav: {},
    profile: null,
    activeMeals: [],
    activeMealTotal: 0,
    activeMealDetail: null,
    showActiveMealDetail: false,
    unreadCount: 0,
    dishes: [],
    showEditor: false,
    showFeatureUsage: false,
    enhancedFeaturesEnabled: false,
    pointsEnabled: false,
    dishTotal: 0,
    mealHistoryTotal: 0,
    pendingShoppingCount: 0,
    featureUsageTotal: 0,
    featureUsage: [],
  },

  onLoad() {
    this.setData({ nav: getApp().globalData.nav });
  },

  onShow() {
    this.load();
  },

  async load() {
    const data = await api.bootstrap();
    const config = runtimeConfig();
    const enhancedFeaturesEnabled = Boolean(config && config.enhancedFeaturesEnabled);
    const [dishes, activeMealPage, mealHistories, usagePage] = await Promise.all([
      api.listDishes({ page: 1, pageSize: 1 }),
      api.listMeals({ page: 1, pageSize: 100, scope: "active" }),
      api.listMeals({ page: 1, pageSize: 1, scope: "history" }),
      enhancedFeaturesEnabled
        ? api.listFeatureUsages({ page: 1, pageSize: 20 })
        : Promise.resolve({ total: 0, list: [] }),
    ]);
    let pendingShoppingCount = 0;
    if (data.activeMeal && data.activeMeal.status === "confirmed") {
      try {
        const shoppingList = await api.getShoppingList({ showError: false });
        pendingShoppingCount = shoppingList.pendingCount;
      } catch (error) {
        pendingShoppingCount = 0;
      }
    }
    this.setData({
      profile: {
        ...data.profile,
        savedAvatarUrl: data.profile.avatarUrl,
      },
      activeMeals: activeMealPage.list.map((meal) => ({
        ...meal,
        statusLabel: mealStatusLabel(meal.status),
      })),
      activeMealTotal: activeMealPage.total,
      unreadCount: data.unreadCount,
      dishes: data.dishes,
      dishTotal: dishes.total,
      mealHistoryTotal: mealHistories.total,
      pendingShoppingCount,
      enhancedFeaturesEnabled,
      pointsEnabled: pointsEnabled(),
      featureUsageTotal: usagePage.total,
      featureUsage: usagePage.list.map((item) => ({
        id: item.id,
        name: item.feature,
        cost: `${item.pointCost} 积分`,
        createdAt: item.createdAt,
      })),
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
    this.setData({
      profile: { ...profile, savedAvatarUrl: profile.avatarUrl },
      showEditor: false,
    });
    wx.showToast({ title: "资料已更新", icon: "success" });
  },

  openPoints() { go("/pages/points/index"); },
  openCheckin() { go("/pages/checkin/index"); },
  createMeal() { go("/pages/meal/create"); },
  joinMeal() { go("/pages/meal/vote", { state: "join" }); },
  async openActiveMeal(event) {
    const mealId = String(event.currentTarget.dataset.id || "");
    if (!mealId) return;
    const meal = await api.getMeal(mealId);
    if (meal.createdByMe) {
      if (meal.status === "collecting") {
        go("/pages/meal/invite");
        return;
      }
      if (meal.status === "closed") {
        go("/pages/meal/stats");
        return;
      }
      if (meal.status === "confirmed") {
        go("/pages/shopping/list");
        return;
      }
    }
    if (meal.status === "collecting") {
      go("/pages/meal/vote", { mealId });
      return;
    }
    const displayDishes = (meal.finalDishes.length ? meal.finalDishes : meal.candidates)
      .map((item) => ({
        ...item,
        displayId: item.id || item.candidateId,
      }));
    this.setData({
      activeMealDetail: {
        ...meal,
        statusLabel: mealStatusLabel(meal.status),
        displayDishes,
      },
      showActiveMealDetail: true,
    });
  },
  closeActiveMealDetail() {
    this.setData({ showActiveMealDetail: false, activeMealDetail: null });
  },
  openShopping() { go("/pages/shopping/list"); },
  openNotifications() { go("/pages/notifications/index"); },
  openAllDishes() { go("/pages/dish/list"); },
  openMealHistory() { go("/pages/meal/history"); },
  openFeatureUsage() { this.setData({ showFeatureUsage: true }); },
  closeFeatureUsage() { this.setData({ showFeatureUsage: false }); },
  noop() {},
});
