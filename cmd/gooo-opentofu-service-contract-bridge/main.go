package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bridge "github.com/kimjooyoon/gooo-opentofu-service-contract-bridge/internal/bridge"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	goooPath := flags.String("gooo", "", "path to one project main.gooo")
	infraPath := flags.String("infra", "", "path to one OpenTofu file or directory")
	contractPath := flags.String("contract", "", "path to one OpenAPI YAML or JSON document")
	suiteRoot := flags.String("suite-root", "", "directory containing closed, unknown, and refuted projects")
	outputDir := flags.String("output", "", "absolute empty caller-owned output directory")
	inventoryRoot := flags.String("inventory-root", "", "source repository root to inventory")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	if *outputDir == "" {
		return fmt.Errorf("-output is required")
	}
	if (*suiteRoot == "" && (*goooPath == "" || *infraPath == "" || *contractPath == "")) || (*suiteRoot != "" && (*goooPath != "" || *infraPath != "" || *contractPath != "")) {
		return fmt.Errorf("provide either -suite-root or all of -gooo, -infra, and -contract")
	}
	start := time.Now()
	var projection bridge.Projection
	var err error
	if *suiteRoot != "" {
		projection, err = bridge.RunSuite(*suiteRoot)
	} else {
		projection, err = runSingle(*goooPath, *infraPath, *contractPath)
	}
	if err != nil {
		bridge.WriteOperationalRefuted(*outputDir, err)
		return err
	}
	root := *inventoryRoot
	if root == "" {
		root, err = os.Getwd()
		if err != nil {
			bridge.WriteOperationalRefuted(*outputDir, err)
			return err
		}
	}
	inventory, err := bridge.BuildInventory(root)
	if err != nil {
		bridge.WriteOperationalRefuted(*outputDir, err)
		return err
	}
	manifest, err := bridge.GenerateArtifacts(projection, inventory, *outputDir, time.Since(start))
	if err != nil {
		bridge.WriteOperationalRefuted(*outputDir, err)
		return err
	}
	fmt.Printf("generated %d artifacts for %s: %s\n", manifest.GeneratedArtifactCount, manifest.Project, filepath.ToSlash(*outputDir))
	return nil
}

func runSingle(goooPath, infraPath, contractPath string) (bridge.Projection, error) {
	authority, goooRaw, err := bridge.ParseGooo(goooPath)
	if err != nil {
		return bridge.Projection{}, err
	}
	infra, err := bridge.ReadBoundedOpenTofu(infraPath)
	if err != nil {
		return bridge.Projection{}, err
	}
	api, err := bridge.ReadBoundedOpenAPI(contractPath)
	if err != nil {
		return bridge.Projection{}, err
	}
	openAPIRaw, err := os.ReadFile(contractPath)
	if err != nil {
		return bridge.Projection{}, err
	}
	sources := []bridge.ProjectSource{{
		Project: authority.Project,
		Gooo:    bridge.DigestFor(goooPath, goooRaw),
		HCL:     infra.Files,
		OpenAPI: bridge.DigestFor(contractPath, openAPIRaw),
	}}
	return bridge.ProjectInputs(authority, infra, api, sources)
}
