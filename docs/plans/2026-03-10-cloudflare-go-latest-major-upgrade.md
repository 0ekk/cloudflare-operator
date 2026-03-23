# Cloudflare Go Latest Major Upgrade Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Upgrade to the latest major version of `cloudflare-go`, treat Cloudflare official API as source of truth, and adapt project code with test-first workflow.

**Architecture:** Keep all SDK-specific adaptations inside `internal/clients/cf` so controllers and CRDs stay stable. Use red-green-refactor per compatibility area: tunnel lifecycle, tunnel configuration, routes/virtual networks, and warp connector paths.

**Tech Stack:** Go 1.25, controller-runtime, cloudflare-go (latest major), gomock, Ginkgo/Gomega, go test.

---

### Task 1: Environment and Version Discovery

**Files:**
- Modify: `go.mod`, `go.sum`
- Verify: `internal/clients/cf/*`

**Step 1: Run version discovery command**
- Run `go list -m -versions github.com/cloudflare/cloudflare-go` and major-path variant if required.

**Step 2: Capture baseline failures**
- Run `go test ./internal/clients/cf/...` with writable cache envs.
- Record compile/runtime failures for later mapping.

### Task 2: TDD for Compatibility Surface

**Files:**
- Modify/Test: `internal/clients/cf/*_test.go`

**Step 1: Add or adjust failing tests first**
- Focus on API call signatures and response model transformations used by:
  - Tunnel lifecycle (`CreateTunnelWithParams`, `DeleteTunnelByID`, `GetTunnelToken`)
  - Tunnel config (`GetTunnelConfiguration`, `UpdateTunnelConfiguration`)
  - Network (`VirtualNetwork`, `TunnelRoute`)
  - WARP connector (`CreateWARPConnector`, token fetch)

**Step 2: Run targeted tests and confirm failures are due to SDK incompatibility**
- Use package-level `go test` selectors for the touched tests.

### Task 3: Minimal Production Adaptation

**Files:**
- Modify: `internal/clients/cf/*.go`
- Optional regenerate: `internal/clients/cf/mock/mock_client.go`

**Step 1: Update imports/types/signatures to latest major SDK**
- Keep outward project interface stable where possible.

**Step 2: Implement compatibility adapters in `internal/clients/cf`**
- Convert new SDK request/response shapes to existing local structs.

**Step 3: Make failing tests pass with minimal code**

### Task 4: Full Verification

**Files:**
- Verify: `internal/clients/cf`, selected controller packages using cf client

**Step 1: Run focused tests**
- `go test ./internal/clients/cf/...`

**Step 2: Run broader regression**
- `go test ./internal/controller/...`

**Step 3: Run lint/check if needed for API signature changes**
- `make lint-full` (or targeted golangci run).

### Task 5: Deliverables

**Step 1: Summarize incompatible changes handled**
- List API/signature migrations and behavior changes.

**Step 2: Document remaining gaps versus official API**
- Highlight tunnel subresources still not modeled in CRDs/controllers.

---

## Progress Status (updated 2026-03-11)

- Task 1: ✅ Completed
- Task 2: ✅ Completed
- Task 3: ✅ Completed (v6-first compatibility in `internal/clients/cf`)
- Task 4: ✅ Completed (`go test ./internal/clients/cf/...`, `go test ./internal/controller/...`, `go test $(go list ./... | grep -v /e2e)`)
- Task 5: ✅ Completed

Overall completion: **100% (for this migration scope)**

## Incompatible Changes Handled

- Added parallel v6 client wiring (`CloudflareV6`) while preserving existing `cloudflare.API` usage for compatibility fallback.
- Updated tunnel lifecycle and lookup paths to v6-first calls:
  - create/delete/get/list by name/token/creds
- Updated tunnel configuration get/update to v6 endpoints, with controlled fallback when legacy-only fields are present.
- Updated Zero Trust network virtual-network and route operations to v6 request/response types.
- Updated WARP connector create/token/delete to v6 endpoints.
- Added package-level v6 behavior tests to lock request paths and payload mapping for tunnel/network/connector/config flows.

## Tunnel API Coverage Against cloudflare-go/v6.8.0

Implemented in project (cloudflared):
- `Tunnels.Cloudflared.New/List/Get/Delete`
- `Tunnels.Cloudflared.Token.Get`
- `Tunnels.Cloudflared.Configurations.Get/Update` (with safe fallback to legacy SDK for unsupported config parts)

Implemented in project (WARP connector):
- `Tunnels.WARPConnector.New/Delete`
- `Tunnels.WARPConnector.Token.Get`

Not yet exposed by current project service layer (available in v6 SDK):
- `Tunnels.Cloudflared.Edit`
- `Tunnels.Cloudflared.Connections.Get/Delete`
- `Tunnels.Cloudflared.Connectors.Get`
- `Tunnels.Cloudflared.Management.New`
- `Tunnels.WARPConnector.List/Get/Edit`

Notes:
- These are currently API-surface gaps, not migration blockers for existing CRDs/controllers.
- Tunnel configuration update intentionally falls back to legacy SDK when fields cannot be represented safely in v6 request structs.
