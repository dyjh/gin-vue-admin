const api = require("../../services/api");
const { go, home } = require("../../utils/navigation");

Page({
  data: {
    id: "",
    dish: null,
  },

  onLoad(options) {
    this.setData({ id: options.id || "dish-1" });
  },

  onShow() {
    this.load();
  },

  async load() {
    const dish = await api.getDish(this.data.id);
    this.setData({ dish });
  },

  edit() {
    go("/pages/dish/edit", { id: this.data.id });
  },

  async toggleDiscoverable(event) {
    const discoverable = event.detail.value;
    try {
      const dish = await api.setDishDiscoverable(this.data.id, discoverable);
      this.setData({ dish });
    } catch (error) {
      this.setData({ "dish.discoverable": false });
    }
  },

  remove() {
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
});
