# API

This project uses Ent schemas in `src/ent/schema`, MySQL 8, and Atlas versioned migrations in `migrations/`.
Migration files and `migrations/atlas.sum` must be committed together. Never edit a migration after it has
been applied to production; create a new migration to correct or reverse a change.

## Posts API

All post, like, and comment endpoints require the `session` cookie issued by `/api/auth/login` or
`/api/auth/register`. The cookie is HTTP-only, Secure, SameSite=Strict, and its Redis session expires
after seven days. `GET /api/auth/me` verifies the current session; `POST /api/auth/logout` revokes it
and clears the cookie.

- `GET /api/posts?page=1&limit=20` returns a stable, newest-first page (maximum page size 100);
  `GET /api/posts/mine?page=1&limit=20` returns the current user's posts,
  `GET /api/posts/mine/summary` returns their total post, like, and comment counts, and
  `GET /api/posts/:postID` returns one post.
- `POST /api/posts` with `{ "title": "...", "content": "..." }`
- `PUT /api/posts/:postID` with `{ "title": "...", "content": "..." }`
- `DELETE /api/posts/:postID` (only the author can update or delete a post)
- `POST /api/posts/:postID/likes` to like; `DELETE /api/posts/:postID/likes` to unlike
- `GET /api/posts/:postID/comments` returns a nested `replies` tree
- `POST /api/posts/:postID/comments` with `{ "content": "...", "parent_id": 123 }`;
  omit `parent_id` for a top-level comment
- `PUT /api/comments/:commentID` with `{ "content": "..." }`
- `DELETE /api/comments/:commentID` (deletes the owned comment and promotes its direct replies)

Post/comment writes and comment deletion are restricted to their author. A user can like each post once.
The migration `20260926170505_create_posts_comments_and_likes.sql` adds the storage tables.

Post, comment, and like each have their own domain type, Ent schema, repository, service, and handler.
Request and response DTOs are split by resource; services map domain entities to response DTOs before
returning data to handlers. Deleting a post removes its comments and likes in one transaction; deleting
a comment preserves other authors' replies by promoting direct replies to its parent. Post titles are
limited to 200 characters, post content to 10,000, and comment content to 5,000.

## Install tools

- Go: install the version required by `go.mod`.
- Make: install GNU Make (available in WSL/Linux/macOS; on Windows use WSL or install Make).
- Atlas CLI:
  - macOS/Linux: `curl -sSf https://atlasgo.sh | sh` or `brew install ariga/tap/atlas`
  - Windows: download the latest `atlas-windows-amd64` binary from
    <https://release.ariga.io/atlas/atlas-windows-amd64-latest.exe>, rename it to `atlas.exe`, and add its
    directory to `PATH`.
- Docker is required for the temporary MySQL dev database used by Atlas to calculate schema diffs. Run
  Atlas from the host/WSL environment so it can invoke Docker and load the Ent schema from this Go module.

## Configure the database URL

Copy `.env.example` to `.env`, edit `ATLAS_DB_URL` as needed, then export it before running Make:

```sh
set -a
. ./.env
set +a
```

Use a MySQL URL such as `mysql://user:password@127.0.0.1:3306/kyoani_db`. URL-encode special
characters in credentials. Do not commit `.env`; use your deployment secret store for staging/production.
The actual URL is read from `ATLAS_DB_URL`; no database credentials are stored in `atlas.hcl`.

## Configure the application

The API loads `src/config/config.yml` when present. Environment variables override the corresponding
nested YAML fields, so production settings can be supplied through the hosting provider without
committing credentials. If the YAML file is missing, configuration starts empty and environment
variables can provide the settings instead. Non-empty environment values override YAML. Malformed YAML
or invalid non-empty integer environment values cause startup to fail with an error.
Invalid non-empty boolean environment values also cause startup to fail.

| Environment variable | Config field | Type |
| --- | --- | --- |
| `SERVER_HOST` | `server.host` | string |
| `SERVER_PORT` | `server.port` | integer |
| `CORS_ALLOWED_ORIGINS` | `cors.allowed_origins` | comma-separated origins |
| `DB_HOST` | `database.host` | string |
| `DB_PORT` | `database.port` | integer |
| `DB_USER` | `database.user` | string |
| `DB_PASSWORD` | `database.password` | string |
| `DB_NAME` | `database.name` | string |
| `DB_TLS` | `database.tls` | boolean |
| `PASSWORD_COST` | `password.cost` | integer |
| `REDIS_HOST` | `redis.host` | string |
| `REDIS_PORT` | `redis.port` | integer |
| `REDIS_PASSWORD` | `redis.password` | string |
| `REDIS_DB` | `redis.db` | integer |
| `REDIS_TLS` | `redis.tls` | boolean |

For local runs, export the variables from `.env` before starting the API; the application does not
load `.env` automatically. Vercel and other hosting platforms expose configured environment variables
to the application automatically. `PORT` is used as the server port when `SERVER_PORT` is not set.
Set `SERVER_HOST` to `0.0.0.0` in deployments that require binding to all network interfaces.
Comma-separated CORS origins are trimmed around each origin.

## Deploy on Vercel with Aiven MySQL

Deploy the API as its own Vercel project with the project root set to `api` and the Go framework
preset. Configure these environment variables in Vercel (use the Aiven service's current
credentials; do not commit them):

| Variable | Value |
| --- | --- |
| `SERVER_HOST` | `0.0.0.0` |
| `DB_HOST` | Aiven hostname |
| `DB_PORT` | Aiven port |
| `DB_USER` | Aiven username |
| `DB_PASSWORD` | Aiven password |
| `DB_NAME` | `defaultdb` |
| `DB_TLS` | `true` |
| `CORS_ALLOWED_ORIGINS` | Exact deployed frontend origin, including `https://` |

The API also requires a reachable Redis service for sessions. Configure `REDIS_HOST`,
`REDIS_PORT`, `REDIS_PASSWORD`, and `REDIS_TLS` from a managed Redis provider; the server will
fail during startup if Redis cannot be reached.

Deploy the React app as a second Vercel project with the project root set to `react`. Set
`VITE_API_URL` to the deployed API's base URL ending in `/api`, for example
`https://your-api.vercel.app/api`. This is a frontend build-time variable, so redeploy the
frontend after changing it. Do not use `localhost` as the deployed API URL.

## Standard schema change workflow

1. Edit a schema under `src/ent/schema/*.go`.
2. Regenerate Ent code: `go generate ./src/ent`.
3. Create a versioned SQL migration: `make migrate-diff name=add_user_phone_number`.
4. Review the generated SQL carefully, especially `ALTER`, `DROP`, nullability, and data-preservation behavior.
5. Commit the Ent schema/code, migration SQL, and `migrations/atlas.sum` together. Do not rewrite a migration
   that has reached production.
6. Check and apply pending migrations: `make migrate-status`, then `make migrate-apply`.

`make migrate-lint` validates migration checksums and lints the latest migration. Run it in CI. There is
currently no CI/CD pipeline in this repository; deployment should run `make migrate-apply` with
`ATLAS_DB_URL` supplied from CI secrets before starting the service. Keep production migration approval
separate from application startup.

`make migrate-down` uses Atlas to revert the latest migration and can cause data loss. Use it only for a
deliberate local/staging rollback after reviewing the plan. Prefer a new forward migration in shared or
production environments.

## Worked example: add `phone_number` to User

The `User` Ent schema now declares `phone_number` as nullable. After updating the schema:

```sh
go generate ./src/ent
make migrate-diff name=add_user_phone_number
```

Review the generated SQL. It should add a nullable `phone_number` column to `users`; the example migration
is in `migrations/20260926170000_add_user_phone_number.sql`. Commit that file and the updated
`migrations/atlas.sum` with the schema change. Start MySQL, then check/apply it:

```sh
make migrate-status
make migrate-apply
```

The schema change and migration file are already included as a runnable example in this repository.

## Atlas Cloud

For a team, optionally push the migration directory to an Atlas Schema Registry and configure protected
deployment approvals. This centralizes migration history and review; keep database URLs in CI secrets.
