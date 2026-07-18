const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/prep-ai-cat-timer-v1.png";

const prepSteps = [
  {
    title: "鸡腿肉先焯水",
    color: "#D65D54",
    bg: "#FFF1EF",
    current: "鸡腿肉冷水下锅，加姜片焯水",
    parallel: "泡香菇；冬瓜去皮切块",
    done: "水开撇净浮沫，捞出鸡腿肉",
  },
  {
    title: "接着焖鸡腿肉",
    color: "#D28A2D",
    bg: "#FFF5E7",
    current: "鸡腿肉与香菇入锅，按菜谱开始焖煮",
    parallel: "西兰花分小朵洗净；蒜切末",
    done: "转小火后可离灶，进入下一步",
  },
  {
    title: "再煮冬瓜丸子汤",
    color: "#5795AD",
    bg: "#EEF7FA",
    current: "冬瓜汤煮开，丸子逐个下锅",
    parallel: "番茄切块；鸡蛋下锅前再打散",
    done: "丸子浮起后转小火保温",
  },
  {
    title: "开饭前炒快手菜",
    color: "#7B8781",
    bg: "#F1F3F2",
    current: "先焯西兰花快炒，再做番茄炒蛋",
    parallel: "检查鸡腿饭和汤的咸淡",
    done: "快手菜完成后立即上桌",
  },
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
    text(70, 96, "备菜提醒", 24, 800, "#151F1A"),
    text(70, 124, "安排顺序，从容备菜", 13, 400, "#6F7D76"),
    "<rect y='224' width='418' height='1016' fill='#F8FAF7'/>",
  ].join("");
}

function mealSummary(status = "菜单已确认") {
  return [
    "<g filter='url(#shadow)'><rect x='16' y='242' width='386' height='88' rx='8' fill='#FFFFFF' stroke='#E9EFEB'/></g>",
    text(34, 274, "周六家庭聚餐", 17, 850, "#1B241F"),
    "<rect x='307' y='254' width='77' height='26' rx='13' fill='#EAF7EF'/><circle cx='321' cy='267' r='3.5' fill='#16B96F'/>",
    text(349, 271, status, 10, 750, "#159B55", "text-anchor='middle'"),
    text(34, 304, "4 道菜 · 共 5 份 · 4 人参与", 10, 400, "#7B8982"),
  ].join("");
}

function sparkleIcon(cx, cy, color = "#159B55", accent = "#6FC89B") {
  return [
    "<path d='M" + cx + " " + (cy - 9) + "c1.5 5 4 7.5 9 9-5 1.5-7.5 4-9 9-1.5-5-4-7.5-9-9 5-1.5 7.5-4 9-9z' fill='none' stroke='" + color + "' stroke-width='1.5' stroke-linejoin='round'/>",
    "<path d='M" + (cx + 13) + " " + (cy - 11) + "c.7 2.4 1.8 3.5 4.2 4.2-2.4.7-3.5 1.8-4.2 4.2-.7-2.4-1.8-3.5-4.2-4.2 2.4-.7 3.5-1.8 4.2-4.2z' fill='" + accent + "'/>",
  ].join("");
}

function dishInputCard(x, y, name, servings) {
  return [
    "<rect x='" + x + "' y='" + y + "' width='178' height='52' rx='8' fill='#FFFFFF' stroke='#E8EEE9'/>",
    "<circle cx='" + (x + 20) + "' cy='" + (y + 26) + "' r='10' fill='#EAF7EF'/><path d='M" + (x + 16) + " " + (y + 26) + "l3 3 6-7' fill='none' stroke='#159B55' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/>",
    text(x + 38, y + 23, name, 11, 800, "#27362E"),
    text(x + 38, y + 40, servings, 9, 400, "#7B8982"),
  ].join("");
}

function readyContent() {
  return [
    mealSummary(),
    text(24, 374, "生成备菜提醒", 20, 850, "#1B241F"),
    text(24, 399, "AI 将读取最终菜单、做法、备注和制作份数", 11, 400, "#6F7D76"),
    "<g filter='url(#shadow)'><rect x='16' y='420' width='386' height='170' rx='8' fill='#FFFFFF' stroke='#E9EFEB'/></g>",
    "<circle cx='47' cy='454' r='18' fill='#EAF7EF'/>",
    sparkleIcon(47, 454),
    text(78, 451, "本次将分析 4 道菜", 14, 850, "#1B241F"),
    text(78, 472, "生成准备顺序和并行处理建议", 9, 400, "#7B8982"),
    "<line x1='32' y1='492' x2='386' y2='492' stroke='#EEF1EE'/>",
    text(34, 520, "参考信息", 10, 750, "#56645D"),
    text(112, 520, "完整做法步骤 · 菜谱备注 · 最终份数", 10, 500, "#44534B"),
    text(34, 548, "数据用途", 10, 750, "#56645D"),
    text(112, 548, "仅用于生成本次饭局备菜提醒", 10, 500, "#44534B"),
    text(34, 576, "结果影响", 10, 750, "#56645D"),
    text(112, 576, "不会修改菜单、份数或采购清单", 10, 500, "#44534B"),
    "<rect x='16' y='610' width='386' height='106' rx='8' fill='#F1F8F3'/>",
    text(80, 648, "128", 22, 850, "#1B241F", "text-anchor='middle'"),
    text(80, 669, "当前积分", 9, 400, "#7B8982", "text-anchor='middle'"),
    "<line x1='140' y1='630' x2='140' y2='688' stroke='#DFE9E2'/>",
    text(209, 648, "- 8", 22, 850, "#C7882C", "text-anchor='middle'"),
    text(209, 669, "本次消耗", 9, 400, "#7B8982", "text-anchor='middle'"),
    "<line x1='278' y1='630' x2='278' y2='688' stroke='#DFE9E2'/>",
    text(342, 648, "120", 22, 850, "#159B55", "text-anchor='middle'"),
    text(342, 669, "生成后剩余", 9, 400, "#7B8982", "text-anchor='middle'"),
    text(209, 701, "生成失败或超时将自动退还积分", 9, 500, "#728079", "text-anchor='middle'"),
    text(24, 756, "本次菜单", 17, 850, "#1B241F"),
    dishInputCard(24, 774, "香菇鸡腿饭", "2 份"),
    dishInputCard(216, 774, "番茄炒蛋", "1 份"),
    dishInputCard(24, 838, "冬瓜丸子汤", "1 份"),
    dishInputCard(216, 838, "蒜蓉西兰花", "1 份"),
    "<g filter='url(#topShadow)'><rect x='0' y='1012' width='418' height='228' fill='#FFFFFF'/></g>",
    text(24, 1052, "本次生成消耗 8 积分", 12, 800, "#1B241F"),
    text(394, 1052, "当前 128", 10, 500, "#7B8982", "text-anchor='end'"),
    "<rect x='24' y='1070' width='370' height='54' rx='27' fill='url(#green)'/>",
    sparkleIcon(132, 1097, "#FFFFFF", "#DDF5E8"),
    text(234, 1104, "生成备菜提醒", 15, 850, "#FFFFFF", "text-anchor='middle'"),
    text(209, 1154, "提醒仅作参考，可忽略或重新生成", 9, 400, "#8A9690", "text-anchor='middle'"),
    "<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>",
  ].join("");
}

function prepStepRow(y, item, index, muted) {
  const opacity = muted ? " opacity='0.42'" : "";
  const hasNext = index < prepSteps.length - 1;
  return [
    "<g" + opacity + ">",
    hasNext ? "<line x1='40' y1='" + (y + 34) + "' x2='40' y2='" + (y + 138) + "' stroke='#DDE8E1' stroke-width='2'/>" : "",
    "<circle cx='40' cy='" + (y + 22) + "' r='12' fill='" + item.bg + "' stroke='" + item.color + "' stroke-width='1'/>",
    text(40, y + 27, String(index + 1), 11, 850, item.color, "text-anchor='middle'"),
    text(64, y + 26, item.title, 14, 850, "#1B241F"),
    "<rect x='64' y='" + (y + 40) + "' width='48' height='20' rx='5' fill='#FFF1EF'/>",
    text(88, y + 54, "现在做", 9, 750, "#C95850", "text-anchor='middle'"),
    text(120, y + 54, item.current, 10, 550, "#35443C"),
    "<rect x='64' y='" + (y + 68) + "' width='48' height='20' rx='5' fill='#EAF7EF'/>",
    text(88, y + 82, "同时做", 9, 750, "#159B55", "text-anchor='middle'"),
    text(120, y + 82, item.parallel, 10, 550, "#35443C"),
    "<circle cx='70' cy='" + (y + 108) + "' r='6' fill='#F1F4F2'/><path d='M67 " + (y + 108) + "l2 2 4-5' fill='none' stroke='#7B8982' stroke-width='1.2' stroke-linecap='round' stroke-linejoin='round'/>",
    text(82, y + 112, "完成标志：" + item.done, 9, 400, "#7B8982"),
    index < prepSteps.length - 1 ? "<line x1='64' y1='" + (y + 124) + "' x2='386' y2='" + (y + 124) + "' stroke='#EEF1EE'/>" : "",
    "</g>",
  ].join("");
}

function resultContent(ignored = false) {
  const parts = [
    mealSummary(ignored ? "提醒已忽略" : "提醒已生成"),
    text(24, 370, ignored ? "已忽略的顺序" : "建议操作顺序", 20, 850, "#1B241F"),
    text(394, 369, "今天 10:24 生成", 9, 400, "#7B8982", "text-anchor='end'"),
    "<rect x='16' y='386' width='386' height='36' rx='8' fill='" + (ignored ? "#F3F5F4" : "#EFF8F2") + "'/>",
    "<path d='M34 404h10m-4-4 4 4-4 4' fill='none' stroke='" + (ignored ? "#8A9690" : "#159B55") + "' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/>",
    text(54, 409, ignored ? "这份顺序不会再主动展示" : "按菜谱依赖排序，等待时穿插其他准备", 10, 650, ignored ? "#6F7D76" : "#365146"),
    "<g filter='url(#shadow)'><rect x='16' y='430' width='386' height='532' rx='8' fill='#FFFFFF' stroke='#E9EFEB'/></g>",
  ];
  prepSteps.forEach((item, index) => parts.push(prepStepRow(438 + index * 128, item, index, ignored)));
  parts.push(
    "<rect x='16' y='974' width='386' height='40' rx='8' fill='" + (ignored ? "#F4F6F5" : "#F1F8F3") + "'/>",
    "<circle cx='36' cy='994' r='8' fill='" + (ignored ? "#E7EAE8" : "#DFF3E7") + "'/><path d='M32 994l3 3 5-6' fill='none' stroke='" + (ignored ? "#8A9690" : "#159B55") + "' stroke-width='1.4' stroke-linecap='round' stroke-linejoin='round'/>",
    text(52, 998, ignored ? "需要时可重新生成一份新顺序" : "仅重排原菜谱步骤，不新增食材或做法", 9, 500, "#6F7D76"),
    "<g filter='url(#topShadow)'><rect x='0' y='1030' width='418' height='210' fill='#FFFFFF'/></g>"
  );
  if (ignored) {
    parts.push(
      "<rect x='24' y='1070' width='370' height='52' rx='26' fill='url(#green)'/>",
      "<path d='M122 1096a10 10 0 1 1 3 7M122 1096v-7h7' fill='none' stroke='#FFFFFF' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'/>",
      text(236, 1103, "重新生成提醒", 15, 850, "#FFFFFF", "text-anchor='middle'"),
      text(209, 1152, "重新生成将再次消耗 8 积分", 9, 400, "#8A9690", "text-anchor='middle'")
    );
  } else {
    parts.push(
      "<rect x='24' y='1070' width='126' height='50' rx='25' fill='#FFFFFF' stroke='#CCD6D0'/>",
      text(87, 1101, "忽略本次", 13, 800, "#66736C", "text-anchor='middle'"),
      "<rect x='160' y='1070' width='234' height='50' rx='25' fill='url(#green)'/>",
      "<path d='M218 1095a9 9 0 1 1 3 6M218 1095v-6h6' fill='none' stroke='#FFFFFF' stroke-width='1.7' stroke-linecap='round' stroke-linejoin='round'/>",
      text(300, 1101, "重新生成", 14, 850, "#FFFFFF", "text-anchor='middle'"),
      text(209, 1152, "重新生成将再次消耗 8 积分", 9, 400, "#8A9690", "text-anchor='middle'")
    );
  }
  parts.push("<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>");
  return parts.join("");
}
function regenerateOverlay() {
  return [
    "<rect x='0' y='0' width='418' height='1240' fill='#172019' fill-opacity='0.38'/>",
    "<path d='M0 748Q0 732 16 732H402Q418 732 418 748V1240H0Z' fill='#FFFFFF'/>",
    "<rect x='177' y='748' width='64' height='5' rx='2.5' fill='#DDE4DF'/>",
    "<circle cx='49' cy='804' r='20' fill='#EAF7EF'/><path d='M42 805a8 8 0 1 1 3 6M42 805v-6h6' fill='none' stroke='#159B55' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'/>",
    text(80, 800, "重新生成备菜提醒？", 20, 850, "#1B241F"),
    text(80, 824, "新结果生成成功后将替换当前提醒", 10, 400, "#7B8982"),
    "<line x1='24' y1='852' x2='394' y2='852' stroke='#EEF1EE'/>",
    text(82, 892, "128", 21, 850, "#1B241F", "text-anchor='middle'"),
    text(82, 914, "当前积分", 9, 400, "#7B8982", "text-anchor='middle'"),
    "<line x1='139' y1='870' x2='139' y2='920' stroke='#EEF1EE'/>",
    text(209, 892, "- 8", 21, 850, "#C7882C", "text-anchor='middle'"),
    text(209, 914, "本次消耗", 9, 400, "#7B8982", "text-anchor='middle'"),
    "<line x1='279' y1='870' x2='279' y2='920' stroke='#EEF1EE'/>",
    text(345, 892, "120", 21, 850, "#159B55", "text-anchor='middle'"),
    text(345, 914, "生成后剩余", 9, 400, "#7B8982", "text-anchor='middle'"),
    "<rect x='24' y='946' width='370' height='58' rx='8' fill='#F1F8F3'/>",
    text(42, 971, "生成失败或超时会自动退还积分", 10, 750, "#44534B"),
    text(42, 990, "旧提醒在新结果成功前仍会保留", 9, 400, "#7B8982"),
    "<rect x='24' y='1032' width='126' height='52' rx='26' fill='#FFFFFF' stroke='#CCD6D0'/>",
    text(87, 1065, "取消", 14, 800, "#5E6B64", "text-anchor='middle'"),
    "<rect x='160' y='1032' width='234' height='52' rx='26' fill='url(#green)'/>",
    text(277, 1065, "消耗 8 积分生成", 14, 850, "#FFFFFF", "text-anchor='middle'"),
    "<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>",
  ].join("");
}

function mainSvg(state) {
  const body = state === "ready" ? readyContent() : resultContent(state === "ignored");
  return [
    "<svg width='" + WIDTH + "' height='" + HEIGHT + "' viewBox='0 0 " + WIDTH + " " + HEIGHT + "' xmlns='http://www.w3.org/2000/svg'>",
    defs(),
    chrome(),
    body,
    "</svg>",
  ].join("");
}

function modalSvg() {
  return [
    "<svg width='" + WIDTH + "' height='" + HEIGHT + "' viewBox='0 0 " + WIDTH + " " + HEIGHT + "' xmlns='http://www.w3.org/2000/svg'>",
    defs(),
    regenerateOverlay(),
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
  const resultSvg = mainSvg("result");
  const outputs = [
    ["miniApp/.preview/18-prep_ai-ready-preview.png", await renderScreen(hero, mainSvg("ready"))],
    ["miniApp/.preview/18-prep_ai-result-preview.png", await renderScreen(hero, resultSvg)],
    ["miniApp/.preview/18-prep_ai-regenerate-preview.png", await renderScreen(hero, resultSvg, modalSvg())],
    ["miniApp/.preview/18-prep_ai-ignored-preview.png", await renderScreen(hero, mainSvg("ignored"))],
  ];
  outputs.forEach(([file, buffer]) => fs.writeFileSync(file, buffer));
  console.log(JSON.stringify({ outputs: outputs.map(([file]) => file), width: WIDTH, height: HEIGHT }, null, 2));
}

render().catch((error) => {
  console.error(error);
  process.exit(1);
});




