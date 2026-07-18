const api = require("../../services/api");
const { go, home } = require("../../utils/navigation");

Page({
  data: {
    meal: null,
    closing: false,
  },

  onLoad() {
    this.load();
  },

  async load() {
    const meal = await api.getCurrentMeal();
    this.setData({ meal });
  },

  copyCode() {
    wx.setClipboardData({ data: this.data.meal.code });
  },

  openVote() {
    go("/pages/meal/vote");
  },

  closeMeal() {
    wx.showModal({
      title: "提前关闭点餐？",
      content: "关闭后参与者不能继续修改点菜，可以进入统计并确认菜单。",
      confirmText: "提前关闭",
      confirmColor: "#D8564F",
      success: async ({ confirm }) => {
        if (!confirm) return;
        this.setData({ closing: true });
        try {
          const meal = await api.closeMeal(this.data.meal.id);
          this.setData({ meal });
          wx.showToast({ title: "点餐已关闭", icon: "success" });
        } finally {
          this.setData({ closing: false });
        }
      },
    });
  },

  openStats() {
    go("/pages/meal/stats");
  },

  home,

  onShareAppMessage() {
    return {
      title: `来参加“${this.data.meal.name}”点菜`,
      path: `/pages/meal/vote?state=join&code=${this.data.meal.code}`,
    };
  },
});
