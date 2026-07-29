const api = require("../../services/api");
const { go } = require("../../utils/navigation");
const { resolveAssetUrl } = require("../../utils/assets");
const { getFeature } = require("../../utils/features");

function getHomeBadge(dish) {
  if (Number(dish.recipeCount) > 0) return { label: "菜谱内", tone: "blue" };
  if (Number(dish.mealCount) > 0) return { label: "常点", tone: "warm" };
  if (dish.status === "draft") return { label: "草稿", tone: "gray" };
  return { label: "已完善", tone: "green" };
}

function decorateDishes(dishes = []) {
  return dishes.map((dish) => ({
    ...dish,
    homeBadge: getHomeBadge(dish),
  }));
}

function filterDishes(dishes, query, category) {
  const keyword = (query || "").trim().toLowerCase();
  return dishes.filter((dish) => {
    if (category && dish.category !== category) return false;
    if (!keyword) return true;

    const ingredients = (dish.ingredients || []).map((item) => item.name).join(" ");
    const searchText = [dish.name, dish.category, ...(dish.tags || []), ingredients].join(" ").toLowerCase();
    return searchText.includes(keyword);
  }).slice(0, 3);
}

Page({
  data: {
    heroImage: resolveAssetUrl("/assets/images/home-approved-header-v3.jpg"),

    loading: true,
    query: "",
    dishFilter: "recent",
    categoryLabel: "",
    dishTotal: 0,
    allDishes: [],
    frequentDishes: [],
    dishes: [],
    recommendations: [],
    suggestionFeature: null,
  },


  onShow() {
    this.load();
  },

  onPullDownRefresh() {
    this.load().finally(() => wx.stopPullDownRefresh());
  },

  onUnload() {
    clearTimeout(this.dishSearchTimer);
  },

  async load() {
    this.setData({ loading: true });
    try {
      const data = await api.bootstrap();
      const allDishes = decorateDishes(data.dishes);
      this.setData({
        dishTotal: Number(data.dishTotal ?? allDishes.length),
        dishFilter: "recent",
        categoryLabel: "",
        allDishes,
        frequentDishes: [],
        dishes: filterDishes(allDishes, this.data.query, ""),
        recommendations: (data.recommendations || []).slice(0, 3),
        suggestionFeature: getFeature("meal_suggest"),
      });
      if (this.data.query) {
        await this.loadVisibleDishes();
      }
    } finally {
      this.setData({ loading: false });
    }
  },

  search(event) {
    const query = event.detail.value;
    const source = this.data.dishFilter === "frequent"
      ? this.data.frequentDishes
      : this.data.allDishes;
    this.setData({
      query,
      dishes: filterDishes(source, query, this.data.categoryLabel),
    });
    clearTimeout(this.dishSearchTimer);
    const requestId = (this.dishSearchRequest || 0) + 1;
    this.dishSearchRequest = requestId;
    this.dishSearchTimer = setTimeout(() => {
      this.loadVisibleDishes({ requestId });
    }, 280);
  },

  async loadVisibleDishes(options = {}) {
    const requestId = options.requestId || (this.dishSearchRequest || 0) + 1;
    this.dishSearchRequest = requestId;
    const query = this.data.query;
    const filter = this.data.dishFilter;
    const category = this.data.categoryLabel;
    const result = await api.listDishes({
      page: 1,
      pageSize: 3,
      q: query,
      category,
      sort: filter === "frequent" ? "frequent" : "recent",
    });
    if (requestId !== this.dishSearchRequest) return;

    const dishes = decorateDishes(result.list);
    const patch = { dishes };
    if (!query && !category && filter === "recent") {
      patch.allDishes = dishes;
    }
    if (!query && !category && filter === "frequent") {
      patch.frequentDishes = dishes;
    }
    this.setData(patch);
  },

  async setDishFilter(event) {
    const filter = event.currentTarget.dataset.filter;
    if (filter === "category") {
      this.chooseCategory();
      return;
    }
    if (filter === "frequent") {
      this.setData({
        dishFilter: "frequent",
        categoryLabel: "",
      });
      await this.loadVisibleDishes();
      return;
    }

    this.setData({
      dishFilter: "recent",
      categoryLabel: "",
    });
    await this.loadVisibleDishes();
  },

  chooseCategory() {
    const categories = Array.from(new Set(this.data.allDishes.map((dish) => dish.category)));
    wx.showActionSheet({
      itemList: ["全部分类"].concat(categories),
      success: ({ tapIndex }) => {
        const categoryLabel = tapIndex === 0 ? "" : categories[tapIndex - 1];
        this.setData({
          dishFilter: categoryLabel ? "category" : "recent",
          categoryLabel,
        });
        this.loadVisibleDishes();
      },
    });
  },

  openManual() {
    go("/pages/dish/add-entry", { tab: "manual" });
  },

  openDishLibrary() {
    go("/pages/dish/list");
  },

  openDish(event) {
    go("/pages/dish/detail", { id: event.currentTarget.dataset.id });
  },

  openRecommendation(event) {
    go("/pages/recommend/dish-detail", { id: event.currentTarget.dataset.id });
  },

  openRecommendations() {
    go("/pages/recommend/dishes");
  },

  async copyRecommendation(event) {
    const id = event.currentTarget.dataset.id;
    const current = this.data.recommendations.find((item) => item.recommendationId === id);
    if (current && current.copied) {
      wx.showToast({ title: "已在菜品库中", icon: "none" });
      return;
    }

    await api.copyRecommendation(id);
    this.setData({
      recommendations: this.data.recommendations.map((item) => (
        item.recommendationId === id ? { ...item, copied: true } : item
      )),
    });
    wx.showToast({ title: "已加入菜品库", icon: "success" });
  },

  openWhatToEat() {
    go("/pages/ideas/what-to-eat");
  },
});
