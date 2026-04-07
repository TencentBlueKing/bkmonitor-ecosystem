# 指标类型

## 1. 计数器（Counter）

计数器类型代表一种样本数据**单调递增**的指标，在没有发生重置的情况下只增不减，其样本值应该是不断增大的。

例如，使用 Counter 类型的指标来表示服务的请求数、已完成的请求数或错误的发生次数等。在 Prometheus client SDK 中，主要使用 Inc() 和 Add(float64) 这 2 个函数。

## 2. 仪表盘（Gauge）

Gauge 类型表示可以任意上升或下降的单个数值的指标。

通常用于记录**当前**的 CPU、内存使用情况等测量值，也用于可变的`“计数”`，例如并发请求数（每个统计周期之间都可能存在差异）。

## 3. 直方图（Histogram）

直方图会采集观测值（通常是请求时长或响应大小等）并按可配置的桶进行计数。它还会计算所有观测值的总和。

在程序接口服务中，由于 1 秒的请求通常都不止一个，比如 1 秒有 1000 个请求，950 个请求平均响应在 10 ms 以下，50 个请求在 50 ms 以上。

使用 Counter 或者 Gauge 类型，都不适合求一组数据的最大值，最小值或平均值作为最终的指标统计。

例如，使用最大值，会漏掉最小值；反之一样；使用平均值，则整体数值不符合实际情况，产生了长尾效应。

当一种新的指标类型产生，需要一组数据按照其分布规律去组合排列。

```shell
# HELP prometheus_tsdb_compaction_chunk_size_bytes Final size of chunks on their first compaction
# TYPE prometheus_tsdb_compaction_chunk_size_bytes histogram
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="32"}                5
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="48"}               25
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="72"}               35
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="108"}              39
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="162"}              39
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="243"}            1503
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="364.5"}          1673
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="546.75"}         1774
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="820.125"}        1810
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="1230.1875"}      1853
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="1845.28125"}     1856
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="2767.921875"}    1856
prometheus_tsdb_compaction_chunk_size_bytes_bucket{le="+Inf"}           1856
prometheus_tsdb_compaction_chunk_size_bytes_sum                       471152
prometheus_tsdb_compaction_chunk_size_bytes_count                       1856
```

直方图类型的有 3 种指标名，假设指标是 `<basename>`，则其指标由以下组成：

* 样本的值分布在 Bucket 中的数量，命名为 `<basename>_bucket{le="<上边界>"}`，有多个样本分布在不同区间。

* 所有样本的总和，命名为 `<basename>_sum`。

* 所有样本的总数，命名为 `<basename>_count`，其值和 `<basename>_bucket{le="+Inf"}` 相同。

## 4. 摘要（Summary）

摘要与直方图类似，它会对持续产生的数据（如请求耗时、响应大小等）进行实时抽样分析。在固定长度移动时间窗口内，它不仅记录数据点的**总数量和数值总和**，还能快速计算出当前窗口内数据的分位值（例如中位数、75% 分位数等）。

```shell
# HELP prometheus_target_interval_length_seconds Actual intervals between scrapes.
# TYPE prometheus_target_interval_length_seconds summary
prometheus_target_interval_length_seconds{interval="15s",quantile="0.01"} 14.99828357
prometheus_target_interval_length_seconds{interval="15s",quantile="0.05"} 14.99869915
prometheus_target_interval_length_seconds{interval="15s",quantile="0.5"}  15.000018812
prometheus_target_interval_length_seconds{interval="15s",quantile="0.9"}  15.00112985
prometheus_target_interval_length_seconds{interval="15s",quantile="0.99"} 15.001921368
prometheus_target_interval_length_seconds_sum{interval="15s"}             7455.018291232004
prometheus_target_interval_length_seconds_count{interval="15s"}           497
```

摘要类型的有 3 种指标名，假设指标是 `<basename>`，则其指标由以下组成：

* 样本的值分布在 Bucket 中的数量，命名为 `<basename>{quantile="<分位数>"}`。

* 所有样本的总和，命名为 `<basename>_sum`。

* 所有样本的总数，命名为 `<basename>_count` ，其值和 `<basename>_bucket{le="Inf"}` 相同。

## 5. 了解更多

进一步了解以下内容：

* 了解如何进行 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/cookbook/Quickstarts/metrics/http/README.md" target="_blank">自定义指标 HTTP 上报</a>。