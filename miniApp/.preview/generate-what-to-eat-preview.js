const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const INPUT_HEIGHT = 1320;
const RESULT_HEIGHT = 1360;

function defs() {
  return `
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261A" flood-opacity="0.05"/>
    </filter>
    <linearGradient id="green" x1="0" x2="1">
      <stop offset="0" stop-color="#20B866"/>
      <stop offset="1" stop-color="#0A8E52"/>
    </linearGradient>
  </defs>`;
}

function header(height) {
  return `
  <text x="44" y="39" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="#151F1A">9:41</text>
  <g fill="#151F1A">
    <rect x="326" y="32" width="4" height="8" rx="1"/>
    <rect x="332" y="28" width="4" height="12" rx="1"/>
    <rect x="338" y="24" width="4" height="16" rx="1"/>
    <path d="M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0" fill="none" stroke="#151F1A" stroke-width="2" stroke-linecap="round"/>
    <rect x="373" y="28" width="21" height="11" rx="3" fill="none" stroke="#151F1A" stroke-width="1.6"/>
    <rect x="395" y="31" width="2" height="5" rx="1"/>
    <rect x="376" y="31" width="15" height="5" rx="1"/>
  </g>
  <g>
    <rect x="312" y="60" width="82" height="36" rx="18" fill="#FFFFFF" fill-opacity="0.94" stroke="#E7E9E5"/>
    <circle cx="338" cy="76" r="2.4" fill="#151F1A"/>
    <circle cx="348" cy="76" r="2.4" fill="#151F1A"/>
    <line x1="359" y1="66" x2="359" y2="86" stroke="#E7E9E5"/>
    <circle cx="378" cy="76" r="9" fill="none" stroke="#151F1A" stroke-width="2.6"/>
  </g>
  <path d="M48 79L38 89l10 10M39 89h20" fill="none" stroke="#151F1A" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="23" font-weight="800" fill="#151F1A">不知道吃什么</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6F7D76">按口味和记录帮你推荐</text>
  <rect y="224" width="${WIDTH}" height="${height - 224}" fill="#FEFEFA"/>`;
}

function quotaCard(exhausted, locked) {
  if (locked) {
    return `
  <g filter="url(#shadow)"><rect x="24" y="238" width="370" height="112" rx="16" fill="#FFF7DC" stroke="#F3E5B9"/></g>
  <text x="43" y="274" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#5D4A28">再打卡 2 天即可解锁</text>
  <text x="43" y="304" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8C7955">累计打卡满 7 天后，推荐会更贴合你的口味</text>
  <rect x="43" y="320" width="278" height="8" rx="4" fill="#F2E4BC"/>
  <rect x="43" y="320" width="199" height="8" rx="4" fill="#D8A33C"/>
  <text x="365" y="328" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="800" fill="#8B6D31">5 / 7 天</text>`;
  }

  if (exhausted) {
    return `
  <g filter="url(#shadow)"><rect x="24" y="238" width="370" height="112" rx="16" fill="#FFF7DC" stroke="#F3E5B9"/></g>
  <circle cx="58" cy="286" r="20" fill="#FFE9A9"/>
  <path d="M49 286h18M58 277v18" stroke="#D6952D" stroke-width="2" stroke-linecap="round"/>
  <text x="91" y="278" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#5D4A28">今日免费次数已用完</text>
  <text x="91" y="304" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8C7955">可按当前积分配置继续推荐</text>
  <text x="91" y="328" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#9A8763">推荐失败不会扣除积分</text>`;
  }

  return `
  <g filter="url(#shadow)"><rect x="24" y="238" width="370" height="112" rx="16" fill="#FFF7DC" stroke="#F3E5B9"/></g>
  <text x="43" y="274" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#5D4A28">今日免费推荐</text>
  <text x="43" y="316" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="28" font-weight="850" fill="#D6952D">剩余 2 次</text>
  <rect x="289" y="257" width="76" height="30" rx="15" fill="#FFFFFF" fill-opacity="0.72"/>
  <text x="327" y="277" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#8B6D31">累计打卡 12 天</text>
  <text x="289" y="312" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#9A8763">仅成功推荐才计次</text>`;
}

function conditionContent() {
  return `
  <g filter="url(#shadow)"><rect x="24" y="374" width="370" height="526" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <text x="43" y="412" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">今天想怎么吃</text>
  <text x="43" y="438" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">条件可多选，也可以直接开始推荐</text>

  <text x="43" y="484" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#3A4740">用餐场景</text>
  <rect x="43" y="502" width="323" height="42" rx="14" fill="#F1F6F2"/>
  <rect x="151" y="505" width="104" height="36" rx="12" fill="#FFFFFF" stroke="#CFE4D6"/>
  <text x="97" y="529" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#708078">午餐</text>
  <text x="203" y="529" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#159B55">晚餐</text>
  <text x="311" y="529" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#708078">夜宵</text>

  <text x="43" y="590" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#3A4740">一起吃的人</text>
  <rect x="43" y="608" width="88" height="36" rx="18" fill="#F1F6F2"/>
  <text x="87" y="631" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#65746C">一人食</text>
  <rect x="141" y="608" width="88" height="36" rx="18" fill="#E9F8EF" stroke="#B7DEC6"/>
  <text x="185" y="631" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#159B55">两人餐</text>
  <rect x="239" y="608" width="88" height="36" rx="18" fill="#F1F6F2"/>
  <text x="283" y="631" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#65746C">多人餐</text>

  <text x="43" y="690" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#3A4740">今天的口味</text>
  <rect x="43" y="708" width="72" height="36" rx="18" fill="#E9F8EF" stroke="#B7DEC6"/>
  <text x="79" y="731" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#159B55">快手</text>
  <rect x="125" y="708" width="72" height="36" rx="18" fill="#E9F8EF" stroke="#B7DEC6"/>
  <text x="161" y="731" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#159B55">下饭</text>
  <rect x="207" y="708" width="72" height="36" rx="18" fill="#F1F6F2"/>
  <text x="243" y="731" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#65746C">清淡</text>
  <rect x="289" y="708" width="72" height="36" rx="18" fill="#F1F6F2"/>
  <text x="325" y="731" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#65746C">汤菜</text>
  <rect x="43" y="756" width="72" height="36" rx="18" fill="#F1F6F2"/>
  <text x="79" y="779" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#65746C">少油</text>
  <rect x="125" y="756" width="72" height="36" rx="18" fill="#F1F6F2"/>
  <text x="161" y="779" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#65746C">素菜</text>

  <line x1="43" y1="824" x2="373" y2="824" stroke="#EEF1EB"/>
  <text x="43" y="858" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#3A4740">忌口与偏好</text>
  <text x="349" y="858" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#159B55">少油 · 不太辣</text>
  <path d="M359 850l6 6-6 6" fill="none" stroke="#9AA49E" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>`;
}

function profileCard() {
  return `
  <g filter="url(#shadow)"><rect x="24" y="924" width="370" height="132" rx="16" fill="#EDF7F8" stroke="#D8EBEE"/></g>
  <circle cx="54" cy="961" r="17" fill="#D7EEF1"/>
  <path d="M48 964c6-10 12-10 14-12-1 8-5 14-14 14 2-2 4-4 7-6" fill="none" stroke="#4F9CB0" stroke-width="1.8" stroke-linecap="round"/>
  <text x="82" y="958" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#31525A">你的口味线索</text>
  <text x="82" y="982" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6E858A">常选快手菜，最近更偏好鸡肉和少油做法</text>
  <rect x="43" y="1002" width="74" height="28" rx="14" fill="#FFFFFF" fill-opacity="0.82"/>
  <text x="80" y="1021" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#4B737C">快手较多</text>
  <rect x="127" y="1002" width="74" height="28" rx="14" fill="#FFFFFF" fill-opacity="0.82"/>
  <text x="164" y="1021" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#4B737C">常选鸡肉</text>
  <rect x="211" y="1002" width="62" height="28" rx="14" fill="#FFFFFF" fill-opacity="0.82"/>
  <text x="242" y="1021" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#4B737C">少油</text>`;
}

function inputSvg(pointsModal, locked = false) {
  const modal = pointsModal
    ? `
  <rect x="0" y="0" width="${WIDTH}" height="${INPUT_HEIGHT}" fill="#10261A" fill-opacity="0.24"/>
  <g filter="url(#shadow)"><rect x="24" y="850" width="370" height="390" rx="20" fill="#FFFFFF" stroke="#E8EEE9"/></g>
  <circle cx="209" cy="914" r="32" fill="#FFF1E4"/>
  <circle cx="209" cy="914" r="17" fill="none" stroke="#D98942" stroke-width="2"/>
  <text x="209" y="920" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#D98942">积</text>
  <text x="209" y="978" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#1B241F">继续推荐吗？</text>
  <text x="209" y="1010" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#7B8982">今日 2 次免费推荐已经用完</text>
  <rect x="51" y="1034" width="316" height="74" rx="14" fill="#FFF8F1" stroke="#F3E3D3"/>
  <text x="72" y="1064" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#71543A">本次消耗</text>
  <text x="347" y="1064" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="850" fill="#D98942">2 积分</text>
  <text x="72" y="1090" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#9A7B5D">当前积分 18 · 推荐失败自动退还</text>
  <rect x="51" y="1132" width="132" height="48" rx="24" fill="#FFFFFF" stroke="#C9D8CE"/>
  <text x="117" y="1162" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#5F7067">取消</text>
  <rect x="199" y="1132" width="168" height="48" rx="24" fill="url(#green)"/>
  <text x="283" y="1162" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#FFFFFF">消耗 2 积分继续</text>
  <text x="209" y="1214" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#97A29C">积分消耗由后台配置</text>`
    : "";

  const body = locked
    ? `<g opacity="0.42">${conditionContent()}${profileCard()}</g>`
    : `${conditionContent()}${profileCard()}`;
  const action = locked
    ? `
  <rect x="26" y="1082" width="366" height="54" rx="27" fill="#DDE6E0"/>
  <text x="209" y="1116" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#8B9991">累计打卡 7 天后可用</text>`
    : `
  <rect x="26" y="1082" width="366" height="54" rx="27" fill="url(#green)"/>
  <text x="209" y="1116" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#FFFFFF">开始推荐</text>`;

  return `
<svg width="${WIDTH}" height="${INPUT_HEIGHT}" viewBox="0 0 ${WIDTH} ${INPUT_HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  ${defs()}
  ${header(INPUT_HEIGHT)}
  ${quotaCard(pointsModal, locked)}
  ${body}
  ${action}
  <text x="209" y="1164" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">${locked ? "继续打卡，解锁后会结合你的口味记录" : "会结合你的打卡、菜品标签和点菜记录"}</text>
  <text x="209" y="1284" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#A0AAA4">AI 推荐仅作参考</text>
  ${modal}
</svg>`;
}

function resultSvg() {
  return `
<svg width="${WIDTH}" height="${RESULT_HEIGHT}" viewBox="0 0 ${WIDTH} ${RESULT_HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  ${defs()}
  ${header(RESULT_HEIGHT)}

  <g filter="url(#shadow)"><rect x="24" y="238" width="370" height="86" rx="16" fill="#FFF7DC" stroke="#F3E5B9"/></g>
  <text x="43" y="270" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#5D4A28">今日免费推荐</text>
  <text x="43" y="300" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="20" font-weight="850" fill="#D6952D">剩余 1 次</text>
  <text x="368" y="287" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8B7751">晚餐 · 两人餐 · 快手下饭</text>

  <g filter="url(#shadow)"><rect x="24" y="348" width="370" height="530" rx="18" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <text x="43" y="386" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">为你推荐</text>
  <rect x="287" y="367" width="79" height="28" rx="14" fill="#E9F8EF"/>
  <text x="326" y="386" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">AI生成推荐</text>
  <text x="43" y="616" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="27" font-weight="850" fill="#151F1A">宫保鸡丁</text>
  <text x="43" y="647" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#7B8982">川菜 · 经典家常菜 · 适合 2 人</text>
  <rect x="43" y="666" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="76" y="686" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#159B55">下饭</text>
  <rect x="119" y="666" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="152" y="686" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#159B55">快手</text>
  <rect x="195" y="666" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="228" y="686" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#159B55">微辣</text>
  <line x1="43" y1="724" x2="373" y2="724" stroke="#EEF1EB"/>
  <text x="43" y="759" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#3A4740">适合今天的理由</text>
  <text x="43" y="789" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#536159">你选择了快手和下饭，宫保鸡丁是做法明确的经典川菜。</text>
  <text x="43" y="816" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#536159">两人份容易调整，辣度也可以按偏好降低。</text>
  <circle cx="51" cy="848" r="10" fill="#E9F8EF"/>
  <path d="M46 848l3 3 6-7" fill="none" stroke="#159B55" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="70" y="853" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#63736B">经典川菜 · 标准菜名</text>

  <g filter="url(#shadow)"><rect x="24" y="902" width="370" height="150" rx="16" fill="#EDF7F8" stroke="#D8EBEE"/></g>
  <text x="43" y="938" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#31525A">为什么推荐它</text>
  <circle cx="49" cy="974" r="4" fill="#4F9CB0"/>
  <text x="63" y="979" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#55737A">与你选择的“快手、下饭”条件接近</text>
  <circle cx="49" cy="1010" r="4" fill="#4F9CB0"/>
  <text x="63" y="1015" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#55737A">菜名、菜系、食材和步骤已通过基础校验</text>

  <rect x="26" y="1080" width="366" height="54" rx="27" fill="url(#green)"/>
  <text x="209" y="1114" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#FFFFFF">查看完整做法</text>
  <rect x="26" y="1150" width="238" height="50" rx="25" fill="#FFFFFF" stroke="#9FCFB2"/>
  <text x="145" y="1181" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#159B55">换一道</text>
  <text x="346" y="1181" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#7B8982">先不考虑</text>
  <text x="209" y="1238" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">查看不会保存，保存到菜品库后才算采用</text>
  <text x="209" y="1324" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" fill="#A0AAA4">推荐结果仅供参考，可随时调整</text>
</svg>`;
}

async function roundedPhoto(source, width, height, radius, border) {
  const mask = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`);
  const frame = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect x="0.5" y="0.5" width="${width - 1}" height="${height - 1}" rx="${radius}" fill="none" stroke="${border}"/></svg>`);
  return sharp(source)
    .resize({ width, height, fit: "cover", position: "center" })
    .composite([{ input: mask, blend: "dest-in" }, { input: frame }])
    .png()
    .toBuffer();
}

async function heroCard() {
  const source = sharp("miniApp/.preview/assets/what-to-eat-cat-fridge.png");
  const meta = await source.metadata();
  const cropWidth = Math.floor(meta.width * 0.87);
  const cropHeight = Math.floor(meta.height * 0.87);
  const left = Math.floor((meta.width - cropWidth) * 0.5);
  const top = Math.floor((meta.height - cropHeight) * 0.42);
  const artwork = await source
    .extract({ left, top, width: cropWidth, height: cropHeight })
    .resize({ width: 418, height: 224, fit: "cover", position: "center" })
    .png()
    .toBuffer();
  return artwork;
}

async function render(output, height, svg, hero, photoLayer) {
  const layers = [
    { input: hero, left: 0, top: 0 },
    { input: Buffer.from(svg), left: 0, top: 0 },
  ];
  if (photoLayer) layers.push(photoLayer);
  const buffer = await sharp({ create: { width: WIDTH, height, channels: 4, background: "#FEFEFA" } })
    .composite(layers)
    .png()
    .toBuffer();
  await sharp(buffer).png().toFile(output + ".tmp");
  fs.renameSync(output + ".tmp", output);
  return sharp(output).metadata();
}

async function main() {
  const hero = await heroCard();
  const resultPhoto = await roundedPhoto(
    "miniApp/.preview/assets/what-to-eat-kung-pao-chicken.png",
    332,
    170,
    18,
    "#E6EEE6",
  );

  const inputOutput = "miniApp/.preview/07-what_to_eat-input-preview.png";
  const resultOutput = "miniApp/.preview/07-what_to_eat-result-preview.png";
  const pointsOutput = "miniApp/.preview/07-what_to_eat-points-preview.png";
  const lockedOutput = "miniApp/.preview/07-what_to_eat-locked-preview.png";

  const inputMeta = await render(inputOutput, INPUT_HEIGHT, inputSvg(false), hero);
  const resultMeta = await render(resultOutput, RESULT_HEIGHT, resultSvg(), hero, { input: resultPhoto, left: 43, top: 414 });
  const pointsMeta = await render(pointsOutput, INPUT_HEIGHT, inputSvg(true), hero);
  const lockedMeta = await render(lockedOutput, INPUT_HEIGHT, inputSvg(false, true), hero);

  console.log(JSON.stringify({
    outputs: [inputOutput, resultOutput, pointsOutput, lockedOutput],
    input: { width: inputMeta.width, height: inputMeta.height },
    result: { width: resultMeta.width, height: resultMeta.height },
    points: { width: pointsMeta.width, height: pointsMeta.height },
    locked: { width: lockedMeta.width, height: lockedMeta.height },
  }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
