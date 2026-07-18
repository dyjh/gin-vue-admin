const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;
const heroSource = "miniApp/.preview/assets/profile-cat-washing-v3.png";

function defs() {
  return `<defs>
    <filter id="shadow" x="-10%" y="-20%" width="120%" height="160%"><feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261A" flood-opacity="0.055"/></filter>
    <filter id="tabShadow" x="-10%" y="-40%" width="120%" height="150%"><feDropShadow dx="0" dy="-4" stdDeviation="7" flood-color="#163322" flood-opacity="0.06"/></filter>
    <linearGradient id="green" x1="0" x2="1"><stop offset="0" stop-color="#20B866"/><stop offset="1" stop-color="#0A8E52"/></linearGradient>
  </defs>`;
}

function profileIdentity(profileReady = true) {
  if (!profileReady) {
    return `
  <circle cx="58" cy="105" r="31" fill="#FFFFFF" stroke="#E4EEE7" stroke-width="3"/><circle cx="58" cy="105" r="26" fill="#E8F7EE"/>
  <circle cx="58" cy="99" r="7" fill="none" stroke="#159B55" stroke-width="1.8"/><path d="M45 119c2-9 8-13 13-13s11 4 13 13" fill="none" stroke="#159B55" stroke-width="1.8" stroke-linecap="round"/>
  <circle cx="79" cy="123" r="10" fill="#159B55"/><path d="M74 123h10M79 118v10" fill="none" stroke="#FFFFFF" stroke-width="1.7" stroke-linecap="round"/>
  <text x="102" y="101" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="850" fill="#1B241F">完善个人资料</text>
  <text x="102" y="128" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#159B55">设置头像和昵称</text>
  <path d="M204 101l5 5-5 5" fill="none" stroke="#8A958F" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>`;
  }

  return `
  <circle cx="58" cy="105" r="31" fill="#FFFFFF" stroke="#E4EEE7" stroke-width="3"/><circle cx="58" cy="105" r="26" fill="#E8F7EE"/>
  <text x="58" y="114" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="23" font-weight="850" fill="#159B55">厨</text>
  <text x="102" y="101" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="20" font-weight="850" fill="#1B241F">厨房小记</text>
  <text x="102" y="128" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">点击编辑头像昵称</text>
  <path d="M205 101l5 5-5 5" fill="none" stroke="#8A958F" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>`;
}

function header(profileReady = true) {
  return `
  <rect x="0" y="0" width="418" height="224" fill="#FFFFFF" fill-opacity="0.1"/>
  <text x="44" y="39" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="#151F1A">9:41</text>
  <g fill="#151F1A"><rect x="326" y="32" width="4" height="8" rx="1"/><rect x="332" y="28" width="4" height="12" rx="1"/><rect x="338" y="24" width="4" height="16" rx="1"/>
    <path d="M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0" fill="none" stroke="#151F1A" stroke-width="2" stroke-linecap="round"/>
    <rect x="373" y="28" width="21" height="11" rx="3" fill="none" stroke="#151F1A" stroke-width="1.6"/><rect x="395" y="31" width="2" height="5" rx="1"/><rect x="376" y="31" width="15" height="5" rx="1"/>
  </g>
  <g><rect x="312" y="60" width="82" height="36" rx="18" fill="#FFFFFF" fill-opacity="0.9" stroke="#E7E9E5"/><circle cx="338" cy="76" r="2.4" fill="#151F1A"/><circle cx="348" cy="76" r="2.4" fill="#151F1A"/><line x1="359" y1="66" x2="359" y2="86" stroke="#E7E9E5"/><circle cx="378" cy="76" r="9" fill="none" stroke="#151F1A" stroke-width="2.6"/></g>
  ${profileIdentity(profileReady)}
  <rect x="0" y="148" width="418" height="76" fill="#FFFFFF" fill-opacity="0.93"/>
  <line x1="0" y1="148" x2="418" y2="148" stroke="#E3EAE5"/><line x1="209" y1="161" x2="209" y2="211" stroke="#DFE6E1"/>
  <text x="110" y="182" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#159B55">128</text>
  <text x="110" y="204" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#6F7D76">当前积分</text>
  <text x="308" y="182" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">18 天</text>
  <text x="308" y="204" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#6F7D76">累计打卡</text>
  <rect y="224" width="${WIDTH}" height="${HEIGHT - 224}" fill="#F8FAF7"/>`;
}

function profileCard() {
  return `
  <g filter="url(#shadow)"><rect x="24" y="188" width="370" height="150" rx="18" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <circle cx="72" cy="229" r="34" fill="#FFFFFF" stroke="#E4EEE7" stroke-width="4"/><circle cx="72" cy="229" r="28" fill="#E8F7EE"/>
  <text x="72" y="238" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="850" fill="#159B55">厨</text>
  <text x="116" y="220" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">厨房小记</text>
  <text x="116" y="247" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">微信用户</text>
  <path d="M365 219l6 6-6 6" fill="none" stroke="#9AA49E" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
  <line x1="43" y1="270" x2="375" y2="270" stroke="#EEF1EE"/><line x1="209" y1="284" x2="209" y2="321" stroke="#E8EDE9"/>
  <text x="108" y="303" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#159B55">128</text>
  <text x="108" y="325" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">当前积分</text>
  <text x="293" y="303" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#1B241F">18 天</text>
  <text x="293" y="325" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#6F7D76">累计打卡</text>`;
}

function checkinCard() {
  return `
  <g filter="url(#shadow)"><rect x="24" y="390" width="370" height="88" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <circle cx="66" cy="434" r="24" fill="#E9F8EF"/>
  <path d="M57 421h18v22H57zM61 417h10v5H61M61 431l4 4 8-9" fill="none" stroke="#159B55" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="104" y="425" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#1B241F">做菜打卡</text>
  <text x="104" y="450" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">记录今天的一餐，可获得打卡积分</text>
  <rect x="313" y="410" width="64" height="36" rx="18" fill="url(#green)"/><text x="345" y="434" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#FFFFFF">去打卡</text>`;
}

function mealActions() {
  return `
  <text x="24" y="516" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">饭局</text>
  <rect x="24" y="532" width="180" height="64" rx="15" fill="url(#green)"/>
  <path d="M50 549h18v18H50zM55 545h8v5h-8M72 552h8m-4-4v8" fill="none" stroke="#FFFFFF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="128" y="570" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#FFFFFF">创建饭局</text>
  <rect x="214" y="532" width="180" height="64" rx="15" fill="#FFFFFF" stroke="#B9DFC9"/>
  <path d="M241 548h19v20h-19zM246 553h9M246 558h9M264 551l4 4 7-8" fill="none" stroke="#159B55" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="318" y="570" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#159B55">加入饭局</text>`;
}

function activeMeal() {
  return `
  <text x="24" y="638" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">进行中的饭局</text>
  <text x="370" y="637" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#159B55">查看全部</text><path d="M379 629l5 5-5 5" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <g filter="url(#shadow)"><rect x="24" y="654" width="370" height="130" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <text x="43" y="686" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">周六家宴</text>
  <rect x="313" y="667" width="62" height="28" rx="14" fill="#E9F8EF"/><circle cx="326" cy="681" r="4" fill="#20B866"/><text x="349" y="685" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">收集中</text>
  <text x="43" y="712" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">3 人已参与 · 6 道候选菜</text>
  <rect x="43" y="730" width="332" height="38" rx="11" fill="#F3F7F4"/>
  <circle cx="62" cy="749" r="9" fill="#FFF2D9"/><path d="M62 744v6l4 2" fill="none" stroke="#C98A24" stroke-width="1.5" stroke-linecap="round"/>
  <text x="80" y="754" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#526159">周五 20:00 前关闭点菜</text><path d="M354 744l5 5-5 5" fill="none" stroke="#8A958F" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>`;
}

function services() {
  const rows = [
    { y: 842, label: "历史饭局", meta: "12 场", icon: "history" },
    { y: 893, label: "采购清单", meta: "3 项待购买", icon: "cart" },
    { y: 944, label: "通知中心", meta: "2 条未读", icon: "bell" },
    { y: 995, label: "AI 使用记录", meta: "本月 8 次", icon: "spark" },
  ];
  const icons = {
    history: (y) => `<path d="M48 ${y + 24}a10 10 0 1 0 3-7M48 ${y + 15}v8h8M58 ${y + 22}v7l5 3"/>`,
    cart: (y) => `<path d="M48 ${y + 17}h4l3 14h12l3-10H54M57 ${y + 36}h.1M67 ${y + 36}h.1"/>`,
    bell: (y) => `<path d="M51 ${y + 30}h18l-3-4v-5a6 6 0 0 0-12 0v5zM58 ${y + 34}c1 3 4 3 5 0"/>`,
    spark: (y) => `<path d="M60 ${y + 16}l2 6 6 2-6 2-2 6-2-6-6-2 6-2zM69 ${y + 31}l1 3 3 1-3 1-1 3-1-3-3-1 3-1z"/>`,
  };
  return `<text x="24" y="824" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">常用服务</text>
  <g filter="url(#shadow)"><rect x="24" y="838" width="370" height="204" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  ${rows.map((row, index) => `<g fill="none" stroke="#159B55" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">${icons[row.icon](row.y)}</g>
    <text x="90" y="${row.y + 31}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="750" fill="#1B241F">${row.label}</text>
    <text x="354" y="${row.y + 31}" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="${index === 1 ? "#C98A24" : index === 2 ? "#C84D49" : "#8A958F"}">${row.meta}</text>
    <path d="M371 ${row.y + 23}l5 5-5 5" fill="none" stroke="#A6B0AA" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>${index < rows.length - 1 ? `<line x1="90" y1="${row.y + 50}" x2="376" y2="${row.y + 50}" stroke="#EEF1EE"/>` : ""}`).join("")}`;
}

function nav() {
  return `
  <g filter="url(#tabShadow)"><rect x="0" y="1118" width="418" height="122" fill="#FFFFFF"/></g>
  <line x1="0" y1="1118" x2="418" y2="1118" stroke="#E8EEE9"/>
  <rect x="321" y="1118" width="24" height="3" rx="1.5" fill="#159B55"/>
  <g fill="none" stroke="#8F9A94" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <path d="M74 1145l11-9 11 9v14H89v-8h-8v8h-7z"/>
    <path d="M198 1137c7-3 11-1 11 4v16c-3-4-7-5-11-3zM220 1137c-7-3-11-1-11 4v16c3-4 7-5 11-3z"/>
  </g>
  <g fill="none" stroke="#159B55" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <circle cx="333" cy="1143" r="7.5"/>
    <path d="M321 1159c2-7 7-10 12-10s10 3 12 10"/>
  </g>
  <text x="85" y="1184" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">首页</text>
  <text x="209" y="1184" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">菜谱</text>
  <text x="333" y="1184" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="800" fill="#159B55">我的</text>
  <rect x="147" y="1212" width="124" height="5" rx="2.5" fill="#202521" fill-opacity="0.74"/>`;
}

function integratedContent() {
  const serviceRows = [
    { y: 662, label: "历史饭局", meta: "12 场", icon: "history" },
    { y: 724, label: "采购清单", meta: "3 项待购买", icon: "cart" },
    { y: 786, label: "通知中心", meta: "2 条未读", icon: "bell" },
    { y: 848, label: "AI 使用记录", meta: "本月 8 次", icon: "spark" },
  ];
  const icons = {
    history: (y) => `<path d="M49 ${y + 25}a9 9 0 1 0 3-7M49 ${y + 17}v7h7M59 ${y + 23}v6l4 3"/>`,
    cart: (y) => `<path d="M49 ${y + 18}h4l3 13h11l3-9H55M58 ${y + 36}h.1M67 ${y + 36}h.1"/>`,
    bell: (y) => `<path d="M52 ${y + 30}h17l-3-4v-5a6 6 0 0 0-12 0v5zM58 ${y + 34}c1 3 4 3 5 0"/>`,
    spark: (y) => `<path d="M60 ${y + 17}l2 6 6 2-6 2-2 6-2-6-6-2 6-2zM69 ${y + 31}l1 3 3 1-3 1-1 3-1-3-3-1 3-1z"/>`,
  };
  return `
  <path d="M0 260A24 24 0 0 1 24 236H394A24 24 0 0 1 418 260V1064H0Z" fill="#FFFFFF"/>
  <text x="24" y="278" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="20" font-weight="850" fill="#1B241F">常用操作</text>
  <rect x="24" y="296" width="370" height="102" rx="15" fill="#F5F9F6"/>
  <line x1="147" y1="312" x2="147" y2="382" stroke="#E4ECE6"/><line x1="270" y1="312" x2="270" y2="382" stroke="#E4ECE6"/>
  <circle cx="86" cy="327" r="18" fill="#159B55"/><path d="M78 318h16v19H78zM82 314h8v5h-8M81 328l4 4 7-8" fill="none" stroke="#FFFFFF" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="209" cy="327" r="18" fill="#E5F6EC"/><path d="M201 321h15v15h-15zM205 317h7v5h-7M219 322h7m-3.5-3.5v7" fill="none" stroke="#159B55" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="332" cy="327" r="18" fill="#E5F6EC"/><path d="M324 319h16v18h-16zM328 324h8M328 329h8M343 321l4 4 6-7" fill="none" stroke="#159B55" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="86" y="363" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">做菜打卡</text><text x="86" y="383" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#159B55">记录一餐</text>
  <text x="209" y="363" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">创建饭局</text><text x="209" y="383" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">发起点菜</text>
  <text x="332" y="363" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">加入饭局</text><text x="332" y="383" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">输入点餐码</text>
  <text x="24" y="438" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="20" font-weight="850" fill="#1B241F">进行中的饭局</text><text x="370" y="437" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#159B55">查看全部</text><path d="M379 429l5 5-5 5" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <rect x="24" y="456" width="370" height="132" rx="15" fill="#F4F8F5" stroke="#E7EEE9"/>
  <text x="43" y="490" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">周六家宴</text>
  <rect x="313" y="469" width="62" height="28" rx="14" fill="#E3F5EA"/><circle cx="326" cy="483" r="4" fill="#20B866"/><text x="349" y="487" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">收集中</text>
  <text x="43" y="516" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">3 人已参与 · 6 道候选菜</text>
  <line x1="43" y1="534" x2="375" y2="534" stroke="#E1E9E3"/>
  <circle cx="58" cy="558" r="9" fill="#FFF2D9"/><path d="M58 553v6l4 2" fill="none" stroke="#C98A24" stroke-width="1.5" stroke-linecap="round"/><text x="76" y="563" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#526159">周五 20:00 前关闭点菜</text><path d="M357 553l5 5-5 5" fill="none" stroke="#8A958F" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="24" y="638" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="20" font-weight="850" fill="#1B241F">更多服务</text>
  ${serviceRows.map((row, index) => `<rect x="43" y="${row.y + 10}" width="36" height="36" rx="10" fill="#EFF8F2"/><g fill="none" stroke="#159B55" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">${icons[row.icon](row.y)}</g><text x="96" y="${row.y + 34}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="750" fill="#1B241F">${row.label}</text><text x="354" y="${row.y + 34}" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="${index === 1 ? "#C98A24" : index === 2 ? "#C84D49" : "#8A958F"}">${row.meta}</text><path d="M371 ${row.y + 26}l5 5-5 5" fill="none" stroke="#A6B0AA" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>${index < serviceRows.length - 1 ? `<line x1="96" y1="${row.y + 61}" x2="376" y2="${row.y + 61}" stroke="#EEF1EE"/>` : ""}`).join("")}`;
}

function conventionalContent(hasActiveMeal = true) {
  const rows = [
    { y: 590, label: "全部菜品", meta: "26 道", icon: "dish", metaColor: "#8A958F" },
    { y: 646, label: "历史饭局", meta: "12 场", icon: "history", metaColor: "#8A958F" },
    { y: 702, label: "采购清单", meta: "3 项待购买", icon: "cart", metaColor: "#C98A24" },
    { y: 758, label: "通知中心", meta: "2 条未读", icon: "bell", metaColor: "#C84D49" },
    { y: 814, label: "AI 使用记录", meta: "本月 8 次", icon: "spark", metaColor: "#8A958F" },
  ];
  const icons = {
    dish: (y) => `<rect x="53" y="${y + 15}" width="6" height="6" rx="1"/><rect x="63" y="${y + 15}" width="6" height="6" rx="1"/><rect x="53" y="${y + 25}" width="6" height="6" rx="1"/><rect x="63" y="${y + 25}" width="6" height="6" rx="1"/>`,
    history: (y) => `<path d="M51 ${y + 22}a8 8 0 1 0 3-6M51 ${y + 15}v6h6M60 ${y + 21}v5l4 3"/>`,
    cart: (y) => `<path d="M51 ${y + 16}h3l3 12h10l3-8H56M59 ${y + 33}h.1M67 ${y + 33}h.1"/>`,
    bell: (y) => `<path d="M53 ${y + 27}h15l-3-4v-4a5 5 0 0 0-10 0v4zM59 ${y + 31}c1 2 3 2 4 0"/>`,
    spark: (y) => `<path d="M61 ${y + 15}l2 5 5 2-5 2-2 5-2-5-5-2 5-2zM68 ${y + 28}l1 3 3 1-3 1-1 3-1-3-3-1 3-1z"/>`,
  };
  return `
  <rect y="224" width="418" height="894" fill="#FFFFFF"/>
  <rect y="224" width="418" height="8" fill="#F1F4F2"/>

  ${hasActiveMeal ? `
  <text x="24" y="280" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="850" fill="#1B241F">进行中的饭局</text><text x="370" y="279" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#159B55">查看全部</text><path d="M387 271l5 5-5 5" fill="none" stroke="#159B55" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="24" y="318" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="850" fill="#1B241F">周六家宴</text>
  <rect x="313" y="298" width="62" height="28" rx="14" fill="#E9F8EF"/><circle cx="326" cy="312" r="4" fill="#20B866"/><text x="349" y="316" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="750" fill="#159B55">收集中</text>
  <text x="24" y="343" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">3 人已参与 · 6 道候选菜</text>
  <circle cx="32" cy="369" r="8" fill="#FFF2D9"/><path d="M32 364v5l4 2" fill="none" stroke="#C98A24" stroke-width="1.4" stroke-linecap="round"/><text x="49" y="374" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#526159">周五 20:00 前关闭点菜</text><path d="M382 364l5 5-5 5" fill="none" stroke="#8A958F" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>` : `
  <text x="24" y="280" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="850" fill="#1B241F">进行中的饭局</text>
  <circle cx="45" cy="333" r="19" fill="#E9F8EF"/>
  <path d="M38 326h14v13H38zM41 323v5M49 323v5M38 330h14" fill="none" stroke="#159B55" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="76" y="327" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#1B241F">暂无进行中的饭局</text>
  <text x="76" y="351" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">创建饭局，邀请大家一起点菜</text>
  <text x="370" y="339" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="750" fill="#159B55">去创建</text><path d="M387 330l5 5-5 5" fill="none" stroke="#159B55" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>`}
  <line x1="0" y1="392" x2="418" y2="392" stroke="#E5EAE7"/>

  <text x="24" y="427" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="850" fill="#1B241F">常用操作</text>
  <line x1="147" y1="444" x2="147" y2="525" stroke="#EDF1EE"/><line x1="270" y1="444" x2="270" y2="525" stroke="#EDF1EE"/>
  <circle cx="86" cy="465" r="18" fill="#159B55"/><path d="M79 457h14v17H79zM82 454h8v4h-8M81 466l4 4 7-8" fill="none" stroke="#FFFFFF" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="209" cy="465" r="18" fill="#E7F6ED"/><path d="M202 460h14v14h-14zM205 456h7v5h-7M219 461h7m-3.5-3.5v7" fill="none" stroke="#159B55" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="332" cy="465" r="18" fill="#E7F6ED"/><path d="M325 458h14v16h-14zM328 463h8M328 468h8M342 460l4 4 6-7" fill="none" stroke="#159B55" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="86" y="503" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">做菜打卡</text><text x="86" y="522" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#159B55">记录一餐</text>
  <text x="209" y="503" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">创建饭局</text><text x="209" y="522" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">发起点菜</text>
  <text x="332" y="503" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#1B241F">加入饭局</text><text x="332" y="522" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="10" fill="#7B8982">输入点餐码</text>
  <line x1="0" y1="548" x2="418" y2="548" stroke="#E5EAE7"/>

  <text x="24" y="582" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="850" fill="#1B241F">更多操作</text>
  ${rows.map((row, index) => `<rect x="24" y="${row.y + 8}" width="34" height="34" rx="9" fill="#EFF8F2"/><g transform="translate(-19 0)" fill="none" stroke="#159B55" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">${icons[row.icon](row.y)}</g><text x="75" y="${row.y + 31}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="750" fill="#1B241F">${row.label}</text><text x="369" y="${row.y + 31}" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="${row.metaColor}">${row.meta}</text><path d="M386 ${row.y + 23}l5 5-5 5" fill="none" stroke="#A6B0AA" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>${index < rows.length - 1 ? `<line x1="0" y1="${row.y + 55}" x2="418" y2="${row.y + 55}" stroke="#EEF1EE"/>` : ""}`).join("")}`;
}

function profileEditorOverlay() {
  return `
  <rect x="0" y="0" width="418" height="1240" fill="#102018" fill-opacity="0.3"/>
  <path d="M0 714Q0 688 26 688H392Q418 688 418 714V1240H0Z" fill="#FFFFFF"/>
  <rect x="181" y="700" width="56" height="5" rx="2.5" fill="#DDE5E0"/>
  <text x="24" y="746" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="850" fill="#1B241F">完善个人资料</text>
  <text x="24" y="773" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">使用微信头像昵称快捷填写，也可以自行设置</text>
  <path d="M374 733l14 14M388 733l-14 14" fill="none" stroke="#78857E" stroke-width="1.8" stroke-linecap="round"/>
  <line x1="24" y1="798" x2="394" y2="798" stroke="#EEF1EE"/>

  <text x="24" y="851" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="750" fill="#1B241F">头像</text>
  <text x="271" y="851" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#159B55">选择微信头像</text>
  <circle cx="346" cy="840" r="30" fill="#E8F7EE" stroke="#D9E9DF" stroke-width="2"/>
  <text x="346" y="849" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#159B55">厨</text>
  <circle cx="367" cy="861" r="10" fill="#159B55" stroke="#FFFFFF" stroke-width="2"/><path d="M362 858h10v7h-10zM365 856h4l1 2" fill="none" stroke="#FFFFFF" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><circle cx="367" cy="861.5" r="2" fill="none" stroke="#FFFFFF" stroke-width="1.2"/>
  <path d="M386 835l5 5-5 5" fill="none" stroke="#9AA49E" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <line x1="24" y1="888" x2="394" y2="888" stroke="#EEF1EE"/>

  <text x="24" y="932" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="750" fill="#1B241F">昵称</text>
  <rect x="88" y="902" width="306" height="50" rx="8" fill="#F8FAF8" stroke="#DDE6E0"/>
  <text x="106" y="934" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#1B241F">厨房小记</text>
  <circle cx="373" cy="927" r="8" fill="#E5F6EB"/><path d="M369 927l3 3 5-6" fill="none" stroke="#159B55" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="24" y="980" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#7B8982">点击昵称输入框后，可从微信昵称快捷填写</text>

  <rect x="24" y="1010" width="370" height="48" rx="24" fill="url(#green)"/>
  <text x="209" y="1041" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#FFFFFF">保存</text>
  <text x="209" y="1094" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">暂不设置</text>`;
}

function pageSvg(hasActiveMeal = true, profileReady = true, showProfileEditor = false) {
  return `<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">${defs()}${header(profileReady)}${conventionalContent(hasActiveMeal)}${nav()}${showProfileEditor ? profileEditorOverlay() : ""}</svg>`;
}

async function build() {
  const hero = await sharp(heroSource).resize(418, 224, { fit: "fill" }).modulate({ brightness: 1.02, saturation: 0.9 }).png().toBuffer();
  const page = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(pageSvg()), left: 0, top: 0 }]).png().toBuffer();
  const emptyPage = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(pageSvg(false)), left: 0, top: 0 }]).png().toBuffer();
  const unsetPage = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(pageSvg(true, false)), left: 0, top: 0 }]).png().toBuffer();
  const editorPage = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#F8FAF7" } }).composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(pageSvg(true, true, true)), left: 0, top: 0 }]).png().toBuffer();
  fs.writeFileSync("miniApp/.preview/11-profile-preview.png", page);
  fs.writeFileSync("miniApp/.preview/11-profile-empty-preview.png", emptyPage);
  fs.writeFileSync("miniApp/.preview/11-profile-unset-preview.png", unsetPage);
  fs.writeFileSync("miniApp/.preview/11-profile-edit-preview.png", editorPage);
  console.log(JSON.stringify({ outputs: ["miniApp/.preview/11-profile-preview.png", "miniApp/.preview/11-profile-empty-preview.png", "miniApp/.preview/11-profile-unset-preview.png", "miniApp/.preview/11-profile-edit-preview.png"], width: WIDTH, height: HEIGHT }, null, 2));
}

build().catch((error) => { console.error(error); process.exit(1); });
