const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/meal-invite-cat-code-v3.png";
const dishes = [
  { name: "香菇鸡腿饭", image: "miniApp/.preview/assets/dish-detail-cover-realistic.png" },
  { name: "番茄炒蛋", image: "miniApp/.preview/assets/recommended-tomato-egg.png" },
  { name: "蒜蓉西兰花", image: "miniApp/.preview/assets/recommended-garlic-broccoli.png" },
  { name: "冬瓜丸子汤", image: "miniApp/.preview/assets/recommended-winter-melon-soup.png" },
];

function defs() {
  return `<defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="160%"><feDropShadow dx="0" dy="7" stdDeviation="12" flood-color="#10261A" flood-opacity="0.055"/></filter>
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
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="800" fill="#151F1A">饭局点餐码</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">邀请大家来点菜</text>
  <rect y="224" width="418" height="1016" fill="#F8FAF7"/>`;
}

function codeBoxes() {
  const digits = ["7", "3", "6", "4", "2", "8"];
  return digits.map((digit, index) => {
    const x = 42 + index * 57;
    return `<rect x="${x}" y="441" width="46" height="58" rx="8" fill="#F3F8F4" stroke="#CFE4D7"/><text x="${x + 23}" y="480" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="25" font-weight="850" fill="#173326">${digit}</text>`;
  }).join("");
}

function mainContent() {
  return `
  <g filter="url(#shadow)"><rect x="16" y="242" width="386" height="92" rx="8" fill="#FFFFFF" stroke="#E9EFEB"/></g>
  <text x="34" y="276" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="850" fill="#1B241F">周六家庭聚餐</text>
  <rect x="303" y="255" width="80" height="28" rx="14" fill="#EAF7EF"/><circle cx="319" cy="269" r="4" fill="#20B866"/><text x="351" y="273" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">收集中</text>
  <rect x="34" y="295" width="26" height="26" rx="7" fill="#EAF7EF"/>
  <g fill="none" stroke="#159B55" stroke-width="1.3">
    <rect x="40" y="301" width="5" height="5" rx="1"/>
    <rect x="49" y="301" width="5" height="5" rx="1"/>
    <rect x="40" y="310" width="5" height="5" rx="1"/>
    <rect x="49" y="310" width="5" height="5" rx="1"/>
  </g>
  <text x="70" y="313" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">4 道候选菜</text>
  <rect x="180" y="295" width="26" height="26" rx="7" fill="#EAF7EF"/>
  <circle cx="193" cy="308" r="7" fill="none" stroke="#159B55" stroke-width="1.3"/>
  <path d="M193 304v5l3 2" fill="none" stroke="#159B55" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="216" y="313" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">今天 21:00 截止</text>

  <g filter="url(#shadow)"><rect x="16" y="350" width="386" height="428" rx="8" fill="#FFFFFF" stroke="#E9EFEB"/></g>
  <text x="34" y="388" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">点餐码</text>
  <rect x="302" y="367" width="82" height="28" rx="14" fill="#EFF8F2"/><path d="M317 377h12v9h-12zM320 374h6v4" fill="none" stroke="#159B55" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><text x="355" y="386" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="750" fill="#159B55">可加入</text>
  <text x="34" y="420" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">让朋友在小程序中输入这组号码</text>
  ${codeBoxes()}
  <rect x="142" y="521" width="134" height="38" rx="19" fill="#FFFFFF" stroke="#B9DFC9"/>
  <path d="M163 533h12v14h-12zM168 529h12v14" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="224" y="546" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="800" fill="#159B55">复制点餐码</text>
  <line x1="34" y1="582" x2="384" y2="582" stroke="#EEF1EE"/>
  <rect x="34" y="602" width="350" height="70" rx="8" fill="#EFF8F2"/>
  <circle cx="62" cy="637" r="17" fill="#DFF3E7"/><circle cx="62" cy="637" r="8" fill="none" stroke="#159B55" stroke-width="1.5"/><path d="M62 632v6l4 2" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round"/>
  <text x="88" y="630" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#1B241F">剩余 3 小时 18 分</text>
  <text x="88" y="651" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#6F7D76">到期后自动关闭，点餐码不能重新开启</text>
  <rect x="34" y="692" width="350" height="54" rx="27" fill="url(#green)"/>
  <path d="M145 718l7-4M145 720l7 4" fill="none" stroke="#FFFFFF" stroke-width="1.7" stroke-linecap="round"/>
  <circle cx="141" cy="719" r="3" fill="none" stroke="#FFFFFF" stroke-width="1.7"/>
  <circle cx="155" cy="712" r="3" fill="none" stroke="#FFFFFF" stroke-width="1.7"/>
  <circle cx="155" cy="726" r="3" fill="none" stroke="#FFFFFF" stroke-width="1.7"/>
  <text x="231" y="726" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#FFFFFF">分享给微信好友</text>

  <text x="16" y="824" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">候选菜</text>
  <text x="386" y="822" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">共 4 道</text>
  <g filter="url(#shadow)"><rect x="16" y="840" width="386" height="136" rx="8" fill="#FFFFFF" stroke="#E9EFEB"/></g>
  ${dishes.map((dish, index) => { const x = 26 + index * 95; return `<text x="${x + 41}" y="958" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="9" font-weight="700" fill="#53615A">${dish.name}</text>`; }).join("")}

  <g filter="url(#topShadow)"><rect x="0" y="1048" width="418" height="192" fill="#FFFFFF"/></g>
  <rect x="24" y="1070" width="370" height="50" rx="25" fill="#FFFFFF" stroke="#E5AAA6"/>
  <circle cx="146" cy="1095" r="10" fill="none" stroke="#C9534E" stroke-width="1.5"/>
  <rect x="142" y="1091" width="8" height="8" rx="1.5" fill="#C9534E"/>
  <text x="233" y="1102" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#C9534E">提前关闭点餐</text>
  <text x="209" y="1150" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#8B9690">关闭后参与者不能继续修改点餐</text>
  <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>`;
}

function mainSvg() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${chrome()}${mainContent()}</svg>`;
}

function closeOverlaySvg() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}
    <rect width="418" height="1240" fill="#132119" fill-opacity="0.42"/>
    <path d="M0 810Q0 784 26 784H392Q418 784 418 810V1240H0Z" fill="#FFFFFF"/>
    <rect x="181" y="796" width="56" height="5" rx="2.5" fill="#DDE5E0"/>
    <circle cx="209" cy="857" r="29" fill="#FFF0EF"/>
    <circle cx="209" cy="857" r="13" fill="none" stroke="#C9534E" stroke-width="2.2"/>
    <rect x="203" y="851" width="12" height="12" rx="2" fill="#C9534E"/>
    <text x="209" y="914" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">提前关闭点餐？</text>
    <text x="209" y="944" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6F7D76">关闭后参与者不能新增或修改点餐，且无法重新开启。</text>
    <text x="209" y="966" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6F7D76">你仍可查看点菜结果并确认最终菜单。</text>
    <rect x="24" y="994" width="370" height="62" rx="8" fill="#F7F9F7"/>
    <circle cx="52" cy="1025" r="6" fill="#20B866"/><text x="70" y="1021" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="800" fill="#1B241F">当前仍在收集中</text><text x="70" y="1041" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">原定今天 21:00 自动关闭</text>
    <rect x="24" y="1080" width="174" height="52" rx="26" fill="#F1F4F2"/>
    <text x="111" y="1113" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="750" fill="#59665F">继续收集</text>
    <rect x="220" y="1080" width="174" height="52" rx="26" fill="#D45B53"/>
    <text x="307" y="1113" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#FFFFFF">确认关闭</text>
    <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>
  </svg>`;
}

async function roundedImage(source, width, height, radius) {
  const mask = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`);
  return sharp(source).resize(width, height, { fit: "cover", position: "centre" }).composite([{ input: mask, blend: "dest-in" }]).png().toBuffer();
}

async function prepareHero() {
  return sharp(heroSource).resize(WIDTH, 224, { fit: "cover", position: "centre" }).modulate({ brightness: 1.03, saturation: 0.86 }).png().toBuffer();
}

async function render() {
  const hero = await prepareHero();
  const dishPhotos = await Promise.all(dishes.map((dish) => roundedImage(dish.image, 82, 82, 7)));
  const base = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite([
    { input: hero, left: 0, top: 0 },
    { input: Buffer.from(mainSvg()), left: 0, top: 0 },
    ...dishPhotos.map((input, index) => ({ input, left: 26 + index * 95, top: 852 })),
  ]).png().toBuffer();
  const close = await sharp(base).composite([{ input: Buffer.from(closeOverlaySvg()), left: 0, top: 0 }]).png().toBuffer();
  fs.writeFileSync("miniApp/.preview/14-meal_invite-preview.png", base);
  fs.writeFileSync("miniApp/.preview/14-meal_invite-close-confirm-preview.png", close);
  console.log(JSON.stringify({ outputs: ["miniApp/.preview/14-meal_invite-preview.png", "miniApp/.preview/14-meal_invite-close-confirm-preview.png"], width: WIDTH, height: HEIGHT }, null, 2));
}

render().catch((error) => { console.error(error); process.exit(1); });
