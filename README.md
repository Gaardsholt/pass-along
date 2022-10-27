# Pass-along

The main application uses port `8080`.

 `/healthz` and `/metrics` endpoints uses port `8888`.


## Server config

The following config can be set via environment variables
| Tables                          | Required | Default   |
| ------------------------------- | :------: | --------- |
| [SERVER_SALT](#SERVER_SALT)     |          |           |
| [DATABASE_TYPE](#DATABASE_TYPE) |          | in-memory |
| [REDIS_SERVER](#REDIS_SERVER)   |          | localhost |
| [REDIS_PORT](#REDIS_PORT)       |          | 6379      |
| [SERVER_PORT](#SERVER_PORT)     |          | 8080      |
| [HEALTH_PORT](#HEALTH_PORT)     |          | 8888      |
| [LOG_LEVEL](#LOG_LEVEL)         |          | info      |


### SERVER_SALT
For extra security you can add your own salt when encrypting the data.

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

The reponse will be the ID of your secret, which can be used to fetch it again.

## Fetch a secret

To fetch you secret again to a GET request to `http://localhost:8080/api/<your-secret-id-goes-here>`

For example:
```bash
curl --request GET \
  --url http://localhost:8080/api/Jsm9nDvKVhtAQEfz1Bukx7jHeKIBpPV8kX0B_a4w2rEqAke0MYJ_uvGc30s6o85TiIn-qeBm_9S55ajlDzysRw
```


