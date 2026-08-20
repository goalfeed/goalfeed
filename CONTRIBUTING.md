# Contributing to Goalfeed

Thanks for looking at the code. This is a small, mostly-solo project — a Go
service, a bundled React UI, and a set of per-league integrations against
unofficial sports APIs. This doc gets a dev environment running and explains
the one thing most contributors will actually want to do: add a league.

## Which repo does this belong in?

Goalfeed is three repos. Filing an issue or PR in the right one saves everyone
time:

- **This repo, [`goalfeed/goalfeed`](https://github.com/goalfeed/goalfeed)** —
  the Go service: league polling, goal detection, the Home Assistant
  integration, the REST/WebSocket API, and the React web UI in `web/frontend`.
  Bugs in goal detection, the API, the config system, or the web UI belong
  here.
- **[`goalfeed/hassio-goalfeed-repository`](https://github.com/goalfeed/hassio-goalfeed-repository)** —
  the Home Assistant Supervisor add-on packaging (Docker image, `config.json`,
  `run.sh`). Issues about installing the add-on, the add-on's Configuration
  tab, or add-on versioning/updates belong there, not here.
- **[`goalfeed/goalfeed.ca`](https://github.com/goalfeed/goalfeed.ca)** — the
  documentation/marketing site. Typos, broken links, or missing docs on
  [goalfeed.ca](https://goalfeed.ca) belong there.

If you're not sure, open it here — it's the most actively maintained of the
three and it's easy to redirect from here.

## Development setup

Requires **Go 1.24+** (`go.mod` pins `go 1.24.0`) and **Node 20+** (the CI
frontend job uses Node 20; see `.github/workflows/frontend-ci.yml`).

```bash
git clone https://github.com/goalfeed/goalfeed.git
cd goalfeed
cp config.example.yaml config.yaml   # edit in your HA URL/token and teams to watch
make build                            # builds web/frontend (npm ci && npm run build), then the Go binary
./goalfeed --web
```

`make` targets, from the `Makefile`:

| Target | Does |
|---|---|
| `make build` | `npm ci && npm run build` in `web/frontend`, then `go build -o goalfeed .` |
| `make frontend` | Frontend only (`npm ci && npm run build`) |
| `make backend` | Go binary only (`GO111MODULE=on go build -o goalfeed .`) |
| `make dev` | `bash dev.sh dev` — fswatch-based hot reload of both frontend and backend, see `DEV.md` |
| `make clean` | Removes the `goalfeed` binary and `web/frontend/build` |

For iterative work, `./dev.sh dev` (documented in `DEV.md`) watches
`web/frontend/src/**` and any `.go` file, rebuilds whichever side changed, and
restarts the server at `http://localhost:8080`. It needs `fswatch`
(`brew install fswatch` / `apt-get install fswatch`); without it, use
`./dev.sh rebuild` manually after each change.

`docker-compose.yml` in this repo does **not** run Goalfeed — it only starts a
local Home Assistant container (`ghcr.io/home-assistant/home-assistant:stable`
on port 8123) so you have something to point Goalfeed's `home_assistant.url`
at while developing. On macOS/Windows, point that HA instance at
`http://host.docker.internal:8080` to reach your locally-running Goalfeed.

### Running tests

```bash
go test ./...                 # Go test suite
cd web/frontend && npm test   # React test suite
```

### The coverage gate

`.github/workflows/coverage-check.yml` runs on every PR. It checks out
`origin/main`, runs the Go test suite there to get a baseline coverage
percentage (over `./models`, `./targets/homeassistant`, `./targets/memoryStore`,
the `nhl`/`mlb`/`iihf` league services, `./utils`, `./config`, `./clients/...`,
plus a named set of `main.go` tests), then checks out your PR commit and runs
the same suite. The PR fails if coverage drops by more than **1 percentage
point** from that baseline. In practice: if you add code, add tests for it: a
big new untested file (a new league client, say) can drop the aggregate
enough to fail the gate even if nothing existing broke.

Other CI you need to pass: `test.yml` (`go build ./...`, then Go tests, then a
Node 20 frontend build/test pass), `frontend-ci.yml` (React tests + build).
`docs.yml` regenerates the Swagger spec from source annotations and opens its
own PR if it's out of date — you don't need to run `swag` yourself, but if
you change any `// @...` annotation in `web/api/server.go` expect a follow-up
docs PR from the bot.

## Architecture: clients → services → targets

```
clients/leagues/<league>/        services/leagues/<league>/         targets/
  thin HTTP client for the    →    implements leagues.ILeagueService,  →  homeassistant (events + sensors)
  upstream API, unmarshals         diffs old vs. new GameState,           applog        (JSONL event log)
  into league response models      produces models.Event               notify        (WebSocket broadcast hooks)
```

- **`clients/leagues/<league>/`** — one package per league. A `<league>_client.go`
  with an `interface.go` (`I<League>ApiClient`) for testability, response
  models (`response_models.go`, `schedule_response_models.go`), and a
  `mock_client.go`/`mock_responses.go` pair generated with `go.uber.org/mock`
  (see `go generate ./...` in `.goreleaser.yml`'s before-hooks). Clients call
  the upstream API and unmarshal the response — they don't know anything
  about Goalfeed's internal `models.Game`/`models.Event` shapes.
- **`services/leagues/<league>/`** — implements `leagues.ILeagueService`
  (below), translating client responses into `models.Game`/`models.GameState`,
  and containing the score-diff logic that turns a state change into
  `models.Event`s.
- **`targets/`** — output sinks. `homeassistant` posts events/sensors to Home
  Assistant; `applog` appends a durable JSONL log and can broadcast to
  WebSocket clients; `memoryStore` is the process-local in-memory game store
  (not a database — state resets on restart); `notify` is a small set of
  function-pointer hooks that let `targets/*` push to WebSocket clients
  without `targets` importing `web/api` (which would be an import cycle).

Every league service call is fire-and-forget-goroutine-plus-channel: a method
takes a `ret chan T`, launches a goroutine that does the HTTP call, and sends
the result on the channel. Clients follow the same pattern. Match it in new
code rather than introducing a different concurrency style.

## Adding a new league

This is the highest-value contribution this project can take, and the most
involved one. Concretely, adding league `X` means:

### 1. The client package (`clients/leagues/x/`)

- `interface.go` — define `IXApiClient` with whatever methods the upstream
  API needs (get schedule/games by date, get a single game's live state,
  list teams). Look at `clients/leagues/nhl/interface.go` or
  `clients/leagues/mlb/interface.go` for the shape other leagues use.
- `x_client.go` — the real implementation: build the URL, make the HTTP
  call, unmarshal JSON into response-model structs.
- `response_models.go` (and `schedule_response_models.go` if the
  schedule/live-game shapes differ) — structs matching the upstream API's
  actual JSON. Don't reuse `models.Game`/`models.GameState` here; keep the
  wire format separate from Goalfeed's internal model.
- `mock_client.go` / `mock_responses.go` — a mock implementing `IXApiClient`
  (generated via `go.uber.org/mock`, see an existing league's `//go:generate`
  directive) plus canned response fixtures, so `services/leagues/x` can be
  unit-tested without hitting the real upstream API.

Find the upstream API first. Every league Goalfeed supports today
(`api-web.nhle.com` for NHL, `statsapi.mlb.com` for MLB, ESPN's public site
API for NFL, CFL's own scoreboard JSON plus a third-party BetGenius widget
feed for play detail, `realtime.iihf.com` for IIHF) is an **unofficial,
undocumented API** — the same one the league's own website or app calls, read
by watching network requests, not a published/versioned public API. Expect to
do the same reconnaissance for a new league, and expect it to be able to
change shape without notice — that's a real, ongoing maintenance cost, not a
one-time integration cost.

### 2. The service package (`services/leagues/x/`)

Implement `leagues.ILeagueService` (`services/leagues/interface.go`):

```go
package leagues

import "goalfeed/models"

type ILeagueService interface {
	GetLeagueName() string
	GetActiveGames(ret chan []models.Game)
	GetUpcomingGames(ret chan []models.Game)
	GetGamesByDate(date string, ret chan []models.Game)
	GetGameUpdate(game models.Game, ret chan models.GameUpdate)
	GetEvents(update models.GameUpdate, ret chan []models.Event)
}
```

`GetEvents` is where goal detection actually happens. Look at
`services/leagues/nhl` or `services/leagues/mlb` for the baseline pattern
(diff `OldState.Home/Away.Score` against `NewState.Home/Away.Score`, emit one
`models.Event` per point of increase) — that's what NHL/MLB/NFL/CFL do today.
If you can do better than a raw score diff (the (untracked, in-progress)
Olympics service is the only one that currently sets `Event.Type` and
`Event.Description` from real play-by-play data instead of just diffing
scores), that's a genuine improvement over the existing pattern, not just
parity with it.

Write `services/leagues/x/x_test.go` against the mock client from step 1 —
this is what keeps the coverage gate happy and is the only way to test
score-diff logic without a live game.

### 3. Wiring it in

- `models/league.go` — add a `LeagueIdX` constant. Note that league ID `3`
  (EPL) is already reserved and unimplemented — don't reuse it; add a new ID.
- `main.go`'s `initialize()` — construct the client and service and add
  `leagueServices[models.LeagueIdX] = x.XService{Client: xClients.XApiClient{}}`
  to the map. **This is the step that actually makes the league live** — IIHF
  has a complete client and service today but is missing exactly this line,
  which is why it's queryable through the API but never fires a live Home
  Assistant event. Don't repeat that half-wired state for a new league unless
  you mean to.
- `web/api/server.go` — add the league to whatever switch statements dispatch
  by `leagueId` (`getGamesByDate`, `getUpcomingGames`, `getLeagues`/
  `updateLeagueConfig`, `getAllTeams`) so the REST API and the team-picker UI
  can see it.
- Config: add a `watch.x` YAML/env key (`GOALFEED_WATCH_X`) by following the
  existing `watch.nfl`/`watch.olympic_men` pattern in the config loader. A CLI
  flag is optional — NFL and Olympic hockey don't have one; NHL/MLB/CFL do.
- `web/frontend/src/components/EventFeed.tsx` — add an icon/color mapping for
  the new league if you want it to render distinctly in the event feed (this
  is purely cosmetic; EPL has one of these with zero backend behind it, which
  is a trap to avoid, not a pattern to copy).

### 4. What "done" looks like

A league is fully wired when: it's in `leagueServices` in `main.go` (live
polling + goal detection), the REST endpoints recognize its league ID, the
frontend can pick teams for it, and `go test ./services/leagues/x/...` passes
with real coverage of the score-diff/event logic. If you're only adding
schedule/history lookups without live polling, say so plainly in the PR
description — don't let it look more finished than it is (this is exactly the
IIHF situation today, and it should be called out, not hidden).

## Code style

- Standard `gofmt`/`go vet` — no separate linter config in this repo today.
  Run `gofmt -l .` before opening a PR; CI will build but won't currently
  fail a PR purely for formatting, so treat this as a courtesy, not a gate.
- Match the channel-based, fire-and-forget-goroutine pattern already used
  throughout `clients/` and `services/` (see above) rather than introducing
  a different concurrency idiom in one corner of the codebase.
- Keep upstream wire-format structs (`response_models.go`) separate from
  `models.*` — don't have a client package return `models.Game` directly.
- Frontend: TypeScript, functional components, Tailwind utility classes
  (`web/frontend/tailwind.config.js`). Follow the existing component
  structure in `web/frontend/src/components/` rather than introducing a new
  state-management library.

## PR process

1. Fork, branch, make your change.
2. `go test ./...` and, if you touched the frontend, `cd web/frontend && npm test`.
3. Open the PR against `main`. Describe *what* changed and *why*, and call out
   anything that's partially done (see "What done looks like" above) rather
   than letting CI passing imply full completeness.
4. CI must pass: `test.yml` (Go build + test, Node 20 frontend build/test),
   `frontend-ci.yml`, and `coverage-check.yml` (no more than a 1-point
   coverage regression against `main`). `docs.yml` will open its own
   follow-up PR if Swagger annotations changed — you don't need to do
   anything for that one.
5. Small, focused PRs review faster than large ones, especially for a
   solo-maintained project — if you're adding a whole league, it's fine (and
   expected) to be a large PR, but keep unrelated refactors out of it.

## Reporting bugs / requesting features

Use the issue forms in `.github/ISSUE_TEMPLATE/` rather than a blank issue —
they ask for the version, run mode (add-on/Docker/binary/source), and
league/team config up front, which is almost always what's needed to
reproduce a polling or event-firing bug. See [`SECURITY.md`](SECURITY.md)
instead of a public issue for anything that looks like a vulnerability.
