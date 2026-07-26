const api = require("../../services/api");
const { getFeature } = require("../../utils/features");

function nowLabel() {
  const now = new Date();
  return `今天 ${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;
}

Page({
  data: {
    meal: null,
    menuItems: [],
    dishCount: 0,
    totalServings: 0,
    plan: null,
    ignored: false,
    showRegenerate: false,
    generating: false,
    generatedAt: "今天 10:24",
    points: 0,
    cost: 0,
    remainingPoints: 0,
    canGenerate: false,
    prepFeature: null,
  },

  async onLoad() {
    await api.getRuntimeConfig();
    const prepFeature = getFeature("prep_sequence");
    if (!prepFeature) {
      wx.showToast({ title: "该功能暂不可用", icon: "none" });
      setTimeout(() => wx.navigateBack(), 300);
      return;
    }
    this.setData({ prepFeature });
    await this.loadMeal();
  },

  async loadMeal() {
    try {
      const current = await api.getCurrentMeal();
      if (!current.meal) return;
      const meal = current.meal;
      const menuItems = meal.finalDishes;
      const quote = await api.getPrepPlanQuote(meal.id);
      this.setData({
        meal,
        menuItems,
        dishCount: menuItems.length,
        totalServings: menuItems.reduce((sum, item) => sum + item.finalServings, 0),
        points: quote.pointBalance,
        cost: quote.pointCost,
        remainingPoints: Math.max(quote.pointBalance - quote.pointCost, 0),
        canGenerate: quote.canGenerate,
      });
    } catch (error) {
      // Keep the approved static summary visible if the meal refresh fails.
    }
  },

  generate() {
    if (!this.data.generating && this.data.canGenerate) this.runGenerate(false);
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
      const excludedPlanIds = isRegenerate && this.data.plan ? [this.data.plan.id] : [];
      const plan = await api.generatePrepPlan({
        mealId: this.data.meal.id,
        excludedPlanIds,
      });
      this.setData({
        plan: {
          ...plan,
          steps: plan.steps.map((step) => ({
            ...step,
            parallelText: step.parallelActions.join("；"),
          })),
        },
        ignored: false,
        showRegenerate: false,
        generatedAt: nowLabel(),
        points: plan.usage.pointBalance,
        remainingPoints: Math.max(plan.usage.pointBalance - this.data.cost, 0),
      });
      wx.showToast({ title: isRegenerate ? "已重新生成备菜提醒" : "备菜提醒已生成", icon: "success" });
    } catch (error) {
      wx.showToast({ title: "生成失败，积分已退还", icon: "none" });
    } finally {
      this.setData({ generating: false });
    }
  },

  async ignore() {
    await api.submitPrepPlanFeedback(this.data.plan.id, "ignored");
    this.setData({ ignored: true });
    wx.showToast({ title: "已忽略本次提醒", icon: "none" });
  },
});
