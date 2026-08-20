## What changed

<!-- Summarize the change. If this adds/modifies a league integration, name it. -->

## Why

<!-- What problem does this solve, or what bug does it fix? Link an issue if there is one. -->

## How was this tested?

<!-- go test ./..., npm test in web/frontend, manual testing against a live game or a fixture, etc. -->

## Checklist

- [ ] `go test ./...` passes locally, and `gofmt -l .` prints nothing
- [ ] `cd web/frontend && npm test` passes locally (if frontend code changed)
- [ ] Added/updated tests for any new documented behavior (the coverage-check CI gate
      compares against `main`)
- [ ] **No new API endpoint (or existing one) accepts a URL, hostname, or filesystem path
      from a request without validating it.** This is the exact shape of the
      `POST /api/homeassistant/config` token-exfiltration bug fixed in this project: an
      unvalidated URL let an attacker point outbound Home Assistant calls — and the real
      access token — at a host they controlled. Any code that builds a URL/path from
      user input calls `validateOutboundHAURL` (or an equivalent, justified check), not
      just at the handler but at the actual outbound call site.
- [ ] No credentials, Home Assistant access tokens, `SUPERVISOR_TOKEN` values, or real
      `config.yaml` contents appear in the diff, a test fixture, or a log line.
- [ ] If this touches a league integration, `CONTRIBUTING.md`'s "what done looks like"
      applies — I've said in this description whether it's fully wired (live polling +
      goal detection) or partial (e.g. schedule/history lookups only), and any league
      named as supported in `README.md`/docs is actually registered in `main.go`'s
      polling loop (`initialize()`/`leagueServices`), not just present as a client/service
      package.
- [ ] If this changes Swagger annotations in `web/api/server.go`, I know `docs.yml` will
      open a follow-up PR to regenerate the spec — no action needed from me
