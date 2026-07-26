const request = require("./request");
const auth = require("./auth");
const { uploadImage, normalizeDishPayload } = require("./upload");

const api = {
  bootstrap: async () => {
    const data = await request({ url: "/bootstrap" });
    auth.updateRuntimeConfig(data.runtimeConfig);
    return data;
  },
  getRuntimeConfig: async () => {
    const data = await request({ url: "/runtime-config" });
    auth.updateRuntimeConfig(data);
    return data;
  },
  getProfile: () => request({ url: "/profile" }),
  updateProfile: async (data) => {
    const payload = { nickname: data.nickname };
    if (data.avatarUrl && data.avatarUrl !== data.savedAvatarUrl) {
      const avatar = await uploadImage(data.avatarUrl, "profile_avatar");
      payload.avatarFileId = avatar.fileId;
    } else if (data.avatarFileId) {
      payload.avatarFileId = data.avatarFileId;
    }
    return request({ url: "/profile", method: "PUT", data: payload });
  },
  getMetadata: () => request({ url: "/metadata" }),

  listDishes: (data) => request({ url: "/dishes", data }),
  getDish: (id) => request({ url: `/dishes/${id}` }),
  createDish: async (data) => request({
    url: "/dishes",
    method: "POST",
    data: await normalizeDishPayload(data),
  }),
  updateDish: async (id, data) => request({
    url: `/dishes/${id}`,
    method: "PUT",
    data: await normalizeDishPayload(data),
  }),
  deleteDish: (id) => request({ url: `/dishes/${id}`, method: "DELETE" }),
  setDishDiscoverable: (id, discoverable) => request({
    url: `/dishes/${id}/discoverability`,
    method: "PUT",
    data: { discoverable },
  }),
  parseDish: async (data) => {
    const image = data.image
      ? await uploadImage(data.image, "dish_extract")
      : null;
    const payload = {};
    if (data.text) payload.text = data.text;
    if (image) payload.imageFileId = image.fileId;
    return request({ url: "/assist/dish-extraction", method: "POST", data: payload });
  },
  generateDishCover: (data) => request({
    url: "/assist/dish-covers",
    method: "POST",
    data,
  }),

  listRecommendations: (data) => request({ url: "/recommendations", data }),
  getRecommendation: (id) => request({ url: `/recommendations/${id}` }),
  copyRecommendation: (id) => request({
    url: `/recommendations/${id}/copy`,
    method: "POST",
  }),

  getWhatToEatStatus: () => request({ url: "/meal-suggestions/status" }),
  generateWhatToEat: (data) => request({
    url: "/meal-suggestions",
    method: "POST",
    data,
  }),
  getMealSuggestion: (id) => request({ url: `/meal-suggestions/${id}` }),
  saveSuggestedDish: (id, suggestionDishId) => request({
    url: `/meal-suggestions/${id}/copy`,
    method: "POST",
    data: { suggestionDishId },
  }),
  submitMealSuggestionFeedback: (id, action) => request({
    url: `/meal-suggestions/${id}/feedback`,
    method: "POST",
    data: { action },
  }),

  listRecipes: () => request({ url: "/recipes" }),
  getRecipe: (id) => request({ url: `/recipes/${id}` }),
  createRecipe: (data) => request({ url: "/recipes", method: "POST", data }),
  updateRecipe: (id, data) => request({
    url: `/recipes/${id}`,
    method: "PUT",
    data,
  }),
  deleteRecipe: (id) => request({ url: `/recipes/${id}`, method: "DELETE" }),
  addRecipeDishes: (id, dishIds) => request({
    url: `/recipes/${id}/dishes`,
    method: "POST",
    data: { dishIds },
  }),
  removeRecipeDish: (recipeId, dishId) => request({
    url: `/recipes/${recipeId}/dishes/${dishId}`,
    method: "DELETE",
  }),

  getCheckinCalendar: (month) => request({
    url: "/checkins/calendar",
    data: { month },
  }),
  listCheckins: (data) => request({ url: "/checkins", data }),
  createCheckin: async (data) => {
    const image = await uploadImage(data.image, "checkin");
    return request({
      url: "/checkins",
      method: "POST",
      data: {
        dishName: data.dishName,
        imageFileId: image.fileId,
        note: data.note,
      },
    });
  },

  getCurrentMeal: () => request({ url: "/meals/current" }),
  listMeals: (data) => request({ url: "/meals", data }),
  getMeal: (id) => request({ url: `/meals/${id}` }),
  createMeal: (data) => request({ url: "/meals", method: "POST", data }),
  previewMealByCode: (data) => request({
    url: "/meals/lookup",
    method: "POST",
    data,
  }),
  joinMeal: (data) => request({ url: "/meals/join", method: "POST", data }),
  recordMealFinalResultSubscription: (id, templateId, authorizationResult) => request({
    url: `/meals/${id}/final-result-subscriptions`,
    method: "POST",
    data: { templateId, authorizationResult },
  }),
  closeMeal: (id) => request({ url: `/meals/${id}/close`, method: "POST" }),
  cancelMeal: (id, data = {}) => request({
    url: `/meals/${id}/cancel`,
    method: "POST",
    data,
  }),
  listMealCandidates: (id, data) => request({
    url: `/meals/${id}/candidates`,
    data,
  }),
  removeMealCandidate: (mealId, candidateId) => request({
    url: `/meals/${mealId}/candidates/${candidateId}`,
    method: "DELETE",
  }),
  getMyMealVotes: (id) => request({ url: `/meals/${id}/votes/me` }),
  saveVotes: (id, candidateIds) => request({
    url: `/meals/${id}/votes/me`,
    method: "PUT",
    data: { candidateIds },
  }),
  getMealStats: (id) => request({ url: `/meals/${id}/stats` }),
  confirmMeal: (id, data) => request({
    url: `/meals/${id}/confirm`,
    method: "POST",
    data,
  }),
  completeMeal: (id) => request({
    url: `/meals/${id}/complete`,
    method: "POST",
  }),

  getShoppingList: (options = {}) => request({
    url: "/shopping-lists/current",
    ...options,
  }),
  getShoppingListById: (id) => request({ url: `/shopping-lists/${id}` }),
  getSharedShoppingList: (shareToken) => request({
    url: `/public/shopping-lists/${encodeURIComponent(shareToken)}`,
    auth: false,
    showError: false,
  }),
  addShoppingItem: (data) => request({
    url: "/shopping-lists/current/items",
    method: "POST",
    data,
  }),
  updateShoppingItem: (id, data) => request({
    url: `/shopping-lists/current/items/${id}`,
    method: "PUT",
    data,
  }),
  deleteShoppingItem: (id) => request({
    url: `/shopping-lists/current/items/${id}`,
    method: "DELETE",
  }),
  exportShoppingListText: (id) => request({
    url: `/shopping-lists/${id}/export-text`,
  }),

  getPrepPlanQuote: (mealId) => request({
    url: "/prep-plans/quote",
    data: { mealId },
  }),
  generatePrepPlan: (data) => request({
    url: "/prep-plans",
    method: "POST",
    data,
  }),
  getPrepPlan: (id) => request({ url: `/prep-plans/${id}` }),
  submitPrepPlanFeedback: (id, action) => request({
    url: `/prep-plans/${id}/feedback`,
    method: "POST",
    data: { action },
  }),

  getPointsSummary: () => request({ url: "/points/summary" }),
  listPointEntries: (data) => request({ url: "/points/entries", data }),
  listFeatureUsages: (data) => request({ url: "/feature-usages", data }),
  getNotificationSummary: () => request({ url: "/notifications/summary" }),
  listNotifications: (data) => request({ url: "/notifications", data }),
  readNotification: (id) => request({
    url: `/notifications/${id}/read`,
    method: "POST",
  }),
  readAllNotifications: () => request({
    url: "/notifications/read-all",
    method: "POST",
  }),
  uploadImage,
};

module.exports = api;
