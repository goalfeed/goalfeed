---
title: 'Flaky: services/leagues/nhl TestGetEvents fails ~1 in 3 under go test -shuffle=on'
date: "2026-11-11"
time: "12:42:52"
tags:
    - tests
    - flaky
    - nhl
summary: MockNHLApiClient keeps home/away scores in package-level vars (mock_client.go), so test order changes TestGetEvents' expected score diff
type: task
status: todo
effort: S
---

Found while reviewing nhl-team-stats (Daisy, 2026-11-11); reproduced on main @ 8f67af23: 4 of 12 shuffled runs fail at nhl_test.go:27-29.

Cause: `clients/leagues/nhl/mock_client.go` stores scores in package-level `homeScore`/`awayScore`, set by value-receiver `SetHomeScore`/`SetAwayScore`. Any earlier test that sets them changes what TestGetEvents sees.

Fix: move the scores onto the mock struct (pointer receiver) or reset them in each test. CI does not shuffle, so it passes there today.
