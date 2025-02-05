package api_server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func parseGitHubURL(repoURL string) (string, string, error) {
	parsed, err := url.Parse(repoURL)
	if err != nil {
		log.Printf("Error: Failed to parse repo URL %s: %v", repoURL, err)
		return "", "", fmt.Errorf("failed to parse repo: %s, err: %w", repoURL, err)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		log.Printf("Error: Invalid GitHub URL structure: %s", repoURL)
		return "", "", fmt.Errorf("invalid URL: failed to parse GitHub repository")
	}

	owner, repo := parts[0], parts[1]
	log.Printf("Parsed GitHub Owner: %s, Repo: %s", owner, repo)

	return owner, repo, nil
}

func getRepoFiles(owner, repo string, scanFiles []string) ([]RepoFile, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/", owner, repo)
	log.Printf("Getting repository files from GitHub: %s", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Error: Failed to fetch repository files from %s: %v", apiURL, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("github api failure: failed to fetch contents of file using github api")
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var contents []RepoFile
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		log.Printf("Failed to decode json response from github: %v", err)
		return nil, err
	}

	//Can be optimised
	var files []RepoFile
	for _, file := range contents {
		for _, f := range scanFiles {
			if file.Path == f {
				files = append(files, file)
			}
		}
	}

	log.Printf("Total matching files found: %d", len(files))
	return files, nil
}
