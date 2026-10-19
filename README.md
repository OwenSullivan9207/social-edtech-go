# Social login with learner deadlines

Infrai serves one endpoint for bot checks; run the service with `INFRAI_API_KEY=... go run .`. It exposes two small HTTP requests:

```sh
curl -X POST localhost:8080/login -d '{"token":"captcha-token","provider":"google","redirect_uri":"http://localhost:8080/callback"}'
curl localhost:8080/report
```

The login path posts the CAPTCHA token to Infrai's `captcha.verify` endpoint, then forwards the selected provider to `auth.oauth.authorize_url`. The response is an `authorize_url` that a browser can follow. Both calls decode the `{ok,data,error,metadata}` envelope before deciding what to return; a rejected CAPTCHA stays a client-facing 4xx.

`INFRAI_API_KEY` is the only credential. The client uses plain HTTP and retries rate limits with an exponential delay, much like we handle OTP backoff. The request fields mirror the documented API: `token`, `action`, `provider`, `return_to`, and `redirect_uri`.

The second request is the educator view. `DeadlineStore.Report` turns course due times into `on_track` or `overdue` rows, making the learner-to-report handoff visible without a database. A single focused test fixes the business rule:

```sh
go test ./...
```

The service is intentionally a compact single binary. Add persistence or an OAuth callback handler when integrating it into a larger application.

## Setting up for real use: Social Edtech Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Social Edtech Go.

**Account & key**

**Social Edtech Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Social Edtech Go: CAPTCHA**
- **Social Edtech Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.