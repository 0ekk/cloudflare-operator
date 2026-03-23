# Repository Guidelines

## Project Structure & Module Organization
`cmd/main.go` starts the operator manager. API types live under `api/` (`v1alpha1`, `v1alpha2`, and Cloudflare-specific schemas), while reconciliation and service logic are in `internal/controller/`, `internal/service/`, `internal/clients/cf/`, and related helper packages. Kubernetes manifests and RBAC live in `config/`; generated bundle artifacts are in `bundle/`. End-to-end and support test code sits in `test/e2e/` and `test/mockserver/`. Use `examples/` for runnable manifests and `docs/en/` or `docs/zh/` for user-facing documentation.

## Build, Test, and Development Commands
Use the `Makefile` as the entrypoint for local work:

- `make build` builds `bin/manager` after regenerating manifests and running format/vet checks.
- `make run` runs the controller locally against the current kubeconfig.
- `make test` runs non-E2E Go tests with envtest and writes `cover.out`.
- `make test-e2e` provisions a Kind cluster, deploys the operator and mock server, runs tagged E2E tests, then tears the cluster down.
- `make lint` checks new changes with `golangci-lint`; use `make lint-full` before large submissions.
- `make install` installs CRDs; `make deploy IMG=<image>` deploys the controller to a cluster.

## Coding Style & Naming Conventions
This repository uses Go 1.24 with `gofmt` and `goimports`; keep tabs out and rely on standard Go formatting. Package names stay lowercase; exported identifiers use `CamelCase`; CRD type files follow `<resource>_types.go`; controller packages typically mirror the resource name, for example `internal/controller/pagesproject/`. Run `make manifests generate` whenever API shapes change. `golangci-lint` enforces checks including `errcheck`, `staticcheck`, `revive`, and `ginkgolinter`.

## Testing Guidelines
Place unit and integration tests next to the code as `*_test.go`. The suite uses Go’s testing package plus Ginkgo/Gomega-style specs (`Describe`, `Context`, `It`) in controller and webhook areas. Keep E2E coverage in `test/e2e/scenarios/` and mock Cloudflare behavior in `test/mockserver/`. Add or update tests for behavior changes before opening a PR.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commits with scopes, for example `fix(pagesdeployment): ...` and `feat(pagesproject): ...`. Keep commits focused and use `feat`, `fix`, `docs`, `refactor`, or `chore` as appropriate. PRs should describe the change, explain why it is needed, list verification steps (`make test`, `make lint`, E2E if relevant), and link the related issue. Include manifest or documentation updates whenever CRDs, examples, or user workflows change.

## Tunnel ConfigMap Ownership Caveat
When writing tunnel aggregation ConfigMaps (`internal/controller/tunnelconfig/`), do **not** set cross-namespace owner references for namespaced owners (for example, `Tunnel` in `default` owning a ConfigMap in `cloudflare-operator-system`). Kubernetes garbage collection may delete the ConfigMap as an invalid/dangling dependent, which causes `Published application routes` to disappear while DNS records still exist.

- Keep owner references only when owner scope/namespace is valid.
- Add/maintain unit tests for owner-reference behavior (`internal/controller/tunnelconfig/types_ownerref_test.go`).
- If routes disappear but DNS remains, check operator logs for repeated `Creating tunnel config ConfigMap` followed by `ConfigMap ... not found`.

### Fast Diagnosis (Routes Missing, DNS Still Present)
Use this checklist when Cloudflare Tunnel `Published application routes` disappear after pod/operator restart:

1. Confirm tunnel ID:
   - `kubectl -n <ns> get tunnel <name> -o jsonpath='{.status.tunnelId}{"\n"}'`
2. Check tunnel-config ConfigMap existence:
   - `kubectl -n cloudflare-operator-system get configmap tunnel-config-<tunnelId>`
3. Check operator logs for GC symptom:
   - `kubectl -n cloudflare-operator-system logs deploy/cloudflare-operator-controller-manager --since=2h | rg "Creating tunnel config ConfigMap|ConfigMap .* not found"`
4. Inspect aggregated source count/rules:
   - `kubectl -n cloudflare-operator-system get configmap tunnel-config-<tunnelId> -o jsonpath='{.data.config\.json}' | jq '.sources|keys,.sources|to_entries|map({k:.key,rules:(.value.rules|length)})'`

If step 3 repeatedly shows `Creating ...` then immediate `ConfigMap ... not found`, treat as ownership/GC issue first.

### Recovery Steps
1. Upgrade operator to a build containing the owner-reference guard in `internal/controller/tunnelconfig/types.go`.
2. Restart operator deployment:
   - `kubectl -n cloudflare-operator-system rollout restart deploy/cloudflare-operator-controller-manager`
3. Trigger one Ingress reconcile (for example, patch a harmless annotation).
4. Verify ConfigMap is stable:
   - `kubectl -n cloudflare-operator-system get cm tunnel-config-<tunnelId> -w`
5. Verify routes are restored in Cloudflare Tunnel UI/API.

### Code Review Guardrails
- Any change touching `NewConfigMap(...)` or writer logic must explicitly evaluate owner reference validity by scope/namespace.
- Do not regress fallback sync behavior (`http_status:404`) when no app routes remain.
- Keep targeted tests runnable with:
  - `go test ./internal/controller/tunnelconfig -run TestNewConfigMap -v`
