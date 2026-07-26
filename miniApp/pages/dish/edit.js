const api = require("../../services/api");
const { getFeature } = require("../../utils/features");

Page({
  data: {
    id: "",
    dish: null,
    coverFeature: null,
  },

  async onLoad(options) {
    const id = options.id || "dish-1";
    this.setData({ id });
    const [dish] = await Promise.all([
      api.getDish(id),
      api.getRuntimeConfig(),
    ]);
    this.setData({ dish, coverFeature: getFeature("cover_create") });
  },

  async save(event) {
    await api.updateDish(this.data.id, event.detail.form);
    wx.showToast({ title: "修改已保存", icon: "success" });
    setTimeout(() => wx.navigateBack(), 350);
  },
});
