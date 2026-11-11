---
title: 'Global viper config is read and written concurrently (go test -race: 10 races on main)'
date: "2026-11-11"
time: "13:12:10"
tags:
    - race
    - config
    - tests
summary: viper is not goroutine-safe; the HA config API and tests call viper.Set while publish/send goroutines read config, so go test -race fails in goalfeed and web/api
type: task
status: todo
effort: M
---

Measured on fleet-node, go1.27.4, 2026-11-11, with go test -race ./...:

- main @ 8f67af23: 10 races. They are in package goalfeed (the known main_test viper.Set vs ticker race) and in web/api, where tests call viper.Set while the goroutines that setHomeAssistantConfig starts (go PublishBaselineForMonitoredTeams, refreshActiveGamesInternal) read config.
- nhl-team-stats @ 637d45d6: 22 races, of the same class. Each publish now does more config reads (broadcastDelay) and there are more NHL baseline entities, so the window is wider.
- targets/homeassistant is race-free on both, at 20 rounds each.

Production counterpart: POST /api/homeassistant/config calls viper.Set while live sends read home_assistant.url and access_token from other goroutines. Today this is a data race, not just a test artifact.

Fix: put config reads and writes behind a sync.RWMutex in the config package (add config.Set), and route the API handler, main.go and tests through it instead of calling viper directly. Then make go test -race ./... a CI job.
