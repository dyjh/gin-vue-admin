const { getState, mutate } = require("./store");
const { paginate } = require("../utils/format");

const SHOPPING_LIST_ID = "shopping-list-current";
const SHOPPING_SHARE_TOKEN = "sl_8f3c2a71d46e9b05";

function nextId(prefix) {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
}

function findById(list, id, label) {
  const item = list.find((entry) => entry.id === id);
  if (!item) throw new Error(`${label || "内容"}不存在`);
  return item;
}

function match(path, pattern) {
  const result = path.match(pattern);
  return result ? result.slice(1) : null;
}

function enrichRecipe(recipe, state) {
  const dishes = recipe.dishIds.map((id) => state.dishes.find((dish) => dish.id === id)).filter(Boolean);
  const { coverImages, ...summary } = recipe;
  return {
    ...summary,
    dishes,
    dishCount: dishes.length,
    coverUrl: dishes[0] ? dishes[0].image : "",
  };
}

function mealView(state) {
  return {
    ...state.meal,
    candidates: state.meal.candidateIds.map((id) => state.dishes.find((dish) => dish.id === id)).filter(Boolean),
  };
}

function mealHistoryView(meal, state) {
  const finalDishes = (meal.finalMenu || []).map((entry) => {
    const dish = state.dishes.find((item) => item.id === entry.dishId);
    return dish ? { ...dish, servings: entry.servings } : null;
  }).filter(Boolean);
  return {
    ...meal,
    finalDishes,
    finalDishCount: finalDishes.length,
    totalServings: finalDishes.reduce((total, dish) => total + dish.servings, 0),
    coverImage: finalDishes[0] ? finalDishes[0].image : "",
  };
}

function shoppingListView(state, options = {}) {
  const items = state.shoppingItems;
  return {
    id: SHOPPING_LIST_ID,
    meal: state.meal,
    items,
    pendingCount: items.filter((item) => !item.completed).length,
    completedCount: items.filter((item) => item.completed).length,
    ...(options.includeShareToken ? { shareToken: SHOPPING_SHARE_TOKEN } : {}),
    ...(options.readOnly ? { readOnly: true } : {}),
  };
}

function handleGet(url, data, state) {
  if (url === "/bootstrap") {
    return {
      profile: state.profile,
      dishes: state.dishes.slice(0, 3),
      recommendations: state.recommendations.slice(0, 5),
      meal: mealView(state),
      unreadCount: state.notifications.filter((item) => !item.read).length,
    };
  }

  if (url === "/dishes") {
    let list = state.dishes;
    if (data.q) {
      const keyword = String(data.q).trim().toLowerCase();
      list = list.filter((item) => {
        const searchable = [item.name, item.category, ...(item.tags || [])]
          .filter(Boolean)
          .join(" ")
          .toLowerCase();
        return searchable.includes(keyword);
      });
    }
    if (data.category && data.category !== "全部") list = list.filter((item) => item.category === data.category);
    if (data.status) list = list.filter((item) => item.status === data.status);
    return paginate(list, data.page || 1, data.pageSize || 20);
  }

  const dishMatch = match(url, /^\/dishes\/([^/]+)$/);
  if (dishMatch) return findById(state.dishes, dishMatch[0], "菜品");

  if (url === "/recommendations") {
    let list = state.recommendations;
    if (data.category && data.category !== "全部") list = list.filter((item) => item.category === data.category);
    return paginate(list, data.page || 1, data.pageSize || 20);
  }

  const recommendationMatch = match(url, /^\/recommendations\/([^/]+)$/);
  if (recommendationMatch) return findById(state.recommendations, recommendationMatch[0], "推荐菜品");

  if (url === "/recipes") return state.recipes.map((recipe) => enrichRecipe(recipe, state));
  const recipeMatch = match(url, /^\/recipes\/([^/]+)$/);
  if (recipeMatch) return enrichRecipe(findById(state.recipes, recipeMatch[0], "菜谱"), state);

  if (url === "/checkins") return paginate(state.checkins, data.page || 1, data.pageSize || 20);
  if (url === "/meals") {
    const list = data.scope === "history"
      ? (state.mealHistories || []).map((meal) => mealHistoryView(meal, state))
      : [mealView(state)];
    return paginate(list, data.page || 1, data.pageSize || 20);
  }
  if (url === "/meals/current") return mealView(state);

  const mealDetailMatch = match(url, /^\/meals\/([^/]+)$/);
  if (mealDetailMatch) {
    if (state.meal.id === mealDetailMatch[0]) return mealView(state);
    return mealHistoryView(findById(state.mealHistories || [], mealDetailMatch[0], "饭局"), state);
  }

  const statsMatch = match(url, /^\/meals\/([^/]+)\/stats$/);
  if (statsMatch) {
    return {
      meal: mealView(state),
      dishes: state.meal.candidateIds.map((id) => {
        const dish = findById(state.dishes, id, "菜品");
        return {
          ...dish,
          votes: state.meal.votes[id] || 0,
          suggestedServings: Math.max(1, Math.ceil((state.meal.votes[id] || 1) / dish.serving)),
          finalServings: Math.max(1, Math.ceil((state.meal.votes[id] || 1) / dish.serving)),
        };
      }),
    };
  }

  if (url === "/shopping-lists/current") {
    return shoppingListView(state, { includeShareToken: true });
  }

  const sharedShoppingMatch = match(url, /^\/shopping-lists\/shared\/([^/]+)$/);
  if (sharedShoppingMatch) {
    if (sharedShoppingMatch[0] !== SHOPPING_SHARE_TOKEN) throw new Error("分享链接无效或已失效");
    return shoppingListView(state, { readOnly: true });
  }

  if (url === "/ai/what-to-eat/status") {
    return {
      unlocked: state.profile.checkinDays >= 7,
      checkinDays: state.profile.checkinDays,
      unlockDays: 7,
      freeQuota: 2,
      pointBalance: state.profile.points,
      pointCost: 2,
    };
  }

  if (url === "/points/summary") {
    const earned = state.pointEntries.filter((entry) => entry.amount > 0 && entry.type !== "refund").reduce((sum, entry) => sum + entry.amount, 0);
    const spent = Math.abs(state.pointEntries.filter((entry) => entry.type === "spent").reduce((sum, entry) => sum + entry.amount, 0));
    return { balance: state.profile.points, earned, spent };
  }

  if (url === "/points/entries") {
    const list = data.type && data.type !== "all"
      ? state.pointEntries.filter((entry) => entry.type === data.type)
      : state.pointEntries;
    return paginate(list, data.page || 1, data.pageSize || 20);
  }

  if (url === "/notifications/summary") {
    return {
      unreadCount: state.notifications.filter((item) => !item.read).length,
      total: state.notifications.length,
    };
  }

  if (url === "/notifications") {
    const list = data.scope === "unread"
      ? state.notifications.filter((item) => !item.read)
      : state.notifications;
    return paginate(list, data.page || 1, data.pageSize || 20);
  }

  throw new Error(`Mock GET route not found: ${url}`);
}

function handlePost(url, data, state) {
  if (url === "/dishes") {
    const dish = {
      id: nextId("dish"),
      status: "draft",
      discoverable: false,
      sourceLocked: false,
      ingredients: [],
      steps: [],
      tags: [],
      ...data,
    };
    mutate((current) => {
      current.dishes.unshift(dish);
      return current;
    });
    return dish;
  }

  if (url === "/ai/dish-recognition") {
    return {
      name: "虾仁蒸蛋",
      category: "家常菜",
      tags: ["清淡", "适合孩子"],
      serving: 2,
      description: "口感细嫩，虾仁鲜甜，适合作为清淡的一餐。",
      ingredients: [
        { id: nextId("ingredient"), name: "鸡蛋", amount: "2 个" },
        { id: nextId("ingredient"), name: "虾仁", amount: "80 克" }
      ],
      steps: [
        { id: nextId("step"), text: "鸡蛋加温水打散并过筛。", image: "" },
        { id: nextId("step"), text: "放入虾仁，盖盘蒸约 10 分钟。", image: "" }
      ],
      image: "/assets/images/what-to-eat-kung-pao-chicken.jpg"
    };
  }

  if (url === "/ai/dish-cover") {
    return { image: "/assets/images/dish-detail-cover-realistic.jpg", cost: 6 };
  }

  const copyMatch = match(url, /^\/recommendations\/([^/]+)\/copy$/);
  if (copyMatch) {
    const source = findById(state.recommendations, copyMatch[0], "推荐菜品");
    const dish = {
      ...source,
      id: nextId("dish"),
      discoverable: false,
      sourceLocked: true,
      status: "usable",
      meta: `${source.category} · 已加入菜品库`,
    };
    delete dish.author;
    delete dish.copied;
    delete dish.sourceType;
    mutate((current) => {
      current.dishes.unshift(dish);
      const target = current.recommendations.find((item) => item.id === source.id);
      if (target) target.copied = true;
      return current;
    });
    return dish;
  }

  if (url === "/recipes") {
    const recipe = { id: nextId("recipe"), name: "新菜谱", note: "", dishIds: [], ...data };
    mutate((current) => {
      current.recipes.unshift(recipe);
      return current;
    });
    return enrichRecipe(recipe, getState());
  }

  if (url === "/checkins") {
    const checkin = { id: nextId("checkin"), date: new Date().toISOString().slice(0, 10), rewarded: true, ...data };
    mutate((current) => {
      current.checkins.unshift(checkin);
      current.profile.points += 5;
      current.profile.checkinDays += 1;
      current.pointEntries.unshift({
        id: nextId("point"),
        type: "earned",
        title: "每日首次做菜打卡",
        description: checkin.dishName || "完成做菜打卡",
        amount: 5,
        createdAt: "刚刚",
      });
      return current;
    });
    return { checkin, rewardPoints: 5 };
  }

  if (url === "/meals") {
    const meal = {
      id: nextId("meal"),
      name: "新的家庭饭局",
      code: String(Math.floor(100000 + Math.random() * 900000)),
      status: "collecting",
      deadline: "明天 20:00",
      participantCount: 1,
      selectedIds: [],
      votes: {},
      ...data,
    };
    mutate((current) => {
      current.meal = meal;
      return current;
    });
    return mealView(getState());
  }

  if (url === "/meals/lookup") {
    const validMockCode = /^\d{6}$/.test(String(data.code || ""));
    if (!validMockCode) return Promise.reject(new Error("没有找到这个饭局，请检查点餐码"));
    const meal = mealView(state);
    return {
      id: meal.id,
      name: meal.name,
      status: meal.status,
      deadlineAt: meal.deadlineAt || meal.deadline,
      participantCount: meal.participantCount,
      candidateCount: meal.candidateIds.length,
    };
  }

  if (url === "/meals/join") {
    const validMockCode = /^\d{6}$/.test(String(data.code || ""));
    return validMockCode ? mealView(state) : Promise.reject(new Error("没有找到这个饭局，请检查点餐码"));
  }

  const closeMatch = match(url, /^\/meals\/([^/]+)\/close$/);
  if (closeMatch) {
    mutate((current) => {
      current.meal.status = "closed";
      return current;
    });
    return mealView(getState());
  }

  const confirmMatch = match(url, /^\/meals\/([^/]+)\/confirm$/);
  if (confirmMatch) {
    mutate((current) => {
      current.meal.status = "confirmed";
      current.meal.finalMenu = data.dishes || [];
      return current;
    });
    return { meal: mealView(getState()), shoppingListCreated: true };
  }

  if (url === "/shopping-lists/current/items") {
    const item = { id: nextId("shop"), completed: false, note: "", ...data };
    mutate((current) => {
      current.shoppingItems.unshift(item);
      return current;
    });
    return item;
  }

  if (url === "/ai/what-to-eat") {
    const people = Math.min(Math.max(Number(data.people) || 1, 1), 20);
    const suggestedDishCount = people === 1
      ? 1
      : people === 2
        ? 2
        : Math.min(5, Math.ceil(people / 2) + 1);
    const source = state.recommendations.find((item) => item.name === "青椒牛柳") || state.recommendations[0];
    const primaryDish = {
      ...source,
      recommendationId: "ai-recommendation-1-dish-0",
      image: "/assets/images/what-to-eat-kung-pao-chicken.jpg",
      name: "宫保鸡丁",
      category: "川菜",
      tags: ["下饭", "快手", "微辣"],
      serving: people,
    };
    const companionNames = ["蒜蓉西兰花", "冬瓜丸子汤", "番茄炒蛋", "香菇鸡腿饭"];
    const companionDishes = companionNames
      .map((name) => state.recommendations.find((item) => item.name === name))
      .filter(Boolean)
      .slice(0, suggestedDishCount - 1)
      .map((dish, index) => ({
        ...dish,
        recommendationId: "ai-recommendation-1-dish-" + (index + 1),
        serving: people,
      }));
    const dishes = [primaryDish, ...companionDishes];
    return {
      id: "ai-recommendation-1",
      source: "ai",
      sourceLabel: "AI生成推荐",
      reason: "已结合口味条件和用餐人数安排菜品数量与搭配。",
      people,
      suggestedDishCount: dishes.length,
      dish: dishes[0],
      dishes,
      cost: data.useFreeQuota ? 0 : 2,
    };
  }

  const saveAiMatch = match(url, /^\/ai\/recommendations\/([^/]+)\/save$/);
  if (saveAiMatch) {
    const matchIndex = saveAiMatch[0].match(/-dish-(\d+)$/);
    const dishIndex = matchIndex ? Number(matchIndex[1]) : 0;
    const companionNames = ["蒜蓉西兰花", "冬瓜丸子汤", "番茄炒蛋", "香菇鸡腿饭"];
    const sourceName = dishIndex === 0 ? "青椒牛柳" : companionNames[dishIndex - 1];
    const source = state.dishes.find((item) => item.name === sourceName) || state.dishes[0];
    const dish = dishIndex === 0
      ? { ...source, id: nextId("dish"), name: "宫保鸡丁", discoverable: false, sourceLocked: true, image: "/assets/images/what-to-eat-kung-pao-chicken.jpg" }
      : { ...source, id: nextId("dish"), discoverable: false, sourceLocked: true };
    mutate((current) => {
      current.dishes.unshift(dish);
      return current;
    });
    return dish;
  }

  if (url === "/ai/prep-plans") {
    return {
      id: nextId("prep"),
      cost: 8,
      steps: [
        { id: "prep-1", title: "鸡腿肉先焯水", detail: "鸡腿肉冷水下锅，加姜片焯水", parallel: "泡香菇；冬瓜去皮切块", done: "水开撇净浮沫，捞出鸡腿肉" },
        { id: "prep-2", title: "接着焖鸡腿肉", detail: "鸡腿肉与香菇入锅，按菜谱开始焖煮", parallel: "西兰花分小朵洗净；蒜切末", done: "转小火后可离灶，进入下一步" },
        { id: "prep-3", title: "再煮冬瓜丸子汤", detail: "冬瓜汤煮开，丸子逐个下锅", parallel: "番茄切块；鸡蛋下锅前再打散", done: "丸子浮起后转小火保温" },
        { id: "prep-4", title: "开饭前炒快手菜", detail: "先焯西兰花快炒，再做番茄炒蛋", parallel: "检查鸡腿饭和汤的咸淡", done: "快手菜完成后立即上桌" }
      ]
    };
  }

  const readMatch = match(url, /^\/notifications\/([^/]+)\/read$/);
  if (readMatch) {
    mutate((current) => {
      const notice = current.notifications.find((item) => item.id === readMatch[0]);
      if (notice) notice.read = true;
      return current;
    });
    return { success: true };
  }

  if (url === "/notifications/read-all") {
    mutate((current) => {
      current.notifications.forEach((item) => { item.read = true; });
      return current;
    });
    return { success: true };
  }

  throw new Error(`Mock POST route not found: ${url}`);
}

function handlePut(url, data, state) {
  if (url === "/profile") {
    mutate((current) => {
      current.profile = { ...current.profile, ...data };
      return current;
    });
    return getState().profile;
  }

  const discoverMatch = match(url, /^\/dishes\/([^/]+)\/discoverable$/);
  if (discoverMatch) {
    mutate((current) => {
      const dish = findById(current.dishes, discoverMatch[0], "菜品");
      if (dish.sourceLocked && data.discoverable) throw new Error("从推荐加入的菜品不能设为允许被发现");
      dish.discoverable = Boolean(data.discoverable);
      return current;
    });
    return findById(getState().dishes, discoverMatch[0], "菜品");
  }

  const dishMatch = match(url, /^\/dishes\/([^/]+)$/);
  if (dishMatch) {
    mutate((current) => {
      const dish = findById(current.dishes, dishMatch[0], "菜品");
      Object.assign(dish, data, { id: dish.id });
      return current;
    });
    return findById(getState().dishes, dishMatch[0], "菜品");
  }

  const recipeMatch = match(url, /^\/recipes\/([^/]+)$/);
  if (recipeMatch) {
    mutate((current) => {
      const recipe = findById(current.recipes, recipeMatch[0], "菜谱");
      Object.assign(recipe, data, { id: recipe.id });
      return current;
    });
    return enrichRecipe(findById(getState().recipes, recipeMatch[0], "菜谱"), getState());
  }

  const voteMatch = match(url, /^\/meals\/([^/]+)\/votes\/me$/);
  if (voteMatch) {
    mutate((current) => {
      current.meal.selectedIds = data.dishIds || [];
      return current;
    });
    return mealView(getState());
  }

  const shoppingMatch = match(url, /^\/shopping-lists\/current\/items\/([^/]+)$/);
  if (shoppingMatch) {
    mutate((current) => {
      const item = findById(current.shoppingItems, shoppingMatch[0], "采购项");
      Object.assign(item, data, { id: item.id });
      return current;
    });
    return findById(getState().shoppingItems, shoppingMatch[0], "采购项");
  }

  throw new Error(`Mock PUT route not found: ${url}`);
}

function handleDelete(url, state) {
  const dishMatch = match(url, /^\/dishes\/([^/]+)$/);
  if (dishMatch) {
    mutate((current) => {
      current.dishes = current.dishes.filter((dish) => dish.id !== dishMatch[0]);
      current.recipes.forEach((recipe) => {
        recipe.dishIds = recipe.dishIds.filter((id) => id !== dishMatch[0]);
      });
      return current;
    });
    return { success: true };
  }

  const recipeMatch = match(url, /^\/recipes\/([^/]+)$/);
  if (recipeMatch) {
    mutate((current) => {
      current.recipes = current.recipes.filter((recipe) => recipe.id !== recipeMatch[0]);
      return current;
    });
    return { success: true };
  }

  const shoppingMatch = match(url, /^\/shopping-lists\/current\/items\/([^/]+)$/);
  if (shoppingMatch) {
    mutate((current) => {
      current.shoppingItems = current.shoppingItems.filter((item) => item.id !== shoppingMatch[0]);
      return current;
    });
    return { success: true };
  }

  throw new Error(`Mock DELETE route not found: ${url}`);
}

function handle(options) {
  const method = String(options.method || "GET").toUpperCase();
  const url = options.url;
  const data = options.data || {};
  const state = getState();

  return new Promise((resolve, reject) => {
    setTimeout(() => {
      try {
        let result;
        if (method === "GET") result = handleGet(url, data, state);
        else if (method === "POST") result = handlePost(url, data, state);
        else if (method === "PUT") result = handlePut(url, data, state);
        else if (method === "DELETE") result = handleDelete(url, state);
        else throw new Error(`Mock method not supported: ${method}`);

        if (result && typeof result.then === "function") {
          result.then(resolve).catch(reject);
          return;
        }
        resolve(result);
      } catch (error) {
        if (options.showError !== false) wx.showToast({ title: error.message || "操作失败", icon: "none" });
        reject(error);
      }
    }, 120);
  });
}

module.exports = { handle };
