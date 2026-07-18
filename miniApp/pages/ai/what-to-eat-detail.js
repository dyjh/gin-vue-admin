const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    result: null,
    savedDishId: "",
  },

  async onLoad() {
    const result = await api.generateWhatToEat({ tags: ["下饭", "少油"], people: 2, timeLimit: "30 分钟内", preview: true, useFreeQuota: true });
    this.setData({ result });
  },

  async save() {
    if (this.data.savedDishId) {
      go("/pages/dish/detail", { id: this.data.savedDishId });
      return;
    }
    const dish = await api.saveAiDish(this.data.result.id);
    this.setData({ savedDishId: dish.id });
    wx.showToast({ title: "已保存到菜品库", icon: "success" });
  },

  change() {
    wx.navigateBack();
  },
});
