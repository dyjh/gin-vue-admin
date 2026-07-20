const api = require("../../services/api");
const { go } = require("../../utils/navigation");

function filterRecipes(recipes, query) {
  const keyword = String(query || "").trim().toLowerCase();
  if (!keyword) return recipes;

  return recipes.filter((recipe) => {
    const dishNames = (recipe.dishes || []).map((dish) => dish.name).join(" ");
    return [recipe.name, recipe.note, dishNames]
      .filter(Boolean)
      .join(" ")
      .toLowerCase()
      .includes(keyword);
  });
}

Page({
  data: {
    recipes: [],
    filteredRecipes: [],
    query: "",
    showCreator: false,
    draftName: "",
    draftNote: "",
    canCreate: false,
    creating: false,
  },

  onShow() {
    this.load();
  },

  async load() {
    const response = await api.listRecipes();
    const recipes = response.map((recipe) => ({
      ...recipe,
      coverUrl:
        (recipe.dishes && recipe.dishes[0] && recipe.dishes[0].image)
        || recipe.coverUrl
        || (recipe.coverImages && recipe.coverImages[0])
        || (recipe.coverUrls && recipe.coverUrls[0])
        || "",
    }));
    this.setData({
      recipes,
      filteredRecipes: filterRecipes(recipes, this.data.query),
    });
  },

  search(event) {
    const query = event.detail.value;
    this.setData({
      query,
      filteredRecipes: filterRecipes(this.data.recipes, query),
    });
  },

  clearSearch() {
    this.setData({
      query: "",
      filteredRecipes: this.data.recipes,
    });
  },

  create() {
    this.setData({
      showCreator: true,
      draftName: "",
      draftNote: "",
      canCreate: false,
      creating: false,
    });
  },

  closeCreator() {
    if (this.data.creating) return;
    this.setData({ showCreator: false });
  },

  creatorNameInput(event) {
    const draftName = event.detail.value;
    this.setData({
      draftName,
      canCreate: Boolean(draftName.trim()),
    });
  },

  creatorNoteInput(event) {
    this.setData({ draftNote: event.detail.value });
  },

  async submitCreate() {
    const name = this.data.draftName.trim();
    const note = this.data.draftNote.trim();
    if (!name || this.data.creating) return;

    this.setData({ creating: true });
    try {
      const recipe = await api.createRecipe({ name, note });
      this.setData({
        showCreator: false,
        creating: false,
      });
      go("/pages/recipes/detail", { id: recipe.id });
    } catch (error) {
      this.setData({ creating: false });
    }
  },

  open(event) {
    go("/pages/recipes/detail", { id: event.currentTarget.dataset.id });
  },

  noop() {},
});