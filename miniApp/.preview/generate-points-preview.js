const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/points-cat-jar-v1.png";

const transactions = [
  { group: "今天", time: "10:24", title: "AI 备菜提醒", desc: "周六家庭聚餐", amount: -8, type: "spent", icon: "sparkle" },
  { group: "今天", time: "09:16", title: "做菜打卡", desc: "今日首次打卡奖励", amount: 10, type: "earned", icon: "check" },
  { group: "今天", time: "08:42", title: "AI 菜品解析失败退还", desc: "关联 08:41 AI 菜品解析", amount: 5, type: "refund", icon: "refund" },
  { group: "今天", time: "08:41", title: "AI 菜品解析", desc: "文本识别菜品信息", amount: -5, type: "spent", icon: "scan" },
  { group: "昨天", time: "20:05", title: "AI 生成菜品封面", desc: "香菇鸡腿饭", amount: -6, type: "spent", icon: "image" },
  { group: "昨天", time: "18:42", title: "做菜打卡", desc: "每日首次打卡奖励", amount: 10, type: "earned", icon: "check" },
  { group: "7 月 16 日", time: "14:30", title: "系统发放", desc: "活动奖励", amount: 20, type: "earned", icon: "gift" },
];

const tabs = [
  { key: "all", label: "全部" },
  { key: "earned", label: "获得" },
  { key: "spent", label: "消耗" },
  { key: "refund", label: "退还" },
];

function text(x, y, value, size, weight, fill, extra = "") {
  return "<text x='" + x + "' y='" + y + "' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='" + size + "' font-weight='" + weight + "' fill='" + fill + "' " + extra + ">" + value + "</text>";
}

function defs() {
  return [
    "<defs>",
    "<filter id='shadow' x='-8%' y='-20%' width='116%' height='160%'><feDropShadow dx='0' dy='5' stdDeviation='9' flood-color='#10261A' flood-opacity='0.045'/></filter>",
    "</defs>",
  ].join("");
}

function chrome() {
  return [
    text(44, 39, "9:41", 15, 700, "#151F1A"),
    "<g fill='#151F1A'><rect x='326' y='32' width='4' height='8' rx='1'/><rect x='332' y='28' width='4' height='12' rx='1'/><rect x='338' y='24' width='4' height='16' rx='1'/>",
    "<path d='M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0' fill='none' stroke='#151F1A' stroke-width='2' stroke-linecap='round'/>",
    "<rect x='373' y='28' width='21' height='11' rx='3' fill='none' stroke='#151F1A' stroke-width='1.6'/><rect x='395' y='31' width='2' height='5' rx='1'/><rect x='376' y='31' width='15' height='5' rx='1'/></g>",
    "<g><rect x='312' y='60' width='82' height='36' rx='18' fill='#FFFFFF' fill-opacity='0.92' stroke='#E7E9E5'/><circle cx='338' cy='76' r='2.4' fill='#151F1A'/><circle cx='348' cy='76' r='2.4' fill='#151F1A'/><line x1='359' y1='66' x2='359' y2='86' stroke='#E7E9E5'/><circle cx='378' cy='76' r='9' fill='none' stroke='#151F1A' stroke-width='2.6'/></g>",
    "<path d='M48 79L38 89l10 10M39 89h20' fill='none' stroke='#151F1A' stroke-width='2.8' stroke-linecap='round' stroke-linejoin='round'/>",
    text(70, 96, "积分明细", 24, 800, "#151F1A"),
    text(70, 124, "每一笔变化都有记录", 13, 400, "#6F7D76"),
    "<rect y='224' width='418' height='1016' fill='#F8FAF7'/>",
  ].join("");
}

function summary() {
  return [
    "<g filter='url(#shadow)'><rect x='16' y='242' width='386' height='128' rx='8' fill='#FFFFFF' stroke='#E9EFEB'/></g>",
    text(34, 274, "当前积分", 11, 650, "#6F7D76"),
    text(34, 322, "128", 36, 850, "#159B55"),
    text(34, 346, "可用于 AI 功能", 10, 400, "#7B8982"),
    "<line x1='204' y1='260' x2='204' y2='352' stroke='#EEF1EE'/>",
    text(269, 300, "+40", 20, 850, "#159B55", "text-anchor='middle'"),
    text(269, 324, "本月获得", 9, 400, "#7B8982", "text-anchor='middle'"),
    text(351, 300, "-19", 20, 850, "#C7882C", "text-anchor='middle'"),
    text(351, 324, "本月消耗", 9, 400, "#7B8982", "text-anchor='middle'"),
  ].join("");
}

function tabBar(active) {
  const parts = ["<rect x='0' y='390' width='418' height='58' fill='#FFFFFF'/><line x1='16' y1='447' x2='402' y2='447' stroke='#EEF1EE'/>"];
  tabs.forEach((tab, index) => {
    const center = 64 + index * 96.5;
    const isActive = tab.key === active;
    parts.push(text(center, 425, tab.label, 13, isActive ? 800 : 500, isActive ? "#159B55" : "#7B8982", "text-anchor='middle'"));
    if (isActive) parts.push("<rect x='" + (center - 18) + "' y='444' width='36' height='3' rx='1.5' fill='#16B96F'/>");
  });
  return parts.join("");
}

function rowIcon(x, y, icon, type) {
  const palette = type === "spent" ? { bg: "#FFF5E7", color: "#C7882C" } : type === "refund" ? { bg: "#EEF7FA", color: "#5795AD" } : { bg: "#EAF7EF", color: "#159B55" };
  const parts = ["<circle cx='" + x + "' cy='" + y + "' r='19' fill='" + palette.bg + "'/><g fill='none' stroke='" + palette.color + "' stroke-width='1.7' stroke-linecap='round' stroke-linejoin='round'>"];
  if (icon === "sparkle") parts.push("<path d='M" + x + " " + (y - 10) + "c1.5 5 4 7.5 9 9-5 1.5-7.5 4-9 9-1.5-5-4-7.5-9-9 5-1.5 7.5-4 9-9z'/>");
  if (icon === "check") parts.push("<rect x='" + (x - 8) + "' y='" + (y - 8) + "' width='16' height='16' rx='4'/><path d='M" + (x - 4) + " " + y + "l3 3 6-7'/>");
  if (icon === "refund") parts.push("<path d='M" + (x + 7) + " " + (y - 5) + "a9 9 0 1 0 1 9M" + (x + 7) + " " + (y - 5) + "v7h-7'/>");
  if (icon === "scan") parts.push("<path d='M" + (x - 8) + " " + (y - 3) + "v-5h5M" + (x + 8) + " " + (y - 3) + "v-5h-5M" + (x - 8) + " " + (y + 3) + "v5h5M" + (x + 8) + " " + (y + 3) + "v5h-5M" + (x - 4) + " " + y + "h8'/>");
  if (icon === "image") parts.push("<rect x='" + (x - 9) + "' y='" + (y - 8) + "' width='18' height='16' rx='3'/><circle cx='" + (x + 4) + "' cy='" + (y - 3) + "' r='2'/><path d='M" + (x - 6) + " " + (y + 5) + "l5-5 4 4 2-2 3 3'/>");
  if (icon === "gift") parts.push("<rect x='" + (x - 8) + "' y='" + (y - 2) + "' width='16' height='11' rx='2'/><path d='M" + x + " " + (y - 2) + "v11M" + (x - 10) + " " + (y - 6) + "h20v4h-20zM" + x + " " + (y - 6) + "c-2-7-9-4-5 0M" + x + " " + (y - 6) + "c2-7 9-4 5 0'/>");
  parts.push("</g>");
  return parts.join("");
}

function transactionRow(y, item, isLast) {
  const amountColor = item.type === "spent" ? "#C7882C" : item.type === "refund" ? "#5795AD" : "#159B55";
  const amount = (item.amount > 0 ? "+" : "") + item.amount;
  return [
    rowIcon(38, y + 35, item.icon, item.type),
    text(68, y + 29, item.title, 13, 800, "#24342C"),
    text(68, y + 50, item.desc, 9, 400, "#7B8982"),
    text(382, y + 30, amount, 16, 850, amountColor, "text-anchor='end'"),
    text(382, y + 51, item.time, 9, 400, "#97A19C", "text-anchor='end'"),
    isLast ? "" : "<line x1='68' y1='" + (y + 75) + "' x2='402' y2='" + (y + 75) + "' stroke='#EEF1EE'/>",
  ].join("");
}

function visibleTransactions(active) {
  if (active === "all") return transactions;
  return transactions.filter((item) => item.type === active);
}

function listContent(active) {
  const items = visibleTransactions(active);
  const groups = [];
  items.forEach((item) => {
    let group = groups.find((entry) => entry.name === item.group);
    if (!group) {
      group = { name: item.group, items: [] };
      groups.push(group);
    }
    group.items.push(item);
  });
  const parts = ["<rect x='0' y='448' width='418' height='792' fill='#FFFFFF'/>"];
  let y = 468;
  groups.forEach((group) => {
    parts.push(text(24, y + 16, group.name, 11, 750, "#66736C"));
    parts.push(text(394, y + 16, group.items.length + " 笔", 9, 400, "#9AA49F", "text-anchor='end'"));
    y += 24;
    group.items.forEach((item, index) => {
      parts.push(transactionRow(y, item, index === group.items.length - 1));
      y += 76;
    });
    y += 4;
  });
  if (active !== "all" && y < 900) {
    const labels = { earned: "本月共获得 40 积分", spent: "本月共消耗 19 积分", refund: "本月共退还 5 积分" };
    parts.push("<rect x='24' y='" + (y + 12) + "' width='370' height='46' rx='8' fill='#F5F8F6'/>");
    parts.push(text(42, y + 40, labels[active], 10, 650, "#65736B"));
  }
  parts.push("<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>");
  return parts.join("");
}

function screenSvg(active) {
  return [
    "<svg width='" + WIDTH + "' height='" + HEIGHT + "' viewBox='0 0 " + WIDTH + " " + HEIGHT + "' xmlns='http://www.w3.org/2000/svg'>",
    defs(), chrome(), summary(), tabBar(active), listContent(active), "</svg>",
  ].join("");
}

async function prepareHero() {
  return sharp(heroSource).resize(WIDTH, 224, { fit: "cover", position: "centre" }).modulate({ brightness: 1.015, saturation: 0.92 }).png().toBuffer();
}

async function renderScreen(hero, active) {
  return sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } })
    .composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(screenSvg(active)), left: 0, top: 0 }])
    .png()
    .toBuffer();
}

async function render() {
  const hero = await prepareHero();
  const outputs = [];
  for (const state of tabs.map((tab) => tab.key)) {
    const suffix = state === "all" ? "" : "-" + state;
    const file = "miniApp/.preview/19-points" + suffix + "-preview.png";
    outputs.push([file, await renderScreen(hero, state)]);
  }
  outputs.forEach(([file, buffer]) => fs.writeFileSync(file, buffer));
  console.log(JSON.stringify({ outputs: outputs.map(([file]) => file), width: WIDTH, height: HEIGHT }, null, 2));
}

render().catch((error) => {
  console.error(error);
  process.exit(1);
});

