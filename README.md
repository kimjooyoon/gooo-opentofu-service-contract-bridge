# Gooo OpenTofu Service Contract Bridge

This repository is a deliberately small, read-only bridge between a Gooo semantic graph and service-shaped OpenTofu/OpenAPI observations. It generates a human dossier and machine manifest for matches, unresolved evidence, and known contradictions without applying infrastructure or editing the source project.

The `.gooo` file is the semantic authority for `InfrastructureClaim`, `ServiceContractClaim`, `MappingRule`, `EvidenceBinding`, `DriftDecision`, and the fixed 12 meta activities. Go is only the bounded reader, projector, generator, and CLI. The suite has three small projects:

| Example | Cases | Purpose |
| --- | ---: | --- |
| `examples/closed` | 4 CLOSED | type/name/scope evidence is present and coherent |
| `examples/unknown` | 4 UNKNOWN | missing resource, operation, attribute, or operationId |
| `examples/refuted` | 4 REFUTED | known type, name, scope, or method contradiction |

The suite denominator is fixed at 12 cases: 4 CLOSED, 4 UNKNOWN, and 4 REFUTED. The decision order is `REFUTED > UNKNOWN > CLOSED`. Every UNKNOWN carries `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`. No aggregate score or percentage is emitted. Improvement is UNKNOWN without a matched before/after observation pair, and external user utility is UNKNOWN without a user observation.

## Use

The conformance run writes only to an absolute, empty, caller-owned directory outside the source repository:

```text
go run ./cmd/gooo-opentofu-service-contract-bridge \
  -suite-root ./examples \
  -inventory-root . \
  -output /tmp/gooo-bridge-output
```

The output contains exactly four generated artifacts: `dossier.md`, `manifest.json`, `mapping-report.json`, and `projected-claims.json`. The manifest records source digests, line counts, generated artifact counts, `wall_ms`, `peak_rss_kib`, fixed cells, decisions, and the zero-mutation boundary.

For one project, provide `-gooo`, `-infra`, and `-contract` instead of `-suite-root`. The release and CI evidence path uses the suite so that the fixed 12-case denominator is always exercised.

## OpenTofu boundary

OpenTofu is not treated as Terraform by implication. The reader is based on the official OpenTofu configuration description and accepts only native syntax containing top-level `resource "TYPE" "NAME" { ... }` blocks with literal, single-line string/boolean/number arguments. It does not execute OpenTofu and does not resolve providers.

The following are outside this bounded reader: `module`, `provider`, `data`, `variable`, `output`, `locals`, nested blocks, expressions, JSON configuration, full HCL, `tofu init`, `tofu validate`, `tofu plan`, `tofu apply`, state/backend access, provider installation, lockfiles, credentials, and network activity. A reader boundary failure is preserved as `OPERATIONAL_REFUTED`; it is never silently treated as a semantic match.

## OpenAPI boundary

The reader accepts an OpenAPI 3.x JSON document or a small YAML subset containing `openapi`, `paths`, HTTP methods, and optional `operationId`. It observes path/method/operationId only. It is not a full OpenAPI validator and does not resolve references or inspect components, schemas, servers, security, webhooks, callbacks, or network responses.

## Authority and evidence

The input repository is read-only. The CLI never calls OpenTofu, never performs provider resolution, and never writes source, lockfile, `.git`, or cloud state. Generated files are written only to the caller-owned output directory. If a run fails after producing artifacts, preceding artifacts are retained and an `operational-refuted.json` record may be added to the output directory.

The fixed proof cells are FOUNDATION, COHERENCE, and REGRESSION at 4/4/4. The fixed indicator cells are DRIVER, OUTCOME, and GUARDRAIL at 4/4/4. The `.gooo` source binds one fixed activity to each cell pair.

## Evidence basis

- [OpenTofu Configuration Syntax](https://opentofu.org/docs/language/syntax/configuration/)
- [OpenTofu Language Documentation](https://opentofu.org/docs/v1.10/language/)
- [OpenTofu official GitHub repository](https://github.com/opentofu/opentofu)
- [OpenAPI Specification v3.1.0](https://spec.openapis.org/oas/v3.1.0)
- [OpenAPI Specification repository](https://github.com/OAI/OpenAPI-Specification)

See [the protocol](docs/protocol.md) for the grammar, projection rules, output contract, and release boundary. The root README is intentionally excluded from the inventory reported by the CLI.
