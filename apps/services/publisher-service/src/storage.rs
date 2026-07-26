use std::collections::HashMap;
use std::sync::atomic::{AtomicI64, Ordering};
use std::sync::Arc;

use chrono::Utc;
use tokio::sync::RwLock;
use tracing::info;
use uuid::Uuid;

use crate::models::{Snapshot, SnapshotStatus};

/// `SnapshotStore` manages snapshot metadata and payload storage.
///
/// For MVP, metadata is kept in-memory. A future iteration will add a
/// PostgreSQL/JSONB backend and migrate payloads to `MinIO` (S3-compatible).
#[derive(Clone)]
pub struct SnapshotStore {
    /// In-memory snapshot metadata store (keyed by snapshot ID).
    snapshots: Arc<RwLock<HashMap<String, Snapshot>>>,
    /// Total bytes stored across all snapshots.
    total_bytes: Arc<AtomicI64>,
}

impl SnapshotStore {
    /// Creates a new empty `SnapshotStore`.
    pub fn new() -> Self {
        SnapshotStore {
            snapshots: Arc::new(RwLock::new(HashMap::new())),
            total_bytes: Arc::new(AtomicI64::new(0)),
        }
    }

    /// Creates a snapshot record and stores the payload bytes.
    /// Returns the created Snapshot.
    pub async fn create_snapshot(
        &self,
        ontology_id: &str,
        branch_id: Option<&str>,
        commit_id: Option<&str>,
        payload_bytes: &[u8],
        format: &str,
    ) -> Snapshot {
        let id = format!("snap_{}", Uuid::new_v4());
        let now = Utc::now();
        let size: i64 = payload_bytes.len().try_into().unwrap_or(i64::MAX);

        let storage_path = format!(
            "snapshots/{}/{}_{}.{}",
            ontology_id,
            id,
            now.format("%Y%m%d_%H%M%S"),
            format.replace('/', "-"),
        );

        let snapshot = Snapshot {
            id: id.clone(),
            ontology_id: ontology_id.to_string(),
            branch_id: branch_id.map(ToString::to_string),
            commit_id: commit_id.map(ToString::to_string),
            status: SnapshotStatus::Published,
            created_at: now,
            size_bytes: size,
            storage_path: storage_path.clone(),
            format: format.to_string(),
        };

        info!(
            snapshot_id = %id,
            ontology_id = %ontology_id,
            size_bytes = size,
            storage_path = %storage_path,
            format = %format,
            "Snapshot created"
        );

        self.total_bytes.fetch_add(size, Ordering::Relaxed);
        self.snapshots
            .write()
            .await
            .insert(id.clone(), snapshot.clone());

        snapshot
    }

    /// Retrieves a snapshot by ID.
    pub async fn get_snapshot(&self, id: &str) -> Option<Snapshot> {
        self.snapshots.read().await.get(id).cloned()
    }

    /// Lists all snapshots for a given ontology, ordered by creation time descending.
    pub async fn list_snapshots(&self, ontology_id: &str) -> Vec<Snapshot> {
        let guard = self.snapshots.read().await;
        let mut snapshots: Vec<Snapshot> = guard
            .values()
            .filter(|s| s.ontology_id == ontology_id)
            .cloned()
            .collect();
        snapshots.sort_by_key(|b| std::cmp::Reverse(b.created_at));
        snapshots
    }

    /// Lists all snapshots across all ontologies.
    pub async fn list_all_snapshots(&self) -> Vec<Snapshot> {
        let mut snapshots: Vec<Snapshot> = self.snapshots.read().await.values().cloned().collect();
        snapshots.sort_by_key(|b| std::cmp::Reverse(b.created_at));
        snapshots
    }

    /// Retires (soft-deletes) a snapshot by marking its status as Retired.
    pub async fn retire_snapshot(&self, id: &str) -> bool {
        let mut guard = self.snapshots.write().await;
        if let Some(snapshot) = guard.get_mut(id) {
            if snapshot.status == SnapshotStatus::Retired {
                return false; // already retired
            }
            snapshot.status = SnapshotStatus::Retired;
            info!(snapshot_id = %id, "Snapshot retired");
            true
        } else {
            false
        }
    }

    /// Returns the total number of snapshots.
    pub async fn count(&self) -> usize {
        self.snapshots.read().await.len()
    }

    /// Returns the total bytes stored.
    pub fn total_storage_bytes(&self) -> i64 {
        self.total_bytes.load(Ordering::Relaxed)
    }
}

impl Default for SnapshotStore {
    fn default() -> Self {
        Self::new()
    }
}
