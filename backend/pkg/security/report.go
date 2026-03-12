package security

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	ReportVersion = "1.0.0"
	ReportFormat  = "pentagi-security-report"
)

type ReportSeverity string

const (
	SeverityCritical ReportSeverity = "critical"
	SeverityHigh     ReportSeverity = "high"
	SeverityMedium   ReportSeverity = "medium"
	SeverityLow      ReportSeverity = "low"
	SeverityInfo     ReportSeverity = "info"
)

type ReportStatus string

const (
	StatusCompleted ReportStatus = "completed"
	StatusFailed    ReportStatus = "failed"
	StatusPartial   ReportStatus = "partial"
	StatusRunning   ReportStatus = "running"
	StatusPending   ReportStatus = "pending"
)

type VulnerabilityCategory string

const (
	CategoryInjection      VulnerabilityCategory = "injection"
	CategoryAuth           VulnerabilityCategory = "authentication"
	CategoryAccessControl  VulnerabilityCategory = "access_control"
	CategoryCrypto         VulnerabilityCategory = "cryptography"
	CategoryDataExposure   VulnerabilityCategory = "data_exposure"
	CategoryDoS            VulnerabilityCategory = "denial_of_service"
	CategoryMisconfig      VulnerabilityCategory = "misconfiguration"
	CategoryXSS            VulnerabilityCategory = "xss"
	CategoryCSRF           VulnerabilityCategory = "csrf"
	CategorySSRF           VulnerabilityCategory = "ssrf"
	CategoryLFI            VulnerabilityCategory = "lfi"
	CategoryRCE            VulnerabilityCategory = "rce"
	CategoryInfoDisclosure VulnerabilityCategory = "information_disclosure"
	CategoryOther          VulnerabilityCategory = "other"
)

type CVSSScore struct {
	Vector        string  `json:"vector,omitempty"`
	BaseScore     float64 `json:"base_score"`
	ImpactScore   float64 `json:"impact_score,omitempty"`
	Exploitability float64 `json:"exploitability_score,omitempty"`
}

type Evidence struct {
	Type        string            `json:"type"`
	Content     string            `json:"content"`
	ContentType string            `json:"content_type"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type Remediation struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	References  []string `json:"references,omitempty"`
	Priority    int      `json:"priority"`
}

type Vulnerability struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	Category         VulnerabilityCategory `json:"category"`
	Severity         ReportSeverity      `json:"severity"`
	CVSS            *CVSSScore          `json:"cvss,omitempty"`
	AffectedHost     string              `json:"affected_host"`
	AffectedPort     int                 `json:"affected_port,omitempty"`
	AffectedPath     string              `json:"affected_path,omitempty"`
	AffectedService  string              `json:"affected_service,omitempty"`
	Evidence         []Evidence          `json:"evidence,omitempty"`
	Remediation      []Remediation       `json:"remediation,omitempty"`
	References       []string            `json:"references,omitempty"`
	CWE              string              `json:"cwe,omitempty"`
	CVE              string              `json:"cve,omitempty"`
	FirstSeen       time.Time           `json:"first_seen"`
	LastSeen        time.Time           `json:"last_seen"`
	Confirmed       bool                `json:"confirmed"`
	FalsePositive   bool                `json:"false_positive,omitempty"`
	Metadata        map[string]string   `json:"metadata,omitempty"`
}

type HostInfo struct {
	Host          string            `json:"host"`
	IPAddresses   []string          `json:"ip_addresses,omitempty"`
	Ports         []int             `json:"ports,omitempty"`
	Services      map[int]string    `json:"services,omitempty"`
	OS            string            `json:"os,omitempty"`
	OSAccuracy    int               `json:"os_accuracy,omitempty"`
	Hostname      string            `json:"hostname,omitempty"`
	MacAddress    string            `json:"mac_address,omitempty"`
	LastSeen      time.Time         `json:"last_seen"`
	Status        string            `json:"status"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type ToolExecution struct {
	ToolName       string            `json:"tool_name"`
	ToolVersion    string            `json:"tool_version,omitempty"`
	Command        string            `json:"command,omitempty"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        time.Time         `json:"end_time"`
	Duration       float64           `json:"duration_seconds"`
	ExitCode       int               `json:"exit_code"`
	Success        bool              `json:"success"`
	OutputSize     int64             `json:"output_size_bytes"`
	Error          string            `json:"error,omitempty"`
	SandboxUsed    bool              `json:"sandbox_used"`
	SandboxType    string            `json:"sandbox_type,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type Finding struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Severity       ReportSeverity    `json:"severity"`
	Confidence     int               `json:"confidence"`
	Host           string            `json:"host,omitempty"`
	Port           int               `json:"port,omitempty"`
	Data           json.RawMessage   `json:"data,omitempty"`
	Source         string            `json:"source"`
	Timestamp      time.Time         `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type SecurityReport struct {
	ReportMeta    ReportMetadata    `json:"metadata"`
	Summary       ReportSummary     `json:"summary"`
	Scope         ReportScope       `json:"scope"`
	Hosts         []HostInfo        `json:"hosts,omitempty"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
	Findings      []Finding         `json:"findings,omitempty"`
	ToolExecutions []ToolExecution  `json:"tool_executions,omitempty"`
	Timeline      []TimelineEvent   `json:"timeline,omitempty"`
	Recommendations []string        `json:"recommendations,omitempty"`
	Conclusion    string            `json:"conclusion,omitempty"`
	Signatures    []Signature       `json:"signatures,omitempty"`
}

type ReportMetadata struct {
	Version       string    `json:"version"`
	Format        string    `json:"format"`
	Generator     string    `json:"generator"`
	GeneratedAt   time.Time `json:"generated_at"`
	ReportID      string    `json:"report_id"`
	FlowID        int64     `json:"flow_id,omitempty"`
	TaskID        int64     `json:"task_id,omitempty"`
	SubtaskID     int64     `json:"subtask_id,omitempty"`
	SchemaVersion string    `json:"schema_version"`
}

type ReportSummary struct {
	Status              ReportStatus `json:"status"`
	TotalHosts          int          `json:"total_hosts"`
	TotalPorts          int          `json:"total_ports"`
	TotalVulnerabilities int         `json:"total_vulnerabilities"`
	CriticalCount       int          `json:"critical_count"`
	HighCount           int          `json:"high_count"`
	MediumCount         int          `json:"medium_count"`
	LowCount            int          `json:"low_count"`
	InfoCount           int          `json:"info_count"`
	StartTime           time.Time    `json:"start_time"`
	EndTime             time.Time    `json:"end_time,omitempty"`
	Duration            float64      `json:"duration_seconds"`
	SuccessRate         float64      `json:"success_rate"`
	ToolsUsed           []string     `json:"tools_used"`
}

type ReportScope struct {
	TargetHosts     []string `json:"target_hosts,omitempty"`
	TargetPorts     []int    `json:"target_ports,omitempty"`
	ExcludedHosts   []string `json:"excluded_hosts,omitempty"`
	ScanType        string   `json:"scan_type"`
	Authorized      bool     `json:"authorized"`
	AuthorizationRef string  `json:"authorization_ref,omitempty"`
	TestWindow      string   `json:"test_window,omitempty"`
	RulesOfEngagement string `json:"rules_of_engagement,omitempty"`
}

type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	Source      string    `json:"source,omitempty"`
	Severity    string    `json:"severity,omitempty"`
}

type Signature struct {
	Signer    string    `json:"signer"`
	Timestamp time.Time `json:"timestamp"`
	Hash      string    `json:"hash"`
	Algorithm string    `json:"algorithm"`
}

type ReportBuilder struct {
	report *SecurityReport
}

func NewReportBuilder(flowID int64) *ReportBuilder {
	now := time.Now()
	return &ReportBuilder{
		report: &SecurityReport{
			ReportMeta: ReportMetadata{
				Version:       ReportVersion,
				Format:        ReportFormat,
				Generator:     "PentAGI Security Scanner",
				GeneratedAt:   now,
				ReportID:      generateReportID(),
				FlowID:        flowID,
				SchemaVersion: "2024-01",
			},
			Summary: ReportSummary{
				Status:    StatusRunning,
				StartTime: now,
				ToolsUsed: []string{},
			},
			Hosts:           []HostInfo{},
			Vulnerabilities: []Vulnerability{},
			Findings:        []Finding{},
			ToolExecutions:  []ToolExecution{},
			Timeline:        []TimelineEvent{},
			Recommendations: []string{},
		},
	}
}

func (b *ReportBuilder) SetTask(taskID int64) *ReportBuilder {
	b.report.ReportMeta.TaskID = taskID
	return b
}

func (b *ReportBuilder) SetSubtask(subtaskID int64) *ReportBuilder {
	b.report.ReportMeta.SubtaskID = subtaskID
	return b
}

func (b *ReportBuilder) SetScope(scope ReportScope) *ReportBuilder {
	b.report.Scope = scope
	return b
}

func (b *ReportBuilder) SetScanType(scanType string) *ReportBuilder {
	b.report.Scope.ScanType = scanType
	return b
}

func (b *ReportBuilder) AddHost(host HostInfo) *ReportBuilder {
	b.report.Hosts = append(b.report.Hosts, host)
	b.report.Summary.TotalHosts = len(b.report.Hosts)
	b.report.Summary.TotalPorts += len(host.Ports)
	b.addTimelineEvent("host_discovered", fmt.Sprintf("Host discovered: %s", host.Host), "info")
	return b
}

func (b *ReportBuilder) AddVulnerability(vuln Vulnerability) *ReportBuilder {
	if vuln.FirstSeen.IsZero() {
		vuln.FirstSeen = time.Now()
	}
	if vuln.LastSeen.IsZero() {
		vuln.LastSeen = time.Now()
	}
	if vuln.ID == "" {
		vuln.ID = generateVulnerabilityID(vuln.Name, vuln.AffectedHost)
	}

	b.report.Vulnerabilities = append(b.report.Vulnerabilities, vuln)
	b.report.Summary.TotalVulnerabilities = len(b.report.Vulnerabilities)
	b.updateSeverityCounts()
	b.addTimelineEvent("vulnerability_found", fmt.Sprintf("Vulnerability found: %s on %s", vuln.Name, vuln.AffectedHost), string(vuln.Severity))
	return b
}

func (b *ReportBuilder) AddFinding(finding Finding) *ReportBuilder {
	if finding.Timestamp.IsZero() {
		finding.Timestamp = time.Now()
	}
	if finding.ID == "" {
		finding.ID = generateFindingID(finding.Type, finding.Title)
	}
	b.report.Findings = append(b.report.Findings, finding)
	b.addTimelineEvent("finding", fmt.Sprintf("Finding: %s", finding.Title), string(finding.Severity))
	return b
}

func (b *ReportBuilder) AddToolExecution(exec ToolExecution) *ReportBuilder {
	if exec.StartTime.IsZero() {
		exec.StartTime = time.Now()
	}
	if exec.EndTime.IsZero() {
		exec.EndTime = time.Now()
	}
	exec.Duration = exec.EndTime.Sub(exec.StartTime).Seconds()

	b.report.ToolExecutions = append(b.report.ToolExecutions, exec)
	b.report.Summary.ToolsUsed = appendUnique(b.report.Summary.ToolsUsed, exec.ToolName)
	return b
}

func (b *ReportBuilder) AddRecommendation(recommendation string) *ReportBuilder {
	b.report.Recommendations = append(b.report.Recommendations, recommendation)
	return b
}

func (b *ReportBuilder) SetConclusion(conclusion string) *ReportBuilder {
	b.report.Conclusion = conclusion
	return b
}

func (b *ReportBuilder) SetStatus(status ReportStatus) *ReportBuilder {
	b.report.Summary.Status = status
	return b
}

func (b *ReportBuilder) Finalize() *SecurityReport {
	now := time.Now()
	b.report.ReportMeta.GeneratedAt = now
	b.report.Summary.EndTime = now
	b.report.Summary.Duration = now.Sub(b.report.Summary.StartTime).Seconds()

	b.calculateSuccessRate()

	b.addTimelineEvent("report_completed", "Security report generation completed", "info")

	return b.report
}

func (b *ReportBuilder) updateSeverityCounts() {
	b.report.Summary.CriticalCount = 0
	b.report.Summary.HighCount = 0
	b.report.Summary.MediumCount = 0
	b.report.Summary.LowCount = 0
	b.report.Summary.InfoCount = 0

	for _, v := range b.report.Vulnerabilities {
		switch v.Severity {
		case SeverityCritical:
			b.report.Summary.CriticalCount++
		case SeverityHigh:
			b.report.Summary.HighCount++
		case SeverityMedium:
			b.report.Summary.MediumCount++
		case SeverityLow:
			b.report.Summary.LowCount++
		case SeverityInfo:
			b.report.Summary.InfoCount++
		}
	}
}

func (b *ReportBuilder) calculateSuccessRate() {
	if len(b.report.ToolExecutions) == 0 {
		b.report.Summary.SuccessRate = 0
		return
	}

	successCount := 0
	for _, exec := range b.report.ToolExecutions {
		if exec.Success {
			successCount++
		}
	}
	b.report.Summary.SuccessRate = float64(successCount) / float64(len(b.report.ToolExecutions)) * 100
}

func (b *ReportBuilder) addTimelineEvent(eventType, description, severity string) {
	b.report.Timeline = append(b.report.Timeline, TimelineEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		Description: description,
		Severity:    severity,
	})
}

func (r *SecurityReport) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}
	return string(data), nil
}

func (r *SecurityReport) ToJSONCompact() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}
	return string(data), nil
}

func (r *SecurityReport) GetSummaryJSON() (string, error) {
	summary := struct {
		Metadata  ReportMetadata `json:"metadata"`
		Summary   ReportSummary  `json:"summary"`
		Conclusion string        `json:"conclusion,omitempty"`
	}{
		Metadata:   r.ReportMeta,
		Summary:    r.Summary,
		Conclusion: r.Conclusion,
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal summary: %w", err)
	}
	return string(data), nil
}

func ParseReport(jsonData string) (*SecurityReport, error) {
	var report SecurityReport
	if err := json.Unmarshal([]byte(jsonData), &report); err != nil {
		return nil, fmt.Errorf("failed to parse report: %w", err)
	}
	return &report, nil
}

func generateReportID() string {
	return fmt.Sprintf("RPT-%d-%s", time.Now().Unix(), randomString(8))
}

func generateVulnerabilityID(name, host string) string {
	return fmt.Sprintf("VULN-%s-%s-%s", sanitizeID(name), sanitizeID(host), randomString(6))
}

func generateFindingID(findingType, title string) string {
	return fmt.Sprintf("FIND-%s-%s-%s", sanitizeID(findingType), sanitizeID(title), randomString(6))
}

func sanitizeID(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		}
	}
	if len(result) > 20 {
		result = result[:20]
	}
	return result
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func SeverityFromString(s string) ReportSeverity {
	switch s {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	default:
		return SeverityInfo
	}
}

func CategoryFromString(s string) VulnerabilityCategory {
	switch s {
	case "injection":
		return CategoryInjection
	case "authentication", "auth":
		return CategoryAuth
	case "access_control":
		return CategoryAccessControl
	case "cryptography", "crypto":
		return CategoryCrypto
	case "data_exposure":
		return CategoryDataExposure
	case "denial_of_service", "dos":
		return CategoryDoS
	case "misconfiguration", "misconfig":
		return CategoryMisconfig
	case "xss":
		return CategoryXSS
	case "csrf":
		return CategoryCSRF
	case "ssrf":
		return CategorySSRF
	case "lfi":
		return CategoryLFI
	case "rce":
		return CategoryRCE
	case "information_disclosure", "info_disclosure":
		return CategoryInfoDisclosure
	default:
		return CategoryOther
	}
}
