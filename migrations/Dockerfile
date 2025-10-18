FROM postgres:14.1-alpine

# Copy the migration SQL into the image so the job can run it directly.
COPY 001_initial_schema.up.sql /migrations/001_initial_schema.up.sql

WORKDIR /migrations

CMD ["/bin/sh", "-c", "psql -h yb-tservers.go-yuga-dist.svc.cluster.local -p 5433 -U $PGUSER -d $PGDATABASE -f /migrations/001_initial_schema.up.sql -v ON_ERROR_STOP=1"]
