use tonic::{Request, Response, Status};
use tracing::info;
use vedo_shared::protos::publisher::publisher_service_server::PublisherService;
use vedo_shared::protos::publisher;

/// Stub gRPC server for publisher-service.
/// Full implementation will be added in a later phase.
pub struct PublisherGrpcServer;

#[tonic::async_trait]
impl PublisherService for PublisherGrpcServer {
    async fn publish_snapshot(&self, _req: Request<publisher::PublishSnapshotRequest>) -> Result<Response<publisher::PublishSnapshotResponse>, Status> {
        info!("gRPC PublishSnapshot called (stub)");
        Err(Status::unimplemented("publisher-service not yet fully implemented"))
    }
    async fn get_published(&self, _req: Request<publisher::GetPublishedRequest>) -> Result<Response<publisher::GetPublishedResponse>, Status> {
        info!("gRPC GetPublished called (stub)");
        Err(Status::unimplemented("publisher-service not yet fully implemented"))
    }
    async fn list_published(&self, _req: Request<publisher::ListPublishedRequest>) -> Result<Response<publisher::ListPublishedResponse>, Status> {
        info!("gRPC ListPublished called (stub)");
        Err(Status::unimplemented("publisher-service not yet fully implemented"))
    }
    async fn unpublish_snapshot(&self, _req: Request<publisher::UnpublishSnapshotRequest>) -> Result<Response<publisher::UnpublishSnapshotResponse>, Status> {
        info!("gRPC UnpublishSnapshot called (stub)");
        Err(Status::unimplemented("publisher-service not yet fully implemented"))
    }
}
