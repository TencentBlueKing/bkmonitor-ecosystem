// Tencent is pleased to support the open source community by making 蓝鲸智云 - 监控平台 (BlueKing - Monitor) available.
// Copyright (C) 2017-2025 Tencent. All rights reserved.
// Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.

const http = require('http');

function startClient(config) {
    const options = {
        hostname: config.serverAddress,
        port: config.serverPort,
        path: '/tasks',
        method: 'GET',
    };
    const timer = setInterval(() => {
        const req = http.request(options, (res) => {
            res.resume();
        });
        req.on('error', (err) => {
            console.error(`[client] got error -> ${err.message}`);
        });
        req.end();
    }, 3000);

    return async () => {
        clearInterval(timer);
    };
}

module.exports = { startClient };
