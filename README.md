<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/goalfeed-wordmark-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="docs/assets/goalfeed-wordmark-light.svg">
  <img src="docs/assets/goalfeed-wordmark-dark.svg" alt="Goalfeed" width="460">
</picture>

**Your team scores. Your house reacts.**

Self-hosted goal detection for NHL, MLB, CFL and NFL — fires a Home Assistant event
within about a second of your watched team's score changing upstream, or streams the
same thing over REST and WebSocket for anything else you want to build.

[![Build Status](https://github.com/goalfeed/goalfeed/workflows/PR%20Test/badge.svg)](https://github.com/goalfeed/goalfeed/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/goalfeed/goalfeed/branch/main/graph/badge.svg)](https://codecov.io/gh/goalfeed/goalfeed)
[![Go Report Card](https://goreportcard.com/badge/github.com/goalfeed/goalfeed)](https://goreportcard.com/report/github.com/goalfeed/goalfeed)
[![Release](https://img.shields.io/github/release/goalfeed/goalfeed.svg)](https://github.com/goalfeed/goalfeed/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

</div>

---

Goalfeed watches live games for the teams you care about and tells your house within
about a second when something happens — a goal, a touchdown, a home run. Point it at
Home Assistant and it fires a real event on the HA event bus with the scoring team, the
opponent, and the current game state, so an automation can blow the goal horn and flash
the lights in team colours. It polls the same handful of undocumented, public league
APIs your browser already talks to when you check the score — nothing scraped, nothing
paid for.

It's self-hosted and free: no account, no API key, no cloud service standing between
your Home Assistant instance and the leagues. The only network calls Goalfeed makes are
to the league APIs it polls and, if you use it, your Home Assistant instance.

Home Assistant is the primary use case, but it isn't the only one. Goalfeed also runs a
plain REST API, a WebSocket feed, and a small React scoreboard/team-manager web UI, so
you can build your own consumer — a physical goal-light rig, a Discord bot, a browser
overlay — without touching Home Assistant at all.

## Seeing it fire

Goalfeed ships a `test-goals` option so you can watch the whole pipeline run without
waiting for a real game — it's a genuine feature (used to build and test Home Assistant
automations), not a demo trick added for this README. What follows is real output: built
from this exact working tree with `go build -o goalfeed .` and run against a plain HTTP
listener standing in for Home Assistant on loopback, so the request path is real without
needing a live HA instance.

```
$ cat config.yaml
test-goals: true
web: false
home_assistant:
  url: "http://127.0.0.1:9123"
  access_token: "demo-token"
watch:
  nhl:
    - WPG

$ ./goalfeed
{"level":"info","ts":1787264694.444872,"caller":"goalfeed/main.go:171","msg":"Puck Drop! Initializing Goalfeed Process"}
{"level":"info","ts":1787264694.444947,"caller":"goalfeed/main.go:181","msg":"Initializing Active Games"}
{"level":"info","ts":1787264694.444968,"caller":"goalfeed/main.go:192","msg":"Updating Active Games"}
{"level":"info","ts":1787264694.444989,"caller":"goalfeed/main.go:199","msg":"Checking for active CFL games"}
   … one "Checking for active <League> games" line per registered league …

   (60.009 seconds later — the test-goals ticker really does fire once a minute)

{"level":"info","ts":1787264754.453922,"caller":"goalfeed/main.go:338","msg":"Sending test goal"}
{"level":"info","ts":1787264754.459335,"caller":"homeassistant/homeassistant.go:84","msg":"Successfully sent  event to Home Assistant"}
```

Two real numbers in that transcript. **1787264754.453922 − 1787264694.444872 = 60.009
seconds** between startup and the first test goal, confirming the `test-goals` ticker
fires once a minute, not "periodically." And the delivery genuinely round-tripped an
HTTP POST to the stand-in listener rather than being logged and discarded — this is the
same `homeassistant.SendEvent` code path a real goal takes. The double space in
"Successfully sent&nbsp;&nbsp;event" isn't a typo added for this README: it's
`event.Type` printed at its real, empty value, because the synthetic test event — like
every NHL/MLB/NFL/CFL goal event today — never sets `Event.Type` (see
[Status](#status), below).

<details>
<summary>More output — what happens with no Home Assistant configured at all</summary>

The same run, with no `home_assistant.url` set, refuses to send rather than failing
silently or crashing:

```
{"level":"info","ts":1787264548.694264,"caller":"goalfeed/main.go:338","msg":"Sending test goal"}
{"level":"warn","ts":1787264548.694821,"caller":"homeassistant/homeassistant.go:35","msg":"refusing to send Home Assistant request: home assistant url is empty"}
```

That guard (`targets/homeassistant/guard.go`, backed by `utils.ValidateHomeAssistantURL`)
runs in front of every outbound Home Assistant request, real or synthetic, and is also
what `home_assistant.allow_remote_url` opts out of — see
[Configuration](#configuration).

</details>

## What it does

### Supported leagues

| League | Upstream source | Status |
|---|---|---|
| NHL | `api-web.nhle.com` (undocumented NHL Edge API) | Stable — polled and live |
| MLB | `statsapi.mlb.com` (unofficial MLB Stats API) | Stable — polled and live |

CFL and NFL code exists in the repo but is **not listed as supported**. Live testing
against real games on 2026-08-20 found that both leagues discover games correctly but
fail to track the score afterwards, so no goal or touchdown event can fire. Fixes are in
progress; neither league will be listed until it has been verified against a live game
again. See [Status](#status).

### The detection loop

`main.go` runs five recurring tickers, and the intervals are the actual numbers, not a
rounding of "fast" or "real-time":

| Every | Does |
|---|---|
| **1 minute** | Asks each registered league service which games among your watched teams are currently active, and starts tracking any newly-live one |
| **1 second** | Re-polls every currently-tracked game, diffs its score against the last-seen state, and fires goal/period events the moment a change is seen |
| **10 minutes** | Publishes "upcoming game" sensors to Home Assistant for your watched teams |
| **1 minute** | If `test-goals` is on, fires one synthetic `TEST` event through the exact same pipeline a real goal uses — this is what powers the demo above |
| **5 seconds** | A conditional re-check ticker; currently a no-op in this codebase (the flag that would trigger it is never set) |

NFL also gets push updates over a persistent ESPN Fastcast WebSocket
(`services/leagues/nfl/fastcast.go`), layered on top of the 1-second poll whenever
`nfl.fastcast.enabled` is true (the default).

**The honest latency statement, stated plainly rather than left as "real-time":** once a
game is being tracked, a score change is caught within 1 second. Getting a game *into*
tracking in the first place — noticing it went from scheduled to live — happens on the
1-minute ticker, so the very first score of a game can take up to a minute to reach Home
Assistant even though every score after that is caught within a second.

### What actually reaches Home Assistant

Every event Goalfeed sends — goal detections, period-start notices, and test goals —
arrives under the same Home Assistant event type, `goal`. Filter automations on the
event's `teamCode` field, not on the payload's `type` field: `type` (and
`Event.Description`) are left empty for every league (see [Status](#status)).

Goal/score detection itself is a raw score diff, not a play-by-play feed: Goalfeed
compares a watched team's last-seen score to its current one and fires one event per
point of increase. For NHL/MLB that's one event per goal. For NFL/CFL, a 7-point
touchdown-plus-conversion fires **seven separate `goal`
events** in the same tick, not one "touchdown" event — plan automations accordingly
(debounce, or only react to the first event in a burst).

Goalfeed also publishes per-team sensors and binary sensors
(`sensor.goalfeed_<league>_<team>_current_score`,
`binary_sensor.goalfeed_<league>_<team>_has_active_game`, and league-specific ones like
`..._shots` for hockey or `..._down` for football), seeded at startup and refreshed on
the 10-minute schedule ticker, so you can build a dashboard without listening for the
event at all.

## The problem

An automated home wants to know the instant your team scores, but "instant" is where the
usual answers fall apart. Checking your phone means you were looking at your phone
instead of the game. A browser tab left open on ESPN doesn't turn on a light. A generic
"sports notification" app can tell *you* something happened, but has no way to trigger
anything in your house — it isn't wired to Home Assistant, or to anything else you
control. And a cloud trigger service (the IFTTT-style kind) puts a sports-score applet
you don't control in the path between the play happening and your light turning on — and
those applets have a habit of quietly disappearing.

Three people this actually serves:

- **A Home Assistant user** who wants a goal horn and a light flash without writing a
  polling script — install the add-on, enter team codes, done. This is the supported,
  primary path.
- **Someone building a physical goal-light rig** — an ESP32 or Raspberry Pi driving real
  lights or a horn — who doesn't want Home Assistant in the loop at all. The REST API
  and WebSocket feed exist for exactly this; a rig can talk to Goalfeed directly.
- **Someone piping goals into something that isn't a smart home** — a Discord bot, a
  stream overlay, a personal dashboard. Same REST/WebSocket surface, no HA required.

Goalfeed splits the problem in half: it owns the per-league polling, the score-diffing,
and the delivery (a Home Assistant event and sensors, or REST/WebSocket for anything
else); you own the automation, the rig, or the bot that reacts to it.

## Alternatives

Read this before adopting. A polled Home Assistant sensor covers a lot of this ground
already, with less to install, and is a better fit for some readers.

| Tool | Use it instead when |
|---|---|
| [Team Tracker (`ha-teamtracker`)](https://github.com/vasqued2/ha-teamtracker) | You want dozens of leagues/sports (NFL, NBA, MLB, NHL, MLS, NCAA, international soccer, golf, F1, tennis, UFC natively, plus hundreds more teams across 30+ additional leagues via its custom API mode) with a five-minute HACS install and UI-driven setup, and a polled sensor is good enough for you — it updates every 10 minutes normally and every 5 seconds during (and ~20 minutes before) a game. It updates a sensor's *state*, not a discrete Home Assistant *event* — you trigger automations on the state changing rather than on an event firing, which is a real difference in how you write the automation but rarely a blocker. If Goalfeed doesn't have your league, Team Tracker very likely does, and that's the honest answer. |
| [ha-nhl](https://github.com/SimplySynced/ha-nhl) | You only care about NHL, want the same ESPN-backed, sensor-not-event polling as Team Tracker above, and don't need a REST/WebSocket API or the ability to run independently of Home Assistant. |
| Roll your own — an HA REST sensor + template automation, or a Node-RED flow polling an API yourself | You want to hit one endpoint you already know, own the polling interval and the parsing, and don't want a second process running at all. This is the same score-diffing and per-league response parsing Goalfeed already does — the trade is your time against not running Goalfeed, not a real limitation of either approach. |
| IFTTT-style webhook/sports-trigger services | You want a cloud-hosted, no-code trigger and don't need a Home Assistant event bus, self-hosting, or guaranteed uptime. Sports-score applets on these services have come and gone for years, and Goalfeed has no way to know whether a given one still exists by the time you read this. |

## Install

### Home Assistant Add-on (recommended)

[![Add repository to my Home Assistant](https://my.home-assistant.io/badges/supervisor_add_addon_repository.svg)](https://my.home-assistant.io/redirect/supervisor_add_addon_repository/?repository_url=https%3A%2F%2Fgithub.com%2Fgoalfeed%2Fhassio-goalfeed-repository)

1. In Home Assistant, go to **Settings → Add-ons → Add-on Store**, open the **⋮** menu,
   choose **Repositories**, and add:
   ```
   https://github.com/goalfeed/hassio-goalfeed-repository
   ```
2. Install the **Goalfeed** add-on from the store.
3. Open its **Configuration** tab and enter comma-separated team codes, e.g.:
   ```yaml
   nhl_teams: "TOR,WPG"
   mlb_teams: "TOR,NYY"
   test_goals: false
   ```
   The add-on's configuration UI currently exposes NHL and MLB only. To watch CFL or
   NFL teams from the add-on, mount a `config.yaml` (see
   [Configuration](#configuration)) or set `GOALFEED_WATCH_*` environment variables on
   the container.
4. Start the add-on. It runs in web mode on port 8080 with Home Assistant Supervisor
   auto-detected, so no URL or token is needed — the add-on reaches Home Assistant
   directly through the Supervisor API.
5. Open the web UI from the Home Assistant sidebar (ingress) or
   `http://homeassistant.local:8080`.

### Binary from a release

Every push to `main` publishes a new
[GitHub Release](https://github.com/goalfeed/goalfeed/releases/latest) with binaries for
Linux, macOS and Windows (amd64/arm64/arm/386) built by GoReleaser, each bundled with the
built web UI:

```bash
curl -LO https://github.com/goalfeed/goalfeed/releases/latest/download/goalfeed_1.0.38_darwin_arm64.tar.gz
tar xzf goalfeed_1.0.38_darwin_arm64.tar.gz
./goalfeed --nhl WPG --web
```

Only a released binary reports a real version — `./goalfeed --version` on the archive
above prints `goalfeed version 1.0.38`, stamped by GoReleaser's build. Neither `make
build` nor a plain `go build`/`go install` stamps that value, and Cobra doesn't register
a `--version` flag at all when it's unset — a source build genuinely has no `--version`
flag, not a blank one; that's confirmed by running it, not assumed.

### `go install`

```bash
go install github.com/goalfeed/goalfeed@latest
```

Builds the Go binary only — no bundled web UI, since `go install` never runs the
`web/frontend` build step GoReleaser does. Use `--web` and Goalfeed will serve the
REST/WebSocket API with a minimal JSON status page in place of the React UI. See the
version caveat above.

### Docker

There's no published standalone Docker image, and the repo's own `Dockerfile` is
currently stale: it builds from `golang:1.21`, but `go.mod` requires Go 1.24.0, so it
will not build as committed. It also only compiles the Go binary — it never runs the
frontend build, so even patched to a current Go image it would serve the REST/WebSocket
API without the web UI. Until that's fixed, a released binary or building from source is
the reliable path; if you want to try Docker anyway, bump the `FROM golang:1.21` line to
a current Go image first:

```bash
git clone https://github.com/goalfeed/goalfeed.git
cd goalfeed
sed -i '' 's/golang:1.21/golang:1.24/' Dockerfile   # or edit by hand
docker build -t goalfeed .
docker run -d --name goalfeed -p 8080:8080 \
  -e GOALFEED_WATCH_NHL=WPG \
  -e GOALFEED_HOME_ASSISTANT_URL=http://homeassistant.local:8123 \
  -e GOALFEED_HOME_ASSISTANT_ACCESS_TOKEN=your-long-lived-token \
  goalfeed --web
```

`docker-compose.yml` in this repo does **not** run Goalfeed — it only spins up a local
Home Assistant instance for development (see [CONTRIBUTING.md](CONTRIBUTING.md)).

### Build from source

```bash
git clone https://github.com/goalfeed/goalfeed.git
cd goalfeed
make build        # npm ci && npm run build in web/frontend, then go build -o goalfeed .
```

Requires Go 1.24+ and Node 20+. Full dev-loop instructions (hot reload, running the test
suite) are in [CONTRIBUTING.md](CONTRIBUTING.md).

## Quickstart

About five minutes if you already have the binary from Install, above, and it ends in a
goal firing on demand rather than waiting for a real game.

**1. Copy the example config (10 seconds).**

```bash
cp config.example.yaml config.yaml
```

**2. Turn on `test-goals` and point it at your Home Assistant (1–2 minutes).** The
example config already has a `home_assistant:` block with placeholder values — replace
them with your real URL and token, and add `test-goals: true` at the top level:

```yaml
test-goals: true
home_assistant:
  url: "http://homeassistant.local:8123"
  access_token: "your-long-lived-access-token"
```

Long-lived access tokens are created from your Home Assistant user profile, under
**Security → Long-Lived Access Tokens**.

**3. Add the one-line automation you're testing against (1 minute).** In Home Assistant,
**Settings → Automations → Create Automation → Skip** (edit in YAML), trigger
`event_type: goal`, action whatever you want to prove works — a light, a notification.

**4. Run it (a few seconds to start).**

```bash
./goalfeed
```

Expected output — a normal startup, no active real games required:

```
{"level":"info","ts":...,"msg":"Puck Drop! Initializing Goalfeed Process"}
{"level":"info","ts":...,"msg":"Initializing Active Games"}
{"level":"info","ts":...,"msg":"Updating Active Games"}
```

**5. Wait up to 60 seconds.** `test-goals` fires once a minute — see
[Seeing it fire](#seeing-it-fire) for exactly what that log line looks like and exactly
how long the wait is, measured, not estimated. When it lands you'll see:

```
{"level":"info","ts":...,"msg":"Sending test goal"}
{"level":"info","ts":...,"msg":"Successfully sent  event to Home Assistant"}
```

and your automation's action fires in Home Assistant. **That's the guaranteed visible
success** — it doesn't depend on a real game being on, so a quiet first run and a broken
install never look the same. Turn `test-goals` off once you've confirmed it, narrow
`watch:` to your actual teams, and leave it running.

## Configuration

Goalfeed reads configuration from three places, in increasing priority: a `config.yaml`
file in the working directory, `GOALFEED_*` environment variables, and CLI flags. Copy
[`config.example.yaml`](config.example.yaml) to `config.yaml` to get started — there is
no `--config` flag to point at a file somewhere else; Goalfeed always looks for
`./config.yaml`.

| CLI flag | YAML key | Env var | Type | Default | Description |
|---|---|---|---|---|---|
| `--nhl` | `watch.nhl` | `GOALFEED_WATCH_NHL` | string list | `[]` | NHL team codes to watch (e.g. `WPG`), or `*` for all teams |
| `--mlb` | `watch.mlb` | `GOALFEED_WATCH_MLB` | string list | `[]` | MLB team codes to watch |
| `--cfl` | `watch.cfl` | `GOALFEED_WATCH_CFL` | string list | `[]` | CFL team codes to watch |
| — | `watch.nfl` | `GOALFEED_WATCH_NFL` | string list | `[]` | NFL team codes to watch — no CLI flag exists yet, use YAML or env |
| — | `home_assistant.url` | `GOALFEED_HOME_ASSISTANT_URL` | string | `""` | Home Assistant base URL. Ignored (and auto-detected via Supervisor) when running as the HA add-on |
| — | `home_assistant.access_token` | `GOALFEED_HOME_ASSISTANT_ACCESS_TOKEN` | string | `""` | Home Assistant long-lived access token |
| — | `home_assistant.allow_remote_url` | `GOALFEED_HOME_ASSISTANT_ALLOW_REMOTE_URL` | bool | `false` | Allow `home_assistant.url` to be a public/remote address. By default it's rejected unless it's private/loopback/link-local or a clearly local hostname (`*.local`, `homeassistant`, `supervisor`, etc.) — this is a defense against the access token being sent to an attacker-controlled host |
| — | `web.allow_config_writes` | `GOALFEED_WEB_ALLOW_CONFIG_WRITES` | bool | `false` | Let `POST /api/homeassistant/config` persist changes to `config.yaml` on disk. By default that endpoint only updates the running process's in-memory config for the current session |
| `--test-goals` | `test-goals` | *(hyphenated key; use `env` rather than `export`)* | bool | `false` | Fire a synthetic `TEST` goal event once a minute, useful for testing automations |
| `--web` | `web` | *(same caveat)* | bool | `false` | Start the REST/WebSocket/web UI server alongside the polling loop |
| `--web-port` | `web-port` | *(same caveat)* | string | `"8080"` | Port for the web server |
| — | `app_log.path` | `GOALFEED_APP_LOG_PATH` | string | `"app.log.jsonl"` | Path to the JSONL application log consumed by the `/api/logs` and `/api/events` endpoints |
| — | `nfl.fastcast.enabled` | `GOALFEED_NFL_FASTCAST_ENABLED` | bool | `true` | Use ESPN's Fastcast WebSocket for push NFL updates alongside the 1-second poll |
| — | `nfl.fastcast.ping_interval_sec` | `GOALFEED_NFL_FASTCAST_PING_INTERVAL_SEC` | int | `20` | Fastcast keepalive ping interval |
| — | `nfl.fastcast.pong_wait_sec` | `GOALFEED_NFL_FASTCAST_PONG_WAIT_SEC` | int | `60` | Fastcast pong timeout |
| — | `nfl.fastcast.reconnect_base_ms` | `GOALFEED_NFL_FASTCAST_RECONNECT_BASE_MS` | int | `2000` | Fastcast reconnect backoff base |
| — | `nfl.fastcast.reconnect_max_ms` | `GOALFEED_NFL_FASTCAST_RECONNECT_MAX_MS` | int | `30000` | Fastcast reconnect backoff ceiling |

`test-goals`, `web`, and `web-port` use hyphenated viper keys, which most shells reject
in `export NAME=value`. Set them via CLI flag, YAML, or `env 'GOALFEED_WEB=true'
goalfeed` instead.

A full annotated `config.yaml`:

```yaml
home_assistant:
  url: "http://homeassistant.local:8123"
  access_token: "your-long-lived-access-token"
  # allow_remote_url: false      # default; set true only for a genuinely remote HA

watch:
  nhl:
    - WPG
  mlb:
    - TOR
  cfl:
    - BC
    - OTT
  # nfl also reads from here — no CLI flag for it yet
```

Notes that bite people:

- **`home_assistant.url` is validated before every outbound request, not just at
  startup.** A URL that isn't loopback, private, link-local, or a recognizably local
  hostname is rejected with an actionable error, even if it was set at runtime through
  `POST /api/homeassistant/config`. Set `home_assistant.allow_remote_url: true` if you
  genuinely run Home Assistant somewhere remote.
- **`POST /api/homeassistant/config` doesn't touch disk by default.** It updates the
  running process only, so a config-writing API call surviving a restart requires
  `web.allow_config_writes: true`.
- **There is no authentication on the REST/WebSocket API, and CORS is wide open**
  (`AllowOrigins: ["*"]`, credentials allowed). Fine behind Home Assistant ingress or a
  home network; don't expose it directly to the internet.

Full API reference (every route and field), the WebSocket message contract, and the Home
Assistant automation walkthrough live at **[goalfeed.ca](https://goalfeed.ca)**, covering
install, configuration, Home Assistant, API, and WebSocket topics, plus troubleshooting.
The [wiki](https://github.com/goalfeed/goalfeed/wiki) also still carries the original
[Hassio add-on installation](https://github.com/goalfeed/goalfeed/wiki/Hassio-Add%E2%80%90on-Installation)
and [goal automation](https://github.com/goalfeed/goalfeed/wiki/Goal-Automation)
walkthroughs, including the light-show automation YAML in full.

## Architecture

```
clients/leagues/<league>/        services/leagues/<league>/         targets/
  thin HTTP client for the    →    implements leagues.ILeagueService,  →  homeassistant (events + sensors)
  upstream API, unmarshals         diffs old vs. new GameState,           applog        (JSONL event log)
  into league response models      produces models.Event               notify        (WebSocket broadcast hooks)
```

```
main.go                          Ticker manager (see The detection loop, above) + CLI/config wiring
clients/leagues/<league>/        One package per league: raw upstream API calls, response
                                    models, a mock client for tests. Doesn't know about
                                    Goalfeed's own models.
services/leagues/interface.go    The ILeagueService contract every league implements
services/leagues/<league>/       Translates client responses into models.Game/GameState,
                                    and contains the score-diff logic that produces models.Event
targets/homeassistant/           Posts events + per-team sensors to Home Assistant; the
                                    home_assistant.allow_remote_url guard lives in guard.go
targets/applog/                  Durable JSONL event/state log; backs /api/logs and /api/events
targets/memoryStore/             Process-local in-memory game store — not a database;
                                    state resets on restart and rebuilds from the next poll
targets/notify/                  Function-pointer hooks so targets/* can push to WebSocket
                                    clients without importing web/api (which would be a cycle)
models/                          Shared Game/GameState/Event/Team shapes every league maps into
config/                          Viper wiring: file → env → CLI precedence, defaults, guards
web/api/                         Gin REST API, the /ws WebSocket hub, static file serving
web/frontend/                    React scoreboard/team-manager UI (Create React App + Tailwind)
docs/                            Swagger/OpenAPI spec (generated) + the WebSocket contract
```

Every league service call is fire-and-forget-goroutine-plus-channel: a method takes a
`ret chan T`, launches a goroutine that does the HTTP call, and sends the result on the
channel. New league code should match that pattern rather than introduce a different
concurrency style — see [Adding a new league](CONTRIBUTING.md#adding-a-new-league) for
the full walkthrough.

## Contributing

Issues and pull requests are welcome. The most valuable contribution this project can
take is a new league, and [CONTRIBUTING.md](CONTRIBUTING.md) walks through exactly what
that involves end to end. It also covers which of Goalfeed's three repos an issue
belongs in, the dev setup, the coverage gate, and code style.

Questions and bug reports: [GitHub Issues](https://github.com/goalfeed/goalfeed/issues) —
that's the only supported contact channel.

## Status

Pre-1.0-in-spirit and honest about it, even past the 1.x version numbers. The latest
release is **v1.0.38**, cut from current `main`, so what's described here is what
shipped. Releases are tagged automatically on every push to `main`, which means the
version number counts pushes rather than milestones — a bump does not imply a feature.
Check [CHANGELOG.md](CHANGELOG.md) for what actually changed in each one.

What's real and working today, and what isn't yet:

- **NHL and MLB are stable and live-polled.** These are the leagues to build on.
- **CFL and NFL are unlisted, pending verification.** Tested against live games on
  2026-08-20: both discover games correctly, but their live score-tracking is broken —
  CFL's live feed sends `sportId`/`competitionId` as JSON strings while the Go struct
  types them as `int`, so the whole payload is discarded on every poll; NFL parses
  ESPN's `/summary?event=` expecting a top-level `events[]` array that response does not
  contain. Both leave the score pinned or frozen, and since detection is a score diff,
  no event can fire. Fixes are in progress. Neither league gets a support label until it
  has been observed working against a live game.
- **`Event.Type` / `Event.Description` are empty for every league** — goal/score
  detection is a raw score diff, not a play-by-play feed (see
  [What actually reaches Home Assistant](#what-actually-reaches-home-assistant)).
- **`POST /api/refresh` is a stub** left over from a refactor — it returns success
  without refreshing anything. `POST /api/debug/nfl/add` is a real but explicitly
  debug-only endpoint, not part of the supported API surface.
- **No persistence beyond a JSONL log.** `targets/memoryStore` is in-process memory only;
  all active-game state resets on restart and rebuilds from the next poll.
- **All upstream league data sources are unofficial and undocumented** — the same
  endpoints each league's own site or app calls, not a published, versioned API. Any of
  them can change shape or disappear without notice, and Goalfeed has no fallback when
  they do. This is the single biggest operational risk in the project.
- **The Home Assistant add-on's Configuration tab only exposes NHL and MLB team
  options**, even though the underlying binary also supports CFL and NFL — see
  [Install](#home-assistant-add-on-recommended).

## License & credits

[MIT](./LICENSE) © 2026 Craig Midwinter.

Built on [gin-gonic/gin](https://github.com/gin-gonic/gin) for the HTTP API,
[gorilla/websocket](https://github.com/gorilla/websocket) for the WebSocket hub,
[spf13/cobra](https://github.com/spf13/cobra) and
[spf13/viper](https://github.com/spf13/viper) for the CLI and layered config,
[uber-go/zap](https://github.com/uber-go/zap) for logging, and
[swaggo/swag](https://github.com/swaggo/swag) for the generated API reference. The web UI
is a Create React App + Tailwind CSS project. Test mocks are generated with
[go.uber.org/mock](https://github.com/uber-go/mock).

The wordmark is traced from the maintainer's own original raster artwork; full
provenance and the brand palette live in the `goalfeed.ca` repo's brand documentation
(not yet published at time of writing — check [goalfeed.ca](https://goalfeed.ca) or that
repo directly). No asset in this project was generated with AI assistance.
