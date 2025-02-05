package api_server

type RepoFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

type FileData struct {
	Content string `json:"content"`
}

type Vulnerability struct {
	ID             string      `json:"id"`
	Severity       string      `json:"severity"`
	CVSS           float64     `json:"cvss"`
	Status         string      `json:"status"`
	PackageName    string      `json:"package_name"`
	CurrentVersion string      `json:"current_version"`
	FixedVersion   string      `json:"fixed_version"`
	Description    string      `json:"description"`
	PublishedDate  string      `json:"published_date"`
	Link           string      `json:"link"`
	RiskFactors    interface{} `json:"risk_factors"`
}

type ScanRequest struct {
	RepoURL string   `json:"repo_url"`
	Files   []string `json:"files"`
}

type QueryRequest struct {
	Filters map[string]string `json:"filters"`
}

type ScanResult struct {
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

type ScanFile struct {
	ScanResults ScanResult `json:"scanResults"`
}

type Result struct {
	FilePath string
	Error    error
}
