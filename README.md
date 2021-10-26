# Pass-along

> :warning: Very much work in progress !


## TODO:
* Add some server config
* Security review?


The main application uses port `8080`.

 `/healthz` and `/metrics` endpoints uses port `8888`.


## Server config

The following config can be set via environment variables
| Tables                        | Required | Default   |
| ----------------------------- | :------: | --------- |
| [SERVERSALT](#SERVERSALT)     |          |           |
| [DATABASETYPE](#DATABASETYPE) |          | in-memory |
| [REDISSERVER](#REDISSERVER)   |          | localhost |
| [REDISPORT](#REDISPORT)       |          | 6379      |


### SERVERSALT
For extra security you can add your own salt when encrypting the data.

### DATABASETYPE
Can either be `in-memory` or `redis`.

### REDISSERVER
Address to your redis server.

### REDISPORT
Used to specify the port your redis server is using.


## Create a new secret

```bash
curl --request POST \
  --url http://localhost:8080/ \
  --header 'Content-Type: application/json' \
  --data '{
	"content": "some super secret stuff goes here",
	"expires_in": 10
}'
```

`expires_in` is number of seconds until it expires.

The reponse will be the ID of your secret, which can be used to fetch it again.

## Fetch a secret

To fetch you secret again to a GET request to `http://localhost:8080/<your-secret-id-goes-here>`

For example:
```bash
curl --request GET \
  --url http://localhost:8080/Jsm9nDvKVhtAQEfz1Bukx7jHeKIBpPV8kX0B_a4w2rEqAke0MYJ_uvGc30s6o85TiIn-qeBm_9S55ajlDzysRw
```
