# SyntaxDevopsSquad - WhoKnows Migration Project

Welcome to the **SyntaxDevopsSquad** main repository. This project is part of our 2026 DevOps module at KEA, focusing on migrating a legacy Python Flask application to Go while learning DevOps practices including automation, CI/CD, and infrastructure as code.

## Live Application

| | Current (Terraform + Azure + DigitalOcean) | Original (Azure only, pre-Terraform) |
|---|---|---|
| **App** | [https://syntax-reborndev.com/](https://syntax-reborndev.com/) | [https://original.syntax-reborndev.com/](https://original.syntax-reborndev.com/) |
| **Monitoring** | [https://monitor.syntax-reborndev.com/](https://monitor.syntax-reborndev.com/) | [https://original-monitor.syntax-reborndev.com/](https://original-monitor.syntax-reborndev.com/) |

---

## Project Overview

**WhoKnows** is a web application for searching and managing wiki-style pages with user authentication. We have successfully migrated the application from Python/Flask to Go as part of our DevOps learning journey. Our team of 4 developers has implemented modern DevOps practices including containerization, automated CI/CD pipelines, and cloud deployment.

### Core Functionality
- **User Authentication:** Registration, login, session management, and password reset
- **Page Management:** Create, read, and search wiki-style pages (PostgreSQL full-text search with `tsvector`)
- **Security:** CSRF protection, middleware, and breach response tooling
- **Database:** PostgreSQL with native full-text search
- **Health Check:** `GET /health` endpoint for uptime monitoring and watchdog integration

### Team Members
- **CodeByNajib** (NajibGPT)
- **AceS0**
- **MarcusLieberH**
- **Daniel23894** (Daniel Søgaard)

---

## Tech Stack

### Backend
- **Language:** Go 1.25.0
- **Database:** PostgreSQL 16 with `github.com/lib/pq`
- **Session Management:** Gorilla Sessions
- **Legacy:** Python Flask (original implementation, kept for reference)

### Infrastructure & DevOps
- **Cloud Platforms:** Azure (app VM) + DigitalOcean (monitoring VM)
- **Containerization:** Docker + Docker Compose (dev & prod)
- **CI/CD:** GitHub Actions (`ci.yml`, `cd.yml`, `dependabot-auto-merge.yml`)
- **Infrastructure as Code:** Terraform (Azure VM, network, firewall, DigitalOcean droplet, Cloudflare DNS)
- **Configuration Management:** Ansible (Docker, Nginx, fail2ban, UFW, swap, disk mount)
- **DNS:** Cloudflare (automatic A-record update on deploy)
- **Persistent Storage:** Azure Managed Disk (Postgres data), DigitalOcean Volume (Prometheus data) — both managed outside Terraform lifecycle
- **Remote State:** Terraform state stored in Azure Blob Storage
- **Server Security:** fail2ban, UFW
- **Linting:** `golangci-lint`
- **Code Quality:** SonarCloud (Automatic Analysis)
- **Version Control:** Git with Conventional Commits
- **Development Environment:** WSL (Ubuntu 22.04)

### Monitoring Stack
- **Metrics:** Prometheus (scrapes `/metrics` on app VM port 8080)
- **Dashboards:** Grafana (auto-provisioned with datasource + 3 dashboards via Ansible)
- **Watchdog:** Cron job on monitoring VM — checks `/health` every 5 minutes, auto-restarts app via SSH after 3 consecutive failures
- **Deployment:** Separate DigitalOcean VM for resilience (survives Azure app VM destroy)

### Database Schema
- **users table:** User authentication and profiles
- **pages table:** Wiki-style content storage with `tsvector` full-text search and GIN index

---

## Project Structure

```
devops-syntaxsquad/
├── README.md
├── docker-compose.yml                   # Development environment
├── docker-compose.prod.yml              # Production environment
├── docs/
│   ├── openapi.yaml                     # API specification
│   └── mandatory/
│       ├── BRANCHING_STRATEGY.md        # Git branching documentation
│       ├── dependency_graph.dot         # System architecture (source)
│       ├── mandatory_ii.md              # DevOps reflection task II
│       ├── monitoring_repo_prompt.md
│       └── technical_audit.md           # Technical audit report
├── implementations/
│   ├── go/                              # Active Go implementation
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── schema.sql
│   │   ├── backend/
│   │   │   ├── main.go
│   │   │   ├── routes.go
│   │   │   ├── routes_test.go
│   │   │   ├── database.go
│   │   │   ├── database_test.go
│   │   │   ├── integration_test.go
│   │   │   ├── middleware.go
│   │   │   ├── migrations.go
│   │   │   ├── metrics.go
│   │   │   ├── metrics_test.go
│   │   │   ├── security.go
│   │   │   ├── security_test.go
│   │   │   └── entrypoint.sh
│   │   ├── scripts/
│   │   │   ├── deploy.sh
│   │   │   ├── deploy_compose.sh
│   │   │   ├── migration.sh
│   │   │   ├── setup.sh
│   │   │   └── breach_response.sh
│   │   ├── static/
│   │   │   └── style.css
│   │   └── templates/
│   │       ├── layout.html
│   │       ├── search.html
│   │       ├── login.html
│   │       ├── register.html
│   │       ├── reset-password.html
│   │       └── about.html
│   └── python/                          # Legacy Flask implementation (reference only)
│       ├── Makefile
│       ├── schema.sql
│       ├── run_forever.sh
│       └── backend/
│           ├── app.py
│           ├── app_tests.py
│           └── requirements.txt
├── terraform/
│   ├── main.tf                          # Azure app VM + Cloudflare DNS
│   ├── monitoring.tf                    # DigitalOcean monitoring VM + DO Volume attachment
│   ├── outputs.tf
│   ├── variables.tf
│   ├── terraform.tfvars.example
│   ├── scripts/
│   │   ├── bootstrap-disk.sh            # One-time: create Azure Managed Disk (Postgres)
│   │   └── bootstrap-do-volume.sh       # One-time: create DO Volume (Prometheus)
│   └── ansible/
│       ├── playbook.yml                 # App VM setup
│       ├── monitoring-playbook.yml      # Monitoring VM setup
│       ├── deploy.sh                    # Full deploy (Terraform + both Ansible playbooks)
│       └── grafana-provisioning/        # Auto-provisioned datasource + dashboards
│           ├── datasources/
│           │   └── prometheus.yml
│           └── dashboards/
│               ├── dashboard.yml
│               ├── whoknows-auth.json
│               ├── whoknows-requests.json
│               └── whoknows-overview.json
├── server-config/
│   └── fail2ban-jail.local              # Server security configuration
└── .github/
    └── workflows/
        ├── ci.yml                       # Continuous Integration
        ├── cd.yml                       # Continuous Deployment
        └── dependabot-auto-merge.yml
```

---

## Getting Started

### Prerequisites

**Required:**
- Go 1.25.0 or higher
- Docker & Docker Compose
- Git
- WSL/Linux environment (for Windows users)

**For infrastructure deployment:**
- Azure CLI (`az`) — authenticated
- Terraform
- Ansible (WSL only)
- `doctl` (DigitalOcean CLI)

### Environment Variables

Copy `.env.example` to `.env` in the project root and fill in real values:

```env
SECRET_KEY=replace-with-a-long-random-secret
CSRF_RELAXED=false
POSTGRES_DB=whoknows
POSTGRES_USER=whoknows
POSTGRES_PASSWORD=replace-with-a-strong-password
IMAGE_NAME=ghcr.io/syntaxdevopssquad-sds/whoknows-go
IMAGE_TAG=latest
GHCR_USER=your-github-username
GHCR_PAT=your-github-pat-with-read:packages-scope
GRAFANA_PASSWORD=replace-with-a-strong-password
```

Copy `terraform/terraform.tfvars.example` to `terraform/terraform.tfvars` and fill in:

```hcl
subscription_id        = "your-azure-subscription-id"
vm_name                = "whoknows-vm"
cloudflare_api_token   = "your-cloudflare-api-token"
cloudflare_zone_id     = "your-cloudflare-zone-id"
do_token               = "your-digitalocean-api-token"
do_ssh_key_fingerprint = "your-ssh-key-fingerprint-from-do-dashboard"
```

### Running with Docker (recommended)

```bash
# Development
docker compose up

# Production
docker compose -f docker-compose.prod.yml up
```

### Running locally

1. **Clone the repository:**
```bash
git clone https://github.com/SyntaxDevopsSquad-SDS/devops-syntaxsquad.git
cd devops-syntaxsquad/implementations/go
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Start PostgreSQL:**
```bash
docker compose up postgres -d
```

4. **Set environment variables:**
```bash
export DATABASE_URL=postgres://whoknows:whoknows@localhost:5432/whoknows?sslmode=disable
export SECRET_KEY=your-secret-key-here
```

5. **Run the application:**
```bash
go run ./backend/...
```

6. **Access the application:**
Open your browser and navigate to `http://localhost:8080`

### CSRF Simulation Mode

The Go backend protects login/register with CSRF tokens by default. For controlled black-box simulations that call API endpoints directly, you can relax this check:

```env
CSRF_RELAXED=true
```

Use `CSRF_RELAXED=false` in normal/prod operation. Enable `CSRF_RELAXED=true` only when simulation tooling cannot handle form CSRF flow.

---

## Infrastructure Deployment

### First-time setup (run once)

```bash
# 1. Create persistent Azure Managed Disk (Postgres data)
bash terraform/scripts/bootstrap-disk.sh

# 2. Create persistent DigitalOcean Volume (Prometheus data)
doctl auth init
bash terraform/scripts/bootstrap-do-volume.sh
```

### Deploy everything

```bash
# From WSL
cd terraform/ansible
sed -i 's/\r//' deploy.sh
bash deploy.sh
```

This runs `terraform apply`, then Ansible on both VMs. Cloudflare DNS is updated automatically.

### Destroy infrastructure

```bash
cd terraform
terraform destroy -auto-approve
```

Postgres and Prometheus data are **not deleted** — they live on persistent disks outside Terraform.

---

## Monitoring (Prometheus + Grafana)

The Go backend exposes metrics at `GET /metrics` (port 8080) and a health check at `GET /health`.

Prometheus and Grafana run on a separate DigitalOcean VM. Grafana dashboards and the Prometheus datasource are auto-provisioned via Ansible on every deploy — no manual setup required.

Live Grafana dashboard: [https://monitor.syntax-reborndev.com/](https://monitor.syntax-reborndev.com/)

### Available Metrics

- `whoknows_http_requests_total{method,path,status}`
- `whoknows_http_request_duration_seconds{method,path}`
- `whoknows_login_attempts_total{outcome}` — `success|failure`
- `whoknows_registrations_total{outcome}` — `success|validation_error|failure`
- `whoknows_searches_total{source,language,query,outcome}` — `source: web|api`, `outcome: success|failure`

### Prometheus Query Examples

```promql
# Total HTTP requests in the last 5 minutes
sum(increase(whoknows_http_requests_total[5m]))

# Successful logins in the last 1 hour
increase(whoknows_login_attempts_total{outcome="success"}[1h])

# Successful registrations in the last 1 hour
increase(whoknows_registrations_total{outcome="success"}[1h])

# Searches for a specific term in the last 1 hour
increase(whoknows_searches_total{query="fortran"}[1h])
```

### Watchdog

The monitoring VM runs a cron job every 5 minutes that checks `GET /health` on the app VM. After 3 consecutive failures it SSH's into the app VM and runs `docker compose restart` automatically.

---

## Development Workflow

### Git Commit Conventions

We follow **Conventional Commits** for clean and readable history:

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New functionality | `feat: add user authentication` |
| `fix` | Bug fixes | `fix: resolve database connection issue` |
| `refactor` | Code optimization | `refactor: improve error handling` |
| `docs` | Documentation | `docs: update README with setup steps` |
| `ci` | CI/CD changes | `ci: add Docker build workflow` |
| `test` | Tests | `test: add integration tests` |
| `style` | Code formatting | `style: format Go code with gofmt` |
| `chore` | Maintenance | `chore: reorganize docs folder` |

**Format:** `<type>: <description>`

### Branch Strategy

See [`docs/mandatory/BRANCHING_STRATEGY.md`](docs/mandatory/BRANCHING_STRATEGY.md) for the full strategy.

We follow **GitHub Flow**:

- `main` - Production-ready code, always deployable
- `feat/*` - New features (branch from main, PR back to main)
- `fix/*` - Bug fixes (branch from main, PR back to main)
- `ci/*` - CI/CD changes (branch from main, PR back to main)
- `chore/*` - Maintenance and housekeeping

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run unit tests
go test ./...
```

---

## Project Milestones

### Week 1-2: Foundation - Completed
- [x] Legacy Python codebase analysis
- [x] Dependency graph creation
- [x] Framework selection (Go)
- [x] OpenAPI specification
- [x] Kanban board setup (GitHub Projects)
- [x] Initial Go project structure

### Week 3: Deployment & Cloud - Completed
- [x] GitHub Actions CI/CD setup
- [x] Azure VM deployment
- [x] SSH configuration
- [x] Production deployment
- [x] Custom domain

### Week 4-5: Quality & Containerization - Completed
- [x] Linting setup (`golangci-lint`)
- [x] Branch protection rules
- [x] Docker containerization
- [x] Docker Compose (dev + prod)
- [x] Integration tests
- [x] Dependabot with auto-merge

### Week 6-7: Continuous Delivery - Completed
- [x] Continuous Delivery pipeline (`cd.yml`)
- [x] Docker Compose production deployment
- [x] Security hardening (fail2ban, CSRF, middleware)
- [x] Password reset flow

### Week 8+: Advanced Topics - Completed
- [x] PostgreSQL migration (from SQLite)
- [x] Monitoring and observability (Prometheus + Grafana)
- [x] SonarCloud code quality analysis
- [x] Terraform infrastructure as code (Azure + DigitalOcean + Cloudflare)
- [x] Ansible configuration management (app VM + monitoring VM)
- [x] Persistent cloud storage (Azure Managed Disk + DigitalOcean Volume)
- [x] Grafana dashboard provisioning via Ansible
- [x] Watchdog auto-recovery via cron + SSH
- [x] Health endpoint (`/health`)

---

## Contributing

1. Create a new branch from `main` following the branch naming convention
2. Make your changes
3. Write/update tests
4. Follow commit conventions
5. Create a Pull Request targeting `main`

---

## License

This project is part of KEA's DevOps module 2026.

---

**Course:** DevOps 2026
**Institution:** KEA Kobenhavn
**Instructor:** Anders Latif
