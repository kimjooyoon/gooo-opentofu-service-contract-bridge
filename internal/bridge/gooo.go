package bridge

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseGooo(path string) (Authority, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Authority{}, nil, err
	}
	authority, err := parseGoooBytes(raw)
	if err != nil {
		return Authority{}, nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return authority, raw, nil
}

func parseGoooBytes(raw []byte) (Authority, error) {
	authority := Authority{
		Infrastructure:   []InfrastructureClaim{},
		ServiceContracts: []ServiceContractClaim{},
		Mappings:         []MappingRule{},
		Evidence:         []EvidenceBinding{},
		Decisions:        []DriftDecision{},
		Cases:            []CaseDeclaration{},
		Activities:       []MetaActivity{},
	}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		tokens, err := tokenize(scanner.Text())
		if err != nil {
			return Authority{}, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		if len(tokens) == 0 {
			continue
		}
		switch tokens[0] {
		case "schema":
			if len(tokens) != 2 || authority.Schema != "" {
				return Authority{}, fmt.Errorf("line %d: invalid schema declaration", lineNumber)
			}
			authority.Schema = tokens[1]
		case "project":
			if len(tokens) != 2 || authority.Project != "" {
				return Authority{}, fmt.Errorf("line %d: invalid project declaration", lineNumber)
			}
			authority.Project = tokens[1]
		case "infrastructure_claim":
			values, err := keyValues(tokens[1:])
			if err != nil {
				return Authority{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			authority.Infrastructure = append(authority.Infrastructure, InfrastructureClaim{
				ID:           values["id"],
				Service:      values["service"],
				ResourceType: values["resource_type"],
				ResourceName: values["resource_name"],
				Scope:        values["scope"],
				Capability:   values["capability"],
			})
		case "service_contract_claim":
			values, err := keyValues(tokens[1:])
			if err != nil {
				return Authority{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			authority.ServiceContracts = append(authority.ServiceContracts, ServiceContractClaim{
				ID:           values["id"],
				Service:      values["service"],
				Contract:     values["contract"],
				OperationID:  values["operation_id"],
				Method:       strings.ToUpper(values["method"]),
				Path:         values["path"],
				ContractType: values["contract_type"],
				Name:         values["name"],
				Scope:        values["scope"],
			})
		case "mapping_rule":
			values, err := keyValues(tokens[1:])
			if err != nil {
				return Authority{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			authority.Mappings = append(authority.Mappings, MappingRule{
				ID:                    values["id"],
				InfrastructureClaim:  values["infrastructure_claim"],
				ServiceContractClaim: values["service_contract_claim"],
				Match:                 splitList(values["match"]),
			})
		case "evidence_binding":
			values, err := keyValues(tokens[1:])
			if err != nil {
				return Authority{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			authority.Evidence = append(authority.Evidence, EvidenceBinding{
				ID:                    values["id"],
				MappingRule:           values["mapping_rule"],
				InfrastructureLocator: values["infrastructure_locator"],
				ContractLocator:       values["contract_locator"],
				Evidence:              values["evidence"],
			})
		case "drift_decision":
			values, err := keyValues(tokens[1:])
			if err != nil {
				return Authority{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			decision := DriftDecision{
				ID:          values["id"],
				MappingRule: values["mapping_rule"],
				Expected:    values["expected"],
				Precedence:  splitPrecedence(values["precedence"]),
			}
			if values["unknown_stage"] != "" || values["unknown_step"] != "" || values["unknown_class"] != "" || values["unknown_reason"] != "" || values["next_operation"] != "" || values["blocked_by"] != "" || hasKey(values, "blocked_by") {
				decision.Unknown = &Unknown{
					Stage:         values["unknown_stage"],
					Step:          values["unknown_step"],
					Reason:        values["unknown_reason"],
					UnknownClass:  values["unknown_class"],
					NextOperation: values["next_operation"],
					BlockedBy:     splitListAllowEmpty(values["blocked_by"]),
				}
			}
			authority.Decisions = append(authority.Decisions, decision)
		case "case":
			values, err := keyValues(tokens[1:])
			if err != nil {
				return Authority{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			authority.Cases = append(authority.Cases, CaseDeclaration{
				ID:          values["id"],
				MappingRule: values["mapping_rule"],
				Expected:    values["expected"],
			})
		case "meta_activity":
			values, err := keyValues(tokens[1:])
			if err != nil {
				return Authority{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			ordinal, err := strconv.Atoi(values["ordinal"])
			if err != nil {
				return Authority{}, fmt.Errorf("line %d: ordinal must be an integer", lineNumber)
			}
			authority.Activities = append(authority.Activities, MetaActivity{
				Ordinal: ordinal,
				ID:      values["id"],
				Name:    values["name"],
				Proof:   values["proof"],
				Indicator: values["indicator"],
				Input:   values["input"],
				Output:  values["output"],
			})
		default:
			return Authority{}, fmt.Errorf("line %d: unsupported declaration %s", lineNumber, tokens[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return Authority{}, err
	}
	if err := validateAuthority(authority); err != nil {
		return Authority{}, err
	}
	return authority, nil
}

func tokenize(line string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	var quote rune
	escaped := false
	flush := func() {
		if current.Len() > 0 || quote != 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for _, r := range line {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if quote != 0 {
			if r == '\\' {
				escaped = true
			} else if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}
		if r == '#' {
			break
		}
		if r == '"' || r == '\'' {
			quote = r
			continue
		}
		if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
			flush()
			continue
		}
		current.WriteRune(r)
	}
	if escaped || quote != 0 {
		return nil, errors.New("unterminated quoted value")
	}
	flush()
	return tokens, nil
}

func keyValues(fields []string) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid key/value %q", field)
		}
		if _, exists := values[parts[0]]; exists {
			return nil, fmt.Errorf("duplicate key %s", parts[0])
		}
		values[parts[0]] = parts[1]
	}
	return values, nil
}

func hasKey(values map[string]string, key string) bool {
	_, ok := values[key]
	return ok
}

func splitList(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitListAllowEmpty(value string) []string {
	if value == "" {
		return []string{}
	}
	return splitList(value)
}

func splitPrecedence(value string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ">")
}

func validateAuthority(authority Authority) error {
	if authority.Schema != Schema || authority.Project == "" {
		return errors.New(".gooo authority must declare the bridge schema and project")
	}
	if len(authority.Cases) != 4 {
		return fmt.Errorf("each project must declare exactly four cases, got %d", len(authority.Cases))
	}
	if len(authority.Activities) != 12 {
		return fmt.Errorf(".gooo authority must declare exactly 12 meta activities, got %d", len(authority.Activities))
	}
	for index, activity := range authority.Activities {
		if activity.Ordinal != index+1 || activity.ID == "" || activity.Name != FixedActivityNames[index] || activity.Input == "" || activity.Output == "" {
			return fmt.Errorf("activity %d does not match the fixed activity denominator", index+1)
		}
		wantProof := FixedProofCells[index/4]
		wantIndicator := FixedIndicatorCells[index/4]
		if activity.Proof != wantProof || activity.Indicator != wantIndicator {
			return fmt.Errorf("activity %d must bind proof %s and indicator %s", index+1, wantProof, wantIndicator)
		}
	}
	if len(authority.Infrastructure) != 4 || len(authority.ServiceContracts) != 4 || len(authority.Mappings) != 4 || len(authority.Evidence) != 4 || len(authority.Decisions) != 4 {
		return errors.New("each project must declare four claims, mappings, bindings, and decisions")
	}
	if err := validateUniqueIDs(authority); err != nil {
		return err
	}
	infra := make(map[string]bool)
	for _, claim := range authority.Infrastructure {
		if claim.ID == "" || claim.Service == "" || claim.ResourceType == "" || claim.ResourceName == "" || claim.Scope == "" || claim.Capability == "" {
			return fmt.Errorf("infrastructure claim %s is incomplete", claim.ID)
		}
		infra[claim.ID] = true
	}
	contracts := make(map[string]bool)
	for _, claim := range authority.ServiceContracts {
		if claim.ID == "" || claim.Service == "" || claim.Contract == "" || claim.Method == "" || claim.Path == "" || claim.ContractType == "" || claim.Name == "" || claim.Scope == "" {
			return fmt.Errorf("service contract claim %s is incomplete", claim.ID)
		}
		contracts[claim.ID] = true
	}
	mappings := make(map[string]bool)
	for _, rule := range authority.Mappings {
		if rule.ID == "" || !infra[rule.InfrastructureClaim] || !contracts[rule.ServiceContractClaim] || len(rule.Match) == 0 {
			return fmt.Errorf("mapping rule %s is incomplete or references an unknown claim", rule.ID)
		}
		for _, field := range rule.Match {
			if field != "type" && field != "name" && field != "scope" && field != "method" {
				return fmt.Errorf("mapping rule %s uses unsupported match field %s", rule.ID, field)
			}
		}
		mappings[rule.ID] = true
	}
	evidence := make(map[string]bool)
	for _, binding := range authority.Evidence {
		if binding.ID == "" || !mappings[binding.MappingRule] || binding.InfrastructureLocator == "" || binding.ContractLocator == "" || binding.Evidence == "" {
			return fmt.Errorf("evidence binding %s is incomplete or references an unknown mapping", binding.ID)
		}
		evidence[binding.MappingRule] = true
	}
	if len(evidence) != len(authority.Mappings) {
		return errors.New("each mapping rule must have exactly one evidence binding")
	}
	decisions := make(map[string]DriftDecision)
	for _, decision := range authority.Decisions {
		if decision.ID == "" || !mappings[decision.MappingRule] || !validDecision(decision.Expected) || len(decision.Precedence) != 3 || decision.Precedence[0] != DecisionRefuted || decision.Precedence[1] != DecisionUnknown || decision.Precedence[2] != DecisionClosed {
			return fmt.Errorf("drift decision %s is incomplete or has invalid precedence", decision.ID)
		}
		if decision.Expected == DecisionUnknown {
			if decision.Unknown == nil || !decision.Unknown.Valid() {
				return fmt.Errorf("unknown decision %s must contain all six unknown fields", decision.ID)
			}
		} else if decision.Unknown != nil {
			return fmt.Errorf("non-unknown decision %s must not carry unknown fields", decision.ID)
		}
		decisions[decision.MappingRule] = decision
	}
	if len(decisions) != len(authority.Mappings) {
		return errors.New("each mapping rule must have exactly one drift decision")
	}
	seenCaseRules := map[string]bool{}
	for _, declaration := range authority.Cases {
		if declaration.ID == "" || !mappings[declaration.MappingRule] || !validDecision(declaration.Expected) {
			return fmt.Errorf("case %s is incomplete or references an unknown mapping", declaration.ID)
		}
		decision, ok := decisions[declaration.MappingRule]
		if !ok || decision.Expected != declaration.Expected || !evidence[declaration.MappingRule] {
			return fmt.Errorf("case %s is not fully bound by evidence and drift decision", declaration.ID)
		}
		if seenCaseRules[declaration.MappingRule] {
			return fmt.Errorf("mapping rule %s must bind exactly one case", declaration.MappingRule)
		}
		seenCaseRules[declaration.MappingRule] = true
	}
	if len(seenCaseRules) != len(authority.Mappings) {
		return errors.New("each mapping rule must have exactly one case")
	}
	return nil
}

func validateUniqueIDs(authority Authority) error {
	seen := map[string]string{}
	check := func(kind, id string) error {
		if previous, ok := seen[id]; ok {
			return fmt.Errorf("duplicate semantic id %s in %s and %s", id, previous, kind)
		}
		seen[id] = kind
		return nil
	}
	for _, claim := range authority.Infrastructure {
		if err := check("infrastructure_claim", claim.ID); err != nil {
			return err
		}
	}
	for _, claim := range authority.ServiceContracts {
		if err := check("service_contract_claim", claim.ID); err != nil {
			return err
		}
	}
	for _, mapping := range authority.Mappings {
		if err := check("mapping_rule", mapping.ID); err != nil {
			return err
		}
	}
	for _, binding := range authority.Evidence {
		if err := check("evidence_binding", binding.ID); err != nil {
			return err
		}
	}
	for _, decision := range authority.Decisions {
		if err := check("drift_decision", decision.ID); err != nil {
			return err
		}
	}
	for _, declaration := range authority.Cases {
		if err := check("case", declaration.ID); err != nil {
			return err
		}
	}
	for _, activity := range authority.Activities {
		if err := check("meta_activity", activity.ID); err != nil {
			return err
		}
	}
	return nil
}

func validDecision(decision string) bool {
	return decision == DecisionClosed || decision == DecisionUnknown || decision == DecisionRefuted
}
