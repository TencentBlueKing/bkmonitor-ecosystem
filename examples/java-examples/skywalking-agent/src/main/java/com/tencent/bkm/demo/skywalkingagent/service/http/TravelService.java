// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

package com.tencent.bkm.demo.skywalkingagent.service.http;

import com.tencent.bkm.demo.skywalkingagent.service.http.exception.FileNotFoundException;
import com.tencent.bkm.demo.skywalkingagent.service.http.exception.MySQLConnectTimeoutException;
import com.tencent.bkm.demo.skywalkingagent.service.http.exception.NetworkUnreachableException;
import com.tencent.bkm.demo.skywalkingagent.service.http.exception.UserNotFoundException;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Service;
import org.springframework.scheduling.annotation.Async;

import java.util.List;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Random;


@Service
public class TravelService {
    private static final String[] COUNTRIES = {
            "United States", "Canada", "United Kingdom", "Germany",
            "France", "Japan", "Australia", "China", "India", "Brazil"
    };
    private static final float ERROR_RATE = 0.1F;
    private static final Logger logger = LogManager.getLogger(TravelService.class);
    private final ApplicationContext applicationContext;

    public TravelService(ApplicationContext applicationContext) {
        this.applicationContext = applicationContext;
    };

    public String handle() {
        List<String> countries = choiceCountries();
        logger.info("get countries -> {}", countries);

        this.parallelVisit(countries);
        this.serialVisit(countries);
        return "Travel Success";
    };

    private List<String> choiceCountries() {
        ArrayList<String> copyCountries = new ArrayList<>(List.of(COUNTRIES));
        Collections.shuffle(copyCountries, new Random());
        return copyCountries.subList(0, 3);
    };

    private void parallelVisit(List<String> countries) {
        // 获取当前类的 bean 对象
        TravelService travelBean = this.applicationContext.getBean(TravelService.class);
        for (String country : countries) {
            travelBean.parallelTask(country);
        }
    };

    private void serialVisit(List<String> countries) {
        logger.info("Travel serialVisit start");
        for (String country : countries) {
            this.visit(country);
        }
        logger.info("Travel serialVisit end");
    };

    @Async
    public void parallelTask(String country) {
        logger.info("Travel parallelTask start");
        this.visit(country);
        logger.info("Travel parallelTask end");
    };

    private void visit(String country) {
        // Traces（调用链）- 模拟任务耗时
        // 随机休眠 200 ms~500 ms
        this.doSomething();
        // Traces（调用链）- 模拟错误
        // 10% 的概率抛出一个随机的错误
        this.tracesRandomErrorDemo();
    };

    private void doSomething() {
        Random random = new Random();
        int sleepTime = 200 + random.nextInt(300);
        try {
            Thread.sleep(sleepTime);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    };

    // Traces（调用链）- 异常事件、状态
    // Refer:
    private void tracesRandomErrorDemo() {
        try {
            throwRandomError();
        } catch (RuntimeException e) {
            logger.error("[tracesRandomErrorDemo] got error -> {}", e.getMessage());
            throw e;
        }
    };

    private void throwRandomError() {
        Random random = new Random();
        if (random.nextFloat() < ERROR_RATE) {
            int index = random.nextInt(4);
            switch (index) {
                case 0:
                    throw new FileNotFoundException();
                case 1:
                    throw new MySQLConnectTimeoutException();
                case 2:
                    throw new NetworkUnreachableException();
                case 3:
                    throw new UserNotFoundException();
                default:
                    break;
            }
        }
    };

}