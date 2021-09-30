# share-a-password

> :warning: Very much work in progress !


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
  --url http://localhost:8080/<your-secret-id-goes-here>
```

