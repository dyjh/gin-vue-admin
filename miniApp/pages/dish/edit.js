const api = require("../../services/api");

Page({
  data: {
    id: "",
    dish: null,
  },

  async onLoad(options) {
    const id = options.id || "dish-1";
    this.setData({ id });
    const dish = await api.getDish(id);
    this.setData({ dish });
  },

  async save(event) {
    await api.updateDish(this.data.id, event.detail.form);
    wx.showToast({ title: "修改已保存", icon: "success" });
    setTimeout(() => wx.navigateBack(), 350);
  },
});
