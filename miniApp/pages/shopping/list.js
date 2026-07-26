const api = require("../../services/api");
const { go, home } = require("../../utils/navigation");
const { SHARE_IMAGES } = require("../../config/share");
const { getFeature } = require("../../utils/features");

const UNIT_OPTIONS = ["个", "克", "千克", "斤", "颗", "袋", "盒", "瓶", "把", "根", "瓣", "包", "毫升", "升"];
const UNITS_BY_LENGTH = [...UNIT_OPTIONS].sort((left, right) => right.length - left.length);

function splitAmount(value = "") {
  const amount = String(value).trim();
  const unit = UNITS_BY_LENGTH.find((candidate) => amount.endsWith(candidate)) || "";
  return {
    quantity: unit ? amount.slice(0, -unit.length).trim() : amount,
    unit,
  };
}

Page({
  data: {
    shareToken: "",
    meal: null,
    loading: true,
    items: [],
    scope: "pending",
    pendingCount: 0,
    completedCount: 0,
    finishing: false,
    showEditor: false,
    editingId: "",
    unitOptions: UNIT_OPTIONS,
    unitIndex: 0,
    form: { name: "", quantity: "", unit: "", note: "" },
    prepFeature: null,
  },

  onLoad() {
    this.load();
  },

  async load() {
    this.setData({ loading: true });
    try {
      const [data, current] = await Promise.all([
        api.getShoppingList({ showError: false }),
        api.getCurrentMeal(),
        api.getRuntimeConfig(),
      ]);
      const activeMeal = current.meal && current.meal.id === data.meal.id
        ? current.meal
        : data.meal;
      this.setData({
        ...data,
        meal: activeMeal,
        prepFeature: getFeature("prep_sequence"),
      });
    } catch (error) {
      this.setData({
        meal: null,
        items: [],
        pendingCount: 0,
        completedCount: 0,
      });
    } finally {
      this.setData({ loading: false });
    }
  },

  setScope(event) {
    this.setData({ scope: event.currentTarget.dataset.scope });
  },

  async toggle(event) {
    const id = event.currentTarget.dataset.id;
    const item = this.data.items.find((entry) => entry.id === id);
    await api.updateShoppingItem(id, { completed: !item.completed });
    await this.load();
  },

  openEdit(event) {
    const id = event.currentTarget.dataset.id;
    const item = this.data.items.find((entry) => entry.id === id);
    const amount = splitAmount(item.amount);
    this.setData({
      showEditor: true,
      editingId: id,
      unitIndex: Math.max(UNIT_OPTIONS.indexOf(amount.unit), 0),
      form: { name: item.name, quantity: amount.quantity, unit: amount.unit, note: item.note || "" },
    });
  },

  openAdd() {
    this.setData({
      showEditor: true,
      editingId: "",
      unitIndex: 0,
      form: { name: "", quantity: "", unit: "", note: "" },
    });
  },

  closeEditor() {
    this.setData({ showEditor: false });
  },

  noop() {},

  home,

  inputField(event) {
    const field = event.currentTarget.dataset.field;
    this.setData({ [`form.${field}`]: event.detail.value });
  },

  selectUnit(event) {
    const unitIndex = Number(event.detail.value);
    this.setData({ unitIndex, "form.unit": UNIT_OPTIONS[unitIndex] });
  },

  async saveItem() {
    const { name, quantity, unit, note } = this.data.form;
    if (!name.trim()) {
      wx.showToast({ title: "请填写采购项名称", icon: "none" });
      return;
    }
    if (!quantity.trim()) {
      wx.showToast({ title: "请填写数量", icon: "none" });
      return;
    }
    if (!unit) {
      wx.showToast({ title: "请选择单位", icon: "none" });
      return;
    }

    const editingId = this.data.editingId;
    const payload = {
      name: name.trim(),
      amount: `${quantity.trim()} ${unit}`,
      note: note.trim(),
    };
    if (editingId) await api.updateShoppingItem(editingId, payload);
    else await api.addShoppingItem(payload);
    this.setData({ showEditor: false, editingId: "" });
    await this.load();
    wx.showToast({ title: editingId ? "采购项已更新" : "采购项已添加", icon: "success" });
  },

  async removeItem() {
    await api.deleteShoppingItem(this.data.editingId);
    this.setData({ showEditor: false });
    await this.load();
    wx.showToast({ title: "采购项已删除", icon: "success" });
  },

  async copy() {
    const exported = await api.exportShoppingListText(this.data.id);
    wx.setClipboardData({ data: exported.text });
  },

  openPrep() {
    go("/pages/meal/prep-guide");
  },

  openCompleteConfirm() {
    const meal = this.data.meal;
    if (
      !meal
      || meal.status !== "confirmed"
      || !meal.createdByMe
      || this.data.finishing
    ) return;

    wx.showModal({
      title: "结束这场饭局？",
      content: "结束后本场饭局会进入历史记录，采购清单仍会保留；只有结束当前饭局后，才能创建下一场。",
      confirmText: "确认结束",
      confirmColor: "#159B55",
      success: ({ confirm }) => {
        if (confirm) this.completeMeal();
      },
    });
  },

  async completeMeal() {
    if (this.data.finishing) return;
    this.setData({ finishing: true });
    try {
      await api.completeMeal(this.data.meal.id);
      wx.showToast({ title: "饭局已结束", icon: "success" });
      setTimeout(() => go("/pages/meal/history"), 350);
    } finally {
      this.setData({ finishing: false });
    }
  },

  onShareAppMessage() {
    const token = encodeURIComponent(this.data.shareToken || "");
    return {
      title: `${this.data.meal.name}采购清单`,
      path: token ? `/pages/shopping/share?token=${token}` : "/pages/shopping/share",
      imageUrl: SHARE_IMAGES.shoppingList,
    };
  },
});
