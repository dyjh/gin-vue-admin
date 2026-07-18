const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/profile-cat-spoon-v2.png";
const dishes = [
  { name: "香菇鸡腿饭", meta: "主食 · 家常菜", votes: "3 人想吃", image: "miniApp/.preview/assets/dish-detail-cover-realistic.png" },
  { name: "番茄炒蛋", meta: "下饭 · 快手", votes: "2 人想吃", image: "miniApp/.preview/assets/recommended-tomato-egg.png" },
  { name: "蒜蓉西兰花", meta: "素菜 · 少油", votes: "2 人想吃", image: "miniApp/.preview/assets/recommended-garlic-broccoli.png" },
  { name: "冬瓜丸子汤", meta: "汤菜 · 适合多人", votes: "1 人想吃", image: "miniApp/.preview/assets/recommended-winter-melon-soup.png" },
];

function defs() {
  return `<defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="160%"><feDropShadow dx="0" dy="7" stdDeviation="12" flood-color="#10261A" flood-opacity="0.055"/></filter>
    <filter id="topShadow" x="-10%" y="-50%" width="120%" height="180%"><feDropShadow dx="0" dy="-4" stdDeviation="8" flood-color="#163322" flood-opacity="0.06"/></filter>
    <linearGradient id="green" x1="0" x2="1"><stop offset="0" stop-color="#20B866"/><stop offset="1" stop-color="#0A8E52"/></linearGradient>
  </defs>`;
}

function chrome(title, subtitle) {
  return `
  <text x="44" y="39" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="#151F1A">9:41</text>
  <g fill="#151F1A"><rect x="326" y="32" width="4" height="8" rx="1"/><rect x="332" y="28" width="4" height="12" rx="1"/><rect x="338" y="24" width="4" height="16" rx="1"/>
    <path d="M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0" fill="none" stroke="#151F1A" stroke-width="2" stroke-linecap="round"/>
    <rect x="373" y="28" width="21" height="11" rx="3" fill="none" stroke="#151F1A" stroke-width="1.6"/><rect x="395" y="31" width="2" height="5" rx="1"/><rect x="376" y="31" width="15" height="5" rx="1"/>
  </g>
  <g><rect x="312" y="60" width="82" height="36" rx="18" fill="#FFFFFF" fill-opacity="0.92" stroke="#E7E9E5"/><circle cx="338" cy="76" r="2.4" fill="#151F1A"/><circle cx="348" cy="76" r="2.4" fill="#151F1A"/><line x1="359" y1="66" x2="359" y2="86" stroke="#E7E9E5"/><circle cx="378" cy="76" r="9" fill="none" stroke="#151F1A" stroke-width="2.6"/></g>
  <path d="M48 79L38 89l10 10M39 89h20" fill="none" stroke="#151F1A" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="800" fill="#151F1A">${title}</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">${subtitle}</text>
  <rect y="224" width="418" height="1016" fill="#F8FAF7"/>`;
}

function statusChip(label, mode = "green") {
  const palette = mode === "green"
    ? { bg: "#EAF7EF", dot: "#20B866", text: "#159B55" }
    : mode === "red"
      ? { bg: "#FFF0EF", dot: "#D45B53", text: "#C9534E" }
      : { bg: "#F0F3F1", dot: "#89958F", text: "#6F7D76" };
  return `<rect x="310" y="244" width="76" height="26" rx="13" fill="${palette.bg}"/><circle cx="326" cy="257" r="3.5" fill="${palette.dot}"/><text x="357" y="261" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="750" fill="${palette.text}">${label}</text>`;
}

function mealInfo(label = "收集中", mode = "green", detail = "今天 21:00 截止") {
  return `
  <rect x="0" y="224" width="418" height="126" fill="#FFFFFF"/>
  <text x="24" y="264" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="850" fill="#1B241F">周六家庭聚餐</text>
  ${statusChip(label, mode)}
  <circle cx="35" cy="300" r="10" fill="#EAF7EF"/><text x="35" y="304" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="850" fill="#159B55">厨</text>
  <text x="52" y="304" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">厨房小记发起</text>
  <circle cx="198" cy="299" r="7" fill="none" stroke="#8A9690" stroke-width="1.3"/><path d="M198 295v5l3 2" fill="none" stroke="#8A9690" stroke-width="1.3" stroke-linecap="round"/>
  <text x="212" y="304" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">${detail}</text>
  <text x="24" y="332" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#8B9690">4 道候选菜 · 已有 4 人加入</text>
  <rect x="0" y="350" width="418" height="8" fill="#F8FAF7"/>`;
}

function codeBoxes() {
  const digits = ["7", "3", "6", "4", "2", "8"];
  return digits.map((digit, index) => {
    const x = 42 + index * 57;
    return `<rect x="${x}" y="350" width="46" height="58" rx="8" fill="#F3F8F4" stroke="#BFDCCA"/><text x="${x + 23}" y="389" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="25" font-weight="850" fill="#173326">${digit}</text>`;
  }).join("");
}

function joinSvg() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${chrome("加入饭局", "输入点餐码，一起选菜")}
    <g filter="url(#shadow)"><rect x="16" y="246" width="386" height="360" rx="8" fill="#FFFFFF" stroke="#E9EFEB"/></g>
    <circle cx="54" cy="286" r="20" fill="#EAF7EF"/><g fill="none" stroke="#159B55" stroke-width="1.5"><rect x="45" y="277" width="6" height="6" rx="1"/><rect x="57" y="277" width="6" height="6" rx="1"/><rect x="45" y="289" width="6" height="6" rx="1"/><rect x="57" y="289" width="6" height="6" rx="1"/></g>
    <text x="86" y="284" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">输入 6 位点餐码</text>
    <text x="86" y="306" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">点餐码由饭局发起人分享</text>
    ${codeBoxes()}
    <rect x="34" y="430" width="350" height="42" rx="8" fill="#EFF8F2"/>
    <circle cx="55" cy="451" r="8" fill="#20B866"/><path d="M51 451l3 3 5-6" fill="none" stroke="#FFFFFF" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
    <text x="73" y="455" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">已找到“周六家庭聚餐”</text>
    <rect x="34" y="492" width="350" height="54" rx="27" fill="url(#green)"/>

    <text x="209" y="526" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#FFFFFF">加入饭局</text>
    <text x="209" y="574" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#8B9690">加入后可以多选想吃的菜，关闭前还能修改</text>

    <text x="24" y="656" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="850" fill="#1B241F">加入后可以</text>
    <rect x="16" y="676" width="386" height="154" rx="8" fill="#FFFFFF" stroke="#E9EFEB"/>
    <circle cx="48" cy="714" r="16" fill="#EAF7EF"/><path d="M41 714l5 5 9-11" fill="none" stroke="#159B55" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
    <text x="78" y="710" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">多选自己想吃的菜</text>
    <text x="78" y="730" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">你的选择会帮助发起人准备菜单</text>
    <line x1="78" y1="752" x2="384" y2="752" stroke="#EEF1EE"/>
    <circle cx="48" cy="789" r="16" fill="#EAF7EF"/><circle cx="48" cy="789" r="7" fill="none" stroke="#159B55" stroke-width="1.4"/><path d="M48 785v5l3 2" fill="none" stroke="#159B55" stroke-width="1.4" stroke-linecap="round"/>
    <text x="78" y="785" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">点餐关闭前随时修改</text>
    <text x="78" y="805" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">重复输入点餐码会回到已有选择</text>
    <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>
  </svg>`;
}

function choiceControl(y, selected, disabled) {
  const cy = y + 54;
  if (disabled) {
    return selected
      ? `<circle cx="376" cy="${cy}" r="14" fill="#DDE3DF"/><path d="M369 ${cy}l5 5 9-11" fill="none" stroke="#7F8B85" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>`
      : `<circle cx="376" cy="${cy}" r="14" fill="#F6F7F6" stroke="#DDE3DF"/><path d="M371 ${cy}h10" fill="none" stroke="#AAB3AE" stroke-width="1.6" stroke-linecap="round"/>`;
  }
  return selected
    ? `<circle cx="376" cy="${cy}" r="14" fill="#16B96F"/><path d="M369 ${cy}l5 5 9-11" fill="none" stroke="#FFFFFF" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"/>`
    : `<circle cx="376" cy="${cy}" r="14" fill="#FFFFFF" stroke="#16A863" stroke-width="1.5"/><path d="M370 ${cy}h12M376 ${cy - 6}v12" fill="none" stroke="#16A863" stroke-width="1.6" stroke-linecap="round"/>`;
}

function categoryRail(disabled = false) {
  const categories = [
    { name: "全部", count: 4, selected: true },
    { name: "主食", count: 1 },
    { name: "家常菜", count: 1 },
    { name: "素菜", count: 1 },
    { name: "汤菜", count: 1 },
  ];
  return categories.map((category, index) => {
    const y = 430 + index * 64;
    const textColor = disabled ? "#707B75" : category.selected ? "#159B55" : "#48554E";
    const countColor = disabled ? "#A0AAA5" : category.selected ? "#6E9E82" : "#98A29D";
    return `${category.selected ? `<rect x="0" y="${y}" width="96" height="60" fill="${disabled ? "#ECEFED" : "#E9F6ED"}"/><rect x="0" y="${y + 15}" width="3" height="30" rx="1.5" fill="${disabled ? "#AAB3AE" : "#16B96F"}"/>` : ""}<text x="48" y="${y + 25}" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="${category.selected ? 850 : 700}" fill="${textColor}">${category.name}</text><text x="48" y="${y + 43}" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="9" fill="${countColor}">${category.count} 道</text>`;
  }).join("");
}

function dishRows(disabled = false) {
  return dishes.map((dish, index) => {
    const y = 468 + index * 112;
    const selected = index === 0 || index === 2;
    return `${index > 0 ? `<line x1="112" y1="${y}" x2="390" y2="${y}" stroke="#EEF1EE"/>` : ""}<text x="198" y="${y + 35}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="850" fill="#1B241F">${dish.name}</text><text x="198" y="${y + 57}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">${dish.meta}</text><text x="198" y="${y + 82}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="9" fill="${disabled ? "#8C9691" : "#159B55"}">${dish.votes}</text>${choiceControl(y, selected, disabled)}`;
  }).join("");
}

function orderingPanel(disabled = false) {
  return `
    <rect x="0" y="358" width="418" height="650" fill="#FFFFFF"/>
    <text x="24" y="390" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">候选菜</text>
    <text x="394" y="389" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="${disabled ? "#7B8982" : "#159B55"}">${disabled ? "我的选择 2 道" : "已选 2 道"}</text>
    <text x="24" y="412" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">${disabled ? "点餐已结束，仅查看本次选择" : "按分类浏览，可以多选，截止前还能修改"}</text>
    <rect x="0" y="430" width="96" height="578" fill="#F5F7F5"/>
    ${categoryRail(disabled)}
    <text x="112" y="455" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#1B241F">全部菜品</text>
    <text x="390" y="455" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#8A9690">4 道</text>
    <line x1="112" y1="468" x2="390" y2="468" stroke="#EEF1EE"/>
    ${dishRows(disabled)}
  `;
}

function voteSvg(closed = false) {
  const label = closed ? "已关闭" : "收集中";
  const mode = closed ? "neutral" : "green";
  const detail = closed ? "今天 21:00 到期关闭" : "今天 21:00 截止";
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${chrome("参与点菜", closed ? "点餐已结束，查看你的选择" : "选出你最想吃的菜")}
    ${mealInfo(label, mode, detail)}
    ${orderingPanel(closed)}
    <rect x="0" y="1008" width="418" height="232" fill="#FFFFFF"/>
    ${closed
      ? `<text x="24" y="1083" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="850" fill="#1B241F">我的选择 · 2 道</text>
        <text x="24" y="1104" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="9" fill="#7B8982">本次点餐已结束</text>
        <rect x="226" y="1058" width="168" height="54" rx="27" fill="#FFFFFF" stroke="#C8D2CC"/>
        <path d="M270 1078l-6 7 6 7M265 1085h14" fill="none" stroke="#59665F" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
        <text x="330" y="1091" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#59665F">返回首页</text>`
      : `<text x="24" y="1083" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="850" fill="#1B241F">已选 2 道</text>
        <text x="24" y="1104" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="9" fill="#6F7D76">关闭前仍可修改</text>
        <rect x="226" y="1058" width="168" height="54" rx="27" fill="url(#green)"/>
        <path d="M264 1081l5 5 9-11" fill="none" stroke="#FFFFFF" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"/>
        <text x="331" y="1091" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="850" fill="#FFFFFF">保存点菜</text>`}
    <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>
  </svg>`;
}

function cancelledSvg() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${chrome("参与点菜", "查看饭局当前状态")}
    ${mealInfo("已取消", "red", "发起人已取消")}
    <g filter="url(#shadow)"><rect x="16" y="374" width="386" height="390" rx="8" fill="#FFFFFF" stroke="#E9EFEB"/></g>
    <circle cx="209" cy="470" r="38" fill="#FFF0EF"/><circle cx="209" cy="470" r="16" fill="none" stroke="#C9534E" stroke-width="2"/><path d="M202 463l14 14M216 463l-14 14" fill="none" stroke="#C9534E" stroke-width="2" stroke-linecap="round"/>
    <text x="209" y="548" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#1B241F">饭局已取消</text>
    <text x="209" y="580" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6F7D76">发起人取消了这次饭局</text>
    <text x="209" y="604" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6F7D76">本次点餐结果不再保留</text>
    <line x1="54" y1="642" x2="364" y2="642" stroke="#EEF1EE"/>
    <rect x="50" y="674" width="318" height="50" rx="25" fill="#FFFFFF" stroke="#C8D2CC"/>
    <path d="M157 693l-6 6 6 6M152 699h13" fill="none" stroke="#59665F" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
    <text x="234" y="706" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#59665F">返回首页</text>
    <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>
  </svg>`;
}

async function roundedImage(source, width, height, radius) {
  const mask = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`);
  return sharp(source).resize(width, height, { fit: "cover", position: "centre" }).composite([{ input: mask, blend: "dest-in" }]).png().toBuffer();
}

async function prepareHero() {
  const scene = await sharp(heroSource).resize(585, 313, { fit: "fill" }).extract({ left: 67, top: 69, width: 418, height: 224 }).modulate({ brightness: 1.03, saturation: 0.9 }).png().toBuffer();
  const leftPlant = await sharp(heroSource).extract({ left: 0, top: 520, width: 360, height: 350 }).resize(82, 84, { fit: "fill" }).modulate({ brightness: 1.03, saturation: 0.9 }).png().toBuffer();
  const rightPlant = await sharp(heroSource).extract({ left: 1400, top: 520, width: 314, height: 350 }).resize(72, 80, { fit: "fill" }).modulate({ brightness: 1.03, saturation: 0.9 }).png().toBuffer();
  return sharp(scene).composite([
    { input: leftPlant, left: 18, top: 133 },
    { input: rightPlant, left: 346, top: 135 },
  ]).png().toBuffer();
}

async function renderScreen(hero, svg, dishPhotos = []) {
  return sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite([
    { input: hero, left: 0, top: 0 },
    { input: Buffer.from(svg), left: 0, top: 0 },
    ...dishPhotos.map((input, index) => ({ input, left: 114, top: 484 + index * 112 })),
  ]).png().toBuffer();
}

async function render() {
  const hero = await prepareHero();
  const dishPhotos = await Promise.all(dishes.map((dish) => roundedImage(dish.image, 72, 72, 7)));
  const join = await renderScreen(hero, joinSvg());
  const collecting = await renderScreen(hero, voteSvg(false), dishPhotos);
  const closed = await renderScreen(hero, voteSvg(true), dishPhotos);
  const cancelled = await renderScreen(hero, cancelledSvg());
  const outputs = [
    ["miniApp/.preview/15-meal_vote-join-preview.png", join],
    ["miniApp/.preview/15-meal_vote-preview.png", collecting],
    ["miniApp/.preview/15-meal_vote-closed-preview.png", closed],
    ["miniApp/.preview/15-meal_vote-cancelled-preview.png", cancelled],
  ];
  outputs.forEach(([file, buffer]) => fs.writeFileSync(file, buffer));
  console.log(JSON.stringify({ outputs: outputs.map(([file]) => file), width: WIDTH, height: HEIGHT }, null, 2));
}

render().catch((error) => { console.error(error); process.exit(1); });
