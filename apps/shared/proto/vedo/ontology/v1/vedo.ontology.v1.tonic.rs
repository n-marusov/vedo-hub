// @generated
/// Generated client implementations.
pub mod ontology_service_client {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    ///
    #[derive(Debug, Clone)]
    pub struct OntologyServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl OntologyServiceClient<tonic::transport::Channel> {
        /// Attempt to create a new client by connecting to a given endpoint.
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> OntologyServiceClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::Body>,
        T::Error: Into<StdError>,
        T::ResponseBody: Body<Data = Bytes> + std::marker::Send + 'static,
        <T::ResponseBody as Body>::Error: Into<StdError> + std::marker::Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_origin(inner: T, origin: Uri) -> Self {
            let inner = tonic::client::Grpc::with_origin(inner, origin);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> OntologyServiceClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T::ResponseBody: Default,
            T: tonic::codegen::Service<
                http::Request<tonic::body::Body>,
                Response = http::Response<
                    <T as tonic::client::GrpcService<tonic::body::Body>>::ResponseBody,
                >,
            >,
            <T as tonic::codegen::Service<
                http::Request<tonic::body::Body>,
            >>::Error: Into<StdError> + std::marker::Send + std::marker::Sync,
        {
            OntologyServiceClient::new(InterceptedService::new(inner, interceptor))
        }
        /// Compress requests with the given encoding.
        ///
        /// This requires the server to support it otherwise it might respond with an
        /// error.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.send_compressed(encoding);
            self
        }
        /// Enable decompressing responses.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.accept_compressed(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_decoding_message_size(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_encoding_message_size(limit);
            self
        }
        /** Class operations
*/
        pub async fn create_class(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateClassRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateClassResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/CreateClass",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "CreateClass"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_class(
            &mut self,
            request: impl tonic::IntoRequest<super::GetClassRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetClassResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/GetClass",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("vedo.ontology.v1.OntologyService", "GetClass"));
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn list_classes(
            &mut self,
            request: impl tonic::IntoRequest<super::ListClassesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListClassesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/ListClasses",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "ListClasses"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn update_class(
            &mut self,
            request: impl tonic::IntoRequest<super::UpdateClassRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateClassResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/UpdateClass",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "UpdateClass"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn delete_class(
            &mut self,
            request: impl tonic::IntoRequest<super::DeleteClassRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteClassResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/DeleteClass",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "DeleteClass"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Property operations
*/
        pub async fn create_property(
            &mut self,
            request: impl tonic::IntoRequest<super::CreatePropertyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreatePropertyResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/CreateProperty",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "CreateProperty"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_property(
            &mut self,
            request: impl tonic::IntoRequest<super::GetPropertyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetPropertyResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/GetProperty",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "GetProperty"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn list_properties(
            &mut self,
            request: impl tonic::IntoRequest<super::ListPropertiesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListPropertiesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/ListProperties",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "ListProperties"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn update_property(
            &mut self,
            request: impl tonic::IntoRequest<super::UpdatePropertyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdatePropertyResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/UpdateProperty",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "UpdateProperty"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn delete_property(
            &mut self,
            request: impl tonic::IntoRequest<super::DeletePropertyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeletePropertyResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/DeleteProperty",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "DeleteProperty"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Individual operations
*/
        pub async fn create_individual(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateIndividualRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateIndividualResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/CreateIndividual",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ontology.v1.OntologyService",
                        "CreateIndividual",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_individual(
            &mut self,
            request: impl tonic::IntoRequest<super::GetIndividualRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetIndividualResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/GetIndividual",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "GetIndividual"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn list_individuals(
            &mut self,
            request: impl tonic::IntoRequest<super::ListIndividualsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListIndividualsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/ListIndividuals",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ontology.v1.OntologyService",
                        "ListIndividuals",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn update_individual(
            &mut self,
            request: impl tonic::IntoRequest<super::UpdateIndividualRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateIndividualResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/UpdateIndividual",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ontology.v1.OntologyService",
                        "UpdateIndividual",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn delete_individual(
            &mut self,
            request: impl tonic::IntoRequest<super::DeleteIndividualRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteIndividualResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/DeleteIndividual",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ontology.v1.OntologyService",
                        "DeleteIndividual",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Ontology metadata
*/
        pub async fn create_ontology(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateOntologyResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/CreateOntology",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "CreateOntology"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_ontology(
            &mut self,
            request: impl tonic::IntoRequest<super::GetOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetOntologyResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/GetOntology",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "GetOntology"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn list_ontologies(
            &mut self,
            request: impl tonic::IntoRequest<super::ListOntologiesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListOntologiesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/ListOntologies",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "ListOntologies"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn update_ontology(
            &mut self,
            request: impl tonic::IntoRequest<super::UpdateOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateOntologyResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/UpdateOntology",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "UpdateOntology"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn delete_ontology(
            &mut self,
            request: impl tonic::IntoRequest<super::DeleteOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteOntologyResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/DeleteOntology",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "DeleteOntology"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Atomic batch operations
*/
        pub async fn apply_sequence(
            &mut self,
            request: impl tonic::IntoRequest<super::ApplySequenceRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ApplySequenceResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/ApplySequence",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "ApplySequence"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Query
*/
        pub async fn execute_sparql(
            &mut self,
            request: impl tonic::IntoRequest<super::ExecuteSparqlRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ExecuteSparqlResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/ExecuteSPARQL",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "ExecuteSPARQL"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn execute_cypher(
            &mut self,
            request: impl tonic::IntoRequest<super::ExecuteCypherRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ExecuteCypherResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/ExecuteCYPHER",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "ExecuteCYPHER"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** GraphQL
*/
        pub async fn execute_graph_ql(
            &mut self,
            request: impl tonic::IntoRequest<super::ExecuteGraphQlRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ExecuteGraphQlResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::unknown(
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/vedo.ontology.v1.OntologyService/ExecuteGraphQL",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.ontology.v1.OntologyService", "ExecuteGraphQL"),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod ontology_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with OntologyServiceServer.
    #[async_trait]
    pub trait OntologyService: std::marker::Send + std::marker::Sync + 'static {
        /** Class operations
*/
        async fn create_class(
            &self,
            request: tonic::Request<super::CreateClassRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateClassResponse>,
            tonic::Status,
        >;
        ///
        async fn get_class(
            &self,
            request: tonic::Request<super::GetClassRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetClassResponse>,
            tonic::Status,
        >;
        ///
        async fn list_classes(
            &self,
            request: tonic::Request<super::ListClassesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListClassesResponse>,
            tonic::Status,
        >;
        ///
        async fn update_class(
            &self,
            request: tonic::Request<super::UpdateClassRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateClassResponse>,
            tonic::Status,
        >;
        ///
        async fn delete_class(
            &self,
            request: tonic::Request<super::DeleteClassRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteClassResponse>,
            tonic::Status,
        >;
        /** Property operations
*/
        async fn create_property(
            &self,
            request: tonic::Request<super::CreatePropertyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreatePropertyResponse>,
            tonic::Status,
        >;
        ///
        async fn get_property(
            &self,
            request: tonic::Request<super::GetPropertyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetPropertyResponse>,
            tonic::Status,
        >;
        ///
        async fn list_properties(
            &self,
            request: tonic::Request<super::ListPropertiesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListPropertiesResponse>,
            tonic::Status,
        >;
        ///
        async fn update_property(
            &self,
            request: tonic::Request<super::UpdatePropertyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdatePropertyResponse>,
            tonic::Status,
        >;
        ///
        async fn delete_property(
            &self,
            request: tonic::Request<super::DeletePropertyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeletePropertyResponse>,
            tonic::Status,
        >;
        /** Individual operations
*/
        async fn create_individual(
            &self,
            request: tonic::Request<super::CreateIndividualRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateIndividualResponse>,
            tonic::Status,
        >;
        ///
        async fn get_individual(
            &self,
            request: tonic::Request<super::GetIndividualRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetIndividualResponse>,
            tonic::Status,
        >;
        ///
        async fn list_individuals(
            &self,
            request: tonic::Request<super::ListIndividualsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListIndividualsResponse>,
            tonic::Status,
        >;
        ///
        async fn update_individual(
            &self,
            request: tonic::Request<super::UpdateIndividualRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateIndividualResponse>,
            tonic::Status,
        >;
        ///
        async fn delete_individual(
            &self,
            request: tonic::Request<super::DeleteIndividualRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteIndividualResponse>,
            tonic::Status,
        >;
        /** Ontology metadata
*/
        async fn create_ontology(
            &self,
            request: tonic::Request<super::CreateOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateOntologyResponse>,
            tonic::Status,
        >;
        ///
        async fn get_ontology(
            &self,
            request: tonic::Request<super::GetOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetOntologyResponse>,
            tonic::Status,
        >;
        ///
        async fn list_ontologies(
            &self,
            request: tonic::Request<super::ListOntologiesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListOntologiesResponse>,
            tonic::Status,
        >;
        ///
        async fn update_ontology(
            &self,
            request: tonic::Request<super::UpdateOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::UpdateOntologyResponse>,
            tonic::Status,
        >;
        ///
        async fn delete_ontology(
            &self,
            request: tonic::Request<super::DeleteOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteOntologyResponse>,
            tonic::Status,
        >;
        /** Atomic batch operations
*/
        async fn apply_sequence(
            &self,
            request: tonic::Request<super::ApplySequenceRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ApplySequenceResponse>,
            tonic::Status,
        >;
        /** Query
*/
        async fn execute_sparql(
            &self,
            request: tonic::Request<super::ExecuteSparqlRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ExecuteSparqlResponse>,
            tonic::Status,
        >;
        ///
        async fn execute_cypher(
            &self,
            request: tonic::Request<super::ExecuteCypherRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ExecuteCypherResponse>,
            tonic::Status,
        >;
        /** GraphQL
*/
        async fn execute_graph_ql(
            &self,
            request: tonic::Request<super::ExecuteGraphQlRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ExecuteGraphQlResponse>,
            tonic::Status,
        >;
    }
    ///
    #[derive(Debug)]
    pub struct OntologyServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> OntologyServiceServer<T> {
        pub fn new(inner: T) -> Self {
            Self::from_arc(Arc::new(inner))
        }
        pub fn from_arc(inner: Arc<T>) -> Self {
            Self {
                inner,
                accept_compression_encodings: Default::default(),
                send_compression_encodings: Default::default(),
                max_decoding_message_size: None,
                max_encoding_message_size: None,
            }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> InterceptedService<Self, F>
        where
            F: tonic::service::Interceptor,
        {
            InterceptedService::new(Self::new(inner), interceptor)
        }
        /// Enable decompressing requests with the given encoding.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.accept_compression_encodings.enable(encoding);
            self
        }
        /// Compress responses with the given encoding, if the client supports it.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.send_compression_encodings.enable(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.max_decoding_message_size = Some(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.max_encoding_message_size = Some(limit);
            self
        }
    }
    impl<T, B> tonic::codegen::Service<http::Request<B>> for OntologyServiceServer<T>
    where
        T: OntologyService,
        B: Body + std::marker::Send + 'static,
        B::Error: Into<StdError> + std::marker::Send + 'static,
    {
        type Response = http::Response<tonic::body::Body>;
        type Error = std::convert::Infallible;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(
            &mut self,
            _cx: &mut Context<'_>,
        ) -> Poll<std::result::Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            match req.uri().path() {
                "/vedo.ontology.v1.OntologyService/CreateClass" => {
                    #[allow(non_camel_case_types)]
                    struct CreateClassSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::CreateClassRequest>
                    for CreateClassSvc<T> {
                        type Response = super::CreateClassResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateClassRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::create_class(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = CreateClassSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/GetClass" => {
                    #[allow(non_camel_case_types)]
                    struct GetClassSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::GetClassRequest>
                    for GetClassSvc<T> {
                        type Response = super::GetClassResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetClassRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::get_class(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = GetClassSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/ListClasses" => {
                    #[allow(non_camel_case_types)]
                    struct ListClassesSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::ListClassesRequest>
                    for ListClassesSvc<T> {
                        type Response = super::ListClassesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListClassesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::list_classes(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ListClassesSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/UpdateClass" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateClassSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::UpdateClassRequest>
                    for UpdateClassSvc<T> {
                        type Response = super::UpdateClassResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::UpdateClassRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::update_class(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateClassSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/DeleteClass" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteClassSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::DeleteClassRequest>
                    for DeleteClassSvc<T> {
                        type Response = super::DeleteClassResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeleteClassRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::delete_class(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = DeleteClassSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/CreateProperty" => {
                    #[allow(non_camel_case_types)]
                    struct CreatePropertySvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::CreatePropertyRequest>
                    for CreatePropertySvc<T> {
                        type Response = super::CreatePropertyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreatePropertyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::create_property(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = CreatePropertySvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/GetProperty" => {
                    #[allow(non_camel_case_types)]
                    struct GetPropertySvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::GetPropertyRequest>
                    for GetPropertySvc<T> {
                        type Response = super::GetPropertyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetPropertyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::get_property(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = GetPropertySvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/ListProperties" => {
                    #[allow(non_camel_case_types)]
                    struct ListPropertiesSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::ListPropertiesRequest>
                    for ListPropertiesSvc<T> {
                        type Response = super::ListPropertiesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListPropertiesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::list_properties(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ListPropertiesSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/UpdateProperty" => {
                    #[allow(non_camel_case_types)]
                    struct UpdatePropertySvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::UpdatePropertyRequest>
                    for UpdatePropertySvc<T> {
                        type Response = super::UpdatePropertyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::UpdatePropertyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::update_property(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdatePropertySvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/DeleteProperty" => {
                    #[allow(non_camel_case_types)]
                    struct DeletePropertySvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::DeletePropertyRequest>
                    for DeletePropertySvc<T> {
                        type Response = super::DeletePropertyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeletePropertyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::delete_property(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = DeletePropertySvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/CreateIndividual" => {
                    #[allow(non_camel_case_types)]
                    struct CreateIndividualSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::CreateIndividualRequest>
                    for CreateIndividualSvc<T> {
                        type Response = super::CreateIndividualResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateIndividualRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::create_individual(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = CreateIndividualSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/GetIndividual" => {
                    #[allow(non_camel_case_types)]
                    struct GetIndividualSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::GetIndividualRequest>
                    for GetIndividualSvc<T> {
                        type Response = super::GetIndividualResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetIndividualRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::get_individual(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = GetIndividualSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/ListIndividuals" => {
                    #[allow(non_camel_case_types)]
                    struct ListIndividualsSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::ListIndividualsRequest>
                    for ListIndividualsSvc<T> {
                        type Response = super::ListIndividualsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListIndividualsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::list_individuals(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ListIndividualsSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/UpdateIndividual" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateIndividualSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::UpdateIndividualRequest>
                    for UpdateIndividualSvc<T> {
                        type Response = super::UpdateIndividualResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::UpdateIndividualRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::update_individual(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateIndividualSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/DeleteIndividual" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteIndividualSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::DeleteIndividualRequest>
                    for DeleteIndividualSvc<T> {
                        type Response = super::DeleteIndividualResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeleteIndividualRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::delete_individual(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = DeleteIndividualSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/CreateOntology" => {
                    #[allow(non_camel_case_types)]
                    struct CreateOntologySvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::CreateOntologyRequest>
                    for CreateOntologySvc<T> {
                        type Response = super::CreateOntologyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateOntologyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::create_ontology(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = CreateOntologySvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/GetOntology" => {
                    #[allow(non_camel_case_types)]
                    struct GetOntologySvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::GetOntologyRequest>
                    for GetOntologySvc<T> {
                        type Response = super::GetOntologyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetOntologyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::get_ontology(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = GetOntologySvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/ListOntologies" => {
                    #[allow(non_camel_case_types)]
                    struct ListOntologiesSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::ListOntologiesRequest>
                    for ListOntologiesSvc<T> {
                        type Response = super::ListOntologiesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListOntologiesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::list_ontologies(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ListOntologiesSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/UpdateOntology" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateOntologySvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::UpdateOntologyRequest>
                    for UpdateOntologySvc<T> {
                        type Response = super::UpdateOntologyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::UpdateOntologyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::update_ontology(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateOntologySvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/DeleteOntology" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteOntologySvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::DeleteOntologyRequest>
                    for DeleteOntologySvc<T> {
                        type Response = super::DeleteOntologyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeleteOntologyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::delete_ontology(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = DeleteOntologySvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/ApplySequence" => {
                    #[allow(non_camel_case_types)]
                    struct ApplySequenceSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::ApplySequenceRequest>
                    for ApplySequenceSvc<T> {
                        type Response = super::ApplySequenceResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ApplySequenceRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::apply_sequence(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ApplySequenceSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/ExecuteSPARQL" => {
                    #[allow(non_camel_case_types)]
                    struct ExecuteSPARQLSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::ExecuteSparqlRequest>
                    for ExecuteSPARQLSvc<T> {
                        type Response = super::ExecuteSparqlResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ExecuteSparqlRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::execute_sparql(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ExecuteSPARQLSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/ExecuteCYPHER" => {
                    #[allow(non_camel_case_types)]
                    struct ExecuteCYPHERSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::ExecuteCypherRequest>
                    for ExecuteCYPHERSvc<T> {
                        type Response = super::ExecuteCypherResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ExecuteCypherRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::execute_cypher(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ExecuteCYPHERSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ontology.v1.OntologyService/ExecuteGraphQL" => {
                    #[allow(non_camel_case_types)]
                    struct ExecuteGraphQLSvc<T: OntologyService>(pub Arc<T>);
                    impl<
                        T: OntologyService,
                    > tonic::server::UnaryService<super::ExecuteGraphQlRequest>
                    for ExecuteGraphQLSvc<T> {
                        type Response = super::ExecuteGraphQlResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ExecuteGraphQlRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as OntologyService>::execute_graph_ql(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ExecuteGraphQLSvc(inner);
                        let codec = tonic_prost::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                _ => {
                    Box::pin(async move {
                        let mut response = http::Response::new(
                            tonic::body::Body::default(),
                        );
                        let headers = response.headers_mut();
                        headers
                            .insert(
                                tonic::Status::GRPC_STATUS,
                                (tonic::Code::Unimplemented as i32).into(),
                            );
                        headers
                            .insert(
                                http::header::CONTENT_TYPE,
                                tonic::metadata::GRPC_CONTENT_TYPE,
                            );
                        Ok(response)
                    })
                }
            }
        }
    }
    impl<T> Clone for OntologyServiceServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self {
                inner,
                accept_compression_encodings: self.accept_compression_encodings,
                send_compression_encodings: self.send_compression_encodings,
                max_decoding_message_size: self.max_decoding_message_size,
                max_encoding_message_size: self.max_encoding_message_size,
            }
        }
    }
    /// Generated gRPC service name
    pub const SERVICE_NAME: &str = "vedo.ontology.v1.OntologyService";
    impl<T> tonic::server::NamedService for OntologyServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
