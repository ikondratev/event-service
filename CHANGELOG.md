# Changelog

## [Unreleased]
### Changed

## [0.6.0] - 2026-08-15
### Added

- Add Idempotency Key.
- Add validation for event dto.
- Table, written in the same transaction as `idempotency_keys`.

## [0.5.0] - 2026-08-11
### Added

- Kafka integration: `deploy/docker-compose.yml` (Kafka + Kafka UI).
- Kafka settings (`brokers`, `topics`, `poll_interval`, `batch_size`, `flush_timeout`).
- Transactional Outbox: `outbox` table, written in the same transaction as `events`.

## [0.4.0] - 2026-08-07

### Added

- PostgreSQL adapter and connection settings.
- Event domain model and `EventRepo` repository.
- DTOs and mappers for the HTTP API.
- Create and list events via API.

## [0.3.0] - 2026-08-05

### Added

- Graceful HTTP server shutdown on SIGINT/SIGTERM.
- Server timeouts from settings (`waiting_shutdown`, read/write/idle).

## [0.2.0] - 2026-08-04

### Added

- HTTP router (gorilla/mux).
- Handlers: health/service and events.
- Request logging middleware.
- Application skeleton (`application`, `settings`, `logger`).

## [0.1.0] - 2026-08-04

### Added

- Initial commit: Go service scaffold, `cmd/http/main.go`.

[Unreleased]: https://github.com/ikondratev/event-service/compare/v0.6.0...HEAD
[0.5.0]: https://github.com/ikondratev/event-service/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/ikondratev/event-service/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/ikondratev/event-service/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/ikondratev/event-service/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ikondratev/event-service/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ikondratev/event-service/releases/tag/v0.1.0
