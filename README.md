# go
# Instructions for Compilation, Execution, and Deployment

## 4.1 Prerequisites

- Go installed (version 1.18 or higher).
- SQLite installed (optional, for database inspection).
- Tienes que teneir pre instaldo "gcc" de 64

## 4.2 Compilation Instructions

1. Navigate to the project directory:
   ```bash
   cd /path/to/project
   ```

2. Initialize the Go module (if not already created):
   ```bash
   go mod init proyecto_final
   ```

3. Download the dependencies:
   ```bash
   go mod tidy
   ```

4. Compile the project:
   ```bash
   go build -o proyecto_final
   ```

## 4.3 Execution Instructions

1. Run the compiled file:
   ```bash
   ./proyecto_final
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8081
   ```

## 4.4 Deployment Instructions

To deploy the project on a server:

1. Upload the project files to the server.
2. Configure a systemd service or equivalent to keep the process running.
3. Ensure ports 8080 (API) and 8081 (client) are open.

