const api = require("../../services/api");

const DEFAULT_MENU = [
  { id: "dish-1", name: "香菇鸡腿饭", servings: 2 },
  { id: "dish-2", name: "番茄炒蛋", servings: 1 },
  { id: "dish-4", name: "冬瓜丸子汤", servings: 1 },
  { id: "dish-3", name: "蒜蓉西兰花", servings: 1 },
];

function nowLabel() {
  const now = new Date();
  return `今天 ${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;
}

Page({
  data: {
    meal: { id: "meal-1", name: "周六家庭聚餐", participantCount: 4 },
    menuItems: DEFAULT_MENU,
    dishCount: 4,
    totalServings: 5,
    plan: null,
    ignored: false,
    showRegenerate: false,
    generating: false,
    generatedAt: "今天 10:24",
    points: 128,
    cost: 8,
    remainingPoints: 120,
  },

  onLoad() {
    this.loadMeal();
  },

  async loadMeal() {
    try {
      const meal = await api.getCurrentMeal();
      const candidates = Array.isArray(meal.candidates) ? meal.candidates.slice(0, 4) : [];
      const menuItems = candidates.length
        ? candidates.map((dish, index) => ({ id: dish.id, name: dish.name, servings: index === 0 ? 2 : 1 }))
        : DEFAULT_MENU;
      this.setData({
        meal,
        menuItems,
        dishCount: menuItems.length,
        totalServings: menuItems.reduce((sum, item) => sum + item.servings, 0),
      });
    } catch (error) {
      // Keep the approved static summary visible if the meal refresh fails.
    }
  },

  generate() {
    if (!this.data.generating) this.runGenerate(false);
  },

  openRegenerate() {
    this.setData({ showRegenerate: true });
  },

  closeRegenerate() {
    if (!this.data.generating) this.setData({ showRegenerate: false });
  },

  noop() {},

  confirmRegenerate() {
    if (!this.data.generating) this.runGenerate(true);
  },

  async runGenerate(isRegenerate) {
    this.setData({ generating: true });
    try {
      const plan = await api.generatePrepPlan({ mealId: this.data.meal.id || "meal-1" });
      this.setData({
        plan,
        ignored: false,
        showRegenerate: false,
        generatedAt: nowLabel(),
      });
      wx.showToast({ title: isRegenerate ? "已重新生成备菜提醒" : "备菜提醒已生成", icon: "success" });
    } catch (error) {
      wx.showToast({ title: "生成失败，积分已退还", icon: "none" });
    } finally {
      this.setData({ generating: false });
    }
  },

  ignore() {
    this.setData({ ignored: true });
    wx.showToast({ title: "已忽略本次提醒", icon: "none" });
  },
});