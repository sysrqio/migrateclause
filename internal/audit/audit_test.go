package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvaluateCompliantFixture(t *testing.T) {
	m, err := LoadManifest(filepath.Join("..", "..", "fixtures", "compliant.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	r := Evaluate(m)
	if !r.Passed {
		t.Fatalf("expected pass, got findings: %+v", r.Findings)
	}
}

func TestEvaluateViolationsFixture(t *testing.T) {
	m, err := LoadManifest(filepath.Join("..", "..", "fixtures", "violations.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	r := Evaluate(m)
	if r.Passed {
		t.Fatal("expected violations")
	}
	if len(r.Findings) < 2 {
		t.Fatalf("expected multiple findings, got %d", len(r.Findings))
	}
}

func TestLoadManifestMissingID(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(f, []byte("switching: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadManifest(f)
	if err == nil {
		t.Fatal("expected error for missing contract_id")
	}
}
