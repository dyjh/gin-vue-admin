const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    meal: null,
    dishes: [],
    confirming: false,
  },

  async onLoad() {
    const data = await api.getMealStats("meal-1");
    this.setData(data);
  },

  adjust(event) {
    const index = event.currentTarget.dataset.index;
    const delta = Number(event.currentTarget.dataset.delta);
    const value = Math.max(1, this.data.dishes[index].finalServings + delta);
    this.setData({ [`dishes[${index}].finalServings`]: value });
  },

  async confirm() {
    this.setData({ confirming: true });
    try {
      await api.confirmMeal(this.data.meal.id, {
        dishes: this.data.dishes.map((dish) => ({ dishId: dish.id, servings: dish.finalServings })),
      });
      wx.showToast({ title: "菜单已确认", icon: "success" });
      setTimeout(() => go("/pages/shopping/list"), 300);
    } finally {
      this.setData({ confirming: false });
    }
  },
});
