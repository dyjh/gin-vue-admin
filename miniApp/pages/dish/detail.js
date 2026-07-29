const api = require("../../services/api");
const { go, home } = require("../../utils/navigation");

Page({
  data: {
    id: "",
    dish: null,
    togglingDiscoverable: false,
  },

  onLoad(options) {
    this.setData({ id: options.id || "dish-1" });
    wx.hideShareMenu();
  },

  onShow() {
    this.load();
  },

  async load() {
    const dish = await api.getDish(this.data.id);
    this.setData({ dish });
    this.syncShareMenu(dish);
  },

  syncShareMenu(dish) {
    if (dish?.discoverable) {
      wx.showShareMenu({ menus: ["shareAppMessage", "shareTimeline"] });
      return;
    }
    wx.hideShareMenu();
  },

  edit() {
    if (!this.data.dish?.ownedByMe) return;
    go("/pages/dish/edit", { id: this.data.id });
  },

  explainDiscoverabilityLock() {
    wx.showModal({
      title: "来源保护",
      content: "这道菜由平台推荐或他人菜品加入，为保护原始来源，不能再次设为允许被发现。你仍可正常编辑并在自己的菜品库中使用。",
      showCancel: false,
      confirmText: "知道了",
    });
  },

  async toggleDiscoverable(event) {
    if (this.data.togglingDiscoverable || !this.data.dish?.ownedByMe) return;
    const discoverable = event.detail.value;
    const previous = this.data.dish.discoverable;
    this.setData({
      togglingDiscoverable: true,
      "dish.discoverable": discoverable,
    });
    try {
      const dish = await api.setDishDiscoverable(this.data.id, discoverable);
      this.setData({ dish });
      this.syncShareMenu(dish);
    } catch (error) {
      this.setData({ "dish.discoverable": previous });
      this.syncShareMenu({ ...this.data.dish, discoverable: previous });
    } finally {
      this.setData({ togglingDiscoverable: false });
    }
  },

  remove() {
    if (!this.data.dish?.ownedByMe) return;
    wx.showModal({
      title: "删除这道菜？",
      content: "删除后不会继续出现在菜谱和饭局候选中。",
      confirmText: "删除",
      confirmColor: "#D8564F",
      success: async ({ confirm }) => {
        if (!confirm) return;
        await api.deleteDish(this.data.id);
        wx.showToast({ title: "菜品已删除", icon: "success" });
        setTimeout(home, 300);
      },
    });
  },

  onShareAppMessage() {
    const dish = this.data.dish;
    if (!dish?.discoverable) return {};
    return {
      title: dish.name,
      path: `/pages/dish/detail?id=${encodeURIComponent(dish.id)}`,
      imageUrl: dish.coverUrl,
    };
  },

  onShareTimeline() {
    const dish = this.data.dish;
    if (!dish?.discoverable) return {};
    return {
      title: dish.name,
      query: `id=${encodeURIComponent(dish.id)}`,
      imageUrl: dish.coverUrl,
    };
  },
});