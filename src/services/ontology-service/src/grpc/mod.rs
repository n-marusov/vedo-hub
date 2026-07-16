//! gRPC server implementation for the ontology service.
//!
//! Implements the tonic `OntologyService` trait generated from the proto
//! definitions in `vedo-shared`. All methods are currently stubs returning
//! `unimplemented` — domain logic will be wired in later phases.

use std::sync::Arc;

use tonic::{Request, Response, Status};
use tracing::info;

use crate::AppState;
use vedo_shared::protos::ontology::ontology_service_server::OntologyService;
use vedo_shared::protos::ontology::*;

/// gRPC server that implements the `OntologyService` tonic trait.
///
/// Holds a reference to the shared application state (Neo4j pool) so that
/// future implementations can delegate to domain-layer functions.
pub struct OntologyGrpcServer {
    pub state: Arc<AppState>,
}

#[tonic::async_trait]
impl OntologyService for OntologyGrpcServer {
    // ---------------------------------------------------------------------------
    // Class operations
    // ---------------------------------------------------------------------------

    async fn create_class(
        &self,
        _request: Request<CreateClassRequest>,
    ) -> Result<Response<CreateClassResponse>, Status> {
        info!("gRPC CreateClass called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn get_class(
        &self,
        _request: Request<GetClassRequest>,
    ) -> Result<Response<GetClassResponse>, Status> {
        info!("gRPC GetClass called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn list_classes(
        &self,
        _request: Request<ListClassesRequest>,
    ) -> Result<Response<ListClassesResponse>, Status> {
        info!("gRPC ListClasses called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn update_class(
        &self,
        _request: Request<UpdateClassRequest>,
    ) -> Result<Response<UpdateClassResponse>, Status> {
        info!("gRPC UpdateClass called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn delete_class(
        &self,
        _request: Request<DeleteClassRequest>,
    ) -> Result<Response<DeleteClassResponse>, Status> {
        info!("gRPC DeleteClass called");
        Err(Status::unimplemented("not yet implemented"))
    }

    // ---------------------------------------------------------------------------
    // Property operations
    // ---------------------------------------------------------------------------

    async fn create_property(
        &self,
        _request: Request<CreatePropertyRequest>,
    ) -> Result<Response<CreatePropertyResponse>, Status> {
        info!("gRPC CreateProperty called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn get_property(
        &self,
        _request: Request<GetPropertyRequest>,
    ) -> Result<Response<GetPropertyResponse>, Status> {
        info!("gRPC GetProperty called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn list_properties(
        &self,
        _request: Request<ListPropertiesRequest>,
    ) -> Result<Response<ListPropertiesResponse>, Status> {
        info!("gRPC ListProperties called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn update_property(
        &self,
        _request: Request<UpdatePropertyRequest>,
    ) -> Result<Response<UpdatePropertyResponse>, Status> {
        info!("gRPC UpdateProperty called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn delete_property(
        &self,
        _request: Request<DeletePropertyRequest>,
    ) -> Result<Response<DeletePropertyResponse>, Status> {
        info!("gRPC DeleteProperty called");
        Err(Status::unimplemented("not yet implemented"))
    }

    // ---------------------------------------------------------------------------
    // Individual operations
    // ---------------------------------------------------------------------------

    async fn create_individual(
        &self,
        _request: Request<CreateIndividualRequest>,
    ) -> Result<Response<CreateIndividualResponse>, Status> {
        info!("gRPC CreateIndividual called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn get_individual(
        &self,
        _request: Request<GetIndividualRequest>,
    ) -> Result<Response<GetIndividualResponse>, Status> {
        info!("gRPC GetIndividual called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn list_individuals(
        &self,
        _request: Request<ListIndividualsRequest>,
    ) -> Result<Response<ListIndividualsResponse>, Status> {
        info!("gRPC ListIndividuals called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn update_individual(
        &self,
        _request: Request<UpdateIndividualRequest>,
    ) -> Result<Response<UpdateIndividualResponse>, Status> {
        info!("gRPC UpdateIndividual called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn delete_individual(
        &self,
        _request: Request<DeleteIndividualRequest>,
    ) -> Result<Response<DeleteIndividualResponse>, Status> {
        info!("gRPC DeleteIndividual called");
        Err(Status::unimplemented("not yet implemented"))
    }

    // ---------------------------------------------------------------------------
    // Ontology metadata operations
    // ---------------------------------------------------------------------------

    async fn create_ontology(
        &self,
        _request: Request<CreateOntologyRequest>,
    ) -> Result<Response<CreateOntologyResponse>, Status> {
        info!("gRPC CreateOntology called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn get_ontology(
        &self,
        _request: Request<GetOntologyRequest>,
    ) -> Result<Response<GetOntologyResponse>, Status> {
        info!("gRPC GetOntology called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn list_ontologies(
        &self,
        _request: Request<ListOntologiesRequest>,
    ) -> Result<Response<ListOntologiesResponse>, Status> {
        info!("gRPC ListOntologies called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn update_ontology(
        &self,
        _request: Request<UpdateOntologyRequest>,
    ) -> Result<Response<UpdateOntologyResponse>, Status> {
        info!("gRPC UpdateOntology called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn delete_ontology(
        &self,
        _request: Request<DeleteOntologyRequest>,
    ) -> Result<Response<DeleteOntologyResponse>, Status> {
        info!("gRPC DeleteOntology called");
        Err(Status::unimplemented("not yet implemented"))
    }

    // ---------------------------------------------------------------------------
    // ApplySequence — atomic batch operations
    // ---------------------------------------------------------------------------

    async fn apply_sequence(
        &self,
        _request: Request<ApplySequenceRequest>,
    ) -> Result<Response<ApplySequenceResponse>, Status> {
        info!("gRPC ApplySequence called");
        Err(Status::unimplemented("not yet implemented"))
    }

    // ---------------------------------------------------------------------------
    // Query operations
    // ---------------------------------------------------------------------------

    async fn execute_sparql(
        &self,
        _request: Request<ExecuteSPARQLRequest>,
    ) -> Result<Response<ExecuteSPARQLResponse>, Status> {
        info!("gRPC ExecuteSPARQL called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn execute_cypher(
        &self,
        _request: Request<ExecuteCYPHERRequest>,
    ) -> Result<Response<ExecuteCYPHERResponse>, Status> {
        info!("gRPC ExecuteCYPHER called");
        Err(Status::unimplemented("not yet implemented"))
    }

    async fn execute_graphql(
        &self,
        _request: Request<ExecuteGraphQLRequest>,
    ) -> Result<Response<ExecuteGraphQLResponse>, Status> {
        info!("gRPC ExecuteGraphQL called");
        Err(Status::unimplemented("not yet implemented"))
    }
}
