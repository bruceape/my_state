# API
Exists as a basic gateway to LLMs. Will have a hand rolled auth system that a client can authenticate against.

## Usage
### Requirements
- Go v1.25
- Postgres DB

### Environment Variables
Expects a the following environment variables to be set.
- OPENAI_API_KEY
- OPENAI_PROJECT_ID
- OPENAI_ORG_ID
- DATABASE_URL

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
