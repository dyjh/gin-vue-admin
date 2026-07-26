const api = require("../../services/api");
const { go, home } = require("../../utils/navigation");

function getDeadlineTarget(meal) {
  const raw = meal.deadlineAt || "";
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

function formatDeadline(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return `${date.getMonth() + 1}月${date.getDate()}日 ${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

Page({
  data: {
    meal: null,
    codeDigits: [],
    candidateDishes: [],
    availableCandidateCount: 0,
    remainingLabel: "",
    closing: false,
    cancelling: false,
    showCandidates: false,
    candidateSheetOffset: 0,
    candidateSheetDragStart: 0,
    candidateSheetDragging: false,
    showCloseConfirm: false,
    removingCandidateId: "",
  },

  onLoad() {
    this.load();
  },

  async load() {
    const current = await api.getCurrentMeal();
    if (!current.meal) {
      this.setData({ meal: null, candidateDishes: [], availableCandidateCount: 0 });
      return;
    }
    const candidates = await api.listMealCandidates(current.meal.id);
    this.applyMeal(current.meal, candidates.list);
  },

  applyMeal(meal, candidates = this.data.candidateDishes) {
    const codeDigits = String(meal.code || "")
      .padEnd(6, " ")
      .slice(0, 6)
      .split("");

    this.setData({
      meal: {
        ...meal,
        deadlineLabel: formatDeadline(meal.deadlineAt),
      },
      codeDigits,
      candidateDishes: candidates,
      availableCandidateCount: candidates.filter((item) => item.available).length,
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
    if (this.data.removingCandidateId) return;
    this.setData({
      showCandidates: false,
      candidateSheetOffset: 0,
      candidateSheetDragging: false,
    });
  },

  requestRemoveCandidate(event) {
    const candidateId = String(event.currentTarget.dataset.id || "");
    const candidate = this.data.candidateDishes.find((item) => item.id === candidateId);
    if (
      !candidate
      || !candidate.available
      || !this.data.meal
      || !this.data.meal.createdByMe
      || this.data.meal.status !== "collecting"
      || this.data.removingCandidateId
    ) return;
    if (this.data.availableCandidateCount <= 1) {
      wx.showToast({ title: "饭局至少保留一道候选菜", icon: "none" });
      return;
    }
    wx.showModal({
      title: `移除“${candidate.name}”？`,
      content: "移除后参与者将不再看到这道候选菜，已有点选也不再计入统计。",
      confirmText: "确认移除",
      confirmColor: "#D8564F",
      success: ({ confirm }) => {
        if (confirm) this.removeCandidate(candidateId);
      },
    });
  },

  async removeCandidate(candidateId) {
    if (this.data.removingCandidateId) return;
    this.setData({ removingCandidateId: candidateId });
    try {
      await api.removeMealCandidate(this.data.meal.id, candidateId);
      const candidateDishes = this.data.candidateDishes.filter(
        (item) => item.id !== candidateId
      );
      this.setData({
        candidateDishes,
        availableCandidateCount: candidateDishes.filter((item) => item.available).length,
        "meal.candidateCount": candidateDishes.length,
      });
      wx.showToast({ title: "候选菜已移除", icon: "success" });
    } finally {
      this.setData({ removingCandidateId: "" });
    }
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
    if (this.data.cancelling) return;
    this.setData({ showCloseConfirm: true });
  },

  closeCloseConfirm() {
    if (this.data.closing) return;
    this.setData({ showCloseConfirm: false });
  },

  async confirmCloseMeal() {
    if (this.data.closing || this.data.cancelling) return;
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

  openCancelConfirm() {
    const meal = this.data.meal;
    if (
      !meal
      || !meal.createdByMe
      || !["collecting", "closed"].includes(meal.status)
      || this.data.closing
      || this.data.cancelling
    ) return;

    const statusDescription = meal.status === "collecting"
      ? "点餐仍在收集中，取消后点餐码立即失效。"
      : "点餐虽已关闭，但饭局仍在进行中。";
    wx.showModal({
      title: "取消这场饭局？",
      content: `${statusDescription}本场将记为已取消，之后才能创建下一场饭局。`,
      confirmText: "确认取消",
      confirmColor: "#D8564F",
      success: ({ confirm }) => {
        if (confirm) this.cancelMeal();
      },
    });
  },

  async cancelMeal() {
    if (this.data.cancelling || this.data.closing) return;
    this.setData({ cancelling: true });
    try {
      await api.cancelMeal(this.data.meal.id, {});
      wx.showToast({ title: "饭局已取消", icon: "success" });
      setTimeout(home, 350);
    } finally {
      this.setData({ cancelling: false });
    }
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
