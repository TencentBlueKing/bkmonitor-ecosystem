#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

import argparse
import os
import re
import sys
from pathlib import Path
from typing import Any

# MIT License template
MIT_LICENSE_TEMPLATE = """Tencent is pleased to support the open source community by making \
蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
Copyright (C) 2017-2025 Tencent. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License."""

LANGUAGE_CONFIG: dict[str, dict[str, Any]] = {
    "python": {
        "extensions": [".py"],
        "comment_line": "#",
    },
    "go": {
        "extensions": [".go"],
        "comment_line": "//",
    },
    "java": {
        "extensions": [".java"],
        "comment_line": "//",
    },
    "cpp": {
        "extensions": [".cpp", ".cc", ".cxx", ".c", ".h", ".hpp"],
        "comment_line": "//",
    },
    "javascript": {
        "extensions": [".js", ".jsx", ".ts", ".tsx"],
        "comment_line": "//",
    },
    "shell": {
        "extensions": [".sh", ".bash"],
        "comment_line": "#",
    },
}

# 用于识别旧协议头的关键词模式
LICENSE_PATTERNS = [
    r"Licensed under the MIT License",
    r"蓝鲸智云.*监控平台",
    r"BlueKing.*Monitor",
    r"Copyright.*THL A29 Limited",
    r"Tencent is pleased to support",
]


def detect_language(file_path: str) -> str | None:
    """Detect programming language based on file extension."""
    ext: str = Path(file_path).suffix.lower()
    for lang, config in LANGUAGE_CONFIG.items():
        if ext in config["extensions"]:
            return lang
    return None


def is_special_header_line(line: str) -> bool:
    """
    Check if line is a special header line that should be preserved.

    This includes:
    - Shebang lines: #!/usr/bin/env python3
    - Encoding declarations: # -*- coding: utf-8 -*-
    - Vim modelines: # vim: set ...
    - Emacs modelines: # -*- mode: python -*-
    """
    stripped: str = line.strip()
    if not stripped:
        return False

    # Shebang
    if stripped.startswith("#!"):
        return True

    # Python encoding declaration: # -*- coding: utf-8 -*- or # coding: utf-8 or # encoding: utf-8
    if re.match(r'^#.*?(-\*-.*?)?coding[=:]\s*[-\w.]+', stripped):
        return True

    # Vim modeline
    if re.match(r'^#\s*vim?:', stripped, re.IGNORECASE):
        return True

    # Emacs modeline (file variables)
    if re.match(r'^#.*-\*-.*-\*-', stripped):
        return True

    return False


def get_skip_lines_count(lines: list[str]) -> int:
    """
    Get the number of special header lines to skip at the beginning of the file.

    These lines (shebang, encoding declaration, etc.) should be preserved
    before the license header.
    """
    skip = 0
    for line in lines:
        if is_special_header_line(line):
            skip += 1
        else:
            break
    return skip


def has_license_header(content: str) -> bool:
    """Check if content already has MIT license header."""
    return any(re.search(pattern, content) for pattern in LICENSE_PATTERNS)


def generate_license_header(lang: str, config: dict[str, Any]) -> str:
    """Generate license header for specific language."""
    lines: list[str] = MIT_LICENSE_TEMPLATE.strip().split("\n")
    comment_line: str = config["comment_line"]

    # 使用行注释风格生成 header
    header_lines: list[str] = [f"{comment_line} {line}" for line in lines]
    return "\n".join(header_lines)


def insert_license_header(file_path: str, dry_run: bool = False) -> tuple[bool, str]:
    """
    Insert license header into file if missing.

    Returns:
        Tuple of (modified, message)
    """
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
            lines = content.split("\n")
    except UnicodeDecodeError:
        return False, f"Skipped (binary or non-UTF8 file): {file_path}"
    except Exception as e:  # pylint: disable=broad-except
        return False, f"Error reading {file_path}: {e}"

    # Check if already has license
    if has_license_header(content):
        return False, f"Already has license: {file_path}"

    # Detect language
    lang: str | None = detect_language(file_path)
    if not lang:
        return False, f"Unsupported file type: {file_path}"

    config: dict[str, Any] = LANGUAGE_CONFIG[lang]

    # Generate license header
    license_header: str = generate_license_header(lang, config)

    # Handle special header lines (shebang, encoding declaration, etc.)
    insert_position: int = get_skip_lines_count(lines)

    # Build new content
    new_content: str
    if insert_position > 0:
        # Has shebang
        before: str = "\n".join(lines[:insert_position])
        after: str = "\n".join(lines[insert_position:])
        new_content = f"{before}\n{license_header}\n\n{after}"
    else:
        # No shebang
        new_content = f"{license_header}\n\n{content}"

    # Remove excessive blank lines
    new_content = re.sub(r"\n{4,}", "\n\n\n", new_content)

    if dry_run:
        return True, f"Would add license to: {file_path}"

    # Write back
    try:
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(new_content)
        return True, f"Added license to: {file_path}"
    except Exception as e:  # pylint: disable=broad-except
        return False, f"Error writing {file_path}: {e}"


def find_source_files(root_dir: str, extensions: list[str] | None = None) -> list[str]:
    """Find all source files in directory tree."""
    source_files: list[str] = []

    # Collect all extensions if not specified
    if extensions is None:
        extensions = []
        config: dict[str, Any]
        for config in LANGUAGE_CONFIG.values():
            extensions.extend(config["extensions"])
    for root, _, files in os.walk(root_dir):
        file: str
        for file in files:
            file_path: str = os.path.join(root, file)

            # Check extension
            if any(file.endswith(ext) for ext in extensions):
                source_files.append(file_path)

    return source_files


def main() -> int:
    parser: argparse.ArgumentParser = argparse.ArgumentParser(
        description="Check and insert MIT license headers in source files"
    )
    parser.add_argument(
        "files",
        nargs="*",
        help="Files to check (for pre-commit mode). If not provided, scans entire project.",
    )
    parser.add_argument(
        "--fix",
        action="store_true",
        help="Insert missing license headers (default is check-only)",
    )
    parser.add_argument(
        "--all",
        action="store_true",
        help="Process all files in the project",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be changed without actually modifying files",
    )

    args: argparse.Namespace = parser.parse_args()

    # Determine files to process
    if args.all or not args.files:
        print("Scanning for source files in current directory...")
        files_to_check = find_source_files(".")
        print(f"Found {len(files_to_check)} source files")
    else:
        # 只处理支持的代码文件（能检测到语言的文件）
        files_to_check = [f for f in args.files if detect_language(f)]

    if not files_to_check:
        print("No files to check")
        return 0

    # Process files
    modified_files: list[str] = []
    skipped_files: list[str] = []
    error_files: list[str] = []

    for file_path in files_to_check:
        if not os.path.exists(file_path):
            continue

        if args.fix or args.dry_run:
            # Fix mode: insert missing license headers
            modified, message = insert_license_header(file_path, dry_run=args.dry_run)
            print(message)
            if modified:
                modified_files.append(file_path)
            elif "Error" in message:
                error_files.append(file_path)
            else:
                skipped_files.append(file_path)
        else:
            # Check-only mode
            try:
                with open(file_path, "r", encoding="utf-8") as f:
                    content = f.read()
                lang = detect_language(file_path)
                if lang and not has_license_header(content):
                    print(f"Missing license header: {file_path}")
                    modified_files.append(file_path)
            except Exception as e:  # pylint: disable=broad-except
                print(f"Error checking {file_path}: {e}")
                error_files.append(file_path)

    # Summary
    print("\n" + "=" * 60)
    print("Summary:")
    print(f"  Total files checked: {len(files_to_check)}")
    print(f"  Files missing license: {len(modified_files)}")
    print(f"  Files skipped: {len(skipped_files)}")
    print(f"  Errors: {len(error_files)}")

    if modified_files and not args.fix and not args.dry_run:
        print("\nRun with --fix to add missing license headers")
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
