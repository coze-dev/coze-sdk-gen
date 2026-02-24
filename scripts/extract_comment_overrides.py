#!/usr/bin/env python3
import ast
import argparse
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import yaml


def module_name_from_path(path: Path) -> str:
    parts = list(path.parts)
    if parts[-1] == "__init__.py":
        parts = parts[:-1]
    else:
        parts[-1] = parts[-1][:-3]
    return ".".join(parts)


def is_enum_class(node: ast.ClassDef) -> bool:
    for base in node.bases:
        name = ""
        if isinstance(base, ast.Name):
            name = base.id
        elif isinstance(base, ast.Attribute):
            name = base.attr
        if name in {"Enum", "IntEnum", "DynamicStrEnum"}:
            return True
    return False


def extract_comment_lines(lines: List[str], lineno: int, class_start: int) -> Tuple[List[str], bool]:
    index = lineno - 1
    if index < 0 or index >= len(lines):
        return [], False

    inline_lines: List[str] = []
    line = lines[index]
    if "#" in line:
        inline = line.split("#", 1)[1].strip()
        if inline:
            inline_lines = [inline]

    prev_comments: List[str] = []
    i = index - 1
    while i >= class_start:
        text = lines[i].strip()
        if not text:
            if prev_comments:
                break
            i -= 1
            continue
        if text.startswith("#"):
            prev_comments.append(text[1:].strip())
            i -= 1
            continue
        break

    if prev_comments:
        prev_comments.reverse()
        return [c for c in prev_comments if c], False
    return inline_lines, len(inline_lines) > 0


def walk_class(
    module: str,
    node: ast.ClassDef,
    lines: List[str],
    source: str,
    class_prefix: str,
    class_docstrings: Dict[str, str],
    class_docstring_styles: Dict[str, str],
    method_docstrings: Dict[str, str],
    method_docstring_styles: Dict[str, str],
    field_comments: Dict[str, List[str]],
    inline_field_comments: Dict[str, str],
    enum_member_comments: Dict[str, List[str]],
    inline_enum_member_comments: Dict[str, str],
) -> None:
    class_name = node.name if not class_prefix else f"{class_prefix}.{node.name}"
    class_key = f"{module}.{class_name}"
    class_doc = ast.get_docstring(node, clean=True)
    if class_doc:
        class_docstrings[class_key] = class_doc
        style = detect_docstring_style(source, node)
        if style:
            class_docstring_styles[class_key] = style

    enum_class = is_enum_class(node)
    class_start = max(node.lineno - 1, 0)
    for item in node.body:
        if isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
            key = f"{class_key}.{item.name}"
            method_doc = ast.get_docstring(item, clean=True)
            if method_doc:
                method_docstrings[key] = method_doc
                style = detect_docstring_style(source, item)
                if style:
                    method_docstring_styles[key] = style
            else:
                method_docstrings[key] = ""
            continue

        if isinstance(item, ast.ClassDef):
            walk_class(
                module,
                item,
                lines,
                source,
                class_name,
                class_docstrings,
                class_docstring_styles,
                method_docstrings,
                method_docstring_styles,
                field_comments,
                inline_field_comments,
                enum_member_comments,
                inline_enum_member_comments,
            )
            continue

        target_name: Optional[str] = None
        target_lineno: Optional[int] = None
        if isinstance(item, ast.Assign) and len(item.targets) == 1 and isinstance(item.targets[0], ast.Name):
            target_name = item.targets[0].id
            target_lineno = item.lineno
        elif isinstance(item, ast.AnnAssign) and isinstance(item.target, ast.Name):
            target_name = item.target.id
            target_lineno = item.lineno
        if not target_name or not target_lineno:
            continue

        comments, inline = extract_comment_lines(lines, target_lineno, class_start)
        if not comments:
            continue
        key = f"{class_key}.{target_name}"
        if enum_class:
            if inline and len(comments) == 1:
                inline_enum_member_comments[key] = comments[0]
            else:
                enum_member_comments[key] = comments
        else:
            if inline and len(comments) == 1:
                inline_field_comments[key] = comments[0]
            else:
                field_comments[key] = comments


def collect_class_method_keys(module: str, node: ast.ClassDef, class_prefix: str, out: set[str]) -> None:
    class_name = node.name if not class_prefix else f"{class_prefix}.{node.name}"
    class_key = f"{module}.{class_name}"
    for item in node.body:
        if isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
            out.add(f"{class_key}.{item.name}")
            continue
        if isinstance(item, ast.ClassDef):
            collect_class_method_keys(module, item, class_name, out)


def collect_method_keys(py_path: Path, module: str) -> set[str]:
    source = py_path.read_text(encoding="utf-8")
    tree = ast.parse(source)
    keys: set[str] = set()
    for node in tree.body:
        if isinstance(node, ast.ClassDef):
            collect_class_method_keys(module, node, "", keys)
    return keys


def extract_from_file(
    py_path: Path,
    module: str,
    class_docstrings: Dict[str, str],
    class_docstring_styles: Dict[str, str],
    method_docstrings: Dict[str, str],
    method_docstring_styles: Dict[str, str],
    field_comments: Dict[str, List[str]],
    inline_field_comments: Dict[str, str],
    enum_member_comments: Dict[str, List[str]],
    inline_enum_member_comments: Dict[str, str],
) -> None:
    source = py_path.read_text(encoding="utf-8")
    lines = source.splitlines()
    tree = ast.parse(source)
    for node in tree.body:
        if isinstance(node, ast.ClassDef):
            walk_class(
                module,
                node,
                lines,
                source,
                "",
                class_docstrings,
                class_docstring_styles,
                method_docstrings,
                method_docstring_styles,
                field_comments,
                inline_field_comments,
                enum_member_comments,
                inline_enum_member_comments,
            )


def detect_docstring_style(source: str, node: ast.AST) -> str:
    body = getattr(node, "body", None)
    if not body:
        return ""
    first = body[0]
    if not isinstance(first, ast.Expr):
        return ""
    value = getattr(first, "value", None)
    if not isinstance(value, ast.Constant) or not isinstance(value.value, str):
        return ""
    raw = ast.get_source_segment(source, value)
    if not raw:
        return ""
    text = raw.lstrip()
    i = 0
    while i < len(text) and text[i] in "rRuUbBfF":
        i += 1
    text = text[i:]
    quote = ""
    if text.startswith('"""'):
        quote = '"""'
    elif text.startswith("'''"):
        quote = "'''"
    if not quote:
        return ""
    after = text[len(quote):]
    if after.startswith("\n") or after.startswith("\r\n"):
        return "block"
    return "inline"


def main() -> None:
    parser = argparse.ArgumentParser(description="Extract comment/docstring overrides from legacy SDK.")
    parser.add_argument("--legacy", required=True, help="Legacy SDK root, e.g. exist-repo/coze-py/cozepy")
    parser.add_argument("--generated", required=True, help="Generated SDK root, e.g. exist-repo/coze-py-generated/cozepy")
    parser.add_argument("--out", required=True, help="Output yaml path")
    args = parser.parse_args()

    legacy_root = Path(args.legacy).resolve()
    generated_root = Path(args.generated).resolve()
    out_path = Path(args.out).resolve()

    class_docstrings: Dict[str, str] = {}
    class_docstring_styles: Dict[str, str] = {}
    method_docstrings: Dict[str, str] = {}
    method_docstring_styles: Dict[str, str] = {}
    field_comments: Dict[str, List[str]] = {}
    inline_field_comments: Dict[str, str] = {}
    enum_member_comments: Dict[str, List[str]] = {}
    inline_enum_member_comments: Dict[str, str] = {}

    generated_files = {
        p.relative_to(generated_root).as_posix()
        for p in generated_root.rglob("*.py")
    }
    for rel in sorted(generated_files):
        legacy_file = legacy_root / rel
        if not legacy_file.exists():
            continue
        module = module_name_from_path(Path("cozepy") / Path(rel))
        extract_from_file(
            legacy_file,
            module,
            class_docstrings,
            class_docstring_styles,
            method_docstrings,
            method_docstring_styles,
            field_comments,
            inline_field_comments,
            enum_member_comments,
            inline_enum_member_comments,
        )

    generated_method_keys: set[str] = set()
    for rel in sorted(generated_files):
        generated_file = generated_root / rel
        if not generated_file.exists():
            continue
        module = module_name_from_path(Path("cozepy") / Path(rel))
        generated_method_keys.update(collect_method_keys(generated_file, module))
    for key in sorted(generated_method_keys):
        if key not in method_docstrings:
            method_docstrings[key] = ""

    out_path.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "class_docstrings": dict(sorted(class_docstrings.items())),
        "class_docstring_styles": dict(sorted(class_docstring_styles.items())),
        "method_docstrings": dict(sorted(method_docstrings.items())),
        "method_docstring_styles": dict(sorted(method_docstring_styles.items())),
        "field_comments": dict(sorted(field_comments.items())),
        "inline_field_comments": dict(sorted(inline_field_comments.items())),
        "enum_member_comments": dict(sorted(enum_member_comments.items())),
        "inline_enum_member_comments": dict(sorted(inline_enum_member_comments.items())),
    }
    out_path.write_text(
        yaml.safe_dump(payload, sort_keys=False, allow_unicode=True, width=1000),
        encoding="utf-8",
    )


if __name__ == "__main__":
    main()
