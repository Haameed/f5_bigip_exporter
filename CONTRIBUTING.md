# Contributing to BIG-IP Exporter

First off — **thank you** for taking the time to contribute! 🎉
This project thrives on community participation, and every contribution counts.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Adding a New Collector](#adding-a-new-collector)
- [Coding Guidelines](#coding-guidelines)
- [Commit & PR Guidelines](#commit--pr-guidelines)

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md).
By participating, you are expected to uphold it.

## How Can I Contribute?

- 🐛 **Report bugs** — open an [issue](https://github.com/Haameed/bigip_exporter/issues)
  with steps to reproduce, your BIG-IP version (TMOS), and exporter version.
- 💡 **Suggest features / new metrics** — describe the iControl REST endpoint and
  the metrics you'd like to see.
- 📖 **Improve documentation** — typos, clarifications, examples are all welcome.
- 🔧 **Submit code** — bug fixes, new collectors, tests.

## Development Setup

    git clone https://github.com/Haameed/bigip_exporter.git
    cd bigip_exporter
    make build
    make test

Run locally against a device:

    make run   # uses config-example.yml — edit it first

## Adding a New Collector

Collectors live in `pkg/probe/`. To add one:

1. Create `pkg/probe/<name>.go` with a function matching this signature:

       func GetMyProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
           // 1. Define prometheus.Desc for each metric
           // 2. Define the JSON structs for the iControl REST response
           // 3. c.Get("/mgmt/tm/...", &resp)
           // 4. Build and return []prometheus.Metric
       }

2. Register it in the `allProbes` slice in `pkg/probe/probe.go`:

       {"MyProbe", GetMyProbe},

The framework runs every collector concurrently and merges the results, so you
don't need to manage goroutines for the top level.

## Coding Guidelines

- Run `make check` (fmt + vet + test) before pushing.
- Follow the
  [Prometheus metric naming conventions](https://prometheus.io/docs/practices/naming/):
  - Use base units (bytes, seconds) where the source data allows.
  - Use the `_total` suffix for counters.
  - Provide clear, human-readable `HELP` text.
- Keep the `target` label on every metric.
- Prefer `slog` for structured logging in new code.

## Commit & PR Guidelines

- Write descriptive commit messages
  ([Conventional Commits](https://www.conventionalcommits.org/) preferred:
  `feat:`, `fix:`, `docs:`, `refactor:`...).
- One logical change per PR.
- Include/update tests where it makes sense.
- Make sure CI passes before requesting review.

Thanks again! ❤️
