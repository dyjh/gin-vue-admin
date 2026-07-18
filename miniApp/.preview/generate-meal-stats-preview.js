const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/meal-stats-cat-checklist-v1.png";
const dishes = [
  { name: "香菇鸡腿饭", meta: "基础份量 2 人", votes: "3 人想吃", suggest: 2, final: 2, image: "miniApp/.preview/assets/dish-detail-cover-realistic.png" },
  { name: "番茄炒蛋", meta: "基础份量 2 人", votes: "2 人想吃", suggest: 1, final: 1, image: "miniApp/.preview/assets/recommended-tomato-egg.png" },
  { name: "蒜蓉西兰花", meta: "基础份量 2 人", votes: "2 人想吃", suggest: 1, final: 1, image: "miniApp/.preview/assets/recommended-garlic-broccoli.png" },
  { name: "冬瓜丸子汤", meta: "基础份量 3 人", votes: "1 人想吃", suggest: 1, final: 1, image: "miniApp/.preview/assets/recommended-winter-melon-soup.png" },
];

function defs() {
  return [
    "<defs>",
    "<filter id='shadow' x='-8%' y='-20%' width='116%' height='160%'><feDropShadow dx='0' dy='6' stdDeviation='10' flood-color='#10261A' flood-opacity='0.05'/></filter>",
    "<filter id='topShadow' x='-10%' y='-50%' width='120%' height='180%'><feDropShadow dx='0' dy='-4' stdDeviation='8' flood-color='#163322' flood-opacity='0.055'/></filter>",
    "<linearGradient id='green' x1='0' x2='1'><stop offset='0' stop-color='#20B866'/><stop offset='1' stop-color='#0A8E52'/></linearGradient>",
    "</defs>",
  ].join("");
}

function chrome(title, subtitle) {
  return [
    "<text x='44' y='39' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='15' font-weight='700' fill='#151F1A'>9:41</text>",
    "<g fill='#151F1A'><rect x='326' y='32' width='4' height='8' rx='1'/><rect x='332' y='28' width='4' height='12' rx='1'/><rect x='338' y='24' width='4' height='16' rx='1'/>",
    "<path d='M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0' fill='none' stroke='#151F1A' stroke-width='2' stroke-linecap='round'/>",
    "<rect x='373' y='28' width='21' height='11' rx='3' fill='none' stroke='#151F1A' stroke-width='1.6'/><rect x='395' y='31' width='2' height='5' rx='1'/><rect x='376' y='31' width='15' height='5' rx='1'/></g>",
    "<g><rect x='312' y='60' width='82' height='36' rx='18' fill='#FFFFFF' fill-opacity='0.92' stroke='#E7E9E5'/><circle cx='338' cy='76' r='2.4' fill='#151F1A'/><circle cx='348' cy='76' r='2.4' fill='#151F1A'/><line x1='359' y1='66' x2='359' y2='86' stroke='#E7E9E5'/><circle cx='378' cy='76' r='9' fill='none' stroke='#151F1A' stroke-width='2.6'/></g>",
    "<path d='M48 79L38 89l10 10M39 89h20' fill='none' stroke='#151F1A' stroke-width='2.8' stroke-linecap='round' stroke-linejoin='round'/>",
    "<text x='70' y='96' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='24' font-weight='800' fill='#151F1A'>" + title + "</text>",
    "<text x='70' y='124' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='13' fill='#6F7D76'>" + subtitle + "</text>",
    "<rect y='224' width='418' height='1016' fill='#F8FAF7'/>",
  ].join("");
}

function mealSummary() {
  return [
    "<g filter='url(#shadow)'><rect x='16' y='242' width='386' height='88' rx='8' fill='#FFFFFF' stroke='#E9EFEB'/></g>",
    "<text x='34' y='274' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='17' font-weight='850' fill='#1B241F'>周六家庭聚餐</text>",
    "<rect x='317' y='254' width='67' height='26' rx='13' fill='#FFF5E9'/><circle cx='331' cy='267' r='3.5' fill='#E9A141'/>",
    "<text x='357' y='271' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='10' font-weight='750' fill='#C87F27'>待确认</text>",
    "<circle cx='42' cy='304' r='7' fill='none' stroke='#7D8983' stroke-width='1.3'/><path d='M42 300v5l3 2' fill='none' stroke='#7D8983' stroke-width='1.3' stroke-linecap='round'/>",
    "<text x='56' y='308' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='10' fill='#6F7D76'>点餐已关闭</text>",
    "<circle cx='151' cy='303' r='7' fill='#EAF7EF'/><path d='M147 303h8M151 299v8' stroke='#159B55' stroke-width='1.2' stroke-linecap='round'/>",
    "<text x='165' y='308' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='10' fill='#6F7D76'>4 人参与 · 4 道候选</text>",
  ].join("");
}

function servingStepper(y, value) {
  return [
    "<text x='264' y='" + (y + 72) + "' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#8A9690'>最终份数</text>",
    "<rect x='264' y='" + (y + 80) + "' width='122' height='32' rx='7' fill='#F7FAF8' stroke='#DDE7E1'/>",
    "<path d='M276 " + (y + 96) + "h10' stroke='#718078' stroke-width='1.7' stroke-linecap='round'/>",
    "<text x='325' y='" + (y + 101) + "' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='14' font-weight='850' fill='#1B241F'>" + value + "</text>",
    "<path d='M365 " + (y + 96) + "h10M370 " + (y + 91) + "v10' stroke='#159B55' stroke-width='1.7' stroke-linecap='round'/>",
    "<line x1='298' y1='" + (y + 86) + "' x2='298' y2='" + (y + 106) + "' stroke='#E3E9E5'/>",
    "<line x1='351' y1='" + (y + 86) + "' x2='351' y2='" + (y + 106) + "' stroke='#E3E9E5'/>",
  ].join("");
}

function dishCards() {
  return dishes.map((dish, index) => {
    const y = 466 + index * 128;
    return [
      "<g filter='url(#shadow)'><rect x='16' y='" + y + "' width='386' height='118' rx='8' fill='#FFFFFF' stroke='#E9EFEB'/></g>",
      "<text x='114' y='" + (y + 32) + "' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='15' font-weight='850' fill='#1B241F'>" + dish.name + "</text>",
      "<text x='114' y='" + (y + 55) + "' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='10' font-weight='750' fill='#159B55'>" + dish.votes + "</text>",
      "<text x='114' y='" + (y + 76) + "' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#7B8982'>" + dish.meta + "</text>",
      "<rect x='114' y='" + (y + 87) + "' width='74' height='22' rx='11' fill='#EAF7EF'/>",
      "<text x='151' y='" + (y + 102) + "' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' font-weight='750' fill='#159B55'>建议 " + dish.suggest + " 份</text>",
      "<text x='322' y='" + (y + 31) + "' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#6F7D76'>做</text>",
      "<rect x='345' y='" + (y + 17) + "' width='41' height='23' rx='11.5' fill='#20B866'/><circle cx='374' cy='" + (y + 28.5) + "' r='9' fill='#FFFFFF'/>",
      servingStepper(y, dish.final),
    ].join("");
  }).join("");
}

function confirmOverlay() {
  return [
    "<rect x='0' y='0' width='418' height='1240' fill='#172019' fill-opacity='0.38'/>",
    "<path d='M0 738Q0 722 16 722H402Q418 722 418 738V1240H0Z' fill='#FFFFFF'/>",
    "<rect x='177' y='738' width='64' height='5' rx='2.5' fill='#DDE4DF'/>",
    "<circle cx='49' cy='790' r='20' fill='#EAF7EF'/><path d='M41 790l6 6 11-14' fill='none' stroke='#159B55' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'/>",
    "<text x='80' y='786' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='20' font-weight='850' fill='#1B241F'>确认最终菜单？</text>",
    "<text x='80' y='810' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='10' fill='#7B8982'>确认后将生成饭局快照和采购清单</text>",
    "<line x1='24' y1='838' x2='394' y2='838' stroke='#EEF1EE'/>",
    "<text x='82' y='873' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='19' font-weight='850' fill='#1B241F'>4</text><text x='82' y='893' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#7B8982'>道菜</text>",
    "<line x1='139' y1='858' x2='139' y2='898' stroke='#EEF1EE'/>",
    "<text x='209' y='873' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='19' font-weight='850' fill='#1B241F'>5</text><text x='209' y='893' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#7B8982'>制作份数</text>",
    "<line x1='279' y1='858' x2='279' y2='898' stroke='#EEF1EE'/>",
    "<text x='345' y='873' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='19' font-weight='850' fill='#1B241F'>4</text><text x='345' y='893' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#7B8982'>参与人数</text>",
    "<line x1='24' y1='920' x2='394' y2='920' stroke='#EEF1EE'/>",
    "<circle cx='39' cy='954' r='8' fill='#EAF7EF'/><path d='M35 954l3 3 5-6' fill='none' stroke='#159B55' stroke-width='1.4' stroke-linecap='round' stroke-linejoin='round'/><text x='56' y='958' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='11' fill='#48554E'>采购清单将按最终份数汇总配料</text>",
    "<circle cx='39' cy='986' r='8' fill='#EAF7EF'/><path d='M35 986l3 3 5-6' fill='none' stroke='#159B55' stroke-width='1.4' stroke-linecap='round' stroke-linejoin='round'/><text x='56' y='990' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='11' fill='#48554E'>后续调整份数时，采购数量同步重算</text>",
    "<rect x='24' y='1034' width='136' height='52' rx='26' fill='#FFFFFF' stroke='#CBD4CE'/><text x='92' y='1067' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='14' font-weight='800' fill='#59665F'>再检查一下</text>",
    "<rect x='170' y='1034' width='224' height='52' rx='26' fill='url(#green)'/><text x='282' y='1067' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='14' font-weight='850' fill='#FFFFFF'>确认并生成</text>",
    "<text x='209' y='1120' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#8A9690'>生成后饭局状态更新为菜单已确认</text>",
    "<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>",
  ].join("");
}

function mainSvg(showConfirm) {
  return [
    "<svg width='" + WIDTH + "' height='" + HEIGHT + "' viewBox='0 0 " + WIDTH + " " + HEIGHT + "' xmlns='http://www.w3.org/2000/svg'>",
    defs(),
    chrome("确认菜单", "查看统计，调整最终制作份数"),
    mealSummary(),
    "<rect x='16' y='342' width='386' height='58' rx='8' fill='#EFF8F2'/>",
    "<circle cx='45' cy='371' r='15' fill='#DFF3E7'/><path d='M40 371l4 4 7-9' fill='none' stroke='#159B55' stroke-width='1.7' stroke-linecap='round' stroke-linejoin='round'/>",
    "<text x='70' y='367' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='12' font-weight='800' fill='#1B241F'>系统已按基础份量给出建议</text>",
    "<text x='70' y='386' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#6F7D76'>关闭“做”可移出菜单，最终份数可手动调整</text>",
    "<text x='24' y='434' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='19' font-weight='850' fill='#1B241F'>点菜统计</text>",
    "<text x='394' y='433' text-anchor='end' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='10' fill='#7B8982'>按想吃人数排序</text>",
    "<text x='24' y='454' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='10' fill='#8A9690'>确认每道菜是否制作，并调整最终份数</text>",
    dishCards(),
    "<g filter='url(#topShadow)'><rect x='0' y='986' width='418' height='254' fill='#FFFFFF'/></g>",
    "<text x='24' y='1027' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='13' font-weight='850' fill='#1B241F'>最终菜单 4 道 · 共 5 份</text>",
    "<rect x='24' y='1050' width='370' height='54' rx='27' fill='url(#green)'/><path d='M92 1073l5 5 9-11' fill='none' stroke='#FFFFFF' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'/>",
    "<text x='226' y='1084' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='15' font-weight='850' fill='#FFFFFF'>确认菜单并生成采购清单</text>",
    "<text x='209' y='1132' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='9' fill='#8B9690'>确认后生成饭局快照，采购清单可继续编辑</text>",
    "<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>",
    showConfirm ? confirmOverlay() : "",
    "</svg>",
  ].join("");
}

function modalSvg() {
  return [
    "<svg width='" + WIDTH + "' height='" + HEIGHT + "' viewBox='0 0 " + WIDTH + " " + HEIGHT + "' xmlns='http://www.w3.org/2000/svg'>",
    defs(),
    confirmOverlay(),
    "</svg>",
  ].join("");
}

async function roundedImage(source, width, height, radius) {
  const mask = Buffer.from("<svg width='" + width + "' height='" + height + "' xmlns='http://www.w3.org/2000/svg'><rect width='" + width + "' height='" + height + "' rx='" + radius + "' fill='#fff'/></svg>");
  return sharp(source).resize(width, height, { fit: "cover", position: "centre" }).composite([{ input: mask, blend: "dest-in" }]).png().toBuffer();
}

async function prepareHero() {
  return sharp(heroSource).resize(WIDTH, 224, { fit: "cover", position: "centre" }).modulate({ brightness: 1.02, saturation: 0.92 }).png().toBuffer();
}

async function renderScreen(hero, svg, dishPhotos, overlay) {
  return sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite([
    { input: hero, left: 0, top: 0 },
    { input: Buffer.from(svg), left: 0, top: 0 },
    ...dishPhotos.map((input, index) => ({ input, left: 28, top: 489 + index * 128 })),
    ...(overlay ? [{ input: Buffer.from(overlay), left: 0, top: 0 }] : []),
  ]).png().toBuffer();
}

async function render() {
  const hero = await prepareHero();
  const dishPhotos = await Promise.all(dishes.map((dish) => roundedImage(dish.image, 72, 72, 7)));
  const outputs = [
    ["miniApp/.preview/16-meal_stats-preview.png", await renderScreen(hero, mainSvg(false), dishPhotos)],
    ["miniApp/.preview/16-meal_stats-confirm-preview.png", await renderScreen(hero, mainSvg(false), dishPhotos, modalSvg())],
  ];
  outputs.forEach(([file, buffer]) => fs.writeFileSync(file, buffer));
  console.log(JSON.stringify({ outputs: outputs.map(([file]) => file), width: WIDTH, height: HEIGHT }, null, 2));
}

render().catch((error) => { console.error(error); process.exit(1); });
