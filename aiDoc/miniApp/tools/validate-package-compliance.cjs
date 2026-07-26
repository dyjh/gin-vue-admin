const fs = require("fs");
const path = require("path");

const packageRoot = path.resolve(__dirname, "..", "..", "..", "miniApp");
const projectConfig = JSON.parse(fs.readFileSync(path.join(packageRoot, "project.config.json"), "utf8"));
const ignored = new Set((projectConfig.packOptions?.ignore || []).map((item) => item.value));
const sourceExtensions = new Set([".js", ".json", ".wxml", ".wxss", ".wxs", ".svg", ".txt", ".md"]);
const forbiddenTerms = [
  { name: "uppercase term", expression: /(?:^|[^A-Za-z])AI(?:[^A-Za-z]|$)|[a-z0-9_]AI(?:[A-Z0-9_]|$)/ },
  { name: "camel-case term", expression: /[a-z0-9_]Ai(?:[A-Z0-9_]|$)|[a-z0-9_]ai(?=[A-Z0-9_])/ },
  { name: "standalone lowercase term", expression: /(?:^|[^A-Za-z])ai(?:[^A-Za-z]|$)/i },
  { name: "Chinese technology term", expression: /人工智能|大模型|模型生成|智能生成|智能推荐/ },
  { name: "supplier or model term", expression: /OpenAI|DeepSeek|GPT|千问|通义|Qwen/i },
];
const violations = [];

// isIgnored 按微信项目配置判断文件是否明确排除在上传包之外。
function isIgnored(relativePath) {
  const normalized = relativePath.split(path.sep).join("/");
  return [...ignored].some((entry) => normalized === entry || normalized.startsWith(`${entry}/`));
}

// visit 收集实际可能进入上传包的文件，不扫描已显式忽略的开发资源。
function visit(directory, files = []) {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const fullPath = path.join(directory, entry.name);
    const relativePath = path.relative(packageRoot, fullPath);
    if (isIgnored(relativePath)) continue;
    if (entry.isDirectory()) visit(fullPath, files);
    else files.push(fullPath);
  }
  return files;
}

// inspectText 检查文本内容并报告首个命中的准确行号。
function inspectText(file, relativePath) {
  if (!sourceExtensions.has(path.extname(file).toLowerCase())) return;
  const lines = fs.readFileSync(file, "utf8").split(/\r?\n/);
  lines.forEach((line, index) => {
    for (const rule of forbiddenTerms) {
      if (rule.expression.test(line)) {
        violations.push(`${relativePath}:${index + 1} ${rule.name}`);
        break;
      }
    }
  });
}

for (const file of visit(packageRoot)) {
  const relativePath = path.relative(packageRoot, file).split(path.sep).join("/");
  for (const rule of forbiddenTerms) {
    if (rule.expression.test(relativePath)) {
      violations.push(`${relativePath} path ${rule.name}`);
      break;
    }
  }
  inspectText(file, relativePath);
}

if (violations.length) {
  console.error(`Mini-program upload package compliance violations:\n${violations.join("\n")}`);
  process.exit(1);
}

console.log(JSON.stringify({
  scannedFiles: visit(packageRoot).length,
  ignoredEntries: [...ignored].sort(),
  violations: 0,
}, null, 2));
