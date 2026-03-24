// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

const config = require("./src/config.js");
const { setupOtlp, getLogger } = require("./src/otlp.js");

const stopOtlp = setupOtlp(config);

const { startServer } = require("./src/server.js");
const { startClient } = require("./src/client.js");

const logger = getLogger();

const stopServer = startServer(config, logger);
const stopClient = startClient(config, logger);

const gracefulShutdown = () => {
    Promise.all([
        stopClient(),
        stopServer(),
        stopOtlp(),
    ]).then(() => {
        console.log("[Main] 👋");
        process.exit(0);
    }).catch((error) => {
        console.error("[main] 🤔👋", error);
        process.exit(1);
    });
};

process.on("SIGINT", () => gracefulShutdown("SIGINT"));
process.on("SIGTERM", () => gracefulShutdown("SIGTERM"));

console.log("[main] 🚀");
