const api = require("../../services/api");
const { go } = require("../../utils/navigation");
const { getFeature } = require("../../utils/features");

Page({
  data: {
    result: null,
    savedDishId: "",
    people: 2,
    suggestionFeature: null,
  },

  async onLoad(options = {}) {
    await api.getRuntimeConfig();
    const suggestionFeature = getFeature("meal_suggest");
    if (!suggestionFeature) {
      wx.showToast({ title: "该功能暂不可用", icon: "none" });
      setTimeout(() => wx.navigateBack(), 300);
      return;
    }
    const people = Math.min(Math.max(Number(options.people) || 2, 1), 20);
    const dishIndex = Math.max(Number(options.dishIndex) || 0, 0);
    const generated = await api.getMealSuggestion(options.id);
    const dishes = Array.isArray(generated.dishes) && generated.dishes.length
      ? generated.dishes
      : [generated.dish];
    const dish = dishes[dishIndex] || dishes[0];
    const result = {
      ...generated,
      dish,
    };
    this.setData({
      result,
      people,
      suggestionFeature,
      savedDishId: dish.copied && dish.dishId ? dish.dishId : "",
    });
  },

  async save() {
    if (this.data.savedDishId) {
      go("/pages/dish/detail", { id: this.data.savedDishId });
      return;
    }
    const result = await api.saveSuggestedDish(
      this.data.result.id,
      this.data.result.dish.id,
    );
    this.setData({ savedDishId: result.dish.id });
    wx.showToast({ title: "已保存到菜品库", icon: "success" });
  },

  change() {
    wx.navigateBack();
  },
});
