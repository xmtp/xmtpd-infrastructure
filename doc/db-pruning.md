# Tutorial: Managing Data Retention in XMTPD

## Overview

XMTPD automatically deletes old messages based on system-defined expiration rules. These rules are set **per message type** and configured by the **payer**, not the node operator. As a node operator, **you don't need to configure expiration policies or handle billing** — the system takes care of that.

However, you **do need to run the prune process** to ensure expired data is removed from your database. If you don’t, your node may experience **unnecessary data growth and degraded performance**.

## Message Retention Defaults

The following are typical system-managed retention periods:

- **Key Packages**: expire after 90 days
- **Welcome Messages**: expire after 90 days
- **Chat Messages**: expire after 30 days
- **Commit Messages**: retained indefinitely (for now)

These durations are automatically enforced based on the retention policy attached to each message when it is created.

## Your Responsibility as a Node Operator

To keep your database healthy and disk usage under control, **you must regularly delete expired messages**.

### Run the Prune Job

Use the official pruning tool provided by XMTP:

```bash
docker run ghcr.io/xmtp/xmtpd-prune
```

This tool:
- Deletes messages that have passed their expiration time
- Only deletes messages that are safe to remove (i.e. those that have been processed)

## Automate It
You can run the prune job via:
- A **cron job** on your host machine
- A **Kubernetes CronJob** in your cluster
- Any other task scheduler (e.g. systemd timer, CI/CD pipeline)

We recommend running the prune job **at least once a day** to avoid storage bloat.

## Required Settings

When running `xmtpd-prune`, the following arguments are **required**:

| CLI Flag                                               | Environment Variable                             | Description                                         |
|--------------------------------------------------------|--------------------------------------------------|-----------------------------------------------------|
| `--db.writer-connection-string`                        | `XMTPD_DB_WRITER_CONNECTION_STRING`              | Database writer connection string                   |
| `--contracts.settlement-chain.node-registry-address`   | `XMTPD_SETTLEMENT_CHAIN_NODE_REGISTRY_ADDRESS`   | Node registry contract address to determine DB name |
| `--signer.private-key`                                 | `XMTPD_SIGNER_PRIVATE_KEY`                       | Private key used to determine DB name          |

### Validation Rule

All three fields **must be provided**. If any is missing, the job will fail with:  
`missing required arguments: ...`

## Optional Settings

| CLI Flag                    | Environment Variable             | Description                                  | Default |
|-----------------------------|----------------------------------|----------------------------------------------|---------|
| `--prune.max-prune-cycles` | `XMTPD_PRUNE_MAX_CYCLES`         | Maximum number of pruning cycles per run     | `10`    |
| `--prune.dry-run`          | `XMTPD_PRUNE_DRY_RUN`            | Run in dry mode (no actual deletions)        | `false` |


## What Happens If You Don't Prune?
- Your node will continue to store expired data
- Disk usage will grow unnecessarily
- Performance of database queries may degrade over time
- Other nodes will not be affected — this is local to your setup

## Summary
- Retention is handled by the system, not the node operator
- Pruning is your responsibility
- Run `ghcr.io/xmtp/xmtpd-prune` regularly
- Prevent data bloat and keep your node performant