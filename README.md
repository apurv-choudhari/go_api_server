# Go API Server

## Description 
This project presents a Go API server with two endpoints 
1. **/scan** 
   - It scans GitHub repository ( https://github.com/velancio/vulnerability_scans ) for JSON files and store their contents in MySQL database.
2. **/query**
   - It restrives JSON payloads stored in MySQL database using key-value filters.

This project is containerized using Docker, ensuring a consistent and reproducible environment. The container runs two interdependent services—a database and an API server. While running multiple services within a single container violates the Single Responsibility Principle (SRP), this architecture is a project requirement. To handle the lifecycle of both services efficiently, the project leverages Supervisor, a process control system that ensures:

- Both services start automatically when the container is launched.
- Services are restarted if they crash or exit unexpectedly.
- Logs and process states are managed centrally.

By using Supervisor, the project avoids the complexities of managing multiple containers while ensuring both services remain available and functional within a single containerized environment.

## Database
The project uses MariaDB database. By default, when a MariaDB container is initialized, it assigns a well-known default root password, which poses a security risk. To mitigate this, the root password is automatically changed during container startup via an SQL initialization script. The pre-configured initializations script init.sql is placed in /docker-entrypoint-initdb.d/. The script creates updates root password and creates database and its scema.

Why /docker-entrypoint-initdb.d/ is Special?
MariaDB (and MySQL) containers have a special behavior: any .sql, .sh, or .sql.gz files placed inside /docker-entrypoint-initdb.d/ are automatically executed when the container starts for the first time. This directory is part of the official MariaDB Docker entrypoint script, which ensures:

- Scripts run only if the database is empty (i.e., on first container startup).
- The database is initialized before the main service starts accepting queries.
- Custom scripts (like init.sql) can be used for automated configuration.

## API Endpoints and Concurrency Management
The /scan endpoint is responsible for processing multiple JSON files simultaneously. To achieve efficient parallel execution, it uses a thread pool with 5 worker threads:

- Each incoming request can include multiple JSON files.
- The API server dispatches each file to one of the 5 worker threads.
- These threads process the files concurrently, reducing request processing time.
- Thread safety is ensured to prevent race conditions while inserting data into the database.

Unlike /scan, the /query endpoint is a read-only operation. Since database queries are already optimized for fast retrieval, concurrency is not needed. This keeps the implementation simple and avoids unnecessary overhead.


## Execution Steps 

```
docker pull apurv98/go_api_server
docker run -p 8000:8000 apurv98/go_api_server
```