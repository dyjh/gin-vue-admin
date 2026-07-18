const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    name: "周六家庭聚餐",
    deadline: "明天 20:00",
    source: "dishes",
    dishes: [],
    recipes: [],
    recipeIndex: 0,
    selectedIds: [],
    loading: false,
  },

  async onLoad(options) {
    const [dishes, recipes] = await Promise.all([
      api.listDishes({ page: 1, pageSize: 100, status: "usable" }),
      api.listRecipes(),
    ]);
    let selectedIds = dishes.list.slice(0, 4).map((dish) => dish.id);
    let source = "dishes";
    let recipeIndex = 0;
    if (options.recipeId) {
      const index = recipes.findIndex((recipe) => recipe.id === options.recipeId);
      if (index >= 0) {
        source = "recipe";
        recipeIndex = index;
        selectedIds = [...recipes[index].dishIds];
      }
    }
    this.setData({ dishes: dishes.list, recipes, selectedIds, source, recipeIndex });
  },

  inputName(event) {
    this.setData({ name: event.detail.value });
  },

  chooseDeadline(event) {
    const deadlines = ["今天 20:00", "今天 21:00", "明天 12:00", "明天 20:00"];
    this.setData({ deadline: deadlines[event.detail.value] });
  },

  setSource(event) {
    const source = event.currentTarget.dataset.source;
    if (source === "recipe" && this.data.recipes.length) {
      this.setData({ source, selectedIds: [...this.data.recipes[this.data.recipeIndex].dishIds] });
    } else {
      this.setData({ source });
    }
  },

  chooseRecipe(event) {
    const recipeIndex = Number(event.detail.value);
    this.setData({ recipeIndex, selectedIds: [...this.data.recipes[recipeIndex].dishIds] });
  },

  toggleDish(event) {
    const id = event.detail.id;
    const selectedIds = [...this.data.selectedIds];
    const index = selectedIds.indexOf(id);
    if (index >= 0) selectedIds.splice(index, 1);
    else selectedIds.push(id);
    this.setData({ selectedIds });
  },

  async create() {
    if (!this.data.name.trim()) {
      wx.showToast({ title: "请填写饭局名称", icon: "none" });
      return;
    }
    if (!this.data.selectedIds.length) {
      wx.showToast({ title: "至少选择一道候选菜", icon: "none" });
      return;
    }
    this.setData({ loading: true });
    try {
      await api.createMeal({
        name: this.data.name,
        deadline: this.data.deadline,
        candidateIds: this.data.selectedIds,
      });
      wx.showToast({ title: "饭局已创建", icon: "success" });
      setTimeout(() => go("/pages/meal/invite"), 300);
    } finally {
      this.setData({ loading: false });
    }
  },
});
