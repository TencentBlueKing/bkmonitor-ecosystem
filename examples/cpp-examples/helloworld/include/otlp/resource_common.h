// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

//
// Created by sandrincai on 2024/9/10.
//

// NOCC:build/header_guard(工具误报:)
#ifndef HELLOWORLD_INCLUDE_OTLP_RESOURCE_COMMON_H_
#define HELLOWORLD_INCLUDE_OTLP_RESOURCE_COMMON_H_

// C 系统头文件
#include <sys/utsname.h>
#include <unistd.h>

// C++ 系统头文件
#include <fstream>
#include <string>

// 第三方库
#include "opentelemetry/sdk/resource/resource.h"
#include "opentelemetry/sdk/resource/semantic_conventions.h"

// 本地头文件
#include "config.h"


namespace resource_sdk = opentelemetry::sdk::resource;

namespace internal {
    std::string GetProcessId() { return std::to_string(getpid()); }

    std::string GetOperatingSystem() {
        struct utsname buffer{};
        if (uname(&buffer) != 0) {
            return "unknown";
        }
        return std::string(buffer.sysname) + " " + buffer.release;
    }

    std::string GetHostName() {
        char hostname[1024];
        gethostname(hostname, sizeof(hostname));
        hostname[1023] = '\0';
        return {hostname};
    }

    resource_sdk::Resource CreateResource(const Config &config) {
        auto defaultResource = resource_sdk::Resource::GetDefault();
        auto resourceAttributes = resource_sdk::ResourceAttributes{
                // ❗️❗【非常重要】应用服务唯一标识
                {resource_sdk::SemanticConventions::kServiceName, config.ServiceName},
                {resource_sdk::SemanticConventions::kProcessPid,  GetProcessId()},
                {resource_sdk::SemanticConventions::kOsType,      GetOperatingSystem()},
                {resource_sdk::SemanticConventions::kHostName,    GetHostName()},
        };
        return defaultResource.Merge(resource_sdk::Resource::Create(resourceAttributes));
    }
}  // namespace internal

// NOCC:build/header_guard(工具误报:)
#endif  // HELLOWORLD_INCLUDE_OTLP_RESOURCE_COMMON_H_
