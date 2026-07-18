const api = require("../../services/api");
const { home } = require("../../utils/navigation");

Page({
  data: {
    state: "collecting",
    code: "",
    meal: null,
    categories: [],
    category: "全部",
    selectedIds: [],
    saving: false,
  },

  async onLoad(options) {
    this.setData({ state: options.state || "collecting", code: options.code || "" });
    if (options.state !== "join") await this.load();
  },

  async load() {
    const meal = await api.getCurrentMeal();
    const categories = ["全部", ...Array.from(new Set(meal.candidates.map((dish) => dish.category)))];
    this.setData({
      meal,
      categories,
      selectedIds: [...meal.selectedIds],
      state: meal.status === "cancelled" ? "cancelled" : meal.status === "collecting" ? "collecting" : "closed",
    });
  },

  inputCode(event) {
    this.setData({ code: event.detail.value.replace(/\D/g, "").slice(0, 6) });
  },

  async join() {
    if (this.data.code.length !== 6) {
      wx.showToast({ title: "请输入 6 位点餐码", icon: "none" });
      return;
    }
    try {
      const meal = await api.joinMeal({ code: this.data.code });
      const categories = ["全部", ...Array.from(new Set(meal.candidates.map((dish) => dish.category)))];
      this.setData({ meal, categories, selectedIds: [...meal.selectedIds], state: "collecting" });
    } catch (error) {
      // The request layer already presents the reason.
    }
  },

  selectCategory(event) {
    this.setData({ category: event.currentTarget.dataset.category });
  },

  toggleDish(event) {
    const id = event.detail.id;
    const selectedIds = [...this.data.selectedIds];
    const index = selectedIds.indexOf(id);
    if (index >= 0) selectedIds.splice(index, 1);
    else selectedIds.push(id);
    this.setData({ selectedIds });
  },

  async save() {
    this.setData({ saving: true });
    try {
      await api.saveVotes(this.data.meal.id, this.data.selectedIds);
      wx.showToast({ title: "点菜已保存", icon: "success" });
    } finally {
      this.setData({ saving: false });
    }
  },

  home,
});
