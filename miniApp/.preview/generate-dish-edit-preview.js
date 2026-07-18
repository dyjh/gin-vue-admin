const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1588;

function pageSvg() {
  return `
<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261a" flood-opacity="0.05"/>
    </filter>
    <linearGradient id="green" x1="0" x2="1">
      <stop offset="0" stop-color="#20B866"/>
      <stop offset="1" stop-color="#0A8E52"/>
    </linearGradient>
    <linearGradient id="coverBg" x1="0" x2="1">
      <stop offset="0" stop-color="#FFF5DC"/>
      <stop offset="1" stop-color="#EEF9F0"/>
    </linearGradient>
  </defs>

  <rect y="224" width="${WIDTH}" height="${HEIGHT - 224}" fill="#FEFEFA"/>

  <rect x="58" y="66" width="210" height="66" rx="18" fill="#FFF7E8" opacity="0.92"/>
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="800" fill="#151F1A">编辑菜品</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">调整资料，保存后同步到菜品库</text>

  <g filter="url(#shadow)">
    <rect x="24" y="238" width="370" height="254" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="273" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">菜品图片</text>
  <rect x="313" y="255" width="52" height="24" rx="12" fill="#FFF4E6"/>
  <text x="339" y="272" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#D97812">必填</text>
  <rect x="43" y="296" width="102" height="86" rx="16" fill="url(#coverBg)" stroke="#E6EEE6"/>
  <ellipse cx="94" cy="341" rx="33" ry="16" fill="#DCE8E0" stroke="#AABDB1"/>
  <ellipse cx="94" cy="337" rx="27" ry="12" fill="#F8DCA5"/>
  <circle cx="85" cy="335" r="6" fill="#F0A765"/>
  <circle cx="101" cy="336" r="6" fill="#F0A765"/>
  <path d="M79 329c10-9 23-9 32 0" fill="none" stroke="#FFFFFF" stroke-width="2" stroke-linecap="round"/>
  <path d="M108 323l11-8" stroke="#63B86E" stroke-width="2" stroke-linecap="round"/>
  <rect x="160" y="305" width="91" height="36" rx="18" fill="#FFFFFF" stroke="#159B55"/>
  <text x="206" y="329" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">更换图片</text>
  <rect x="263" y="305" width="101" height="36" rx="18" fill="#E9F8EF"/>
  <text x="314" y="329" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">AI重生成</text>
  <text x="160" y="364" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">当前图片可继续替换或重新生成</text>
  <line x1="43" y1="405" x2="373" y2="405" stroke="#EEF1EB"/>
  <text x="43" y="434" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#1B241F">状态</text>
  <rect x="94" y="416" width="58" height="28" rx="14" fill="#E9F8EF"/>
  <text x="123" y="435" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#159B55">草稿</text>
  <text x="43" y="466" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#7B8982">保存为可用后，可加入菜谱和饭局</text>

  <g filter="url(#shadow)">
    <rect x="24" y="512" width="370" height="340" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="548" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">基础信息</text>
  <line x1="43" y1="566" x2="373" y2="566" stroke="#EEF1EB"/>
  <text x="43" y="604" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="700" fill="#1B241F">菜名</text>
  <rect x="134" y="581" width="236" height="42" rx="18" fill="#FFFFFF" stroke="#E5ECE6"/>
  <text x="151" y="607" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" fill="#1B241F">香菇鸡腿饭</text>
  <line x1="43" y1="642" x2="373" y2="642" stroke="#EEF1EB"/>
  <text x="43" y="681" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="700" fill="#1B241F">分类</text>
  <rect x="134" y="660" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="167" y="680" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">主食</text>
  <rect x="213" y="660" width="78" height="30" rx="15" fill="#E9F8EF"/>
  <text x="252" y="680" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">家常菜</text>
  <path d="M357 668l7 7l-7 7" fill="none" stroke="#96A29B" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  <line x1="43" y1="708" x2="373" y2="708" stroke="#EEF1EB"/>
  <text x="43" y="747" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="700" fill="#1B241F">标签</text>
  <rect x="134" y="726" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="167" y="746" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">下饭</text>
  <rect x="213" y="726" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="246" y="746" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">快手</text>
  <rect x="292" y="726" width="67" height="30" rx="15" fill="#E9F8EF"/>
  <text x="326" y="746" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">少油</text>
  <line x1="43" y1="774" x2="373" y2="774" stroke="#EEF1EB"/>
  <text x="43" y="780" dy="36" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="700" fill="#1B241F">基础份量</text>
  <rect x="134" y="789" width="236" height="42" rx="18" fill="#FFFFFF" stroke="#E5ECE6"/>
  <text x="151" y="815" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" fill="#1B241F">1份适合 2 人</text>
  <path d="M350 804l7 7l-7 7" fill="none" stroke="#96A29B" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>

  <g filter="url(#shadow)">
    <rect x="24" y="876" width="370" height="270" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="912" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">配料列表</text>
  <rect x="296" y="893" width="70" height="26" rx="13" fill="#E9F8EF"/>
  <text x="331" y="911" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">按1份</text>
  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" fill="#3A4740">
    <text x="52" y="964">鸡腿肉</text>
    <text x="315" y="964" text-anchor="end">220g</text>
    <path d="M354 952l7 7l-7 7" fill="none" stroke="#96A29B" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    <line x1="43" y1="985" x2="373" y2="985" stroke="#EEF1EB"/>
    <text x="52" y="1018">香菇</text>
    <text x="315" y="1018" text-anchor="end">120g</text>
    <path d="M354 1006l7 7l-7 7" fill="none" stroke="#96A29B" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    <line x1="43" y1="1039" x2="373" y2="1039" stroke="#EEF1EB"/>
    <text x="52" y="1072">胡萝卜</text>
    <text x="315" y="1072" text-anchor="end">半根</text>
    <path d="M354 1060l7 7l-7 7" fill="none" stroke="#96A29B" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  </g>
  <rect x="145" y="1094" width="128" height="36" rx="18" fill="#F1FAF4"/>
  <text x="209" y="1117" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="#159B55">＋ 添加配料</text>

  <g filter="url(#shadow)">
    <rect x="24" y="1170" width="370" height="300" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="1206" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">做法步骤</text>
  <text x="43" y="1232" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">可只写文字，步骤图可选</text>
  <rect x="294" y="1191" width="72" height="26" rx="13" fill="#E9F8EF"/>
  <text x="330" y="1209" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">图片可选</text>
  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#3A4740">
    <circle cx="52" cy="1265" r="10" fill="#1F9D4E"/>
    <text x="52" y="1270" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">1</text>
    <text x="78" y="1270">鸡腿切块，香菇切片</text>
    <line x1="43" y1="1294" x2="373" y2="1294" stroke="#EEF1EB"/>
    <circle cx="52" cy="1320" r="10" fill="#1F9D4E"/>
    <text x="52" y="1325" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">2</text>
    <text x="78" y="1325">少油煎香，加入米饭焖熟</text>
    <line x1="43" y1="1349" x2="373" y2="1349" stroke="#EEF1EB"/>
    <circle cx="52" cy="1375" r="10" fill="#1F9D4E"/>
    <text x="52" y="1380" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">3</text>
    <text x="78" y="1380">出锅前撒少许葱花</text>
  </g>
  <rect x="146" y="1408" width="126" height="34" rx="17" fill="#F1FAF4"/>
  <text x="209" y="1430" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">＋ 添加步骤</text>

  <rect x="0" y="1490" width="${WIDTH}" height="98" fill="#FEFEFA"/>
  <rect x="26" y="1510" width="132" height="50" rx="25" fill="#FFFFFF" stroke="#9FCFB2"/>
  <text x="92" y="1541" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#159B55">保存草稿</text>
  <rect x="174" y="1510" width="218" height="50" rx="25" fill="url(#green)"/>
  <text x="283" y="1541" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#FFFFFF">保存为可用菜品</text>
  <text x="209" y="1575" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">可用后才能加入菜谱和饭局</text>
</svg>`;
}

async function main() {
  const source = "miniApp/.preview/references/baseline-subpages.png";
  const output = "miniApp/.preview/03-dish_edit-preview.png";
  const top = await sharp(source)
    .extract({ left: 12, top: 10, width: 394, height: 214 })
    .resize({ width: WIDTH, height: 224, fit: "fill" })
    .png()
    .toBuffer();

  const buffer = await sharp({
    create: {
      width: WIDTH,
      height: HEIGHT,
      channels: 4,
      background: "#FEFEFA",
    },
  })
    .composite([
      { input: top, left: 0, top: 0 },
      { input: Buffer.from(pageSvg()), left: 0, top: 0 },
    ])
    .png()
    .toBuffer();

  await sharp(buffer).png().toFile(output + ".tmp");
  fs.renameSync(output + ".tmp", output);
  const meta = await sharp(output).metadata();
  console.log(JSON.stringify({ output, width: meta.width, height: meta.height }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
