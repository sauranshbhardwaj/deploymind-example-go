---
name: deploymind
description: Deploy this app to DeployMind, the company's internal cloud. Use whenever the user asks to deploy, ship, publish, redeploy, roll back, or check the status/logs of this app. Workflow is validate → push → plan → present → approve → deploy. Never read or print .env.
---

# DeployMind deploy skill

You are the builder's coding agent. DeployMind gives this app a stable private URL behind the
company login, a managed Postgres, and secrets, from one strict config file: `deploymind.yaml`.
The platform is deterministic: it validates the config, builds the pinned commit, and promotes
only what passes. You do the adapting; it does the running.

## Hard rules

1. **Never read, cat, print, or paste `.env`** (or any secret value). The only thing you run with it
   is `deploymind secrets upload .env`, which prints names and digests only. If the builder pastes
   a secret into chat, tell them to put it in `.env` instead.
2. **Present the plan and wait for the builder's go-ahead before `deploymind deploy`.**
   The plan lists files, dependencies, env names, resources, members, and warnings.
3. Work in this order, every time: `validate` → commit → push → `plan` → present → `deploy`.
   `plan` and `deploy` refuse dirty or unpushed trees; do not use `--allow-dirty` unless the
   builder explicitly asks.
4. The first deploy of an app is private (`visibility: private`); widen it later.
5. Relay the platform's messages verbatim; they are written for you.

## Commands

| Step | Command | What to tell the builder |
|---|---|---|
| Log in (once) | `deploymind login --org <org>` | Open the printed URL, confirm the code. |
| Adopt the app | `deploymind init` | Creates `deploymind.yaml`, gitignores `.env`, registers the app. Review the manifest with them. |
| Check config | `deploymind validate` | Fix every issue it lists; codes are stable (e.g. `env_reserved`). |
| Publish code | `git add -A && git commit -m ... && git push` | The platform fetches the pushed commit. |
| Plan | `deploymind plan` | Show the summary and every warning verbatim. Ask: "Deploy this?" |
| Secrets | `deploymind secrets upload .env` | Only when `plan` lists env names. Show the digests, never values. |
| Deploy | `deploymind deploy` | Streams narration; ends with the URL. On failure see below. |
| Afterwards | `deploymind status`, `deploymind logs [--tail]`, `deploymind releases`, `deploymind rollback --to <id> --confirm` | Health with a plain reason; console output; history; rollback. |

## When something is refused

- `drift`: the manifest or commit changed since the plan. Commit, push, run `deploymind plan` again.
- `missing_secrets`: the message lists the names; ask the builder to put values in `.env`, then `deploymind secrets upload .env`.
- `manifest_invalid` / validate issues: fix `deploymind.yaml`; the issue path and code say exactly what.
- `framework_unavailable` (`next`), `execution_unavailable` (`serverless`), `primitive_unavailable` (files, llm, cron, connectors, egress): not in this MVP; use `framework: custom` with your own Dockerfile if the app cannot fit `go`, `python`, `vite`, or `static`.
- Build failed: `deploymind releases` shows the failing release; the failure reason is buildah's last error line. Fix the code or the Dockerfile and redeploy.
- Health failed (promotion): the previous release stays live. Check `health.path` returns 200 on `run.port`, and that the app listens on `0.0.0.0:$PORT`.

## The manifest (deploymind.yaml), MVP fields

```yaml
schema: 1
app: my-app                 # immutable slug → https://my-app.<org>.<base>
framework: go               # go | python | vite | static | custom  (next: Beta)
build:
  dockerfile: Dockerfile    # only for framework: custom
  command: null             # override the template's build command (single line)
run:
  command: null             # override the start command (single line)
  port: 3000
health:
  path: /                   # must return 200 on run.port
  timeout_seconds: 30
resources: { cpu: 0.5, memory: 512Mi, ports: [] }
visibility: private         # private | org  (public: Beta)
access:
  default_role: viewer
  routes: [{ path: /admin/*, role: editor }]
  members: [{ user: alice@company.com, role: editor }]   # owner is the creator
env: [API_KEY]              # NAMES only
primitives:
  db: { enabled: true }     # injects DATABASE_URL (managed Postgres)
```

Templates: `go` builds `.` (or `build.command` producing `/out/app`); `python` installs
`requirements.txt` or `pyproject.toml` and runs `uvicorn main:app`; `vite` runs the lockfile's
package manager, serves `dist/` with SPA fallback; `static` serves the repo root; `custom` uses
your Dockerfile. Containers run with every Linux capability dropped and no privilege escalation: listen on a port above 1024, and avoid binaries that carry file capabilities (the official `caddy` image needs `RUN setcap -r /usr/bin/caddy`; the official `nginx` image cannot bind :80, so use `nginxinc/nginx-unprivileged` or busybox `httpd`). Every pod gets `PORT`, `DEPLOYMIND_APP`, `DEPLOYMIND_ORG`, `DEPLOYMIND_RELEASE`,
and `DEPLOYMIND_IDENTITY_KEY` (public key to verify the `X-DeployMind-Identity` header the
proxy adds: email, role, org, app).
