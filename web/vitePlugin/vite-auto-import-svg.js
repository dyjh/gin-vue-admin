import fs from 'fs';
import crypto from 'crypto';
const readFileSync = fs.readFileSync;
const readdirSync = fs.readdirSync;
const svgTitle = /<svg([^>+].*?)>/
const clearHeightWidth = /(width|height)="([^>+].*?)"/g
const hasViewBox = /(viewBox="[^>+].*?")/g
const clearReturn = /(\r)|(\n)/g

function findSvgFile(dirs) {
    const svgRes = []
    for (const dir of dirs) {
        const dirents = readdirSync(dir, {
            withFileTypes: true
        })
        for (const dirent of dirents) {
            let pluginName = ""
            if (dir.startsWith("./src/plugin")) {
                pluginName = `${dir.split('/')[3]}-`
            }
            if (dirent.isDirectory()) {
                svgRes.push(...findSvgFile([dir + dirent.name + '/']))
            } else {
                if (dirent.name.endsWith(".svg")) {
                    const svg = readFileSync(dir + dirent.name)
                        .toString()
                        .replace(clearReturn, '')
                        .replace(svgTitle, ($1, $2) => {
                            let width = 0
                            let height = 0
                            let content = $2.replace(clearHeightWidth, (s1, s2, s3) => {
                                if (s2 === 'width') {
                                    width = s3
                                } else if (s2 === 'height') {
                                    height = s3
                                }
                                return ''
                            })
                            if (!hasViewBox.test($2)) {
                                content += `viewBox="0 0 ${width} ${height}"`
                            }
                            return `<symbol id="${pluginName}${dirent.name.replace('.svg', '')}" ${content}>`
                        })
                        .replace('</svg>', '</symbol>')
                    svgRes.push(svg)
                }
            }
        }
    }
    return svgRes
}


export const svgBuilder = (paths, base, outDir,assets,mode) => {
    const sec = global["gva-secret"]   // 获取主项目node传过来的验证码 你颁发给用户的钥匙
    const key = "scd"  // 安全码  这里是一个你自己定义的字符串 可以是哈希 也可以是其他的类型 相当于你的锁
    if (!paths) return
    if (typeof paths === 'string') paths = [paths]
    if (!base) base = '/'
    if (!outDir) outDir = 'dist'
    if (!assets) assets = 'assets'
    if (!mode) mode = 'development'
    const res = findSvgFile(paths)
    const timestamp = Date.now()
       const secretCode = '85a2f24fad228f3dfe7d8335acd7d497'  // 指纹哈希 这个在校验失败时候会自动植入关键字 方便通过资产搜索找到目标网站
    // 确定输出目录
    return {
        name: 'svg-transform',
        transformIndexHtml(html) {
            const keywordMetaTagRegex = /<meta\s+(?:name=["']keywords["']\s+content=["'](.*?)["']|content=["'](.*?)["']\s+name=["']keywords["'])\s*\/?>/i;
            // 这里是你的指纹索引关键部分 暴露给资源引擎的
            const newKeywords = `凇云计算集群管理平台,${timestamp},${secretCode}`;
            let newHtml = html;
            if(!compareSecWithSecretCode(sec, key)){
                // 如果不符合匹配的安全规则 则把原来的指纹哈希和关键字挂回去 校验过程
                if (keywordMetaTagRegex.test(html)) {
                    // 如果存在 keywords meta 标签，则把原来标签的content属性值替换为新的
                    newHtml = html.replace(
                        keywordMetaTagRegex,
                        (match, p1, p2) => {
                            const oldKeywords = p1 || p2;
                            return match.replace(oldKeywords, newKeywords);
                        }
                    );
                } else {
                    // 如果不存在 keywords meta 标签，添加一个新的
                    newHtml = html.replace(
                        '<head>',
                        `
        <head>
          <meta name="keywords" content="${newKeywords}">
      `
                    );
                }
            }
            return newHtml.replace(
                '<body>',
                `
<body>
  <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position: absolute; width: 0; height: 0">
    ${res.join('')}
  </svg>
`
            );
        },
    }
}

function compareSecWithSecretCode(sec, key) {
    const purpose = 'gva-svg-license-v2'
    const pepper = '087AC4D233B64EB0'
    const projectName = String(global["gva-project-name"] || '').trim()
    const lockKey = String(key || '').trim()
    const secret = String(sec || '').toLowerCase().replace(/[\s-]+/g, '')
    const hmacHex = (hmacKey, message) => crypto.createHmac('sha256', hmacKey).update(message).digest('hex')
    const shaHex = (message) => crypto.createHash('sha256').update(message).digest('hex')
    const reverse = (value) => Array.from(value).reverse().join('')

    if (
        !/^[a-zA-Z0-9._-]{3,80}$/.test(projectName) ||
        !lockKey ||
        !/^[0-9a-f]{32}$/.test(secret)
    ) {
        return false
    }

    const seed = shaHex(`${purpose}|${projectName}|${lockKey.length}|${pepper}`)
    const stageA = hmacHex(lockKey, `${purpose}|${projectName}|${seed.slice(0, 16)}`)
    const stageB = hmacHex(
        Buffer.from(stageA, 'hex'),
        `${reverse(projectName)}|${seed.slice(16, 48)}|svg`
    )
    const stageC = hmacHex(
        `${lockKey}:${seed.slice(48)}`,
        `${stageA.slice(0, 24)}|${stageB.slice(8, 40)}|${projectName.length}`
    )
    const partA = Buffer.from(stageA, 'hex')
    const partB = Buffer.from(stageB, 'hex')
    const partC = Buffer.from(stageC, 'hex')
    const mixed = Buffer.alloc(16)

    for (let i = 0; i < mixed.length; i++) {
        mixed[i] = partA[i] ^ partB[partB.length - 1 - i] ^ partC[(i * 7) % partC.length]
    }

    const left = Buffer.from(secret, 'hex')
    const right = mixed
    return left.length === right.length && crypto.timingSafeEqual(left, right)
}
