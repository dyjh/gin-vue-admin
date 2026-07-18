const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    loading: true,
    query: "",
    dishes: [],
    recommendations: [],
    meal: null,
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
      this.setData({
        dishes: data.dishes,
        recommendations: data.recommendations,
        meal: data.meal,
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  search(event) {
    this.setData({ query: event.detail.value });
  },

  openAdd() {
    wx.showActionSheet({
      itemList: ["菜品录入", "AI 识别"],
      success: ({ tapIndex }) => {
        go("/pages/dish/add-entry", { tab: tapIndex === 1 ? "ai" : "manual" });
      },
    });
  },

  openDish(event) {
    go("/pages/dish/detail", { id: event.detail.id });
  },

  openRecommendation(event) {
    go("/pages/recommend/dish-detail", { id: event.detail.id });
  },

  openRecommendations() {
    go("/pages/recommend/dishes");
  },

  async copyRecommendation(event) {
    const dish = await api.copyRecommendation(event.detail.id);
    wx.showToast({ title: "已加入菜品库", icon: "success" });
    go("/pages/dish/detail", { id: dish.id });
  },

  openWhatToEat() {
    go("/pages/ai/what-to-eat");
  },
});
