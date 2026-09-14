// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

const express = require('express');
const { wrapWithLabels } = require('./profiling.js');

const retained = [];

function intensiveTask(durationMs = 500) {
    const start = Date.now();
    let total = 0;
    while (Date.now() - start < durationMs) {
        const numbers = [];
        for (let i = 0; i < 10000; i += 1) {
            numbers.push(Math.random());
        }
        total += numbers.reduce((sum, value) => sum + value, 0);
        if (retained.length < 32) {
            retained.push(Buffer.alloc(256 * 1024));
        } else {
            retained.shift();
        }
    }
    return total;
}

function startServer(config) {
    const app = express();

    app.get('/tasks', (req, res) => {
        let result;
        // Wall / CPU profile 动态标签，用于区分不同代码路径
        wrapWithLabels({ handler: 'tasks' }, () => {
            result = intensiveTask();
        });
        res.json({ result });
    });

    const server = app.listen(config.serverPort, '0.0.0.0', () => {
        console.log(`[server] listening on ${config.serverPort}`);
    });

    return () => new Promise((resolve, reject) => {
        server.close((err) => {
            if (err) {
                reject(err);
                return;
            }
            resolve();
        });
    });
}

module.exports = { startServer };
