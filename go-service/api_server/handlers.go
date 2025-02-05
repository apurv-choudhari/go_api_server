package api_server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func ScanRepoHandler(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.RepoURL == "" || len(req.Files) == 0 {
		http.Error(w, "Missing repo_url or files", http.StatusBadRequest)
		return
	}

	owner, repo, err := parseGitHubURL(req.RepoURL)
	if err != nil {
		http.Error(w, "Invalid GitHub repository URL", http.StatusBadRequest)
		return
	}
	log.Println("Repo:", req.RepoURL, "Files:", req.Files)
	repoFiles, err := getRepoFiles(owner, repo, req.Files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := PopulateDB(dbConn, repoFiles); err != nil {
		http.Error(w, "Failed to populate database", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"repository": owner + "/" + repo,
		"file_count": len(repoFiles),
		"files":      repoFiles,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func QueryVulnHandler(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	severity, ok := req.Filters["severity"]
	if !ok || severity == "" {
		http.Error(w, "Missing severity filter", http.StatusBadRequest)
		return
	}

	results, err := SevereVuln(dbConn, severity)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
