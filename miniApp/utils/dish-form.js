const FALLBACK_UNIT = "克";

function clone(value) {
  return JSON.parse(JSON.stringify(value || {}));
}

function enabledCatalogItems(items) {
  return (Array.isArray(items) ? items : []).filter((item) => (
    item &&
    item.enabled !== false &&
    typeof item.name === "string" &&
    item.name.trim()
  ));
}

function catalogNames(items) {
  return enabledCatalogItems(items).map((item) => item.name.trim());
}

function resolveCatalogName(value, items) {
  const normalized = String(value || "").trim();
  const match = enabledCatalogItems(items).find((item) => (
    item.name === normalized || item.id === normalized
  ));
  return match ? match.name.trim() : "";
}

function getDishFormOptions(metadata = {}) {
  return {
    categories: catalogNames(metadata.dishCategories),
    tagOptions: catalogNames(metadata.dishTags),
    ingredientUnits: (Array.isArray(metadata.ingredientUnits) ? metadata.ingredientUnits : [])
      .map((item) => String(item || "").trim())
      .filter(Boolean),
  };
}

function defaultUnit(units) {
  if (units.includes(FALLBACK_UNIT)) return FALLBACK_UNIT;
  return units[0] || FALLBACK_UNIT;
}

function createDishForm(value = {}, metadata = {}) {
  const source = clone(value);
  const options = getDishFormOptions(metadata);
  const hasCategoryCatalog = Array.isArray(metadata.dishCategories);
  const hasTagCatalog = Array.isArray(metadata.dishTags);
  const hasUnitCatalog = Array.isArray(metadata.ingredientUnits);
  const category = hasCategoryCatalog
    ? resolveCatalogName(source.category, metadata.dishCategories) || options.categories[0] || ""
    : source.category || "";
  const tags = hasTagCatalog
    ? [...new Set((Array.isArray(source.tags) ? source.tags : [])
      .map((item) => resolveCatalogName(item, metadata.dishTags))
      .filter(Boolean))]
      .slice(0, 3)
    : Array.isArray(source.tags) ? source.tags : [];
  const unitFallback = defaultUnit(options.ingredientUnits);
  const ingredients = Array.isArray(source.ingredients) && source.ingredients.length
    ? source.ingredients.map((item, index) => ({
      id: item.id || `ingredient-${index + 1}`,
      name: item.name || "",
      amount: item.amount || "",
      unit: hasUnitCatalog && !options.ingredientUnits.includes(item.unit)
        ? unitFallback
        : item.unit || unitFallback,
      note: item.note || "",
    }))
    : [{ id: "ingredient-1", name: "", amount: "", unit: unitFallback, note: "" }];
  const steps = Array.isArray(source.steps) && source.steps.length
    ? source.steps.map((item, index) => ({
      id: item.id || `step-${index + 1}`,
      text: item.text || "",
      imageUrl: item.imageUrl || "",
      imageFileId: item.imageFileId || "",
    }))
    : [{ id: "step-1", text: "", imageUrl: "", imageFileId: "" }];

  return {
    name: "",
    category: "",
    tags: [],
    serving: 2,
    description: "",
    coverUrl: "",
    coverFileId: "",
    ...source,
    category,
    tags,
    ingredients,
    steps,
  };
}

function validateDishForm(form, options) {
  if (!form.coverUrl) return "请先选择菜品封面图";
  if (!String(form.name || "").trim()) return "请填写菜名";
  if (!options.categories.length) return "菜品分类加载失败，请稍后重试";
  if (!options.categories.includes(form.category)) return "请选择有效的菜品分类";
  if ((form.tags || []).some((tag) => !options.tagOptions.includes(tag))) {
    return "请选择有效的菜品标签";
  }
  if ((form.ingredients || []).some((item) => (
    !String(item.name || "").trim() ||
    !String(item.amount || "").trim() ||
    !String(item.unit || "").trim()
  ))) {
    return "请完整填写配料名称、用量和单位";
  }
  if ((form.ingredients || []).some((item) => !options.ingredientUnits.includes(item.unit))) {
    return "请选择有效的配料单位";
  }
  const incompleteStep = (form.steps || []).find((item) => (
    !String(item.text || "").trim() && (item.imageUrl || item.imageFileId)
  ));
  if (incompleteStep) return "添加了步骤图时也要填写做法";
  return "";
}

module.exports = {
  createDishForm,
  getDishFormOptions,
  validateDishForm,
};
