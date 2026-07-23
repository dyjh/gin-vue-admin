const api = require("../../services/api");
const { go, home } = require("../../utils/navigation");

function getDeadlineTarget(meal) {
  const raw = meal.deadlineAt || meal.deadline || "";
  const parsed = Date.parse(raw);
  if (!Number.isNaN(parsed)) return new Date(parsed);

  const match = String(raw).match(/^(今天|明天)\s*(\d{1,2}):(\d{2})$/);
  if (!match) return null;

  const target = new Date();
  target.setHours(Number(match[2]), Number(match[3]), 0, 0);
  if (match[1] === "明天") target.setDate(target.getDate() + 1);
  return target;
}

function getRemainingLabel(meal) {
  const target = getDeadlineTarget(meal);
  if (!target) return "等待大家点菜";

  const remainingMinutes = Math.max(0, Math.ceil((target.getTime() - Date.now()) / 60000));
  if (remainingMinutes === 0) return "点餐即将结束";

  const hours = Math.floor(remainingMinutes / 60);
  const minutes = remainingMinutes % 60;
  return hours
    ? "剩余 " + hours + " 小时 " + minutes + " 分"
    : "剩余 " + minutes + " 分";
}

Page({
  data: {
    meal: null,
    codeDigits: [],
    candidateDishes: [],
    remainingLabel: "",
    closing: false,
    showCandidates: false,
    candidateSheetOffset: 0,
    candidateSheetDragStart: 0,
    candidateSheetDragging: false,
    showCloseConfirm: false,
  },

  onLoad() {
    this.load();
  },

  async load() {
    const meal = await api.getCurrentMeal();
    this.applyMeal(meal);
  },

  applyMeal(meal) {
    const codeDigits = String(meal.code || "")
      .padEnd(6, " ")
      .slice(0, 6)
      .split("");

    this.setData({
      meal,
      codeDigits,
      candidateDishes: meal.candidates || [],
      remainingLabel: getRemainingLabel(meal),
    });
  },

  copyCode() {
    wx.setClipboardData({ data: String(this.data.meal.code || "") });
  },

  openVote() {
    go("/pages/meal/vote");
  },

  openCandidates() {
    this.setData({
      showCandidates: true,
      candidateSheetOffset: 0,
      candidateSheetDragging: false,
    });
  },

  closeCandidates() {
    this.setData({
      showCandidates: false,
      candidateSheetOffset: 0,
      candidateSheetDragging: false,
    });
  },

  startCandidateDrag(event) {
    const touch = event.touches && event.touches[0];
    if (!touch) return;
    this.setData({
      candidateSheetDragStart: touch.clientY,
      candidateSheetDragging: true,
    });
  },

  moveCandidateDrag(event) {
    const touch = event.touches && event.touches[0];
    if (!touch || !this.data.candidateSheetDragging) return;
    const offset = Math.max(0, touch.clientY - this.data.candidateSheetDragStart);
    this.setData({ candidateSheetOffset: Math.min(offset, 180) });
  },

  endCandidateDrag() {
    if (!this.data.candidateSheetDragging) return;
    if (this.data.candidateSheetOffset >= 56) {
      this.closeCandidates();
      return;
    }
    this.setData({
      candidateSheetOffset: 0,
      candidateSheetDragging: false,
    });
  },

  openCloseConfirm() {
    this.setData({ showCloseConfirm: true });
  },

  closeCloseConfirm() {
    if (this.data.closing) return;
    this.setData({ showCloseConfirm: false });
  },

  async confirmCloseMeal() {
    if (this.data.closing) return;
    this.setData({ closing: true });
    try {
      const meal = await api.closeMeal(this.data.meal.id);
      this.applyMeal(meal);
      this.setData({ showCloseConfirm: false });
      wx.showToast({ title: "点餐已关闭", icon: "success" });
    } finally {
      this.setData({ closing: false });
    }
  },

  openStats() {
    go("/pages/meal/stats");
  },

  noop() {},

  home,

  onShareAppMessage() {
    return {
      title: "来参加“" + this.data.meal.name + "”点菜",
      path: "/pages/meal/vote?state=join&code=" + this.data.meal.code,
    };
  },
});
