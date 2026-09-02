package bridge

const (
	Schema          = "gooo/opentofu-service-contract-bridge/v1"
	DecisionClosed  = "CLOSED"
	DecisionUnknown = "UNKNOWN"
	DecisionRefuted = "REFUTED"
	Version         = "v0.1.0"
)

var FixedActivityNames = []string{
	"ReadGoooAuthority",
	"ReadBoundedOpenTofu",
	"ReadBoundedOpenAPI",
	"InventorySources",
	"ProjectClaims",
	"BindEvidence",
	"EvaluateMappings",
	"ClassifyDecision",
	"PreserveUnknown",
	"RenderHumanDossier",
	"RenderMachineManifest",
	"AuditOutputBoundary",
}

var FixedProofCells = []string{"FOUNDATION", "COHERENCE", "REGRESSION"}

var FixedIndicatorCells = []string{"DRIVER", "OUTCOME", "GUARDRAIL"}

type Unknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func (u Unknown) Valid() bool {
	return u.Stage != "" && u.Step != "" && u.Reason != "" &&
		u.UnknownClass != "" && u.NextOperation != "" && u.BlockedBy != nil
}

type InfrastructureClaim struct {
	ID           string `json:"id"`
	Service      string `json:"service"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	Scope        string `json:"scope"`
	Capability   string `json:"capability"`
}

type ServiceContractClaim struct {
	ID           string `json:"id"`
	Service      string `json:"service"`
	Contract     string `json:"contract"`
	OperationID  string `json:"operation_id"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	ContractType string `json:"contract_type"`
	Name         string `json:"name"`
	Scope        string `json:"scope"`
}

type MappingRule struct {
	ID                    string   `json:"id"`
	InfrastructureClaim  string   `json:"infrastructure_claim"`
	ServiceContractClaim string   `json:"service_contract_claim"`
	Match                 []string `json:"match"`
}

type EvidenceBinding struct {
	ID                    string `json:"id"`
	MappingRule           string `json:"mapping_rule"`
	InfrastructureLocator string `json:"infrastructure_locator"`
	ContractLocator       string `json:"contract_locator"`
	Evidence              string `json:"evidence"`
}

type DriftDecision struct {
	ID          string   `json:"id"`
	MappingRule string   `json:"mapping_rule"`
	Expected    string   `json:"expected"`
	Precedence  []string `json:"precedence"`
	Unknown     *Unknown `json:"unknown,omitempty"`
}

type CaseDeclaration struct {
	ID          string `json:"id"`
	MappingRule string `json:"mapping_rule"`
	Expected    string `json:"expected"`
}

type MetaActivity struct {
	Ordinal  int    `json:"ordinal"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Proof    string `json:"proof"`
	Indicator string `json:"indicator"`
	Input    string `json:"input"`
	Output   string `json:"output"`
}

type Authority struct {
	Schema             string                 `json:"schema"`
	Project            string                 `json:"project"`
	Infrastructure     []InfrastructureClaim  `json:"infrastructure_claims"`
	ServiceContracts   []ServiceContractClaim `json:"service_contract_claims"`
	Mappings           []MappingRule          `json:"mapping_rules"`
	Evidence           []EvidenceBinding      `json:"evidence_bindings"`
	Decisions          []DriftDecision        `json:"drift_decisions"`
	Cases              []CaseDeclaration      `json:"cases"`
	Activities         []MetaActivity         `json:"activities"`
}

type InfraResource struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	File       string            `json:"file"`
	Line       int               `json:"line"`
	Attributes map[string]string `json:"attributes"`
}

type InfraInput struct {
	Resources []InfraResource `json:"resources"`
	Files     []FileDigest    `json:"files"`
}

type OpenAPIOperation struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	OperationID string `json:"operation_id"`
	File        string `json:"file"`
	Line        int    `json:"line"`
}

type OpenAPIInput struct {
	Version    string             `json:"version"`
	Operations []OpenAPIOperation `json:"operations"`
	File       string             `json:"file"`
	Lines      int                `json:"lines"`
}

type FileDigest struct {
	Path      string `json:"path"`
	Bytes     int    `json:"bytes"`
	Lines     int    `json:"lines"`
	SHA256    string `json:"sha256"`
}

type ObservedEvidence struct {
	Infrastructure *InfraResource    `json:"infrastructure,omitempty"`
	Contract       *OpenAPIOperation  `json:"contract,omitempty"`
}

type CaseResult struct {
	ID             string            `json:"id"`
	MappingRule    string            `json:"mapping_rule"`
	Expected       string            `json:"expected"`
	Decision       string            `json:"decision"`
	Matches        []string          `json:"matches"`
	Contradictions []string          `json:"contradictions"`
	Missing        []string          `json:"missing"`
	Unknown        *Unknown          `json:"unknown,omitempty"`
	Evidence       ObservedEvidence `json:"evidence"`
}

type Projection struct {
	Project   string         `json:"project"`
	Authority Authority       `json:"authority"`
	Infra     InfraInput      `json:"infrastructure_input"`
	OpenAPI   OpenAPIInput    `json:"openapi_input"`
	Cases     []CaseResult    `json:"cases"`
	Sources   []ProjectSource `json:"sources"`
}

type ProjectSource struct {
	Project string       `json:"project"`
	Gooo    FileDigest   `json:"gooo"`
	HCL     []FileDigest `json:"hcl"`
	OpenAPI FileDigest   `json:"openapi"`
}

type Cell struct {
	ID        string `json:"id"`
	Proof     string `json:"proof"`
	Indicator string `json:"indicator"`
	Activity  string `json:"activity"`
}

type UnknownRecord struct {
	CaseID      string  `json:"case_id"`
	MappingRule string  `json:"mapping_rule"`
	Unknown     Unknown `json:"unknown"`
}

type Artifact struct {
	Name   string `json:"name"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type Inventory struct {
	Root                  string `json:"root"`
	DescendantFolders     int    `json:"descendant_folders"`
	DescendantFiles       int    `json:"descendant_files"`
	GoFiles               int    `json:"go_files"`
	GoPhysicalLines       int    `json:"go_physical_lines"`
	GoooFiles             int    `json:"gooo_files"`
	GoooPhysicalLines     int    `json:"gooo_physical_lines"`
	HCLFiles              int    `json:"hcl_files"`
	HCLPhysicalLines      int    `json:"hcl_physical_lines"`
	OpenAPIFiles          int    `json:"openapi_files"`
	OpenAPIPhysicalLines  int    `json:"openapi_physical_lines"`
	RootREADMEExcluded    bool   `json:"root_readme_excluded"`
}

type RuntimeMetrics struct {
	WallMS      int    `json:"wall_ms"`
	PeakRSSKiB  int    `json:"peak_rss_kib"`
	PeakRSSState string `json:"peak_rss_state"`
}

type MutationBoundary struct {
	RepositoryWrites            int  `json:"repository_writes"`
	LocalTestExecutions         int  `json:"local_test_executions"`
	CrossProjectRequiredGates   int  `json:"cross_project_required_gates"`
	SourceMutations             int  `json:"source_mutations"`
	LockfileMutations           int  `json:"lockfile_mutations"`
	NetworkProviderResolutions  int  `json:"network_provider_resolutions"`
	OpenTofuInitPlanApplyRuns   int  `json:"opentofu_init_plan_apply_runs"`
	CallerOwnedOutputOnly       bool `json:"caller_owned_output_only"`
}

type Status struct {
	State         string  `json:"state"`
	Reason        string  `json:"reason"`
	Unknown       Unknown `json:"unknown"`
	Before        *string `json:"before,omitempty"`
	After         *string `json:"after,omitempty"`
}

type Manifest struct {
	Schema              string            `json:"schema"`
	Product             string            `json:"product"`
	Release             string            `json:"release"`
	Project             string            `json:"project"`
	AuthorityScope      string            `json:"authority_scope"`
	InputBoundary       string            `json:"input_boundary"`
	OutputBoundary      string            `json:"output_boundary"`
	Source              SourceManifest    `json:"source"`
	SupportedScope      []string          `json:"supported_scope"`
	UnsupportedScope    []string          `json:"unsupported_scope"`
	Denominator         Denominator       `json:"denominator"`
	CaseCounts          map[string]int    `json:"case_counts"`
	Cases               []CaseResult      `json:"cases"`
	UnknownClaims       []UnknownRecord   `json:"unknown_claims"`
	GeneratedArtifacts  []Artifact        `json:"generated_artifacts"`
	GeneratedArtifactCount int             `json:"generated_artifact_count"`
	ArtifactDigestsExcludeManifest bool    `json:"artifact_digests_exclude_manifest"`
	Inventory           Inventory         `json:"inventory"`
	Metrics             RuntimeMetrics    `json:"metrics"`
	MutationBoundary    MutationBoundary `json:"mutation_boundary"`
	Improvement         Status            `json:"improvement"`
	ExternalUtility     Status            `json:"external_user_utility"`
	OperationalState    string            `json:"operational_state"`
}

type SourceManifest struct {
	Projects []ProjectSource `json:"projects"`
}

type Denominator struct {
	TotalCases      int              `json:"total_cases"`
	DecisionOrder   []string         `json:"decision_order"`
	ProofCells      []Cell           `json:"proof_cells"`
	IndicatorCells  []Cell           `json:"indicator_cells"`
	Activities      []MetaActivity   `json:"activities"`
	ProofCounts     map[string]int   `json:"proof_counts"`
	IndicatorCounts map[string]int   `json:"indicator_counts"`
}
