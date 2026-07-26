const api = require("../../services/api");
const { home } = require("../../utils/navigation");

function normalizeCode(value) {
  if (value === undefined || value === null) return "";
  return String(value).replace(/\D/g, "").slice(0, 6);
}

function formatDeadline(value) {
  const deadline = new Date(value);
  if (Number.isNaN(deadline.getTime())) return "";
  const now = new Date();
  const tomorrow = new Date(now);
  tomorrow.setDate(tomorrow.getDate() + 1);
  const sameDay = (left, right) => left.getFullYear() === right.getFullYear()
    && left.getMonth() === right.getMonth()
    && left.getDate() === right.getDate();
  const day = sameDay(deadline, now) ? "今天" : sameDay(deadline, tomorrow) ? "明天" : `${deadline.getMonth() + 1}月${deadline.getDate()}日`;
  return `${day} ${String(deadline.getHours()).padStart(2, "0")}:${String(deadline.getMinutes()).padStart(2, "0")}`;
}

Page({
  data: {
    state: "collecting",
    code: "",

    checkingCode: false,
    codeMatched: false,
    matchedMealName: "",
    codeError: "",
    meal: null,
    categories: [],
    category: "全部",
    selectedIds: [],
    joining: false,
    saving: false,
    subscriptionAvailable: false,
    subscriptionButtonText: "接收最终结果提醒",
    subscribing: false,
    subscriptionRecorded: false,
  },

  async onLoad(options) {
    const code = normalizeCode(options && options.code);
    this.mealId = options && options.mealId ? String(options.mealId) : "";
    this.setData({ state: options.state || "collecting", code });
    if (options.state !== "join") await this.load();
    else if (code.length === 6) await this.lookupCode(code);
  },
  async load() {
    const meal = this.mealId
      ? await api.getMeal(this.mealId)
      : (await api.getCurrentMeal()).meal;
    if (!meal) {
      this.setData({ meal: null, state: "closed", categories: [], selectedIds: [] });
      return;
    }
    await this.applyMeal(meal);
  },

  async applyMeal(meal) {
    const [candidateResponse, voteResponse] = await Promise.all([
      api.listMealCandidates(meal.id),
      api.getMyMealVotes(meal.id),
    ]);
    const candidates = candidateResponse.list;
    const categories = [
      { category: "全部", count: candidates.length },
      ...candidateResponse.categories,
    ];
    const app = getApp();
    const subscriptionConfig = app.globalData.runtimeConfig
      && app.globalData.runtimeConfig.mealFinalResultSubscription;
    this.setData({
      meal: { ...meal, candidates, deadlineLabel: formatDeadline(meal.deadlineAt) },
      categories,
      selectedIds: voteResponse.candidateIds.filter((id) => {
        const candidate = candidates.find((item) => item.id === id);
        return candidate && candidate.available;
      }),
      state: meal.status === "cancelled" ? "cancelled" : meal.status === "collecting" ? "collecting" : "closed",
      subscriptionAvailable: Boolean(
        !meal.createdByMe
        && !meal.finalResultSubscriptionAccepted
        && subscriptionConfig
        && subscriptionConfig.enabled
        && subscriptionConfig.templateId
      ),
      subscriptionButtonText: subscriptionConfig && subscriptionConfig.buttonText
        ? subscriptionConfig.buttonText
        : "接收最终结果提醒",
      subscriptionRecorded: Boolean(meal.finalResultSubscriptionAccepted),
    });
  },

  inputCode(event) {
    const code = normalizeCode(event && event.detail && event.detail.value);
    this.setData({ code, codeMatched: false, matchedMealName: "", codeError: "", checkingCode: false });
    if (code.length === 6) this.lookupCode(code);
    return code;
  },


  async lookupCode(code) {
    this.setData({ checkingCode: true, codeMatched: false, codeError: "" });
    try {
      const preview = await api.previewMealByCode({ code });
      if (this.data.code !== code) return false;
      if (preview.status !== "collecting") {
        this.setData({ checkingCode: false, codeError: preview.status === "cancelled" ? "这个饭局已取消" : "这个饭局已停止点餐" });
        return false;
      }
      this.setData({ checkingCode: false, codeMatched: true, matchedMealName: preview.name });
      return true;
    } catch (error) {
      if (this.data.code === code) this.setData({ checkingCode: false, codeMatched: false, codeError: "没有找到这个饭局，请检查点餐码" });
      return false;
    }
  },

  async join() {
    if (this.data.code.length !== 6) {
      wx.showToast({ title: "请输入 6 位点餐码", icon: "none" });
      return;
    }
    const matched = this.data.codeMatched || await this.lookupCode(this.data.code);
    if (!matched) return;
    this.setData({ joining: true });
    try {
      const meal = await api.joinMeal({ code: this.data.code });
      await this.applyMeal(meal);
      wx.showToast({ title: "已加入饭局", icon: "success" });
    } catch (error) {
      // The request layer already presents the reason.
    } finally {
      this.setData({ joining: false });
    }
  },
  selectCategory(event) {
    this.setData({ category: event.currentTarget.dataset.category });
  },

  toggleDish(event) {
    const id = event.detail.id;
    const candidate = this.data.meal.candidates.find((item) => item.id === id);
    if (!candidate || !candidate.available) {
      wx.showToast({ title: "这道候选菜已失效", icon: "none" });
      return;
    }
    const selectedIds = [...this.data.selectedIds];
    const index = selectedIds.indexOf(id);
    if (index >= 0) selectedIds.splice(index, 1);
    else selectedIds.push(id);
    this.setData({ selectedIds });
  },

  async requestFinalResultSubscription() {
    if (this.data.subscribing || !this.data.subscriptionAvailable) return;
    const app = getApp();
    const config = app.globalData.runtimeConfig
      && app.globalData.runtimeConfig.mealFinalResultSubscription;
    if (!config || !config.templateId || typeof wx.requestSubscribeMessage !== "function") {
      wx.showToast({ title: "当前无法申请提醒", icon: "none" });
      return;
    }
    this.setData({ subscribing: true });
    try {
      const result = await new Promise((resolve, reject) => {
        // 只有用户主动点击本按钮时才调用一次微信订阅授权接口。
        wx.requestSubscribeMessage({
          tmplIds: [config.templateId],
          success: resolve,
          fail: reject,
        });
      });
      const authorizationResult = result[config.templateId];
      if (!["accept", "reject", "ban"].includes(authorizationResult)) {
        throw new Error("微信未返回有效授权结果");
      }
      const recorded = await api.recordMealFinalResultSubscription(
        this.data.meal.id,
        config.templateId,
        authorizationResult,
      );
      this.setData({
        subscriptionRecorded: recorded.accepted,
        subscriptionAvailable: !recorded.accepted,
      });
      wx.showToast({
        title: recorded.accepted ? "已开启本次提醒" : "已记录授权选择",
        icon: recorded.accepted ? "success" : "none",
      });
    } catch (error) {
      wx.showToast({ title: "未能登记提醒授权", icon: "none" });
    } finally {
      this.setData({ subscribing: false });
    }
  },

  async save() {
    this.setData({ saving: true });
    try {
      await api.saveVotes(this.data.meal.id, this.data.selectedIds);
      wx.showToast({ title: "点菜已保存", icon: "success" });
    } finally {
      this.setData({ saving: false });
    }
  },

  home,
});
