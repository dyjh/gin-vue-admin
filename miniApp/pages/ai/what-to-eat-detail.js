const api = require("../../services/api");
const { go } = require("../../utils/navigation");
const { resolveAssetUrl } = require("../../utils/assets");

Page({
  data: {
    firstStepImage: resolveAssetUrl("/assets/images/ai-kung-pao-chicken-step.jpg"),
    result: null,
    savedDishId: "",
    people: 2,
  },

  async onLoad(options = {}) {
    const people = Math.min(Math.max(Number(options.people) || 2, 1), 20);
    const dishIndex = Math.max(Number(options.dishIndex) || 0, 0);
    const generated = await api.generateWhatToEat({
      tags: ["下饭", "少油"],
      people,
      timeLimitMinutes: 30,
      preview: true,
      useFreeQuota: true,
    });
    const dishes = Array.isArray(generated.dishes) && generated.dishes.length
      ? generated.dishes
      : [generated.dish];
    const dish = dishes[dishIndex] || dishes[0];
    const result = {
      ...generated,
      id: dish.recommendationId || generated.id,
      dish,
    };
    this.setData({ result, people });
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
