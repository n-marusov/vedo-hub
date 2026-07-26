// @generated
/// Generated client implementations.
pub mod ai_orchestration_service_client {
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
    pub struct AiOrchestrationServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl AiOrchestrationServiceClient<tonic::transport::Channel> {
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
    impl<T> AiOrchestrationServiceClient<T>
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
        ) -> AiOrchestrationServiceClient<InterceptedService<T, F>>
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
            AiOrchestrationServiceClient::new(
                InterceptedService::new(inner, interceptor),
            )
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
        /** NL→OWL — Convert natural language text to OWL ontology structure (unary)
*/
        pub async fn generate_owl(
            &mut self,
            request: impl tonic::IntoRequest<super::GenerateOwlRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GenerateOwlResponse>,
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/GenerateOWL",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ai_orchestration.v1.AIOrchestrationService",
                        "GenerateOWL",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** NaturalLanguageQuery — Answer NL questions grounded in ontology context (unary)
*/
        pub async fn natural_language_query(
            &mut self,
            request: impl tonic::IntoRequest<super::NaturalLanguageQueryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NaturalLanguageQueryResponse>,
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/NaturalLanguageQuery",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ai_orchestration.v1.AIOrchestrationService",
                        "NaturalLanguageQuery",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** RefineOntology — Iterative refinement with server-streaming updates
*/
        pub async fn refine_ontology(
            &mut self,
            request: impl tonic::IntoRequest<super::RefineOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::RefineOntologyResponse>>,
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/RefineOntology",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ai_orchestration.v1.AIOrchestrationService",
                        "RefineOntology",
                    ),
                );
            self.inner.server_streaming(req, path, codec).await
        }
        /** Complete — AI-assisted completion with server-streaming tokens
*/
        pub async fn complete(
            &mut self,
            request: impl tonic::IntoRequest<super::CompleteRequest>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::CompleteResponse>>,
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/Complete",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ai_orchestration.v1.AIOrchestrationService",
                        "Complete",
                    ),
                );
            self.inner.server_streaming(req, path, codec).await
        }
        /** Policy check — Used by document-extractor (Hybrid Model)
*/
        pub async fn check_policy(
            &mut self,
            request: impl tonic::IntoRequest<super::CheckPolicyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckPolicyResponse>,
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/CheckPolicy",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ai_orchestration.v1.AIOrchestrationService",
                        "CheckPolicy",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** LLM usage audit — Centralized usage logging
*/
        pub async fn log_llm_usage(
            &mut self,
            request: impl tonic::IntoRequest<super::LogLlmUsageRequest>,
        ) -> std::result::Result<
            tonic::Response<super::LogLlmUsageResponse>,
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/LogLLMUsage",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.ai_orchestration.v1.AIOrchestrationService",
                        "LogLLMUsage",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod ai_orchestration_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with AiOrchestrationServiceServer.
    #[async_trait]
    pub trait AiOrchestrationService: std::marker::Send + std::marker::Sync + 'static {
        /** NL→OWL — Convert natural language text to OWL ontology structure (unary)
*/
        async fn generate_owl(
            &self,
            request: tonic::Request<super::GenerateOwlRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GenerateOwlResponse>,
            tonic::Status,
        >;
        /** NaturalLanguageQuery — Answer NL questions grounded in ontology context (unary)
*/
        async fn natural_language_query(
            &self,
            request: tonic::Request<super::NaturalLanguageQueryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::NaturalLanguageQueryResponse>,
            tonic::Status,
        >;
        /// Server streaming response type for the RefineOntology method.
        type RefineOntologyStream: tonic::codegen::tokio_stream::Stream<
                Item = std::result::Result<super::RefineOntologyResponse, tonic::Status>,
            >
            + std::marker::Send
            + 'static;
        /** RefineOntology — Iterative refinement with server-streaming updates
*/
        async fn refine_ontology(
            &self,
            request: tonic::Request<super::RefineOntologyRequest>,
        ) -> std::result::Result<
            tonic::Response<Self::RefineOntologyStream>,
            tonic::Status,
        >;
        /// Server streaming response type for the Complete method.
        type CompleteStream: tonic::codegen::tokio_stream::Stream<
                Item = std::result::Result<super::CompleteResponse, tonic::Status>,
            >
            + std::marker::Send
            + 'static;
        /** Complete — AI-assisted completion with server-streaming tokens
*/
        async fn complete(
            &self,
            request: tonic::Request<super::CompleteRequest>,
        ) -> std::result::Result<tonic::Response<Self::CompleteStream>, tonic::Status>;
        /** Policy check — Used by document-extractor (Hybrid Model)
*/
        async fn check_policy(
            &self,
            request: tonic::Request<super::CheckPolicyRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckPolicyResponse>,
            tonic::Status,
        >;
        /** LLM usage audit — Centralized usage logging
*/
        async fn log_llm_usage(
            &self,
            request: tonic::Request<super::LogLlmUsageRequest>,
        ) -> std::result::Result<
            tonic::Response<super::LogLlmUsageResponse>,
            tonic::Status,
        >;
    }
    ///
    #[derive(Debug)]
    pub struct AiOrchestrationServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> AiOrchestrationServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>>
    for AiOrchestrationServiceServer<T>
    where
        T: AiOrchestrationService,
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/GenerateOWL" => {
                    #[allow(non_camel_case_types)]
                    struct GenerateOWLSvc<T: AiOrchestrationService>(pub Arc<T>);
                    impl<
                        T: AiOrchestrationService,
                    > tonic::server::UnaryService<super::GenerateOwlRequest>
                    for GenerateOWLSvc<T> {
                        type Response = super::GenerateOwlResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GenerateOwlRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AiOrchestrationService>::generate_owl(&inner, request)
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
                        let method = GenerateOWLSvc(inner);
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/NaturalLanguageQuery" => {
                    #[allow(non_camel_case_types)]
                    struct NaturalLanguageQuerySvc<T: AiOrchestrationService>(
                        pub Arc<T>,
                    );
                    impl<
                        T: AiOrchestrationService,
                    > tonic::server::UnaryService<super::NaturalLanguageQueryRequest>
                    for NaturalLanguageQuerySvc<T> {
                        type Response = super::NaturalLanguageQueryResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::NaturalLanguageQueryRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AiOrchestrationService>::natural_language_query(
                                        &inner,
                                        request,
                                    )
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
                        let method = NaturalLanguageQuerySvc(inner);
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/RefineOntology" => {
                    #[allow(non_camel_case_types)]
                    struct RefineOntologySvc<T: AiOrchestrationService>(pub Arc<T>);
                    impl<
                        T: AiOrchestrationService,
                    > tonic::server::ServerStreamingService<super::RefineOntologyRequest>
                    for RefineOntologySvc<T> {
                        type Response = super::RefineOntologyResponse;
                        type ResponseStream = T::RefineOntologyStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RefineOntologyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AiOrchestrationService>::refine_ontology(
                                        &inner,
                                        request,
                                    )
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
                        let method = RefineOntologySvc(inner);
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
                        let res = grpc.server_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ai_orchestration.v1.AIOrchestrationService/Complete" => {
                    #[allow(non_camel_case_types)]
                    struct CompleteSvc<T: AiOrchestrationService>(pub Arc<T>);
                    impl<
                        T: AiOrchestrationService,
                    > tonic::server::ServerStreamingService<super::CompleteRequest>
                    for CompleteSvc<T> {
                        type Response = super::CompleteResponse;
                        type ResponseStream = T::CompleteStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CompleteRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AiOrchestrationService>::complete(&inner, request)
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
                        let method = CompleteSvc(inner);
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
                        let res = grpc.server_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/vedo.ai_orchestration.v1.AIOrchestrationService/CheckPolicy" => {
                    #[allow(non_camel_case_types)]
                    struct CheckPolicySvc<T: AiOrchestrationService>(pub Arc<T>);
                    impl<
                        T: AiOrchestrationService,
                    > tonic::server::UnaryService<super::CheckPolicyRequest>
                    for CheckPolicySvc<T> {
                        type Response = super::CheckPolicyResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CheckPolicyRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AiOrchestrationService>::check_policy(&inner, request)
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
                        let method = CheckPolicySvc(inner);
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
                "/vedo.ai_orchestration.v1.AIOrchestrationService/LogLLMUsage" => {
                    #[allow(non_camel_case_types)]
                    struct LogLLMUsageSvc<T: AiOrchestrationService>(pub Arc<T>);
                    impl<
                        T: AiOrchestrationService,
                    > tonic::server::UnaryService<super::LogLlmUsageRequest>
                    for LogLLMUsageSvc<T> {
                        type Response = super::LogLlmUsageResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::LogLlmUsageRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as AiOrchestrationService>::log_llm_usage(
                                        &inner,
                                        request,
                                    )
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
                        let method = LogLLMUsageSvc(inner);
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
    impl<T> Clone for AiOrchestrationServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "vedo.ai_orchestration.v1.AIOrchestrationService";
    impl<T> tonic::server::NamedService for AiOrchestrationServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
