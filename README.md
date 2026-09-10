# MiStatic

A highly minimal, on-demand static site hosting platform powered by PocketBase, SvelteKit, and Go.

My main motivation was wanting to deploy quick static sites for demos, or testing, or AI POCs and with most self-hosting solutions like Caprover id need to include a dockerfile for nginx or something to bundle up a web server.

Now, I can click "new site", drop an index.html or a zip and its live.

Since its built on Pocketbase, you instantly get API support (for MCP, or skills or whatever) easy social logins and much more.

## Features

- Custom domain support
- Subpath support so you dont need to bother with wildcard certs and stuff
- No containers needed for your projects, pure static files
- No need to edit any nginx config files
- Perfomant request logging with forwarded headers support
- Mutable deployments (edit files in a live deployment)
- Private sites with SSO
- Deployments and rollbacks

## Screenshots

![Create New Site](./screenshots/create-new-site.png)
![Manage Site](./screenshots/manage-site.png)
![File Browser](./screenshots/file-browser.png)

## Docker Deployment

We use a unified multi-stage `Dockerfile`.
When running in production, map both `/app/sites` and `/pb/pb_data` to volumes. The
first preserves site deployments, while `/pb/pb_data` preserves PocketBase's SQLite
database and other application data.

```bash
docker build -t mistatic .
docker run -p 8090:8090 \
	-v ./sites:/app/sites \
	-v ./pb_data:/pb/pb_data \
	-e APP_DOMAIN=app.example.com \
	-e ROOT_DOMAIN=example.com \
	mistatic
```

**Note on SSL:** MiStatic routes HTTP traffic. Put it behind an auto-SSL reverse proxy like **Caddy** or **Traefik** for custom domain HTTPS.

## Architecture

- **PocketBase (Go):** Manages metadata, file routing, and serves the dashboard UI.
- **SvelteKit (Static HTML):** Embedded into the Go binary. A lightweight UI to manage your sites.
- **Tailwind CSS:** GitHub-inspired styles for a clean, minimal interface.

## Local Development Setup

### 1. Configure OS hosts file
To test routing locally, you'll need to fake DNS for your local domains. Edit `/etc/hosts` (Linux/Mac) or `C:\Windows\System32\drivers\etc\hosts` (Windows) and add:

```
127.0.0.1 mistatic.local
127.0.0.1 app.mistatic.local
127.0.0.1 my-project.mistatic.local
```
*(Add additional lines for any custom domains or subdomains you create.)*

### 2. Run the application

Start both the PocketBase backend and SvelteKit frontend simultaneously:

```bash
npm run dev
```

### 3. Access

- **Dashboard:** [http://app.mistatic.local:8090](http://app.mistatic.local:8090)
- **Hosted Sites:** `<subdomain>.mistatic.local:8090`
- **PocketBase Admin:** [http://app.mistatic.local:8090/_/](http://app.mistatic.local:8090/_/)
