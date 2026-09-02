package bridge

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func RunProject(projectDir string) (Projection, error) {
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return Projection{}, err
	}
	goooPath := filepath.Join(projectDir, "main.gooo")
	infraPath := filepath.Join(projectDir, "main.tf")
	openAPIPath := filepath.Join(projectDir, "openapi.yaml")
	authority, goooRaw, err := ParseGooo(goooPath)
	if err != nil {
		return Projection{}, err
	}
	infra, err := ReadBoundedOpenTofu(infraPath)
	if err != nil {
		return Projection{}, err
	}
	api, err := ReadBoundedOpenAPI(openAPIPath)
	if err != nil {
		return Projection{}, err
	}
	openAPIRaw, err := os.ReadFile(openAPIPath)
	if err != nil {
		return Projection{}, err
	}
	sources := []ProjectSource{{
		Project: authority.Project,
		Gooo:    digestFor(goooPath, goooRaw),
		HCL:     infra.Files,
		OpenAPI: digestFor(openAPIPath, openAPIRaw),
	}}
	return ProjectInputs(authority, infra, api, sources)
}

func RunSuite(root string) (Projection, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Projection{}, err
	}
	projectNames := []string{"closed", "unknown", "refuted"}
	combined := Authority{
		Schema:           Schema,
		Project:          "examples-suite",
		Infrastructure:   []InfrastructureClaim{},
		ServiceContracts: []ServiceContractClaim{},
		Mappings:         []MappingRule{},
		Evidence:         []EvidenceBinding{},
		Decisions:        []DriftDecision{},
		Cases:            []CaseDeclaration{},
		Activities:       []MetaActivity{},
	}
	projection := Projection{Project: combined.Project, Authority: combined, Cases: []CaseResult{}, Sources: []ProjectSource{}}
	projection.Infra.Resources = []InfraResource{}
	projection.Infra.Files = []FileDigest{}
	projection.OpenAPI.Operations = []OpenAPIOperation{}
	for index, projectName := range projectNames {
		project, err := RunProject(filepath.Join(root, projectName))
		if err != nil {
			return Projection{}, err
		}
		if index == 0 {
			combined.Activities = append([]MetaActivity(nil), project.Authority.Activities...)
			projection.OpenAPI.Version = project.OpenAPI.Version
		}
		combined.Infrastructure = append(combined.Infrastructure, project.Authority.Infrastructure...)
		combined.ServiceContracts = append(combined.ServiceContracts, project.Authority.ServiceContracts...)
		combined.Mappings = append(combined.Mappings, project.Authority.Mappings...)
		combined.Evidence = append(combined.Evidence, project.Authority.Evidence...)
		combined.Decisions = append(combined.Decisions, project.Authority.Decisions...)
		combined.Cases = append(combined.Cases, project.Authority.Cases...)
		projection.Infra.Resources = append(projection.Infra.Resources, project.Infra.Resources...)
		projection.Infra.Files = append(projection.Infra.Files, project.Infra.Files...)
		projection.OpenAPI.Operations = append(projection.OpenAPI.Operations, project.OpenAPI.Operations...)
		projection.OpenAPI.Lines += project.OpenAPI.Lines
		projection.Cases = append(projection.Cases, project.Cases...)
		projection.Sources = append(projection.Sources, project.Sources...)
	}
	if len(projection.Cases) != 12 {
		return Projection{}, fmt.Errorf("suite must contain exactly 12 cases, got %d", len(projection.Cases))
	}
	counts := decisionCounts(projection.Cases)
	if counts[DecisionClosed] != 4 || counts[DecisionUnknown] != 4 || counts[DecisionRefuted] != 4 {
		return Projection{}, fmt.Errorf("suite decision denominator must be 4/4/4, got %#v", counts)
	}
	projection.Authority = combined
	return projection, nil
}

func decisionCounts(cases []CaseResult) map[string]int {
	counts := map[string]int{DecisionClosed: 0, DecisionUnknown: 0, DecisionRefuted: 0}
	for _, result := range cases {
		counts[result.Decision]++
	}
	return counts
}

func GenerateArtifacts(projection Projection, inventory Inventory, outputDir string, elapsed time.Duration) (Manifest, error) {
	if err := EnsureOutputDirectory(outputDir, inventory.Root); err != nil {
		return Manifest{}, err
	}
	if len(projection.Authority.Activities) != 12 {
		return Manifest{}, errors.New("generated manifest requires the fixed 12 meta activities")
	}
	manifest, err := buildManifest(projection, inventory, elapsed)
	if err != nil {
		return Manifest{}, err
	}
	projectedRaw, err := JSON(projection)
	if err != nil {
		return Manifest{}, err
	}
	projectedRaw = append(projectedRaw, '\n')
	report := struct {
		Schema         string       `json:"schema"`
		Project        string       `json:"project"`
		DecisionOrder  []string     `json:"decision_order"`
		Cases          []CaseResult `json:"cases"`
	}{Schema: Schema, Project: projection.Project, DecisionOrder: []string{DecisionRefuted, DecisionUnknown, DecisionClosed}, Cases: projection.Cases}
	reportRaw, err := JSON(report)
	if err != nil {
		return Manifest{}, err
	}
	reportRaw = append(reportRaw, '\n')
	dossierRaw := []byte(renderDossier(projection, manifest))
	paths := []struct {
		name string
		raw  []byte
	}{
		{name: "projected-claims.json", raw: projectedRaw},
		{name: "mapping-report.json", raw: reportRaw},
		{name: "dossier.md", raw: dossierRaw},
	}
	for _, artifact := range paths {
		if err := writeOutputFile(filepath.Join(outputDir, artifact.name), artifact.raw); err != nil {
			return Manifest{}, err
		}
	}
	manifest.GeneratedArtifacts = []Artifact{}
	for _, artifact := range paths {
		raw, err := os.ReadFile(filepath.Join(outputDir, artifact.name))
		if err != nil {
			return Manifest{}, err
		}
		digest := digestFor(filepath.Join(outputDir, artifact.name), raw)
		manifest.GeneratedArtifacts = append(manifest.GeneratedArtifacts, Artifact{Name: artifact.name, Bytes: len(raw), SHA256: digest.SHA256})
	}
	manifest.GeneratedArtifactCount = len(paths) + 1
	manifest.ArtifactDigestsExcludeManifest = true
	manifestRaw, err := JSON(manifest)
	if err != nil {
		return Manifest{}, err
	}
	manifestRaw = append(manifestRaw, '\n')
	if err := writeOutputFile(filepath.Join(outputDir, "manifest.json"), manifestRaw); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func buildManifest(projection Projection, inventory Inventory, elapsed time.Duration) (Manifest, error) {
	caseCounts := decisionCounts(projection.Cases)
	proofCounts := map[string]int{FixedProofCells[0]: 0, FixedProofCells[1]: 0, FixedProofCells[2]: 0}
	indicatorCounts := map[string]int{FixedIndicatorCells[0]: 0, FixedIndicatorCells[1]: 0, FixedIndicatorCells[2]: 0}
	for _, activity := range projection.Authority.Activities {
		proofCounts[activity.Proof]++
		indicatorCounts[activity.Indicator]++
	}
	if proofCounts[FixedProofCells[0]] != 4 || proofCounts[FixedProofCells[1]] != 4 || proofCounts[FixedProofCells[2]] != 4 || indicatorCounts[FixedIndicatorCells[0]] != 4 || indicatorCounts[FixedIndicatorCells[1]] != 4 || indicatorCounts[FixedIndicatorCells[2]] != 4 {
		return Manifest{}, errors.New("proof and indicator cells must each be 4/4/4")
	}
	proofCells, indicatorCells := makeCells(projection.Authority.Activities)
	unknownClaims := []UnknownRecord{}
	for _, result := range projection.Cases {
		if result.Decision == DecisionUnknown {
			if result.Unknown == nil || !result.Unknown.Valid() {
				return Manifest{}, fmt.Errorf("unknown case %s lacks all six unknown fields", result.ID)
			}
			unknownClaims = append(unknownClaims, UnknownRecord{CaseID: result.ID, MappingRule: result.MappingRule, Unknown: *result.Unknown})
		}
	}
	peakRSS, rssState := peakRSSKiB()
	return Manifest{
		Schema:         Schema,
		Product:        "Gooo OpenTofu Service Contract Bridge",
		Release:        Version,
		Project:        projection.Project,
		AuthorityScope: "GOOO_CLAIMS_ONLY",
		InputBoundary:  "READ_ONLY_SOURCE_REPOSITORY",
		OutputBoundary: "ABSOLUTE_CALLER_OWNED_EMPTY_DIRECTORY_OUTSIDE_SOURCE_REPOSITORY",
		Source:         SourceManifest{Projects: projection.Sources},
		SupportedScope: []string{
			".gooo InfrastructureClaim, ServiceContractClaim, MappingRule, EvidenceBinding, DriftDecision authority",
			"fixed 12 meta activities with proof FOUNDATION/COHERENCE/REGRESSION 4/4/4",
			"fixed 12 meta activities with indicator DRIVER/OUTCOME/GUARDRAIL 4/4/4",
			"OpenTofu resource blocks with two labels and literal single-line arguments",
			"OpenAPI 3.x JSON or bounded YAML paths and HTTP operations",
			"CLOSED, UNKNOWN, REFUTED dossier classification with REFUTED > UNKNOWN > CLOSED",
		},
		UnsupportedScope: []string{
			"Terraform identity or semantic equivalence",
			"OpenTofu module, provider, data, variable, output, locals, nested blocks, expressions, JSON configuration, or full HCL",
			"tofu init, validate, plan, apply, destroy, state, provider installation, or network resolution",
			"source, lockfile, .git, backend, credentials, or cloud mutation",
			"full OpenAPI validation, components, references, schemas, servers, security, webhooks, and callbacks",
			"aggregate roll-up outputs",
			"improvement without a matched before/after observation pair",
			"external user utility without a user observation",
		},
		Denominator: Denominator{
			TotalCases:      len(projection.Cases),
			DecisionOrder:   []string{DecisionRefuted, DecisionUnknown, DecisionClosed},
			ProofCells:      proofCells,
			IndicatorCells:  indicatorCells,
			Activities:      projection.Authority.Activities,
			ProofCounts:     proofCounts,
			IndicatorCounts: indicatorCounts,
		},
		CaseCounts:             caseCounts,
		Cases:                  projection.Cases,
		UnknownClaims:          unknownClaims,
		GeneratedArtifacts:     []Artifact{},
		GeneratedArtifactCount: 4,
		ArtifactDigestsExcludeManifest: true,
		Inventory:              inventory,
		Metrics:                RuntimeMetrics{WallMS: elapsed.Milliseconds(), PeakRSSKiB: peakRSS, PeakRSSState: rssState},
		MutationBoundary: MutationBoundary{
			RepositoryWrites: 0, LocalTestExecutions: 0, CrossProjectRequiredGates: 0,
			SourceMutations: 0, LockfileMutations: 0, NetworkProviderResolutions: 0,
			OpenTofuInitPlanApplyRuns: 0, CallerOwnedOutputOnly: true,
		},
		Improvement: Status{
			State:  DecisionUnknown,
			Reason: "no matched before/after observation pair exists",
			Unknown: Unknown{Stage: "evidence", Step: "compare", Reason: "no matched before/after observation pair exists", UnknownClass: "NO_BEFORE_AFTER_PAIR", NextOperation: "capture a matched before and after observation", BlockedBy: []string{"before_observation", "after_observation"}},
		},
		ExternalUtility: Status{
			State:  DecisionUnknown,
			Reason: "no external user observation exists",
			Unknown: Unknown{Stage: "evidence", Step: "observe-user-utility", Reason: "no external user observation exists", UnknownClass: "NO_USER_OBSERVATION", NextOperation: "collect a user utility observation", BlockedBy: []string{"external_user_observation"}},
		},
		OperationalState: "SUCCESSFUL_READ_ONLY_PROJECTION",
	}, nil
}

func makeCells(activities []MetaActivity) ([]Cell, []Cell) {
	proofCells := make([]Cell, 0, len(activities))
	indicatorCells := make([]Cell, 0, len(activities))
	for index, activity := range activities {
		proofCells = append(proofCells, Cell{ID: fmt.Sprintf("proof-cell-%02d", index+1), Proof: activity.Proof, Indicator: "", Activity: activity.Name})
		indicatorCells = append(indicatorCells, Cell{ID: fmt.Sprintf("indicator-cell-%02d", index+1), Proof: "", Indicator: activity.Indicator, Activity: activity.Name})
	}
	return proofCells, indicatorCells
}

func renderDossier(projection Projection, manifest Manifest) string {
	var builder strings.Builder
	builder.WriteString("# Gooo OpenTofu Service Contract Bridge dossier\n\n")
	builder.WriteString("This is a read-only projection from `.gooo` semantic authority, bounded OpenTofu observations, and a bounded OpenAPI observation. It does not run OpenTofu and does not claim Terraform equivalence.\n\n")
	fmt.Fprintf(&builder, "- Project: `%s`\n- Release schema: `%s`\n- Authority scope: `%s`\n- Generated artifact count: `%d`\n- Wall time: `%d ms`\n- Peak RSS: `%d KiB` (%s)\n", projection.Project, Schema, manifest.AuthorityScope, len(projection.Cases)+3, manifest.Metrics.WallMS, manifest.Metrics.PeakRSSKiB, manifest.Metrics.PeakRSSState)
	builder.WriteString("\n## Decision dossier\n\n")
	builder.WriteString("| Case | Expected | Observed | Evidence / unknown detail |\n| --- | --- | --- | --- |\n")
	for _, result := range projection.Cases {
		detail := strings.Join(result.Contradictions, "; ")
		if detail == "" {
			detail = strings.Join(result.Matches, ", ")
		}
		if result.Unknown != nil {
			detail = fmt.Sprintf("%s; next=%s; blocked_by=%s", result.Unknown.UnknownClass, result.Unknown.NextOperation, strings.Join(result.Unknown.BlockedBy, ","))
		}
		fmt.Fprintf(&builder, "| `%s` | `%s` | `%s` | %s |\n", result.ID, result.Expected, result.Decision, detail)
	}
	builder.WriteString("\nKnown contradiction is REFUTED before UNKNOWN and CLOSED. Every UNKNOWN retains stage, step, reason, unknown_class, next_operation, and blocked_by.\n\n")
	builder.WriteString("## Fixed denominator\n\n")
	fmt.Fprintf(&builder, "Cases: %d. CLOSED=%d, UNKNOWN=%d, REFUTED=%d. Proof cells: FOUNDATION=%d, COHERENCE=%d, REGRESSION=%d. Indicator cells: DRIVER=%d, OUTCOME=%d, GUARDRAIL=%d.\n\n", manifest.Denominator.TotalCases, manifest.CaseCounts[DecisionClosed], manifest.CaseCounts[DecisionUnknown], manifest.CaseCounts[DecisionRefuted], manifest.Denominator.ProofCounts["FOUNDATION"], manifest.Denominator.ProofCounts["COHERENCE"], manifest.Denominator.ProofCounts["REGRESSION"], manifest.Denominator.IndicatorCounts["DRIVER"], manifest.Denominator.IndicatorCounts["OUTCOME"], manifest.Denominator.IndicatorCounts["GUARDRAIL"])
	builder.WriteString("No aggregate score or percentage is emitted. Improvement and external user utility remain UNKNOWN because no qualifying observations exist.\n\n")
	builder.WriteString("## Boundary and measurements\n\n")
	fmt.Fprintf(&builder, "The input repository is read-only. Runtime repository writes=%d, local test executions=%d, cross-project required gates=%d, OpenTofu init/plan/apply runs=%d, network provider resolutions=%d. Root README is excluded from inventory.\n\n", manifest.MutationBoundary.RepositoryWrites, manifest.MutationBoundary.LocalTestExecutions, manifest.MutationBoundary.CrossProjectRequiredGates, manifest.MutationBoundary.OpenTofuInitPlanApplyRuns, manifest.MutationBoundary.NetworkProviderResolutions)
	fmt.Fprintf(&builder, "Inventory: folders=%d, files=%d, Go files/lines=%d/%d, `.gooo` files/lines=%d/%d, HCL files/lines=%d/%d, OpenAPI files/lines=%d/%d.\n\n", manifest.Inventory.DescendantFolders, manifest.Inventory.DescendantFiles, manifest.Inventory.GoFiles, manifest.Inventory.GoPhysicalLines, manifest.Inventory.GoooFiles, manifest.Inventory.GoooPhysicalLines, manifest.Inventory.HCLFiles, manifest.Inventory.HCLPhysicalLines, manifest.Inventory.OpenAPIFiles, manifest.Inventory.OpenAPIPhysicalLines)
	builder.WriteString("## Official basis\n\n")
	builder.WriteString("- [OpenTofu Configuration Syntax](https://opentofu.org/docs/language/syntax/configuration/)\n- [OpenTofu Language Documentation](https://opentofu.org/docs/v1.10/language/)\n- [OpenTofu official GitHub repository](https://github.com/opentofu/opentofu)\n- [OpenAPI Specification v3.1.0](https://spec.openapis.org/oas/v3.1.0)\n- [OpenAPI Specification repository](https://github.com/OAI/OpenAPI-Specification)\n")
	return builder.String()
}

func EnsureOutputDirectory(path, repositoryRoot string) error {
	if !filepath.IsAbs(path) {
		return errors.New("output directory must be an absolute caller-owned path")
	}
	path = filepath.Clean(path)
	if repositoryRoot != "" {
		root, err := filepath.Abs(repositoryRoot)
		if err != nil {
			return err
		}
		if isWithin(root, path) {
			return errors.New("output directory must be outside the source repository")
		}
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("output path must be a directory")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return errors.New("caller-owned output directory must be empty")
	}
	return nil
}

func isWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func WriteOperationalRefuted(outputDir string, operationErr error) {
	if outputDir == "" || !filepath.IsAbs(outputDir) || operationErr == nil {
		return
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return
	}
	path := filepath.Join(outputDir, "operational-refuted.json")
	if _, err := os.Stat(path); err == nil {
		return
	}
	record := fmt.Sprintf("{\n  \"schema\": %q,\n  \"state\": \"OPERATIONAL_REFUTED\",\n  \"reason\": %q,\n  \"preceding_artifacts_preserved\": true\n}\n", Schema, operationErr.Error())
	_ = os.WriteFile(path, []byte(record), 0o644)
}

func SortSources(sources []ProjectSource) {
	sort.Slice(sources, func(i, j int) bool { return sources[i].Project < sources[j].Project })
}
