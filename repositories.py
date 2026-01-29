GROUP_REPOS = [
        {
            "name": "Example Group",
            "gitLinks": ["https://<git_link>"],
            "backend": "http(s)://<IP_DOMAIN>/<APIURL>",
            "frontend": "http(s)://<IP_DOMAIN>/<FrontEndURL>",
            "monitoring": "http(s)://<IP_DOMAIN>/<MonitoringURL>",
            "stack": ["Flask", "Svelte", "CouchDB", "Redis"],
            "documentation": ["link to documentation", "another link if it applies", "et cetera"],
            "sla": "link to sla",
        }
]
