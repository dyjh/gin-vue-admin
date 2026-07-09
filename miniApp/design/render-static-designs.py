import json
import re
from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter, ImageFont


ROOT = Path(__file__).resolve().parents[1]
DESIGN_DIR = ROOT / "design"
SCREEN_FILE = ROOT / "common" / "screens.js"
CAT_FILE = ROOT / "assets" / "cat-mascot.png"

W, H, S = 390, 844, 2

COLORS = {
    "bg_top": "#fffaf2",
    "bg_bottom": "#f8fbf6",
    "card": "#ffffff",
    "ink": "#21362d",
    "muted": "#798881",
    "line": "#edf0e8",
    "green": "#18bf6b",
    "green_dark": "#079757",
    "mint": "#eaf8ef",
    "orange": "#ffad68",
    "orange_pale": "#fff3df",
    "yellow": "#ffd86a",
    "coral": "#ff8f7a",
    "blue": "#82cbd9",
    "blue_pale": "#e9f7fb",
    "gray": "#f2f4ef",
}

TONE = {
    "green": ("#eaf8ef", "#079757"),
    "blue": ("#e9f7fb", "#248aa2"),
    "orange": ("#fff3df", "#d36b22"),
    "coral": ("#fff0ec", "#d65e50"),
    "yellow": ("#fff8d9", "#b87900"),
    "gray": ("#f2f4ef", "#7d8a83"),
}

PAGE_ORDER = [
    "home",
    "dishAdd",
    "dishAiParse",
    "dishImage",
    "dishDetail",
    "dishSearch",
    "dishRecommendations",
    "recipeList",
    "recipeEdit",
    "recipeDetail",
    "profile",
    "points",
    "aiRecords",
    "checkin",
    "notifications",
    "parties",
    "partyCreate",
    "partyCode",
    "partyJoin",
    "partyOrder",
    "partySummary",
    "purchaseList",
    "aiPrepTips",
    "whatToEat",
    "disabled",
]


def camel_to_kebab(name):
    return re.sub(r"(?<!^)(?=[A-Z])", "-", name).lower()


def load_screens():
    text = SCREEN_FILE.read_text(encoding="utf-8")
    match = re.search(r"const screens = (\{[\s\S]*?\});\s*module\.exports", text)
    if not match:
        raise RuntimeError("Unable to parse miniApp/common/screens.js")
    return json.loads(match.group(1))


def pick_font(bold=False):
    candidates = [
        "C:/Windows/Fonts/msyhbd.ttc" if bold else "C:/Windows/Fonts/msyh.ttc",
        "C:/Windows/Fonts/simhei.ttf",
        "C:/Windows/Fonts/simsun.ttc",
    ]
    for item in candidates:
        if Path(item).exists():
            return item
    return None


FONT_REGULAR = pick_font(False)
FONT_BOLD = pick_font(True) or FONT_REGULAR
FONT_CACHE = {}


def font(size, bold=False):
    key = (size, bold)
    if key not in FONT_CACHE:
        path = FONT_BOLD if bold else FONT_REGULAR
        if path:
            FONT_CACHE[key] = ImageFont.truetype(path, int(size * S))
        else:
            FONT_CACHE[key] = ImageFont.load_default()
    return FONT_CACHE[key]


def c(value):
    if isinstance(value, tuple):
        return value
    value = value.lstrip("#")
    return tuple(int(value[i : i + 2], 16) for i in (0, 2, 4))


def sc(value):
    return int(round(value * S))


def rect(x, y, w, h):
    return [sc(x), sc(y), sc(x + w), sc(y + h)]


class Renderer:
    def __init__(self):
        self.image = Image.new("RGB", (sc(W), sc(H)), c(COLORS["bg_top"]))
        self.draw = ImageDraw.Draw(self.image)
        self.background()

    def background(self):
        top = c(COLORS["bg_top"])
        bottom = c(COLORS["bg_bottom"])
        px = self.image.load()
        for y in range(sc(H)):
            t = y / max(sc(H) - 1, 1)
            color = tuple(int(top[i] * (1 - t) + bottom[i] * t) for i in range(3))
            for x in range(sc(W)):
                px[x, y] = color
        overlay = Image.new("RGBA", self.image.size, (0, 0, 0, 0))
        od = ImageDraw.Draw(overlay)
        od.ellipse(rect(-44, 42, 170, 160), fill=(255, 216, 106, 42))
        od.ellipse(rect(294, -42, 150, 150), fill=(130, 203, 217, 34))
        od.ellipse(rect(270, 420, 110, 110), fill=(24, 191, 107, 16))
        self.image = Image.alpha_composite(self.image.convert("RGBA"), overlay).convert("RGB")
        self.draw = ImageDraw.Draw(self.image)

    def rounded(self, x, y, w, h, r=8, fill="#ffffff", outline=None, width=1, shadow=False):
        if shadow:
            layer = Image.new("RGBA", self.image.size, (0, 0, 0, 0))
            ld = ImageDraw.Draw(layer)
            ld.rounded_rectangle(rect(x, y + 3, w, h), radius=sc(r), fill=(39, 76, 55, 20))
            layer = layer.filter(ImageFilter.GaussianBlur(sc(8)))
            self.image = Image.alpha_composite(self.image.convert("RGBA"), layer).convert("RGB")
            self.draw = ImageDraw.Draw(self.image)
        self.draw.rounded_rectangle(
            rect(x, y, w, h),
            radius=sc(r),
            fill=c(fill),
            outline=c(outline) if outline else None,
            width=sc(width),
        )

    def text(self, x, y, value, size=14, color=None, bold=False, max_width=None, max_lines=None, line_gap=4):
        color = c(color or COLORS["ink"])
        f = font(size, bold)
        if not max_width:
            self.draw.text((sc(x), sc(y)), value, font=f, fill=color)
            box = self.draw.textbbox((sc(x), sc(y)), value, font=f)
            return (box[3] - box[1]) / S
        lines = self.wrap(value, f, max_width, max_lines)
        yy = y
        line_h = size * 1.35
        for line in lines:
            self.draw.text((sc(x), sc(yy)), line, font=f, fill=color)
            yy += line_h + line_gap
        return yy - y - line_gap

    def wrap(self, value, f, max_width, max_lines=None):
        max_px = sc(max_width)
        lines = []
        current = ""
        for ch in value:
            trial = current + ch
            if self.draw.textlength(trial, font=f) <= max_px or not current:
                current = trial
            else:
                lines.append(current)
                current = ch
                if max_lines and len(lines) == max_lines:
                    lines[-1] = lines[-1][:-1] + "…" if len(lines[-1]) > 1 else "…"
                    return lines
        if current:
            lines.append(current)
        if max_lines and len(lines) > max_lines:
            lines = lines[:max_lines]
            lines[-1] = lines[-1][:-1] + "…" if len(lines[-1]) > 1 else "…"
        return lines

    def line(self, points, fill="#21362d", width=2):
        self.draw.line([(sc(x), sc(y)) for x, y in points], fill=c(fill), width=sc(width), joint="curve")

    def status_nav(self, title, back=False):
        self.text(18, 7, "9:41", 12, COLORS["ink"], True)
        self.text(300, 8, "5G  ▮▮▮  84%", 10.5, "#5f6f66", True)
        y = 28
        if back:
            self.rounded(14, y + 6, 34, 34, 17, fill="#ffffff", outline="#eef1eb")
            self.text(26, y + 5, "‹", 28, COLORS["ink"], False)
        self.text_center(195, y + 13, title, 17, COLORS["ink"], True)
        self.rounded(322, y + 9, 54, 27, 14, fill="#ffffff", outline="#e8ede6")
        for i in range(3):
            self.draw.ellipse(rect(343 + i * 8, y + 20, 4, 4), fill=c("#8a958f"))

    def text_center(self, x, y, value, size=14, color=None, bold=False):
        f = font(size, bold)
        box = self.draw.textbbox((0, 0), value, font=f)
        self.draw.text((sc(x) - (box[2] - box[0]) // 2, sc(y)), value, font=f, fill=c(color or COLORS["ink"]))

    def badge(self, x, y, text, tone="green", min_w=52):
        bg, fg = TONE.get(tone or "green", TONE["green"])
        text_w = self.draw.textlength(text, font=font(12, True)) / S
        w = max(min_w, text_w + 22)
        self.rounded(x, y, w, 28, 14, fill=bg)
        self.text_center(x + w / 2, y + 6, text, 12, fg, True)
        return w

    def button(self, x, y, w, h, text, primary=True):
        if primary:
            self.rounded(x, y, w, h, h / 2, fill=COLORS["green"], shadow=True)
            self.text_center(x + w / 2, y + (h - 18) / 2 - 1, text, 15.5, "#ffffff", True)
        else:
            self.rounded(x, y, w, h, h / 2, fill="#ffffff", outline="#cfeddc", shadow=True)
            self.text_center(x + w / 2, y + (h - 18) / 2 - 1, text, 15, COLORS["green_dark"], True)

    def cat(self, x, y, scale=1.0):
        if CAT_FILE.exists():
            asset = Image.open(CAT_FILE).convert("RGBA")
            target_w = max(1, sc(150 * scale))
            ratio = target_w / asset.width
            target_h = max(1, int(asset.height * ratio))
            asset = asset.resize((target_w, target_h), Image.Resampling.LANCZOS)

            shadow = Image.new("RGBA", self.image.size, (0, 0, 0, 0))
            sd = ImageDraw.Draw(shadow)
            sd.ellipse(
                [sc(x + 22 * scale), sc(y + 122 * scale), sc(x + 132 * scale), sc(y + 143 * scale)],
                fill=(120, 84, 42, 24),
            )
            shadow = shadow.filter(ImageFilter.GaussianBlur(sc(5 * scale)))
            base = Image.alpha_composite(self.image.convert("RGBA"), shadow)
            base.alpha_composite(asset, (sc(x), sc(y)))
            self.image = base.convert("RGB")
            self.draw = ImageDraw.Draw(self.image)
            return

        def sx(v):
            return x + v * scale

        def sy(v):
            return y + v * scale

        orange = "#f6b45f"
        orange_dark = "#d98a38"
        outline = "#b87332"
        cream = "#fff6e8"
        apron = "#7fa35d"
        self.draw.ellipse([sc(sx(22)), sc(sy(128)), sc(sx(138)), sc(sy(148))], fill=(120, 84, 42, 26))
        self.draw.arc([sc(sx(112)), sc(sy(78)), sc(sx(168)), sc(sy(152))], 250, 80, fill=c(outline), width=sc(7 * scale))
        self.draw.arc([sc(sx(116)), sc(sy(82)), sc(sx(158)), sc(sy(142))], 250, 80, fill=c(orange), width=sc(16 * scale))
        self.line([(sx(138), sy(103)), (sx(151), sy(112))], orange_dark, 4 * scale)
        self.line([(sx(139), sy(121)), (sx(153), sy(128))], orange_dark, 4 * scale)
        self.draw.polygon([(sc(sx(43)), sc(sy(47))), (sc(sx(30)), sc(sy(8))), (sc(sx(73)), sc(sy(35)))], fill=c(orange), outline=c(outline))
        self.draw.polygon([(sc(sx(104)), sc(sy(35))), (sc(sx(145)), sc(sy(9))), (sc(sx(131)), sc(sy(50)))], fill=c(orange), outline=c(outline))
        self.draw.ellipse([sc(sx(38)), sc(sy(32)), sc(sx(132)), sc(sy(116))], fill=c(orange), outline=c(outline), width=sc(3 * scale))
        self.draw.ellipse([sc(sx(54)), sc(sy(68)), sc(sx(118)), sc(sy(118))], fill=c(cream))
        for dx, rot in [(68, -18), (80, 0), (92, 18)]:
            self.line([(sx(dx), sy(42)), (sx(dx - rot * 0.12), sy(62))], orange_dark, 4 * scale)
        self.line([(sx(42), sy(64)), (sx(24), sy(58))], orange_dark, 3 * scale)
        self.line([(sx(43), sy(78)), (sx(24), sy(78))], orange_dark, 3 * scale)
        self.line([(sx(126), sy(64)), (sx(144), sy(58))], orange_dark, 3 * scale)
        self.line([(sx(125), sy(78)), (sx(144), sy(78))], orange_dark, 3 * scale)
        self.draw.ellipse([sc(sx(62)), sc(sy(70)), sc(sx(73)), sc(sy(84))], fill=c("#3b2d22"))
        self.draw.ellipse([sc(sx(99)), sc(sy(70)), sc(sx(110)), sc(sy(84))], fill=c("#3b2d22"))
        self.draw.ellipse([sc(sx(65)), sc(sy(72)), sc(sx(68)), sc(sy(75))], fill=c("#ffffff"))
        self.draw.ellipse([sc(sx(102)), sc(sy(72)), sc(sx(105)), sc(sy(75))], fill=c("#ffffff"))
        self.draw.ellipse([sc(sx(54)), sc(sy(88)), sc(sx(69)), sc(sy(96))], fill=(238, 139, 120, 70))
        self.draw.ellipse([sc(sx(103)), sc(sy(88)), sc(sx(118)), sc(sy(96))], fill=(238, 139, 120, 70))
        self.draw.polygon([(sc(sx(82)), sc(sy(86))), (sc(sx(88)), sc(sy(92))), (sc(sx(94)), sc(sy(86)))], fill=c("#8d5b35"))
        self.draw.arc([sc(sx(72)), sc(sy(88)), sc(sx(88)), sc(sy(104))], 8, 92, fill=c("#8d5b35"), width=sc(2 * scale))
        self.draw.arc([sc(sx(88)), sc(sy(88)), sc(sx(104)), sc(sy(104))], 88, 172, fill=c("#8d5b35"), width=sc(2 * scale))
        self.line([(sx(38), sy(87)), (sx(6), sy(82))], "#8d5b35", 1.8 * scale)
        self.line([(sx(38), sy(98)), (sx(8), sy(102))], "#8d5b35", 1.8 * scale)
        self.line([(sx(132), sy(87)), (sx(160), sy(82))], "#8d5b35", 1.8 * scale)
        self.line([(sx(132), sy(98)), (sx(158), sy(102))], "#8d5b35", 1.8 * scale)
        self.draw.rounded_rectangle([sc(sx(58)), sc(sy(112)), sc(sx(126)), sc(sy(150))], radius=sc(18 * scale), fill=c("#fff7ed"), outline=c(outline), width=sc(3 * scale))
        self.draw.rounded_rectangle([sc(sx(72)), sc(sy(116)), sc(sx(113)), sc(sy(154))], radius=sc(8 * scale), fill=c(apron), outline=c("#5d7d43"), width=sc(2 * scale))
        self.line([(sx(78), sy(117)), (sx(66), sy(100))], "#5d7d43", 3 * scale)
        self.line([(sx(108), sy(117)), (sx(122), sy(100))], "#5d7d43", 3 * scale)
        self.draw.ellipse([sc(sx(64)), sc(sy(104)), sc(sx(96)), sc(sy(132))], fill=c("#fff7ed"), outline=c(outline), width=sc(3 * scale))
        self.line([(sx(75), sy(116)), (sx(85), sy(124))], "#d7a371", 1.8 * scale)
        self.draw.ellipse([sc(sx(112)), sc(sy(138)), sc(sx(145)), sc(sy(164))], fill=c("#fff7ed"), outline=c(outline), width=sc(3 * scale))

    def hero(self, hero):
        self.rounded(0, 0, W, 220, 0, fill="#fff5d6")
        overlay = Image.new("RGBA", self.image.size, (0, 0, 0, 0))
        od = ImageDraw.Draw(overlay)
        od.rectangle(rect(0, 0, W, 220), fill=(227, 248, 235, 180))
        od.ellipse(rect(20, 74, 8, 8), fill=(255, 216, 106, 190))
        od.ellipse(rect(334, 62, 8, 8), fill=(255, 143, 122, 160))
        od.ellipse(rect(300, 136, 8, 8), fill=(130, 203, 217, 170))
        self.image = Image.alpha_composite(self.image.convert("RGBA"), overlay).convert("RGB")
        self.draw = ImageDraw.Draw(self.image)
        self.text(18, 122, hero["title"], 28, COLORS["ink"], True, max_width=228, max_lines=1)
        self.text(20, 160, hero["desc"], 13, "#5f6f66", False, max_width=220, max_lines=2)
        self.cat(238, 72, 0.95)
        return 232

    def title_block(self, data, y):
        self.text(16, y, data["title"], 26, COLORS["ink"], True, max_width=250, max_lines=2)
        self.text(16, y + 42, data["desc"], 13, COLORS["muted"], False, max_width=245, max_lines=3)
        self.cat(266, y - 16, 0.78)
        return y + 96

    def detail_hero(self, data, y):
        self.rounded(16, y, 358, 172, 8, fill="#fff0dc", shadow=True)
        self.cat(246, y + 22, 0.62)
        self.text(34, y + 104, data["title"], 28, COLORS["ink"], True)
        self.text(34, y + 140, data["desc"], 13, "#68786f")
        return y + 188

    def profile_card(self, y):
        self.rounded(16, y, 358, 104, 8, fill="#ffffff", shadow=True)
        self.rounded(30, y + 14, 78, 76, 8, fill="#fff3dd")
        self.cat(24, y + 16, 0.5)
        self.text(122, y + 22, "微信用户", 22, COLORS["ink"], True)
        self.text(122, y + 56, "积分 128 · 已打卡 7 天", 13, COLORS["muted"])
        self.badge(304, y + 38, "可推荐")
        return y + 118

    def search(self, text, y):
        self.rounded(16, y, 358, 44, 22, fill="#ffffff", outline="#eef1eb", shadow=True)
        self.draw.ellipse(rect(32, y + 14, 15, 15), outline=c("#93a09a"), width=sc(2))
        self.line([(45, y + 27), (52, y + 34)], "#93a09a", 2)
        self.text(62, y + 12, text, 14, COLORS["muted"], max_width=290, max_lines=1)
        return y + 56

    def actions(self, actions, y):
        gap = 10
        w = (358 - gap) / 2
        for i, action in enumerate(actions):
            self.button(16 + i * (w + gap), y, w, 44, action["text"], action.get("variant") == "primary")
        return y + 58

    def chips(self, items, y):
        x = 16
        for item in items:
            text = item["text"] if isinstance(item, dict) else str(item)
            tone = item.get("tone", "green") if isinstance(item, dict) else "green"
            bg, fg = TONE.get(tone, TONE["green"])
            tw = self.draw.textlength(text, font=font(12, True)) / S
            w = tw + 24
            if x + w > 374:
                break
            self.rounded(x, y, w, 30, 15, fill=bg)
            self.text_center(x + w / 2, y + 7, text, 12, fg, True)
            x += w + 8
        return y + 42

    def section_head(self, title, more, y):
        if not title and not more:
            return y
        self.text(16, y + 4, title or "", 20 if title else 16, COLORS["ink"], True)
        if more:
            self.text(316, y + 8, more, 13, COLORS["green_dark"], True)
        return y + 40

    def thumb(self, x, y, mark, tone="green", image_path=None, w=48, h=48):
        bg, fg = TONE.get(tone or "green", TONE["green"])
        self.rounded(x, y, w, h, 8, fill="#fbf4e6", outline="#d8c9a8")
        if image_path:
            source = ROOT / image_path.lstrip("/")
            if source.exists():
                asset = Image.open(source).convert("RGBA")
                target = (sc(w), sc(h))
                asset_ratio = asset.width / asset.height
                target_ratio = target[0] / target[1]
                if asset_ratio > target_ratio:
                    new_h = target[1]
                    new_w = int(new_h * asset_ratio)
                else:
                    new_w = target[0]
                    new_h = int(new_w / asset_ratio)
                asset = asset.resize((new_w, new_h), Image.Resampling.LANCZOS)
                left = (new_w - target[0]) // 2
                top = (new_h - target[1]) // 2
                asset = asset.crop((left, top, left + target[0], top + target[1]))
                mask = Image.new("L", target, 0)
                md = ImageDraw.Draw(mask)
                md.rounded_rectangle([0, 0, target[0], target[1]], radius=sc(8), fill=255)
                base = self.image.convert("RGBA")
                base.paste(asset, (sc(x), sc(y)), mask)
                self.image = base.convert("RGB")
                self.draw = ImageDraw.Draw(self.image)
                return
        self.text_center(x + w / 2, y + (h - 24) / 2, mark, 17, fg, True)

    def rows(self, section, y, purchase=False):
        y = self.section_head(section.get("title"), section.get("more"), y)
        items = section.get("items", [])
        row_h = 64 if purchase else 72
        card_h = max(8, len(items) * row_h + 8)
        self.rounded(16, y, 358, card_h, 8, fill="#ffffff", shadow=True)
        yy = y + 4
        for index, item in enumerate(items):
            if yy > H - 84:
                break
            if index:
                self.line([(28, yy), (362, yy)], COLORS["line"], 1)
            if purchase:
                bg, fg = TONE.get(item.get("tone", "green"), TONE["green"])
                self.rounded(28, yy + 18, 22, 22, 11, fill=bg)
                self.text_center(39, yy + 21, item.get("mark", ""), 10, fg, True)
                tx = 60
                self.text(tx, yy + 13, item["title"], 15, COLORS["ink"], True, max_width=220, max_lines=1)
                self.text(tx, yy + 36, item["desc"], 12, COLORS["muted"], False, max_width=220, max_lines=1)
            else:
                self.thumb(28, yy + 12, item.get("mark", ""), item.get("tone", "green"), item.get("image"), 48, 48)
                tx = 88
                self.text(tx, yy + 14, item["title"], 17, COLORS["ink"], True, max_width=190, max_lines=1)
                self.text(tx, yy + 42, item["desc"], 12.5, COLORS["muted"], False, max_width=190, max_lines=1)
            self.badge(304, yy + (18 if purchase else 22), item.get("badge", ""), item.get("badgeTone", "green"))
            yy += row_h
        return y + card_h + 14

    def recommend(self, section, y):
        y = self.section_head(section.get("title"), section.get("more"), y)
        for item in section.get("items", []):
            if y > H - 82:
                break
            self.rounded(16, y, 358, 92, 8, fill="#ffffff", shadow=True)
            self.thumb(30, y + 18, item.get("mark", ""), item.get("tone", "green"), item.get("image"), 72, 54)
            self.text(94, y + 16, item["title"], 17, COLORS["ink"], True, max_width=178, max_lines=1)
            self.text(94, y + 42, item["desc"], 12.5, COLORS["muted"], False, max_width=188, max_lines=1)
            self.rounded(94, y + 64, 84, 30, 15, fill=COLORS["green"])
            self.text_center(136, y + 71, item.get("action", "加入菜品库"), 11.5, "#ffffff", True)
            y += 102
        return y

    def fields(self, section, y):
        y = self.section_head(section.get("title"), section.get("more"), y)
        fields = section.get("fields", [])
        heights = [94 if f.get("large") else 62 for f in fields]
        card_h = sum(heights) + max(len(fields) - 1, 0) * 8 + 28
        self.rounded(16, y, 358, card_h, 8, fill="#ffffff", shadow=True)
        yy = y + 14
        for fdata, h in zip(fields, heights):
            if yy > H - 80:
                break
            self.text(30, yy, fdata["label"], 12, COLORS["muted"], True)
            self.rounded(30, yy + 20, 330, h - 24, 8, fill="#f8fbf6", outline="#e7ede6")
            self.text(42, yy + 31, fdata["value"], 13.5, "#5f6f66" if fdata.get("large") else COLORS["ink"], True, max_width=300, max_lines=3 if fdata.get("large") else 1)
            yy += h + 8
        return y + card_h + 14

    def stats(self, section, y):
        items = section.get("items", [])
        gap = 10
        w = (358 - 2 * gap) / 3
        for i, item in enumerate(items[:3]):
            x = 16 + i * (w + gap)
            self.rounded(x, y, w, 70, 8, fill="#ffffff", shadow=True)
            self.text_center(x + w / 2, y + 14, item["value"], 22, COLORS["ink"], True)
            self.text_center(x + w / 2, y + 44, item["label"], 12, COLORS["muted"], True)
        return y + 86

    def tiles(self, section, y):
        y = self.section_head(section.get("title"), section.get("more"), y)
        items = section.get("items", [])
        gap = 10
        w = (358 - gap) / 2
        h = 112
        for idx, item in enumerate(items):
            row = idx // 2
            col = idx % 2
            x = 16 + col * (w + gap)
            yy = y + row * (h + gap)
            if yy > H - 88:
                break
            self.rounded(x, yy, w, h, 8, fill="#ffffff", shadow=True)
            self.text(x + 14, yy + 14, item["title"], 17, COLORS["ink"], True, max_width=w - 28, max_lines=1)
            self.text(x + 14, yy + 44, item["desc"], 12, COLORS["muted"], False, max_width=w - 28, max_lines=2)
            if item.get("chips"):
                self.rounded(x + 14, yy + 82, 58, 24, 12, fill=COLORS["mint"])
                self.text_center(x + 43, yy + 88, item["chips"][0], 10.5, COLORS["green_dark"], True)
        rows = (len(items) + 1) // 2
        return y + rows * h + max(rows - 1, 0) * gap + 14

    def calendar(self, section, y):
        self.rounded(16, y, 358, 182, 8, fill="#ffffff", shadow=True)
        x0, y0 = 28, y + 14
        gap = 8
        size = (334 - 6 * gap) / 7
        for idx, day in enumerate(section.get("days", [])):
            x = x0 + (idx % 7) * (size + gap)
            yy = y0 + (idx // 7) * (34 + gap)
            fill = COLORS["green"] if day.get("today") else COLORS["mint"] if day.get("done") else "#f8fbf6"
            txt = "#ffffff" if day.get("today") else COLORS["green_dark"] if day.get("done") else "#76867e"
            self.rounded(x, yy, size, 34, 8, fill=fill)
            self.text_center(x + size / 2, yy + 8, day["label"], 12, txt, True)
        return y + 196

    def code(self, section, y):
        self.rounded(16, y, 358, 112, 8, fill="#f6fcf8", outline="#bfead1", shadow=True)
        self.text_center(195, y + 22, section["code"], 34, COLORS["green_dark"], True)
        self.text_center(195, y + 74, section["desc"], 12, COLORS["muted"], True)
        return y + 128

    def notice(self, section, y):
        y = self.section_head(section.get("title"), None, y)
        self.rounded(16, y, 358, 78, 8, fill="#fff8e8", outline="#ffddb7")
        self.text(30, y + 14, section["text"], 12, "#815625", True, max_width=330, max_lines=3)
        return y + 92

    def image_mock(self, y):
        self.rounded(16, y, 358, 210, 8, fill="#fff2df", shadow=True)
        self.draw.ellipse(rect(116, y + 132, 160, 50), fill=c("#ffffff"), outline=c("#42534a"), width=sc(3))
        self.draw.ellipse(rect(156, y + 72, 76, 50), fill=c(COLORS["coral"]))
        self.draw.ellipse(rect(176, y + 48, 44, 40), fill=c(COLORS["yellow"]))
        self.rounded(132, y + 116, 120, 18, 9, fill=COLORS["green"])
        return y + 224

    def center_state(self, data, y):
        self.cat(116, 212, 1.0)
        self.text_center(195, 402, data["title"], 24, COLORS["ink"], True)
        self.text(54, 446, data["desc"], 14, COLORS["muted"], False, max_width=282, max_lines=3)
        self.rounded(38, 536, 314, 84, 8, fill="#fff8e8", outline="#ffddb7")
        self.text(54, 552, data["notice"], 12, "#815625", True, max_width=284, max_lines=3)

    def bottom_ai(self):
        y = 844 - 82 - 78
        self.rounded(16, y, 358, 78, 8, fill="#ffffff", outline="#d7f1e2", shadow=True)
        self.text(30, y + 14, "不知道吃什么？", 17, COLORS["ink"], True)
        self.text(30, y + 44, "按口味和记录帮你推荐", 12, COLORS["muted"])
        self.rounded(280, y + 24, 76, 30, 15, fill=COLORS["green"])
        self.text_center(318, y + 31, "试试推荐", 12, "#ffffff", True)

    def bottom_actions(self, actions):
        y = 782
        x = 16
        if len(actions) == 1:
            self.button(x, y, 358, 44, actions[0]["text"], actions[0].get("variant") == "primary")
            return
        gap = 10
        widths = [150, 198]
        for i, action in enumerate(actions[:2]):
            self.button(x, y, widths[i], 44, action["text"], action.get("variant") == "primary")
            x += widths[i] + gap

    def tabbar(self, active):
        y = 762
        self.rounded(0, y, W, 82, 0, fill="#ffffff", outline="#e8ede6")
        tabs = [("home", "首页", "⌂"), ("recipe", "菜谱", "食"), ("profile", "我的", "人")]
        for idx, (key, label, icon) in enumerate(tabs):
            x = 52 + idx * 112
            selected = key == active
            color = COLORS["green_dark"] if selected else "#87928d"
            self.rounded(x + 19, y + 10, 30, 30, 15, fill=COLORS["mint"] if selected else "#ffffff", outline="#dfe5df")
            self.text_center(x + 34, y + 16, icon, 14, color, True)
            self.text_center(x + 34, y + 47, label, 12, color, True)

    def toast(self, text):
        self.rounded(82, 734, 226, 36, 18, fill="#21362d")
        self.text_center(195, 743, text, 13, "#ffffff", True)

    def render(self, screen):
        y = 82
        if screen.get("hero"):
            y = self.hero(screen["hero"])
        self.status_nav(screen.get("navTitle", ""), screen.get("back", False))
        if screen.get("titleBlock"):
            y = self.title_block(screen["titleBlock"], y)
        if screen.get("detailHero"):
            y = self.detail_hero(screen["detailHero"], y)
        if screen.get("profileCard"):
            y = self.profile_card(y)
        if screen.get("search"):
            y = self.search(screen["search"], y)
        if screen.get("actions"):
            y = self.actions(screen["actions"], y)
        if screen.get("chips"):
            y = self.chips(screen["chips"], y)
        for section in screen.get("sections", []):
            kind = section.get("type")
            if kind == "rows":
                y = self.rows(section, y)
            elif kind == "recommend":
                y = self.recommend(section, y)
            elif kind == "fields":
                y = self.fields(section, y)
            elif kind == "stats":
                y = self.stats(section, y)
            elif kind == "tiles":
                y = self.tiles(section, y)
            elif kind == "purchaseRows":
                y = self.rows(section, y, purchase=True)
            elif kind == "calendar":
                y = self.calendar(section, y)
            elif kind == "code":
                y = self.code(section, y)
            elif kind == "notice":
                y = self.notice(section, y)
            elif kind == "imageMock":
                y = self.image_mock(y)
        if screen.get("centerState"):
            self.center_state(screen["centerState"], y)
        if screen.get("bottomAI"):
            self.bottom_ai()
        if screen.get("bottomActions"):
            self.bottom_actions(screen["bottomActions"])
        if screen.get("toast"):
            self.toast(screen["toast"])
        if screen.get("tab"):
            self.tabbar(screen["tab"])
        return self.image


def main():
    screens = load_screens()
    files = []
    for index, key in enumerate(PAGE_ORDER, 1):
        renderer = Renderer()
        image = renderer.render(screens[key])
        name = f"{index:02d}-{camel_to_kebab(key)}.png"
        out = DESIGN_DIR / name
        image.save(out)
        files.append(name)
    manifest = {
        "count": len(files),
        "size": [W * S, H * S],
        "source": "miniApp/common/screens.js",
        "files": files,
    }
    (DESIGN_DIR / "design-manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Generated {len(files)} PNG design images in {DESIGN_DIR}")


if __name__ == "__main__":
    main()
