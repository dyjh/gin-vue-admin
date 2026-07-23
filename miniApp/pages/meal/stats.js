const api = require("../../services/api");
const { go } = require("../../utils/navigation");

function summarize(dishes) {
  return dishes.reduce((summary, dish) => {
    if (!dish.included) return summary;
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
    const data = await api.getMealStats("meal-1");
    const dishes = data.dishes.map((dish) => ({ ...dish, included: dish.included !== false }));
    this.setData({ ...data, dishes, ...summarize(dishes) });
  },

  toggleIncluded(event) {
    const index = Number(event.currentTarget.dataset.index);
    const dish = this.data.dishes[index];
    if (!dish) return;

    const included = !dish.included;
    const servingDelta = included ? dish.finalServings : -dish.finalServings;
    this.setData({
      [`dishes[${index}].included`]: included,
      includedDishCount: this.data.includedDishCount + (included ? 1 : -1),
      totalServings: this.data.totalServings + servingDelta,
    });
  },

  adjust(event) {
    const index = Number(event.currentTarget.dataset.index);
    const delta = Number(event.currentTarget.dataset.delta);
    const dish = this.data.dishes[index];
    if (!dish || !dish.included) return;

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
          .filter((dish) => dish.included)
          .map((dish) => ({ dishId: dish.id, servings: dish.finalServings })),
      });
      this.setData({ showConfirm: false });
      wx.showToast({ title: "采购清单已生成", icon: "success" });
      setTimeout(() => go("/pages/shopping/list"), 450);
    } finally {
      this.setData({ confirming: false });
    }
  },
});