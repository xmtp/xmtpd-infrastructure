# Runbook: Migrate from SHA-based database name to a fixed name

## Background

Older versions of xmtpd automatically named the node's database using a Keccak256 hash of the node's private key and the settlement chain node registry address, producing names like `xmtpd_a3f2c10d9e8b`. This naming scheme has several drawbacks:

- The name is opaque — you cannot tell which node owns which database by looking at it.
- Rotating the node's private key changes the derived name, silently leaving the old data behind.
- Infrastructure tooling (monitoring, backups) cannot predict the name in advance.

As of xmtpd v0.1.0, the hash-based fallback has been removed. The database name is now sourced exclusively from:

1. `XMTPD_DB_NAME_OVERRIDE` (env var / `--db.name-override` flag), or
2. The database name embedded in `XMTPD_DB_WRITER_CONNECTION_STRING` if the override is not set.

This runbook covers migrating a running node from the old `xmtpd_<sha>` name to a fixed name (e.g. `xmtpddata`) without data loss.

---

## Prerequisites

- Direct SQL access to the node's PostgreSQL instance (Aurora writer endpoint, or equivalent).
- The ability to update the node's environment variables and restart it.
- A maintenance window is recommended but not strictly required — `ALTER DATABASE … RENAME` is a metadata-only operation and completes instantly.

---

## Step 1 — Identify the current database name

Connect to the node's PostgreSQL instance and run:

```sql
SELECT datname FROM pg_database WHERE datname LIKE 'xmtpd_%';
```

You should see exactly one row, e.g.:

```
      datname
-------------------
 xmtpd_a3f2c10d9e8b
```

Note this name — you will need it in Step 3.

If there is also a `xmtpddata` row, it is an empty placeholder database created by your cloud provider at cluster provisioning time. Verify it is empty before proceeding:

```sql
\c xmtpddata
\dt
-- Expected: "Did not find any relations."
```

---

## Step 2 — Configure the fixed database name

Set `XMTPD_DB_NAME_OVERRIDE` to your chosen fixed name (e.g. `xmtpddata`) in your deployment configuration **before** restarting the node.

### Kubernetes (Helm)

In your `values.yaml`:

```yaml
databaseName: "xmtpddata"
```

This propagates to all xmtpd pods (api, sync, indexer, reporting) and the prune CronJob automatically.

### AWS Terraform

In your module calls:

```hcl
module "xmtpd_api" {
  source  = "..."
  db_name = "xmtpddata"
  # ...
}

module "xmtpd_worker" {
  source  = "..."
  db_name = "xmtpddata"
  # ...
}

module "xmtpd_prune" {
  source  = "..."
  db_name = "xmtpddata"
  # ...
}
```

**Do not apply or restart yet** — do the database rename in Step 3 first.

---

## Step 3 — Rename the database

While the node is still running (or after stopping it), execute the rename on the PostgreSQL writer endpoint.

If an empty `xmtpddata` placeholder exists, drop it first:

```sql
DROP DATABASE xmtpddata;
```

Then rename:

```sql
ALTER DATABASE "xmtpd_a3f2c10d9e8b" RENAME TO "xmtpddata";
```

Replace `xmtpd_a3f2c10d9e8b` with the actual name found in Step 1.

> **Note:** `ALTER DATABASE … RENAME` requires that no other sessions are connected to the database being renamed. If connections are open, terminate them first:
>
> ```sql
> SELECT pg_terminate_backend(pid)
> FROM pg_stat_activity
> WHERE datname = 'xmtpd_a3f2c10d9e8b'
>   AND pid <> pg_backend_pid();
> ```

---

## Step 4 — Apply the configuration and restart

Deploy the updated configuration from Step 2. The node will restart and connect to `xmtpddata`.

### Kubernetes

```sh
helm upgrade <release-name> ./helm/xmtpd -f your-values.yaml
```

### AWS Terraform

```sh
terraform apply
```

ECS will drain the old task and start a new one with `XMTPD_DB_NAME_OVERRIDE=xmtpddata`.

---

## Step 5 — Verify

Check the node logs for a successful connection message:

```
{"level":"info","msg":"successfully connected to database","namespace":"xmtpddata"}
```

Confirm the node is healthy and replicating:

```sh
# Kubernetes
kubectl logs -l app.kubernetes.io/role=xmtpd-server-api --tail=50

# AWS (CloudWatch)
aws logs tail /ecs/xmtpd-api --follow
```

---

## Rollback

If the node fails to start after the rename, rename the database back and redeploy the previous image:

```sql
ALTER DATABASE "xmtpddata" RENAME TO "xmtpd_a3f2c10d9e8b";
```

Then remove `XMTPD_DB_NAME_OVERRIDE` from your config and redeploy.
