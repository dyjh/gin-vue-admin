const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/shopping-list-cat-basket-v1.png";

const pendingItems = [
  { name: "土豆", source: "来自 土豆丝", amount: "4 个" },
  { name: "青椒", source: "来自 土豆丝、青椒牛柳", amount: "3 个" },
  { name: "西兰花", source: "来自 蒜蓉西兰花", amount: "2 颗" },
  { name: "鸡腿肉", source: "来自 香菇鸡腿饭", amount: "600 克" },
  { name: "鸡蛋", source: "来自 番茄炒蛋", amount: "6 个" },
  { name: "大米", source: "来自 香菇鸡腿饭", amount: "400 克" },
  { name: "食用油", source: "3 道菜汇总", amount: "45 毫升" },
  { name: "姜", source: "来自 冬瓜丸子汤", amount: "20 克" },
];

const completedItems = [
  { name: "番茄", source: "来自 番茄炒蛋", amount: "4 个" },
  { name: "香菇", source: "来自 香菇鸡腿饭", amount: "8 朵" },
  { name: "蒜", source: "2 道菜汇总", amount: "1 头" },
  { name: "冬瓜", source: "来自 冬瓜丸子汤", amount: "500 克" },
];

function defs() {
  return [
    "<defs>",
    "<filter id='shadow' x='-8%' y='-20%' width='116%' height='160%'><feDropShadow dx='0' dy='5' stdDeviation='9' flood-color='#10261A' flood-opacity='0.045'/></filter>",
    "<filter id='topShadow' x='-10%' y='-50%' width='120%' height='180%'><feDropShadow dx='0' dy='-4' stdDeviation='8' flood-color='#163322' flood-opacity='0.055'/></filter>",
    "<linearGradient id='green' x1='0' x2='1'><stop offset='0' stop-color='#20B866'/><stop offset='1' stop-color='#0A8E52'/></linearGradient>",
    "</defs>",
  ].join("");
}

function text(x, y, value, size, weight, fill, extra = "") {
  return "<text x='" + x + "' y='" + y + "' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='" + size + "' font-weight='" + weight + "' fill='" + fill + "' " + extra + ">" + value + "</text>";
}

function chrome() {
  return [
    text(44, 39, "9:41", 15, 700, "#151F1A"),
    "<g fill='#151F1A'><rect x='326' y='32' width='4' height='8' rx='1'/><rect x='332' y='28' width='4' height='12' rx='1'/><rect x='338' y='24' width='4' height='16' rx='1'/>",
    "<path d='M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0' fill='none' stroke='#151F1A' stroke-width='2' stroke-linecap='round'/>",
    "<rect x='373' y='28' width='21' height='11' rx='3' fill='none' stroke='#151F1A' stroke-width='1.6'/><rect x='395' y='31' width='2' height='5' rx='1'/><rect x='376' y='31' width='15' height='5' rx='1'/></g>",
    "<g><rect x='312' y='60' width='82' height='36' rx='18' fill='#FFFFFF' fill-opacity='0.92' stroke='#E7E9E5'/><circle cx='338' cy='76' r='2.4' fill='#151F1A'/><circle cx='348' cy='76' r='2.4' fill='#151F1A'/><line x1='359' y1='66' x2='359' y2='86' stroke='#E7E9E5'/><circle cx='378' cy='76' r='9' fill='none' stroke='#151F1A' stroke-width='2.6'/></g>",
    "<path d='M48 79L38 89l10 10M39 89h20' fill='none' stroke='#151F1A' stroke-width='2.8' stroke-linecap='round' stroke-linejoin='round'/>",
    text(70, 96, "采购清单", 24, 800, "#151F1A"),
    text(70, 124, "按清单采购，买完随手勾选", 13, 400, "#6F7D76"),
    "<rect y='224' width='418' height='1016' fill='#F8FAF7'/>",
  ].join("");
}

function copyIcon(x, y, color) {
  return [
    "<rect x='" + (x + 4) + "' y='" + y + "' width='11' height='13' rx='2' fill='none' stroke='" + color + "' stroke-width='1.5'/>",
    "<rect x='" + x + "' y='" + (y + 4) + "' width='11' height='13' rx='2' fill='none' stroke='" + color + "' stroke-width='1.5'/>",
  ].join("");
}

function shareIcon(x, y, color) {
  return [
    "<circle cx='" + x + "' cy='" + (y + 8) + "' r='2.5' fill='none' stroke='" + color + "' stroke-width='1.5'/>",
    "<circle cx='" + (x + 12) + "' cy='" + (y + 2) + "' r='2.5' fill='none' stroke='" + color + "' stroke-width='1.5'/>",
    "<circle cx='" + (x + 12) + "' cy='" + (y + 14) + "' r='2.5' fill='none' stroke='" + color + "' stroke-width='1.5'/>",
    "<path d='M" + (x + 2) + " " + (y + 7) + "l8-4M" + (x + 2) + " " + (y + 9) + "l8 4' fill='none' stroke='" + color + "' stroke-width='1.5' stroke-linecap='round'/>",
  ].join("");
}

function summary(purchased = 4) {
  const progress = purchased / 12;
  return [
    "<g filter='url(#shadow)'><rect x='16' y='242' width='386' height='170' rx='8' fill='#FFFFFF' stroke='#E9EFEB'/></g>",
    text(34, 274, "周六家庭聚餐", 17, 850, "#1B241F"),
    "<rect x='321' y='254' width='63' height='26' rx='13' fill='#EAF7EF'/><circle cx='334' cy='267' r='3.5' fill='#16B96F'/>",
    text(357, 271, "采购中", 10, 750, "#159B55", "text-anchor='middle'"),
    text(34, 299, "4 道菜 · 12 项食材", 10, 400, "#7B8982"),
    text(34, 326, "采购进度", 10, 750, "#48554E"),
    text(384, 326, "已买 " + purchased + " / 12", 10, 750, "#159B55", "text-anchor='end'"),
    "<rect x='34' y='338' width='350' height='7' rx='3.5' fill='#E9EFEB'/>",
    "<rect x='34' y='338' width='" + Math.round(350 * progress) + "' height='7' rx='3.5' fill='#20B866'/>",
    "<line x1='34' y1='362' x2='384' y2='362' stroke='#EEF1EE'/>",
    copyIcon(79, 378, "#159B55"),
    text(108, 392, "复制清单", 11, 750, "#159B55"),
    "<line x1='209' y1='374' x2='209' y2='400' stroke='#E5ECE7'/>",
    shareIcon(272, 378, "#159B55"),
    text(302, 392, "分享给家人", 11, 750, "#159B55"),
  ].join("");
}

function actionBar() {
  return [
    text(24, 449, "采购项", 19, 850, "#1B241F"),
    text(394, 448, "共 12 项", 10, 500, "#7B8982", "text-anchor='end'"),
  ].join("");
}

function tabs(active) {
  const pending = active === "pending";
  return [
    "<line x1='16' y1='506' x2='402' y2='506' stroke='#E4EBE6'/>",
    text(113, 491, "待购买 8", 12, pending ? 800 : 500, pending ? "#159B55" : "#7B8982", "text-anchor='middle'"),
    text(305, 491, "已购买 4", 12, pending ? 500 : 800, pending ? "#7B8982" : "#159B55", "text-anchor='middle'"),
    "<rect x='" + (pending ? 70 : 262) + "' y='503' width='86' height='3' rx='1.5' fill='#20B866'/>",
  ].join("");
}

function emptyCircle(cx, cy) {
  return "<circle cx='" + cx + "' cy='" + cy + "' r='10' fill='#FFFFFF' stroke='#16B96F' stroke-width='1.7'/>";
}

function checkedCircle(cx, cy) {
  return "<circle cx='" + cx + "' cy='" + cy + "' r='10' fill='#16B96F'/><path d='M" + (cx - 4) + " " + cy + "l3 3 6-7' fill='none' stroke='#FFFFFF' stroke-width='1.7' stroke-linecap='round' stroke-linejoin='round'/>";
}


function itemRow(y, item, checked = false, emphasize = false) {
  const primary = checked ? "#8A9690" : "#1B241F";
  const secondary = checked ? "#A2ACA7" : "#7B8982";
  return [
    emphasize ? "<rect x='16' y='" + y + "' width='386' height='64' fill='#F5FBF7'/>" : "",
    checked ? checkedCircle(36, y + 31) : emptyCircle(36, y + 31),
    text(58, y + 26, item.name, 14, checked ? 650 : 800, primary, checked ? "text-decoration='line-through'" : ""),
    text(58, y + 47, item.source, 9, 400, secondary),
    text(390, y + 37, item.amount, 12, 800, checked ? "#8A9690" : "#425149", "text-anchor='end'"),
    "<line x1='58' y1='" + (y + 63) + "' x2='402' y2='" + (y + 63) + "' stroke='#EEF1EE'/>",
  ].join("");
}

function pendingList(extraChecked = false) {
  let y = 518;
  const parts = [];
  pendingItems.forEach((item, index) => {
    const isFirst = index === 0;
    parts.push(itemRow(y, item, extraChecked && isFirst, extraChecked && isFirst));
    y += 64;
  });
  return parts.join("");
}

function completedList() {
  let y = 518;
  const parts = [];
  completedItems.forEach((item) => {
    parts.push(itemRow(y, item, true));
    y += 64;
  });
  parts.push("<rect x='16' y='" + (y + 18) + "' width='386' height='64' rx='8' fill='#F1F7F3'/>");
  parts.push("<circle cx='46' cy='" + (y + 50) + "' r='14' fill='#E1F4E9'/><path d='M40 " + (y + 50) + "l4 4 8-10' fill='none' stroke='#159B55' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'/>");
  parts.push(text(70, y + 45, "这些已经买好啦", 12, 800, "#2E4036"));
  parts.push(text(70, y + 63, "误勾时可再次点击恢复为待购买", 9, 400, "#7B8982"));
  return parts.join("");
}

function bottomBar(count = 8) {
  return [
    "<g filter='url(#topShadow)'><rect x='0' y='1080' width='418' height='160' fill='#FFFFFF'/></g>",
    text(24, 1113, "待购买 " + count + " 项", 12, 800, "#1B241F"),
    text(394, 1113, "清单修改只影响本次饭局", 9, 400, "#8A9690", "text-anchor='end'"),
    "<rect x='24' y='1130' width='370' height='52' rx='26' fill='url(#green)'/>",
    "<path d='M135 1156h14M142 1149v14' fill='none' stroke='#FFFFFF' stroke-width='1.9' stroke-linecap='round'/>",
    text(226, 1163, "新增采购项", 15, 850, "#FFFFFF", "text-anchor='middle'"),
    "<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>",
  ].join("");
}

function mainSvg(mode = "pending", extraChecked = false) {
  const completed = mode === "completed";
  const purchased = extraChecked ? 5 : 4;
  return [
    "<svg width='" + WIDTH + "' height='" + HEIGHT + "' viewBox='0 0 " + WIDTH + " " + HEIGHT + "' xmlns='http://www.w3.org/2000/svg'>",
    defs(),
    chrome(),
    summary(purchased),
    actionBar(),
    tabs(completed ? "completed" : "pending"),
    completed ? completedList() : pendingList(extraChecked),
    bottomBar(extraChecked ? 7 : 8),
    "</svg>",
  ].join("");
}

function inputField(y, label, value, width = 370, x = 24, muted = false) {
  return [
    text(x, y, label, 10, 750, "#57645D"),
    "<rect x='" + x + "' y='" + (y + 10) + "' width='" + width + "' height='48' rx='8' fill='#F8FAF8' stroke='#DDE7E1'/>",
    text(x + 16, y + 40, value, 13, muted ? 400 : 650, muted ? "#A2ACA7" : "#1B241F"),
  ].join("");
}

function sheetOverlay(kind) {
  const edit = kind === "edit";
  return [
    "<rect x='0' y='0' width='418' height='1240' fill='#172019' fill-opacity='0.38'/>",
    "<path d='M0 676Q0 660 16 660H402Q418 660 418 676V1240H0Z' fill='#FFFFFF'/>",
    "<rect x='177' y='674' width='64' height='5' rx='2.5' fill='#DDE4DF'/>",
    text(24, 724, edit ? "编辑采购项" : "新增采购项", 20, 850, "#1B241F"),
    "<circle cx='378' cy='714' r='16' fill='#F4F7F5'/><path d='M373 709l10 10M383 709l-10 10' stroke='#65736B' stroke-width='1.6' stroke-linecap='round'/>",
    inputField(758, "食材名称", edit ? "土豆" : "例如：小葱", 370, 24, !edit),
    inputField(836, "数量", edit ? "4" : "请输入", 178, 24, !edit),
    inputField(836, "单位", edit ? "个" : "请选择", 184, 210, !edit),
    text(24, 928, "备注（选填）", 10, 750, "#57645D"),
    "<rect x='24' y='938' width='370' height='66' rx='8' fill='#F8FAF8' stroke='#DDE7E1'/>",
    text(40, 967, edit ? "个头中等即可" : "补充品牌、规格或其他要求", 12, edit ? 500 : 400, edit ? "#48554E" : "#A2ACA7"),
    edit ? "<rect x='24' y='1024' width='370' height='44' rx='8' fill='#FFF7F6' stroke='#F3D7D3'/><path d='M139 1038h12M142 1038l1 16h8l1-16M145 1034h4' fill='none' stroke='#D85D52' stroke-width='1.4' stroke-linecap='round'/><text x='220' y='1052' text-anchor='middle' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='12' font-weight='750' fill='#D85D52'>删除采购项</text>" : "",
    "<rect x='24' y='1090' width='126' height='52' rx='26' fill='#FFFFFF' stroke='#CCD6D0'/>",
    text(87, 1123, "取消", 14, 800, "#5E6B64", "text-anchor='middle'"),
    "<rect x='160' y='1090' width='234' height='52' rx='26' fill='url(#green)'/>",
    text(277, 1123, edit ? "保存修改" : "添加到清单", 14, 850, "#FFFFFF", "text-anchor='middle'"),
    "<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>",
  ].join("");
}

function modalSvg(kind) {
  return [
    "<svg width='" + WIDTH + "' height='" + HEIGHT + "' viewBox='0 0 " + WIDTH + " " + HEIGHT + "' xmlns='http://www.w3.org/2000/svg'>",
    defs(),
    sheetOverlay(kind),
    "</svg>",
  ].join("");
}

async function prepareHero() {
  return sharp(heroSource)
    .resize(WIDTH, 224, { fit: "cover", position: "centre" })
    .modulate({ brightness: 1.02, saturation: 0.92 })
    .png()
    .toBuffer();
}

async function renderScreen(hero, svg, overlay) {
  const layers = [
    { input: hero, left: 0, top: 0 },
    { input: Buffer.from(svg), left: 0, top: 0 },
  ];
  if (overlay) layers.push({ input: Buffer.from(overlay), left: 0, top: 0 });
  return sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } })
    .composite(layers)
    .png()
    .toBuffer();
}

async function render() {
  const hero = await prepareHero();
  const base = mainSvg("pending", false);
  const outputs = [
    ["miniApp/.preview/17-shopping_list-preview.png", await renderScreen(hero, base)],
    ["miniApp/.preview/17-shopping_list-completed-preview.png", await renderScreen(hero, mainSvg("completed", false))],
    ["miniApp/.preview/17-shopping_list-checked-preview.png", await renderScreen(hero, mainSvg("pending", true))],
    ["miniApp/.preview/17-shopping_list-edit-preview.png", await renderScreen(hero, base, modalSvg("edit"))],
    ["miniApp/.preview/17-shopping_list-add-preview.png", await renderScreen(hero, base, modalSvg("add"))],
  ];
  outputs.forEach(([file, buffer]) => fs.writeFileSync(file, buffer));
  console.log(JSON.stringify({ outputs: outputs.map(([file]) => file), width: WIDTH, height: HEIGHT }, null, 2));
}

render().catch((error) => {
  console.error(error);
  process.exit(1);
});



