"""Guard bilingual pages, repository metadata and rendered diagram markup."""

import re
import unittest
from html.parser import HTMLParser
from pathlib import Path

from mkdocs.config import load_config


ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"
SITE = ROOT / "site"
REPOSITORY = "https://github.com/longxi-go/nemo"
OWNER = "The longxi-go/nemo Authors"


class Markup(HTMLParser):
    def __init__(self):
        super().__init__()
        self.repositories = []
        self.diagram_count = 0
        self.authors = []
        self.copyright = []
        self.in_copyright = False

    def handle_starttag(self, tag, attrs):
        attributes = dict(attrs)
        classes = attributes.get("class", "").split()
        if "mermaid" in classes:
            self.diagram_count += 1
        if tag == "a" and attributes.get("data-md-component") == "source":
            self.repositories.append(attributes.get("href"))
        if tag == "meta" and attributes.get("name") == "author":
            self.authors.append(attributes.get("content"))
        if "md-copyright__highlight" in classes:
            self.in_copyright = True

    def handle_endtag(self, tag):
        if tag == "div":
            self.in_copyright = False

    def handle_data(self, data):
        if self.in_copyright:
            self.copyright.append(data)


def pages():
    return sorted(path for path in DOCS.rglob("*.md") if not path.name.endswith(".zh.md"))


def target(source):
    relative = source.relative_to(DOCS)
    chinese = source.name.endswith(".zh.md")
    stem = source.name.removesuffix(".zh.md" if chinese else ".md")
    output = SITE / ("zh" if chinese else "") / relative.parent
    return output / ("index.html" if stem == "index" else f"{stem}/index.html")


def diagram_structure(block):
    """Compare graph identities and branches independently of translated labels."""
    lines = []
    for line in block.splitlines():
        line = line.strip()
        if line.startswith("%%"):
            continue
        line = re.sub(r'"[^"]*"', '""', line)
        line = re.sub(r"\|[^|]*\|", "||", line)
        if line.startswith("participant "):
            line = line.partition(" as ")[0]
        elif re.match(r"^(alt|else|opt)\b", line):
            line = line.split()[0]
        else:
            line = line.partition(":")[0]
        lines.append(line)
    return lines


class DocumentationBuild(unittest.TestCase):
    def test_all_pages_have_translations_and_matching_examples(self):
        for english in pages():
            chinese = english.with_name(english.stem + ".zh.md")
            with self.subTest(page=english.name):
                self.assertTrue(chinese.is_file())
                examples = r"```(?!mermaid)(\w+)\n(.*?)```"
                self.assertEqual(
                    re.findall(examples, english.read_text(), re.S),
                    re.findall(examples, chinese.read_text(), re.S),
                )

    def test_native_repository_component_and_authorship_on_every_page(self):
        config = load_config(config_file=str(ROOT / "mkdocs.yml"))
        self.assertEqual(config.repo_url, REPOSITORY)
        self.assertEqual(config.repo_name, "longxi-go/nemo")
        self.assertEqual(config.site_author, OWNER)
        self.assertEqual(config.theme["name"], "material")
        self.assertEqual(config.theme["icon"]["repo"], "fontawesome/brands/github")
        self.assertNotIn("custom_dir", config.theme)
        for source in DOCS.rglob("*.md"):
            with self.subTest(page=source.name):
                markup = Markup()
                markup.feed(target(source).read_text())
                self.assertTrue(markup.repositories)
                self.assertEqual(set(markup.repositories), {REPOSITORY})
                self.assertEqual(markup.authors, [OWNER])
                self.assertEqual(
                    "".join(markup.copyright).strip(),
                    f"Copyright © 2026 - present {OWNER}",
                )

    def test_native_repository_bundle_supports_release_stars_and_forks(self):
        bundles = list((SITE / "assets" / "javascripts").glob("bundle*.js"))
        self.assertTrue(bundles)
        javascript = "\n".join(path.read_text() for path in bundles)
        for token in ("/releases/latest", "stargazers_count", "forks_count", "md-source__facts"):
            with self.subTest(token=token):
                self.assertIn(token, javascript)

    def test_bilingual_diagrams_reach_the_site_as_mermaid(self):
        total = 0
        for english in pages():
            chinese = english.with_name(english.stem + ".zh.md")
            blocks = [
                re.findall(r"```mermaid\n(.*?)```", path.read_text(), re.S)
                for path in (english, chinese)
            ]
            with self.subTest(page=english.name):
                self.assertEqual(
                    [diagram_structure(block) for block in blocks[0]],
                    [diagram_structure(block) for block in blocks[1]],
                )
            for source, diagrams in zip((english, chinese), blocks):
                total += len(diagrams)
                with self.subTest(page=source.name):
                    markup = Markup()
                    markup.feed(target(source).read_text())
                    self.assertEqual(markup.diagram_count, len(diagrams))
        self.assertGreater(total, 0)


if __name__ == "__main__":
    unittest.main()
