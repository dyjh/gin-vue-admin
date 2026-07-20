const api = require("../../services/api");
const { go, home } = require("../../utils/navigation");

function filterDishes(dishes, query, category) {
  const keyword = String(query || "").trim().toLowerCase();
  return dishes.filter((dish) => {
    const inCategory = category === "全部" || dish.category === category;
    const searchable = [dish.name, dish.category, dish.meta, ...(dish.tags || [])].join(" ").toLowerCase();
    return inCategory && (!keyword || searchable.includes(keyword));
  });
}

Page({
  data: {
    id: "",
    recipe: null,
    allDishes: [],
    filteredDishes: [],
    dishCategories: ["全部"],
    dishFilterQuery: "",
    dishFilterCategory: "全部",
    showAdd: false,
    selectedIds: [],
    swipeId: "",
    touchStartX: 0,
  },

  onLoad(options) {
    this.setData({ id: options.id || "recipe-1" });
  },

  onShow() {
    this.load();
  },

  async load() {
    const [recipe, result] = await Promise.all([
      api.getRecipe(this.data.id),
      api.listDishes({ page: 1, pageSize: 100, status: "usable" }),
    ]);
    const dishCategories = ["全部"];
    result.list.forEach((dish) => {
      if (dish.category && !dishCategories.includes(dish.category)) dishCategories.push(dish.category);
    });
    this.setData({
      recipe,
      allDishes: result.list,
      filteredDishes: result.list,
      dishCategories,
      selectedIds: recipe.dishIds,
    });
  },

  editName() {
    wx.showModal({
      title: "编辑菜谱",
      editable: true,
      placeholderText: "菜谱名称",
      content: this.data.recipe.name,
      confirmColor: "#159B55",
      success: async ({ confirm, content }) => {
        if (!confirm) return;
        const recipe = await api.updateRecipe(this.data.id, { name: content.trim() || this.data.recipe.name });
        this.setData({ recipe });
      },
    });
  },

  editNote(event) {
    this.setData({ "recipe.note": event.detail.value });
  },

  async saveNote() {
    const recipe = await api.updateRecipe(this.data.id, { note: this.data.recipe.note });
    this.setData({ recipe });
    wx.showToast({ title: "备注已保存", icon: "success" });
  },

  touchStart(event) {
    this.setData({ touchStartX: event.touches[0].clientX });
  },

  touchEnd(event) {
    const id = event.currentTarget.dataset.id;
    const distance = event.changedTouches[0].clientX - this.data.touchStartX;
    this.setData({ swipeId: distance < -35 ? id : "" });
  },

  async removeDish(event) {
    const id = event.currentTarget.dataset.id;
    const dishIds = this.data.recipe.dishIds.filter((dishId) => dishId !== id);
    const recipe = await api.updateRecipe(this.data.id, { dishIds });
    this.setData({ recipe, selectedIds: dishIds, swipeId: "" });
  },

  openDish(event) {
    go("/pages/dish/detail", { id: event.currentTarget.dataset.id });
  },

  openAdd() {
    this.setData({
      showAdd: true,
      selectedIds: [...this.data.recipe.dishIds],
      dishFilterQuery: "",
      dishFilterCategory: "全部",
      filteredDishes: this.data.allDishes,
    });
  },

  closeAdd() {
    this.setData({ showAdd: false });
  },

  noop() {},

  inputDishFilter(event) {
    const dishFilterQuery = event.detail.value;
    this.setData({
      dishFilterQuery,
      filteredDishes: filterDishes(this.data.allDishes, dishFilterQuery, this.data.dishFilterCategory),
    });
  },

  clearDishFilter() {
    this.setData({
      dishFilterQuery: "",
      filteredDishes: filterDishes(this.data.allDishes, "", this.data.dishFilterCategory),
    });
  },

  selectDishCategory(event) {
    const dishFilterCategory = event.currentTarget.dataset.category;
    this.setData({
      dishFilterCategory,
      filteredDishes: filterDishes(this.data.allDishes, this.data.dishFilterQuery, dishFilterCategory),
    });
  },

  toggleDish(event) {
    const id = event.detail.id;
    const selectedIds = [...this.data.selectedIds];
    const index = selectedIds.indexOf(id);
    if (index >= 0) selectedIds.splice(index, 1);
    else selectedIds.push(id);
    this.setData({ selectedIds });
  },

  async saveDishes() {
    const recipe = await api.updateRecipe(this.data.id, { dishIds: this.data.selectedIds });
    this.setData({ recipe, showAdd: false });
    wx.showToast({ title: "菜品已更新", icon: "success" });
  },

  useForMeal() {
    go("/pages/meal/create", { recipeId: this.data.id });
  },

  removeRecipe() {
    wx.showModal({
      title: "删除这份菜谱？",
      content: "菜品仍会保留在个人菜品库中。",
      confirmText: "删除",
      confirmColor: "#D8564F",
      success: async ({ confirm }) => {
        if (!confirm) return;
        await api.deleteRecipe(this.data.id);
        wx.showToast({ title: "菜谱已删除", icon: "success" });
        setTimeout(home, 300);
      },
    });
  },
});
