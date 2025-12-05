# XMTPD Systemd Integration

This document explains how to run, test, and debug xmtpd using systemd.
It supports:
- Local sandbox testing via Docker
- Production deployment on Linux systems using systemd

The systemd configuration provided here supports xmtpd-api, xmtpd-workers, and pruner

## Directory Structure
```
systemd/
├── xmtpd-api.service
├── xmtpd-prune.service
├── xmtpd-worker.service
dev/docker/systemd.Dockerfile
xmtpd.env             (optional local env file)
```

Unit files in `systemd/` are copied into `/etc/systemd/system/` in the sandbox image.

## Configuration Layout

All configuration for xmtpd lives under:
```
/etc/xmtpd/
xmtpd.env        # environment variables
testnet.json     # contract configuration
```

Load the env file into your current shell
```
set -a
. /etc/xmtpd/xmtpd.env
set +a
```

Validate the config
```
env | grep XMTPD_
cat /etc/xmtpd/testnet.json | head
```

## Building the Systemd Sandbox Image

The sandbox image boots a real Ubuntu systemd instance and installs the xmtpd release, example systemd unit files, and testnet config.

Build the image:
```
./dev/docker/build-systemd
```

Included in the image:
- Ubuntu 24.04
- systemd as PID 1
- xmtpd v1.0.0 release binary
- /etc/xmtpd/testnet.json from XMTP smart contracts
- /etc/xmtpd/xmtpd.env if provided
- All units from systemd/

## Running the Systemd Sandbox

To run systemd properly inside Docker, run
```
./dev/docker/run-systemd
```

This creates a fully booted Linux system inside Docker.

Open a shell:
```
docker exec -it xmtpd-systemd bash
```


## Inspecting Systemd State
### Verify systemd is PID 1

```
ps -p 1 -o pid,comm,args
```

Expected:
```
1 systemd /lib/systemd/systemd
```

### Overall system state
```
systemctl is-system-running
```

### List installed unit files
```
systemctl list-unit-files --no-pager
```

### List active services
```
systemctl list-units --type=service --no-pager
```

## Working with Units
### Reload unit files
```
systemctl daemon-reload
```

### Start a service
```
systemctl start xmtpd-api.service
systemctl status xmtpd-api.service --no-pager
```

### Stop a service
```
systemctl stop xmtpd-api.service
```

### Enable a service on boot
```
systemctl enable xmtpd-api.service
```

## Inspecting Logs

View recent logs for a service:

```
journalctl -u xmtpd-api.service -n 50 --no-pager
```

View global logs:

```
journalctl -n 200 --no-pager
```

## Editing and Testing Unit Files

1. Modify a file under systemd/ on the host
2. Rebuild the sandbox image
3. Run the container
4. Test the unit:
```
systemctl cat xmtpd-api.service
systemctl daemon-reload
systemctl restart xmtpd-api.service
systemctl status xmtpd-api.service --no-pager
journalctl -u xmtpd-api.service -n 50 --no-pager
```

## Configuring xmtpd via EnvironmentFile

Example unit files include:

`EnvironmentFile=/etc/xmtpd/xmtpd.env`

Example xmtpd.env:
```
# App chain
XMTPD_APP_CHAIN_RPC_URL=https://example-app-rpc
XMTPD_APP_CHAIN_WSS_URL=wss://example-app-wss

# Settlement chain
XMTPD_SETTLEMENT_CHAIN_RPC_URL=https://example-settlement-rpc
XMTPD_SETTLEMENT_CHAIN_WSS_URL=wss://example-settlement-wss

# Database
XMTPD_DB_WRITER_CONNECTION_STRING=postgres://user:pass@host:5432/db?sslmode=disable

# MLS validation
XMTPD_MLS_VALIDATION_GRPC_ADDRESS=mls-validation:50051
```

Systemd passes these values to xmtpd.
