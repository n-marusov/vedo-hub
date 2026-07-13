//! Service layer for versioning service — business logic and orchestration.

pub mod delta_service;
pub mod merge_service;
pub mod state_service;
pub mod sync_client;

pub use delta_service::*;
pub use merge_service::*;
pub use state_service::*;
pub use sync_client::*;
