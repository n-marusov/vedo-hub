//! `PostgreSQL` connection pool manager for versioning-service.
//!
//! Provides:
//! - Connection pool initialization from `DATABASE_URL`
//! - Health check via table existence verification
//! - Migration runner (actual migration is called from main.rs using `sqlx::migrate`!)

use sqlx::postgres::PgPoolOptions;
use sqlx::PgPool;

/// Default `PostgreSQL` connection URL
const DEFAULT_DATABASE_URL: &str = "postgres://vedo:vedo@localhost:5432/vedo_versioning";

/// The shared `PostgreSQL` connection pool.
#[derive(Clone)]
pub struct PgPoolWrapper {
    pool: PgPool,
}

impl PgPoolWrapper {
    /// Returns a reference to the inner `PgPool`.
    pub fn pool(&self) -> &PgPool {
        &self.pool
    }
}

/// Configuration for `PostgreSQL` connection.
pub struct PgConfig {
    pub database_url: String,
    pub max_connections: u32,
}

impl Default for PgConfig {
    fn default() -> Self {
        Self {
            database_url: DEFAULT_DATABASE_URL.to_string(),
            max_connections: 10,
        }
    }
}

impl PgConfig {
    /// Loads config from environment variables, falling back to defaults.
    pub fn from_env() -> Self {
        Self {
            database_url: std::env::var("DATABASE_URL")
                .unwrap_or_else(|_| DEFAULT_DATABASE_URL.to_string()),
            max_connections: std::env::var("PG_MAX_CONNECTIONS")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(10),
        }
    }
}

/// Creates a `PostgreSQL` connection pool with the given config.
pub async fn create_pool(config: &PgConfig) -> Result<PgPoolWrapper, sqlx::Error> {
    let pool = PgPoolOptions::new()
        .max_connections(config.max_connections)
        .connect(&config.database_url)
        .await?;

    tracing::info!(
        max_connections = config.max_connections,
        "Connected to PostgreSQL"
    );

    Ok(PgPoolWrapper { pool })
}

/// Runs manual SQL migrations for the versioning schema.
/// Uses CREATE TABLE IF NOT EXISTS so it can be run idempotently.
///
/// Concurrent invocations are serialized with `pg_advisory_lock` so two
/// processes (or two tests in one binary) cannot race on CREATE/DROP INDEX —
/// a race produced duplicate `pg_class_relname_nsp_index` failures.
#[allow(clippy::too_many_lines)]
pub async fn run_manual_migrations(pool: &PgPool) -> Result<(), sqlx::Error> {
    // Lock key shared by every process that migrates the schema.
    const MIGRATION_LOCK_KEY: i64 = 787_896_734;

    // Single dedicated connection holds the advisory lock for the whole
    // migration run — pool-checked-out connections would not share the lock.
    let mut conn = pool.acquire().await?;

    sqlx::query("SELECT pg_advisory_lock($1)")
        .bind(MIGRATION_LOCK_KEY)
        .execute(&mut *conn)
        .await?;

    let result = run_migrations_locked(&mut conn).await;

    // Always release the lock, even on migration failure.
    sqlx::query("SELECT pg_advisory_unlock($1)")
        .bind(MIGRATION_LOCK_KEY)
        .execute(&mut *conn)
        .await?;

    result
}

/// Executes the migration statements while holding the advisory lock.
/// Separate function keeps the lock acquire/release pair tight and readable.
#[allow(clippy::too_many_lines)]
async fn run_migrations_locked(conn: &mut sqlx::postgres::PgConnection) -> Result<(), sqlx::Error> {
    // Migration 001: Initial schema
    sqlx::query(
        r"
        CREATE TABLE IF NOT EXISTS branches (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            ontology_id UUID NOT NULL,
            head_commit_id UUID,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            is_protected BOOLEAN NOT NULL DEFAULT FALSE
        );
        ",
    )
    .execute(&mut *conn)
    .await?;

    sqlx::query(
        r"
        CREATE TABLE IF NOT EXISTS commits (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            branch_id UUID NOT NULL REFERENCES branches(id),
            parent_commit_id UUID REFERENCES commits(id),
            message TEXT NOT NULL,
            author_id VARCHAR(255) NOT NULL,
            author_name VARCHAR(255) NOT NULL,
            delta JSONB NOT NULL DEFAULT '{}',
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        ",
    )
    .execute(&mut *conn)
    .await?;

    // Migration 002: Add indexes
    sqlx::query("CREATE INDEX IF NOT EXISTS idx_commits_branch_id ON commits(branch_id)")
        .execute(&mut *conn)
        .await?;

    sqlx::query("CREATE INDEX IF NOT EXISTS idx_commits_created_at ON commits(created_at DESC)")
        .execute(&mut *conn)
        .await?;

    sqlx::query("CREATE INDEX IF NOT EXISTS idx_branches_ontology_id ON branches(ontology_id)")
        .execute(&mut *conn)
        .await?;

    sqlx::query("CREATE INDEX IF NOT EXISTS idx_branches_name ON branches(name)")
        .execute(&mut *conn)
        .await?;

    // index from 002_add_commit_indexes.sql — missing from previous manual runner
    sqlx::query("CREATE INDEX IF NOT EXISTS idx_commits_author_id ON commits(author_id)")
        .execute(&mut *conn)
        .await?;

    // Migration 003: State snapshots for fast materialization
    sqlx::query(
        r"
        CREATE TABLE IF NOT EXISTS state_snapshots (
            commit_id UUID PRIMARY KEY,
            branch_id UUID NOT NULL,
            triples JSONB NOT NULL DEFAULT '[]',
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        ",
    )
    .execute(&mut *conn)
    .await?;

    sqlx::query(
        "CREATE INDEX IF NOT EXISTS idx_state_snapshots_branch_id ON state_snapshots(branch_id)",
    )
    .execute(&mut *conn)
    .await?;

    // Composite index for efficient commit listing by branch + creation order
    sqlx::query(
        "CREATE INDEX IF NOT EXISTS idx_commits_branch_created ON commits(branch_id, created_at DESC)",
    )
    .execute(&mut *conn)
    .await?;

    // Migration 004: Add constraints (unique index, foreign keys with cascade)
    sqlx::query(
        "CREATE UNIQUE INDEX IF NOT EXISTS idx_branches_ontology_name ON branches(ontology_id, name)",
    )
    .execute(&mut *conn)
    .await?;

    // Drop the non-unique name index now that we have a unique composite index
    sqlx::query("DROP INDEX IF EXISTS idx_branches_name")
        .execute(&mut *conn)
        .await?;

    // Add foreign key constraints to state_snapshots (idempotent via DO block)
    sqlx::query(
        r"
        DO $$ BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM pg_constraint WHERE conname = 'fk_snapshots_commit'
            ) THEN
                ALTER TABLE state_snapshots
                ADD CONSTRAINT fk_snapshots_commit
                FOREIGN KEY (commit_id) REFERENCES commits(id) ON DELETE CASCADE;
            END IF;
        END $$;
        ",
    )
    .execute(&mut *conn)
    .await?;

    sqlx::query(
        r"
        DO $$ BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM pg_constraint WHERE conname = 'fk_snapshots_branch'
            ) THEN
                ALTER TABLE state_snapshots
                ADD CONSTRAINT fk_snapshots_branch
                FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;
            END IF;
        END $$;
        ",
    )
    .execute(&mut *conn)
    .await?;

    tracing::info!("Manual migrations completed successfully");
    Ok(())
}

/// Runs a health check by verifying that the `branches` and `commits` tables exist.
pub async fn health_check(pool: &PgPoolWrapper) -> bool {
    let query =
        "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'branches')";
    let result: Result<Option<bool>, _> = sqlx::query_scalar(query).fetch_one(pool.pool()).await;

    match result {
        Ok(Some(true)) => {
            let query2 =
                "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'commits')";
            let result2: Result<Option<bool>, _> =
                sqlx::query_scalar(query2).fetch_one(pool.pool()).await;
            matches!(result2, Ok(Some(true)))
        }
        Ok(_) => {
            tracing::warn!("PostgreSQL health check: branches table not found");
            false
        }
        Err(e) => {
            tracing::warn!(error = %e, "PostgreSQL health check failed");
            false
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_config_from_env_defaults() {
        let config = PgConfig::default();
        assert_eq!(config.database_url, DEFAULT_DATABASE_URL);
        assert_eq!(config.max_connections, 10);
    }

    #[test]
    fn test_config_from_env_with_values() {
        let _config = PgConfig::from_env();
    }
}
