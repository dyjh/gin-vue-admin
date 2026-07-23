const api = require("../../services/api");

function decodeShareToken(value = "") {
  try {
    return decodeURIComponent(value);
  } catch (error) {
    return value;
  }
}

Page({
  data: {
    shareToken: "",
    loading: true,
    error: "",
    meal: null,
    items: [],
    pendingItems: [],
    completedItems: [],
    pendingCount: 0,
    completedCount: 0,
    progress: 0,
  },

  onLoad(options = {}) {
    if (wx.hideShareMenu) {
      wx.hideShareMenu({ menus: ["shareAppMessage", "shareTimeline"] });
    }

    const shareToken = decodeShareToken(options.token || "").trim();
    if (!shareToken) {
      this.setData({ loading: false, error: "分享链接缺少必要信息" });
      return;
    }

    this.setData({ shareToken });
    this.load();
  },

  async onPullDownRefresh() {
    await this.load();
    wx.stopPullDownRefresh();
  },

  async load() {
    try {
      const data = await api.getSharedShoppingList(this.data.shareToken);
      const items = Array.isArray(data.items) ? data.items : [];
      const pendingItems = items.filter((item) => !item.completed);
      const completedItems = items.filter((item) => item.completed);
      const completedCount = completedItems.length;

      this.setData({
        ...data,
        items,
        pendingItems,
        completedItems,
        pendingCount: pendingItems.length,
        completedCount,
        progress: items.length ? Math.round((completedCount / items.length) * 100) : 0,
        loading: false,
        error: "",
      });
    } catch (error) {
      this.setData({
        loading: false,
        error: error.message || "分享清单暂时无法查看",
      });
    }
  },
});