const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/meal-create-cat-invite-v1.png";

const selectedDishes = [
  { name: "香菇鸡腿饭", meta: "主食 · 适合 2 人", image: "miniApp/.preview/assets/dish-detail-cover-realistic.png" },
  { name: "番茄炒蛋", meta: "家常菜 · 下饭", image: "miniApp/.preview/assets/recommended-tomato-egg.png" },
  { name: "蒜蓉西兰花", meta: "素菜 · 清淡", image: "miniApp/.preview/assets/recommended-garlic-broccoli.png" },
  { name: "冬瓜丸子汤", meta: "汤菜 · 可提前备", image: "miniApp/.preview/assets/recommended-winter-melon-soup.png" },
];

const allCandidates = [
  { name: "青椒牛柳", meta: "家常菜 · 20 分钟", selected: true, image: "miniApp/.preview/assets/recommended-green-pepper-beef.png" },
  { name: "番茄炒蛋", meta: "下饭 · 适合 2 人", selected: true, image: "miniApp/.preview/assets/recommended-tomato-egg.png" },
  { name: "蒜蓉西兰花", meta: "素菜 · 15 分钟", selected: false, image: "miniApp/.preview/assets/recommended-garlic-broccoli.png" },
  { name: "冬瓜丸子汤", meta: "汤菜 · 适合多人", selected: true, image: "miniApp/.preview/assets/recommended-winter-melon-soup.png" },
];

const recipeCandidates = [
  { name: "青椒牛柳", meta: "家常菜 · 少油", selected: true, image: "miniApp/.preview/assets/recommended-green-pepper-beef.png" },
  { name: "番茄炒蛋", meta: "家常菜 · 下饭", selected: true, image: "miniApp/.preview/assets/recommended-tomato-egg.png" },
  { name: "蒜蓉西兰花", meta: "素菜 · 清淡", selected: false, image: "miniApp/.preview/assets/recommended-garlic-broccoli.png" },
  { name: "冬瓜丸子汤", meta: "汤菜 · 可提前备", selected: true, image: "miniApp/.preview/assets/recommended-winter-melon-soup.png" },
];

function defs() {
  return `<defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="160%"><feDropShadow dx="0" dy="6" stdDeviation="11" flood-color="#10261A" flood-opacity="0.05"/></filter>
    <filter id="topShadow" x="-10%" y="-50%" width="120%" height="180%"><feDropShadow dx="0" dy="-4" stdDeviation="8" flood-color="#163322" flood-opacity="0.06"/></filter>
    <linearGradient id="green" x1="0" x2="1"><stop offset="0" stop-color="#20B866"/><stop offset="1" stop-color="#0A8E52"/></linearGradient>
  </defs>`;
}

function chrome() {
  return `
  <text x="44" y="39" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="#151F1A">9:41</text>
  <g fill="#151F1A"><rect x="326" y="32" width="4" height="8" rx="1"/><rect x="332" y="28" width="4" height="12" rx="1"/><rect x="338" y="24" width="4" height="16" rx="1"/>
    <path d="M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0" fill="none" stroke="#151F1A" stroke-width="2" stroke-linecap="round"/>
    <rect x="373" y="28" width="21" height="11" rx="3" fill="none" stroke="#151F1A" stroke-width="1.6"/><rect x="395" y="31" width="2" height="5" rx="1"/><rect x="376" y="31" width="15" height="5" rx="1"/>
  </g>
  <g><rect x="312" y="60" width="82" height="36" rx="18" fill="#FFFFFF" fill-opacity="0.92" stroke="#E7E9E5"/><circle cx="338" cy="76" r="2.4" fill="#151F1A"/><circle cx="348" cy="76" r="2.4" fill="#151F1A"/><line x1="359" y1="66" x2="359" y2="86" stroke="#E7E9E5"/><circle cx="378" cy="76" r="9" fill="none" stroke="#151F1A" stroke-width="2.6"/></g>
  <path d="M48 79L38 89l10 10M39 89h20" fill="none" stroke="#151F1A" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="800" fill="#151F1A">创建饭局</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">邀请大家一起点菜</text>
  <rect y="224" width="418" height="1016" fill="#F8FAF7"/>`;
}

function sourceButtons() {
  return `
  <rect x="16" y="520" width="186" height="48" rx="8" fill="#FFFFFF" stroke="#B9DFC9"/>
  <g fill="none" stroke="#159B55" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"><rect x="36" y="535" width="17" height="17" rx="2"/><path d="M44.5 535v17M36 543.5h17"/></g>
  <text x="126" y="550" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#159B55">从全部菜品选</text>
  <rect x="216" y="520" width="186" height="48" rx="8" fill="#EAF7EF" stroke="#D8EEE0"/>
  <path d="M237 535c6-2 11 0 13 3v16c-2-3-7-5-13-3zM263 535c-6-2-11 0-13 3v16c2-3 7-5 13-3z" fill="none" stroke="#159B55" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="326" y="550" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#159B55">从菜谱选</text>`;
}

function selectedCards() {
  const positions = [
    { x: 16, y: 584 }, { x: 216, y: 584 },
    { x: 16, y: 744 }, { x: 216, y: 744 },
  ];
  return selectedDishes.map((dish, index) => {
    const { x, y } = positions[index];
    return `
    <g filter="url(#shadow)"><rect x="${x}" y="${y}" width="186" height="146" rx="8" fill="#FFFFFF" stroke="#E9EFEB"/></g>
    <text x="${x + 12}" y="${y + 117}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">${dish.name}</text>
    <text x="${x + 12}" y="${y + 137}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">${dish.meta}</text>`;
  }).join("");
}

function mainPhotoChrome() {
  const positions = [
    { x: 16, y: 584 }, { x: 216, y: 584 },
    { x: 16, y: 744 }, { x: 216, y: 744 },
  ];
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
    ${positions.map(({ x, y }) => `<circle cx="${x + 169}" cy="${y + 17}" r="11" fill="#FFFFFF" fill-opacity="0.94"/><path d="M${x + 165} ${y + 13}l8 8m0-8-8 8" fill="none" stroke="#68756E" stroke-width="1.5" stroke-linecap="round"/>`).join("")}
  </svg>`;
}

function mainContent() {
  return `
  <g filter="url(#shadow)"><rect x="16" y="242" width="386" height="226" rx="8" fill="#FFFFFF" stroke="#E9EFEB"/></g>
  <text x="34" y="277" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="850" fill="#1B241F">饭局信息</text>
  <text x="34" y="330" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="750" fill="#27332D">饭局名称</text>
  <rect x="120" y="299" width="264" height="48" rx="8" fill="#F8FAF8" stroke="#DDE6E0"/>
  <text x="138" y="329" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#1B241F">周六家庭聚餐</text>
  <line x1="34" y1="365" x2="384" y2="365" stroke="#EEF1EE"/>
  <text x="34" y="402" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="750" fill="#27332D">点餐截止</text>
  <rect x="120" y="374" width="264" height="48" rx="8" fill="#F8FAF8" stroke="#DDE6E0"/>
  <path d="M139 388v6M158 388v6M136 397h25M137 391h23v22h-23z" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="176" y="405" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#1B241F">今天 21:00</text>
  <path d="M364 392l6 6-6 6" fill="none" stroke="#9CA7A1" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="120" y="446" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#8B9690">到期后自动停止点餐，不能重新开启</text>

  <text x="16" y="503" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">候选菜</text>
  <rect x="91" y="483" width="68" height="24" rx="12" fill="#EAF7EF"/><text x="125" y="500" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="750" fill="#159B55">已选 4 道</text>
  ${sourceButtons()}
  ${selectedCards()}

  <rect x="16" y="910" width="386" height="64" rx="8" fill="#EFF8F2"/>
  <circle cx="44" cy="942" r="16" fill="#DFF3E7"/>
  <path d="M36 939h16v11H36zM39 936h10l3 3M41 943h6M44 940v6" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="70" y="937" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#1B241F">创建后进入点餐收集</text>
  <text x="70" y="958" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#6F7D76">参与者将从这些候选菜中选择想吃的菜</text>

  <g filter="url(#topShadow)"><rect x="0" y="1076" width="418" height="164" fill="#FFFFFF"/></g>
  <rect x="24" y="1096" width="370" height="52" rx="26" fill="url(#green)"/>
  <path d="M146 1112h20v20h-20zM150 1116h4v4h-4zM158 1116h4v4h-4zM150 1124h4v4h-4zM158 1124h4v4h-4z" fill="none" stroke="#FFFFFF" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="232" y="1129" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#FFFFFF">生成点餐码</text>
  <text x="209" y="1176" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#8B9690">生成后可分享给微信好友</text>
  <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>`;
}

function baseSvg() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${chrome()}${mainContent()}</svg>`;
}

function checkControl(x, y, selected) {
  if (selected) {
    return `<circle cx="${x}" cy="${y}" r="12" fill="#159B55"/><path d="M${x - 6} ${y}l4 4 8-9" fill="none" stroke="#FFFFFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`;
  }
  return `<circle cx="${x}" cy="${y}" r="12" fill="#FFFFFF" stroke="#B9C5BE" stroke-width="1.5"/>`;
}

function sheetHeader(active) {
  return `
    <rect width="418" height="1240" fill="#132119" fill-opacity="0.42"/>
    <path d="M0 374Q0 348 26 348H392Q418 348 418 374V1240H0Z" fill="#FFFFFF"/>
    <rect x="181" y="360" width="56" height="5" rx="2.5" fill="#DDE5E0"/>
    <text x="24" y="408" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">选择候选菜</text>
    <text x="24" y="433" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">只展示可用于饭局的可用菜品</text>
    <circle cx="378" cy="400" r="16" fill="#F2F5F3"/><path d="M373 395l10 10m0-10-10 10" fill="none" stroke="#68746D" stroke-width="1.7" stroke-linecap="round"/>
    <rect x="24" y="454" width="370" height="42" rx="8" fill="#F1F5F2"/>
    <rect x="${active === "all" ? 28 : 211}" y="458" width="179" height="34" rx="6" fill="#FFFFFF"/>
    <text x="117" y="481" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="${active === "all" ? 800 : 600}" fill="${active === "all" ? "#159B55" : "#7B8982"}">全部菜品</text>
    <text x="301" y="481" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="${active === "recipe" ? 800 : 600}" fill="${active === "recipe" ? "#159B55" : "#7B8982"}">从菜谱选</text>`;
}

function candidateRows(candidates, startY) {
  return candidates.map((candidate, index) => {
    const top = startY + index * 98;
    return `
      <text x="118" y="${top + 33}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#1B241F">${candidate.name}</text>
      <text x="118" y="${top + 57}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">${candidate.meta}</text>
      ${checkControl(362, top + 40, candidate.selected)}
      <line x1="118" y1="${top + 84}" x2="394" y2="${top + 84}" stroke="#EEF1EE"/>`;
  }).join("");
}

function fixedConfirm(count) {
  return `
    <g filter="url(#topShadow)"><rect x="0" y="1048" width="418" height="192" fill="#FFFFFF"/></g>
    <text x="24" y="1080" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="750" fill="#526159">当前已选 ${count} 道菜</text>
    <rect x="24" y="1096" width="370" height="52" rx="26" fill="url(#green)"/>
    <text x="209" y="1129" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#FFFFFF">确认候选菜</text>
    <text x="209" y="1176" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#8B9690">重复菜品会自动合并，不会重复添加</text>
    <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>`;
}

function allSheetSvg() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${sheetHeader("all")}
    <rect x="24" y="514" width="370" height="42" rx="21" fill="#F3F6F4"/>
    <circle cx="46" cy="535" r="7" fill="none" stroke="#95A099" stroke-width="1.6"/><path d="M51 540l5 5" fill="none" stroke="#95A099" stroke-width="1.6" stroke-linecap="round"/>
    <text x="66" y="540" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#98A29C">搜索菜名</text>
    ${candidateRows(allCandidates, 574)}${fixedConfirm(4)}
  </svg>`;
}

function recipeSheetSvg() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${sheetHeader("recipe")}
    <rect x="24" y="514" width="370" height="60" rx="8" fill="#EFF8F2"/>
    <path d="M42 531c7-2 13 0 16 4v19c-3-4-9-6-16-4zM74 531c-7-2-13 0-16 4v19c3-4 9-6 16-4z" fill="none" stroke="#159B55" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
    <text x="94" y="540" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">工作日晚餐</text>
    <text x="94" y="560" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">4 道菜 · 快手少油</text>
    <text x="344" y="550" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">更换菜谱</text><path d="M365 543l5 5-5 5" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    ${candidateRows(recipeCandidates, 590)}${fixedConfirm(4)}
  </svg>`;
}

async function roundedImage(source, width, height, radius) {
  const mask = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`);
  return sharp(source).resize(width, height, { fit: "cover", position: "centre" }).composite([{ input: mask, blend: "dest-in" }]).png().toBuffer();
}

async function prepareHero() {
  const scene = await sharp(heroSource).resize(340, 184, { fit: "fill" }).modulate({ brightness: 1.035, saturation: 0.84 }).png().toBuffer();
  const background = Buffer.from(`<svg width="418" height="224" xmlns="http://www.w3.org/2000/svg"><rect width="418" height="224" fill="#FDF8F0"/><rect y="190" width="418" height="34" fill="#F9E2BE"/></svg>`);
  return sharp(background).composite([{ input: scene, left: 8, top: 40 }]).png().toBuffer();
}

async function render() {
  const hero = await prepareHero();
  const mainPhotos = await Promise.all(selectedDishes.map((dish) => roundedImage(dish.image, 178, 88, 6)));
  const mainPositions = [{ left: 20, top: 588 }, { left: 220, top: 588 }, { left: 20, top: 748 }, { left: 220, top: 748 }];
  const base = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite([
    { input: hero, left: 0, top: 0 },
    { input: Buffer.from(baseSvg()), left: 0, top: 0 },
    ...mainPhotos.map((input, index) => ({ input, ...mainPositions[index] })),
    { input: Buffer.from(mainPhotoChrome()), left: 0, top: 0 },
  ]).png().toBuffer();

  const allPhotos = await Promise.all(allCandidates.map((dish) => roundedImage(dish.image, 70, 70, 8)));
  const recipePhotos = await Promise.all(recipeCandidates.map((dish) => roundedImage(dish.image, 70, 70, 8)));
  const all = await sharp(base).composite([
    { input: Buffer.from(allSheetSvg()), left: 0, top: 0 },
    ...allPhotos.map((input, index) => ({ input, left: 36, top: 580 + index * 98 })),
  ]).png().toBuffer();
  const recipe = await sharp(base).composite([
    { input: Buffer.from(recipeSheetSvg()), left: 0, top: 0 },
    ...recipePhotos.map((input, index) => ({ input, left: 36, top: 596 + index * 98 })),
  ]).png().toBuffer();

  fs.writeFileSync("miniApp/.preview/13-meal_create-preview.png", base);
  fs.writeFileSync("miniApp/.preview/13-meal_create-all-dishes-preview.png", all);
  fs.writeFileSync("miniApp/.preview/13-meal_create-recipes-preview.png", recipe);
  console.log(JSON.stringify({ outputs: ["miniApp/.preview/13-meal_create-preview.png", "miniApp/.preview/13-meal_create-all-dishes-preview.png", "miniApp/.preview/13-meal_create-recipes-preview.png"], width: WIDTH, height: HEIGHT }, null, 2));
}

render().catch((error) => { console.error(error); process.exit(1); });
