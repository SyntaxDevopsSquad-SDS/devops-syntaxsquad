# Prompt til separat monitoring-repo (Prometheus + Grafana)

Brug denne prompt direkte i AI/chat i det nye repository:

```text
Lav et komplet, production-minded monitoring repository til WhoKnows med Docker Compose, så det kan deployes på en separat VM.

Krav:
1) Stack
- Prometheus
- Grafana
- Valgfrit: node-exporter (for VM metrics)

2) Struktur
- docker-compose.yml
- prometheus/prometheus.yml
- grafana/provisioning/datasources/datasource.yml
- grafana/provisioning/dashboards/dashboards.yml
- grafana/dashboards/whoknows-overview.json
- .env.example
- README.md med setup/deploy/troubleshooting

3) Prometheus config
- Scrape app backend endpoint: http://<APP_VM_IP>:8080/metrics
- job_name: whoknows-go-backend
- scrape_interval: 15s
- retention: mindst 7d

4) Grafana provisioning
- Automatisk datasource til Prometheus
- Automatisk dashboard import ved startup
- Admin credentials via env vars (ikke hardcoded)

5) Dashboard (whoknows-overview)
Skal indeholde paneler for:
- HTTP requests total (rate + total)
- HTTP requests fordelt på statuskode
- Request latency (p95) baseret på `whoknows_http_request_duration_seconds`
- Successful logins (`whoknows_login_attempts_total{outcome="success"}`)
- Failed logins (`whoknows_login_attempts_total{outcome="failure"}`)
- Successful registrations (`whoknows_registrations_total{outcome="success"}`)
- Registrations der fejler/validering
- Searches total
- Searches fordelt på query-term (topk)
- Mulighed for at filtrere på query og language (Grafana variables hvis muligt)

6) Security + drift
- Brug named volumes til data persistence
- Expose kun nødvendige porte (Prometheus 9090, Grafana 3000)
- Skriv tydeligt i README hvordan firewall/NSG skal åbnes mellem monitoring VM og app VM
- README skal inkludere backup/restore af dashboards og Prometheus data

7) Deploy usability
- Kommandoer i README:
  - docker compose pull
  - docker compose up -d
  - docker compose ps
  - docker compose logs -f
- Inkluder healthcheck for Prometheus og Grafana

8) GitHub
- Tilføj .gitignore
- Tilføj GitHub Actions workflow der validerer compose + Prometheus config på PR
- Workflow må ikke deploye automatisk, kun validate

9) Output format
- Generer alle filerne med konkret indhold
- Forklar kort hvorfor hver fil er nødvendig
- Vis et eksempel på .env filværdier

App metrics der allerede findes i backend:
- whoknows_http_requests_total{method,path,status}
- whoknows_http_request_duration_seconds{method,path}
- whoknows_login_attempts_total{outcome}
- whoknows_registrations_total{outcome}
- whoknows_searches_total{source,language,query,outcome}
```

## Note

Erstat `<APP_VM_IP>` med den IP eller DNS som monitoring-VM kan nå backend på.
