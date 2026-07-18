const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/checkin-cat-camera-v1.png";
const tomatoSource = "miniApp/.preview/assets/recommended-tomato-egg.png";
const soupSource = "miniApp/.preview/assets/recommended-winter-melon-soup.png";

function defs() {
  return `<defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="160%"><feDropShadow dx="0" dy="7" stdDeviation="12" flood-color="#10261A" flood-opacity="0.055"/></filter>
    <filter id="topShadow" x="-10%" y="-50%" width="120%" height="170%"><feDropShadow dx="0" dy="-4" stdDeviation="8" flood-color="#163322" flood-opacity="0.06"/></filter>
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
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="800" fill="#151F1A">做菜打卡</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">记下今天亲手做的一餐</text>
  <rect y="224" width="418" height="1016" fill="#F8FAF7"/>`;
}

function calendarDates() {
  const xs = [48, 101, 154, 207, 260, 313, 366];
  const rows = [
    [null, null, null, 1, 2, 3, 4],
    [5, 6, 7, 8, 9, 10, 11],
    [12, 13, 14, 15, 16, 17, 18],
    [19, 20, 21, 22, 23, 24, 25],
    [26, 27, 28, 29, 30, 31, null],
  ];
  const ys = [362, 402, 442, 482, 522];
  const checked = new Set([2, 5, 8, 11, 13, 16]);
  return rows.map((row, rowIndex) => row.map((day, colIndex) => {
    if (!day) return "";
    const x = xs[colIndex];
    const y = ys[rowIndex];
    if (day === 16) {
      return `<circle cx="${x}" cy="${y - 5}" r="17" fill="#159B55"/><text x="${x}" y="${y}" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#FFFFFF">${day}</text><circle cx="${x}" cy="${y + 16}" r="2" fill="#159B55"/>`;
    }
    return `<text x="${x}" y="${y}" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="${checked.has(day) ? 750 : 500}" fill="#26312B">${day}</text>${checked.has(day) ? `<circle cx="${x}" cy="${y + 12}" r="2.2" fill="#20B866"/>` : ""}`;
  }).join("")).join("");
}

function calendarCard() {
  const weekdays = ["日", "一", "二", "三", "四", "五", "六"];
  const xs = [48, 101, 154, 207, 260, 313, 366];
  return `
  <g filter="url(#shadow)"><rect x="16" y="240" width="386" height="330" rx="10" fill="#FFFFFF" stroke="#EAF0EC"/></g>
  <path d="M43 270l-6 6 6 6" fill="none" stroke="#708078" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="64" y="282" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="850" fill="#1B241F">2026年7月</text>
  <rect x="285" y="258" width="78" height="28" rx="14" fill="#EAF7EF"/><text x="324" y="277" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">本月 7 次</text>
  <path d="M378 270l6 6-6 6" fill="none" stroke="#708078" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
  <line x1="32" y1="300" x2="386" y2="300" stroke="#EEF1EE"/>
  ${weekdays.map((day, index) => `<text x="${xs[index]}" y="329" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="700" fill="${index === 0 || index === 6 ? "#A18C7C" : "#7B8982"}">${day}</text>`).join("")}
  ${calendarDates()}
  <line x1="32" y1="543" x2="386" y2="543" stroke="#EEF1EE"/>
  <circle cx="43" cy="557" r="3" fill="#20B866"/><text x="55" y="561" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">有打卡记录</text><text x="375" y="561" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">累计打卡 18 天</text>`;
}

function records() {
  return `
  <text x="24" y="612" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="850" fill="#1B241F">7月16日</text>
  <rect x="111" y="590" width="58" height="26" rx="13" fill="#EFF7F2"/><text x="140" y="608" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">2 次打卡</text>

  <g filter="url(#shadow)"><rect x="16" y="632" width="386" height="106" rx="8" fill="#FFFFFF" stroke="#EAF0EC"/></g>
  <text x="124" y="663" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="850" fill="#1B241F">番茄炒蛋</text>
  <text x="124" y="687" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">18:42 · 少油，今天很下饭</text>
  <rect x="124" y="701" width="84" height="24" rx="12" fill="#EAF7EF"/><text x="166" y="718" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="750" fill="#159B55">首条 +2 积分</text>
  <path d="M380 679l5 5-5 5" fill="none" stroke="#A1ACA6" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>

  <g filter="url(#shadow)"><rect x="16" y="752" width="386" height="106" rx="8" fill="#FFFFFF" stroke="#EAF0EC"/></g>
  <text x="124" y="783" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="850" fill="#1B241F">冬瓜丸子汤</text>
  <text x="124" y="807" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">12:16 · 清淡暖胃</text>
  <rect x="124" y="821" width="58" height="24" rx="12" fill="#F1F4F2"/><text x="153" y="838" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="700" fill="#6F7D76">已记录</text>
  <path d="M380 799l5 5-5 5" fill="none" stroke="#A1ACA6" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>

  <rect x="16" y="882" width="386" height="68" rx="8" fill="#EFF8F2"/>
  <circle cx="44" cy="916" r="16" fill="#DFF3E7"/><path d="M38 916l4 4 8-9M44 906v-3M44 929v-3M34 916h-3M57 916h-3" fill="none" stroke="#159B55" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="70" y="910" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#1B241F">今天的首次奖励已领取</text>
  <text x="70" y="932" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">仍可继续打卡，后续记录不重复奖励积分</text>`;
}

function bottomAction() {
  return `
  <g filter="url(#topShadow)"><rect x="0" y="1120" width="418" height="120" fill="#FFFFFF"/></g>
  <rect x="24" y="1140" width="370" height="50" rx="25" fill="url(#green)"/>
  <path d="M165 1159h20v15h-20zM170 1155h10l3 4M171 1167a4 4 0 108 0 4 4 0 10-8 0" fill="none" stroke="#FFFFFF" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="226" y="1172" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#FFFFFF">新增打卡</text>
  <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>`;
}

function addOverlay(filled = false) {
  const photoArea = filled
    ? `<rect x="24" y="492" width="370" height="188" rx="8" fill="#F5FAF7" stroke="#CFE4D7"/>`
    : `
  <rect x="24" y="492" width="370" height="188" rx="8" fill="#F5FAF7" stroke="#CFE4D7" stroke-dasharray="6 5"/>
  <circle cx="209" cy="560" r="25" fill="#E2F4E8"/><path d="M196 554h26v19h-26zM202 549h12l3 5M202 562a7 7 0 1014 0 7 7 0 10-14 0" fill="none" stroke="#159B55" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="209" y="612" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">上传做菜照片</text>
  <text x="209" y="635" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">上传后将进行图片内容检查</text>`;
  const dishName = filled
    ? `<text x="42" y="772" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#1B241F">番茄炒蛋</text><circle cx="371" cy="765" r="8" fill="#E5F6EB"/><path d="M367 765l3 3 5-6" fill="none" stroke="#159B55" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>`
    : `<text x="42" y="772" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#9AA49E">输入这道菜的名字</text>`;
  const note = filled
    ? `<text x="42" y="880" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#1B241F">少油，今天很下饭</text><text x="376" y="920" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#A5AEA9">8 / 100</text>`
    : `<text x="42" y="880" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#9AA49E">今天这道菜怎么样？</text><text x="376" y="920" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#A5AEA9">0 / 100</text>`;
  const submit = filled
    ? `<rect x="24" y="982" width="370" height="50" rx="25" fill="url(#green)"/><text x="209" y="1014" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#FFFFFF">完成打卡</text>`
    : `<rect x="24" y="982" width="370" height="50" rx="25" fill="#E4ECE7"/><text x="209" y="1014" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#91A099">请先完善必填项</text>`;

  return `
  <rect x="0" y="0" width="418" height="1240" fill="#102018" fill-opacity="0.3"/>
  <path d="M0 374Q0 348 26 348H392Q418 348 418 374V1240H0Z" fill="#FFFFFF"/>
  <rect x="181" y="360" width="56" height="5" rx="2.5" fill="#DDE5E0"/>
  <text x="24" y="408" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">新增打卡</text>
  <text x="24" y="435" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">7月16日 · 周四</text>
  <path d="M374 389l14 14M388 389l-14 14" fill="none" stroke="#78857E" stroke-width="1.8" stroke-linecap="round"/>

  <text x="24" y="476" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#1B241F">做菜照片</text><rect x="102" y="457" width="40" height="22" rx="11" fill="#FFF3E2"/><text x="122" y="472" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="750" fill="#D08025">必填</text>
  ${photoArea}

  <text x="24" y="724" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#1B241F">菜名</text><rect x="69" y="705" width="40" height="22" rx="11" fill="#FFF3E2"/><text x="89" y="720" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="750" fill="#D08025">必填</text>
  <rect x="24" y="740" width="370" height="50" rx="8" fill="#F8FAF8" stroke="#DDE6E0"/>${dishName}

  <text x="24" y="834" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#1B241F">备注</text><text x="70" y="834" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#9AA49E">选填</text>
  <rect x="24" y="850" width="370" height="88" rx="8" fill="#F8FAF8" stroke="#DDE6E0"/>${note}

  ${submit}
  <text x="209" y="1064" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">每天可多次打卡，仅当天第一条有效记录奖励积分</text>
  <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>`;
}

function filledPhotoChrome() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  <rect x="24" y="492" width="370" height="188" rx="8" fill="none" stroke="#CFE4D7"/>
  <rect x="36" y="504" width="72" height="24" rx="12" fill="#EAF7EF"/><path d="M47 516l4 4 7-8" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><text x="80" y="521" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="750" fill="#159B55">检查通过</text>
  <rect x="302" y="504" width="50" height="24" rx="12" fill="#FFFFFF" fill-opacity="0.92"/><text x="327" y="521" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" font-weight="750" fill="#526159">更换</text>
  <circle cx="372" cy="516" r="12" fill="#FFFFFF" fill-opacity="0.92"/><path d="M368 512h8M370 510h4M369 513l1 8h4l1-8" fill="none" stroke="#D45B50" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>`;
}

function successOverlay() {
  return `
  <rect x="0" y="0" width="418" height="1240" fill="#102018" fill-opacity="0.3"/>
  <path d="M0 562Q0 536 26 536H392Q418 536 418 562V1240H0Z" fill="#FFFFFF"/>
  <rect x="181" y="548" width="56" height="5" rx="2.5" fill="#DDE5E0"/>
  <circle cx="209" cy="618" r="29" fill="#E4F6EA"/><path d="M196 618l9 9 18-21" fill="none" stroke="#159B55" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="209" y="674" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#1B241F">打卡成功</text>
  <text x="209" y="699" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">7月16日 18:42</text>

  <rect x="24" y="724" width="370" height="72" rx="8" fill="#EFF8F2"/>
  <circle cx="58" cy="760" r="18" fill="#DFF3E7"/><text x="58" y="766" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#159B55">+2</text>
  <text x="88" y="754" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">今日首次打卡奖励</text>
  <text x="88" y="778" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">积分已到账，可在积分流水中查看</text>

  <rect x="24" y="816" width="370" height="94" rx="8" fill="#FFFFFF" stroke="#E6ECE8"/>
  <text x="118" y="850" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="850" fill="#1B241F">番茄炒蛋</text>
  <text x="118" y="875" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">少油，今天很下饭</text>
  <rect x="118" y="883" width="46" height="18" rx="9" fill="#EAF7EF"/><text x="141" y="896" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="9" font-weight="700" fill="#159B55">已记录</text>

  <text x="209" y="948" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">去添加菜品只会预填菜名，不会带入打卡照片</text>
  <rect x="24" y="974" width="176" height="48" rx="24" fill="#FFFFFF" stroke="#159B55"/><text x="112" y="1005" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#159B55">去添加菜品</text>
  <rect x="210" y="974" width="184" height="48" rx="24" fill="url(#green)"/><text x="302" y="1005" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="850" fill="#FFFFFF">完成</text>
  <rect x="147" y="1215" width="124" height="5" rx="2.5" fill="#303833"/>`;
}

function pageSvg() {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${chrome()}${calendarCard()}${records()}${bottomAction()}</svg>`;
}

function overlaySvg(state) {
  const content = state === "add"
    ? addOverlay(false)
    : state === "filled"
      ? addOverlay(true)
      : successOverlay();
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${content}</svg>`;
}

async function roundedImage(source, width, height, radius) {
  const mask = Buffer.from(`<svg width="${width}" height="${height}"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`);
  return sharp(source).resize(width, height, { fit: "cover" }).composite([{ input: mask, blend: "dest-in" }]).png().toBuffer();
}

async function render(state) {
  const hero = await sharp(heroSource).resize(418, 224, { fit: "fill" }).modulate({ brightness: 1.025, saturation: 0.9 }).png().toBuffer();
  const tomato = await roundedImage(tomatoSource, 84, 84, 8);
  const soup = await roundedImage(soupSource, 84, 84, 8);
  const successTomato = await roundedImage(tomatoSource, 72, 72, 8);
  const filledPhoto = await roundedImage(tomatoSource, 366, 184, 6);
  const composites = [
    { input: hero, left: 0, top: 0 },
    { input: Buffer.from(pageSvg()), left: 0, top: 0 },
    { input: tomato, left: 28, top: 643 },
    { input: soup, left: 28, top: 763 },
  ];
  if (state !== "main") composites.push({ input: Buffer.from(overlaySvg(state)), left: 0, top: 0 });
  if (state === "filled") {
    composites.push({ input: filledPhoto, left: 26, top: 494 });
    composites.push({ input: Buffer.from(filledPhotoChrome()), left: 0, top: 0 });
  }
  if (state === "success") composites.push({ input: successTomato, left: 36, top: 827 });
  return sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite(composites).png().toBuffer();
}

async function build() {
  const main = await render("main");
  const add = await render("add");
  const filled = await render("filled");
  const success = await render("success");
  fs.writeFileSync("miniApp/.preview/12-checkin_calendar-preview.png", main);
  fs.writeFileSync("miniApp/.preview/12-checkin_calendar-add-preview.png", add);
  fs.writeFileSync("miniApp/.preview/12-checkin_calendar-filled-preview.png", filled);
  fs.writeFileSync("miniApp/.preview/12-checkin_calendar-success-preview.png", success);
  console.log(JSON.stringify({ outputs: ["miniApp/.preview/12-checkin_calendar-preview.png", "miniApp/.preview/12-checkin_calendar-add-preview.png", "miniApp/.preview/12-checkin_calendar-filled-preview.png", "miniApp/.preview/12-checkin_calendar-success-preview.png"], width: WIDTH, height: HEIGHT }, null, 2));
}

build().catch((error) => { console.error(error); process.exit(1); });
