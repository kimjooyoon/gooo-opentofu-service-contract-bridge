package bridge

import (
	"path/filepath"
	"testing"
)

func TestFixedExampleSuite(t *testing.T) {
	projection, err := RunSuite(filepath.Join("..", "..", "examples"))
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Cases) != 12 {
		t.Fatalf("got %d cases, want 12", len(projection.Cases))
	}
	counts := decisionCounts(projection.Cases)
	if counts[DecisionClosed] != 4 || counts[DecisionUnknown] != 4 || counts[DecisionRefuted] != 4 {
		t.Fatalf("got decision counts %#v, want 4/4/4", counts)
	}
	for _, result := range projection.Cases {
		if result.Decision == DecisionUnknown && (result.Unknown == nil || !result.Unknown.Valid()) {
			t.Fatalf("unknown case %s lost its six-field detail", result.ID)
		}
	}
}

func TestPrecedenceRefutesBeforeMissingEvidence(t *testing.T) {
	result := evaluateMapping(
		CaseDeclaration{ID: "case", MappingRule: "map", Expected: DecisionRefuted},
		MappingRule{ID: "map", Match: []string{"type"}},
		InfrastructureClaim{Service: "shop", Capability: "queue"},
		ServiceContractClaim{Service: "shop", ContractType: "http"},
		EvidenceBinding{InfrastructureLocator: "resource.gooo_service.absent", ContractLocator: "path.GET./absent"},
		DriftDecision{Expected: DecisionRefuted},
		nil,
		nil,
	)
	if result.Decision != DecisionRefuted {
		t.Fatalf("got %s, want REFUTED", result.Decision)
	}
}
