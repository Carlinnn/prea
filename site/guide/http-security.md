# HTTP security

Browsers enforce rules that protect your users, and your API has to take part: say which websites may call it, refuse forged requests and send headers that close common holes. Preá runs these protections before any of your code.

::: info What you'll learn
- How to let your frontend call the API (CORS)
- How cross-site request forgery is blocked (CSRF)
- Which security headers every reply carries
- How to get the client's real IP behind a proxy
:::

## Where it lives

The protections are configured in code, in `preas/preas.go`, so they are the same in every environment:

```go
var Security = core.Security{
	CORS: &core.CORS{
		AllowOrigins: []string{"http://localhost:3000", "https://app.example.com"},
	},
	CSRF:    &core.CSRF{},
	Headers: core.DefaultHeaders,
}
```

They run first, before every middleware, in this order: headers, CORS, CSRF, then your own protections. A request they block never reaches your code.

## CORS: who may call the API

A browser only lets a page on `https://app.example.com` call an API on another address if the API allows it. List the addresses of your frontends in `AllowOrigins`:

```go
CORS: &core.CORS{
	AllowOrigins: []string{"https://app.example.com"},
},
```

Preá answers the browser's preflight requests (`OPTIONS`) and adds the `Access-Control-*` headers for the origins you allow.

| Field | Default |
|---|---|
| `AllowOrigins` | none. `"*"` allows any origin |
| `AllowOriginFunc` | a function for rules a list can't express, like any subdomain |
| `AllowMethods` | `GET, POST, PUT, PATCH, DELETE` |
| `AllowHeaders` | `Authorization, Content-Type` |
| `ExposeHeaders` | none |
| `AllowCredentials` | `false`. With `true`, the browser sends cookies |
| `MaxAge` | how long the browser may cache the preflight |

```go
CORS: &core.CORS{
	AllowOriginFunc: func(origin string) bool {
		return strings.HasSuffix(origin, ".example.com")
	},
},
```

`AllowCredentials: true` with the origin `"*"` stops the app at startup. Browsers forbid that combination.

Leaving `CORS` as `nil` adds no headers at all, and browsers block every call from another origin. Servers and tools like `curl` are not affected.

## CSRF: refuse forged requests

Cross-site request forgery is a page on another site making the user's browser send a request to your API, such as a `POST` that deletes something. `CSRF: &core.CSRF{}` blocks unsafe requests (`POST`, `PUT`, `PATCH`, `DELETE`) that a browser sends from another origin, with a `403`.

- **Your CORS origins are trusted.** A frontend you allowed in CORS isn't blocked.
- **Non-browser clients pass.** `curl`, mobile apps and other servers don't send the headers this check uses.

| Field | What it does |
|---|---|
| `TrustedOrigins` | more origins to trust, besides the CORS ones |
| `TrustedOriginFunc` | a function for more complex rules |
| `Bypass` | routes skipped by the check, like `"POST /webhooks/"` |
| `Deny` | your own reply when a request is blocked; it must call `c.Abort()` |

## Security headers

`Headers: core.DefaultHeaders` adds these to every reply:

| Header | Value |
|---|---|
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Content-Security-Policy` | `default-src 'none'; frame-ancestors 'none'` |
| `Cross-Origin-Opener-Policy` | `same-origin` |
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains` |

To change one, copy the map and edit the copy:

```go
headers := maps.Clone(core.DefaultHeaders)
headers["Content-Security-Policy"] = "default-src 'self'"

var Security = core.Security{Headers: headers}
```

## Behind a proxy

Behind a load balancer, every request seems to come from the balancer. To use the client's real address, from `X-Forwarded-For`, list the proxies you trust:

```yaml
# config/application-prod.yaml
server:
  trusted_proxies: ["10.0.0.0/8"]
```

`c.ClientIP()` then returns the real client address, and the request logs show it. Without trusted proxies, Preá trusts none, and a client can't fake its address with the header. `Security.TrustedProxies` in code takes precedence over the setting.

## Your own protections

`Security.Middlewares` runs your protections right after the built-in ones and before every other middleware. Rate limits and IP allowlists go here:

```go
var Security = core.Security{
	CSRF:        &core.CSRF{},
	Headers:     core.DefaultHeaders,
	Middlewares: []any{middleware.RateLimit{}},
}
```

## A different setup for some routes

Each piece also works on its own, for a single preá or route:

```go
publicCORS, err := core.CORS{AllowOrigins: []string{"*"}}.Middleware()
noFrames := core.Headers(map[string]string{"X-Frame-Options": "DENY"})
```

Put them in a module's `Middlewares` or before a route's handler, like any [middleware](./middlewares).

## Next steps

- [Tokens & passwords](./tokens-and-passwords)
- [Deploying](./deploying)
