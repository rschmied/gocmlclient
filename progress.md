# Progress

## Start/Stop API Discrepancy

- [x] Reproduce discrepancy from OpenAPI 2.10 schema (`/labs/{lab_id}/start|stop|wipe`)
- [x] Implement compatibility fallback for lab actions
- [x] Keep existing behavior working for older/other backends (`/labs/{lab_id}/state/{action}`)
- [x] Run unit tests
- [x] Run live integration: create lab, `Lab.Start`, `Lab.Stop`, `Lab.Wipe`
- [x] Update schema docs (`schema-findings.md`) once verified on live instance

Notes:
- `internal/services/lab.go` now tries `labs/{lab_id}/{action}` first; on 404, retries `labs/{lab_id}/state/{action}`.
- This is scoped to lab start/stop/wipe only.

## Schema Parity Follow-Ups

- [x] Update `schema-findings.md` to reflect lab action fix (schema-first + legacy fallback)
- [ ] Align `/import` with OpenAPI: add JSON topology import (keep YAML import as explicit compatibility, if desired)
- [x] Fix `GET /labs` query params: use `show_all` (schema) instead of `data`
- [ ] Handle nullable import warnings (`ImportTopologyResponse.warnings` may be `null`)
- [ ] Expand `models.SystemInformation` to include `timeout` and `features`
- [ ] Capture top-level `InterfaceResponse.mac_address` (schema) in `models.Interface`
- [ ] Reconcile `models.Node` vs schema (priority/pyats; operational-only fields vs top-level)
- [ ] Reconcile `models.ImageDefinition` vs schema (missing/extra fields)

## Terraform Provider Compat

- [x] Add CA certificate support in the new client (`client.WithCACertPEM`)
- [ ] Implement legacy entrypoint in `package gocmlclient`: `New(host string, insecure bool) *Client`
- [ ] Add legacy mutators on `*gocmlclient.Client`: `SetUsernamePassword`, `SetToken`, `UseNamedConfigs`, `SetCACert([]byte)`
- [ ] Ensure legacy defaults match v0.1.2: skip Ready() in constructor; named configs OFF until `UseNamedConfigs()`
- [ ] Add missing compatibility methods on `*pkg/client.Client` (for provider):
  - [ ] `HasLabConverged(ctx, labID string) (bool, error)` (alias to existing converge method)
  - [ ] `ImageDefinitions(ctx) ([]models.ImageDefinition, error)`
  - [ ] `ExtConnectors(ctx) ([]*models.ExtConn, error)`
- [ ] Match provider method signatures for nodes/links (pointer-based legacy types):
  - [ ] `NodeGet(ctx, node *Node) (*Node, error)` and `NodeStart/Stop/Wipe/Destroy(ctx, node *Node) error`
  - [ ] `NodeCreate(ctx, node *Node) (*Node, error)` and `NodeUpdate(ctx, node *Node) (*Node, error)`
  - [ ] `NodeSetConfig(ctx, node *Node, cfg string) error` and `NodeSetNamedConfigs(ctx, node *Node, cfgs []NodeConfig) error`
  - [ ] `LinkGet(ctx, labID, linkID string, deep bool) (*Link, error)`
  - [ ] `LinkCreate(ctx, link *Link) (*Link, error)` and `LinkDestroy(ctx, link *Link) error`
- [ ] Restore legacy model types expected by provider in `package gocmlclient` (do not alias if shapes differ):
  - [ ] `Client`, `Lab`, `LabGroup`, `Node`, `NodeMap`, `NodeConfig`, `SerialDevice`, `Link`, `Interface`, `InterfaceList`, `User`, `Group`, `GroupLab`, `ExtConn`, `ImageDefinition`
  - [ ] Add conversion between legacy models and `pkg/models` in wrappers
- [ ] Re-export legacy errors and state constants in `package gocmlclient`:
  - [ ] `ErrSystemNotReady`, `ErrElementNotFound`
  - [ ] `LabStateDefined/Started/Stopped`, `NodeStateDefined/Stopped`, `IfaceStateDefined/Stopped`
