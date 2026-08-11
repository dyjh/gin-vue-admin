const assert = require("assert");
const {
  createDishForm,
  getDishFormOptions,
  validateDishForm,
} = require("../utils/dish-form");
const { normalizeDishPayload } = require("../services/upload");

const metadata = {
  dishCategories: [
    { id: "category-stir-fry", name: "炒菜", enabled: true },
    { id: "category-staple", name: "主食", enabled: true },
    { id: "category-disabled", name: "停用分类", enabled: false },
  ],
  dishTags: [
    { id: "tag-home", name: "家常菜", enabled: true },
    { id: "tag-quick", name: "快手菜", enabled: true },
    { id: "tag-disabled", name: "停用标签", enabled: false },
  ],
  ingredientUnits: ["克", "个"],
};

async function main() {
  const options = getDishFormOptions(metadata);
  assert.deepStrictEqual(options.categories, ["炒菜", "主食"]);
  assert.deepStrictEqual(options.tagOptions, ["家常菜", "快手菜"]);
  assert.deepStrictEqual(options.ingredientUnits, ["克", "个"]);

  const form = createDishForm({
    name: "番茄炒蛋",
    category: "家常菜",
    tags: ["tag-home", "快手", "停用标签"],
    coverUrl: "/uploads/file/cover.jpg",
    coverFileId: "cover-1",
    ingredients: [{ name: "鸡蛋", amount: "2", unit: "未知单位" }],
    steps: [{ text: "", imageUrl: "", imageFileId: "" }],
  }, metadata);
  assert.strictEqual(form.category, "炒菜");
  assert.deepStrictEqual(form.tags, ["家常菜"]);
  assert.strictEqual(form.ingredients[0].unit, "克");
  assert.strictEqual(validateDishForm(form, options), "");

  const invalidCatalogForm = { ...form, category: "家常菜" };
  assert.strictEqual(validateDishForm(invalidCatalogForm, options), "请选择有效的菜品分类");
  const imageOnlyStepForm = {
    ...form,
    steps: [{ text: "", imageUrl: "wxfile://step.jpg", imageFileId: "" }],
  };
  assert.strictEqual(validateDishForm(imageOnlyStepForm, options), "添加了步骤图时也要填写做法");

  const payload = await normalizeDishPayload({ ...form, status: "usable" });
  assert.strictEqual(payload.category, "炒菜");
  assert.strictEqual(payload.coverFileId, "cover-1");
  assert.deepStrictEqual(payload.steps, []);
  assert.deepStrictEqual(payload.ingredients, [{
    name: "鸡蛋",
    amount: "2",
    unit: "克",
    note: "",
    sortOrder: 1,
  }]);

  console.log(JSON.stringify({
    categories: options.categories.length,
    tags: options.tagOptions.length,
    units: options.ingredientUnits.length,
    blankStepsRemoved: payload.steps.length === 0,
  }));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
