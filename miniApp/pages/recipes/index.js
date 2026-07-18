const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    recipes: [],
  },

  onShow() {
    this.load();
  },

  async load() {
    const recipes = await api.listRecipes();
    this.setData({ recipes });
  },

  create() {
    wx.showModal({
      title: "创建菜谱",
      editable: true,
      placeholderText: "例如：工作日晚餐",
      confirmText: "创建",
      confirmColor: "#159B55",
      success: async ({ confirm, content }) => {
        if (!confirm) return;
        const recipe = await api.createRecipe({ name: content.trim() || "新菜谱" });
        go("/pages/recipes/detail", { id: recipe.id });
      },
    });
  },

  open(event) {
    go("/pages/recipes/detail", { id: event.currentTarget.dataset.id });
  },
});
