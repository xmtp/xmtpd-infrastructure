# MLS validation service: configuring chain RPC endpoints

The MLS validation service verifies Smart Contract Wallet (SCW) signatures as part of identity update validation. To do this, it must make RPC calls to the chain where the smart contract wallet is deployed.

A node processes identity updates from all users on the network. Because any user may have a smart contract wallet on any supported chain, **every node must have working RPC access to every chain in the list below.** Missing or broken endpoints will cause identity updates for affected wallets to fail validation indefinitely.

## The problem with default endpoints

The validation service ships with a set of default public RPC endpoints. These defaults are intentionally conservative — they use free, unauthenticated public nodes so the service works out of the box. In production, they will fail.

Common failure modes observed in the wild:

| Provider | Failure |
|---|---|
| `rpc.ankr.com/*` | HTTP 401 — Ankr now requires an API key for all requests |
| `*.llamarpc.com` | HTTP 429 / Cloudflare error 1015 — rate limited under production load; some subdomains have been removed entirely |

When these calls fail, the validation service returns an error to xmtpd, which logs it and retries the identity update. Without a fix, the retry loop runs indefinitely and no identity updates for affected chains are stored.

## Recommended solution: Alchemy

We already use Alchemy for app chain and settlement chain RPC access throughout xmtpd. Use the same Alchemy API key for the MLS validation service.

### Step 1: Enable all required networks in your Alchemy app

By default, an Alchemy app only has a subset of networks enabled. Before setting the environment variables below, open your app in the [Alchemy dashboard](https://dashboard.alchemy.com) and enable every network listed in the table below.

### Step 2: Set the environment variables

The validation service reads `CHAIN_RPC_<chain_id>` environment variables at startup and uses them to override the compiled-in defaults. Configure all of the following:

| Chain | eip155 ID | Alchemy network slug | Environment variable |
|---|---|---|---|
| Ethereum | 1 | `eth-mainnet` | `CHAIN_RPC_1` |
| Optimism | 10 | `opt-mainnet` | `CHAIN_RPC_10` |
| Polygon | 137 | `polygon-mainnet` | `CHAIN_RPC_137` |
| zkSync Era | 324 | `zksync-mainnet` | `CHAIN_RPC_324` |
| Base | 8453 | `base-mainnet` | `CHAIN_RPC_8453` |
| Arbitrum One | 42161 | `arb-mainnet` | `CHAIN_RPC_42161` |
| Linea | 59144 | `linea-mainnet` | `CHAIN_RPC_59144` |
| World Chain | 480 | `worldchain-mainnet` | `CHAIN_RPC_480` |
| Lens | 232 | `lens-mainnet` | `CHAIN_RPC_232` |
| Abstract | 2741 | `abstract-mainnet` | `CHAIN_RPC_2741` |
| Gnosis | 100 | `gnosis-mainnet` | `CHAIN_RPC_100` |

Alchemy endpoint format:

```
https://<network-slug>.g.alchemy.com/v2/<API_KEY>
```

## Helm configuration

Set the environment variables in your values override file for `mls-validation-service`:

```yaml
env:
  CHAIN_RPC_1: "https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_10: "https://opt-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_137: "https://polygon-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_324: "https://zksync-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_8453: "https://base-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_42161: "https://arb-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_59144: "https://linea-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_480: "https://worldchain-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_232: "https://lens-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_2741: "https://abstract-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  CHAIN_RPC_100: "https://gnosis-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
```

The API key should be stored as a Kubernetes secret and referenced via `valueFrom.secretKeyRef` rather than inlined in plain text.

## Docker / systemd configuration

Pass the variables directly to the container:

```bash
docker run \
  -e CHAIN_RPC_1="https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_10="https://opt-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_137="https://polygon-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_324="https://zksync-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_8453="https://base-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_42161="https://arb-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_59144="https://linea-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_480="https://worldchain-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_232="https://lens-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_2741="https://abstract-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  -e CHAIN_RPC_100="https://gnosis-mainnet.g.alchemy.com/v2/YOUR_API_KEY" \
  ghcr.io/xmtp/mls-validation-service
```

## Keeping this list up to date

The authoritative list of supported chains is maintained in libxmtp:
[`crates/xmtp_id/src/scw_verifier/chain_urls_default.json`](https://github.com/xmtp/libxmtp/blob/main/crates/xmtp_id/src/scw_verifier/chain_urls_default.json)

When new chains are added to that file, a corresponding `CHAIN_RPC_<id>` override should be added to your deployment.
