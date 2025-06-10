# Tutorial: Prune expired messages from your xmtpd database

xmtpd automatically marks messages as expired based on system-defined rules. These rules are set **per message type** and configured by the **payer**, not the node operator.

As a node operator, **you don't need to configure retention policies or handle billing**—the system takes care of that.

However, you **do need to run the prune process** to delete expired messages from your database. If you don’t, your node may experience **unnecessary data growth and degraded performance**.

## Message retention defaults

The following are typical system-defined message retention periods configured by payers:

- **Key packages**: Expire after 90 days
- **Welcome messages**: Expire after 90 days
- **Chat messages**: Expire after 30 days
- **Commit messages**: Retained indefinitely (for now)

xmtpd automatically marks a message as expired based on the retention policy attached upon message creation.

## Your responsibility as a node operator

To keep your database healthy and disk usage under control, **you must regularly prune expired messages**.

### What happens if you don't prune?

- Your node will continue to store expired data
- Disk usage will grow unnecessarily
- Performance of database queries may degrade over time
- Other nodes will not be affected—this is local to your setup

## Run the xmtpd-prune job

Use the official pruning tool provided by XMTP:

```bash
docker run ghcr.io/xmtp/xmtpd-prune
```

This tool:

- Deletes expired messages
- Only deletes messages that are safe to remove (i.e., those that have been processed)

### Automate xmtpd-prune

You can automate the running of the prune job via:

- A **cron job** on your host machine
- A **Kubernetes CronJob** in your cluster
- Any other task scheduler (e.g., systemd timer, CI/CD pipeline)

We recommend running the prune job **at least once a day** to avoid storage bloat.

## Required xmtpd-prune settings

When running `xmtpd-prune`, the following arguments are **required**:

| CLI flag                    | Environment variable                             | Description                                         |
|--------------------------------------------------------|--------------------------------------------------|-----------------------------------------------------|
| `--db.writer-connection-string`                        | `XMTPD_DB_WRITER_CONNECTION_STRING`              | Database writer connection string                   |
| `--contracts.settlement-chain.node-registry-address`   | `XMTPD_SETTLEMENT_CHAIN_NODE_REGISTRY_ADDRESS`   | Node registry contract address to determine DB name |
| `--signer.private-key`                                 | `XMTPD_SIGNER_PRIVATE_KEY`                       | Private key used to determine DB name          |

### Validation rule

All three fields **must be provided**. If any field is missing, the job will fail with: `missing required arguments: ...`

## Optional xmtpd-prune settings

| CLI flag                    | Environment variable             | Description                                  | Default |
|-----------------------------|----------------------------------|----------------------------------------------|---------|
| `--prune.max-prune-cycles` | `XMTPD_PRUNE_MAX_CYCLES`         | Maximum number of pruning cycles per run     | `10`    |
| `--prune.dry-run`          | `XMTPD_PRUNE_DRY_RUN`            | Run in dry mode (no actual deletions)        | `false` |

## Summary

- Message retention policies are handled by the system, not the node operator
- Pruning expired messages is the node operator's responsibility
- Run `ghcr.io/xmtp/xmtpd-prune` regularly to prevent data bloat and keep your node performant
