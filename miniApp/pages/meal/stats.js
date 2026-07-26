const api = require("../../services/api");
const { go } = require("../../utils/navigation");

function summarize(dishes) {
  return dishes.reduce((summary, dish) => {
    if (!dish.selected) return summary;
    return {
      includedDishCount: summary.includedDishCount + 1,
      totalServings: summary.totalServings + dish.finalServings,
    };
  }, { includedDishCount: 0, totalServings: 0 });
}

Page({
  data: {
    meal: null,
    dishes: [],
    includedDishCount: 0,
    totalServings: 0,
    showConfirm: false,
    confirming: false,
  },

  async onLoad() {
    const current = await api.getCurrentMeal();
    if (!current.meal) return;
    const data = await api.getMealStats(current.meal.id);
    this.setData({ ...data, ...summarize(data.dishes) });
  },

  toggleIncluded(event) {
    const index = Number(event.currentTarget.dataset.index);
    const dish = this.data.dishes[index];
    if (!dish) return;

    const selected = !dish.selected;
    const servingDelta = selected ? dish.finalServings : -dish.finalServings;
    this.setData({
      [`dishes[${index}].selected`]: selected,
      includedDishCount: this.data.includedDishCount + (selected ? 1 : -1),
      totalServings: this.data.totalServings + servingDelta,
    });
  },

  adjust(event) {
    const index = Number(event.currentTarget.dataset.index);
    const delta = Number(event.currentTarget.dataset.delta);
    const dish = this.data.dishes[index];
    if (!dish || !dish.selected) return;

    const finalServings = Math.max(1, dish.finalServings + delta);
    if (finalServings === dish.finalServings) return;

    this.setData({
      [`dishes[${index}].finalServings`]: finalServings,
      totalServings: this.data.totalServings + finalServings - dish.finalServings,
    });
  },

  openConfirm() {
    if (!this.data.includedDishCount) {
      wx.showToast({ title: "请至少保留一道菜", icon: "none" });
      return;
    }
    this.setData({ showConfirm: true });
  },

  closeConfirm() {
    if (!this.data.confirming) this.setData({ showConfirm: false });
  },

  noop() {},

  async confirm() {
    this.setData({ confirming: true });
    try {
      await api.confirmMeal(this.data.meal.id, {
        dishes: this.data.dishes
          .map((dish) => ({
            candidateId: dish.candidateId,
            selected: dish.selected,
            finalServings: dish.selected ? dish.finalServings : 0,
          })),
      });
      this.setData({ showConfirm: false });
      wx.showToast({ title: "采购清单已生成", icon: "success" });
      setTimeout(() => go("/pages/shopping/list"), 450);
    } finally {
      this.setData({ confirming: false });
    }
  },
});
