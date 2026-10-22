# -*- coding: utf-8 -*-
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


import tempfile
import unittest
from pathlib import Path

from main import convert_markdown_links_to_html


class ConvertMarkdownLinksTest(unittest.TestCase):
    def test_array_type_and_image_badge_remain_intact(self):
        source = (
            '| `field` | `string[]` | ![Development](https://example.com/badge) |\n'
            '[Reference](https://example.com/reference)\n'
        )
        expected = source.replace(
            '[Reference](https://example.com/reference)',
            '<a href="https://example.com/reference" target="_blank">Reference</a>',
        )
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "fields.md"
            path.write_text(source, encoding="utf-8")
            convert_markdown_links_to_html(str(path))
            self.assertEqual(path.read_text(encoding="utf-8"), expected)
            convert_markdown_links_to_html(str(path))
            self.assertEqual(path.read_text(encoding="utf-8"), expected)
