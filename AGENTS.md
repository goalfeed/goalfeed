# AGENTS.md

Working notes for coding agents. Human-facing docs are in [README.md](README.md)
and at [goalfeed.ca](https://goalfeed.ca) — this file is only the things an agent
gets wrong here.

Goalfeed is a Go service that polls public league APIs for games involving teams
you've named, diffs each game's score against the last-seen value, and fires an
event when it goes up. Its main consumer is Home Assistant; it also exposes a
REST API, a WebSocket feed, and a small React UI. The shape is
`clients/leagues/<league>` (upstream HTTP adapters) → `services/leagues/<league>`
(detection) → `targets/` (Home Assistant, notify, applog, memoryStore).

## Commands (green on a clean checkout)

| | |
|---|---|
| `go build ./...` | Build. Go 1.24+. |
| `go test ./...` | Full suite. Must pass before any PR. |
| `gofmt -l .` | Must print nothing. |
| `go vet ./...` | Must be clean. |
| `golangci-lint run` | Config in `.golangci.yml`. CI gates **new** issues only — there is a large pre-existing backlog, so don't treat a dirty full run as your regression. |
| `make build` | Binary + frontend. `make frontend` / `make backend` individually. |
| `cd web/frontend && npm ci && npm test` | Frontend suite. |

`go test -race ./...` currently fails on a pre-existing race between
`main_test.go`'s `viper.Set` and ticker goroutines reading global viper. It
predates current work and CI does not run `-race`. Don't treat it as yours;
don't "fix" it by weakening the test.

## Invariants a PR must not break

**The league adapter contract.** Every league service implements
`services/leagues/interface.go`:

```go
type ILeagueService interface {
	GetLeagueName() string
	GetActiveGames(ret chan []models.Game)
	GetUpcomingGames(ret chan []models.Game)
	GetGamesByDate(date string, ret chan []models.Game)
	GetGameUpdate(game models.Game, ret chan models.GameUpdate)
	GetEvents(update models.GameUpdate, ret chan []models.Event)
}
```

All of these return through channels, not return values. A new league needs a
client under `clients/leagues/` and a service under `services/leagues/`, and must
be registered in `main.go`'s `leagueServices` map — **a league that isn't in that
map never fires an event**, however complete its code looks. Two leagues have
shipped in exactly that half-wired state.

**Upstream JSON is not typed the way you'd assume.** These are undocumented
vendor APIs that change without notice and are inconsistent about types — the
same field arrives as `"17"` from one and `17` from another. Go's decoder
reports a type mismatch *after* decoding the rest, and code that discards the
whole response on error throws away a correct payload because of one unrelated
field. Prefer tolerant types (see `FlexInt` in `clients/leagues/cfl`), log
decode errors, and never silently return a zero struct.

**The Home Assistant event shape is a downstream contract.** Users' automations
key on it. Events POST to `<ha_url>/api/events/goal` (plus `/game_update` and
`/period_update`), carrying `models.GameEvent`. `Type` and `Description` are
empty for every currently-supported league — automations filter on `teamCode`
and `leagueId`. Changing field names or the endpoint breaks people's houses.

**Detection counts points, not plays.** It is a raw score diff, so a 7-point
score change emits seven separate events. Don't document or imply play-level
semantics that don't exist.

**Never attach the Home Assistant token to an unvalidated URL.** `utils/hurl.go`
validates the HA URL and `targets/homeassistant/guard.go` enforces it at every
outbound call site. This closed a real vulnerability where an unauthenticated
request could redirect the service and exfiltrate a long-lived token. Any new
code path that reaches Home Assistant must go through the guard. Link-local is
rejected deliberately — that range holds the cloud metadata endpoint.

**No team or league logos, wordmarks, or trademarks** anywhere in the repo, the
site, or generated assets. Team codes set in the display face are the visual
shorthand. See `branding/BRAND.md` in the goalfeed.ca repo.

## Security notes

- The web API has **no authentication** and is intended for a trusted home
  network behind Home Assistant ingress. Don't add anything that assumes
  otherwise, and don't widen CORS — it is deliberately localhost-only with
  credentials disabled.
- `web.allow_config_writes` defaults to `false`; the runtime API must not
  persist to disk unless the operator opts in.
- Never commit a real `config.yaml`, a Home Assistant long-lived token, or a
  captured log containing one. `config.yaml` is gitignored; keep it that way.

## Things that will surprise you

- `go.mod` declares `module goalfeed`, not the GitHub path, so
  `go install github.com/goalfeed/goalfeed@latest` cannot work. Renaming touches
  every import; it's tracked, not done.
- There is no `--config` flag. Config is `./config.yaml` in the working
  directory, or `GOALFEED_*` env vars, or flags — in that precedence order.
- CLI flags exist only for `--nhl`, `--mlb`, `--cfl`. NFL is config/env only.
- **Every push to `main` cuts a release** via `.github/workflows/goreleaser.yml`,
  patch-bumping the version. A docs-only commit burns a version number. Batch
  trivial changes.
- `clients/leagues/olympics` and `services/leagues/olympics` may exist locally
  but are deliberately untracked and unwired. Don't add imports for them.
