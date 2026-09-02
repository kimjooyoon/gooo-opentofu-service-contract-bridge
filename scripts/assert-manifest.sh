#!/usr/bin/env bash
set -euo pipefail

manifest=${1:?manifest path is required}
jq -e '
  .schema == "gooo/opentofu-service-contract-bridge/v1"
  and .authority_scope == "GOOO_CLAIMS_ONLY"
  and .input_boundary == "READ_ONLY_SOURCE_REPOSITORY"
  and .denominator.total_cases == 12
  and .denominator.decision_order == ["REFUTED", "UNKNOWN", "CLOSED"]
  and .case_counts.CLOSED == 4
  and .case_counts.UNKNOWN == 4
  and .case_counts.REFUTED == 4
  and .denominator.proof_counts.FOUNDATION == 4
  and .denominator.proof_counts.COHERENCE == 4
  and .denominator.proof_counts.REGRESSION == 4
  and .denominator.indicator_counts.DRIVER == 4
  and .denominator.indicator_counts.OUTCOME == 4
  and .denominator.indicator_counts.GUARDRAIL == 4
  and (.denominator.activities | length) == 12
  and (.denominator.proof_cells | length) == 12
  and (.denominator.indicator_cells | length) == 12
  and (.unknown_claims | length) == 4
  and all(.unknown_claims[]; (.unknown.stage != "" and .unknown.step != "" and .unknown.reason != "" and .unknown.unknown_class != "" and .unknown.next_operation != "" and (.unknown.blocked_by | type) == "array"))
  and .mutation_boundary.repository_writes == 0
  and .mutation_boundary.local_test_executions == 0
  and .mutation_boundary.cross_project_required_gates == 0
  and .mutation_boundary.source_mutations == 0
  and .mutation_boundary.lockfile_mutations == 0
  and .mutation_boundary.network_provider_resolutions == 0
  and .mutation_boundary.opentofu_init_plan_apply_runs == 0
  and .mutation_boundary.caller_owned_output_only == true
  and .improvement.state == "UNKNOWN"
  and .external_user_utility.state == "UNKNOWN"
' "$manifest" > /dev/null

if rg -n -i '"(score|percentage)"|aggregate score|aggregate percentage' "$manifest"; then
  echo "forbidden aggregate field found" >&2
  exit 1
fi
