# Pass-along

[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7427/badge)](https://bestpractices.coreinfrastructure.org/projects/7427)

The main application uses port `8080`.

 `/healthz` and `/metrics` endpoints uses port `8888`.


## Server config

The following config can be set via environment variables
| Tables                                  | Required | Default               |
| --------------------------------------- | :------: | --------------------- |
| [SERVER_SECRET](#SERVER_SECRET)         |    x     |                       |
| [DATABASE_TYPE](#DATABASE_TYPE)         |          | in-memory             |
| [REDIS_SERVER](#REDIS_SERVER)           |          | localhost             |
| [REDIS_PORT](#REDIS_PORT)               |          | 6379                  |
| [SERVER_PORT](#SERVER_PORT)             |          | 8080                  |
| [HEALTH_PORT](#HEALTH_PORT)             |          | 8888                  |
| [LOG_LEVEL](#LOG_LEVEL)                 |          | info                  |
| [VALID_FOR_OPTIONS](#VALID_FOR_OPTIONS) |          | 3600,7200,43200,86400 |
| [KDF_ITERATIONS](#KDF_ITERATIONS)       |          | 600000                |
| [MAX_SECRET_BYTES](#MAX_SECRET_BYTES)   |          | 1048576               |
| [MAX_MULTIPART_BYTES](#MAX_MULTIPART_BYTES) |      | 10485760              |
| [MAX_FILES](#MAX_FILES)                 |          | 10                    |
| [MAX_FILE_SIZE_BYTES](#MAX_FILE_SIZE_BYTES) |      | 2097152               |
| [MAX_FILENAME_LENGTH](#MAX_FILENAME_LENGTH) |      | 255                   |
| [RATE_LIMIT_WINDOW_SECONDS](#RATE_LIMIT_WINDOW_SECONDS) |  | 60          |
| [MAX_REQUESTS_PER_WINDOW](#MAX_REQUESTS_PER_WINDOW) |      | 120          |
| [ENABLE_HSTS](#ENABLE_HSTS)             |          | false                 |
| [HSTS_MAX_AGE_SECONDS](#HSTS_MAX_AGE_SECONDS) |      | 31536000        |


### SERVER_SECRET
Required. Must be at least 32 characters and high entropy.
This value is used as server-side secret material for key derivation.

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
Used to specify loglevels, valid values are: `debug`, `info`, `warn` and `error`

### VALID_FOR_OPTIONS
Which options are available in the UI for secret expiration.
Only these values are accepted server-side for `expires_in`.

### KDF_ITERATIONS
PBKDF2 iteration count for encryption key derivation.

### MAX_SECRET_BYTES
Maximum size of secret text content.

### MAX_MULTIPART_BYTES
Maximum size of full multipart request body.

### MAX_FILES
Maximum number of attached files.

### MAX_FILE_SIZE_BYTES
Maximum size per file attachment.

### MAX_FILENAME_LENGTH
Maximum filename length for attached files.

### RATE_LIMIT_WINDOW_SECONDS
Window size for per-client API rate limits.

### MAX_REQUESTS_PER_WINDOW
Maximum API requests per client within each rate-limit window.

### ENABLE_HSTS
Enable HSTS response header. Keep disabled unless TLS is correctly terminated upstream.

### HSTS_MAX_AGE_SECONDS
HSTS `max-age` value when `ENABLE_HSTS=true`.

## Create a new secret

```bash
curl --request POST \
  --url http://localhost:8080/api \
  --header 'Content-Type: application/json' \
  --data '{
	"content": "some super secret stuff goes here",
	"expires_in": 10
}'
```

`expires_in` is number of seconds until it expires.

The response will be the ID of your secret, which can be used to fetch it again.

Note: this returned value is a token containing both lookup identifier and access key.

## Fetch a secret

To fetch you secret again to a GET request to `http://localhost:8080/api/<your-secret-id-goes-here>`

For example:
```bash
curl --request GET \
  --url http://localhost:8080/api/Jsm9nDvKVhtAQEfz1Bukx7jHeKIBpPV8kX0B_a4w2rEqAke0MYJ_uvGc30s6o85TiIn-qeBm_9S55ajlDzysRw
```

## Security and deployment notes

- Keep `/healthz` and `/metrics` on a non-public network.
- Always run behind TLS (reverse proxy / ingress is supported).
- Security headers and no-store cache controls are enabled by default.
- Rotate `SERVER_SECRET` as part of incident response.
- If a link is leaked, treat the secret as compromised and rotate underlying credentials.
