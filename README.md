# SyntaxDevopsSquad - WhoKnows Migration Project

Welcome to the **SyntaxDevopsSquad** main repository. This project is part of our 2026 DevOps module at KEA, focusing on migrating a legacy Python Flask application to Go while learning DevOps practices including automation, CI/CD, and infrastructure as code.

##  Project Overview

**WhoKnows** is a web application for searching and managing wiki-style pages with user authentication. We are migrating the application from Python/Flask to Go as part of our DevOps learning journey. Our team of 4 developers, none with prior Go experience, is tackling this challenge while implementing modern DevOps practices.

### Core Functionality
- **User Authentication:** Registration, login, and session management
- **Page Management:** Create, read, and search wiki-style pages
- **Database:** SQLite for user and page storage

### Team Members
- **CodeByNajib** (NajibGPT)
- **AceS0**
- **MarcusLieberH**
- **Daniel23894** (Daniel Søgaard)

### Migration Strategy
We are systematically migrating each component from Python to Go:
1. Database layer and schema
2. User authentication system
3. Page management functionality
4. Template rendering
5. Static file serving
6. Testing and validation

##  Tech Stack

### Backend
- **Language:** Go 1.25.0
- **Database:** SQLite with `modernc.org/sqlite`
- **Session Management:** Gorilla Sessions
- **Legacy:** Python Flask (original implementation)

### Legacy Python Dependencies (for reference)
- Flask 3.1.2
- Werkzeug 3.1.5
- Jinja2 3.1.6
- click 8.3.1
- blinker 1.9.0
- itsdangerous 2.2.0
- MarkupSafe 3.0.3
- sqlite3 (built-in)
- hashlib (built-in)

### Infrastructure & DevOps
- **Cloud Platform:** Azure (Azure for Students)
- **IaC:** Terraform
- **Containerization:** Docker
- **CI/CD:** GitHub Actions
- **Development Environment:** WSL (Ubuntu 24.04)
- **Version Control:** Git with Conventional Commits

### Database Schema
- **users table:** User authentication and profiles
- **pages table:** Wiki-style content storage

##  Project Structure

```
devops-syntaxsquad/
├── implementations/
│   ├── python/          # Legacy Flask implementation (reference)
│   │   ├── app.py
│   │   ├── schema.sql
│   │   ├── templates/
│   │   ├── static/
│   │   └── run_forever.sh
│   └── go/              # Go backend implementation (active development)
│       ├── backend/     # Go source code
│       ├── static/      # CSS, JS, images
│       │   └── style.css
│       ├── templates/   # HTML templates
│       │   ├── search.html
│       │   ├── login.html
│       │   ├── register.html
│       │   └── about.html
│       ├── go.mod       # Go dependencies
│       ├── go.sum       # Dependency checksums
│       ├── schema.sql   # Database schema
│       └── whoknows.db  # SQLite database
├── .github/
│   └── workflows/       # CI/CD pipelines
├── docs/
│   └── dependency_graph.dot  # System architecture visualization
└── README.md
```

##  Getting Started

### Prerequisites

**Required Software:**
- Go 1.25.0 or higher
- SQLite3
- Git
- WSL/Linux environment (for Windows users)

**Optional:**
- Docker Desktop
- Azure CLI (`az`)
- Terraform

### Installation

1. **Clone the repository:**
```bash
git clone https://github.com/SyntaxDevopsSquad-SDS/devops-syntaxsquad.git
cd devops-syntaxsquad/implementations/go
```

2. **Verify Go installation:**
```bash
go version
# Should output: go version go1.25.0 or higher
```

3. **Install dependencies:**
```bash
go mod download
```

4. **Initialize the database:**
```bash
sqlite3 whoknows.db < schema.sql
```

5. **Run the application:**
```bash
go run main.go
```

6. **Access the application:**
Open your browser and navigate to `http://localhost:8080`

### CSRF Simulation Mode

The Go backend protects login/register with CSRF tokens by default. For controlled
black-box simulations that call API endpoints directly, you can relax this check:

```env
CSRF_RELAXED=true
```

Guidelines:
- Use `CSRF_RELAXED=false` in normal/prod operation.
- Enable `CSRF_RELAXED=true` only when simulation tooling cannot handle form CSRF flow.

##  Development Workflow

### System Architecture

The original Python Flask application follows this dependency structure:
- **app.py** → Main Flask application
- **Database** → SQLite with users and pages tables
- **Templates** → Jinja2 HTML templates (search, login, register, about)
- **Static** → CSS styling
- **Dependencies** → Flask ecosystem (Werkzeug, Jinja2, Click, etc.)

See `docs/dependency_graph.dot` for the complete system architecture visualization.

### Git Commit Conventions

We follow **Conventional Commits** for clean and readable history:

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New functionality | `feat: add user authentication` |
| `fix` | Bug fixes | `fix: resolve database connection issue` |
| `refactor` | Code optimization | `refactor: improve error handling` |
| `docs` | Documentation | `docs: update README with setup steps` |
| `ci` | CI/CD changes | `ci: add Docker build workflow` |
| `test` | Tests | `test: add unit tests for database methods` |
| `style` | Code formatting | `style: format Go code with gofmt` |

**Format:** `<type>: <description>`

**Example:**
```bash
git commit -m "feat: implement check database exists method"
```

### Branch Strategy

- `main` - Production-ready code
- `develop` - Integration branch
- `feature/*` - New features
- `fix/*` - Bug fixes

### Code Quality

**Before committing:**
```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Check for errors
go vet ./...
```


### DevOps Literature
Required reading throughout the course:
- DevOps Literature I (Week 6)
- DevOps Literature II (Week 7)
- Detecting Agile BS

### Course Topics by Week
1. **Week 1:** Git basics, legacy codebase analysis, dependency graphs
2. **Week 2:** Conventions, OpenAPI, environment variables, framework selection
3. **Week 3:** GitHub Actions, Azure deployment, SSH, CI/CD fundamentals
4. **Week 4:** Software quality, linting, branching strategies, technical debt
5. **Week 5:** Docker fundamentals, containerization, packaging
6. **Week 6:** Docker Compose, Continuous Delivery, Agile & DevOps principles
7. **Week 7:** DevOps culture, psychological safety, pipeline optimization

##  Known Issues

Track issues and feature requests in our [GitHub Issues](https://github.com/SyntaxDevopsSquad-SDS/devops-syntaxsquad/issues).

##  Project Milestones

### Week 1-2: Foundation ( Completed)
- [x] Legacy Python codebase analysis
- [x] Dependency graph creation
- [x] Framework selection (Go)
- [x] OpenAPI specification generation
- [x] Kanban board setup (GitHub Projects)
- [x] Initial Go project structure

### Week 3: Deployment & Cloud ( In Progress)
- [x] GitHub Actions CI/CD setup
- [x] Azure VM deployment
- [ ] SSH configuration
- [ ] Production deployment

### Week 4-5: Quality & Containerization ( Upcoming)
- [ ] Linting setup
- [ ] Branch protection rules
- [ ] README badges
- [ ] Docker containerization
- [ ] Docker Compose configuration
- [ ] Postman monitoring

### Week 6-7: Continuous Delivery ( Planned)
- [ ] Live reload in Docker
- [ ] Workflow optimization
- [ ] Continuous Delivery pipeline
- [ ] Performance optimization

### Week 8+: Advanced Topics ( Future)
- [ ] Complete migration from Python to Go
- [ ] Terraform infrastructure
- [ ] Monitoring and observability
- [ ] Final presentation

##  Contributing

1. Create a new branch from `develop`
2. Make your changes
3. Write/update tests
4. Follow commit conventions
5. Create a Pull Request

##  License

This project is part of EK's DevOps module 2026.

##  Contact

For questions or collaboration, reach out to any team member via GitHub.

---

**Course:** DevOps 2026  
**Institution:** EK København  
**Instructor:** Anders Latif
