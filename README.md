# SyntaxDevopsSquad - SDS 

Welcome to the **SyntaxDevopsSquad** main repository. This project is part of our 2026 DevOps module, focusing on automation, transparency, and professional development conventions.

## Tech Stack
- **Language:** Go 1.24+
- **Database:** SQLite
- **Infrastructure:** Terraform & Azure CLI
- **Runtime:** WSL (Ubuntu 24.04)

## Conventions

### Git Commit Standard
We follow **Conventional Commits** to ensure a clean and readable history:
- `feat`: New functionality.
- `fix`: Bug fixes.
- `refactor`: Code optimization without changing logic.
- `docs`: Documentation (README, reports).
- `ci`: CI/CD (GitHub Actions, build scripts).
- `test`: Unit or integration tests.
- `style`: Formatting and style.

**Format:** `<type>: <description>`  
*Example:* `feat: setup Go project with SQLite and database initialization`

## Getting Started

### Prerequisites
Ensure Go is installed in your WSL/Linux terminal:
```bash
go version
