const api = require("../../services/api");

const TYPE_PRESENTATION = {
  meal: { icon: "calendar", tone: "green" },
  points: { icon: "award", tone: "green" },
  refund: { icon: "award", tone: "blue", iconSize: 20 },
  discover: { icon: "eye-off", tone: "danger" },
  governance: { icon: "info", tone: "danger" },
};

function parseCreatedAt(value) {
  const source = String(value || "").trim();
  const relative = source.match(/^(今天|昨天)\s*(.*)$/);
  if (relative) return { groupLabel: relative[1], timeLabel: relative[2] || "刚刚" };

  if (/^(刚刚|\d+\s*分钟前)$/.test(source)) {
    return { groupLabel: "今天", timeLabel: source };
  }

  const dateTime = source.match(/^(?:\d{4}-)?(\d{2})-(\d{2})\s+(.+)$/);
  if (dateTime) {
    return {
      groupLabel: Number(dateTime[1]) + " 月 " + Number(dateTime[2]) + " 日",
      timeLabel: dateTime[3],
    };
  }

  return { groupLabel: "更早", timeLabel: source };
}

function splitContent(value) {
  const content = String(value || "").trim();
  for (const separator of ["，", "；"]) {
    const index = content.indexOf(separator);
    if (index > 0 && index < content.length - 1) {
      return {
        primaryContent: content.slice(0, index),
        secondaryContent: content.slice(index + 1).replace(/。$/, ""),
      };
    }
  }
  return { primaryContent: content, secondaryContent: "" };
}

function presentNotification(item) {
  const presentationType = TYPE_PRESENTATION[item.type] ? item.type : "governance";
  return {
    ...item,
    ...parseCreatedAt(item.createdAt),
    ...splitContent(item.content),
    presentationType,
    ...TYPE_PRESENTATION[presentationType],
  };
}

function groupNotifications(list) {
  return list.reduce((groups, item) => {
    let group = groups.find((entry) => entry.label === item.groupLabel);
    if (!group) {
      group = { label: item.groupLabel, items: [] };
      groups.push(group);
    }
    group.items.push(item);
    return groups;
  }, []);
}

Page({
  data: {
    scope: "unread",
    unreadCount: 0,
    list: [],
    groups: [],
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
    this.setData({ unreadCount: summary.unreadCount, list: [], groups: [], page: 0, total: 0 });
    await this.loadMore();
  },

  async loadMore() {
    if (this.data.loading) return;
    this.setData({ loading: true });
    try {
      const result = await api.listNotifications({ scope: this.data.scope, page: this.data.page + 1, pageSize: this.data.pageSize });
      const ids = new Set(this.data.list.map((item) => item.id));
      const next = result.list.filter((item) => !ids.has(item.id)).map(presentNotification);
      const list = [...this.data.list, ...next];
      this.setData({
        list,
        groups: groupNotifications(list),
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
    this.setData({ scope, list: [], groups: [], page: 0, total: 0 }, () => this.loadMore());
  },

  async openNotice(event) {
    const id = event.currentTarget.dataset.id;
    const notice = this.data.list.find((item) => item.id === id);
    if (!notice) return;
    if (!notice.read) {
      await api.readNotification(id);
      const list = this.data.scope === "unread"
        ? this.data.list.filter((item) => item.id !== id)
        : this.data.list.map((item) => item.id === id ? { ...item, read: true } : item);
      this.setData({
        unreadCount: Math.max(0, this.data.unreadCount - 1),
        list,
        groups: groupNotifications(list),
      });
    }
    wx.showModal({ title: notice.title, content: notice.content, showCancel: false, confirmColor: "#159B55" });
  },

  async readAll() {
    if (!this.data.unreadCount) return;
    await api.readAllNotifications();
    const list = this.data.scope === "unread" ? [] : this.data.list.map((item) => ({ ...item, read: true }));
    this.setData({
      unreadCount: 0,
      list,
      groups: groupNotifications(list),
      total: this.data.scope === "unread" ? 0 : this.data.total,
    });
    wx.showToast({ title: "已全部标为已读", icon: "success" });
  },
});
