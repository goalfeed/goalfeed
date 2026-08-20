# Changelog

All notable changes to Goalfeed, written for the person running it — what
changed for you, not what changed in the diff. Format follows [Keep a
Changelog](https://keepachangelog.com/en/1.1.0/); versions are tags on this
repo, released via GoReleaser.

Reconstructed from `git log` and GitHub Releases where no changelog existed
before this pass. Every tag from v1.0.21 onward gets its own itemised entry
below, verified against that tag's actual diff. Tags **v0.0.1 through
v1.0.20** (20 tags: v0.0.1–v0.0.15, v1.0.16–v1.0.20; spanning 2023-10-06 to
2023-10-14) predate this changelog
and are **not itemised individually** — they're summarised as one grouped
entry below rather than presented as thin, reconstructed one-liners that
would serve no one. For the full per-tag record from that window, see
[GitHub Releases](https://github.com/goalfeed/goalfeed/releases) or
`git log v0.0.1..v1.0.20`.

## [Unreleased]

### Security

- **Fixed an unauthenticated Home Assistant access-token exfiltration path.**
  `POST /api/homeassistant/config` accepted an arbitrary URL with no
  validation, and every subsequent outbound call to Home Assistant attached
  your real long-lived access token to whatever host was configured there.
  Combined with CORS previously set to `AllowOrigins: ["*"]` with credentials
  allowed, a malicious web page could have pointed your Goalfeed instance at
  an attacker-controlled host and read the token back.
  - The Home Assistant URL is now validated: loopback and private addresses
    only by default. Link-local addresses (`169.254.0.0/16`, `fe80::/10`) are
    rejected even though they're "local," because that range is where the
    cloud metadata endpoint (`169.254.169.254`) lives — the canonical SSRF
    target.
  - Validation runs at all six outbound call sites in
    `targets/homeassistant/` (not just the config handler), so nothing can
    reach an unvalidated URL by a different code path.
  - CORS is now localhost-only, with credentials disabled.
  - `viper.WriteConfig()` (persisting config changes to `config.yaml` on
    disk) is now gated behind an opt-in.
  - Two new config keys, both defaulting to the safe setting:
    `home_assistant.allow_remote_url` (default `false`) and
    `web.allow_config_writes` (default `false`).
  - This affected Docker and standalone binary deployments. The Home
    Assistant add-on was **not** affected — Supervisor sets the Home
    Assistant URL and token via environment variables that take precedence
    over the API-writable config path.

### Fixed

- A data race in `TickerManager.AddTicker`: concurrent `append` calls to an
  unsynchronized slice could corrupt the ticker list under concurrent
  startup. `StartAllTickers` now takes a lock and snapshots the slice before
  starting tickers.
- Removed a stray debug `fmt.Println` of the Home Assistant Supervisor API
  URL left in `main()`.

### Added

- `.golangci.yml` and a `golangci-lint` job in CI (`only-new-issues: true`,
  so existing debt doesn't block new PRs while new issues are still caught).
- Tests pinning behavior that was previously documented but unverified.

### Changed

- README rewritten to the fleet documentation standard.
- Documentation site launched at [goalfeed.ca](https://goalfeed.ca).
- Added an MIT `LICENSE`.

## [1.0.36] — 2026-01-02

### Fixed

- Goal event detection during game updates — a correctness fix in how score
  changes are turned into fired events as a game's state is refreshed.

## [1.0.35] — 2025-12-18

### Added

- Date-based game lookup (`GET /api/games/date/:date` and similar) across
  NHL, MLB, CFL, IIHF, and NFL, so the web UI (and API consumers) can browse
  games on a specific day rather than only "what's live right now."

## [1.0.34] — 2025-12-17

### Changed

- Release workflow: GoReleaser now updates `config.json` (the Home Assistant
  add-on manifest fields it's responsible for) as part of a release. No
  user-facing behavior change.

## [1.0.33] — 2025-12-17

### Changed

- Release workflow: improved version extraction and file-modification checks
  in the GoReleaser pipeline. No user-facing behavior change.

## [1.0.32] — 2025-12-17

### Added

- More test coverage: `TickerManager` refresh behavior, CFL client error
  handling, NHL game-state management. Hardening, not new features.

## [1.0.31] — 2025-12-15

### Fixed

- A GoReleaser release-hook bug (a `cd` in a hook wasn't taking effect
  because it wasn't run through a shell) that had been breaking releases.

## [1.0.30] — 2025-12-15

### Added

- **The bundled web UI.** A React frontend (`web/frontend`) is now built
  into the Goalfeed binary and served alongside the API: a live Scoreboard,
  per-league game detail views (hockey, football, baseball — including a
  base-runner diamond for MLB and drive/possession detail for football and
  CFL), a team manager for the watch list, a live event feed, a Home
  Assistant connection/settings panel, and a log viewer backed by the
  application's JSONL log.
- **NFL support**, including live situation detail (down/distance,
  possession) via ESPN's scoreboard feed.
- **CFL support**, including live drive/possession tracking and one fired
  event per score increment (rather than one event per scoring play type).
- Application logging: structured JSONL logs, exposed over the API and
  rendered in the new log viewer, so a fired (or not-fired) event has a
  visible trail.
- A WebSocket feed (`useWebSocket`) so the frontend updates live instead of
  polling.

This tag folded in a long-lived `ui` branch; the commit history around it is
denser than the feature list above because of that merge, not because this
was an unusually large single change.

## [1.0.29] — 2025-09-06

### Added

- Coverage badges and a coverage-regression CI gate (see `CONTRIBUTING.md`'s
  "coverage gate" section) — PRs now fail if test coverage drops more than 1
  point from `main`.

### Fixed

- A Home Assistant network-error bug surfaced while raising test coverage
  on the integration.

## [1.0.28] — 2025-09-06

### Fixed

- Watching `*` (all teams) in a league's watch list now actually monitors
  every team in that league, rather than silently matching nothing.

## [1.0.27] — 2025-05-12

### Added

- `copilot-instructions.md` and dependency updates. No user-facing behavior
  change.

## [1.0.26] — 2025-05-12

### Added

- Goal events for MLB and NHL now include the opponent team's details, not
  just the scoring team's — useful for automations that want to show
  "WPG vs. TOR" rather than just "WPG."

## [1.0.25] — 2025-05-12

### Added

- Opponent details extended to IIHF goal events, matching NHL/MLB from
  1.0.26.

## [1.0.24] — 2025-05-11

### Added

- Unit tests for league services and game-monitoring functionality.
  Hardening, not new features.

## [1.0.23] — 2024-01-02

### Changed

- README updates. No functional change.

## [1.0.22] — 2023-11-09

### Changed

- Adapted to a new, non-public NHL API after the previous one was retired —
  NHL goal detection kept working across the upstream change.

## [1.0.21] — 2023-10-18

### Changed

- Test-goal firing is now hidden behind an explicit flag, and the team is
  checked before sending an event (previously a test goal could fire
  regardless of configured watch list).

## [0.0.1] – [1.0.20] — 2023-10-06 to 2023-10-14

Everything from the first tagged release through v1.0.20 — 20 tags
(v0.0.1–v0.0.15, v1.0.16–v1.0.20) — grouped into one entry rather than
itemised. Real, verified changes from this window, in no particular order:

- Goal detection for NHL and MLB, delivered to Home Assistant as fired
  events, was the shipped feature set for the whole span.
- The delivery/storage layer changed from Pusher (push) + Redis (state) +
  a Postgres-backed schema (`migrate.up.sql`) to a direct Home Assistant
  target and an in-memory store — all three of the original pieces were
  gone by v1.0.17. If you're running something from this far back, the
  upgrade path is a fresh install, not an in-place migration.
- The release pipeline moved from a plain Docker-build workflow to
  GoReleaser, and v1.0.17–v1.0.20 specifically were release-tooling fixes
  only (auth, tagging, workflow structure) with no behavior change for a
  running instance.
- v0.0.15 and v1.0.16 are the same commit — a version-scheme rename, not a
  code change.

This entry intentionally does not itemise per-tag — commits from this
period include a long-lived branch that was folded back in later, making
precise per-tag attribution unreliable. See the links above for the
authoritative per-tag record.

[Unreleased]: https://github.com/goalfeed/goalfeed/compare/v1.0.36...HEAD
[1.0.36]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.36
[1.0.35]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.35
[1.0.34]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.34
[1.0.33]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.33
[1.0.32]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.32
[1.0.31]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.31
[1.0.30]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.30
[1.0.29]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.29
[1.0.28]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.28
[1.0.27]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.27
[1.0.26]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.26
[1.0.25]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.25
[1.0.24]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.24
[1.0.23]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.23
[1.0.22]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.22
[1.0.21]: https://github.com/goalfeed/goalfeed/releases/tag/v1.0.21
[0.0.1] – [1.0.20]: https://github.com/goalfeed/goalfeed/compare/v0.0.1...v1.0.20
