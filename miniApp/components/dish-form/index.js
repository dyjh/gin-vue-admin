const api = require("../../services/api");

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    value: { type: Object, value: {} },
    mode: { type: String, value: "create" },
  },
  data: {
    form: {},
    categories: ["主食", "家常菜", "素菜", "汤菜"],
    tagOptions: ["下饭", "快手", "少油", "清淡", "可提前备", "适合孩子"],
    generating: false,
    showTagOptions: false,
  },
  observers: {
    value(value) {
      if (value && value.name !== undefined) {
        this.setData({ form: JSON.parse(JSON.stringify(value)) });
      }
    },
  },
  lifetimes: {
    attached() {
      const value = this.data.value || {};
      this.setData({
        form: {
          name: "",
          category: "家常菜",
          tags: [],
          serving: 2,
          description: "",
          image: "",
          ingredients: [{ id: "ingredient-1", name: "", amount: "" }],
          steps: [{ id: "step-1", text: "", image: "" }],
          ...JSON.parse(JSON.stringify(value)),
        },
      });
    },
  },
  methods: {
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
          this.setData({ "form.image": result.tempFiles[0].tempFilePath });
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
        this.setData({ "form.image": result.image });
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
      const ingredients = [...this.data.form.ingredients, { id: `ingredient-${Date.now()}`, name: "", amount: "" }];
      this.setData({ "form.ingredients": ingredients });
    },
    removeIngredient(event) {
      const index = event.currentTarget.dataset.index;
      const ingredients = this.data.form.ingredients.filter((_, current) => current !== index);
      this.setData({ "form.ingredients": ingredients.length ? ingredients : [{ id: `ingredient-${Date.now()}`, name: "", amount: "" }] });
    },
    stepInput(event) {
      const index = event.currentTarget.dataset.index;
      this.setData({ [`form.steps[${index}].text`]: event.detail.value });
    },
    addStep() {
      const steps = [...this.data.form.steps, { id: `step-${Date.now()}`, text: "", image: "" }];
      this.setData({ "form.steps": steps });
    },
    chooseStepImage(event) {
      const index = event.currentTarget.dataset.index;
      wx.chooseMedia({
        count: 1,
        mediaType: ["image"],
        success: (result) => {
          this.setData({ [`form.steps[${index}].image`]: result.tempFiles[0].tempFilePath });
        },
      });
    },
    removeStep(event) {
      const index = event.currentTarget.dataset.index;
      const steps = this.data.form.steps.filter((_, current) => current !== index);
      this.setData({ "form.steps": steps.length ? steps : [{ id: `step-${Date.now()}`, text: "", image: "" }] });
    },
    save(event) {
      const status = event.currentTarget.dataset.status;
      const form = this.data.form;
      if (!form.image) {
        wx.showToast({ title: "请先选择菜品封面图", icon: "none" });
        return;
      }
      if (!form.name.trim()) {
        wx.showToast({ title: "请填写菜名", icon: "none" });
        return;
      }
      this.triggerEvent("save", { form: { ...form, status } });
    },
  },
});
