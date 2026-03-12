# Progress

## Start/Stop API Discrepancy

- [x] Reproduce discrepancy from OpenAPI 2.10 schema (`/labs/{lab_id}/start|stop|wipe`)
- [x] Implement compatibility fallback for lab actions
- [x] Keep existing behavior working for older/other backends (`/labs/{lab_id}/state/{action}`)
- [x] Run unit tests
- [ ] Run live integration: create lab, `Lab.Start`, `Lab.Stop`, `Lab.Wipe`
- [ ] Update schema docs (`schema-findings.md`) once verified on live instance

Notes:
- `internal/services/lab.go` now tries `labs/{lab_id}/state/{action}` first; on 404, retries `labs/{lab_id}/{action}`.
- This is scoped to lab start/stop/wipe only.

## Schema Parity Follow-Ups

- [ ] Update `schema-findings.md` to reflect lab action fix (schema-first + legacy fallback)
- [ ] Align `/import` with OpenAPI: add JSON topology import (keep YAML import as explicit compatibility, if desired)
- [x] Fix `GET /labs` query params: use `show_all` (schema) instead of `data`
- [ ] Handle nullable import warnings (`ImportTopologyResponse.warnings` may be `null`)
- [ ] Expand `models.SystemInformation` to include `timeout` and `features`
- [ ] Capture top-level `InterfaceResponse.mac_address` (schema) in `models.Interface`
- [ ] Reconcile `models.Node` vs schema (priority/pyats; operational-only fields vs top-level)
- [ ] Reconcile `models.ImageDefinition` vs schema (missing/extra fields)
