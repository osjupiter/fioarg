# fioarg

A static web page for building [fio](https://github.com/axboe/fio) benchmark commands by clicking through options.

**→ https://osjupiter.github.io/fioarg/**

## Features

- Covers the common options: `--rw`, `--bs`, `--size`, `--numjobs`, `--iodepth`, `--ioengine`, `--direct`, `--time_based` / `--runtime` / `--ramp_time`, `--filename` / `--directory`, `--rwmixread`, `--output-format`, `--group_reporting`, plus a free-form extras field
- One-click presets (4k random read/write, sequential throughput, mixed 70/30, QD1 latency)
- Live command preview with a copy button
- No build step, no dependencies — a single `index.html`

## Development

Open `index.html` in a browser. That's it.

> **Warning:** Pointing `--filename` at a raw block device (e.g. `/dev/sdX`) with a write workload destroys its data.
