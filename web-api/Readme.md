# API
Exists as a basic gateway to LLMs. Contains hand rolled auth system that a client can authenticate against.

## Usage
### Requirements
- Go v1.25
- Postgres DB

### Environment Variables
Expects the following environment variables to be set.
- OPENAI_API_KEY=""
- OPENAI_PROJECT_ID=""
- OPENAI_ORG_ID=""
- DATABASE_URL=""
- JWT_SECRET=""
- JWT_TTL=15m


### Setup
To run, simply run `go run .` in the root directory.

## Tools
I've included a set of migration helper commands.

### Initialize
Creates the base `schema_migrations` table and other basic setup.

```
go run ./cmd/migrations/initialize
```

### New
Creates a blank schema file with the proper version id and optional name.
The migration file will be stored in `db/migrations`, folliwng the pattern `YYYYMMDDHHMMSS_name.sql`

```
go run ./cmd/migrations/new migration_name
```

### Run
Runs all migrations in the `db/migrations` folder that haven't been applied yet, in chronological order. Expects files to follow the pattern `YYYYMMDDHHMMSS_name.sql`

```
go run ./cmd/migrations/run
```

## Testing Tools
I've included a few bash scripts to manually test the endpoints. I'm sure there's a better way to manually test these. They are available in `testing_tools`
