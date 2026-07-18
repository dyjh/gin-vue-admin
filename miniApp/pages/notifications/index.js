const api = require("../../services/api");

Page({
  data: {
    scope: "unread",
    unreadCount: 0,
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
    const summary = await api.getNotificationSummary();
    this.setData({ unreadCount: summary.unreadCount, list: [], page: 0, total: 0 });
    await this.loadMore();
  },

  async loadMore() {
    if (this.data.loading) return;
    this.setData({ loading: true });
    try {
      const result = await api.listNotifications({ scope: this.data.scope, page: this.data.page + 1, pageSize: this.data.pageSize });
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

  setScope(event) {
    const scope = event.currentTarget.dataset.scope;
    if (scope === this.data.scope) return;
    this.setData({ scope, list: [], page: 0, total: 0 }, () => this.loadMore());
  },

  async openNotice(event) {
    const id = event.currentTarget.dataset.id;
    const notice = this.data.list.find((item) => item.id === id);
    if (!notice.read) {
      await api.readNotification(id);
      this.setData({
        unreadCount: Math.max(0, this.data.unreadCount - 1),
        list: this.data.scope === "unread"
          ? this.data.list.filter((item) => item.id !== id)
          : this.data.list.map((item) => item.id === id ? { ...item, read: true } : item),
      });
    }
    wx.showModal({ title: notice.title, content: notice.content, showCancel: false, confirmColor: "#159B55" });
  },

  async readAll() {
    if (!this.data.unreadCount) return;
    await api.readAllNotifications();
    this.setData({
      unreadCount: 0,
      list: this.data.scope === "unread" ? [] : this.data.list.map((item) => ({ ...item, read: true })),
      total: this.data.scope === "unread" ? 0 : this.data.total,
    });
    wx.showToast({ title: "已全部标为已读", icon: "success" });
  },
});
