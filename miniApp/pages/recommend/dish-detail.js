const api = require("../../services/api");

Page({
  data: {
    id: "",
    dish: null,

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
    if (this.data.dish?.copied) return;
    await api.copyRecommendation(this.data.id);
    this.setData({ "dish.copied": true });
    wx.showToast({ title: "已加入菜品库", icon: "success" });
  },
});
