const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    id: "",
    dish: null,
    personalDishId: "",
  },

  onLoad(options) {
    this.setData({ id: options.id || "recommend-1" });
    this.load();
  },

  async load() {
    const dish = await api.getRecommendation(this.data.id);
    this.setData({ dish });
  },

  async copy() {
    if (this.data.personalDishId) {
      go("/pages/dish/detail", { id: this.data.personalDishId });
      return;
    }
    const dish = await api.copyRecommendation(this.data.id);
    this.setData({ personalDishId: dish.id, "dish.copied": true });
    wx.showToast({ title: "已加入菜品库", icon: "success" });
  },
});
