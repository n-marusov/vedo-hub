// @generated
/// Generated client implementations.
pub mod versioning_service_client {
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
    pub struct VersioningServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl VersioningServiceClient<tonic::transport::Channel> {
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
    impl<T> VersioningServiceClient<T>
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
        ) -> VersioningServiceClient<InterceptedService<T, F>>
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
            VersioningServiceClient::new(InterceptedService::new(inner, interceptor))
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
        /** Commits
*/
        pub async fn create_commit(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateCommitRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateCommitResponse>,
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
                "/vedo.versioning.v1.VersioningService/CreateCommit",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.versioning.v1.VersioningService",
                        "CreateCommit",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_commit(
            &mut self,
            request: impl tonic::IntoRequest<super::GetCommitRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommitResponse>,
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
                "/vedo.versioning.v1.VersioningService/GetCommit",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.versioning.v1.VersioningService", "GetCommit"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn list_commits(
            &mut self,
            request: impl tonic::IntoRequest<super::ListCommitsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListCommitsResponse>,
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
                "/vedo.versioning.v1.VersioningService/ListCommits",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.versioning.v1.VersioningService",
                        "ListCommits",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Branches
*/
        pub async fn create_branch(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateBranchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateBranchResponse>,
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
                "/vedo.versioning.v1.VersioningService/CreateBranch",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.versioning.v1.VersioningService",
                        "CreateBranch",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_branch(
            &mut self,
            request: impl tonic::IntoRequest<super::GetBranchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetBranchResponse>,
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
                "/vedo.versioning.v1.VersioningService/GetBranch",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.versioning.v1.VersioningService", "GetBranch"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn list_branches(
            &mut self,
            request: impl tonic::IntoRequest<super::ListBranchesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListBranchesResponse>,
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
                "/vedo.versioning.v1.VersioningService/ListBranches",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.versioning.v1.VersioningService",
                        "ListBranches",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn delete_branch(
            &mut self,
            request: impl tonic::IntoRequest<super::DeleteBranchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteBranchResponse>,
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
                "/vedo.versioning.v1.VersioningService/DeleteBranch",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.versioning.v1.VersioningService",
                        "DeleteBranch",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn switch_branch(
            &mut self,
            request: impl tonic::IntoRequest<super::SwitchBranchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SwitchBranchResponse>,
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
                "/vedo.versioning.v1.VersioningService/SwitchBranch",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.versioning.v1.VersioningService",
                        "SwitchBranch",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Diff
*/
        pub async fn get_diff(
            &mut self,
            request: impl tonic::IntoRequest<super::GetDiffRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetDiffResponse>,
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
                "/vedo.versioning.v1.VersioningService/GetDiff",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.versioning.v1.VersioningService", "GetDiff"),
                );
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn get_commit_delta(
            &mut self,
            request: impl tonic::IntoRequest<super::GetCommitDeltaRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommitDeltaResponse>,
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
                "/vedo.versioning.v1.VersioningService/GetCommitDelta",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "vedo.versioning.v1.VersioningService",
                        "GetCommitDelta",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Merge
*/
        pub async fn merge(
            &mut self,
            request: impl tonic::IntoRequest<super::MergeRequest>,
        ) -> std::result::Result<tonic::Response<super::MergeResponse>, tonic::Status> {
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
                "/vedo.versioning.v1.VersioningService/Merge",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.versioning.v1.VersioningService", "Merge"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Rollback
*/
        pub async fn rollback(
            &mut self,
            request: impl tonic::IntoRequest<super::RollbackRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RollbackResponse>,
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
                "/vedo.versioning.v1.VersioningService/Rollback",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.versioning.v1.VersioningService", "Rollback"),
                );
            self.inner.unary(req, path, codec).await
        }
        /** Checkout
*/
        pub async fn checkout(
            &mut self,
            request: impl tonic::IntoRequest<super::CheckoutRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckoutResponse>,
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
                "/vedo.versioning.v1.VersioningService/Checkout",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("vedo.versioning.v1.VersioningService", "Checkout"),
                );
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod versioning_service_server {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value,
    )]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with VersioningServiceServer.
    #[async_trait]
    pub trait VersioningService: std::marker::Send + std::marker::Sync + 'static {
        /** Commits
*/
        async fn create_commit(
            &self,
            request: tonic::Request<super::CreateCommitRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateCommitResponse>,
            tonic::Status,
        >;
        ///
        async fn get_commit(
            &self,
            request: tonic::Request<super::GetCommitRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommitResponse>,
            tonic::Status,
        >;
        ///
        async fn list_commits(
            &self,
            request: tonic::Request<super::ListCommitsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListCommitsResponse>,
            tonic::Status,
        >;
        /** Branches
*/
        async fn create_branch(
            &self,
            request: tonic::Request<super::CreateBranchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CreateBranchResponse>,
            tonic::Status,
        >;
        ///
        async fn get_branch(
            &self,
            request: tonic::Request<super::GetBranchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetBranchResponse>,
            tonic::Status,
        >;
        ///
        async fn list_branches(
            &self,
            request: tonic::Request<super::ListBranchesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ListBranchesResponse>,
            tonic::Status,
        >;
        ///
        async fn delete_branch(
            &self,
            request: tonic::Request<super::DeleteBranchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::DeleteBranchResponse>,
            tonic::Status,
        >;
        ///
        async fn switch_branch(
            &self,
            request: tonic::Request<super::SwitchBranchRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SwitchBranchResponse>,
            tonic::Status,
        >;
        /** Diff
*/
        async fn get_diff(
            &self,
            request: tonic::Request<super::GetDiffRequest>,
        ) -> std::result::Result<tonic::Response<super::GetDiffResponse>, tonic::Status>;
        ///
        async fn get_commit_delta(
            &self,
            request: tonic::Request<super::GetCommitDeltaRequest>,
        ) -> std::result::Result<
            tonic::Response<super::GetCommitDeltaResponse>,
            tonic::Status,
        >;
        /** Merge
*/
        async fn merge(
            &self,
            request: tonic::Request<super::MergeRequest>,
        ) -> std::result::Result<tonic::Response<super::MergeResponse>, tonic::Status>;
        /** Rollback
*/
        async fn rollback(
            &self,
            request: tonic::Request<super::RollbackRequest>,
        ) -> std::result::Result<
            tonic::Response<super::RollbackResponse>,
            tonic::Status,
        >;
        /** Checkout
*/
        async fn checkout(
            &self,
            request: tonic::Request<super::CheckoutRequest>,
        ) -> std::result::Result<
            tonic::Response<super::CheckoutResponse>,
            tonic::Status,
        >;
    }
    ///
    #[derive(Debug)]
    pub struct VersioningServiceServer<T> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T> VersioningServiceServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for VersioningServiceServer<T>
    where
        T: VersioningService,
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
                "/vedo.versioning.v1.VersioningService/CreateCommit" => {
                    #[allow(non_camel_case_types)]
                    struct CreateCommitSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::CreateCommitRequest>
                    for CreateCommitSvc<T> {
                        type Response = super::CreateCommitResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateCommitRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::create_commit(&inner, request)
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
                        let method = CreateCommitSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/GetCommit" => {
                    #[allow(non_camel_case_types)]
                    struct GetCommitSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::GetCommitRequest>
                    for GetCommitSvc<T> {
                        type Response = super::GetCommitResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetCommitRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::get_commit(&inner, request).await
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
                        let method = GetCommitSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/ListCommits" => {
                    #[allow(non_camel_case_types)]
                    struct ListCommitsSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::ListCommitsRequest>
                    for ListCommitsSvc<T> {
                        type Response = super::ListCommitsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListCommitsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::list_commits(&inner, request)
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
                        let method = ListCommitsSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/CreateBranch" => {
                    #[allow(non_camel_case_types)]
                    struct CreateBranchSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::CreateBranchRequest>
                    for CreateBranchSvc<T> {
                        type Response = super::CreateBranchResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateBranchRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::create_branch(&inner, request)
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
                        let method = CreateBranchSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/GetBranch" => {
                    #[allow(non_camel_case_types)]
                    struct GetBranchSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::GetBranchRequest>
                    for GetBranchSvc<T> {
                        type Response = super::GetBranchResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetBranchRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::get_branch(&inner, request).await
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
                        let method = GetBranchSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/ListBranches" => {
                    #[allow(non_camel_case_types)]
                    struct ListBranchesSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::ListBranchesRequest>
                    for ListBranchesSvc<T> {
                        type Response = super::ListBranchesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListBranchesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::list_branches(&inner, request)
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
                        let method = ListBranchesSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/DeleteBranch" => {
                    #[allow(non_camel_case_types)]
                    struct DeleteBranchSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::DeleteBranchRequest>
                    for DeleteBranchSvc<T> {
                        type Response = super::DeleteBranchResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DeleteBranchRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::delete_branch(&inner, request)
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
                        let method = DeleteBranchSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/SwitchBranch" => {
                    #[allow(non_camel_case_types)]
                    struct SwitchBranchSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::SwitchBranchRequest>
                    for SwitchBranchSvc<T> {
                        type Response = super::SwitchBranchResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::SwitchBranchRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::switch_branch(&inner, request)
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
                        let method = SwitchBranchSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/GetDiff" => {
                    #[allow(non_camel_case_types)]
                    struct GetDiffSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::GetDiffRequest>
                    for GetDiffSvc<T> {
                        type Response = super::GetDiffResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetDiffRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::get_diff(&inner, request).await
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
                        let method = GetDiffSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/GetCommitDelta" => {
                    #[allow(non_camel_case_types)]
                    struct GetCommitDeltaSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::GetCommitDeltaRequest>
                    for GetCommitDeltaSvc<T> {
                        type Response = super::GetCommitDeltaResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetCommitDeltaRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::get_commit_delta(&inner, request)
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
                        let method = GetCommitDeltaSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/Merge" => {
                    #[allow(non_camel_case_types)]
                    struct MergeSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::MergeRequest> for MergeSvc<T> {
                        type Response = super::MergeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MergeRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::merge(&inner, request).await
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
                        let method = MergeSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/Rollback" => {
                    #[allow(non_camel_case_types)]
                    struct RollbackSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::RollbackRequest>
                    for RollbackSvc<T> {
                        type Response = super::RollbackResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::RollbackRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::rollback(&inner, request).await
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
                        let method = RollbackSvc(inner);
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
                "/vedo.versioning.v1.VersioningService/Checkout" => {
                    #[allow(non_camel_case_types)]
                    struct CheckoutSvc<T: VersioningService>(pub Arc<T>);
                    impl<
                        T: VersioningService,
                    > tonic::server::UnaryService<super::CheckoutRequest>
                    for CheckoutSvc<T> {
                        type Response = super::CheckoutResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CheckoutRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as VersioningService>::checkout(&inner, request).await
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
                        let method = CheckoutSvc(inner);
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
    impl<T> Clone for VersioningServiceServer<T> {
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
    pub const SERVICE_NAME: &str = "vedo.versioning.v1.VersioningService";
    impl<T> tonic::server::NamedService for VersioningServiceServer<T> {
        const NAME: &'static str = SERVICE_NAME;
    }
}
