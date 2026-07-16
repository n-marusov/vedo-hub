pub mod health;
pub mod metrics;
pub mod tracing;

pub use health::*;
pub use metrics::*;
pub use tracing::*;

/// Protocol buffer generated types and gRPC service traits.
///
/// All `include_proto!` calls are placed in a single inner module so that
/// cross-package references (e.g. `super::super::common::v1::ErrorDetail`
/// from within `ontology.v1.rs`) resolve correctly — they all share the
/// same `vedo` root module, which `tonic_build` assumes when compiling
/// all protos together.
///
/// Re-exports at the `v1` level provide shorthand:
/// `protos::versioning::CreateCommitRequest` instead of
/// `protos::generated::vedo::versioning::v1::CreateCommitRequest`.
pub mod protos {
    /// All generated proto modules share this single Rust module so that
    /// cross-package `super::super` references resolve correctly.
    pub(crate) mod generated {
        tonic::include_proto!("vedo.common.v1");
        tonic::include_proto!("vedo.ontology.v1");
        tonic::include_proto!("vedo.versioning.v1");
        tonic::include_proto!("vedo.auth.v1");
        tonic::include_proto!("vedo.commenting.v1");
        tonic::include_proto!("vedo.publisher.v1");
        tonic::include_proto!("vedo.public_browse.v1");
    }

    pub use generated::vedo::auth::v1 as auth;
    pub use generated::vedo::commenting::v1 as commenting;
    pub use generated::vedo::common::v1 as common;
    pub use generated::vedo::ontology::v1 as ontology;
    pub use generated::vedo::public_browse::v1 as public_browse;
    pub use generated::vedo::publisher::v1 as publisher;
    pub use generated::vedo::versioning::v1 as versioning;
}
