const request = require("./request");
const { uploadImage, normalizeDishPayload } = require("./upload");

const api = {
  bootstrap: () => request({ url: "/bootstrap" }),
  updateProfile: async (data) => {
    const payload = { ...data };
    if (data.avatarUrl) {
      const avatar = await uploadImage(data.avatarUrl, "profile_avatar");
      payload.avatarFileId = avatar.fileId;
      payload.avatarUrl = avatar.url;
    }
    return request({ url: "/profile", method: "PUT", data: payload });
  },

  listDishes: (data) => request({ url: "/dishes", data }),
  getDish: (id) => request({ url: `/dishes/${id}` }),
  createDish: async (data) => request({ url: "/dishes", method: "POST", data: await normalizeDishPayload(data) }),
  updateDish: async (id, data) => request({ url: `/dishes/${id}`, method: "PUT", data: await normalizeDishPayload(data) }),
  deleteDish: (id) => request({ url: `/dishes/${id}`, method: "DELETE" }),
  setDishDiscoverable: (id, discoverable) => request({ url: `/dishes/${id}/discoverable`, method: "PUT", data: { discoverable } }),
  parseDish: async (data) => {
    const image = data.image ? await uploadImage(data.image, "dish_recognition") : null;
    return request({ url: "/ai/dish-recognition", method: "POST", data: { ...data, imageFileId: image ? image.fileId : "" } });
  },
  generateDishCover: (data) => request({ url: "/ai/dish-cover", method: "POST", data }),

  listRecommendations: (data) => request({ url: "/recommendations", data }),
  getRecommendation: (id) => request({ url: `/recommendations/${id}` }),
  copyRecommendation: (id) => request({ url: `/recommendations/${id}/copy`, method: "POST" }),

  listRecipes: () => request({ url: "/recipes" }),
  getRecipe: (id) => request({ url: `/recipes/${id}` }),
  createRecipe: (data) => request({ url: "/recipes", method: "POST", data }),
  updateRecipe: (id, data) => request({ url: `/recipes/${id}`, method: "PUT", data }),
  deleteRecipe: (id) => request({ url: `/recipes/${id}`, method: "DELETE" }),

  listCheckins: (data) => request({ url: "/checkins", data }),
  createCheckin: async (data) => {
    const image = await uploadImage(data.image, "checkin");
    return request({ url: "/checkins", method: "POST", data: { ...data, imageFileId: image.fileId, image: image.url } });
  },

  getCurrentMeal: () => request({ url: "/meals/current" }),
  listMeals: (data) => request({ url: "/meals", data }),
  getMeal: (id) => request({ url: `/meals/${id}` }),
  createMeal: (data) => request({ url: "/meals", method: "POST", data }),
  previewMealByCode: (data) => request({ url: "/meals/lookup", method: "POST", data }),
  joinMeal: (data) => request({ url: "/meals/join", method: "POST", data }),
  saveVotes: (id, dishIds) => request({ url: `/meals/${id}/votes/me`, method: "PUT", data: { dishIds } }),
  closeMeal: (id) => request({ url: `/meals/${id}/close`, method: "POST" }),
  getMealStats: (id) => request({ url: `/meals/${id}/stats` }),
  confirmMeal: (id, data) => request({ url: `/meals/${id}/confirm`, method: "POST", data }),

  getShoppingList: () => request({ url: "/shopping-lists/current" }),
  getSharedShoppingList: (shareToken) => request({ url: `/shopping-lists/shared/${encodeURIComponent(shareToken)}`, showError: false }),
  addShoppingItem: (data) => request({ url: "/shopping-lists/current/items", method: "POST", data }),
  updateShoppingItem: (id, data) => request({ url: `/shopping-lists/current/items/${id}`, method: "PUT", data }),
  deleteShoppingItem: (id) => request({ url: `/shopping-lists/current/items/${id}`, method: "DELETE" }),

  getWhatToEatStatus: () => request({ url: "/ai/what-to-eat/status" }),
  generateWhatToEat: (data) => request({ url: "/ai/what-to-eat", method: "POST", data }),
  saveAiDish: (id) => request({ url: `/ai/recommendations/${id}/save`, method: "POST" }),
  generatePrepPlan: (data) => request({ url: "/ai/prep-plans", method: "POST", data }),

  getPointsSummary: () => request({ url: "/points/summary" }),
  listPointEntries: (data) => request({ url: "/points/entries", data }),
  uploadImage,
  getNotificationSummary: () => request({ url: "/notifications/summary" }),
  listNotifications: (data) => request({ url: "/notifications", data }),
  readNotification: (id) => request({ url: `/notifications/${id}/read`, method: "POST" }),
  readAllNotifications: () => request({ url: "/notifications/read-all", method: "POST" }),
};

module.exports = api;
