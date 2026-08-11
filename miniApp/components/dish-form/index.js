const api = require("../../services/api");
const {
  createDishForm,
  getDishFormOptions,
  validateDishForm,
} = require("../../utils/dish-form");

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    value: { type: Object, value: {} },
    mode: { type: String, value: "create" },
    coverFeature: { type: Object, value: null },
  },
  data: {
    form: {},
    metadata: {},
    categories: [],
    tagOptions: [],
    ingredientUnits: [],
    generating: false,
    showTagOptions: false,
  },
  observers: {
    value(value) {
      if (value && value.name !== undefined) {
        this.setData({ form: createDishForm(value, this.data.metadata) });
      }
    },
  },
  lifetimes: {
    attached() {
      this.setData({ form: createDishForm(this.data.value || {}) });
      this.loadMetadata();
    },
  },
  methods: {
    async loadMetadata() {
      try {
        const metadata = await api.getMetadata();
        const options = getDishFormOptions(metadata);
        this.setData({
          metadata,
          ...options,
          form: createDishForm(this.data.form, metadata),
        });
      } catch (error) {
        // 通用请求层已经展示错误提示；保留空分类阻止提交无效目录值。
      }
    },
    inputField(event) {
      const field = event.currentTarget.dataset.field;
      this.setData({ [`form.${field}`]: event.detail.value });
    },
    chooseCategory(event) {
      this.setData({ "form.category": this.data.categories[event.detail.value] });
    },
    chooseServing(event) {
      this.setData({ "form.serving": Number(event.detail.value) + 1 });
    },
    toggleTagPicker() {
      this.setData({ showTagOptions: !this.data.showTagOptions });
    },
    toggleTag(event) {
      const tag = event.currentTarget.dataset.tag;
      const tags = [...(this.data.form.tags || [])];
      const index = tags.indexOf(tag);
      if (index >= 0) tags.splice(index, 1);
      else if (tags.length < 3) tags.push(tag);
      else {
        wx.showToast({ title: "最多选择 3 个标签", icon: "none" });
        return;
      }
      this.setData({ "form.tags": tags });
    },
    chooseCover() {
      wx.chooseMedia({
        count: 1,
        mediaType: ["image"],
        success: (result) => {
          this.setData({
            "form.coverUrl": result.tempFiles[0].tempFilePath,
            "form.coverFileId": "",
          });
        },
      });
    },
    async generateCover() {
      if (!this.data.form.name) {
        wx.showToast({ title: "先填写菜名", icon: "none" });
        return;
      }
      this.setData({ generating: true });
      try {
        const result = await api.generateDishCover({ name: this.data.form.name, description: this.data.form.description });
        this.setData({
          "form.coverUrl": result.url,
          "form.coverFileId": result.fileId,
        });
      } finally {
        this.setData({ generating: false });
      }
    },
    ingredientInput(event) {
      const index = event.currentTarget.dataset.index;
      const field = event.currentTarget.dataset.field;
      this.setData({ [`form.ingredients[${index}].${field}`]: event.detail.value });
    },
    addIngredient() {
      const unit = this.data.ingredientUnits.includes("克")
        ? "克"
        : this.data.ingredientUnits[0] || "";
      const ingredients = [
        ...this.data.form.ingredients,
        { id: `ingredient-${Date.now()}`, name: "", amount: "", unit, note: "" },
      ];
      this.setData({ "form.ingredients": ingredients });
    },
    removeIngredient(event) {
      const index = event.currentTarget.dataset.index;
      const ingredients = this.data.form.ingredients.filter((_, current) => current !== index);
      const unit = this.data.ingredientUnits.includes("克")
        ? "克"
        : this.data.ingredientUnits[0] || "";
      this.setData({
        "form.ingredients": ingredients.length
          ? ingredients
          : [{ id: `ingredient-${Date.now()}`, name: "", amount: "", unit, note: "" }],
      });
    },
    chooseIngredientUnit(event) {
      const index = event.currentTarget.dataset.index;
      const unit = this.data.ingredientUnits[event.detail.value];
      if (unit) this.setData({ [`form.ingredients[${index}].unit`]: unit });
    },
    stepInput(event) {
      const index = event.currentTarget.dataset.index;
      this.setData({ [`form.steps[${index}].text`]: event.detail.value });
    },
    addStep() {
      const steps = [
        ...this.data.form.steps,
        { id: `step-${Date.now()}`, text: "", imageUrl: "", imageFileId: "" },
      ];
      this.setData({ "form.steps": steps });
    },
    chooseStepImage(event) {
      const index = event.currentTarget.dataset.index;
      wx.chooseMedia({
        count: 1,
        mediaType: ["image"],
        success: (result) => {
          this.setData({
            [`form.steps[${index}].imageUrl`]: result.tempFiles[0].tempFilePath,
            [`form.steps[${index}].imageFileId`]: "",
          });
        },
      });
    },
    removeStep(event) {
      const index = event.currentTarget.dataset.index;
      const steps = this.data.form.steps.filter((_, current) => current !== index);
      this.setData({
        "form.steps": steps.length
          ? steps
          : [{ id: `step-${Date.now()}`, text: "", imageUrl: "", imageFileId: "" }],
      });
    },
    save(event) {
      const status = event.currentTarget.dataset.status;
      const form = this.data.form;
      const message = validateDishForm(form, this.data);
      if (message) {
        wx.showToast({ title: message, icon: "none" });
        return;
      }
      this.triggerEvent("save", { form: { ...form, status } });
    },
  },
});
