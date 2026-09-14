// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

const config = require('./src/config.js');
const { startProfiling, stopProfiling } = require('./src/profiling.js');
const { startServer } = require('./src/server.js');
const { startClient } = require('./src/client.js');

startProfiling(config);

const stopServer = startServer(config);
const stopClient = startClient(config);

const gracefulShutdown = () => {
    Promise.all([
        stopClient(),
        stopServer(),
        stopProfiling(),
    ]).then(() => {
        console.log('[main] 👋');
        process.exit(0);
    }).catch((error) => {
        console.error('[main] 🤔👋', error);
        process.exit(1);
    });
};

process.on('SIGINT', gracefulShutdown);
process.on('SIGTERM', gracefulShutdown);

console.log('[main] 🚀');
