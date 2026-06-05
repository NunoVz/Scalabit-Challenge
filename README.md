# Scalabit - DevSecOps Challenge API


## Requirements
- [x] Create a REST API to create, close (destroy), and list issues for a certain GitHub repository.
- [x] Create a REST API to check if a specific PR's pipeline succeeded.
- [x] CI/CD Pipeline (GitHub Actions) for running tests, linting, security checks, and deployment to Minikube.
- [x] Written in Go.
- [x] High test coverage.

---

##  Getting Started (Local Development)

### Prerequisites
- Go 1.26+
- Docker
- Minikube & kubectl (for deployment)
- A GitHub Personal Access Token (PAT)

### 1. Environment Setup
Create a `.env` file in the root of the project and add your GitHub Token. This token needs `repo` permissions to read/write issues and PRs.

```env
GITHUB_TOKEN=ghp_your_personal_access_token_here
```

### 2. Run the Application locally
```bash
go mod tidy
go run ./api/main.go
```
The server will start on `http://localhost:8080`.
A frontend UI is served at `/` to easily interact with the API endpoints.

---

##  API Documentation

### 1. Health Check
Checks if the API is running and responding.
* **URL:** `GET /health`
* **Success Response:** `200 OK` (Body: `OK`)

### 2. List Issues
Retrieves all issues for a given repository.
* **URL:** `GET /repos/{owner}/{repo}/issues`
* **Success Response:** `200 OK`
* **Response Body:** Array of GitHub Issue objects.

### 3. Create Issue
Creates a new issue in the specified repository.
* **URL:** `POST /repos/{owner}/{repo}/issues`
* **Request Body (JSON):**
  ```json
  {
    "title": "My Bug Report",
    "body": "Description of the issue goes here."
  }
  ```
* **Success Response:** `201 Created`
* **Response Body:** The created GitHub Issue object.

### 4. Close (Destroy) Issue
Closes an existing issue in the specified repository (simulating deletion, as GitHub API restricts true deletion to specific scopes).
* **URL:** `DELETE /repos/{owner}/{repo}/issues/{id}`
* **URL Params:** `id` (The issue number, e.g., `1`)
* **Success Response:** `200 OK`
* **Response Body:** 
  ```json
  {
    "message": "Issue 1 successfully closed",
    "title": "My Bug Report"
  }
  ```

### 5. Check PR Pipeline Status
Checks the CI/CD pipeline status of the latest commit in a Pull Request.
* **URL:** `GET /repos/{owner}/{repo}/prs/{id}/status`
* **URL Params:** `id` (The PR number, e.g., `5`)
* **Success Response:** `200 OK`
* **Response Body:**
  ```json
  {
    "pr_number": 5,
    "status": "success"  // Can be: "success", "pending", "failure", or "no_checks_found"
  }
  ```

---

## DevSecOps & CI/CD Architecture

This project focuses on the **Shift-Left Security** approach, catching vulnerabilities and bugs before they  reach the deployment phase.

### 1. Code Quality & Unit Testing
- **GolangCI-Lint:** Enforces Go best practices and code formatting.
- **Unit Testing:** Handlers and business logic are tested natively.
- **Coverage Quality Gate:** The pipeline calculates code coverage natively via Bash script and **fails automatically if coverage drops below 80%**. Entrypoints and third-party wrappers are gracefully excluded from the metric.

### 2. Comprehensive Security Scanning
- **SAST (Static Application Security Testing):** Uses `gosec` to scan the Go source code for insecure coding patterns (e.g., hardcoded credentials, SQL injection).
- **SCA (Software Composition Analysis):** Uses `Trivy` (`fs` mode) to scan third-party Go modules and dependencies for known CVEs. Fails the pipeline on `CRITICAL` or `HIGH` vulnerabilities.
- **Secret Scanning:** Uses `Gitleaks` to perform a full-depth Git history scan, ensuring no API keys or tokens were accidentally committed.
- **IaC Scanning:** Uses `Trivy` (`config` mode) to scan the `Dockerfile` and Kubernetes YAMLs (`k8s/`) for misconfigurations (e.g., missing resource limits, root user execution).
- **Container Image Scanning:** After the Docker image is built locally, `Trivy` scans the final image for OS-level vulnerabilities (Alpine base image) before allowing it into the cluster.

### 3. Kubernetes Hardening & Security Context
The deployment to Minikube adheres to the following security contexts:
- **Non-Root Execution:** The Docker image creates a specific user (UID `10001`). The `deployment.yml` mathematically enforces `runAsUser: 10001` and `runAsNonRoot: true`.
- **Immutable Filesystem:** `readOnlyRootFilesystem: true` prevents attackers from writing malicious scripts to the container's disk.
- **Privilege Escalation:** `allowPrivilegeEscalation: false` and dropping `ALL` capabilities ensures the container is completely isolated from the host node.

### 4. Continuous Deployment (CD) & Smoke Testing
- **Minikube Deployment:** The pipeline boots a Minikube cluster, injects the locally scanned image (`Never` pull policy), and securely mounts the `GITHUB_TOKEN` via Kubernetes Secrets.
- **Rollout Verification:** Kubernetes waits for the Pods to become healthy using Readiness/Liveness probes.
- **External Smoke Test:** Finally, the pipeline acts as an external client, dynamically retrieving the allocated NodePort and issuing a `curl` request. If it does not receive an HTTP 200, the deployment is marked as a failure.

---

## How to Deploy (Minikube)

To deploy the API locally using Minikube, ensure your cluster is running and execute:

```bash
# 0. Start Minikube (if not already running)
minikube start

# 1. Build the Docker Image
docker build -t scalabit-challenge-api:latest .
minikube image load scalabit-challenge-api:latest

# 2. Create the Secret for GitHub Token
kubectl create secret generic github-creds --from-literal=GITHUB_TOKEN=your_token_here

# 3. Apply Kubernetes Manifests
kubectl apply -f k8s/

# 4. Access the API
minikube service scalabit-challenge-api-svc --url
```
