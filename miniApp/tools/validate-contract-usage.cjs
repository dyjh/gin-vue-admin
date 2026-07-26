const fs = require("fs");
const path = require("path");

const root = path.resolve(__dirname, "..");
const api = JSON.parse(fs.readFileSync(path.resolve(root, "..", "aiDoc", "miniApp", "api.json"), "utf8"));
const apiSource = fs.readFileSync(path.join(root, "services", "api.js"), "utf8");
const authSource = fs.readFileSync(path.join(root, "services", "auth.js"), "utf8");
const uploadSource = fs.readFileSync(path.join(root, "services", "upload.js"), "utf8");

function normalize(value) {
  return value
    .replace(/\$\{[^}]+\}/g, "{}")
    .replace(/\{[^}]+\}/g, "{}");
}

const contract = new Set(api.interfaces.map((item) => `${item.method} ${normalize(item.path)}`));
const used = [];
const pattern = /request\(\{\s*url:\s*([\`"])(.*?)\1(?:,\s*method:\s*"([A-Z]+)")?/gs;
let match;
while ((match = pattern.exec(apiSource))) {
  used.push(`${match[3] || "GET"} ${normalize(match[2])}`);
}

const specialImplementations = [
  {
    signature: "POST /auth/wx-login",
    source: authSource,
    path: "/auth/wx-login",
  },
  {
    signature: "POST /uploads/images",
    source: uploadSource,
    path: "/uploads/images",
  },
];
const missingSpecialImplementations = specialImplementations
  .filter((item) => !item.source.includes(item.path))
  .map((item) => item.signature);
const implemented = [...used, ...specialImplementations.map((item) => item.signature)];
const implementedSet = new Set(implemented);
const outsideContract = implemented.filter((signature) => !contract.has(signature));
const uncoveredContracts = [...contract].filter((signature) => !implementedSet.has(signature));
const duplicates = implemented.filter((signature, index) => implemented.indexOf(signature) !== index);

const failures = [];
if (outsideContract.length) {
  failures.push(`Client request paths missing from api.json:\n${[...new Set(outsideContract)].join("\n")}`);
}
if (uncoveredContracts.length) {
  failures.push(`api.json interfaces missing from the client:\n${uncoveredContracts.join("\n")}`);
}
if (duplicates.length) {
  failures.push(`Client implements duplicate request signatures:\n${[...new Set(duplicates)].join("\n")}`);
}
if (missingSpecialImplementations.length) {
  failures.push(`Special request implementations are missing:\n${missingSpecialImplementations.join("\n")}`);
}
if (failures.length) {
  console.error(failures.join("\n\n"));
  process.exit(1);
}

console.log(JSON.stringify({
  contractInterfaces: contract.size,
  clientRequests: used.length,
  specialImplementations: specialImplementations.length,
  coveredInterfaces: implementedSet.size,
}, null, 2));
