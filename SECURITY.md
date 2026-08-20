# Security Policy

## Supported versions

Goalfeed doesn't maintain parallel release branches — there's one supported
version: the [latest release](https://github.com/goalfeed/goalfeed/releases/latest).
If you're running an older binary or an older build of the
[Home Assistant add-on](https://github.com/goalfeed/hassio-goalfeed-repository),
please update before reporting, if you're able to — it's likely the fastest
path to a fix either way.

## Reporting a vulnerability

Please **do not** open a public GitHub issue for a security vulnerability.

Instead, use GitHub's private vulnerability reporting on the relevant repo:

1. Go to the repo's **Security** tab.
2. Click **Report a vulnerability**.
3. Describe the issue, how to reproduce it, and its impact.

- Vulnerabilities in the Go service, the REST/WebSocket API, or the web UI:
  [github.com/goalfeed/goalfeed/security/advisories/new](https://github.com/goalfeed/goalfeed/security/advisories/new)
- Vulnerabilities in the Home Assistant add-on packaging (Docker image,
  add-on manifest, `run.sh`):
  [github.com/goalfeed/hassio-goalfeed-repository/security/advisories/new](https://github.com/goalfeed/hassio-goalfeed-repository/security/advisories/new)

If you're not sure which repo applies, report it against `goalfeed/goalfeed` —
it's the actively maintained one and easy to redirect from there.

This project doesn't publish a contact email for security reports. GitHub's
private advisory flow is the supported channel; it reaches the maintainer
directly and lets us coordinate a fix and disclosure privately before anything
is public.

## Response time

This is a small, largely solo-maintained project, not a company with an SLA.
Expect an initial response within about a week. If you haven't heard back in
two weeks, it's fine to follow up on the same advisory thread.

## Scope

In scope: the Go service in this repo, its REST/WebSocket API, the bundled
React web UI, and the Home Assistant add-on packaging in the companion
`hassio-goalfeed-repository`.

Out of scope: the upstream sports-league APIs Goalfeed polls
(`api-web.nhle.com`, `statsapi.mlb.com`, ESPN's site API, `cflscoreboard.cfl.ca`,
`realtime.iihf.com`, `olympics.com`, and a third-party BetGenius widget feed
used for CFL play detail). Goalfeed doesn't operate any of these; a
vulnerability in one of them isn't a Goalfeed security issue, though a report
of Goalfeed *misusing* one (leaking a credential to it, say) is.

## Known posture — please read before reporting these as new findings

A few things are intentional design decisions for a self-hosted, home-network
tool, not oversights:

- **The REST/WebSocket API has no authentication.** Anyone who can reach
  `--web-port` (default `8080`) can read game state and change which teams
  are watched.
- **The WebSocket endpoint (`/ws`) accepts any origin** and has no rate
  limiting.

Goalfeed is built to run on a trusted home network — typically behind Home
Assistant's ingress (which handles auth itself) or on a LAN you control — not
to be exposed directly to the public internet. If you expose it to the
internet anyway, put it behind your own authenticating reverse proxy; don't
rely on Goalfeed to gate access. We do still want to hear about anything
*beyond* this known posture — e.g. a way to reach another user's Home
Assistant instance, execute code, or exfiltrate the configured
`home_assistant.access_token` — those are genuine vulnerabilities and should
be reported as above.

## Home Assistant access token protections

`POST /api/homeassistant/config` used to accept any `url` with no validation
and always persisted it to `config.yaml` via `viper.WriteConfig()`. Combined
with wide-open CORS (`Access-Control-Allow-Origin: *` plus credentials), that
meant any device on the LAN, or a malicious web page loaded in the same
browser used to manage Goalfeed, could point `home_assistant.url` at an
attacker-controlled host and have Goalfeed's long-lived Home Assistant access
token — which typically grants full control of the user's home — attached to
outbound requests and sent straight to it. This has been fixed:

- **URL validation, enforced twice.** Every `home_assistant.url` — whether set
  via this API, an environment variable, or a hand-edited `config.yaml` — is
  validated before use: it must be an `http`/`https` URL with a non-empty
  host that resolves to a private/loopback/link-local address, or a hostname
  that's clearly local (`*.local`, `homeassistant`, `homeassistant.local`,
  `supervisor`, or a bare single-label hostname). `POST
  /api/homeassistant/config` rejects an invalid URL with `400` before writing
  anything, and — separately, so a config file poisoned some other way is
  still covered — every outbound call to Home Assistant re-validates the
  configured URL immediately before attaching the access token, refusing to
  send if it fails. A genuinely remote HA instance is supported via the
  opt-in `home_assistant.allow_remote_url` config key (default `false`),
  which relaxes the private-address requirement but never the scheme/host
  checks.
- **CORS is no longer wide open.** `Access-Control-Allow-Origin: *` combined
  with credentials — which let any page read this API's responses, including
  the HA config endpoints — has been replaced with an allowlist of local
  origins only (`localhost`/`127.0.0.1`/`[::1]`, any port, covering the React
  dev server and direct local access), and credentials are no longer sent.
- **Config writes from the API are opt-in.** `viper.WriteConfig()` — which
  persisted this unauthenticated endpoint's input to disk across restarts —
  now only runs when the operator has set `web.allow_config_writes: true`.
  By default, a `POST` to this endpoint updates the running process's
  in-memory configuration for that session only.
