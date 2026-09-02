# Protocol v1

## Semantic source

Each example contains one `main.gooo` file. The declaration order is not a score; it is a stable semantic record order. The record kinds are:

- `infrastructure_claim`: service, resource type/name, scope, and capability.
- `service_contract_claim`: OpenAPI contract, operation id, HTTP method/path, contract type, name, and scope.
- `mapping_rule`: the claim pair and fields to compare (`type`, `name`, `scope`, and optionally `method`).
- `evidence_binding`: read-only locators into the bounded OpenTofu and OpenAPI observations.
- `drift_decision`: the authoritative expected decision, fixed precedence, and six-field UNKNOWN detail when expected is UNKNOWN.
- `case`: the fixed case identity and its expected decision.
- `meta_activity`: one of the 12 fixed names and its proof/indicator pair.

The parser rejects incomplete claims, duplicate semantic ids, unknown references, non-fixed activities, and a precedence other than `REFUTED>UNKNOWN>CLOSED`.

## Projection

The projector resolves a resource locator of the form `resource.TYPE.NAME` and an OpenAPI locator of the form `path.METHOD.PATH`. It compares claim fields and, where present, observed literal arguments such as `service_type` and `scope`. A missing locator, missing required literal, or absent optional `operationId` is UNKNOWN. A known claim or observation contradiction is REFUTED even when another piece of evidence is missing. A fully evidenced coherent mapping is CLOSED.

The three example projects are combined in suite mode. The combined result must contain exactly 12 cases with exactly four of each decision. Single-project mode remains available for local consumers, but the release contract is suite mode.

## Generated output

The generator requires an absolute empty output directory outside the source repository. It writes:

1. `projected-claims.json` — parsed semantic authority, observed inputs, and source provenance.
2. `mapping-report.json` — case-by-case result with contradictions, missing evidence, and UNKNOWN detail.
3. `dossier.md` — human-readable decision dossier with boundaries, counts, line counts, and official references.
4. `manifest.json` — machine contract for the run.

The manifest's `generated_artifact_count` includes itself. Its digest list intentionally covers the other three artifacts so that the manifest does not contain a self-referential digest. The manifest records input file/folder counts, Go/`.gooo`/HCL/OpenAPI physical line counts, generated artifact bytes/digests, wall milliseconds, and peak RSS KiB. Root `README.md` is excluded from repository inventory.

Runtime mutation fields are all zero: repository writes, local test executions, cross-project required gates, source mutations, lockfile mutations, network provider resolutions, and OpenTofu init/plan/apply runs. This repository's CI performs the tests and static checks in GitHub Actions; the intended evidence-producing CLI itself performs no OpenTofu execution.

## Operational failures

Unsupported syntax, malformed source, failed conformance, or failed release preparation is `OPERATIONAL_REFUTED`. Existing run output and upstream evidence must be retained. CI uploads the evidence directory with `always()`. The release workflow consumes a successful main-branch evidence artifact, creates a draft release first, uploads exactly four assets, verifies `SHA256SUMS`, then publishes without post-publish asset mutation.
