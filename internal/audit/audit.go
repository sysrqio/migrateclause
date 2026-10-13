package audit

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Manifest is a structured checklist derived from a cloud processing contract
// (manual legal review or SCC mapping). See EU Data Act Art. 25–30.
type Manifest struct {
	ContractID     string         `yaml:"contract_id"`
	Provider       string         `yaml:"provider,omitempty"`
	Customer       string         `yaml:"customer,omitempty"`
	AssessedAt     string         `yaml:"assessed_at,omitempty"`
	Switching      SwitchingBlock `yaml:"switching"`
	Interoperability InteropBlock `yaml:"interoperability,omitempty"`
}

type SwitchingBlock struct {
	MaxNoticePeriodDays              int  `yaml:"max_notice_period_days"`
	MaxTransitionPeriodDays          int  `yaml:"max_transition_period_days"`
	MinDataRetrievalDays             int  `yaml:"min_data_retrieval_days"`
	SpecifiesExportableDataSchedule  bool `yaml:"specifies_exportable_data_schedule"`
	IncludesTerminationAssistance    bool `yaml:"includes_termination_assistance"`
	ProhibitsSwitchingFeesAfter2027  bool `yaml:"prohibits_switching_fees_after_2027"`
	AllowsCostBasedSwitchingFeesNow  bool `yaml:"allows_cost_based_switching_fees_now"`
}

type InteropBlock struct {
	OpenInterfacesForSaaS bool `yaml:"open_interfaces_for_saas"`
}

type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Article  string `json:"article"`
}

type Report struct {
	ContractID string    `json:"contract_id"`
	Provider   string    `json:"provider,omitempty"`
	Passed     bool      `json:"passed"`
	Findings   []Finding `json:"findings"`
	AssessedAt string    `json:"assessed_at"`
}

const (
	maxNoticeDays     = 60
	maxTransitionDays = 30
	minRetrievalDays  = 30
)

func LoadManifest(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.ContractID == "" {
		return nil, fmt.Errorf("contract_id is required")
	}
	return &m, nil
}

func Evaluate(m *Manifest) Report {
	findings := []Finding{}
	add := func(id, sev, msg, art string) {
		findings = append(findings, Finding{ID: id, Severity: sev, Message: msg, Article: art})
	}

	if m.Switching.MaxNoticePeriodDays <= 0 {
		add("SW-001", "critical", "max_notice_period_days must be set", "Art. 25(2)(d)")
	} else if m.Switching.MaxNoticePeriodDays > maxNoticeDays {
		add("SW-002", "critical",
			fmt.Sprintf("notice period %d days exceeds Data Act cap of %d days", m.Switching.MaxNoticePeriodDays, maxNoticeDays),
			"Art. 25(2)(d)")
	}

	if m.Switching.MaxTransitionPeriodDays <= 0 {
		add("SW-003", "critical", "max_transition_period_days must be set", "Art. 25(2)(a)")
	} else if m.Switching.MaxTransitionPeriodDays > maxTransitionDays {
		add("SW-004", "critical",
			fmt.Sprintf("transition period %d days exceeds mandatory maximum of %d calendar days", m.Switching.MaxTransitionPeriodDays, maxTransitionDays),
			"Art. 25(2)(a)")
	}

	if m.Switching.MinDataRetrievalDays < minRetrievalDays {
		add("SW-005", "critical",
			fmt.Sprintf("data retrieval period %d days is below minimum %d days", m.Switching.MinDataRetrievalDays, minRetrievalDays),
			"Art. 25(2)(g)")
	}

	if !m.Switching.SpecifiesExportableDataSchedule {
		add("SW-006", "high", "contract must specify exportable data and digital assets schedule", "Art. 25 / Art. 34")
	}
	if !m.Switching.IncludesTerminationAssistance {
		add("SW-007", "high", "termination / switching assistance clauses missing", "Art. 25(2)(a)")
	}
	if !m.Switching.ProhibitsSwitchingFeesAfter2027 {
		add("SW-008", "high", "must prohibit switching/egress charges from 2027-01-12 (Art. 29)", "Art. 29(1)")
	}

	if m.Interoperability.OpenInterfacesForSaaS {
		// optional field; only flag when explicitly false on SaaS contracts
	} else if m.Interoperability != (InteropBlock{}) && !m.Interoperability.OpenInterfacesForSaaS {
		add("IO-001", "medium", "SaaS contracts should document open interfaces per Art. 30", "Art. 30")
	}

	assessed := m.AssessedAt
	if assessed == "" {
		assessed = time.Now().UTC().Format(time.RFC3339)
	}

	passed := true
	for _, f := range findings {
		if f.Severity == "critical" || f.Severity == "high" {
			passed = false
			break
		}
	}

	return Report{
		ContractID: m.ContractID,
		Provider:   m.Provider,
		Passed:     passed,
		Findings:   findings,
		AssessedAt: assessed,
	}
}
