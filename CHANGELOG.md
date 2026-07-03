# Changelog

## Unreleased

### Breaking changes

- **Structured errors.** Errors now render in the active output format (TOON or JSON) as an `{"error": {code, message, hint?, retryable, details?}}` envelope on **stdout**, replacing plain `error: <message>` text on stderr. Error codes changed from `SCREAMING_CASE` to a documented closed `snake_case` set (see README → Errors). Scripts that parsed stderr text must switch to parsing the stdout envelope.
- **Exit codes.** The `0/1/2/3` scheme is replaced by `0` success, `1` operational error, `70` internal error, `80` usage/parse error. The retry signal moved from exit codes to the `retryable` field in the error envelope.
- **TOON field names.** TOON output now follows the same `json` struct-tag names as JSON output (`id`, `title`, `created_at` instead of `ID`, `Title`, `CreatedAt`), with empty optional fields omitted. TOON field order is alphabetical. Callers parsing the old Go-style field names must update.

### Fixed

- `-o json` / `-o toon` are now honored on error, not just success.
- TOON output no longer silently falls back to JSON for structs containing nil timestamp fields.
- Kong parse errors (unknown command, bad flag) now return the structured envelope with exit 80 instead of plain usage text.
