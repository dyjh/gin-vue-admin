const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    categories: ["全部分类"],
    categoryIndex: 0,
    category: "",
    list: [],
    page: 1,
    pageSize: 20,
    total: 0,
    loading: false,
    loadFailed: false,
  },

  onLoad() {
    this.loadCategories();
    this.load(true);
  },

  onPullDownRefresh() {
    this.load(true).finally(() => wx.stopPullDownRefresh());
  },

  onReachBottom() {
    if (this.data.list.length < this.data.total) this.load(false);
  },

  async load(reset) {
    if (this.data.loading) return;
    const page = reset ? 1 : this.data.page + 1;
    const category = this.data.category;
    this.setData({ loading: true, loadFailed: false });
    try {
      const result = await api.listRecommendations({
        category,
        page,
        pageSize: this.data.pageSize,
      });
      const list = Array.isArray(result && result.list) ? result.list : [];
      const next = list.map((item) => ({
        ...item,
        id: item.recommendationId,
      }));
      this.setData({
        list: reset ? next : [...this.data.list, ...next],
        page: Number(result && result.page) || page,
        total: Number(result && result.total) || 0,
      });
    } catch (error) {
      this.setData({ loadFailed: true });
    } finally {
      this.setData({ loading: false });
    }
  },

  async loadCategories() {
    try {
      const metadata = await api.getMetadata();
      const names = (metadata.dishCategories || [])
        .filter((item) => item && item.enabled !== false && item.name)
        .sort((left, right) => Number(left.sortOrder) - Number(right.sortOrder))
        .map((item) => item.name);
      this.setData({
        categories: ["全部分类", ...Array.from(new Set(names))],
      });
    } catch (error) {
      this.setData({ categories: ["全部分类"] });
    }
  },

  selectCategory(event) {
    const categoryIndex = Number(event.detail.value) || 0;
    const category = categoryIndex === 0 ? "" : this.data.categories[categoryIndex];
    this.setData({ categoryIndex, category }, () => this.load(true));
  },

  open(event) {
    go("/pages/recommend/dish-detail", { id: event.detail.id });
  },
});
