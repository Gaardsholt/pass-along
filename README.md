# Pass-along

[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7427/badge)](https://bestpractices.coreinfrastructure.org/projects/7427)

The external server serves the API and static UI on port `8080` by default.

The internal server serves `/healthz`, `/readyz`, and `/metrics` on port `8888` by default.

## Server config

The following config can be set via environment variables.

| Variable                                                | Required | Default               |
| ------------------------------------------------------- | :------: | --------------------- |
| [SERVER_SALT](#SERVER_SALT)                             |          |                       |
| [DATABASE_TYPE](#DATABASE_TYPE)                         |          | in-memory             |
| [REDIS_SERVER](#REDIS_SERVER)                           |          | localhost             |
| [REDIS_PORT](#REDIS_PORT)                               |          | 6379                  |
| [SERVER_PORT](#SERVER_PORT)                             |          | 8080                  |
| [HEALTH_PORT](#HEALTH_PORT)                             |          | 8888                  |
| [LOG_LEVEL](#LOG_LEVEL)                                 |          | info                  |
| [VALID_FOR_OPTIONS](#VALID_FOR_OPTIONS)                 |          | 3600,7200,43200,86400 |
| [MAX_SECRET_BYTES](#MAX_SECRET_BYTES)                   |          | 10485760              |
| [MAX_FILES](#MAX_FILES)                                 |          | 20                    |
| [MAX_FILE_SIZE_BYTES](#MAX_FILE_SIZE_BYTES)             |          | 104857600             |
| [ENABLE_HSTS](#ENABLE_HSTS)                             |          | false                 |
| [HSTS_MAX_AGE_SECONDS](#HSTS_MAX_AGE_SECONDS)           |          | 31536000              |
| [GRACEFUL_SHUTDOWN_SECONDS](#GRACEFUL_SHUTDOWN_SECONDS) |          | 25                    |
| [READINESS_DRAIN_SECONDS](#READINESS_DRAIN_SECONDS)     |          | 5                     |
| [SHUTDOWN_HARD_SECONDS](#SHUTDOWN_HARD_SECONDS)         |          | 3                     |

### SERVER_SALT

Used as the PBKDF2 salt for secret encryption and decryption.

The current implementation does not require this value and does not validate its length or entropy. For production deployments, set it to a stable, high-entropy value and protect it as secret material. Changing it prevents existing stored secrets from being decrypted.

### DATABASE_TYPE

Can either be `in-memory` or `redis`.

### REDIS_SERVER

Address to your redis server.

### REDIS_PORT

Used to specify the port your redis server is using.

### SERVER_PORT

Listen port for api and ui endpoint.

### HEALTH_PORT

Listen port for health endpoint, used mainly for liveness probes.

### LOG_LEVEL

Used to specify log levels. Valid values are `debug`, `info`, `warn`, and `error`. Unknown values fall back to `info`.

### VALID_FOR_OPTIONS

Which options are available in the UI for secret expiration.
Only these values are accepted server-side for `expires_in`.

### MAX_SECRET_BYTES

Maximum size of secret text content.

### MAX_FILES

Maximum number of attached files.

### MAX_FILE_SIZE_BYTES

Maximum size per file attachment.

### ENABLE_HSTS

Enable HSTS response header. Keep disabled unless TLS is correctly terminated upstream.

### HSTS_MAX_AGE_SECONDS

HSTS `max-age` value when `ENABLE_HSTS=true`.

### GRACEFUL_SHUTDOWN_SECONDS

Total shutdown budget after the process receives `SIGTERM` or `SIGINT`. This includes readiness draining, HTTP shutdown, and hard-cancel delay. Keep this lower than the platform termination grace period so cleanup can complete before the process is killed.

The HTTP shutdown timeout is calculated as `GRACEFUL_SHUTDOWN_SECONDS - READINESS_DRAIN_SECONDS - SHUTDOWN_HARD_SECONDS`.

### READINESS_DRAIN_SECONDS

Time to wait after `/readyz` starts returning `503` and before the public listener is shut down. This gives load balancers and orchestrators time to stop routing new requests to the instance.

### SHUTDOWN_HARD_SECONDS

Time to wait after the graceful shutdown timeout is reached so request contexts can observe cancellation before datastore resources are closed.

### Validation rules

Startup fails if the config is invalid. The implementation validates that:

- `SERVER_PORT` and `HEALTH_PORT` are different.
- `MAX_SECRET_BYTES`, `MAX_FILES`, and `MAX_FILE_SIZE_BYTES` are greater than `0`.
- `VALID_FOR_OPTIONS` is not empty and contains only unique positive values.
- `HSTS_MAX_AGE_SECONDS` is greater than `0` when `ENABLE_HSTS=true`.
- `GRACEFUL_SHUTDOWN_SECONDS`, `READINESS_DRAIN_SECONDS`, and `SHUTDOWN_HARD_SECONDS` are greater than `0`.
- `GRACEFUL_SHUTDOWN_SECONDS` is greater than `READINESS_DRAIN_SECONDS + SHUTDOWN_HARD_SECONDS`.

## Create a new secret

Content-only secrets can be created with JSON.

```bash
curl --request POST \
  --url http://localhost:8080/api \
  --header 'Content-Type: application/json' \
  --data '{
	"content": "some super secret stuff goes here",
	"expires_in": 3600
}'
```

Secrets with file attachments are created with `multipart/form-data`. The `data` field contains the same JSON payload as above, and each attached file is sent as a `files` field.

```bash
curl --request POST \
  --url http://localhost:8080/api \
  --form 'data={"content":"some super secret stuff goes here","expires_in":3600}' \
  --form 'files=@./example.txt'
```

Request fields:

- `content`: secret text content.
- `expires_in`: number of seconds until the secret expires. The value must be one of `VALID_FOR_OPTIONS`.
- `files`: optional map of file names to base64-encoded bytes in JSON requests, or one or more multipart `files` fields in multipart requests.

Either `content` or at least one file is required. `content` must be no larger than `MAX_SECRET_BYTES`. File count is limited by `MAX_FILES`, and each file is limited by `MAX_FILE_SIZE_BYTES`. Multipart requests also have an aggregate parsing limit of `MAX_FILES * MAX_FILE_SIZE_BYTES + MAX_SECRET_BYTES + 1MiB`.

The response status is `201 Created`. The response body is a token that can be used to fetch the secret.

## Fetch a secret

To fetch a secret, send a `GET` request to `http://localhost:8080/api/<token>`.

For example:

```bash
curl --request GET \
  --url http://localhost:8080/api/Jsm9nDvKVhtAQEfz1Bukx7jHeKIBpPV8kX0B_a4w2rE.qAke0MYJ_uvGc30s6o85TiIn-qeBm_9S55ajlDzysRw
```

Example response:

```json
{
  "content": "some super secret stuff goes here",
  "files": {
    "example.txt": "c29tZSBmaWxlIGNvbnRlbnQ="
  },
  "expires": "2026-06-26T12:00:00Z"
}
```

File values are base64-encoded in the JSON response.

If the token is invalid, missing, expired, already read, or cannot decrypt the secret, the API returns `410 Gone` with `secret not found`.

When a secret is fetched before it expires, it is deleted after the read. Expired secrets are deleted when accessed, and the datastore also removes expired secrets in the background.

## Health and metrics

- `GET /healthz` returns `200 OK` when the internal server is running.
- `GET /readyz` returns `200 OK` until shutdown starts, then returns `503 Service Unavailable` during the readiness drain period.
- `GET /metrics` exposes Prometheus metrics on the internal server.

Registered counters include:

- `secrets_read`
- `expired_secrets_read`
- `nonexistent_secrets_read`
- `secrets_created`
- `secrets_created_with_errors`
- `secrets_deleted`

## Security and deployment notes

- Keep `/healthz`, `/readyz`, and `/metrics` on a non-public network.
- Use `/healthz` for liveness probes and `/readyz` for readiness probes.
- Set the platform termination grace period higher than `GRACEFUL_SHUTDOWN_SECONDS`.
- Always run behind TLS (reverse proxy / ingress is supported).
- External routes set `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`, and `Content-Security-Policy` headers.
- API routes set `Cache-Control: no-store, max-age=0`, `Pragma: no-cache`, and `Expires: 0`.
- HSTS is only sent when `ENABLE_HSTS=true` and the request is TLS or has `X-Forwarded-Proto: https`.
- Rotate `SERVER_SALT` as part of incident response, but expect existing stored secrets to become unreadable after rotation.
- If a link is leaked, treat the secret as compromised and rotate underlying credentials.
