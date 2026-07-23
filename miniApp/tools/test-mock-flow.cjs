const assert = require("assert");
const storage = new Map();

global.wx = {
  getStorageSync(key) {
    return storage.get(key);
  },
  setStorageSync(key, value) {
    storage.set(key, value);
  },
  showToast() {},
};

const store = require("../mock/store");
store.reset();
const api = require("../services/api");

async function main() {
  const bootstrap = await api.bootstrap();
  assert.strictEqual(bootstrap.dishes.length, 3);
  assert.strictEqual(bootstrap.recommendations.length, 5);

  const firstPage = await api.listDishes({ page: 1, pageSize: 2 });
  assert.strictEqual(firstPage.pageSize, 2);
  assert.strictEqual(firstPage.list.length, 2);

  const created = await api.createDish({
    name: "测试炖豆腐",
    category: "家常菜",
    tags: ["下饭"],
    serving: 2,
    description: "Mock 流程测试菜品",
    image: "/assets/images/recommended-tomato-egg.png",
    ingredients: [{ id: "ingredient-test", name: "豆腐", amount: "1 盒" }],
    steps: [{ id: "step-test", text: "豆腐切块后炖煮。", image: "" }],
    status: "usable",
  });
  assert.ok(created.id);
  const updated = await api.updateDish(created.id, { ...created, description: "已更新" });
  assert.strictEqual(updated.description, "已更新");
  const discoverable = await api.setDishDiscoverable(created.id, true);
  assert.strictEqual(discoverable.discoverable, true);

  const copied = await api.copyRecommendation("recommend-1");
  assert.strictEqual(copied.sourceLocked, true);
  await assert.rejects(() => api.setDishDiscoverable(copied.id, true));

  const recipes = await api.listRecipes();
  const recipe = await api.updateRecipe(recipes[0].id, { dishIds: [created.id, copied.id] });
  assert.strictEqual(recipe.dishCount, 2);
  assert.strictEqual(recipe.coverUrl, created.image);
  assert.strictEqual(Object.prototype.hasOwnProperty.call(recipe, "coverImages"), false);

  const pointsBefore = await api.getPointsSummary();
  const checkin = await api.createCheckin({
    dishId: created.id,
    dishName: created.name,
    image: "/assets/images/recommended-tomato-egg.png",
    note: "测试打卡",
  });
  assert.strictEqual(checkin.rewardPoints, 5);
  const pointsAfter = await api.getPointsSummary();
  assert.strictEqual(pointsAfter.balance, pointsBefore.balance + 5);

  const meal = await api.createMeal({
    name: "Mock 饭局",
    deadline: "明天 20:00",
    candidateIds: ["dish-1", "dish-2"],
  });
  assert.strictEqual(meal.candidates.length, 2);
  const previewMeal = await api.previewMealByCode({ code: "000000" });
  assert.strictEqual(previewMeal.id, meal.id);
  assert.strictEqual(previewMeal.name, meal.name);

  const joinedMeal = await api.joinMeal({ code: "000000" });
  assert.strictEqual(joinedMeal.id, meal.id);
  await api.saveVotes(meal.id, ["dish-1"]);
  const closed = await api.closeMeal(meal.id);
  assert.strictEqual(closed.status, "closed");
  const stats = await api.getMealStats(meal.id);
  assert.strictEqual(stats.dishes.length, 2);
  const confirmed = await api.confirmMeal(meal.id, { dishes: [{ dishId: "dish-1", servings: 2 }] });
  assert.strictEqual(confirmed.shoppingListCreated, true);

  const historyFirstPage = await api.listMeals({ scope: "history", page: 1, pageSize: 5 });
  const historySecondPage = await api.listMeals({ scope: "history", page: 2, pageSize: 5 });
  assert.strictEqual(historyFirstPage.total, 12);
  assert.strictEqual(historyFirstPage.list.length, 5);
  assert.strictEqual(historySecondPage.list.length, 5);
  assert.ok(historyFirstPage.list[0].finalDishes.length > 0);
  assert.ok(historyFirstPage.list[0].coverImage);

  const shopping = await api.getShoppingList();
  assert.ok(Array.isArray(shopping.items));
  assert.ok(shopping.shareToken);
  const sharedShopping = await api.getSharedShoppingList(shopping.shareToken);
  assert.strictEqual(sharedShopping.readOnly, true);
  assert.strictEqual(sharedShopping.id, shopping.id);
  const shoppingItem = await api.addShoppingItem({ name: "测试采购项", amount: "1 份", note: "" });
  const checkedItem = await api.updateShoppingItem(shoppingItem.id, { completed: true });
  assert.strictEqual(checkedItem.completed, true);
  await api.deleteShoppingItem(shoppingItem.id);

  const status = await api.getWhatToEatStatus();
  assert.strictEqual(status.unlocked, true);
  const recommendation = await api.generateWhatToEat({ tags: ["下饭"], people: 2, useFreeQuota: true });
  assert.ok(recommendation.dish.name);
  assert.strictEqual(recommendation.dishes.length, 2);
  const singleRecommendation = await api.generateWhatToEat({ tags: ["下饭"], people: 1, useFreeQuota: true });
  assert.strictEqual(singleRecommendation.dishes.length, 1);
  const groupRecommendation = await api.generateWhatToEat({ tags: ["下饭"], people: 6, useFreeQuota: true });
  assert.strictEqual(groupRecommendation.dishes.length, 4);
  const savedAiDish = await api.saveAiDish(recommendation.id);
  assert.strictEqual(savedAiDish.sourceLocked, true);
  const prepPlan = await api.generatePrepPlan({ mealId: meal.id });
  assert.strictEqual(prepPlan.cost, 8);
  assert.strictEqual(prepPlan.steps.length, 4);
  assert.ok(prepPlan.steps.every((step) => step.detail && step.parallel && step.done));

  const pointEntries = await api.listPointEntries({ type: "all", page: 1, pageSize: 20 });
  assert.strictEqual(pointEntries.pageSize, 20);
  assert.ok(pointEntries.total > pointEntries.list.length);

  const notificationSummary = await api.getNotificationSummary();
  assert.ok(notificationSummary.unreadCount > 0);
  const unread = await api.listNotifications({ scope: "unread", page: 1, pageSize: 20 });
  assert.ok(unread.list.every((item) => !item.read));
  await api.readNotification(unread.list[0].id);
  await api.readAllNotifications();
  const afterRead = await api.getNotificationSummary();
  assert.strictEqual(afterRead.unreadCount, 0);

  await api.deleteDish(created.id);
  console.log(JSON.stringify({
    bootstrapDishes: bootstrap.dishes.length,
    createdDish: created.id,
    copiedRecommendation: copied.id,
    recipeDishes: recipe.dishCount,
    mealHistoryTotal: historyFirstPage.total,
    pointPageSize: pointEntries.pageSize,
    notificationUnreadAfterReadAll: afterRead.unreadCount,
  }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
