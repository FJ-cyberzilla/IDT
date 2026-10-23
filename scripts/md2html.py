#!/usr/bin/env python3
"""Convert docs/*.md -> docs/*.html styled with doc.css. No deps."""
import html
import re
import sys
from pathlib import Path

DOCS = Path(__file__).resolve().parent.parent / "docs"

TEMPLATE = """<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
<title>{title} — IDT</title>
<meta name="theme-color" content="#1A1B26">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet">
<link rel="stylesheet" href="doc.css">
</head>
<body>
<div class="bg-effects">
  <div class="glow-orb cyan"></div>
  <div class="glow-orb green"></div>
  <div class="grid-pattern"></div>
</div>
<header class="doc-header">
  <a href="../index.html" class="logo-badge">IDT</a>
  <a href="../index.html" class="back-link">← Back to home</a>
</header>
<main class="doc-main">
<article class="md-content">
{body}
</article>
</main>
<footer class="doc-footer">
  IDT — Intelligent DNS Tool · Created by FJ™ Cybertronic Systems
</footer>
</body>
</html>
"""

def inline(text):
    text = html.escape(text, quote=False)
    # inline code first, protect it
    codes = []
    def stash(m):
        codes.append(m.group(1))
        return f"\x00CODE{len(codes)-1}\x00"
    text = re.sub(r"`([^`]+)`", stash, text)
    # links [t](u)
    text = re.sub(r"\[([^\]]+)\]\(([^)]+)\)",
                  r'<a href="\2">\1</a>', text)
    # bold then italic
    text = re.sub(r"\*\*([^*]+)\*\*", r"<strong>\1</strong>", text)
    text = re.sub(r"(?<!\*)\*([^*]+)\*(?!\*)", r"<em>\1</em>", text)
    # restore code
    for i, c in enumerate(codes):
        text = text.replace(f"\x00CODE{i}\x00", f"<code>{html.escape(c, quote=False)}</code>")
    return text

def md_to_html(md: str) -> str:
    lines = md.split("\n")
    out, i, n = [], 0, len(lines)

    def is_list_start(s):
        return re.match(r"^\s*(?:[-*+]|\d+\.)\s+", s)

    while i < n:
        line = lines[i]
        stripped = line.strip()

        # blank
        if not stripped:
            i += 1
            continue

        # fenced code
        m = re.match(r"^```(\w*)", stripped)
        if m:
            lang = m.group(1)
            i += 1
            buf = []
            while i < n and not lines[i].strip().startswith("```"):
                buf.append(lines[i]); i += 1
            i += 1  # skip closing fence
            code = html.escape("\n".join(buf), quote=False)
            cls = f' class="language-{lang}"' if lang else ""
            out.append(f"<pre><code{cls}>{code}</code></pre>")
            continue

        # heading
        m = re.match(r"^(#{1,6})\s+(.*)$", stripped)
        if m:
            level = len(m.group(1))
            text = inline(m.group(2).rstrip("#").strip())
            anchor = re.sub(r"[^a-z0-9]+", "-", m.group(2).lower()).strip("-")
            out.append(f'<h{level} id="{anchor}">{text}</h{level}>')
            i += 1
            continue

        # hr
        if re.match(r"^(-{3,}|\*{3,}|_{3,})$", stripped):
            out.append("<hr>")
            i += 1
            continue

        # blockquote
        if stripped.startswith(">"):
            buf = []
            while i < n and lines[i].strip().startswith(">"):
                buf.append(re.sub(r"^>\s?", "", lines[i].strip()))
                i += 1
            inner = md_to_html("\n".join(buf))
            out.append(f"<blockquote>{inner}</blockquote>")
            continue

        # table
        if "|" in stripped and i + 1 < n and re.match(r"^\s*\|?[\s:|-]+\|?\s*$", lines[i+1]):
            def cells(s):
                return [c.strip() for c in s.strip().strip("|").split("|")]
            header = cells(stripped)
            i += 2
            rows = []
            while i < n and "|" in lines[i] and lines[i].strip():
                rows.append(cells(lines[i])); i += 1
            t = ["<table><thead><tr>"]
            t += [f"<th>{inline(c)}</th>" for c in header]
            t.append("</tr></thead><tbody>")
            for r in rows:
                t.append("<tr>")
                t += [f"<td>{inline(c)}</td>" for c in r]
                t.append("</tr>")
            t.append("</tbody></table>")
            out.append("".join(t))
            continue

        # ordered list
        m = re.match(r"^(\d+)\.\s+(.*)$", stripped)
        if m:
            items, start = [], int(m.group(1))
            while i < n:
                mm = re.match(r"^\d+\.\s+(.*)$", lines[i].strip())
                if not mm: break
                items.append(inline(mm.group(1)))
                i += 1
            tag = "ol" if start == 1 else f'ol start="{start}"'
            out.append(f"<{tag}>" + "".join(f"<li>{x}</li>" for x in items) + f"</{tag}>")
            continue

        # unordered list
        if is_list_start(stripped) and not re.match(r"^\d", stripped):
            items = []
            while i < n:
                mm = re.match(r"^\s*[-*+]\s+(.*)$", lines[i])
                if not mm: break
                items.append(inline(mm.group(1)))
                i += 1
            out.append("<ul>" + "".join(f"<li>{x}</li>" for x in items) + "</ul>")
            continue

        # paragraph: gather until blank/heading/list
        buf = [stripped]
        i += 1
        while i < n:
            s = lines[i].strip()
            if not s: break
            if re.match(r"^(#{1,6})\s", s): break
            if s.startswith("```"): break
            if is_list_start(s): break
            if s.startswith(">"): break
            if re.match(r"^(-{3,}|\*{3,}|_{3,})$", s): break
            buf.append(s)
            i += 1
        out.append(f"<p>{inline(' '.join(buf))}</p>")

    return "\n".join(out)

def convert(md_path: Path):
    raw = md_path.read_text(encoding="utf-8")
    # strip YAML front matter if present
    if raw.startswith("---"):
        end = raw.find("\n---", 3)
        if end != -1:
            raw = raw[end+4:].lstrip("\n")
    title = md_path.stem.replace("_", " ").title()
    for line in raw.split("\n"):
        if line.startswith("# "):
            title = line[2:].strip()
            break
    body = md_to_html(raw)
    out_path = md_path.with_suffix(".html")
    out_path.write_text(
        TEMPLATE.format(title=html.escape(title), body=body),
        encoding="utf-8",
    )
    print(f"✓ {md_path.name} -> {out_path.name}")

def main():
    targets = sys.argv[1:] or [
        "DOCUMENTATION", "SECURITY", "DISCLAIMER",
        "ARCHITECTURE", "CHANGELOG", "COVERAGE", "TESTING",
    ]
    for name in targets:
        p = DOCS / f"{name}.md"
        if p.exists():
            convert(p)
        else:
            print(f"· skip {name}.md (not found)")

if __name__ == "__main__":
    main()
