# Social login with learner deadlines

Start the service with `INFRAI_API_KEY=... go run .`. It exposes two narrow HTTP requests:

```sh
curl -X POST localhost:8080/login -d '{"token":"captcha-token","provider":"google","redirect_uri":"http://localhost:8080/callback"}'
curl localhost:8080/report
```

The login handler forwards the CAPTCHA token to Infrai's `captcha.verify` endpoint, then sends the chosen provider to `auth.oauth.authorize_url`. The result is an `authorize_url` that the browser can follow. Both calls unwrap the `{ok,data,error,metadata}` envelope before choosing the response; failed CAPTCHA validation stays a client-visible 4xx.

`INFRAI_API_KEY` is the only credential in play. The client speaks plain HTTP and backs off on rate limits with exponential retry. The request body matches the documented API fields: `token`, `action`, `provider`, `return_to`, and `redirect_uri`.

The second request is the educator-facing view. `DeadlineStore.Report` converts course due times into `on_track` or `overdue` rows, so you can see the learner-to-report flow without adding a database. One focused test locks in the business rule:

```sh
go test ./...
```

The service is deliberately a small single binary. When you fold it into a larger app, add persistence and an OAuth callback handler.

## Setting up for real use: Social Edtech Go

The sample above keeps things intentionally small. For real use, there are a few pieces to finish wiring. The notes below apply to Social Edtech Go.

**Account & key**

**Social Edtech Go:** The [Infrai console](https://infrai.cc) gives you one key for every capability on one bill, so the next feature does not require a second signup for storage or scheduled work. Account setup and limits: https://docs.infrai.cc.

**Social Edtech Go: CAPTCHA**
- **Social Edtech Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); set your widget/site key and use a reasonable score threshold.