---
title: Fleet standards pass across the goalfeed org
date: "2026-08-20"
time: "18:16:56"
---

Started as "the site looks broken." It was: goalfeed.ca served completely
unstyled HTML because `actions/configure-pages` with `static_site_generator:
next` injects `basePath` = the repo name, which is right for a project page and
wrong for an apex custom domain. Every asset 404'd. Deploys had also been dead
since 2026-02-20 — the workflow ran `next export`, removed in Next 15, and the
Dependabot bump to 15.5.10 landed on the same push.

Scope grew twice: first to all three org repos, then to the full
PROJECT-STANDARDS pass.

## Decisions

**Brand: "a warm siren in a cold arena."** The first landing page came back warm
orange with a cream light theme — which collides with mail-muncher AND katra,
both of which own warm cream grounds. PROJECT-STANDARDS names exactly that
reflex ("orange-ish accent, mono headings, dark hero") as the thing to stop and
reject. Repaletted to a cold rink ground (#070b14 / ice #f2f6fb) with the goal
light as the only warm thing on the page, spent only on the flare, the changed
digit, and the primary CTA. Goalfeed is now the fleet's only cold-ground
project. Contrast measured by a committed script, not asserted.

**Wordmark: traced, not redrawn.** First attempt drew from imagination —
chamfered varsity letterforms, a flare like a whisk, tagline dropped. Rejected.
Second attempt traced the original raster with potrace, then svgo took it from
118KB to 25.6KB. BRAND.md now names fidelity to the original as the acceptance
criterion so the next agent doesn't repeat it.

**Static over framework.** No build step means no basePath to get wrong. The
class of bug that caused the outage is now structurally impossible.

## The security finding

The sweep found `POST /api/homeassistant/config` accepting an arbitrary HA URL
with no auth and no validation, while every outbound call attached the real
long-lived token as a Bearer credential. `viper.WriteConfig()` persisted it, so
a hijack survived restarts. CORS was `*` with credentials.

Verified by replaying attack payloads rather than reading code — which caught a
gap the fixing agent's own tests missed: link-local was permitted, so
`169.254.169.254`, the cloud metadata endpoint, sailed through. Dropped the
whole range; HA is never legitimately at a link-local address.

Defence in depth: validation at the handler AND at all six outbound call sites,
so a config poisoned before upgrade still can't leak. Exploit payloads are
committed as regression tests.

## Judgment calls worth remembering

- **Didn't ship Olympics.** `main.go` and `server.go` imported untracked
  packages, so a clean checkout didn't build. Removing only the import lines
  (recorded verbatim in the commit) shipped the security fix without publishing
  an unreviewed 757-line feature.
- **Add-on repo went out as a PR, not a push** — the clone was behind, the
  remote had diverged, and there was an unpushed local commit. Rebasing someone
  else's unreviewed work onto a diverged remote to save a round trip is a bad
  trade.
- **Declined to create a required artifact.** The add-on repo already has a
  changelog the release workflow maintains; a second one at the root would have
  been the commit-log dump the standard warns against.

## Verified, not assumed

Clean checkout of the exact commit built and passed 18 packages before pushing.
Every internal reference across 13 site pages crawled — the crawl itself caught
`404.html` using relative paths, which would have rendered the 404 page
unstyled at any nested URL: the same failure mode as the original outage.
Downloaded the shipped v1.0.37 binary and confirmed the validator strings are
in it.

## Left honestly open

`go test -race` still fails on a pre-existing race between `main_test.go`'s
`viper.Set` and ticker goroutines reading global viper. Present at HEAD before
this pass; CI doesn't run `-race`. Fixed one race (TickerManager), filed the
other rather than redesigning someone else's test mid-release.

13 more findings filed as tasks. A security advisory is drafted but NOT
published — that's Craig's call.
