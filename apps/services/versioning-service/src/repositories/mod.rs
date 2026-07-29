//! Repository module for versioning service — data access layer.

pub mod branch_repo;
pub mod commit_repo;

pub use branch_repo::*;
pub use commit_repo::*;

// Re-export traits for use in handlers and tests
pub use branch_repo::BranchRepositoryTrait;
pub use commit_repo::CommitRepositoryTrait;
