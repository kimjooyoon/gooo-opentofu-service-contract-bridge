package bridge

import (
	"fmt"
	"strings"
)

func ProjectInputs(authority Authority, infra InfraInput, api OpenAPIInput, sources []ProjectSource) (Projection, error) {
	projection := Projection{
		Project:   authority.Project,
		Authority: authority,
		Infra:     infra,
		OpenAPI:   api,
		Cases:     []CaseResult{},
		Sources:   sources,
	}
	infraClaims := make(map[string]InfrastructureClaim, len(authority.Infrastructure))
	for _, claim := range authority.Infrastructure {
		infraClaims[claim.ID] = claim
	}
	contractClaims := make(map[string]ServiceContractClaim, len(authority.ServiceContracts))
	for _, claim := range authority.ServiceContracts {
		contractClaims[claim.ID] = claim
	}
	mappings := make(map[string]MappingRule, len(authority.Mappings))
	for _, rule := range authority.Mappings {
		mappings[rule.ID] = rule
	}
	evidence := make(map[string]EvidenceBinding, len(authority.Evidence))
	for _, binding := range authority.Evidence {
		evidence[binding.MappingRule] = binding
	}
	decisions := make(map[string]DriftDecision, len(authority.Decisions))
	for _, decision := range authority.Decisions {
		decisions[decision.MappingRule] = decision
	}
	for _, declaration := range authority.Cases {
		rule := mappings[declaration.MappingRule]
		infraClaim := infraClaims[rule.InfrastructureClaim]
		contractClaim := contractClaims[rule.ServiceContractClaim]
		binding := evidence[rule.ID]
		decision := decisions[rule.ID]
		result := evaluateMapping(declaration, rule, infraClaim, contractClaim, binding, decision, infra.Resources, api.Operations)
		if result.Decision != declaration.Expected {
			return Projection{}, fmt.Errorf("case %s expected %s but observed %s", declaration.ID, declaration.Expected, result.Decision)
		}
		projection.Cases = append(projection.Cases, result)
	}
	return projection, nil
}

func evaluateMapping(declaration CaseDeclaration, rule MappingRule, infraClaim InfrastructureClaim, contractClaim ServiceContractClaim, binding EvidenceBinding, expected DriftDecision, resources []InfraResource, operations []OpenAPIOperation) CaseResult {
	result := CaseResult{
		ID:             declaration.ID,
		MappingRule:    rule.ID,
		Expected:       declaration.Expected,
		Matches:        []string{},
		Contradictions: []string{},
		Missing:        []string{},
	}
	resource := findResource(resources, binding.InfrastructureLocator)
	operation := findOperation(operations, binding.ContractLocator)
	result.Evidence = ObservedEvidence{Infrastructure: resource, Contract: operation}

	if infraClaim.Service != contractClaim.Service {
		result.Contradictions = append(result.Contradictions, "service claim differs")
	}
	for _, field := range rule.Match {
		switch field {
		case "type":
			if infraClaim.Capability != contractClaim.ContractType {
				result.Contradictions = append(result.Contradictions, "type claim differs")
			} else if resource != nil && resource.Attributes["service_type"] != "" && resource.Attributes["service_type"] != infraClaim.Capability {
				result.Contradictions = append(result.Contradictions, "observed service_type differs")
			} else if resource == nil {
				result.Missing = append(result.Missing, "resource.service_type")
			} else if resource.Attributes["service_type"] == "" {
				result.Missing = append(result.Missing, "resource.service_type")
			} else {
				result.Matches = append(result.Matches, "type")
			}
		case "name":
			if infraClaim.ResourceName != contractClaim.Name {
				result.Contradictions = append(result.Contradictions, "name claim differs")
			} else if resource != nil && resource.Name != infraClaim.ResourceName {
				result.Contradictions = append(result.Contradictions, "observed resource name differs")
			} else if operation != nil && contractClaim.OperationID != "" && operation.OperationID != "" && operation.OperationID != contractClaim.OperationID {
				result.Contradictions = append(result.Contradictions, "observed operationId differs")
			} else if operation == nil {
				result.Missing = append(result.Missing, "contract operation")
			} else if contractClaim.OperationID != "" && operation.OperationID == "" {
				result.Missing = append(result.Missing, "contract operationId")
			} else {
				result.Matches = append(result.Matches, "name")
			}
		case "scope":
			if infraClaim.Scope != contractClaim.Scope {
				result.Contradictions = append(result.Contradictions, "scope claim differs")
			} else if resource != nil && resource.Attributes["scope"] != "" && resource.Attributes["scope"] != infraClaim.Scope {
				result.Contradictions = append(result.Contradictions, "observed scope differs")
			} else if resource == nil {
				result.Missing = append(result.Missing, "resource.scope")
			} else if resource.Attributes["scope"] == "" {
				result.Missing = append(result.Missing, "resource.scope")
			} else {
				result.Matches = append(result.Matches, "scope")
			}
		case "method":
			if operation != nil && operation.Method != strings.ToUpper(contractClaim.Method) {
				result.Contradictions = append(result.Contradictions, "observed method differs")
			} else if operation == nil {
				result.Missing = append(result.Missing, "contract method")
			} else {
				result.Matches = append(result.Matches, "method")
			}
		}
	}
	if operation != nil && (operation.Method != strings.ToUpper(contractClaim.Method) || operation.Path != contractClaim.Path) {
		result.Contradictions = append(result.Contradictions, "contract locator differs from claim")
	}
	if resource == nil {
		result.Missing = append(result.Missing, "infrastructure locator")
	}
	if operation == nil {
		result.Missing = append(result.Missing, "contract locator")
	}
	result.Missing = uniqueStrings(result.Missing)
	result.Contradictions = uniqueStrings(result.Contradictions)

	switch {
	case len(result.Contradictions) > 0:
		result.Decision = DecisionRefuted
	case len(result.Missing) > 0:
		result.Decision = DecisionUnknown
		result.Unknown = expected.Unknown
	case len(result.Matches) == len(rule.Match):
		result.Decision = DecisionClosed
	default:
		result.Decision = DecisionUnknown
		result.Unknown = expected.Unknown
	}
	return result
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
