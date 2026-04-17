# SyntaxDevopsSquad - WhoKnows Migration Project

Welcome to the **SyntaxDevopsSquad** main repository. This project is part of our 2026 DevOps module at EK, focusing on migrating a legacy Python Flask application to Go while learning DevOps practices including automation, CI/CD, and infrastructure as code.

## 🌐 Live Application

> **URL:** `<!-- TODO: indsæt domæne her -->`

---

## 📋 Project Overview

**WhoKnows** is a web application for searching and managing wiki-style pages with user authentication. We have successfully migrated the application from Python/Flask to Go as part of our DevOps learning journey. Our team of 4 developers has implemented modern DevOps practices including containerization, automated CI/CD pipelines, and cloud deployment.

### Core Functionality
- **User Authentication:** Registration, login, session management, and password reset
- **Page Management:** Create, read, and search wiki-style pages (FTS5 full-text search)
- **Security:** CSRF protection, middleware, and breach response tooling
- **Database:** SQLite with migration support

### Team Members
- **CodeByNajib** (NajibGPT)
- **AceS0**
- **MarcusLieberH**
- **Daniel23894** (Daniel Søgaard)

---

## 🛠 Tech Stack

### Backend
- **Language:** Go 1.25.0
- **Database:** SQLite with `modernc.org/sqlite`
- **Session Management:** Gorilla Sessions
- **Legacy:** Python Flask (original implementation, kept for reference)

### Infrastructure & DevOps
- **Cloud Platform:** Azure (Azure for Students)
- **Containerization:** Docker + Docker Compose (dev & prod)
- **CI/CD:** GitHub Actions (`ci.yml`, `cd.yml`, `dependabot-auto-merge.yml`)
- **Linting:** `golangci-lint`
- **Server Security:** fail2ban
- **Version Control:** Git with Conventional Commits
- **Development Environment:** WSL (Ubuntu 24.04)

### Database Schema
- **users table:** User authentication and profiles
- **pages table:** Wiki-style content storage with FTS5 full-text search
- **Migrations:** SQL migration system (`migrations/`)

---

## 📁 Project Structure

```
devops-syntaxsquad/
├── docker-compose.yml           # Development environment
├── docker-compose.prod.yml      # Production environment
├── init_db.go                   # Database initialization
├── queries.go                   # SQL queries
├── docs/
│   ├── BRANCHING_STRATEGY.md    # Git branching documentation
│   ├── dependency_graph.dot     # System architecture (source)
│   ├── dependency_graph_picture.svg  # System architecture (visual)
│   ├── openapi.yaml             # API specification
│   └── technical_audit.md      # Technical audit report
├── implementations/
│   ├── go/                      # Active Go implementation
│   │   ├── Dockerfile
│   │   ├── backend/
│   │   │   ├── main.go
│   │   │   ├── routes.go
│   │   │   ├── database.go
│   │   │   ├── database_test.go
│   │   │   ├── integration_test.go
│   │   │   ├── middleware.go
│   │   │   ├── migrations.go
│   │   │   ├── security.go
│   │   │   └── entrypoint.sh
│   │   ├── migrations/
│   │   │   ├── 001_add_fts5.sql
│   │   │   └── 002_add_password_reset.sql
│   │   ├── scripts/
│   │   │   ├── deploy.sh
│   │   │   ├── deploy_compose.sh
│   │   │   ├── migration.sh
│   │   │   ├── setup.sh
│   │   │   └── breach_response.sh
│   │   ├── static/
│   │   │   └── style.css
│   │   ├── templates/
│   │   │   ├── layout.html
│   │   │   ├── search.html
│   │   │   ├── login.html
│   │   │   ├── register.html
│   │   │   ├── reset-password.html
│   │   │   └── about.html
│   │   ├── schema.sql
│   │   ├── go.mod
│   │   └── go.sum
│   └── python/                  # Legacy Flask implementation (reference only)
│       ├── backend/
│       │   ├── app.py
│       │   ├── app_tests.py
│       │   └── requirements.txt
│       ├── schema.sql
│       └── run_forever.sh
├── server-config/
│   └── fail2ban-jail.local      # Server security configuration
└── .github/
    └── workflows/
        ├── ci.yml               # Continuous Integration
        ├── cd.yml               # Continuous Deployment
        └── dependabot-auto-merge.yml
```

---

## 🚀 Getting Started

### Prerequisites

**Required:**
- Go 1.25.0 or higher
- SQLite3
- Git
- Docker & Docker Compose
- WSL/Linux environment (for Windows users)

**Optional:**
- Azure CLI (`az`)
- Terraform

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

3. **Initialize the database:**
```bash
sqlite3 whoknows.db < schema.sql
```

4. **Run the application:**
```bash
go run main.go
```

5. **Access the application:**
Open your browser and navigate to `http://localhost:8080`

### CSRF Simulation Mode

The Go backend protects login/register with CSRF tokens by default. For controlled black-box simulations that call API endpoints directly, you can relax this check:

```env
CSRF_RELAXED=true
```

- Use `CSRF_RELAXED=false` in normal/prod operation.
- Enable `CSRF_RELAXED=true` only when simulation tooling cannot handle form CSRF flow.

---

## Monitoring (Prometheus + Grafana)

The Go backend now exposes a Prometheus endpoint at:

- `GET /metrics` (same host/port as app, default `:8080`)

### Available Metrics

- `whoknows_http_requests_total{method,path,status}`
- `whoknows_http_request_duration_seconds{method,path}`
- `whoknows_login_attempts_total{outcome}`
- `whoknows_registrations_total{outcome}`
- `whoknows_searches_total{source,language,query,outcome}`

`whoknows_searches_total` lets you chart searches for specific terms via the `query` label.

### Prometheus Query Examples

```promql
# Total HTTP requests in the last 5 minutes
sum(increase(whoknows_http_requests_total[5m]))

# Successful logins in the last 1 hour
increase(whoknows_login_attempts_total{outcome="success"}[1h])

# Successful registrations in the last 1 hour
increase(whoknows_registrations_total{outcome="success"}[1h])

# Searches for a specific term (example: "fortran") in the last 1 hour
increase(whoknows_searches_total{query="fortran"}[1h])
```

### Separate Monitoring VM

Prometheus + Grafana can run in a separate repository and on a separate VM.

On that monitoring VM, configure Prometheus to scrape this app endpoint:

```yaml
scrape_configs:
    - job_name: "whoknows-go-backend"
        metrics_path: /metrics
        static_configs:
            - targets: ["<APP_VM_PUBLIC_OR_PRIVATE_IP>:8080"]
```

## 🔄 Development Workflow

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

**Format:** `<type>: <description>`

### Branch Strategy

Se [`docs/BRANCHING_STRATEGY.md`](docs/BRANCHING_STRATEGY.md) for the full strategy.

We follow **GitHub Flow** — simple and effective for our team size:

- `main` — Production-ready code, always deployable
- `feat/*` — New features (branch from main, PR back to main)
- `fix/*` — Bug fixes (branch from main, PR back to main)
- `ci/*` — CI/CD changes (branch from main, PR back to main)

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run unit tests
go test ./...

# Run with verbose output
go test -v ./...
```

---

## ✅ Project Milestones

### Week 1-2: Foundation ✅ Completed
- [x] Legacy Python codebase analysis
- [x] Dependency graph creation
- [x] Framework selection (Go)
- [x] OpenAPI specification
- [x] Kanban board setup (GitHub Projects)
- [x] Initial Go project structure

### Week 3: Deployment & Cloud ✅ Completed
- [x] GitHub Actions CI/CD setup
- [x] Azure VM deployment
- [x] SSH configuration
- [x] Production deployment
- [x] Custom domain

### Week 4-5: Quality & Containerization ✅ Completed
- [x] Linting setup (`golangci-lint`)
- [x] Branch protection rules
- [x] Docker containerization
- [x] Docker Compose (dev + prod)
- [x] Integration tests
- [x] Dependabot with auto-merge

### Week 6-7: Continuous Delivery ✅ Completed
- [x] Continuous Delivery pipeline (`cd.yml`)
- [x] Docker Compose production deployment
- [x] Database migrations
- [x] Security hardening (fail2ban, CSRF, middleware)
- [x] Password reset flow

### Week 8+: Advanced Topics 🔄 In Progress
- [ ] Terraform infrastructure
- [ ] Monitoring and observability
- [ ] Performance optimization
- [ ] Final presentation

---

## 🤝 Contributing

1. Create a new branch from `main` following the branch naming convention
2. Make your changes
3. Write/update tests
4. Follow commit conventions
5. Create a Pull Request targeting `main`

---

## 📄 License

This project is part of EK's DevOps module 2026.

---

**Course:** DevOps 2026  
**Institution:** EK København  
**Instructor:** Anders Latif
