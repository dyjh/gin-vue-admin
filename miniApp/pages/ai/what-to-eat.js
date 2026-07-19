const api = require("../../services/api");
const { go } = require("../../utils/navigation");

const UNLOCK_DAYS = 7;
const FREE_QUOTA = 2;
const POINT_COST = 2;

function normalizeStatus(raw = {}) {
  const unlockDays = Number(raw.unlockDays || UNLOCK_DAYS);
  const checkinDays = Number(raw.checkinDays || 0);
  return {
    unlocked: Boolean(raw.unlocked),
    checkinDays,
    unlockDays,
    freeQuota: Number(raw.freeQuota ?? FREE_QUOTA),
    pointBalance: Number(raw.pointBalance ?? raw.points ?? 0),
    pointCost: Number(raw.pointCost ?? raw.cost ?? POINT_COST),
  };
}

Page({
  data: {
    status: null,
    viewMode: "conditions",
    sceneOptions: ["午餐", "晚餐", "夜宵"],
    scene: "晚餐",
    peopleOptions: [
      { label: "一人食", people: 1 },
      { label: "两人餐", people: 2 },
      { label: "多人餐", people: 4 },
    ],
    people: 2,
    peopleLabel: "两人餐",
    tasteOptions: ["快手", "下饭", "清淡", "汤菜", "少油", "素菜"],
    selectedTags: ["快手", "下饭"],
    selectedTasteText: "快手、下饭",
    preference: "少油 · 不太辣",
    profileTags: ["快手较多", "常选鸡肉", "少油"],
    remainingUnlockDays: 0,
    unlockProgressPercent: 0,
    result: null,
    resultTags: [],
    resultSummary: "",
    resultReason: "",
    resultReasonExtra: "",
    showPointsConfirm: false,
    generating: false,
    excludedRecommendationIds: [],
  },

  async onLoad() {
    try {
      const status = normalizeStatus(await api.getWhatToEatStatus());
      this.setStatus(status);
    } catch (error) {
      wx.showToast({ title: "状态加载失败", icon: "none" });
    }
  },

  setStatus(status) {
    const remainingUnlockDays = Math.max(status.unlockDays - status.checkinDays, 0);
    const unlockProgressPercent = Math.min(Math.round((status.checkinDays / status.unlockDays) * 100), 100);
    this.setData({ status, remainingUnlockDays, unlockProgressPercent });
  },

  chooseScene(event) {
    if (!this.data.status.unlocked) return;
    this.setData({ scene: event.currentTarget.dataset.value });
  },

  choosePeople(event) {
    if (!this.data.status.unlocked) return;
    this.setData({
      people: Number(event.currentTarget.dataset.people),
      peopleLabel: event.currentTarget.dataset.label,
    });
  },

  toggleTag(event) {
    if (!this.data.status.unlocked) return;
    const tag = event.currentTarget.dataset.tag;
    const selectedTags = [...this.data.selectedTags];
    const index = selectedTags.indexOf(tag);
    if (index >= 0) selectedTags.splice(index, 1);
    else selectedTags.push(tag);
    this.setData({
      selectedTags,
      selectedTasteText: selectedTags.length ? selectedTags.join("、") : "不限口味",
    });
  },

  choosePreference() {
    if (!this.data.status.unlocked) return;
    const options = ["少油 · 不太辣", "不吃香菜", "无特殊忌口"];
    wx.showActionSheet({
      itemList: options,
      success: ({ tapIndex }) => this.setData({ preference: options[tapIndex] }),
    });
  },

  generate() {
    if (!this.data.status.unlocked || this.data.generating) return;
    if (this.data.status.freeQuota > 0) {
      this.runGenerate(true);
      return;
    }
    this.setData({ showPointsConfirm: true });
  },

  cancelPointsConfirm() {
    this.setData({ showPointsConfirm: false });
  },

  stopPropagation() {},

  confirmPointsGenerate() {
    if (this.data.status.pointBalance < this.data.status.pointCost) {
      wx.showToast({ title: "积分不足", icon: "none" });
      return;
    }
    this.runGenerate(false);
  },

  async runGenerate(useFreeQuota) {
    this.setData({ generating: true });
    try {
      const result = await api.generateWhatToEat({
        tags: this.data.selectedTags,
        people: this.data.people,
        timeLimitMinutes: 30,
        excludedRecommendationIds: this.data.excludedRecommendationIds,
        useFreeQuota,
      });
      const status = {
        ...this.data.status,
        freeQuota: useFreeQuota ? Math.max(this.data.status.freeQuota - 1, 0) : this.data.status.freeQuota,
        pointBalance: useFreeQuota
          ? this.data.status.pointBalance
          : Math.max(this.data.status.pointBalance - this.data.status.pointCost, 0),
      };
      const dishTags = Array.isArray(result.dish.tags) && result.dish.tags.length
        ? result.dish.tags.slice(0, 3)
        : ["下饭", "快手", "微辣"];
      const selectedTasteText = this.data.selectedTags.length ? this.data.selectedTags.join("、") : "当前偏好";
      this.setData({
        status,
        result,
        resultTags: dishTags,
        resultSummary: this.data.scene + " · " + this.data.peopleLabel + " · " + (this.data.selectedTags.slice(0, 2).join("") || "口味不限"),
        resultReason: "你选择了" + selectedTasteText + "，" + result.dish.name + "是做法明确的经典" + result.dish.category + "。",
        resultReasonExtra: this.data.peopleLabel + "容易调整，口味也可以按偏好处理。",
        selectedTasteText,
        viewMode: "result",
        showPointsConfirm: false,
      });
      wx.pageScrollTo({ scrollTop: 0, duration: 0 });
    } catch (error) {
      wx.showToast({ title: "推荐失败，请稍后重试", icon: "none" });
    } finally {
      this.setData({ generating: false });
    }
  },

  openDetail() {
    if (!this.data.result) return;
    go("/pages/ai/what-to-eat-detail", {
      id: this.data.result.id,
      source: this.data.result.source,
    });
  },

  regenerate() {
    if (!this.data.result) return;
    const excludedRecommendationIds = [
      ...this.data.excludedRecommendationIds,
      this.data.result.id,
    ];
    this.setData({ excludedRecommendationIds });
    this.generate();
  },

  ignore() {
    if (!this.data.result) return;
    this.setData({
      excludedRecommendationIds: [...this.data.excludedRecommendationIds, this.data.result.id],
      result: null,
      viewMode: "conditions",
    });
    wx.pageScrollTo({ scrollTop: 0, duration: 0 });
  },
});
