# Tutorial: Protect your xmtpd QueryApi with rate limits

xmtpd ships with an optional, Redis-backed rate limiter that protects the
read-path `QueryApi` from noisy or abusive edge clients. As a node operator,
you are responsible for deciding whether to enable it and for sizing the
buckets to match your traffic.

This guide walks you through what the rate limiter does, how to turn it on,
and how to tune it safely.

Tracking issue: [xmtp/xmtpd-infrastructure#60](https://github.com/xmtp/xmtpd-infrastructure/issues/60).
Implementation: [xmtp/xmtpd#1938](https://github.com/xmtp/xmtpd/pull/1938).

## What gets rate-limited

The rate limiter wraps the **QueryApi** service only:

- `QueryEnvelopes`
- `SubscribeTopics`
- `GetInboxIds`
- `GetNewestEnvelope`

The following traffic is **not** rate-limited by this subsystem:

- `ReplicationApi` — node-to-node traffic, authenticated separately.
- `PublishApi` — payer-authenticated writes, billed on-chain.
- `NotificationApi` — internal fan-out.
- `MetadataApi` — operator and observability endpoints.

If your workload has hot spots outside `QueryApi`, you need a different
control (nginx `limit_req`, a CDN-level rule, etc.).

## How the limiter decides

Every `QueryApi` call is classified into a **tier**:

| Tier | Who | Policy |
|------|-----|--------|
| Tier 0 | Verified peer nodes (JWT accepted by `ServerAuthInterceptor`) | **Bypass** — no limits. |
| Tier 2 | Everything else (unauthenticated edge clients) | Token-bucket limits apply. |

Tier 2 requests are counted against a **per-IP token bucket** with two
refill windows in series: a short window (per minute) and a long window
(per hour). Both must have tokens for the call to be allowed. The IP is
extracted from the TCP peer address with optional trusted-proxy
`X-Forwarded-For` peeling; IPv6 clients are normalized to their `/64`
prefix so one subscriber cannot burn tokens by rotating addresses inside
their own subnet.

Per-call cost is sub-linear in the amount of work requested:

- `QueryEnvelopes` / `SubscribeTopics` cost `ceil(sqrt(numTopics))` tokens.
- `GetInboxIds` and `GetNewestEnvelope` cost 1 token.

`SubscribeTopics` has an extra `opens-per-minute` sub-limit and a
retrospective drain: long-lived streams pay tokens at close time based on
how long they were held, so a client that opens a subscription and sits on
it for hours will eventually be throttled from opening new ones.

### What happens when a client is over budget

- Unary calls return gRPC `ResourceExhausted` (`connect.CodeResourceExhausted`).
- Subscription opens return the same error **before** any catch-up is done.
- Live subscriptions are also force-closed after either
  `XMTPD_RATE_LIMIT_STREAM_IDLE_TIMEOUT` of silence or
  `XMTPD_RATE_LIMIT_STREAM_MAX_DURATION` wall-clock lifetime, whichever
  comes first.

### What happens when Redis goes away

The limiter is wrapped in a **circuit breaker** that fails *open*: if
Redis is unreachable, Tier 2 traffic is admitted as if the limiter were
disabled, the breaker trips after
`XMTPD_RATE_LIMIT_BREAKER_FAILURE_THRESHOLD` consecutive errors, and
Redis is re-probed after
`XMTPD_RATE_LIMIT_BREAKER_COOLDOWN`. Your node keeps serving; you lose
rate-limit enforcement until Redis comes back.

This is **not** a substitute for Redis reliability. A permanently broken
Redis means permanently unenforced rate limits.

## Before you enable it

You need a Redis instance reachable from the xmtpd process. Any modern
Redis 6+ works. High availability is strongly recommended for production:

- In Kubernetes, a managed Redis (Memorystore, ElastiCache) or a Bitnami
  `redis` chart in sentinel/cluster mode.
- On a single VM, a co-located `redis-server` is fine for testing but
  every Redis outage will flip your node into fail-open mode.

At startup, xmtpd issues a `PING` to Redis and **fails fast** if it is
unreachable. This is intentional: a node that cannot even reach Redis at
boot is almost certainly misconfigured, and you want that loud, not
silent.

## Minimal configuration

Rate limiting is off by default. Set two env vars to enable it:

```sh
XMTPD_RATE_LIMIT_ENABLE=true
XMTPD_REDIS_URL=redis://redis.internal:6379/0
```

That is the minimum. Everything else has a default that is sane for a
medium-sized public node.

## Full configuration reference

All variables below belong to the xmtpd process. They are also available
as CLI flags under the `--rate-limit.*` and `--redis.*` groups if you
prefer flags.

### Redis connection

| Variable | Default | Description |
|---|---|---|
| `XMTPD_REDIS_URL` | _(empty)_ | Redis connection URL. Supports `redis://`, `rediss://` (TLS), and the cluster/sentinel URL forms accepted by `go-redis/v9`. **Required when rate limiting is enabled.** |
| `XMTPD_REDIS_KEY_PREFIX` | `xmtpd:` | Namespace for all xmtpd Redis keys. Useful if you share a Redis with other services. |
| `XMTPD_REDIS_CONNECT_TIMEOUT` | `10s` | How long to wait for the startup PING before giving up. |

### Enabling the interceptor

| Variable | Default | Description |
|---|---|---|
| `XMTPD_RATE_LIMIT_ENABLE` | `false` | Master switch. When false, the interceptor is not wired in and no Redis connection is opened. |

### Tier 2 bucket sizes

| Variable | Default | Description |
|---|---|---|
| `XMTPD_RATE_LIMIT_T2_PER_MINUTE_CAPACITY` | `60` | Per-IP, per-minute token capacity. Each `GetInboxIds` / `GetNewestEnvelope` spends 1 token; a `QueryEnvelopes` over 100 topics spends 10. |
| `XMTPD_RATE_LIMIT_T2_PER_HOUR_CAPACITY` | `1200` | Per-IP, per-hour token capacity, applied in addition to the per-minute bucket. |
| `XMTPD_RATE_LIMIT_T2_SUBSCRIBE_OPENS_PER_MINUTE` | `10` | Dedicated sub-limit: how many times per minute a single IP may **open** a `SubscribeTopics` stream. Independent of the query bucket. |

Both query-bucket limits must have tokens for a call to be allowed. The
hour bucket exists to catch slow burns that slip under the minute bucket.

### Subscription drain

| Variable | Default | Description |
|---|---|---|
| `XMTPD_RATE_LIMIT_DRAIN_INTERVAL_MINUTES` | `5` | Minutes per drain interval. A held stream is billed `DRAIN_AMOUNT` tokens for every interval (or partial interval) it was open at close time. |
| `XMTPD_RATE_LIMIT_DRAIN_AMOUNT` | `1` | Tokens per drain interval. |

With defaults, a stream held for one hour pays `ceil(60/5) * 1 = 12`
retrospective tokens against the query bucket at close. A client that
sits on 100 subscribe streams for an hour will blow through the per-hour
bucket.

### Stream lifetime

| Variable | Default | Description |
|---|---|---|
| `XMTPD_RATE_LIMIT_STREAM_IDLE_TIMEOUT` | `15m` | Cancel a `SubscribeTopics` stream that has had no envelope activity for this long. |
| `XMTPD_RATE_LIMIT_STREAM_MAX_DURATION` | `60m` | Hard cap on subscription stream lifetime regardless of activity. |

Both timers are disabled when set to `0`. Leaving them at the defaults
means well-behaved clients reconnect every hour, which is safer than
letting goroutines and Redis state accumulate indefinitely.

### Circuit breaker

| Variable | Default | Description |
|---|---|---|
| `XMTPD_RATE_LIMIT_BREAKER_FAILURE_THRESHOLD` | `5` | Consecutive Redis errors before the breaker opens. Denials from the limiter are **not** failures; only transport/script errors count. |
| `XMTPD_RATE_LIMIT_BREAKER_COOLDOWN` | `10s` | How long the breaker stays open before letting one probe through. |
| `XMTPD_RATE_LIMIT_REDIS_CALL_TIMEOUT` | `50ms` | Per-call Redis deadline. Exceeding this counts as a breaker failure. |

The defaults are tuned to absorb a Redis restart without blocking
traffic. Lowering `BREAKER_COOLDOWN` below a few seconds is usually a
bad idea — it will hammer Redis during its recovery window.

### Trusted proxy peeling

| Variable | Default | Description |
|---|---|---|
| `XMTPD_RATE_LIMIT_TRUSTED_PROXY_CIDRS` | _(empty)_ | Comma-separated list of CIDRs whose `X-Forwarded-For` header is trusted. Example: `10.0.0.0/8,172.16.0.0/12`. |

If your node sits behind an L7 load balancer (nginx, Envoy, a cloud LB),
set this to the LB subnet. Otherwise every client looks like the
load-balancer IP and shares one bucket.

If you leave it empty, xmtpd uses the raw peer address and ignores
`X-Forwarded-For`. That is the correct behavior when the node is
exposed directly to the internet — an attacker can always forge
`X-Forwarded-For`, so trusting it unconditionally would let anyone pick
their own bucket.

## Example configurations

### Small public node behind nginx on the same VM

```sh
XMTPD_RATE_LIMIT_ENABLE=true
XMTPD_REDIS_URL=redis://127.0.0.1:6379/0
XMTPD_RATE_LIMIT_TRUSTED_PROXY_CIDRS=127.0.0.0/8
# defaults are fine for everything else
```

### Kubernetes with a managed Redis and an ingress controller

```yaml
env:
  - name: XMTPD_RATE_LIMIT_ENABLE
    value: "true"
  - name: XMTPD_REDIS_URL
    valueFrom:
      secretKeyRef:
        name: xmtpd-redis
        key: url
  - name: XMTPD_RATE_LIMIT_TRUSTED_PROXY_CIDRS
    value: "10.0.0.0/8"           # cluster pod CIDR — tune to your cluster
  - name: XMTPD_RATE_LIMIT_T2_PER_MINUTE_CAPACITY
    value: "120"                  # higher; we front-run a CDN
  - name: XMTPD_RATE_LIMIT_T2_PER_HOUR_CAPACITY
    value: "3000"
```

Put the Redis URL in a Secret, not a ConfigMap — if you ever use `rediss://`
with an AUTH token, the URL is credentials.

### High-traffic node that wants to observe first

Enable the limiter but set the buckets so high that normal traffic
never gets denied. Watch the metrics below for a week, then tighten.

```sh
XMTPD_RATE_LIMIT_ENABLE=true
XMTPD_REDIS_URL=redis://redis.internal:6379/0
XMTPD_RATE_LIMIT_T2_PER_MINUTE_CAPACITY=100000
XMTPD_RATE_LIMIT_T2_PER_HOUR_CAPACITY=2000000
XMTPD_RATE_LIMIT_T2_SUBSCRIBE_OPENS_PER_MINUTE=10000
```

## Verifying it is working

### 1. Startup log

With rate limiting enabled, xmtpd logs an enable line at boot:

```
rate limit interceptor enabled  t2_per_minute=60  t2_per_hour=1200  t2_subscribe_opens_per_minute=10
```

If you do not see that line, `XMTPD_RATE_LIMIT_ENABLE` was not parsed as
true.

### 2. Prometheus metrics

The subsystem exposes four metrics on the regular `/metrics` endpoint:

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `xmtpd_rate_limit_decisions_total` | Counter | `service`, `method`, `tier`, `outcome` | Every decision the interceptor makes. `outcome` is one of `allowed`, `denied`, `bypassed`, `failed_open`. |
| `xmtpd_rate_limit_circuit_breaker_state` | Gauge | — | `0` = closed (healthy), `1` = half-open (probing), `2` = open (Redis unreachable, failing open). |
| `xmtpd_rate_limit_circuit_breaker_trips_total` | Counter | — | How many times the breaker has tripped. A non-zero increment means Redis had a bad moment. |
| `xmtpd_rate_limit_stream_terminations_total` | Counter | `reason` | Why subscribe streams closed: `client_close`, `server_close`, `idle`, `max_duration`. |

Useful alerts:

- `increase(xmtpd_rate_limit_circuit_breaker_trips_total[5m]) > 0` — Redis
  flapped; investigate.
- `xmtpd_rate_limit_circuit_breaker_state != 0 for 1m` — Redis is down or
  slow and your node is fail-open.
- `sum by (method) (rate(xmtpd_rate_limit_decisions_total{outcome="denied"}[5m]))` —
  which method is bearing the brunt of denials.
- `rate(xmtpd_rate_limit_stream_terminations_total{reason="max_duration"}[5m])` —
  a burst here means clients are holding streams past the max. Either the
  max is too tight or the client is misbehaving.

### 3. Manual smoke test

From outside the node (or from a cluster pod with network access):

```sh
for i in $(seq 1 200); do
  grpcurl -d '{"inbox_ids": ["test"]}' \
    -import-path . -proto message_api.proto \
    your-node.example.com:5050 \
    xmtp.xmtpv4.message_api.QueryApi/GetInboxIds
done
```

With defaults, you should start seeing `ResourceExhausted` responses after
roughly 60 successful calls in the first minute.

## Operational gotchas

- **Redis is a shared dependency.** Two xmtpd processes pointing at the
  same Redis and the same `XMTPD_REDIS_KEY_PREFIX` share buckets — that is
  usually what you want behind a load balancer. Two unrelated deployments
  sharing a Redis and the same prefix will corrupt each other; set a
  unique prefix per deployment.

- **Client IP accuracy matters more than bucket size.** Misconfigured
  `TRUSTED_PROXY_CIDRS` is the single most common reason the limiter
  under- or over-penalizes users. Verify with a `grpcurl` test from a
  known client IP and check the Redis keys:

  ```sh
  redis-cli --scan --pattern 'xmtpd:rl:t2:q:*' | head
  ```

  You should see keys ending in the actual client IP, not your LB IP.

- **The limiter is a per-node control.** Nothing about Tier 0 or Tier 2 is
  propagated between replicas — a client that shards its traffic across
  `N` nodes gets `N × bucket` effective capacity. If that is a problem for
  you, put the rate limiter at the ingress instead (nginx, Envoy, CDN)
  and leave `XMTPD_RATE_LIMIT_ENABLE=false`.

- **Tier 0 is honored by the `VerifiedNodeRequestCtxKey` context flag set
  by `ServerAuthInterceptor`.** If you replace or disable the auth
  interceptor, every request becomes Tier 2. Keep them wired up together.

- **The drain is retrospective, not live.** A client that opens a stream
  and immediately holds it burns opens-per-minute tokens at open time
  and query-bucket tokens at close time. The query bucket can go
  *negative*, and the client will be denied new work until it refills.
  The debt persists in Redis until paid off — the TTL is sized to cover
  the full refill-from-debt period.

## Turning it off

Set `XMTPD_RATE_LIMIT_ENABLE=false` (or unset the variable) and restart
the node. No Redis cleanup is required; leftover buckets will expire on
their own within one hour at most.
