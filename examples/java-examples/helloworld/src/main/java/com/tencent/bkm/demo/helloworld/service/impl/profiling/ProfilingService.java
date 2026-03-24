// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.helloworld.service.impl.profiling;

import com.tencent.bkm.demo.helloworld.service.Config;
import com.tencent.bkm.demo.helloworld.service.Service;
import io.pyroscope.http.Format;
import io.pyroscope.javaagent.EventType;
import io.pyroscope.javaagent.PyroscopeAgent;

public class ProfilingService implements Service {

    private ProfilingConfig config;
    private io.pyroscope.javaagent.config.Config pyroscopeConfig;

    @Override
    public String getType() {
        return "profiling";
    }

    @Override
    public void init(Config config) throws Exception {
        this.config = new ProfilingConfig(
                config.isEnableProfiling(), config.getServiceName(), config.getProfilingEndpoint(), config.getToken());

        this.pyroscopeConfig = new io.pyroscope.javaagent.config.Config.Builder()
                //❗❗【非常重要】请传入应用 Token
                .setAuthToken(config.getToken())
                //❗❗【非常重要】数据上报地址，请根据页面指引提供的接入地址进行填写
                .setServerAddress(config.getProfilingEndpoint())
                //❗❗【非常重要】应用服务唯一标识
                .setApplicationName(this.config.getServiceName())
                .setProfilingEvent(EventType.ITIMER)
                .setFormat(Format.JFR)
                .build();
    }

    @Override
    public void start() throws Exception {
        if (this.config.getEnabled()) {
            PyroscopeAgent.start(this.pyroscopeConfig);
        }
    }

    @Override
    public void stop() throws Exception {
        if (this.config.getEnabled()) {
            PyroscopeAgent.stop();
        }
    }
}
