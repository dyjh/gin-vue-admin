const api = require("../../services/api");

Page({
  data: {
    summary: null,
    filters: [
      { key: "all", label: "全部" },
      { key: "earned", label: "获得" },
      { key: "spent", label: "消耗" },
      { key: "refund", label: "退还" }
    ],
    filter: "all",
    list: [],
    page: 0,
    pageSize: 20,
    total: 0,
    loading: false,
  },

  onLoad() {
    this.refresh();
  },

  onPullDownRefresh() {
    this.refresh().finally(() => wx.stopPullDownRefresh());
  },

  onReachBottom() {
    if (this.data.list.length < this.data.total) this.loadMore();
  },

  async refresh() {
    const summary = await api.getPointsSummary();
    this.setData({ summary, list: [], page: 0, total: 0 });
    await this.loadMore();
  },

  async loadMore() {
    if (this.data.loading) return;
    this.setData({ loading: true });
    try {
      const result = await api.listPointEntries({ type: this.data.filter, page: this.data.page + 1, pageSize: this.data.pageSize });
      const ids = new Set(this.data.list.map((item) => item.id));
      const next = result.list.filter((item) => !ids.has(item.id));
      this.setData({
        list: [...this.data.list, ...next],
        page: result.page,
        total: result.total,
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  setFilter(event) {
    const filter = event.currentTarget.dataset.filter;
    if (filter === this.data.filter) return;
    this.setData({ filter, list: [], page: 0, total: 0 }, () => this.loadMore());
  },
});
