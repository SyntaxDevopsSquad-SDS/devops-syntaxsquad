# Service Level Agreement (SLA)
## WhoKnows — SyntaxSquad

**Version**: 1.0  
**Effective Date**: May 15, 2026  
**Provider**: SyntaxSquad (CodeByNajib, AceS0, MarcusLieberH, Daniel23894)  
**Service URL**: [https://syntax-reborndev.com](https://syntax-reborndev.com)

---

## 1. Service Scope

This SLA covers the **WhoKnows** web application and its supporting infrastructure. The following services are included:

| Component | Description |
|---|---|
| **Web Application** | Full-stack Go web app accessible at `https://syntax-reborndev.com` |
| **Search API** | `GET /api/search` — full-text search over wiki-style pages |
| **Authentication** | User registration, login, session management, and password reset |
| **Health Endpoint** | `GET /health` — machine-readable availability check |
| **Monitoring Dashboard** | Grafana dashboard at `https://monitor.syntax-reborndev.com` |

The following are **excluded** from this SLA:
- Third-party DNS resolution (Cloudflare, DigitalOcean DNS)
- Client-side network or device issues
- Scheduled maintenance windows (announced ≥ 24 hours in advance)

---

## 2. Performance Metrics (SLIs & SLOs)

### 2.1 Availability (Uptime)

| Metric | Target (SLO) | Measurement |
|---|---|---|
| Monthly Uptime | **≥ 99.0%** | `100 * avg_over_time(up{job="whoknows-go-backend"}[$__range])` via Prometheus |
| Health endpoint reachability | `GET /health` returns HTTP 200 | Prometheus scrape every 15 seconds |

**99.0% uptime** allows for a maximum of **~7 hours 18 minutes** of downtime per month.

### 2.2 Response Time

| Endpoint | Target (p95) |
|---|---|
| `GET /` — Search page | < 500 ms |
| `GET /api/search` | < 1000 ms |
| `POST /login` | < 800 ms |
| `GET /health` | < 100 ms |

Response time is measured at the server side via `whoknows_http_request_duration_seconds` (Prometheus histogram).

### 2.3 Error Rate

- HTTP 5xx responses: **< 1%** of all requests in any given hour

---

## 3. Response and Resolution Times

| Severity | Definition | Initial Response | Resolution Target |
|---|---|---|---|
| **Critical** | Full service unavailability (uptime < 99%) | 1 hour | 4 hours |
| **High** | Core feature degraded (search/auth broken) | 4 hours | 24 hours |
| **Medium** | Non-critical feature impacted, workaround exists | 24 hours | 72 hours |
| **Low** | Minor UI issue or cosmetic bug | 72 hours | Next release |

Incidents are tracked via GitHub Issues on the project repository. Critical incidents trigger automated Grafana alerts via the configured contact points.

---

## 4. Maintenance Windows

- **Scheduled maintenance** will be announced at least **24 hours in advance** via the project GitHub repository.
- Scheduled downtime does not count toward uptime calculations.
- Best-effort target: maintenance between **02:00–04:00 CET** on weekdays.

---

## 5. Security Measures

The following protections are in place for the duration of this agreement:

| Measure | Implementation |
|---|---|
| **CSRF Protection** | Token-based CSRF middleware on all state-changing endpoints |
| **Session Security** | Server-side sessions stored in PostgreSQL; signed cookies via `SECRET_KEY` |
| **Brute-force Protection** | Fail2ban configured on the host server (`server-config/fail2ban-jail.local`) |
| **Transport Security** | HTTPS enforced via TLS (Caddy reverse proxy) |
| **Secrets Management** | No secrets committed to version control; `.env` is gitignored |
| **Breach Response** | Documented and scripted breach response procedure (`scripts/breach_response.sh`) |

Security incidents (e.g. confirmed data breach) will be disclosed to affected users within **72 hours** of discovery.

---

## 6. Compliance Standards

| Standard | Status |
|---|---|
| **HTTPS / TLS** | Enforced — plain HTTP redirected to HTTPS |
| **GDPR (data minimisation)** | Only necessary user data (username, hashed password) is stored |
| **Password storage** | Passwords are hashed using `bcrypt` — never stored in plaintext |
| **Dependency management** | Go modules with pinned versions; reviewed via CI pipeline |

---

## 7. Monitoring and Reporting

- Metrics are collected by **Prometheus** (15-second scrape interval) and visualised in **Grafana**.
- Uptime, request rates, error rates, and session activity are tracked continuously.
- Historical uptime is available on the monitoring dashboard at [https://monitor.syntax-reborndev.com](https://monitor.syntax-reborndev.com).

---

## 8. Limitations and Exclusions

This SLA does not apply under the following circumstances:

- Force majeure (natural disasters, large-scale cloud provider outages)
- DDoS attacks exceeding mitigation capacity
- Breaches caused by compromised user credentials outside our control
- Downtime during announced maintenance windows

---

*This SLA is provided in good faith as part of an academic DevOps project at EK (2026). It reflects real infrastructure and operational practices.*
