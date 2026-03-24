#!/bin/bash
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.


# 检查lint-md是否安装，未安装则自动安装
if ! command -v lint-md &> /dev/null; then
    echo "lint-md未安装，正在自动安装..."
    npm install -g @lint-md/cli
    if [ $? -ne 0 ]; then
        echo "❌ 安装lint-md失败，请手动执行: npm install -g @lint-md/cli"
        exit 1
    fi
    echo "✅ lint-md安装成功"
fi

# 获取暂存的Markdown文件(仅docs目录及其子目录)
changed_md_files=$(git diff --cached --name-only --diff-filter=ACM | grep -E '^docs/.*\.md$')

if [ -n "$changed_md_files" ]; then
    echo "检查Markdown文件规范..."
    any_modified=0  # 跟踪是否有文件被修改
    modified_files=()  # 存储被修复的文件名

    for file in $changed_md_files; do
        echo "检查: $file"

        # 执行检查并修复
        lint-md "$file"
        if [ $? -ne 0 ]; then
            echo "❌ $file 存在规范问题，请修复后再提交"
            exit 1
        fi

        # 检查文件是否被修改
        if ! git diff --quiet -- "$file"; then
            echo "🔄 $file 已自动修复"
            modified_files+=("$file")
            any_modified=1
        fi
    done

    echo "✅ 所有Markdown文件符合规范"

    # 如果有文件被修改，返回非零退出码
    # 这会触发pre-commit的自动处理：重新暂存文件并再次运行钩子
    if [ $any_modified -eq 1 ]; then
        echo "ℹ️ 已修复以下文件: ${modified_files[*]}，请重新暂存，并提交"
        exit 1
    fi
fi