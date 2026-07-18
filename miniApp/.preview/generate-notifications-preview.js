const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/notifications-cat-envelope-v1.png";

const notifications = [
  { group: "今天", time: "10:25", title: "AI 备菜提醒已退款", line1: "本次生成失败，8 积分已退回", line2: "退款流水已记录", type: "refund", unread: true },
  { group: "今天", time: "09:40", title: "点菜已截止", line1: "「周六家庭聚餐」已停止收集", line2: "请确认最终菜单和制作份数", type: "meal", unread: true },
  { group: "今天", time: "08:20", title: "菜品已取消推荐", line1: "「香菇鸡腿饭」已从首页推荐移除", line2: "原因：平台推荐内容调整", type: "governance", unread: true },
  { group: "昨天", time: "16:32", title: "允许被发现已关闭", line1: "「红烧排骨」因封面问题被关闭", line2: "原因：图片包含无关文字", type: "governance", unread: false },
  { group: "昨天", time: "12:08", title: "打卡积分到账", line1: "今日首次打卡奖励 10 积分", line2: "当前积分已更新", type: "points", unread: false },
  { group: "7 月 16 日", time: "18:15", title: "菜单已确认", line1: "「周五晚餐」采购清单已生成", line2: "可从我的页面再次进入", type: "meal", unread: false },
];

function text(x, y, value, size, weight, fill, extra = "") {
  return "<text x='" + x + "' y='" + y + "' font-family='Microsoft YaHei, PingFang SC, sans-serif' font-size='" + size + "' font-weight='" + weight + "' fill='" + fill + "' " + extra + ">" + value + "</text>";
}

function defs() {
  return "<defs><filter id='shadow' x='-8%' y='-20%' width='116%' height='160%'><feDropShadow dx='0' dy='5' stdDeviation='9' flood-color='#10261A' flood-opacity='0.04'/></filter></defs>";
}

function chrome() {
  return [
    text(44, 39, "9:41", 15, 700, "#151F1A"),
    "<g fill='#151F1A'><rect x='326' y='32' width='4' height='8' rx='1'/><rect x='332' y='28' width='4' height='12' rx='1'/><rect x='338' y='24' width='4' height='16' rx='1'/>",
    "<path d='M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0' fill='none' stroke='#151F1A' stroke-width='2' stroke-linecap='round'/>",
    "<rect x='373' y='28' width='21' height='11' rx='3' fill='none' stroke='#151F1A' stroke-width='1.6'/><rect x='395' y='31' width='2' height='5' rx='1'/><rect x='376' y='31' width='15' height='5' rx='1'/></g>",
    "<g><rect x='312' y='60' width='82' height='36' rx='18' fill='#FFFFFF' fill-opacity='0.92' stroke='#E7E9E5'/><circle cx='338' cy='76' r='2.4' fill='#151F1A'/><circle cx='348' cy='76' r='2.4' fill='#151F1A'/><line x1='359' y1='66' x2='359' y2='86' stroke='#E7E9E5'/><circle cx='378' cy='76' r='9' fill='none' stroke='#151F1A' stroke-width='2.6'/></g>",
    "<path d='M48 79L38 89l10 10M39 89h20' fill='none' stroke='#151F1A' stroke-width='2.8' stroke-linecap='round' stroke-linejoin='round'/>",
    text(70, 96, "通知中心", 24, 800, "#151F1A"),
    text(70, 124, "饭局和重要变化都在这里", 13, 400, "#6F7D76"),
    "<rect y='224' width='418' height='1016' fill='#F8FAF7'/>",
  ].join("");
}

function toolbar(state) {
  const allRead = state === "read" || state === "empty";
  const unreadCount = allRead ? 0 : 3;
  return [
    "<rect x='0' y='224' width='418' height='68' fill='#FFFFFF'/>",
    text(24, 265, unreadCount ? unreadCount + " 条未读" : "没有未读通知", 15, 800, "#24342C"),
    text(394, 265, allRead ? "已全部读" : "全部已读", 11, allRead ? 500 : 750, allRead ? "#A2AAA6" : "#159B55", "text-anchor='end'"),
  ].join("");
}

function tabs(active) {
  return [
    "<rect x='0' y='292' width='418' height='54' fill='#FFFFFF'/><line x1='16' y1='345' x2='402' y2='345' stroke='#EEF1EE'/>",
    text(112, 328, "未读", 13, active === "unread" ? 800 : 500, active === "unread" ? "#159B55" : "#7B8982", "text-anchor='middle'"),
    text(306, 328, "全部", 13, active === "all" ? 800 : 500, active === "all" ? "#159B55" : "#7B8982", "text-anchor='middle'"),
    "<rect x='" + (active === "unread" ? 94 : 288) + "' y='342' width='36' height='3' rx='1.5' fill='#16B96F'/>",
  ].join("");
}

function typeIcon(x, y, type, muted) {
  const palettes = {
    refund: { bg: "#EEF7FA", color: "#5795AD" },
    meal: { bg: "#EAF7EF", color: "#159B55" },
    governance: { bg: "#FFF1EF", color: "#D65D54" },
    points: { bg: "#FFF5E7", color: "#C7882C" },
  };
  const p = palettes[type];
  const opacity = muted ? " opacity='0.62'" : "";
  const out = ["<g" + opacity + "><circle cx='" + x + "' cy='" + y + "' r='19' fill='" + p.bg + "'/><g fill='none' stroke='" + p.color + "' stroke-width='1.7' stroke-linecap='round' stroke-linejoin='round'>"];
  if (type === "refund") out.push("<path d='M" + (x + 7) + " " + (y - 5) + "a9 9 0 1 0 1 9M" + (x + 7) + " " + (y - 5) + "v7h-7'/>");
  if (type === "meal") out.push("<rect x='" + (x - 8) + "' y='" + (y - 7) + "' width='16' height='15' rx='3'/><path d='M" + (x - 4) + " " + (y - 10) + "v5M" + (x + 4) + " " + (y - 10) + "v5M" + (x - 4) + " " + (y + 1) + "l3 3 5-6'/>");
  if (type === "governance") out.push("<path d='M" + x + " " + (y - 10) + "l8 3v6c0 6-3 9-8 11-5-2-8-5-8-11v-6z'/><path d='M" + (x - 4) + " " + y + "h8'/>");
  if (type === "points") out.push("<circle cx='" + x + "' cy='" + y + "' r='9'/><path d='M" + x + " " + (y - 5) + "v10M" + (x - 5) + " " + y + "h10'/>");
  out.push("</g></g>");
  return out.join("");
}

function notificationRow(y, item, forceRead, isLast) {
  const unread = item.unread && !forceRead;
  return [
    unread ? "<circle cx='18' cy='" + (y + 27) + "' r='4' fill='#16B96F'/>" : "",
    typeIcon(44, y + 36, item.type, !unread),
    text(74, y + 27, item.title, 13, unread ? 850 : 650, unread ? "#24342C" : "#56645D"),
    text(394, y + 26, item.time, 9, 400, "#9AA49F", "text-anchor='end'"),
    text(74, y + 52, item.line1, 10, 500, unread ? "#526159" : "#7B8982"),
    text(74, y + 73, item.line2, 9, 400, "#8A9690"),
    isLast ? "" : "<line x1='74' y1='" + (y + 95) + "' x2='402' y2='" + (y + 95) + "' stroke='#EEF1EE'/>",
  ].join("");
}

function listContent(state) {
  if (state === "empty") {
    return [
      "<rect x='0' y='346' width='418' height='894' fill='#FFFFFF'/>",
      "<circle cx='209' cy='512' r='34' fill='#EAF7EF'/><path d='M198 519h22M201 519v-12a8 8 0 0 1 16 0v12M205 524c2 5 6 5 8 0' fill='none' stroke='#159B55' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'/>",
      text(209, 574, "暂时没有未读通知", 15, 800, "#35443C", "text-anchor='middle'"),
      text(209, 600, "新的饭局和重要变化会出现在这里", 10, 400, "#8A9690", "text-anchor='middle'"),
      "<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>",
    ].join("");
  }
  const onlyUnread = state === "unread";
  const forceRead = state === "read";
  const items = onlyUnread ? notifications.filter((item) => item.unread) : notifications;
  const groups = [];
  items.forEach((item) => {
    let group = groups.find((entry) => entry.name === item.group);
    if (!group) { group = { name: item.group, items: [] }; groups.push(group); }
    group.items.push(item);
  });
  const out = ["<rect x='0' y='346' width='418' height='894' fill='#FFFFFF'/>"];
  let y = 364;
  groups.forEach((group) => {
    out.push(text(24, y + 16, group.name, 11, 750, "#66736C"));
    y += 24;
    group.items.forEach((item, index) => {
      out.push(notificationRow(y, item, forceRead, index === group.items.length - 1));
      y += 96;
    });
    y += 6;
  });
  out.push("<rect x='147' y='1215' width='124' height='5' rx='2.5' fill='#303833'/>");
  return out.join("");
}

function screenSvg(state) {
  const active = state === "unread" || state === "empty" ? "unread" : "all";
  return ["<svg width='418' height='1240' viewBox='0 0 418 1240' xmlns='http://www.w3.org/2000/svg'>", defs(), chrome(), toolbar(state), tabs(active), listContent(state), "</svg>"].join("");
}

async function prepareHero() {
  return sharp(heroSource).resize(WIDTH, 224, { fit: "cover", position: "centre" }).modulate({ brightness: 1.015, saturation: 0.92 }).png().toBuffer();
}

async function renderScreen(hero, state) {
  return sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } })
    .composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(screenSvg(state)), left: 0, top: 0 }])
    .png().toBuffer();
}

async function render() {
  const hero = await prepareHero();
  const outputs = [];
  for (const state of ["all", "unread", "read", "empty"]) {
    const suffix = state === "all" ? "" : "-" + state;
    const file = "miniApp/.preview/20-notifications" + suffix + "-preview.png";
    outputs.push([file, await renderScreen(hero, state)]);
  }
  outputs.forEach(([file, buffer]) => fs.writeFileSync(file, buffer));
  console.log(JSON.stringify({ outputs: outputs.map(([file]) => file), width: WIDTH, height: HEIGHT }, null, 2));
}

render().catch((error) => { console.error(error); process.exit(1); });

