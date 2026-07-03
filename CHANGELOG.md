# Changelog

## Unreleased

### Breaking changes

- **Structured errors.** Errors now render in the active output format (TOON or JSON) as an `{"error": {code, message, hint?, retryable, details?}}` envelope on **stdout**, replacing plain `error: <message>` text on stderr. Error codes changed from `SCREAMING_CASE` to a documented closed `snake_case` set (see README → Errors). Scripts that parsed stderr text must switch to parsing the stdout envelope.
- **Exit codes.** The `0/1/2/3` scheme is replaced by `0` success, `1` operational error, `70` internal error, `80` usage/parse error. The retry signal moved from exit codes to the `retryable` field in the error envelope.
- **TOON field names.** TOON output now follows the same `json` struct-tag names as JSON output (`id`, `title`, `created_at` instead of `ID`, `Title`, `CreatedAt`), with empty optional fields omitted. TOON field order is alphabetical. Callers parsing the old Go-style field names must update.

### Added

- `--schema` output now advertises the error contract: an `errors` block describing the envelope fields and the closed code set, on every command.

### Fixed

- `--schema` output's `exit_codes` documented the removed `0/1/2/3` scheme; it now reflects `0/1/70/80`.

- `-o json` / `-o toon` are now honored on error, not just success.
- Transient network failures (DNS, refused connection, timeout) and errors wrapped by the API layer now classify into their operational codes (`api_unavailable`, `unauthenticated`, ...) instead of reporting as `internal` with exit 70.
- HTTP 429 rate limiting now maps to `api_unavailable` with `retryable: true`, matching the documented contract, instead of `invalid_argument`.
- A remote command exiting non-zero over `sol ssh` and a rejected `sol push` now report `operation_failed` (exit 1) with the status in `details.exit_code`, instead of `internal` (exit 70).
- An unknown command combined with `--schema` now exits 80 with `invalid_argument`, consistent with other malformed invocations.
- The attached short-flag form (`-ojson`) is honored on error render paths, and invalid `--output` values fall back to the default format instead of producing unformatted output.
- Panic errors now carry the stack trace in `details.stack`.
- TOON output no longer silently falls back to JSON for structs containing nil timestamp fields.
- Kong parse errors (unknown command, bad flag) now return the structured envelope with exit 80 instead of plain usage text.
