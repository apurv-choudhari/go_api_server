ALTER USER 'root'@'localhost' IDENTIFIED BY "rootpass";
FLUSH PRIVILEGES;

CREATE DATABASE IF NOT EXISTS kai_security;

USE kai_security;

CREATE TABLE IF NOT EXISTS vulnerabilities (
    id VARCHAR(20) PRIMARY KEY,
    source_file VARCHAR(100),
    scan_time DATETIME,
    current_version VARCHAR(20),
    cvss VARCHAR(20),
    description TEXT,
    fixed_version VARCHAR(20),
    link VARCHAR(255),
    package_name VARCHAR(100),
    published_date DATETIME,
    severity VARCHAR(20),
    status VARCHAR(20),
    risk_factors VARCHAR(255)
);