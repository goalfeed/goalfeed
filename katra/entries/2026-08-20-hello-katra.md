---
title: Hello, katra
date: "2026-08-20"
time: "16:27:39"
tags:
    - meta
summary: First entry — delete me.
---

Welcome to your dev log. **Write entries as you work** — this file is a *draft*
until you stamp it with a commit, so it shows up in the **In Progress** panel
right away.

## How it works

- `katra new "Title"` starts a draft (a markdown file in `entries/`).
- Write the *why*, not a paraphrased diff. Drop in screenshots and components.
- At commit time it gets stamped with the hash + diffstat and drops into the log.

## Rich components

Embed components by fencing them with a registered language:

```note
This is a callout. Markdown **works** inside it.
```

Add media with `katra capture <file>`, before/after sliders with
`katra compare <before> <after>`, and interactive HTML artifacts with an
`embed` block. Then run `katra serve` and watch it live-reload.
