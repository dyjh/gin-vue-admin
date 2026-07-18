const sharp = require("sharp");

async function main() {
  const input = "miniApp/.preview/references/baseline-home.png";
  const output = "miniApp/.preview/home-preview.png";

  const svg = `
<svg width="108" height="130" viewBox="0 0 108 130" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="circle" x1="24" y1="24" x2="86" y2="93" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="#78CF52"/>
      <stop offset="1" stop-color="#36A845"/>
    </linearGradient>
    <filter id="softShadow" x="-20%" y="-20%" width="140%" height="140%">
      <feDropShadow dx="0" dy="4" stdDeviation="4" flood-color="#16793D" flood-opacity="0.14"/>
    </filter>
  </defs>
  <ellipse cx="50" cy="66" rx="50" ry="60" fill="#FBFEFA"/>
  <circle cx="50" cy="58" r="31" fill="url(#circle)" filter="url(#softShadow)"/>
  <path d="M37 70L62 45" stroke="#FFFFFF" stroke-width="4.4" stroke-linecap="round"/>
  <path d="M59 42L66 49" stroke="#FFFFFF" stroke-width="4.4" stroke-linecap="round"/>
  <path d="M31 43L34 50L41 53L34 56L31 63L28 56L21 53L28 50Z" fill="#FFFFFF" opacity="0.95"/>
  <path d="M64 64L67 69L72 72L67 75L64 80L61 75L56 72L61 69Z" fill="#FFFFFF" opacity="0.95"/>
  <circle cx="47" cy="40" r="2.6" fill="#FFFFFF" opacity="0.88"/>
</svg>`;

  await sharp(input)
    .composite([
      {
        input: Buffer.from(svg),
        left: 55,
        top: 1578,
      },
    ])
    .png()
    .toFile(output);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
