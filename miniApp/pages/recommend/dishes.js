const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    categories: ["全部", "主食", "家常菜", "素菜", "汤菜"],
    category: "全部",
    list: [],
    page: 1,
    pageSize: 20,
    total: 0,
    loading: false,
  },

  onLoad() {
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
    this.setData({ loading: true });
    try {
      const result = await api.listRecommendations({ category: this.data.category, page, pageSize: this.data.pageSize });
      const next = result.list.map((item) => ({
        ...item,
        id: item.recommendationId,
      }));
      this.setData({
        list: reset ? next : [...this.data.list, ...next],
        page: result.page,
        total: result.total,
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  selectCategory(event) {
    this.setData({ category: event.currentTarget.dataset.value }, () => this.load(true));
  },

  open(event) {
    go("/pages/recommend/dish-detail", { id: event.detail.id });
  },
});
