# SyntaxDevopsSquad (SDS) 

Repository for vores DevOps-projekt. Vi bygger en skalerbar web-applikation med Go og Terraform.

## Tech Stack
- **Backend:** Go 1.24+
- **Database:** SQLite
- **Infrastructure:** Terraform
- **Konventioner:** Conventional Commits

## Git Konventioner
Vi bruger formatet: `<type>: <beskrivelse>`
- `feat`: Ny funktionalitet.
- `fix`: Fejlrettelse.
- `ci`: Ændringer i build/workflows.
- `docs`: Dokumentation.

## Kom i gang
1. Initialisér databasen:
   ```bash
   go run init_db.go
