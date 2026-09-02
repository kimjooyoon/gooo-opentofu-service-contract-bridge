# Research basis

Research for the bridge design is intentionally limited to the official OpenTofu documentation, the official OpenTofu GitHub repository, and the official OpenAPI specification sources.

## OpenTofu

The [official configuration syntax documentation](https://opentofu.org/docs/language/syntax/configuration/) describes arguments and blocks, including labeled `resource` blocks and nested bodies. The [official language documentation](https://opentofu.org/docs/v1.10/language/) describes resources as the primary configuration element and explains that expressions can reference and combine values. The bridge therefore accepts only the smallest useful literal subset: two-label resource headers and single-line literal arguments. It does not pretend that this subset is the OpenTofu language.

The [official OpenTofu repository](https://github.com/opentofu/opentofu) is used as project identity and provenance only. No OpenTofu source is vendored, and the bridge does not run the CLI. In particular, a resource-shaped observation is not evidence that a provider is installed, a plan is valid, or an apply is safe.

## OpenAPI

The [OpenAPI Specification v3.1.0](https://spec.openapis.org/oas/v3.1.0) defines the OpenAPI document and the `paths` object containing available paths and operations. It also defines path templating and operation objects. The bridge observes only `openapi`, path keys, HTTP method keys, and optional `operationId`. The [official specification repository](https://github.com/OAI/OpenAPI-Specification) is recorded as the source repository for the published specification.

The bridge does not validate the full OpenAPI document. Components, `$ref`, schemas, request/response bodies, servers, security, webhooks, callbacks, and network behavior are deliberately outside the bounded reader.

## Non-inference rules

The source `.gooo` claims are the semantic authority. HCL and OpenAPI provide observations bound by `EvidenceBinding`; they do not rewrite claims. Missing observations remain UNKNOWN with six fields. Contradictions are REFUTED with precedence over UNKNOWN and CLOSED. No result is converted into a utility claim, improvement claim, score, or percentage.
