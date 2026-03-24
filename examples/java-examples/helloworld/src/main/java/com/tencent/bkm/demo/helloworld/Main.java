// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.helloworld;

import com.tencent.bkm.demo.helloworld.service.Config;
import com.tencent.bkm.demo.helloworld.service.Service;
import com.tencent.bkm.demo.helloworld.service.impl.http.HttpService;
import com.tencent.bkm.demo.helloworld.service.impl.otlp.OtlpService;
import com.tencent.bkm.demo.helloworld.service.impl.profiling.ProfilingService;
import com.tencent.bkm.demo.helloworld.service.impl.querier.QuerierService;

import java.util.List;
import java.util.logging.Logger;


public final class Main {
    private static final Logger logger = Logger.getLogger(Main.class.getName());

    public static void main(String[] args) {
        Config config = new Config();

        List<Service> services = List.of(
                // otlp 放在依赖 opentelemetry 的模块前面进行初始化，避免运行报错
                new OtlpService(),
                new ProfilingService(),
                new HttpService(),
                new QuerierService()
        );

        for (Service service : services) {
            try {
                service.init(config);
            } catch (Exception e) {
                logger.info(String.format("[%s] failed to init: %s", service.getType(), e.getMessage()));
                System.exit(1);
            }
        }

        for (Service service : services) {
            try {
                service.start();
            } catch (Exception e) {
                logger.severe(String.format("[%s] failed to start: %s", service.getType(), e.getMessage()));
                System.exit(1);
            }
            logger.info(String.format("[%s] service started", service.getType()));
        }

        logger.info("[main] 🚀");

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            for (Service service : services) {
                try {
                    service.stop();
                } catch (Exception e) {
                    logger.severe(
                            String.format("[%s] failed to stop: %s", service.getClass().getName(), e.getMessage()));
                    continue;
                }
                logger.info(String.format("[%s] service stopped", service.getClass().getName()));
            }
        }));
    }
}