use tonic::{Request, Response, Status};
use tracing::info;
use vedo_shared::protos::public_browse::v1 as public_browse;
use vedo_shared::protos::public_browse::v1::public_browse_service_server::PublicBrowseService;

/// Stub gRPC server for public-browse-api.
/// Full implementation will be added in a later phase.
pub struct PublicBrowseGrpcServer;

#[tonic::async_trait]
impl PublicBrowseService for PublicBrowseGrpcServer {
    async fn browse_ontology(
        &self,
        _req: Request<public_browse::BrowseOntologyRequest>,
    ) -> Result<Response<public_browse::BrowseOntologyResponse>, Status> {
        info!("gRPC BrowseOntology called (stub)");
        Err(Status::unimplemented(
            "public-browse-api not yet fully implemented",
        ))
    }
    async fn search(
        &self,
        _req: Request<public_browse::SearchRequest>,
    ) -> Result<Response<public_browse::SearchResponse>, Status> {
        info!("gRPC Search called (stub)");
        Err(Status::unimplemented(
            "public-browse-api not yet fully implemented",
        ))
    }
    async fn get_class_hierarchy(
        &self,
        _req: Request<public_browse::GetClassHierarchyRequest>,
    ) -> Result<Response<public_browse::GetClassHierarchyResponse>, Status> {
        info!("gRPC GetClassHierarchy called (stub)");
        Err(Status::unimplemented(
            "public-browse-api not yet fully implemented",
        ))
    }
}
