// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.skywalkingagent.service.querier;

import org.springframework.web.client.RestTemplate;
import org.springframework.stereotype.Service;
import org.springframework.scheduling.annotation.Scheduled;

import org.apache.logging.log4j.Logger;
import org.apache.logging.log4j.LogManager;


@Service
public class QuerierService{
    private final QuerierConfig querierConfig;
    private final Logger logger = LogManager.getLogger(getClass().getName());
    private final RestTemplate restTemplate = new RestTemplate();

    public QuerierService(QuerierConfig querierConfig) {
        this.querierConfig = querierConfig;
    };

    @Scheduled(fixedRate = 1000)
    private void queryHelloWorld() {
        try {
            logger.info("[queryTravel] send request");
            String response = this.restTemplate.getForObject(this.querierConfig.getEndpoint(), String.class);
            logger.info("[queryTravel] received: {}", response);
        } catch (Exception e) {
            logger.error("[queryTravel] got error -> {}", e.getMessage());
        }
    }
}