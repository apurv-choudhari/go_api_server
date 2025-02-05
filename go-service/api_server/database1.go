package api_server

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type SqlOpener func(driverName, dataSourceName string) (*sql.DB, error)

func ConnectDB(sqlOpen SqlOpener) (*sql.DB, error) {
	dsn := "root:rootpass@tcp(localhost:3306)/mysql"
	db, err := sqlOpen("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Database connected successfully.")
	return db, nil
}

func ProcessFile(dbConn *sql.DB, file RepoFile, wg *sync.WaitGroup, results chan<- Result) {
	defer wg.Done()

	log.Printf("Processing file: %s", file.Path)
	startTime := time.Now()

	resp, err := http.Get(file.URL)
	if err != nil {
		log.Printf("Failed to get file conetents %s: %v", file.URL, err)
		results <- Result{file.Path, err}
		return
	}
	defer resp.Body.Close()

	var fileData FileData
	if err := json.NewDecoder(resp.Body).Decode(&fileData); err != nil {
		log.Printf("Failed to decode json response %s: %v", file.Path, err)
		results <- Result{file.Path, err}
		return
	}

	decodedContent, err := base64.StdEncoding.DecodeString(fileData.Content)
	if err != nil {
		log.Printf("Failed to decode Base64 content %s: %v", file.Path, err)
		results <- Result{file.Path, err}
		return
	}

	var scanFiles []ScanFile
	if err := json.Unmarshal(decodedContent, &scanFiles); err != nil {
		log.Printf("Failed to unmarshal json scan results %s: %v", file.Path, err)
		results <- Result{file.Path, err}
		return
	}

	if len(scanFiles) == 0 {
		results <- Result{file.Path, nil}
		return
	}

	usedb := "USE kai_security;"
	dbConn.Exec(usedb)
	insertQuery := `
		INSERT INTO vulnerabilities
		(id, source_file, scan_time, severity, cvss, status, package_name,
		current_version, fixed_version, description, published_date, link, risk_factors)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	tx, err := dbConn.Begin()
	if err != nil {
		results <- Result{file.Path, err}
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		results <- Result{file.Path, err}
		return
	}
	defer stmt.Close()

	for _, vuln := range scanFiles[0].ScanResults.Vulnerabilities {
		publishedTime, err := time.Parse(time.RFC3339, vuln.PublishedDate)
		if err != nil {
			log.Printf("Invalid published date for %s in file %s, using current time: %v", vuln.ID, file.Path, err)
			publishedTime = time.Now()
		}

		riskFactorsJSON, err := json.Marshal(vuln.RiskFactors)
		if err != nil {
			log.Printf("Failed to marshal risk factors for %s in file %s: %v", vuln.ID, file.Path, err)
			riskFactorsJSON = []byte("[]")
		}

		_, err = stmt.Exec(
			vuln.ID,
			file.Path,
			time.Now(),
			vuln.Severity,
			vuln.CVSS,
			vuln.Status,
			vuln.PackageName,
			vuln.CurrentVersion,
			vuln.FixedVersion,
			vuln.Description,
			publishedTime,
			vuln.Link,
			string(riskFactorsJSON),
		)
		if err != nil {
			log.Printf("Error inserting vulnerability %s: %v", vuln.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction %s: %v", file.Path, err)
		results <- Result{file.Path, err}
		return
	}

	log.Printf("Successfully processed file: %s in %v", file.Path, time.Since(startTime))
	results <- Result{file.Path, nil}
}

func PopulateDB(dbConn *sql.DB, files []RepoFile) error {
	if len(files) == 0 {
		log.Println("No files to process. Skipping populateDB.")
		return nil
	}
	log.Printf("Populating DB with %d files", len(files))
	results := make(chan Result, len(files))
	var wg sync.WaitGroup

	numWorkers := 5

	if len(files) < 5 {
		numWorkers = len(files)
	}
	log.Printf("Using %d workers", numWorkers)

	filesChan := make(chan RepoFile, len(files))

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			log.Printf("Worker %d started", workerID)
			for file := range filesChan {
				log.Printf("Worker %d is processing file: %s", workerID, file.Path)
				ProcessFile(dbConn, file, &wg, results)
			}
			log.Printf("Worker %d finished", workerID)
		}(i)
	}

	wg.Add(len(files))
	for _, file := range files {
		filesChan <- file
	}
	close(filesChan)

	go func() {
		wg.Wait()
		close(results)
	}()

	var errors []error
	for result := range results {
		if result.Error != nil {
			log.Printf("Error processing file %s: %v", result.FilePath, result.Error)
			errors = append(errors, result.Error)
		} else {
			log.Printf("Successfully processed file: %s", result.FilePath)
		}
	}

	if len(errors) > 0 {
		return errors[0]
	}

	log.Println("populateDB completed successfully")
	return nil
}

func SevereVuln(dbConn *sql.DB, severity string) ([]map[string]interface{}, error) {
	query := `SELECT * FROM vulnerabilities WHERE severity = ?`
	log.Printf("Executing query: %s with severity: %s", query, severity)

	rows, err := dbConn.Query(query, severity)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	log.Printf("Columns retrieved: %v", columns)

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			if b, ok := values[i].([]byte); ok {
				v = string(b)
			} else {
				v = values[i]
			}
			rowMap[col] = v
		}
		results = append(results, rowMap)
	}
	return results, nil
}
