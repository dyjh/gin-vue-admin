const api = require("../../services/api");
const { go } = require("../../utils/navigation");

const DEADLINE_OPTIONS = [
  { label: "今天 20:00", dayOffset: 0, hour: 20 },
  { label: "今天 21:00", dayOffset: 0, hour: 21 },
  { label: "明天 12:00", dayOffset: 1, hour: 12 },
  { label: "明天 20:00", dayOffset: 1, hour: 20 },
];

function buildDeadlineAt(option) {
  const deadline = new Date();
  deadline.setDate(deadline.getDate() + option.dayOffset);
  deadline.setHours(option.hour, 0, 0, 0);
  return deadline.toISOString();
}

Page({
  data: {
    name: "周六家庭聚餐",
    deadlineOptions: DEADLINE_OPTIONS.map((item) => item.label),
    deadlineIndex: 3,
    deadlineLabel: DEADLINE_OPTIONS[3].label,
    source: "dishes",
    dishes: [],
    visibleDishes: [],
    categories: [],
    category: "全部",
    listHeight: 320,
    recipes: [],
    recipeIndex: 0,
    selectedIds: [],
    confirmOpen: false,
    loading: false,
  },

  async onLoad(options) {
    const [dishes, recipeResponse] = await Promise.all([
      api.listDishes({ page: 1, pageSize: 100, status: "usable" }),
      api.listRecipes(),
    ]);
    const recipes = recipeResponse.list;
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
    this.setData({
      dishes: dishes.list,
      recipes,
      selectedIds,
      source,
      recipeIndex,
      ...this.buildCategoryState(dishes.list, "全部"),
    });
  },

  onReady() {
    this.measureListHeight();
  },

  onResize() {
    this.measureListHeight();
  },

  measureListHeight() {
    wx.nextTick(() => {
      this.createSelectorQuery()
        .select(".category-layout")
        .boundingClientRect((rect) => {
          const listHeight = rect ? Math.floor(rect.height) : 0;
          if (listHeight > 0 && listHeight !== this.data.listHeight) {
            this.setData({ listHeight });
          }
        })
        .exec();
    });
  },

  buildCategoryState(dishes, category) {
    const names = ["全部", ...Array.from(new Set(dishes.map((dish) => dish.category).filter(Boolean)))];
    const categories = names.map((name) => ({
      name,
      count: name === "全部" ? dishes.length : dishes.filter((dish) => dish.category === name).length,
    }));
    return {
      categories,
      category,
      visibleDishes: category === "全部" ? dishes : dishes.filter((dish) => dish.category === category),
    };
  },

  inputName(event) {
    this.setData({ name: event.detail.value });
  },

  chooseDeadline(event) {
    const deadlineIndex = Number(event.detail.value);
    this.setData({
      deadlineIndex,
      deadlineLabel: DEADLINE_OPTIONS[deadlineIndex].label,
    });
  },

  setSource(event) {
    const source = event.currentTarget.dataset.source;
    if (source === "recipe") {
      if (!this.data.recipes.length) {
        wx.showToast({ title: "还没有可用菜谱", icon: "none" });
        return;
      }
      this.setData({
        source,
        selectedIds: [...this.data.recipes[this.data.recipeIndex].dishIds],
        ...this.buildCategoryState(this.data.dishes, "全部"),
      }, () => this.measureListHeight());
      return;
    }
    this.setData({ source }, () => this.measureListHeight());
  },

  chooseRecipe(event) {
    const recipeIndex = Number(event.detail.value);
    this.setData({
      recipeIndex,
      selectedIds: [...this.data.recipes[recipeIndex].dishIds],
      ...this.buildCategoryState(this.data.dishes, "全部"),
    });
  },

  selectCategory(event) {
    this.setData(this.buildCategoryState(this.data.dishes, event.currentTarget.dataset.category));
  },

  toggleDish(event) {
    const id = event.detail.id;
    const selectedIds = [...this.data.selectedIds];
    const index = selectedIds.indexOf(id);
    if (index >= 0) selectedIds.splice(index, 1);
    else selectedIds.push(id);
    this.setData({ selectedIds });
  },

  openConfirm() {
    if (!this.data.selectedIds.length) {
      wx.showToast({ title: "至少选择一道候选菜", icon: "none" });
      return;
    }
    this.setData({ confirmOpen: true });
  },

  closeConfirm() {
    if (!this.data.loading) this.setData({ confirmOpen: false });
  },

  noop() {},

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
      const payload = {
        name: this.data.name,
        deadlineAt: buildDeadlineAt(DEADLINE_OPTIONS[this.data.deadlineIndex]),
        candidateDishIds: this.data.selectedIds,
      };
      if (this.data.source === "recipe") {
        payload.sourceRecipeId = this.data.recipes[this.data.recipeIndex].id;
      }
      await api.createMeal(payload);
      this.setData({ confirmOpen: false });
      wx.showToast({ title: "饭局已创建", icon: "success" });
      setTimeout(() => go("/pages/meal/invite"), 300);
    } finally {
      this.setData({ loading: false });
    }
  },
});
