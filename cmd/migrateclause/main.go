package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/sysrqio/migrateclause/internal/audit"
)

func main() {
	os.Exit(run())
}

func run() int {
	auditCmd := flag.NewFlagSet("audit", flag.ExitOnError)
	manifest := auditCmd.String("manifest", "", "Path to contract switching manifest YAML")
	fixture := auditCmd.String("fixture", "", "Use built-in fixture: compliant|violations")
	output := auditCmd.String("output", "", "Write JSON report to file")
	format := auditCmd.String("format", "table", "Output format: table|json")

	if len(os.Args) < 2 {
		printUsage()
		return 1
	}

	switch os.Args[1] {
	case "audit":
		_ = auditCmd.Parse(os.Args[2:])
		path := *manifest
		if *fixture != "" {
			path = fixturePath(*fixture)
			if path == "" {
				fmt.Fprintf(os.Stderr, "unknown fixture %q\n", *fixture)
				return 1
			}
		}
		if path == "" {
			fmt.Fprintln(os.Stderr, "--manifest or --fixture is required")
			return 1
		}
		m, err := audit.LoadManifest(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		report := audit.Evaluate(m)
		if *output != "" {
			if err := writeJSON(*output, report); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}
		}
		switch *format {
		case "json":
			if err := printJSON(report); err != nil {
				return 1
			}
		default:
			printTable(report)
		}
		if !report.Passed {
			return 2
		}
		return 0
	case "version":
		fmt.Println("migrateclause 0.1.0 (data-act-switching-2026.10)")
		return 0
	default:
		printUsage()
		return 1
	}
}

func fixturePath(name string) string {
	switch name {
	case "compliant":
		return "fixtures/compliant.yaml"
	case "violations":
		return "fixtures/violations.yaml"
	default:
		return ""
	}
}

func printJSON(r audit.Report) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func writeJSON(path string, r audit.Report) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func printTable(r audit.Report) {
	status := "PASS"
	if !r.Passed {
		status = "FAIL"
	}
	fmt.Printf("migrateclause audit: %s (%s)\n", r.ContractID, status)
	if len(r.Findings) == 0 {
		fmt.Println("No findings.")
		return
	}
	for _, f := range r.Findings {
		fmt.Printf("  [%s] %s — %s (%s)\n", f.Severity, f.ID, f.Message, f.Article)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `migrateclause — EU Data Act cloud switching contract evidence CLI

Usage:
  migrateclause audit --manifest contract.yaml [--format json] [--output report.json]
  migrateclause audit --fixture compliant|violations
  migrateclause version

Exit codes: 0=pass, 1=error, 2=policy violation
`)
}
