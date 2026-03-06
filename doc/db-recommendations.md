# PostgreSQL 17+ Configuration Guide for XMTP Node Operators

- [PostgreSQL 17+ Configuration Guide for XMTP Node Operators](#postgresql-17-configuration-guide-for-xmtp-node-operators)
  - [Overview](#overview)
  - [1. Hardware Reference](#1-hardware-reference)
  - [2. Memory](#2-memory)
  - [3. WAL (Write-Ahead Log)](#3-wal-write-ahead-log)
  - [4. Checkpoints](#4-checkpoints)
  - [5. I/O \& Query Planning](#5-io--query-planning)
  - [6. Parallelism](#6-parallelism)
    - [Partitioning](#partitioning)
    - [Statistics](#statistics)
  - [7. Autovacuum](#7-autovacuum)
  - [8. Connections](#8-connections)
  - [9. Logging \& Monitoring](#9-logging--monitoring)
  - [10. Replication \& High Availability](#10-replication--high-availability)
    - [Backups](#backups)
  - [11. Complete postgresql.conf (PostgreSQL 17+, 8 vCPU / 64 GB RAM / SSD)](#11-complete-postgresqlconf-postgresql-17-8-vcpu--64-gb-ram--ssd)
  - [12. Aurora-Specific Notes](#12-aurora-specific-notes)
    - [Parameters managed by Aurora (do not set)](#parameters-managed-by-aurora-do-not-set)
    - [Parameters that still apply on Aurora](#parameters-that-still-apply-on-aurora)
    - [Aurora-specific considerations](#aurora-specific-considerations)
  - [13. Scaling Guide](#13-scaling-guide)

## Overview

This guide provides PostgreSQL configuration recommendations for partners running XMTP nodes. It is based on our production experience running the service on **8 vCPU / 64 GB RAM** instances with SSD storage.

**Minimum supported version:** PostgreSQL **17** (or later). For Aurora deployments, use Aurora PostgreSQL **17-compatible** clusters (or later).

The recommendations target **vanilla PostgreSQL** (self-hosted, RDS standard, or any PostgreSQL-compatible managed service). If you are running AWS Aurora, see the [Aurora-Specific Notes](#12-aurora-specific-notes) section at the end for differences.

**Checking your version and instance:** Confirm version compatibility first:

```sql
SELECT
  current_setting('server_version') AS server_version,
  current_setting('server_version_num')::int AS server_version_num,
  CASE
    WHEN current_setting('server_version_num')::int >= 170000 THEN 'SUPPORTED (PG17+)'
    ELSE 'UNSUPPORTED (requires PG17+)'
  END AS support_status;
```

Then run one of the following parameter comparison queries. Use the **first query** for vanilla PostgreSQL (or any engine); use the **Aurora query** in [§ 12. Aurora-Specific Notes](#12-aurora-specific-notes) for Aurora. Save the query to a file and run e.g. `psql "postgresql://…" -f script.sql`, or paste it into your SQL client.

**Query for vanilla PostgreSQL (all parameters):**

```sql
WITH recommended (category, param_name, recommended_value, recommended_doc) AS (
  VALUES
    -- Memory (values for comparison: shared_buffers/effective_cache in 8kB, work_mem/maintenance in kB)
    ('Memory', 'shared_buffers', '2097152', '16GB'),
    ('Memory', 'effective_cache_size', '6291456', '48GB'),
    ('Memory', 'work_mem', '16384', '16MB'),
    ('Memory', 'maintenance_work_mem', '2097152', '2GB'),
    -- WAL (vanilla PG only; Aurora manages these)
    ('WAL', 'wal_buffers', '8192', '64MB'),
    ('WAL', 'max_wal_size', '8192', '8GB'),
    ('WAL', 'min_wal_size', '2048', '2GB'),
    ('WAL', 'wal_compression', 'on', 'on'),
    ('WAL', 'wal_level', 'replica', 'replica'),
    ('WAL', 'wal_writer_delay', '200', '200ms'),
    ('WAL', 'synchronous_commit', 'on', 'on'),
    ('WAL', 'full_page_writes', 'on', 'on'),
    ('WAL', 'wal_log_hints', 'on', 'on'),
    -- Checkpoints (vanilla PG only)
    ('Checkpoints', 'checkpoint_timeout', '900', '15min'),
    ('Checkpoints', 'checkpoint_completion_target', '0.9', '0.9'),
    ('Checkpoints', 'log_checkpoints', 'on', 'on'),
    -- I/O (SSD)
    ('I/O', 'random_page_cost', '1.1', '1.1'),
    ('I/O', 'seq_page_cost', '1.0', '1.0'),
    ('I/O', 'effective_io_concurrency', '200', '200'),
    ('I/O', 'maintenance_io_concurrency', '200', '200'),
    -- Parallelism
    ('Parallelism', 'max_parallel_workers_per_gather', '4', '4'),
    ('Parallelism', 'max_parallel_maintenance_workers', '4', '4'),
    -- Partitioning & planning
    ('Partitioning', 'enable_partitionwise_join', 'on', 'on'),
    ('Partitioning', 'enable_partitionwise_aggregate', 'on', 'on'),
    ('Statistics', 'default_statistics_target', '200', '200'),
    -- Autovacuum
    ('Autovacuum', 'autovacuum_vacuum_scale_factor', '0.05', '0.05'),
    ('Autovacuum', 'autovacuum_analyze_scale_factor', '0.02', '0.02'),
    ('Autovacuum', 'autovacuum_vacuum_cost_delay', '2', '2ms'),
    ('Autovacuum', 'autovacuum_vacuum_cost_limit', '1000', '1000'),
    ('Autovacuum', 'vacuum_cost_page_miss', '2', '2'),
    -- Connections
    ('Connections', 'max_connections', '5000', '5000'),
    -- Logging (1000 = 1s)
    ('Logging', 'log_min_duration_statement', '1000', '1000 (1s)'),
    ('Logging', 'log_lock_waits', 'on', 'on'),
    ('Logging', 'log_autovacuum_min_duration', '1000', '1000 (1s)'),
    ('Logging', 'log_checkpoints', 'on', 'on'),
    ('Logging', 'track_io_timing', 'on', 'on'),
    -- Replication (optional)
    ('Replication', 'max_wal_senders', '5', '5'),
    ('Replication', 'max_replication_slots', '5', '5'),
    ('Replication', 'hot_standby', 'on', 'on'),
    ('Replication', 'hot_standby_feedback', 'on', 'on')
),
current_settings AS (
  SELECT name, setting, unit
  FROM pg_settings
  WHERE name IN (SELECT param_name FROM recommended)
)
SELECT
  r.category,
  r.param_name AS parameter,
  r.recommended_doc AS recommended,
  COALESCE(c.setting, '(not set / Aurora-managed)') AS current_setting,
  c.unit AS unit,
  CASE
    WHEN c.setting IS NULL THEN 'N/A (Aurora or missing)'
    WHEN r.recommended_value = c.setting THEN 'OK'
    ELSE 'DIFFERS'
  END AS status
FROM recommended r
LEFT JOIN current_settings c ON c.name = r.param_name
ORDER BY r.category, r.param_name;
```

On Aurora, WAL/checkpoint/I/O parameters show as "N/A (Aurora or missing)" because Aurora manages them. For Aurora, use the query in [§ 12](#12-aurora-specific-notes) instead, which only checks parameters that apply.

---

## 1. Hardware Reference

Our production environment runs on **8 vCPU / 64 GB RAM** instances with SSD storage (e.g., `r6g.2xlarge`, `r6i.2xlarge`, `n2-highmem-8`). All parameter values in this guide are tuned for this specific instance size.

---

## 2. Memory

These parameters control how PostgreSQL uses RAM. Getting them right is the single most impactful tuning step. Values are in PostgreSQL's default units: **8 kB pages** for `shared_buffers` and `effective_cache_size`, **kB** for `work_mem` and `maintenance_work_mem`.

| Parameter              | Value   | Notes                                                                                                                                               |
| ---------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `shared_buffers`       | 2097152 | 16 GB in 8 kB pages. 25% of RAM. On dedicated database servers without a separate caching layer, you can go up to 40%. Requires restart.            |
| `effective_cache_size` | 6291456 | 48 GB in 8 kB pages. Planner hint (not a memory allocation). Tells the planner how much data it can expect in OS page cache + shared buffers.       |
| `work_mem`             | 16384   | 16 MB in kB. Memory per sort/hash operation. Each query can use multiple `work_mem` allocations. Conservative but safe with high connection counts. |
| `maintenance_work_mem` | 2097152 | 2 GB in kB. Memory for VACUUM, CREATE INDEX, and ALTER TABLE. Higher values speed up these operations significantly.                                |

---

## 3. WAL (Write-Ahead Log)

WAL configuration directly controls write throughput, crash recovery time, and I/O patterns. These parameters do not exist on Aurora (which replaces WAL with its own storage layer) but are critical for vanilla PostgreSQL. Size values use default units: **8 kB pages** for `wal_buffers`, **MB** for `max_wal_size` and `min_wal_size`, **ms** for `wal_writer_delay`.

| Parameter            | Value   | Notes                                                                                                                                                  |
| -------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `wal_buffers`        | 8192    | 64 MB in 8 kB pages. Shared memory for WAL writes before flushing to disk. Default (~1/32 of `shared_buffers`) is too small for write-heavy workloads. |
| `max_wal_size`       | 8192    | 8 GB in MB. Maximum WAL to accumulate before forcing a checkpoint. Default 1024 causes too-frequent checkpoints under write load.                      |
| `min_wal_size`       | 2048    | 2 GB in MB. Pre-allocate WAL files to avoid filesystem allocation overhead during writes.                                                              |
| `wal_compression`    | on      | Compresses full-page writes in WAL. Reduces WAL volume by 30-50%. On PG17+, use `lz4` for lower CPU cost when available in your build.                 |
| `wal_level`          | replica | Default. Only change to `logical` if you need logical replication.                                                                                     |
| `wal_writer_delay`   | 200     | Milliseconds. Default is fine. Reduce to 10 only for low-latency write requirements.                                                                   |
| `synchronous_commit` | on      | Keep `on` for data safety. Only set to `off` for non-critical write paths where you can tolerate losing the last few transactions on crash.            |
| `full_page_writes`   | on      | Required for crash recovery on ext4/xfs. Only disable on copy-on-write filesystems (ZFS, btrfs) with checksums enabled.                                |
| `wal_log_hints`      | on      | Enables `pg_rewind` for failover scenarios. Negligible performance cost.                                                                               |

---

## 4. Checkpoints

Checkpoints flush dirty pages from shared buffers to disk. Poorly tuned checkpoints cause I/O spikes that impact query latency. `checkpoint_timeout` is in **seconds**.

| Parameter                      | Value | Notes                                                                                                                                             |
| ------------------------------ | ----- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `checkpoint_timeout`           | 900   | Seconds (15 min). Spread checkpoints further apart than the default 300. Reduces I/O spikes at the cost of slightly longer crash recovery.        |
| `checkpoint_completion_target` | 0.9   | Spread checkpoint writes over 90% of the checkpoint interval, smoothing I/O. Set this explicitly to keep behavior consistent across environments. |
| `log_checkpoints`              | on    | Log checkpoint activity. Essential for monitoring checkpoint frequency and duration.                                                              |

---

## 5. I/O & Query Planning

| Parameter                    | HDD             | SSD/NVMe | Notes                                                                                                                                      |
| ---------------------------- | --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `random_page_cost`           | `4.0` (default) | `1.1`    | Tells the planner the relative cost of random I/O. The default 4.0 assumes spinning disks and discourages index scans. Set to 1.1 on SSDs. |
| `seq_page_cost`              | `1.0`           | `1.0`    | Default is fine for all storage types.                                                                                                     |
| `effective_io_concurrency`   | `2`             | `200`    | Number of concurrent I/O requests for bitmap heap scans. NVMe can handle 200+. Default is 1.                                               |
| `maintenance_io_concurrency` | `2`             | `200`    | Same as above, for VACUUM and maintenance.                                                                                                 |

---

## 6. Parallelism

| Parameter                          | Value | Notes                                                |
| ---------------------------------- | ----- | ---------------------------------------------------- |
| `max_parallel_workers_per_gather`  | 4     | Workers per query node. Diminishing returns above 4. |
| `max_parallel_maintenance_workers` | 4     | Workers for VACUUM, CREATE INDEX.                    |

### Partitioning

The XMTP workload uses partitioned tables. These must be enabled:

```text
enable_partitionwise_join = on
enable_partitionwise_aggregate = on
```

Both are `off` by default. They allow the planner to push joins and aggregates down into individual partitions, which can dramatically improve performance.

### Statistics

| Parameter                   | Value | Notes                                                                                                                                        |
| --------------------------- | ----- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `default_statistics_target` | 200   | Higher values give the planner more accurate distribution data for partitioned and TOAST-heavy tables. Cost is slightly longer ANALYZE runs. |

---

## 7. Autovacuum

Autovacuum is critical for PostgreSQL health. The XMTP workload produces significant dead tuples due to TOAST-heavy writes. Aggressive autovacuum tuning prevents table bloat and maintains query performance.

| Parameter                         | Value | Notes                                                                                                       |
| --------------------------------- | ----- | ----------------------------------------------------------------------------------------------------------- |
| `autovacuum_vacuum_scale_factor`  | 0.05  | Trigger vacuum when dead tuples exceed this fraction of live tuples. Default 0.2 is too lax.                |
| `autovacuum_analyze_scale_factor` | 0.02  | Trigger ANALYZE more frequently for up-to-date planner statistics.                                          |
| `autovacuum_vacuum_cost_delay`    | 2     | Milliseconds. Pause between vacuum I/O batches. Lower = faster vacuum. On SSDs, 2 is a good default.        |
| `autovacuum_vacuum_cost_limit`    | 1000  | I/O budget per vacuum cycle. Higher = vacuum processes more pages before pausing. Scale with instance size. |
| `vacuum_cost_page_miss`           | 2     | Cost assigned when vacuum must read a page from disk. Default 10 assumes HDD. Set to 2 on SSDs.             |

---

## 8. Connections

| Parameter         | Value  | Notes                                                                                                                                                                                       |
| ----------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `max_connections` | `5000` | We run 5000 in production. Each connection consumes ~10 MB of RAM. Ensure your instance has enough memory to support the connection count alongside `shared_buffers` and other allocations. |

---

## 9. Logging & Monitoring

| Parameter                     | Value | Notes                                                                                                           |
| ----------------------------- | ----- | --------------------------------------------------------------------------------------------------------------- |
| `log_min_duration_statement`  | 1000  | Milliseconds (1 s). Log queries taking longer than this. Helps identify slow queries without overwhelming logs. |
| `log_lock_waits`              | on    | Log when queries wait longer than `deadlock_timeout` for a lock.                                                |
| `log_autovacuum_min_duration` | 1000  | Milliseconds (1 s). Log vacuum operations that take longer than this.                                           |
| `log_checkpoints`             | on    | Log every checkpoint with timing and buffer statistics.                                                         |
| `track_io_timing`             | on    | Enables I/O timing in `EXPLAIN (ANALYZE, BUFFERS)`. Small overhead, high diagnostic value.                      |

---

## 10. Replication & High Availability

If running a primary + replica setup (recommended for production):

| Parameter               | Value     | Notes                                                                                                                                                                   |
| ----------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `wal_level`             | `replica` | Required for streaming replication. Use `logical` only if you need logical replication.                                                                                 |
| `max_wal_senders`       | `5`       | Maximum number of replication connections. Set to at least `num_replicas + 2` (for backups and monitoring).                                                             |
| `max_replication_slots` | `5`       | Match `max_wal_senders`. Prevents WAL removal before replicas have consumed it.                                                                                         |
| `hot_standby`           | `on`      | Allow read queries on replicas.                                                                                                                                         |
| `hot_standby_feedback`  | `on`      | Replicas report their oldest active transaction to the primary, preventing vacuum from removing rows the replica still needs. Prevents query cancellations on replicas. |

### Backups

For vanilla PostgreSQL, replace managed backup solutions with:

- **pgBackRest** or **barman** for continuous WAL archiving and point-in-time recovery (PITR)
- **pg_basebackup** for ad-hoc base backups
- Configure `archive_mode = on` and `archive_command` to ship WAL to object storage (S3, GCS)

---

## 11. Complete postgresql.conf (PostgreSQL 17+, 8 vCPU / 64 GB RAM / SSD)

This is a ready-to-use configuration block. Parameters match what we run in production.

```ini
# =============================================================================
# XMTP Node - PostgreSQL 17+ Configuration
# Instance: 8 vCPUs, 64 GB RAM, SSD/NVMe storage
# All values in PostgreSQL default units (8 kB pages, kB, MB, ms, s as applicable).
# =============================================================================

# MEMORY (shared_buffers, effective_cache_size in 8 kB pages; work_mem, maintenance_work_mem in kB)
shared_buffers = 2097152
effective_cache_size = 6291456
work_mem = 16384
maintenance_work_mem = 2097152

# WAL (vanilla PostgreSQL only -- Aurora manages these). wal_buffers in 8 kB pages; max/min_wal_size in MB; wal_writer_delay in ms
wal_buffers = 8192
max_wal_size = 8192
min_wal_size = 2048
wal_compression = on                   # on PG17+, consider 'lz4' for lower CPU cost if available
wal_level = replica
synchronous_commit = on
full_page_writes = on
wal_log_hints = on
wal_writer_delay = 200

# CHECKPOINTS (vanilla PostgreSQL only -- Aurora manages these). checkpoint_timeout in seconds
checkpoint_timeout = 900
checkpoint_completion_target = 0.9
log_checkpoints = on

# I/O
random_page_cost = 1.1
effective_io_concurrency = 200         # vanilla PostgreSQL only
maintenance_io_concurrency = 200       # vanilla PostgreSQL only

# PARALLELISM
max_parallel_workers_per_gather = 4
max_parallel_maintenance_workers = 4

# PARTITIONING & PLANNING
enable_partitionwise_join = on
enable_partitionwise_aggregate = on
default_statistics_target = 200

# AUTOVACUUM (autovacuum_vacuum_cost_delay in ms)
autovacuum_vacuum_scale_factor = 0.05
autovacuum_analyze_scale_factor = 0.02
autovacuum_vacuum_cost_delay = 2
autovacuum_vacuum_cost_limit = 1000
vacuum_cost_page_miss = 2

# CONNECTIONS
max_connections = 5000

# LOGGING (log_min_duration_statement, log_autovacuum_min_duration in ms)
log_min_duration_statement = 1000
log_lock_waits = on
log_autovacuum_min_duration = 1000
track_io_timing = on

# REPLICATION (if using replicas)
# max_wal_senders = 5
# max_replication_slots = 5
# hot_standby = on
# hot_standby_feedback = on
```

---

## 12. Aurora-Specific Notes

If you are running **AWS Aurora PostgreSQL** instead of vanilla PostgreSQL, use Aurora PostgreSQL **17-compatible** (or later), and be aware of these differences. To verify your Aurora instance, run the following query (it only checks parameters that apply on Aurora; WAL, checkpoints, and some I/O parameters are managed by Aurora and omitted):

```sql
WITH recommended (category, param_name, recommended_value, recommended_doc) AS (
  VALUES
    -- Memory (25% RAM shared_buffers on Aurora)
    ('Memory', 'shared_buffers', '2097152', '16GB (25% of 64GB)'),
    ('Memory', 'effective_cache_size', '6291456', '48GB'),
    ('Memory', 'work_mem', '16384', '16MB'),
    ('Memory', 'maintenance_work_mem', '2097152', '2GB'),
    -- I/O (Aurora manages effective_io_concurrency / maintenance_io_concurrency)
    ('I/O', 'random_page_cost', '1.1', '1.1'),
    -- Statistics
    ('Statistics', 'default_statistics_target', '200', '200'),
    -- Parallelism
    ('Parallelism', 'max_parallel_workers_per_gather', '4', '4'),
    ('Parallelism', 'max_parallel_maintenance_workers', '4', '4'),
    -- Partitioning
    ('Partitioning', 'enable_partitionwise_join', 'on', 'on'),
    ('Partitioning', 'enable_partitionwise_aggregate', 'on', 'on'),
    -- Autovacuum
    ('Autovacuum', 'autovacuum_vacuum_scale_factor', '0.05', '0.05'),
    ('Autovacuum', 'autovacuum_analyze_scale_factor', '0.02', '0.02'),
    ('Autovacuum', 'autovacuum_vacuum_cost_delay', '2', '2ms'),
    ('Autovacuum', 'autovacuum_vacuum_cost_limit', '1000', '1000'),
    ('Autovacuum', 'vacuum_cost_page_miss', '2', '2'),
    -- Connections
    ('Connections', 'max_connections', '5000', '5000'),
    -- Logging
    ('Logging', 'log_min_duration_statement', '1000', '1000 (1s)'),
    ('Logging', 'log_lock_waits', 'on', 'on'),
    ('Logging', 'log_autovacuum_min_duration', '1000', '1000 (1s)'),
    ('Logging', 'track_io_timing', 'on', 'on')
),
current_settings AS (
  SELECT name, setting, unit
  FROM pg_settings
  WHERE name IN (SELECT param_name FROM recommended)
)
SELECT
  r.category,
  r.param_name AS parameter,
  r.recommended_doc AS recommended,
  COALESCE(c.setting, '(not set)') AS current_setting,
  c.unit AS unit,
  CASE
    WHEN c.setting IS NULL THEN 'MISSING'
    WHEN r.recommended_value = c.setting THEN 'OK'
    ELSE 'DIFFERS'
  END AS status
FROM recommended r
LEFT JOIN current_settings c ON c.name = r.param_name
ORDER BY r.category, r.param_name;
```

### Parameters managed by Aurora (do not set)

Aurora replaces the PostgreSQL storage and WAL subsystems with its own distributed storage layer. The following parameters are either not available or ignored:

- `wal_buffers`, `max_wal_size`, `min_wal_size` -- Aurora has no traditional WAL files
- `checkpoint_timeout`, `checkpoint_completion_target` -- Aurora does not perform PostgreSQL-style checkpoints
- `wal_compression`, `full_page_writes` -- Aurora's log shipping is proprietary
- `effective_io_concurrency`, `maintenance_io_concurrency` -- managed by Aurora
- `huge_pages` -- managed by Aurora
- `wal_level` -- fixed to `replica`

### Parameters that still apply on Aurora

All memory, parallelism, partitioning, autovacuum, connection, and logging parameters from this guide apply to Aurora. Aurora uses the same PostgreSQL query executor.

However, not all of these are modifiable on Aurora. Only parameters explicitly set through an RDS cluster parameter group (or instance parameter group) can be changed. The parameters we configure in our Terraform profiles are:

- `shared_buffers`, `effective_cache_size`, `work_mem`, `maintenance_work_mem`
- `random_page_cost`
- `default_statistics_target`
- `max_parallel_workers_per_gather`, `max_parallel_maintenance_workers`
- `enable_partitionwise_join`, `enable_partitionwise_aggregate`
- `autovacuum_vacuum_scale_factor`, `autovacuum_analyze_scale_factor`, `autovacuum_vacuum_cost_delay`, `autovacuum_vacuum_cost_limit`, `vacuum_cost_page_miss`
- `max_connections`
- `log_min_duration_statement`, `log_lock_waits`, `log_autovacuum_min_duration`, `track_io_timing`

Any parameter not in this list is either managed by Aurora internally or uses the Aurora default and cannot be overridden via the parameter group.

### Aurora-specific considerations

- **`shared_buffers` at 25% of RAM** is optimal on Aurora. Aurora has its own buffer cache layer on top of `shared_buffers`, so going higher wastes memory.
- **`synchronous_commit = off`** on Aurora skips waiting for Aurora's storage acknowledgement. This improves write latency but risks losing the last few committed transactions on crash. Aurora still provides 6-copy durability at the storage layer.
- **`pg_stat_statements`** is available on Aurora but must be added via the `shared_preload_libraries` parameter in the RDS parameter group.

---

## 13. Scaling Guide

- **CPU consistently above 70%** -- move to the next vCPU tier
- **Autovacuum can't keep up** (table bloat growing) -- increase `autovacuum_vacuum_cost_limit`, or scale up
- **Checkpoint warnings in logs** ("checkpoints are occurring too frequently") -- increase `max_wal_size`
- **Replication lag growing** -- scale up primary or check `max_wal_senders`
- **Queries being spilled to disk** -- Increase `work_mem`
