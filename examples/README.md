# Examples

This folder contains runnable examples.

Most examples read configuration from environment variables:

- `CML_BASE_URL` (e.g. `https://cml-controller.example.com`)
- `CML_TOKEN` (preferred)
- `CML_USER` / `CML_PASS` (fallback for username/password examples)

Run an example with:

```bash
go run ./examples/<name>
```

Note: examples may mutate server state. Read the example source before running.
