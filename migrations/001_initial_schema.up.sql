-- Note: Database setup (database creation, user creation, and tablespaces) 
-- is handled by the db-setup-job.yaml

-- Create the regions enum type only if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'region') THEN
        CREATE TYPE region AS ENUM ('IND', 'USA', 'CHN', 'DEU', 'SGP');
    END IF;
END$$;

-- Ensure pgcrypto (or yb_extension equivalent) is available for gen_random_uuid()
-- In Yugabyte, pgcrypto is available; this is safe to run if not present.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- dblink is used to run commands in the 'postgres' maintenance DB when creating
-- databases, users, and tablespaces from within a single script.
CREATE EXTENSION IF NOT EXISTS dblink;

-- Create a global emails table for unique email addresses across all regions
-- Create a global users table to enforce globally unique user ids and emails.
-- This is the authoritative table for user identity and global uniqueness.
CREATE TABLE IF NOT EXISTS global_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_address VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create the users table with geo-partitioning
CREATE TABLE IF NOT EXISTS users (
    region region NOT NULL,
    id UUID DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email_address VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (region, id)
) PARTITION BY LIST (region);

-- Create partitions for each region
CREATE TABLE IF NOT EXISTS users_ind PARTITION OF users FOR VALUES IN ('IND');
CREATE TABLE IF NOT EXISTS users_usa PARTITION OF users FOR VALUES IN ('USA');
CREATE TABLE IF NOT EXISTS users_chn PARTITION OF users FOR VALUES IN ('CHN');
CREATE TABLE IF NOT EXISTS users_deu PARTITION OF users FOR VALUES IN ('DEU');
CREATE TABLE IF NOT EXISTS users_sgp PARTITION OF users FOR VALUES IN ('SGP');

-- Note: Do NOT SET TABLESPACE in YugabyteDB unless you have configured
-- appropriate tablespaces. Pinning partitions to tablespaces is not
-- required for geo-partitioning and can cause migration failures.

-- Create the posts table with geo-partitioning
CREATE TABLE IF NOT EXISTS posts (
    id UUID DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    region region NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT posts_pkey PRIMARY KEY (region, id),
    CONSTRAINT posts_user_fkey FOREIGN KEY (region, user_id) REFERENCES users(region, id)
) PARTITION BY LIST (region);

-- Create partitions for each region
CREATE TABLE IF NOT EXISTS posts_ind PARTITION OF posts FOR VALUES IN ('IND');
CREATE TABLE IF NOT EXISTS posts_usa PARTITION OF posts FOR VALUES IN ('USA');
CREATE TABLE IF NOT EXISTS posts_chn PARTITION OF posts FOR VALUES IN ('CHN');
CREATE TABLE IF NOT EXISTS posts_deu PARTITION OF posts FOR VALUES IN ('DEU');
CREATE TABLE IF NOT EXISTS posts_sgp PARTITION OF posts FOR VALUES IN ('SGP');

-- Note: Avoid tablespace pinning unless explicitly configured in your
-- YugabyteDB cluster. Geo-partitioning is achieved via partitioned tables
-- and placement policies configured at the DB/cluster level.

-- Create the sessions table
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
