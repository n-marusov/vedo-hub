//! Merge handler integration tests (require PostgreSQL).
//!
//! Verifies the F3.828 merge-blocking contract: a merge whose branches
//! conflict on the same triple MUST be rejected with HTTP 409
//! VER-MERGE-CONFLICT instead of silently auto-resolving.

mod common;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use tower::ServiceExt;
use uuid::Uuid;

fn req(method: Method, uri: &str, body: Option<&str>) -> Request<Body> {
    let mut b = Request::builder().method(method).uri(uri);
    if let Some(body) = body {
        b = b.header("Content-Type", "application/json");
        b.body(Body::from(body.to_string())).unwrap()
    } else {
        b.body(Body::empty()).unwrap()
    }
}

/// Creates a branch via the API and returns its ID.
async fn create_branch(app: &axum::Router, ontology_id: &Uuid, name: &str) -> Uuid {
    let body = serde_json::json!({
        "name": name,
        "ontology_id": ontology_id.to_string(),
    });
    let resp = app
        .clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/branches",
            Some(&body.to_string()),
        ))
        .await
        .unwrap();
    assert!(
        resp.status().is_success(),
        "create branch should succeed, got {}",
        resp.status()
    );
    let body_bytes = axum::body::to_bytes(resp.into_body(), usize::MAX)
        .await
        .unwrap();
    let json: serde_json::Value = serde_json::from_slice(&body_bytes).unwrap();
    let id_str = json["id"].as_str().expect("branch response must have 'id'");
    Uuid::parse_str(id_str).expect("branch id must be a valid UUID")
}

/// Creates a commit with a custom delta. Returns the response status.
async fn create_commit_with_delta(
    app: &axum::Router,
    branch_id: &Uuid,
    message: &str,
    delta: serde_json::Value,
) -> StatusCode {
    let body = serde_json::json!({
        "branch_id": branch_id.to_string(),
        "message": message,
        "author_id": "merge-test-author",
        "author_name": "Merge Test Author",
        "delta": delta,
    });
    app.clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/commits",
            Some(&body.to_string()),
        ))
        .await
        .unwrap()
        .status()
}

/// Performs a merge via the API. Returns the response.
async fn merge_branches(
    app: &axum::Router,
    source: &Uuid,
    target: &Uuid,
) -> axum::response::Response {
    let body = serde_json::json!({
        "source_branch_id": source.to_string(),
        "target_branch_id": target.to_string(),
        "message": "merge",
        "author_id": "merge-test-author",
        "author_name": "Merge Test Author",
    });
    app.clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/branches/merge",
            Some(&body.to_string()),
        ))
        .await
        .unwrap()
}

/// Verifies that a merge with conflicting changes on the same triple is
/// blocked with 409 VER-MERGE-CONFLICT (vision F3.828).
///
/// Setup: two branches commit DIFFERENT values for the same triple
/// (both-modified-same-triple is a deterministic conflict in
/// compute_merged_delta). The merge must NOT create a merge commit and must
/// return 409.
#[tokio::test]
async fn test_merge_conflicting_branches_returns_409() {
    common::require_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool.clone());

    let ontology_id = Uuid::new_v4();
    let main_branch = create_branch(&app, &ontology_id, "main").await;

    // Feature branch diverges from main.
    let feat_branch = create_branch(&app, &ontology_id, "feat/conflict").await;

    // Both branches modify the SAME triple with different values.
    let modified_main = serde_json::json!({
        "modified_triples": [
            {"s": "ClassA", "p": "rdfs:label", "old_o": "Class A", "new_o": "Class A (main)"}
        ]
    });
    let modified_feat = serde_json::json!({
        "modified_triples": [
            {"s": "ClassA", "p": "rdfs:label", "old_o": "Class A", "new_o": "Class A (feat)"}
        ]
    });

    let status_main =
        create_commit_with_delta(&app, &main_branch, "main edits label", modified_main).await;
    assert_eq!(status_main, StatusCode::CREATED);
    let status_feat =
        create_commit_with_delta(&app, &feat_branch, "feat edits label", modified_feat).await;
    assert_eq!(status_feat, StatusCode::CREATED);

    // Merge feat into main — must be blocked with 409.
    let resp = merge_branches(&app, &feat_branch, &main_branch).await;
    assert_eq!(
        resp.status(),
        StatusCode::CONFLICT,
        "conflicting merge must be blocked with 409"
    );

    let body_bytes = axum::body::to_bytes(resp.into_body(), usize::MAX)
        .await
        .unwrap();
    let json: serde_json::Value = serde_json::from_slice(&body_bytes).unwrap();
    assert_eq!(json["error"], "VER-MERGE-CONFLICT");

    // No merge commit must have been created on main (blocked before write).
    let merge_count: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM commits WHERE branch_id = $1 AND message = 'merge'",
    )
    .bind(main_branch)
    .fetch_one(pool.pool())
    .await
    .unwrap_or(-1);
    assert_eq!(merge_count, 0, "blocked merge must not create a commit");
}

/// Non-conflicting merge (disjoint triples) must still succeed with 200.
#[tokio::test]
async fn test_merge_non_conflicting_branches_succeeds() {
    common::require_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool.clone());

    let ontology_id = Uuid::new_v4();
    let main_branch = create_branch(&app, &ontology_id, "main").await;
    let feat_branch = create_branch(&app, &ontology_id, "feat/no-conflict").await;

    // Disjoint triples — no conflict.
    let added_main = serde_json::json!({
        "added_triples": [
            {"s": "ClassMain", "p": "rdfs:label", "o": "Main class"}
        ]
    });
    let added_feat = serde_json::json!({
        "added_triples": [
            {"s": "ClassFeat", "p": "rdfs:label", "o": "Feat class"}
        ]
    });

    let s1 = create_commit_with_delta(&app, &main_branch, "main adds class", added_main).await;
    assert_eq!(s1, StatusCode::CREATED);
    let s2 = create_commit_with_delta(&app, &feat_branch, "feat adds class", added_feat).await;
    assert_eq!(s2, StatusCode::CREATED);

    let resp = merge_branches(&app, &feat_branch, &main_branch).await;
    assert_eq!(
        resp.status(),
        StatusCode::OK,
        "non-conflicting merge should succeed"
    );
}
