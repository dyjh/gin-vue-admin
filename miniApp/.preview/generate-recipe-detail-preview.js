const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/recipe-detail-cat-bookmark.png";

const dishes = [
  { name: "青椒牛柳", meta: "家常菜", tag: "少油", time: "20 分钟", image: "miniApp/.preview/assets/recommended-green-pepper-beef.png" },
  { name: "番茄炒蛋", meta: "家常菜", tag: "下饭", time: "20 分钟", image: "miniApp/.preview/assets/recommended-tomato-egg.png" },
  { name: "蒜蓉西兰花", meta: "素菜", tag: "清淡", time: "15 分钟", image: "miniApp/.preview/assets/recommended-garlic-broccoli.png" },
  { name: "冬瓜丸子汤", meta: "汤菜", tag: "可提前备", time: "35 分钟", image: "miniApp/.preview/assets/recommended-winter-melon-soup.png" },
];

const candidates = [
  { name: "香菇鸡腿饭", meta: "主食 · 2 人份", state: "selected", image: "miniApp/.preview/assets/dish-detail-cover-realistic.png" },
  { name: "宫保鸡丁", meta: "下饭 · 家常菜", state: "selected", image: "miniApp/.preview/assets/what-to-eat-kung-pao-chicken.png" },
  { name: "番茄炒蛋", meta: "下饭 · 家常菜", state: "added", image: "miniApp/.preview/assets/recommended-tomato-egg.png" },
];

function defs() {
  return `<defs>
    <filter id="shadow" x="-10%" y="-20%" width="120%" height="155%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261A" flood-opacity="0.055"/>
    </filter>
    <linearGradient id="green" x1="0" x2="1">
      <stop offset="0" stop-color="#20B866"/><stop offset="1" stop-color="#0A8E52"/>
    </linearGradient>
  </defs>`;
}

function header() {
  return `
  <text x="44" y="39" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="#151F1A">9:41</text>
  <g fill="#151F1A">
    <rect x="326" y="32" width="4" height="8" rx="1"/><rect x="332" y="28" width="4" height="12" rx="1"/><rect x="338" y="24" width="4" height="16" rx="1"/>
    <path d="M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0" fill="none" stroke="#151F1A" stroke-width="2" stroke-linecap="round"/>
    <rect x="373" y="28" width="21" height="11" rx="3" fill="none" stroke="#151F1A" stroke-width="1.6"/><rect x="395" y="31" width="2" height="5" rx="1"/><rect x="376" y="31" width="15" height="5" rx="1"/>
  </g>
  <g><rect x="312" y="60" width="82" height="36" rx="18" fill="#FFFFFF" fill-opacity="0.9" stroke="#E7E9E5"/>
    <circle cx="338" cy="76" r="2.4" fill="#151F1A"/><circle cx="348" cy="76" r="2.4" fill="#151F1A"/><line x1="359" y1="66" x2="359" y2="86" stroke="#E7E9E5"/><circle cx="378" cy="76" r="9" fill="none" stroke="#151F1A" stroke-width="2.6"/>
  </g>
  <path d="M39 84H23m0 0 7-7m-7 7 7 7" fill="none" stroke="#151F1A" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="54" y="92" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="850" fill="#151F1A">菜谱详情</text>
  <text x="54" y="119" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6F7D76">整理常做的搭配，也可以继续调整</text>
  <rect y="224" width="${WIDTH}" height="${HEIGHT - 224}" fill="#F8FAF7"/>`;
}

function infoCard() {
  return `
  <g filter="url(#shadow)"><rect x="24" y="246" width="370" height="152" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <text x="43" y="283" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="850" fill="#1B241F">工作日晚餐</text>
  <rect x="297" y="260" width="78" height="30" rx="15" fill="#F1F8F3"/>
  <path d="M315 275l4-4 4 4-4 4zM319 271v8" fill="none" stroke="#159B55" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="344" y="280" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="750" fill="#159B55">编辑</text>
  <rect x="43" y="310" width="332" height="68" rx="12" fill="#F2F8F3"/>
  <text x="59" y="333" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">菜谱备注</text>
  <text x="59" y="357" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#526159">快手、少油，30 分钟内，适合工作日下班后准备。</text>`;
}

function dishCards(swipedIndex = -1) {
  return dishes.map((dish, index) => {
    const top = 468 + index * 137;
    const swiped = index === swipedIndex;
    const shift = swiped ? -74 : 0;
    return `
  <clipPath id="rowClip${index}"><rect x="24" y="${top}" width="370" height="121" rx="15"/></clipPath>
  <g clip-path="url(#rowClip${index})">
    ${swiped ? `<rect x="24" y="${top}" width="370" height="121" rx="15" fill="#D9534F"/><path d="M352 ${top + 40}h14M356 ${top + 35}h6M354 ${top + 40}l1 17h8l1-17" fill="none" stroke="#FFFFFF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/><text x="359" y="${top + 82}" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#FFFFFF">移除</text>` : ""}
    <g transform="translate(${shift} 0)">
      <g filter="url(#shadow)"><rect x="24" y="${top}" width="370" height="121" rx="15" fill="#FFFFFF" stroke="#EEF1EB"/></g>
      <text x="146" y="${top + 34}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#1B241F">${dish.name}</text>
      <text x="146" y="${top + 57}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">${dish.meta}</text>
      <path d="M149 ${top + 80}c8-5 13 1 8 8-6 6-12 1-8-8zm1 7 8-7" fill="none" stroke="#159B55" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>
      <text x="167" y="${top + 91}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="700" fill="#159B55">${dish.tag}</text>
      <circle cx="230" cy="${top + 86}" r="7" fill="none" stroke="#9AA49E" stroke-width="1.3"/><path d="M230 ${top + 82}v5l3 2" fill="none" stroke="#9AA49E" stroke-width="1.3" stroke-linecap="round"/>
      <text x="243" y="${top + 91}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">${dish.time}</text>
      <path d="M368 ${top + 54}l6 6-6 6" fill="none" stroke="#A6B0AA" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
    </g>
  </g>`;
  }).join("");
}

function actions() {
  return `
  <rect x="24" y="1019" width="246" height="46" rx="23" fill="url(#green)"/>
  <path d="M54 1034h14v16H54zM58 1030h6v5h-6M72 1037l4 4 7-8" fill="none" stroke="#FFFFFF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="170" y="1048" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#FFFFFF">用于饭局选菜</text>
  <rect x="282" y="1019" width="112" height="46" rx="23" fill="#FFFFFF" stroke="#E8B8B5"/>
  <path d="M299 1036h12M303 1032h5M301 1036l1 14h8l1-14" fill="none" stroke="#C84D49" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="351" y="1048" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="750" fill="#C84D49">删除菜谱</text>
  <text x="209" y="1100" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#98A29C">删除菜谱不会删除菜品库里的菜品</text>
  <rect x="147" y="1210" width="124" height="5" rx="2.5" fill="#202521" fill-opacity="0.74"/>`;
}

function baseSvg(swipedIndex = -1) {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
    ${defs()}${header()}${infoCard()}
    <text x="24" y="444" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">菜品清单</text>
    <text x="118" y="443" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">4 道菜</text>
    <rect x="276" y="414" width="118" height="40" rx="20" fill="#FFFFFF" stroke="#B9DFC9"/>
    <path d="M296 427v14M289 434h14" fill="none" stroke="#159B55" stroke-width="2" stroke-linecap="round"/>
    <text x="347" y="440" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#159B55">添加菜品</text>
    ${dishCards(swipedIndex)}${actions()}
  </svg>`;
}

function confirmOverlay() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
    <rect width="${WIDTH}" height="${HEIGHT}" fill="#162019" fill-opacity="0.42"/>
    <path d="M0 918A26 26 0 0 1 26 892H392A26 26 0 0 1 418 918V1240H0Z" fill="#FFFFFF"/>
    <rect x="178" y="908" width="62" height="5" rx="2.5" fill="#D9DFDB"/>
    <circle cx="209" cy="965" r="30" fill="#FFF1F0"/>
    <path d="M196 957h26M201 957l2 21h12l2-21M205 950h8" fill="none" stroke="#C84D49" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/>
    <text x="209" y="1025" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#1B241F">删除这份菜谱？</text>
    <text x="209" y="1056" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#748179">删除后无法恢复，但不会删除菜品库里的菜品。</text>
    <rect x="24" y="1092" width="174" height="52" rx="26" fill="#F2F5F3"/>
    <text x="111" y="1125" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="750" fill="#5F6C65">取消</text>
    <rect x="220" y="1092" width="174" height="52" rx="26" fill="#D9534F"/>
    <text x="307" y="1125" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#FFFFFF">确认删除</text>
    <rect x="147" y="1210" width="124" height="5" rx="2.5" fill="#202521" fill-opacity="0.74"/>
  </svg>`;
}

function addDishesOverlay() {
  const rows = candidates.map((candidate, index) => {
    const top = 660 + index * 98;
    const trailing = candidate.state === "added"
      ? `<rect x="322" y="${top + 24}" width="58" height="28" rx="14" fill="#F1F3F2"/><text x="351" y="${top + 43}" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#8B9690">已添加</text>`
      : `<circle cx="352" cy="${top + 38}" r="12" fill="#159B55"/><path d="M346 ${top + 38}l4 4 8-9" fill="none" stroke="#FFFFFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`;
    return `<text x="126" y="${top + 30}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#1B241F">${candidate.name}</text>
      <text x="126" y="${top + 54}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">${candidate.meta}</text>
      ${trailing}<line x1="126" y1="${top + 82}" x2="380" y2="${top + 82}" stroke="#EEF1EE"/>`;
  }).join("");
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
    <rect width="${WIDTH}" height="${HEIGHT}" fill="#162019" fill-opacity="0.42"/>
    <path d="M0 506A26 26 0 0 1 26 480H392A26 26 0 0 1 418 506V1240H0Z" fill="#FFFFFF"/>
    <rect x="178" y="496" width="62" height="5" rx="2.5" fill="#D9DFDB"/>
    <text x="24" y="548" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#1B241F">添加菜品</text>
    <text x="24" y="574" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#748179">从菜品库选择，可同时添加多道菜</text>
    <circle cx="378" cy="544" r="16" fill="#F2F5F3"/><path d="M373 539l10 10m0-10-10 10" fill="none" stroke="#68746D" stroke-width="1.7" stroke-linecap="round"/>
    <rect x="24" y="596" width="370" height="44" rx="22" fill="#F3F6F4"/>
    <circle cx="46" cy="618" r="7" fill="none" stroke="#95A099" stroke-width="1.6"/><path d="M51 623l5 5" fill="none" stroke="#95A099" stroke-width="1.6" stroke-linecap="round"/>
    <text x="66" y="623" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#98A29C">搜索菜名</text>
    ${rows}
    <line x1="0" y1="1086" x2="418" y2="1086" stroke="#E9EEEA"/>
    <text x="24" y="1114" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="750" fill="#526159">已选 2 道菜</text>
    <rect x="24" y="1128" width="370" height="52" rx="26" fill="#159B55"/>
    <text x="209" y="1161" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#FFFFFF">添加 2 道菜</text>
    <rect x="147" y="1210" width="124" height="5" rx="2.5" fill="#202521" fill-opacity="0.74"/>
  </svg>`;
}

async function roundedImage(path, width, height, radius) {
  const mask = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`);
  return sharp(path).resize(width, height, { fit: "cover", position: "centre" }).composite([{ input: mask, blend: "dest-in" }]).png().toBuffer();
}

async function build() {
  const hero = await sharp(heroSource).resize(438, 224, { fit: "fill" }).extract({ left: 20, top: 0, width: 418, height: 224 }).modulate({ brightness: 1.04, saturation: 0.88 }).png().toBuffer();
  const photoBuffers = await Promise.all(dishes.map((dish) => roundedImage(dish.image, 92, 92, 13)));
  const candidatePhotoBuffers = await Promise.all(candidates.map((candidate) => roundedImage(candidate.image, 70, 70, 12)));
  const photoLayers = (swipedIndex = -1) => photoBuffers.map((input, index) => ({ input, left: 40 + (index === swipedIndex ? -74 : 0), top: 482 + index * 137 }));
  const base = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } })
    .composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(baseSvg()), left: 0, top: 0 }, ...photoLayers()])
    .png().toBuffer();
  const swiped = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } })
    .composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(baseSvg(1)), left: 0, top: 0 }, ...photoLayers(1)])
    .png().toBuffer();
  const addDishes = await sharp(base).composite([
    { input: Buffer.from(addDishesOverlay()), left: 0, top: 0 },
    ...candidatePhotoBuffers.map((input, index) => ({ input, left: 40, top: 668 + index * 98 })),
  ]).png().toBuffer();  const confirm = await sharp(base).composite([{ input: Buffer.from(confirmOverlay()), left: 0, top: 0 }]).png().toBuffer();
  fs.writeFileSync("miniApp/.preview/10-recipe_detail-preview.png", base);
  fs.writeFileSync("miniApp/.preview/10-recipe_detail-swipe-remove-preview.png", swiped);
  fs.writeFileSync("miniApp/.preview/10-recipe_detail-add-dishes-preview.png", addDishes);
  fs.writeFileSync("miniApp/.preview/10-recipe_detail-delete-confirm-preview.png", confirm);
  console.log(JSON.stringify({ outputs: ["miniApp/.preview/10-recipe_detail-preview.png", "miniApp/.preview/10-recipe_detail-swipe-remove-preview.png", "miniApp/.preview/10-recipe_detail-add-dishes-preview.png", "miniApp/.preview/10-recipe_detail-delete-confirm-preview.png"], width: WIDTH, height: HEIGHT }, null, 2));
}

build().catch((error) => { console.error(error); process.exit(1); });
