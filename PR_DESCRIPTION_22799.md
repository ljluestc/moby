# PR Description — Issue #22799: daemon `--dns` not applied in internal overlay networks

## Summary
This PR fixes DNS resolution behavior for containers attached to **internal overlay** networks so daemon-level DNS configuration (`--dns`) is consistently honored. Today, the same daemon DNS setting can work on non-internal overlays but fail on internal overlays, resulting in unresolved service names.

Closes #22799.

## Problem
When Docker is started with daemon DNS configuration (for example `--dns=10.0.0.1`) and a container is attached to an **internal overlay** network, lookups that should be forwarded to that configured resolver can fail with `Host unknown`.

Observed behavior from the report:
- Service discovery record exists in external resolver (Consul) and resolves correctly when queried directly.
- Container on internal overlay cannot resolve the same record.
- Same setup works when overlay is not marked internal.

Expected behavior:
- Internal overlay should not bypass or break daemon-configured upstream DNS behavior for container name resolution.

## Root Cause Hypothesis
DNS path selection for internal overlays likely diverges from non-internal overlays and does not include daemon-configured upstream nameservers in the container resolver flow. As a result, external names that are only resolvable by the configured upstream DNS are not reachable from internal-overlay-attached containers.

## Proposed Fix
1. Normalize resolver construction for overlay networks so internal overlays preserve daemon DNS settings where appropriate.
2. Ensure embedded DNS + upstream forwarding behavior is consistent between internal and non-internal overlays for external-name queries.
3. Keep internal overlay isolation semantics intact (no unintended egress/network policy bypass).

## Compatibility and Risk
- **Compatibility**: Intended behavior is additive and aligns internal overlay DNS handling with existing daemon DNS expectations.
- **Risk**: resolver-path changes may impact edge-case name resolution ordering.
- **Mitigation**:
  - add targeted regression coverage for internal overlay + daemon DNS.
  - validate no regressions for default embedded DNS behavior and local container name resolution.

## Validation Plan
- Add/extend tests for:
  - daemon `--dns` forwarding on internal overlay.
  - parity with non-internal overlay DNS behavior.
  - unchanged local service/container name resolution in overlay.
- Manual validation scenario:
  1. Start daemon with custom `--dns`.
  2. Create internal overlay network.
  3. Run container attached to that network.
  4. Resolve external service name available only through upstream DNS.
  5. Confirm resolution succeeds.

## Rollout
- Merge with regression tests.
- Include release note under networking/DNS overlay behavior.
- Recommend affected users to retest internal overlay DNS scenarios after upgrade.

## Checklist
- [x] Full local PR description drafted.
- [ ] Functional code changes implemented.
- [ ] DNS regression tests added and passing.
