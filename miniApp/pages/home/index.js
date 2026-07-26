const api = require("../../services/api");
const { go } = require("../../utils/navigation");
const { resolveAssetUrl } = require("../../utils/assets");
const { getFeature } = require("../../utils/features");

const HOME_BADGES = [
  { label: "已完善", tone: "green" },
  { label: "菜谱内", tone: "blue" },
  { label: "常点", tone: "warm" },
];

function decorateDishes(dishes) {
  return dishes.map((dish, index) => ({
    ...dish,
    homeBadge: HOME_BADGES[index % HOME_BADGES.length],
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
    heroImage: resolveAssetUrl("/assets/images/home-approved-header-v1.jpg"),
    nav: {
      statusBarHeight: 20,
      menuTop: 24,
      menuHeight: 32,
      navHeight: 88,
    },
    loading: true,
    query: "",
    dishFilter: "recent",
    categoryLabel: "",
    allDishes: [],
    dishes: [],
    recommendations: [],
    extractFeature: null,
    coverFeature: null,
    suggestionFeature: null,
  },

  onLoad() {
    this.setData({ nav: getApp().globalData.nav || this.data.nav });
  },

  onShow() {
    this.load();
  },

  onPullDownRefresh() {
    this.load().finally(() => wx.stopPullDownRefresh());
  },

  async load() {
    this.setData({ loading: true });
    try {
      const data = await api.bootstrap();
      const allDishes = decorateDishes(data.dishes);
      this.setData({
        allDishes,
        dishes: filterDishes(allDishes, this.data.query, this.data.categoryLabel),
        recommendations: data.recommendations,
        extractFeature: getFeature("dish_extract"),
        coverFeature: getFeature("cover_create"),
        suggestionFeature: getFeature("meal_suggest"),
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  search(event) {
    const query = event.detail.value;
    this.setData({
      query,
      dishes: filterDishes(this.data.allDishes, query, this.data.categoryLabel),
    });
  },

  setDishFilter(event) {
    const filter = event.currentTarget.dataset.filter;
    if (filter === "category") {
      this.chooseCategory();
      return;
    }

    this.setData({
      dishFilter: filter,
      categoryLabel: "",
      dishes: filterDishes(this.data.allDishes, this.data.query, ""),
    });
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
          dishes: filterDishes(this.data.allDishes, this.data.query, categoryLabel),
        });
      },
    });
  },

  openManual() {
    go("/pages/dish/add-entry", { tab: "manual" });
  },

  openExtraction() {
    go("/pages/dish/add-entry", { tab: "extract" });
  },

  openCoverGenerator() {
    go("/pages/dish/add-entry", { tab: "manual", focus: "cover-generator" });
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
