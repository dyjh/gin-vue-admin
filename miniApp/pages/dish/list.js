const api = require("../../services/api");
const { go } = require("../../utils/navigation");

const categories = [
  { label: "全部", value: "" },
  { label: "主食", value: "主食" },
  { label: "家常菜", value: "家常菜" },
  { label: "素菜", value: "素菜" },
  { label: "汤菜", value: "汤菜" },
];

Page({
  data: {
    categories,
    activeCategory: "",
    keyword: "",
    list: [],
    page: 1,
    pageSize: 20,
    total: 0,
    libraryTotal: 0,
    loading: false,
    loaded: false,
  },

  onShow() {
    this.load(true);
  },

  onUnload() {
    clearTimeout(this.searchTimer);
  },

  loadMore() {
    if (!this.data.loading && this.data.list.length < this.data.total) {
      this.load(false);
    }
  },

  async load(reset) {
    if (!reset && this.data.loading) return;
    const requestId = (this.requestId || 0) + 1;
    this.requestId = requestId;
    const page = reset ? 1 : this.data.page + 1;
    this.setData({ loading: true, ...(reset ? { loaded: false } : {}) });

    try {
      const result = await api.listDishes({
        page,
        pageSize: this.data.pageSize,
        q: this.data.keyword.trim(),
        category: this.data.activeCategory,
      });
      if (requestId !== this.requestId) return;
      const incoming = result.list;
      const combined = reset ? incoming : [...this.data.list, ...incoming];
      const unique = Array.from(new Map(combined.map((item) => [item.id, item])).values());

      this.setData({
        list: unique,
        page: result.page,
        total: result.total,
        libraryTotal: !this.data.activeCategory && !this.data.keyword.trim() ? result.total : this.data.libraryTotal,
        loaded: true,
      });
    } catch (error) {
      if (requestId !== this.requestId) return;
      this.setData({ loaded: true });
      wx.showToast({ title: "菜品加载失败", icon: "none" });
    } finally {
      if (requestId === this.requestId) this.setData({ loading: false });
    }
  },

  selectCategory(event) {
    const activeCategory = event.currentTarget.dataset.category;
    if (activeCategory === this.data.activeCategory) return;
    this.setData({ activeCategory, list: [], total: 0 });
    this.load(true);
  },

  inputSearch(event) {
    const keyword = event.detail.value;
    this.setData({ keyword });
    clearTimeout(this.searchTimer);
    this.searchTimer = setTimeout(() => this.load(true), 300);
  },

  confirmSearch() {
    clearTimeout(this.searchTimer);
    this.load(true);
  },

  clearSearch() {
    clearTimeout(this.searchTimer);
    this.setData({ keyword: "" });
    this.load(true);
  },

  openDish(event) {
    go("/pages/dish/detail", { id: event.detail.id });
  },

  addDish() {
    go("/pages/dish/add-entry");
  },
});
