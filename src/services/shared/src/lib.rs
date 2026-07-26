pub mod health;
pub mod metrics;
pub mod tls;
pub mod tracing;

pub use health::*;
pub use metrics::*;
pub use tracing::*;

/// Protocol buffer generated types and gRPC service traits.
///
/// All `include_proto!` calls are placed in a single module so that
/// cross-package `super::super` references (e.g.
/// `super::super::common::v1::ErrorDetail` from within `ontology::v1`)
/// resolve correctly through the shared `protos` parent module.
pub mod protos {
    pub mod common {
        pub mod v1 {
            tonic::include_proto!("vedo.common.v1");
        }
    }
    pub mod ontology {
        pub mod v1 {
            tonic::include_proto!("vedo.ontology.v1");
        }
    }
    pub mod versioning {
        pub mod v1 {
            tonic::include_proto!("vedo.versioning.v1");
        }
    }
    pub mod auth {
        pub mod v1 {
            tonic::include_proto!("vedo.auth.v1");
        }
    }
    pub mod commenting {
        pub mod v1 {
            tonic::include_proto!("vedo.commenting.v1");
        }
    }
    pub mod publisher {
        pub mod v1 {
            tonic::include_proto!("vedo.publisher.v1");
        }
    }
    pub mod public_browse {
        pub mod v1 {
            tonic::include_proto!("vedo.public_browse.v1");
        }
    }
}
