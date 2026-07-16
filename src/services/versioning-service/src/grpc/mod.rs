use tonic::{Request, Response, Status};
use tracing::info;
use vedo_shared::protos::versioning;
use vedo_shared::protos::versioning::versioning_service_server::VersioningService;

/// Stub gRPC server for versioning-service.
/// Full implementation will be added in a later phase.
pub struct VersioningGrpcServer;

#[tonic::async_trait]
impl VersioningService for VersioningGrpcServer {
    async fn create_commit(
        &self,
        _req: Request<versioning::CreateCommitRequest>,
    ) -> Result<Response<versioning::CreateCommitResponse>, Status> {
        info!("gRPC CreateCommit called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn get_commit(
        &self,
        _req: Request<versioning::GetCommitRequest>,
    ) -> Result<Response<versioning::GetCommitResponse>, Status> {
        info!("gRPC GetCommit called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn list_commits(
        &self,
        _req: Request<versioning::ListCommitsRequest>,
    ) -> Result<Response<versioning::ListCommitsResponse>, Status> {
        info!("gRPC ListCommits called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn create_branch(
        &self,
        _req: Request<versioning::CreateBranchRequest>,
    ) -> Result<Response<versioning::CreateBranchResponse>, Status> {
        info!("gRPC CreateBranch called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn get_branch(
        &self,
        _req: Request<versioning::GetBranchRequest>,
    ) -> Result<Response<versioning::GetBranchResponse>, Status> {
        info!("gRPC GetBranch called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn list_branches(
        &self,
        _req: Request<versioning::ListBranchesRequest>,
    ) -> Result<Response<versioning::ListBranchesResponse>, Status> {
        info!("gRPC ListBranches called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn delete_branch(
        &self,
        _req: Request<versioning::DeleteBranchRequest>,
    ) -> Result<Response<versioning::DeleteBranchResponse>, Status> {
        info!("gRPC DeleteBranch called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn switch_branch(
        &self,
        _req: Request<versioning::SwitchBranchRequest>,
    ) -> Result<Response<versioning::SwitchBranchResponse>, Status> {
        info!("gRPC SwitchBranch called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn get_diff(
        &self,
        _req: Request<versioning::GetDiffRequest>,
    ) -> Result<Response<versioning::GetDiffResponse>, Status> {
        info!("gRPC GetDiff called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn get_commit_delta(
        &self,
        _req: Request<versioning::GetCommitDeltaRequest>,
    ) -> Result<Response<versioning::GetCommitDeltaResponse>, Status> {
        info!("gRPC GetCommitDelta called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn merge(
        &self,
        _req: Request<versioning::MergeRequest>,
    ) -> Result<Response<versioning::MergeResponse>, Status> {
        info!("gRPC Merge called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn rollback(
        &self,
        _req: Request<versioning::RollbackRequest>,
    ) -> Result<Response<versioning::RollbackResponse>, Status> {
        info!("gRPC Rollback called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
    async fn checkout(
        &self,
        _req: Request<versioning::CheckoutRequest>,
    ) -> Result<Response<versioning::CheckoutResponse>, Status> {
        info!("gRPC Checkout called (stub)");
        Err(Status::unimplemented(
            "versioning-service not yet fully implemented",
        ))
    }
}
