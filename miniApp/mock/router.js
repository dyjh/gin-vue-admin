const { getState, mutate } = require("./store");
const { paginate } = require("../utils/format");

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
  return {
    ...recipe,
    dishes: recipe.dishIds.map((id) => state.dishes.find((dish) => dish.id === id)).filter(Boolean),
    dishCount: recipe.dishIds.length,
  };
}

function mealView(state) {
  return {
    ...state.meal,
    candidates: state.meal.candidateIds.map((id) => state.dishes.find((dish) => dish.id === id)).filter(Boolean),
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
  if (url === "/meals/current") return mealView(state);

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
    return {
      meal: state.meal,
      items: state.shoppingItems,
      pendingCount: state.shoppingItems.filter((item) => !item.completed).length,
      completedCount: state.shoppingItems.filter((item) => item.completed).length,
    };
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
    const recipe = { id: nextId("recipe"), name: "新菜谱", note: "", dishIds: [], coverImages: [], ...data };
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

  if (url === "/meals/join") {
    return data.code === state.meal.code
      ? mealView(state)
      : Promise.reject(new Error("没有找到这个饭局，请检查点餐码"));
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
    const dish = state.recommendations.find((item) => item.name === "青椒牛柳") || state.recommendations[0];
    return {
      id: "ai-recommendation-1",
      source: "ai",
      sourceLabel: "AI生成推荐",
      reason: "符合下饭、少油和 30 分钟内完成的条件。",
      dish: { ...dish, image: "/assets/images/what-to-eat-kung-pao-chicken.jpg", name: "宫保鸡丁", category: "川菜", tags: ["下饭", "快手", "微辣"] },
      cost: data.useFreeQuota ? 0 : 8,
    };
  }

  const saveAiMatch = match(url, /^\/ai\/recommendations\/([^/]+)\/save$/);
  if (saveAiMatch) {
    const source = state.dishes.find((item) => item.name === "青椒牛柳") || state.dishes[0];
    const dish = { ...source, id: nextId("dish"), name: "宫保鸡丁", discoverable: false, sourceLocked: true, image: "/assets/images/what-to-eat-kung-pao-chicken.jpg" };
    mutate((current) => {
      current.dishes.unshift(dish);
      return current;
    });
    return dish;
  }

  if (url === "/ai/prep-plans") {
    return {
      id: nextId("prep"),
      cost: 6,
      steps: [
        { id: "prep-1", phase: "现在做", title: "鸡腿冷水下锅焯水", detail: "加姜片，水开后撇去浮沫，约 6 分钟。", parallel: "同时把香菇切片，西兰花切小朵。", done: "鸡腿表面变白且没有明显血沫" },
        { id: "prep-2", phase: "接着做", title: "腌牛肉并烧一锅水", detail: "牛肉加入生抽和淀粉抓匀，静置 10 分钟。", parallel: "水烧开后焯西兰花 1 分钟并过凉。", done: "牛肉吸收料汁，西兰花保持翠绿" },
        { id: "prep-3", phase: "最后做", title: "按耐放程度安排下锅", detail: "先炖汤和焖饭，临开餐再炒牛肉与青菜。", parallel: "等待焖煮时整理台面和调味料。", done: "热菜集中在开餐前 10 分钟完成" }
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
