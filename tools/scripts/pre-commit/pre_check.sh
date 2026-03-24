#!/bin/bash
# Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
# Copyright (C) 2017-2025 Tencent. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

# 如果是 Windows 平台（MSYS/Cygwin），直接跳过
case "$(uname -s)" in
  CYGWIN*|MINGW*|MSYS*)
    exit 0
    ;;
esac

# 优先执行 inner 目录下的脚本（如果存在）
INNER_SCRIPT="tools/scripts/pre-commit/inner/pre_check.sh"
if [ -f "$INNER_SCRIPT" ]; then
  exec "$INNER_SCRIPT" "$@"
fi

# shellcheck disable=SC2006
preci info > /dev/null 2>&1
if [ $? -ne "0" ]; then
  echo "system have not preci，skip preci"
else
  preci scan --pre-commit
fi
