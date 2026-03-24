# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import os
import re
import shutil

from typing import Dict, Any, List, Generator
from dataclasses import dataclass

from context import base
from jinja2 import Environment, FileSystemLoader


class EnvType:
    OPEN = "open"
    INNER = "inner"

try:
    from context import inner
    ENV_TYPE = EnvType.INNER if inner is not None else EnvType.OPEN
except ImportError:
    inner = None
    ENV_TYPE = EnvType.OPEN

TEMPLATE_ROOT_PATH = "templates"
DOCS_PATH = "docs"


@dataclass
class DocsConfig:
    env_type: str
    context: Dict[str, Any]
    root_path: str


def convert_markdown_links_to_html(file_path: str):
    with open(file_path, "r", encoding="utf-8") as f:
        link_pattern = re.compile(r"(?<!!)\[(.+?)]\(([^\s*]*?)\)")
        content: str = link_pattern.sub(r'<a href="\2" target="_blank">\1</a>', f.read())
    with open(file_path, "w", encoding="utf-8") as f:
        f.write(content)


def get_dst_file_path(template_path: str, docs_config: DocsConfig) -> str:
    return os.path.join(docs_config.root_path, os.path.relpath(template_path, TEMPLATE_ROOT_PATH))


def render_template_to_file(template_path: str, docs_config: DocsConfig):
    if docs_config.env_type == EnvType.OPEN and "inner" in template_path.lower():
        return

    dst_file_path = get_dst_file_path(template_path, docs_config)
    os.makedirs(os.path.dirname(dst_file_path), exist_ok=True)
    if not template_path.endswith(".md"):
        shutil.copy2(template_path, dst_file_path)
        return

    convert_markdown_links_to_html(template_path)
    template_obj = Environment(loader=FileSystemLoader(searchpath=".")).get_template(template_path)
    with open(dst_file_path, "w", encoding="utf-8") as f:
        f.write(template_obj.render(**docs_config.context))
    convert_markdown_links_to_html(dst_file_path)


def get_docs_config_from_env() -> List[DocsConfig]:
    open_context: Dict[str, Any] = base.FieldManager.get_context(base.ScopeType.OPEN.value)
    configs: List[DocsConfig] = [
        DocsConfig(
            env_type=EnvType.OPEN,
            context=open_context,
            root_path=os.path.join(DOCS_PATH, EnvType.OPEN)
        )
    ]
    if ENV_TYPE == EnvType.INNER:
        configs.append(DocsConfig(
            env_type=EnvType.INNER,
            context=base.FieldManager.get_context(base.ScopeType.INNER.value),
            root_path=os.path.join(DOCS_PATH, EnvType.INNER)
        ))
    return configs


def get_all_template_paths() -> Generator[str, None, None]:
    return (
        os.path.join(dirpath, filename)
        for dirpath, dirnames, filenames in os.walk(TEMPLATE_ROOT_PATH)
        for filename in filenames
    )


if __name__ == "__main__":
    _docs_configs = get_docs_config_from_env()
    for _template_path in get_all_template_paths():
        for _config_obj in _docs_configs:
            render_template_to_file(_template_path, _config_obj)
