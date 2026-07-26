const api = require("../../services/api");

function pad(value) {
  return String(value).padStart(2, "0");
}

function toDate(value) {
  const matched = String(value || "").match(/^(\d{4})-(\d{2})-(\d{2})/);
  if (!matched) return null;
  return new Date(Number(matched[1]), Number(matched[2]) - 1, Number(matched[3]));
}

function monthLabel(value) {
  const date = toDate(value);
  return date ? `${date.getFullYear()}年${date.getMonth() + 1}月` : "更早";
}

function dateLabel(value) {
  const date = toDate(value);
  if (!date) return "";
  const weekdays = ["周日", "周一", "周二", "周三", "周四", "周五", "周六"];
  return `${date.getMonth() + 1}月${date.getDate()}日 ${weekdays[date.getDay()]}`;
}

function timeLabel(value) {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? ""
    : `${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function formatMeal(meal) {
  const occurredAt = meal.completedAt || meal.cancelledAt || meal.confirmedAt || meal.createdAt;
  return {
    ...meal,
    monthLabel: monthLabel(occurredAt),
    dateLabel: dateLabel(occurredAt),
    timeLabel: timeLabel(occurredAt),
    statusLabel: meal.status === "cancelled" ? "已取消" : "已完成",
  };
}

function groupMeals(list) {
  return list.reduce((groups, meal) => {
    const last = groups[groups.length - 1];
    if (last && last.label === meal.monthLabel) {
      last.items.push(meal);
      return groups;
    }
    groups.push({ label: meal.monthLabel, items: [meal] });
    return groups;
  }, []);
}

Page({
  data: {
    list: [],
    groups: [],
    page: 0,
    pageSize: 5,
    total: 0,
    loading: false,
    initialLoading: true,
    selectedMeal: null,
    selectedShoppingList: null,
    showDetail: false,
  },

  onLoad(options) {
    this.pendingMealId = options && options.mealId ? String(options.mealId) : "";
    this.refresh();
  },

  onPullDownRefresh() {
    this.refresh().finally(() => wx.stopPullDownRefresh());
  },

  onReachBottom() {
    if (this.data.list.length < this.data.total) this.loadMore();
  },

  async refresh() {
    this.setData({
      list: [],
      groups: [],
      page: 0,
      total: 0,
      initialLoading: true,
    });
    await this.loadMore();
    if (this.pendingMealId) {
      const mealId = this.pendingMealId;
      this.pendingMealId = "";
      await this.openMealDetail(mealId);
    }
  },

  async loadMore() {
    if (this.data.loading) return;
    this.setData({ loading: true });
    try {
      const result = await api.listMeals({
        scope: "history",
        page: this.data.page + 1,
        pageSize: this.data.pageSize,
      });
      const ids = new Set(this.data.list.map((item) => item.id));
      const next = result.list.map(formatMeal).filter((item) => !ids.has(item.id));
      const list = [...this.data.list, ...next];
      this.setData({
        list,
        groups: groupMeals(list),
        page: result.page,
        total: result.total,
      });
    } finally {
      this.setData({ loading: false, initialLoading: false });
    }
  },

  async openDetail(event) {
    const id = event.currentTarget.dataset.id;
    await this.openMealDetail(id);
  },

  async openMealDetail(id) {
    const meal = await api.getMeal(id);
    let shoppingList = null;
    if (
      (
        meal.status === "completed"
        || (meal.readOnly && meal.readOnlyReason === "creator_disabled")
      )
      && meal.shoppingListId
    ) {
      shoppingList = await api.getShoppingListById(meal.shoppingListId);
    }
    this.setData({
      selectedMeal: formatMeal(meal),
      selectedShoppingList: shoppingList,
      showDetail: true,
    });
  },

  closeDetail() {
    this.setData({ showDetail: false, selectedShoppingList: null });
  },

  noop() {},
});
