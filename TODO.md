## Follow-ups

- Re-export additional public types from the top-level `gocmlclient` package to avoid consumers importing `pkg/*` for common signatures.
  - Add `type Client = client.Client` to `gocmlclient/gocmlclient.go`.
  - Re-export error types/sentinels from `pkg/errors`:
    - `type APIError = errors.APIError`
    - `type ValidationError = errors.ValidationError`
    - `var Err...` sentinels (e.g. `ErrAPINotFound`, `ErrAPIUnauthorized`, etc.)
  - Consider re-exporting `type Config = client.Config` (optional; useful for helpers that inspect/modify settings).
