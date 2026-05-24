# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-05-24

Initial release.

### Added

- **structure**: initial project scaffold for Go and Python services.
- **models**: domain models for `Service`, `Incident`, and `Alert`.
- **config**: YAML configuration loader and validator.
- **storage**: SQLite persistence layer with migrations.
- **checker**: HTTP health check engine with per-service goroutine scheduling.
- **anomaly**: rolling baseline anomaly detector with cooldown logic.
- **dispatcher**: multi-target alert dispatcher with HTTP bot integration and retry.
- **dashboard**: REST JSON API for service status, incidents, and check results.
- **entrypoint**: CLI entrypoint with graceful shutdown and structured logging.
- **bot-structure**: Python Discord bot project scaffold.
- **bot-server**: Flask HTTP server receiving and validating alerts from Go.
- **bot-embeds**: Discord embed builders for all alert types.
- **bot-commands**: slash commands `/status`, `/history`, `/mute`, `/unmute`, `/force`.
- **bot-core**: `PulseBot` core with incident thread creation and mute-aware dispatch.
- **docker**: multi-stage Dockerfiles and `docker-compose` for full stack deployment.
- **ci**: GitHub Actions pipeline with test, build, and lint jobs.

[0.1.0]: https://github.com/joaopssx/pulse/compare