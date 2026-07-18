const fs = require("fs");
const sharp = require("sharp");

async function insertBlock(sourceBuffer, width, height, insertY, insertH, svg) {
  const top = await sharp(sourceBuffer)
    .extract({ left: 0, top: 0, width, height: insertY })
    .png()
    .toBuffer();
  const bottom = await sharp(sourceBuffer)
    .extract({ left: 0, top: insertY, width, height: height - insertY })
    .png()
    .toBuffer();

  const output = await sharp({
    create: {
      width,
      height: height + insertH,
      channels: 4,
      background: "#FEFEFA",
    },
  })
    .composite([
      { input: top, left: 0, top: 0 },
      { input: Buffer.from(svg), left: 0, top: insertY },
      { input: bottom, left: 0, top: insertY + insertH },
    ])
    .png()
    .toBuffer();

  return { buffer: output, height: height + insertH };
}

function tabOverlaySvg(width, active) {
  const entryActive = active === "entry";
  const aiActive = active === "ai";
  const underlineX = entryActive ? 82 : 263;
  return `
<svg width="${width}" height="68" viewBox="0 0 ${width} 68" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="shadow" x="-8%" y="-30%" width="116%" height="170%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261a" flood-opacity="0.05"/>
    </filter>
  </defs>
  <g filter="url(#shadow)">
    <rect x="24" y="0" width="370" height="56" rx="14" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="119" y="33" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="${entryActive ? "#159B55" : "#8A958F"}">菜品录入</text>
  <text x="300" y="33" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="${aiActive ? "#159B55" : "#8A958F"}">AI识别</text>
  <rect x="${underlineX}" y="52" width="74" height="3" rx="1.5" fill="#159B55"/>
</svg>`;
}

function coverIntroSvg(width, insertH, filled = false) {
  const coverAction = filled ? "更换封面" : "上传封面";
  const aiAction = filled ? "AI重生成" : "AI生成图";
  const coverHint = filled ? "已按识别内容生成封面，可继续调整" : "保存前需上传或 AI 生成封面";
  const introLine1 = filled ? "嫩滑清淡，适合孩子和早餐晚餐" : "可以写口味、亮点、适合场景";
  const introLine2 = filled ? "AI 已填充，可删改或留空" : "也可以先空着";
  const thumbContent = filled
    ? `
  <ellipse cx="89" cy="104" rx="30" ry="14" fill="#DCE8E0" stroke="#AABDB1"/>
  <ellipse cx="89" cy="100" rx="24" ry="10" fill="#F9DDA6"/>
  <circle cx="82" cy="97" r="5" fill="#F2B36D"/>
  <circle cx="94" cy="98" r="5" fill="#F2B36D"/>
  <path d="M75 92c9-9 20-9 28 0" fill="none" stroke="#FFFFFF" stroke-width="2" stroke-linecap="round"/>
  <path d="M99 88l9-7" stroke="#63B86E" stroke-width="2" stroke-linecap="round"/>`
    : `
  <path d="M75 91h22a6 6 0 0 1 6 6v11H69V97a6 6 0 0 1 6-6z" fill="none" stroke="#54B970" stroke-width="2.2" stroke-linejoin="round"/>
  <circle cx="86" cy="85" r="4" fill="none" stroke="#54B970" stroke-width="2"/>
  <path d="M74 109l8-9l6 6l5-5l9 8" fill="none" stroke="#54B970" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`;
  return `
<svg width="${width}" height="${insertH}" viewBox="0 0 ${width} ${insertH}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261a" flood-opacity="0.05"/>
    </filter>
    <linearGradient id="thumbBg" x1="0" x2="1">
      <stop offset="0" stop-color="#FFF7E7"/>
      <stop offset="1" stop-color="#F1FAF1"/>
    </linearGradient>
  </defs>
  <rect width="${width}" height="${insertH}" fill="#FEFEFA"/>
  <g filter="url(#shadow)">
    <rect x="24" y="12" width="370" height="258" rx="14" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="44" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="700" fill="#1B241F">菜品封面</text>
  <rect x="315" y="26" width="50" height="24" rx="12" fill="#FFF4E6"/>
  <text x="340" y="43" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#D97812">必填</text>

  <rect x="43" y="62" width="92" height="74" rx="14" fill="url(#thumbBg)" stroke="#E6EEE6"/>
  ${thumbContent}

  <rect x="153" y="65" width="93" height="34" rx="17" fill="#FFFFFF" stroke="#159B55"/>
  <text x="199" y="87" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">${coverAction}</text>

  <rect x="258" y="65" width="106" height="34" rx="17" fill="#E9F8EF"/>
  <text x="311" y="87" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">${aiAction}</text>

  <text x="153" y="121" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">${coverHint}</text>

  <line x1="43" y1="146" x2="373" y2="146" stroke="#EEF1EB"/>
  <text x="43" y="174" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="700" fill="#1B241F">简介</text>
  <rect x="84" y="156" width="44" height="22" rx="11" fill="#F1F7F2"/>
  <text x="106" y="171" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="11" font-weight="700" fill="#7E8C84">可选</text>
  <rect x="43" y="190" width="330" height="58" rx="14" fill="#F8FBF8" stroke="#E8EEE9"/>
  <text x="60" y="214" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="${filled ? "#3A4740" : "#9BA5A0"}">${introLine1}</text>
  <text x="60" y="236" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#B1BBB5">${introLine2}</text>
</svg>`;
}

function aiFillSvg(width, insertH) {
  return `
<svg width="${width}" height="${insertH}" viewBox="0 0 ${width} ${insertH}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261a" flood-opacity="0.05"/>
    </filter>
    <linearGradient id="green" x1="0" x2="1">
      <stop offset="0" stop-color="#20B866"/>
      <stop offset="1" stop-color="#0A8E52"/>
    </linearGradient>
  </defs>
  <rect width="${width}" height="${insertH}" fill="#FEFEFA"/>
  <g filter="url(#shadow)">
    <rect x="24" y="12" width="370" height="534" rx="14" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="46" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="20" font-weight="700" fill="#1B241F">AI识别菜品</text>
  <rect x="300" y="28" width="66" height="24" rx="12" fill="#E9F8EF"/>
  <text x="333" y="45" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">识别完成</text>
  <text x="43" y="74" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">粘贴文字或上传截图，整理成可编辑菜品资料</text>

  <rect x="43" y="94" width="151" height="76" rx="18" fill="#F7FBF8" stroke="#DCEBE1"/>
  <path d="M103 111h31a5 5 0 0 1 5 5v25H98v-25a5 5 0 0 1 5-5z" fill="none" stroke="#159B55" stroke-width="2" stroke-linejoin="round"/>
  <path d="M106 124h25M106 132h18" stroke="#159B55" stroke-width="2" stroke-linecap="round"/>
  <text x="119" y="158" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">粘贴菜谱</text>

  <rect x="210" y="94" width="156" height="76" rx="18" fill="#F7FBF8" stroke="#DCEBE1"/>
  <rect x="275" y="112" width="32" height="25" rx="5" fill="none" stroke="#159B55" stroke-width="2"/>
  <circle cx="291" cy="123" r="4" fill="none" stroke="#159B55" stroke-width="2"/>
  <path d="M280 137l7-7l5 4l5-5l7 8" fill="none" stroke="#159B55" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="288" y="158" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">上传截图</text>

  <rect x="43" y="190" width="330" height="246" rx="18" fill="#F8FBF8" stroke="#E8EEE9"/>
  <text x="61" y="221" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="700" fill="#1B241F">识别结果</text>
  <rect x="287" y="203" width="64" height="24" rx="12" fill="#FFFFFF"/>
  <text x="319" y="220" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">5项内容</text>

  <line x1="61" y1="240" x2="353" y2="240" stroke="#E3ECE6"/>
  <text x="61" y="270" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#728178">菜名</text>
  <text x="122" y="270" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#1B241F">虾仁蒸蛋</text>
  <line x1="61" y1="288" x2="353" y2="288" stroke="#E3ECE6"/>

  <text x="61" y="318" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#728178">封面</text>
  <text x="122" y="318" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#1B241F">建议生成清淡蒸蛋图</text>
  <line x1="61" y1="336" x2="353" y2="336" stroke="#E3ECE6"/>

  <text x="61" y="366" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#728178">配料</text>
  <text x="122" y="366" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#1B241F">鸡蛋、虾仁、温水</text>
  <line x1="61" y1="384" x2="353" y2="384" stroke="#E3ECE6"/>

  <text x="61" y="414" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#728178">步骤</text>
  <text x="122" y="414" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#1B241F">已拆成 3 步</text>

  <rect x="43" y="458" width="116" height="44" rx="22" fill="#FFFFFF" stroke="#9FCFB2"/>
  <text x="101" y="486" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="#159B55">重新识别</text>
  <rect x="177" y="458" width="196" height="44" rx="22" fill="url(#green)"/>
  <text x="275" y="486" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="700" fill="#FFFFFF">智能填充</text>
  <text x="208" y="528" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">填充后回到菜品录入继续修改</text>
</svg>`;
}

function stepsSvg(width, insertH) {
  return `
<svg width="${width}" height="${insertH}" viewBox="0 0 ${width} ${insertH}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261a" flood-opacity="0.05"/>
    </filter>
  </defs>
  <rect width="${width}" height="${insertH}" fill="#FEFEFA"/>
  <g filter="url(#shadow)">
    <rect x="24" y="12" width="370" height="216" rx="14" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="44" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="700" fill="#1B241F">做法步骤</text>
  <text x="43" y="70" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">可只写文字，步骤图可选</text>
  <rect x="294" y="30" width="72" height="26" rx="13" fill="#E9F8EF"/>
  <text x="330" y="48" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">图片可选</text>

  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#3A4740">
    <circle cx="52" cy="102" r="10" fill="#1F9D4E"/>
    <text x="52" y="107" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">1</text>
    <text x="78" y="107">鸡蛋打散，虾仁处理干净</text>
    <rect x="338" y="86" width="34" height="32" rx="9" fill="#F5FAF6" stroke="#DCEBE1"/>
    <path d="M347 105l5-6l4 4l3-3l5 5" fill="none" stroke="#54B970" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M348 94h11a5 5 0 0 1 5 5v9h-16z" fill="none" stroke="#54B970" stroke-width="1.8" stroke-linejoin="round"/>
    <line x1="43" y1="128" x2="373" y2="128" stroke="#EEF1EB"/>

    <circle cx="52" cy="153" r="10" fill="#1F9D4E"/>
    <text x="52" y="158" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">2</text>
    <text x="78" y="158">温水调匀，倒入碗中上锅蒸</text>
    <rect x="338" y="137" width="34" height="32" rx="9" fill="#F8FBF8" stroke="#E4EEE7"/>
    <path d="M347 156l5-6l4 4l3-3l5 5" fill="none" stroke="#A4B7AB" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M348 145h11a5 5 0 0 1 5 5v9h-16z" fill="none" stroke="#A4B7AB" stroke-width="1.8" stroke-linejoin="round"/>
    <line x1="43" y1="179" x2="373" y2="179" stroke="#EEF1EB"/>

    <circle cx="52" cy="204" r="10" fill="#1F9D4E"/>
    <text x="52" y="209" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">3</text>
    <text x="78" y="209">出锅前点少许葱花</text>
    <rect x="338" y="188" width="34" height="32" rx="9" fill="#F8FBF8" stroke="#E4EEE7"/>
    <path d="M347 207l5-6l4 4l3-3l5 5" fill="none" stroke="#A4B7AB" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
    <path d="M348 196h11a5 5 0 0 1 5 5v9h-16z" fill="none" stroke="#A4B7AB" stroke-width="1.8" stroke-linejoin="round"/>
  </g>
</svg>`;
}

async function main() {
  const source = "miniApp/.preview/references/baseline-subpages.png";
  const aiOutput = "miniApp/.preview/02-dish_add_entry-ai-preview.png";
  const entryOutput = "miniApp/.preview/02-dish_add_entry-entry-preview.png";
  const legacyOutput = "miniApp/.preview/02-dish_add_entry-preview.png";
  const width = 418;
  const baseHeight = 941;

  const fullBleedTop = await sharp(source)
    .extract({ left: 12, top: 10, width: 394, height: 214 })
    .resize({ width, height: 224, fit: "fill" })
    .png()
    .toBuffer();
  const baseBody = await sharp(source)
    .extract({ left: 0, top: 224, width, height: baseHeight - 224 })
    .png()
    .toBuffer();
  const baseBuffer = await sharp({
    create: { width, height: baseHeight, channels: 4, background: "#FEFEFA" },
  })
    .composite([
      { input: fullBleedTop, left: 0, top: 0 },
      { input: baseBody, left: 0, top: 224 },
    ])
    .png()
    .toBuffer();

  let aiBuffer = await sharp(baseBuffer)
    .composite([{ input: Buffer.from(tabOverlaySvg(width, "ai")), left: 0, top: 224 }])
    .png()
    .toBuffer();

  const aiY = 282;
  const aiH = 585;
  const aiTop = await sharp(aiBuffer)
    .extract({ left: 0, top: 0, width, height: aiY })
    .png()
    .toBuffer();
  aiBuffer = await sharp({
    create: {
      width,
      height: baseHeight,
      channels: 4,
      background: "#FEFEFA",
    },
  })
    .composite([
      { input: aiTop, left: 0, top: 0 },
      { input: Buffer.from(aiFillSvg(width, aiH)), left: 0, top: aiY },
    ])
    .png()
    .toBuffer();

  await sharp(aiBuffer).png().toFile(aiOutput + ".tmp");
  fs.renameSync(aiOutput + ".tmp", aiOutput);

  let entryHeight = baseHeight;
  let entryBuffer = await sharp(baseBuffer)
    .composite([{ input: Buffer.from(tabOverlaySvg(width, "entry")), left: 0, top: 224 }])
    .png()
    .toBuffer();

  const coverY = 282;
  const coverH = 284;
  ({ buffer: entryBuffer, height: entryHeight } = await insertBlock(
    entryBuffer,
    width,
    entryHeight,
    coverY,
    coverH,
    coverIntroSvg(width, coverH, true)
  ));

  const stepsY = 800 + coverH;
  const stepsH = 240;
  ({ buffer: entryBuffer, height: entryHeight } = await insertBlock(
    entryBuffer,
    width,
    entryHeight,
    stepsY,
    stepsH,
    stepsSvg(width, stepsH)
  ));

  await sharp(entryBuffer).png().toFile(entryOutput + ".tmp");
  fs.renameSync(entryOutput + ".tmp", entryOutput);

  await sharp(aiBuffer).png().toFile(legacyOutput + ".tmp");
  fs.renameSync(legacyOutput + ".tmp", legacyOutput);

  const aiMeta = await sharp(aiOutput).metadata();
  const entryMeta = await sharp(entryOutput).metadata();
  console.log(
    JSON.stringify(
      {
        outputs: [
          { output: aiOutput, width: aiMeta.width, height: aiMeta.height },
          { output: entryOutput, width: entryMeta.width, height: entryMeta.height },
        ],
      },
      null,
      2
    )
  );
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
