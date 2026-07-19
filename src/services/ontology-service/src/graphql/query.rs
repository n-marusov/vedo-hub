//! GraphQL query resolvers that delegate to repository methods.
//!
//! Each resolver extracts the Neo4j pool from the axum application state,
//! creates the appropriate repository, and converts domain types to
//! GraphQL output types.

use async_graphql::{Context, Object, Result};
use std::sync::Arc;
use uuid::Uuid;

use crate::classes::{self, ClassRepository, ListClassesParams, OwlClassSummary};
use crate::individuals::{IndividualRepository, IndividualSummary, ListIndividualsParams};
use crate::properties::{ListPropertiesParams, PropertyRepository, PropertySummary};
use crate::AppState;

use super::types::*;
use super::versioning_client::VersioningClient;

// ── helpers ────────────────────────────────────────────────────────────────────────

/// Extracts the Neo4j pool from the async-graphql context.
fn pool_from_ctx(ctx: &Context<'_>) -> Result<crate::neo4j::Neo4jPool> {
    let state = ctx
        .data::<Arc<AppState>>()
        .map_err(|e| async_graphql::Error::new(format!("Failed to access app state: {e:?}")))?;
    state
        .neo4j
        .clone()
        .ok_or_else(|| async_graphql::Error::new("Neo4j not configured"))
}

fn map_error<E: std::fmt::Display>(e: E) -> async_graphql::Error {
    async_graphql::Error::new(format!("{e}"))
}

/// Converts an `OwlClass` into a `GqlClass`.
fn gql_class_from(cls: classes::OwlClass) -> GqlClass {
    GqlClass {
        id: cls.id,
        label: cls.label,
        comment: cls.comment,
        parents: cls.parents,
        children: cls.children,
    }
}

/// Converts an `OwlClassSummary` into a `GqlClassSummary`.
fn gql_class_summary_from(s: OwlClassSummary) -> GqlClassSummary {
    GqlClassSummary {
        id: s.id,
        label: s.label,
        comment: s.comment,
        parents: s.parents,
    }
}

/// Converts a `PropertySummary` into a `GqlPropertySummary`.
fn gql_property_summary_from(s: PropertySummary) -> GqlPropertySummary {
    GqlPropertySummary {
        id: s.id,
        label: s.label,
        property_type: s.property_type.into(),
        xsd_type: s.xsd_type,
        domains: s.domains,
    }
}

/// Converts an `IndividualSummary` into a `GqlIndividual`.
fn gql_individual_from_summary(ind: IndividualSummary) -> GqlIndividual {
    GqlIndividual {
        id: ind.id,
        label: ind.label,
        comment: ind.comment,
        class_id: ind.class_id,
        class_label: ind.class_label,
        literal_values: vec![],
        reference_values: vec![],
    }
}

/// Converts an `Individual` (full) into a `GqlIndividual`.
fn gql_individual_from_full(ind: crate::individuals::Individual) -> GqlIndividual {
    GqlIndividual {
        id: ind.id,
        label: ind.label,
        comment: ind.comment,
        class_id: ind.class_id,
        class_label: ind.class_label,
        literal_values: ind
            .literal_values
            .into_iter()
            .map(|lv| GqlLiteralValue {
                property_id: lv.property_id,
                property_label: lv.property_label,
                value: lv.value,
                xsd_type: lv.xsd_type,
                value_id: lv.value_id,
            })
            .collect(),
        reference_values: ind
            .reference_values
            .into_iter()
            .map(|rv| GqlReferenceValue {
                property_id: rv.property_id,
                property_label: rv.property_label,
                target_id: rv.target_id,
                target_label: rv.target_label,
                edge_id: rv.edge_id,
            })
            .collect(),
    }
}

/// Converts a `Property` (full) into a `GqlProperty`.
fn gql_property_from_full(prop: crate::properties::Property) -> GqlProperty {
    GqlProperty {
        id: prop.id,
        label: prop.label,
        comment: prop.comment,
        property_type: prop.property_type.into(),
        domains: prop.domains,
        ranges: prop.ranges,
        xsd_type: prop.xsd_type,
        characteristics: prop.characteristics.into(),
        annotations: prop.annotations.into_iter().map(Into::into).collect(),
    }
}

// ── Query Root ─────────────────────────────────────────────────────────────────────

#[derive(Default)]
pub struct QueryRoot;

#[Object]
impl QueryRoot {
    /// Returns ontology metadata (branch, commit, dirty state).
    ///
    /// Pulls the default branch (`main` when present) and its head commit from
    /// the versioning-service. When the versioning-service is unreachable the
    /// resolver falls back to a safe default so the workspace still loads.
    async fn ontology(&self, _ctx: &Context<'_>, id: String) -> Result<GqlOntology> {
        let client = VersioningClient::from_env();
        let ontology_uuid = Uuid::parse_str(&id).unwrap_or_else(|_| Uuid::nil());

        let (branch, commit) = match client.list_branches(ontology_uuid, None).await {
            Ok(page) => {
                let main = page
                    .items
                    .iter()
                    .find(|b| b.name == "main")
                    .or_else(|| page.items.first());
                match main {
                    Some(b) => (
                        b.name.clone(),
                        b.head_commit_id.map(|c| c.to_string()).unwrap_or_default(),
                    ),
                    None => (String::from("main"), String::new()),
                }
            }
            Err(e) => {
                tracing::warn!(
                    ontology_id = %id,
                    error = %e,
                    "versioning-service unreachable; falling back to default branch"
                );
                (String::from("main"), String::new())
            }
        };

        Ok(GqlOntology {
            id,
            name: String::new(),
            branch,
            commit,
            dirty: false,
        })
    }

    // ── Class Queries ──────────────────────────────────────────────────────────

    /// Retrieve a single class by ID.
    async fn class(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Class ID")] class_id: String,
    ) -> Result<Option<GqlClass>> {
        let pool = pool_from_ctx(ctx)?;
        let repo = ClassRepository::new(pool);
        match repo.get(&ontology_id, &class_id).await {
            Ok(cls) => Ok(Some(gql_class_from(cls))),
            Err(classes::ClassError::NotFound(_)) => Ok(None),
            Err(e) => Err(map_error(e)),
        }
    }

    /// List classes with optional search and pagination.
    async fn classes(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Search filter (matches label)", default)] q: Option<String>,
        #[graphql(default = 0)] page: u64,
        #[graphql(default = 20)] per_page: u64,
    ) -> Result<GqlClassConnection> {
        let pool = pool_from_ctx(ctx)?;
        let repo = ClassRepository::new(pool);
        let params = ListClassesParams {
            q: q.unwrap_or_default(),
            page,
            per_page,
        };
        let result = repo.list(&ontology_id, &params).await.map_err(map_error)?;
        Ok(GqlClassConnection {
            items: result
                .items
                .into_iter()
                .map(gql_class_summary_from)
                .collect(),
            total: result.total,
            page: result.page,
            per_page: result.per_page,
        })
    }

    /// Returns the class hierarchy tree (root classes with lazy children).
    async fn class_tree(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
    ) -> Result<Vec<GqlClassTreeNode>> {
        let pool = pool_from_ctx(ctx)?;
        let repo = ClassRepository::new(pool);
        let params = ListClassesParams {
            q: String::new(),
            page: 0,
            per_page: 500,
        };
        let result = repo
            .get_root_classes(&ontology_id, &params)
            .await
            .map_err(map_error)?;
        Ok(result
            .items
            .into_iter()
            .map(|s| GqlClassTreeNode {
                id: s.id,
                label: s.label,
                children: vec![],
            })
            .collect())
    }

    /// Returns the ancestor chain (breadcrumb) for a class.
    async fn class_ancestors(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Class ID")] class_id: String,
    ) -> Result<Vec<GqlBreadcrumbItem>> {
        let pool = pool_from_ctx(ctx)?;
        let repo = ClassRepository::new(pool);
        let items = repo
            .get_breadcrumb(&ontology_id, &class_id)
            .await
            .map_err(map_error)?;
        Ok(items
            .into_iter()
            .map(|b| GqlBreadcrumbItem {
                id: b.id,
                label: b.label,
            })
            .collect())
    }

    /// Returns the descendant tree for a class.
    async fn class_descendants(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Class ID")] class_id: String,
        #[graphql(default = 10)] max_depth: u64,
    ) -> Result<Vec<GqlClassTreeNode>> {
        let pool = pool_from_ctx(ctx)?;
        let repo = ClassRepository::new(pool);
        let tree = repo
            .get_descendants_tree(&ontology_id, &class_id, max_depth)
            .await
            .map_err(map_error)?;
        // Convert ClassTreeNode → GqlClassTreeNode recursively
        fn convert(node: classes::ClassTreeNode) -> GqlClassTreeNode {
            GqlClassTreeNode {
                id: node.id,
                label: node.label,
                children: node.children.into_iter().map(convert).collect(),
            }
        }
        Ok(tree.into_iter().map(convert).collect())
    }

    /// Returns the graph neighborhood for a class (connected nodes and edges).
    async fn graph_neighborhood(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Class ID")] class_id: String,
        #[graphql(default = 2)] depth: u64,
    ) -> Result<GqlGraphNeighborhood> {
        let pool = pool_from_ctx(ctx)?;
        let repo = ClassRepository::new(pool);
        let nh = repo
            .get_graph_neighborhood(&ontology_id, &class_id, depth)
            .await
            .map_err(map_error)?;
        Ok(GqlGraphNeighborhood {
            nodes: nh
                .nodes
                .into_iter()
                .map(|n| GqlGraphNode {
                    id: n.id,
                    label: n.label,
                })
                .collect(),
            edges: nh
                .edges
                .into_iter()
                .map(|e| GqlGraphEdge {
                    source_id: e.source_id,
                    target_id: e.target_id,
                    property_id: e.property_id,
                    property_label: e.property_label,
                })
                .collect(),
        })
    }

    /// Autocomplete search for classes by label.
    async fn autocomplete_classes(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Search query")] q: String,
        #[graphql(default = 20)] limit: u64,
    ) -> Result<Vec<GqlClassSummary>> {
        let pool = pool_from_ctx(ctx)?;
        let repo = ClassRepository::new(pool);
        let results = repo
            .autocomplete_search(&ontology_id, &q, std::cmp::min(limit, 100))
            .await
            .map_err(map_error)?;
        Ok(results.into_iter().map(gql_class_summary_from).collect())
    }

    // ── Property Queries ───────────────────────────────────────────────────────

    /// Retrieve a single property by ID.
    async fn property(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Property ID")] property_id: String,
    ) -> Result<Option<GqlProperty>> {
        let pool = pool_from_ctx(ctx)?;
        let repo = PropertyRepository::new(pool);
        match repo.get(&ontology_id, &property_id).await {
            Ok(prop) => Ok(Some(gql_property_from_full(prop))),
            Err(crate::properties::PropertyError::NotFound(_)) => Ok(None),
            Err(e) => Err(map_error(e)),
        }
    }

    /// List properties with optional type filter and pagination.
    async fn properties(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Search filter", default)] q: Option<String>,
        #[graphql(desc = "Property type filter")] property_type: Option<GqlPropertyType>,
        #[graphql(default = 0)] page: u64,
        #[graphql(default = 20)] per_page: u64,
    ) -> Result<GqlPropertyConnection> {
        let pool = pool_from_ctx(ctx)?;
        let repo = PropertyRepository::new(pool);
        let type_str = property_type.map(|t| match t {
            GqlPropertyType::Object => "object".to_string(),
            GqlPropertyType::Datatype => "datatype".to_string(),
        });
        let params = ListPropertiesParams {
            q: q.unwrap_or_default(),
            property_type: type_str,
            page,
            per_page,
        };
        let result = repo.list(&ontology_id, &params).await.map_err(map_error)?;
        Ok(GqlPropertyConnection {
            items: result
                .items
                .into_iter()
                .map(gql_property_summary_from)
                .collect(),
            total: result.total,
            page: result.page,
            per_page: result.per_page,
        })
    }

    // ── Individual Queries ──────────────────────────────────────────────────────

    /// Retrieve a single individual by ID (full detail with property values).
    async fn individual(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Individual ID")] individual_id: String,
    ) -> Result<Option<GqlIndividual>> {
        let pool = pool_from_ctx(ctx)?;
        let repo = IndividualRepository::new(pool);
        match repo.get(&ontology_id, &individual_id).await {
            Ok(ind) => Ok(Some(gql_individual_from_full(ind))),
            Err(crate::individuals::IndividualError::NotFound(_)) => Ok(None),
            Err(e) => Err(map_error(e)),
        }
    }

    /// List individuals filtered by class with pagination.
    async fn individuals(
        &self,
        ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Filter by class ID")] class_id: String,
        #[graphql(desc = "Text search", default)] q: Option<String>,
        #[graphql(default = 0)] page: u64,
        #[graphql(default = 20)] per_page: u64,
    ) -> Result<GqlIndividualConnection> {
        let pool = pool_from_ctx(ctx)?;
        let repo = IndividualRepository::new(pool);
        let params = ListIndividualsParams {
            class_id,
            q: q.unwrap_or_default(),
            property_filter: String::new(),
            page,
            per_page,
        };
        let result = repo.list(&ontology_id, &params).await.map_err(map_error)?;
        Ok(GqlIndividualConnection {
            items: result
                .items
                .into_iter()
                .map(gql_individual_from_summary)
                .collect(),
            total: result.total,
            page: result.page,
            per_page: result.per_page,
        })
    }

    // ── Versioning Queries (Commit / Branch) ────────────────────────────────────
    //
    // These resolvers proxy HTTP calls to the versioning-service. The frontend
    // Apollo Client sees them as native GraphQL fields; the gateway still routes
    // write operations (create commit, merge branch, ...) through the REST
    // endpoints exposed under `/api/v1/versioning/*`.

    /// List commit history, optionally filtered by `branch_id`.
    async fn commits(
        &self,
        _ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Filter by branch ID", default)] branch_id: Option<String>,
        #[graphql(default = 0)] page: u64,
        #[graphql(default = 20)] per_page: u64,
    ) -> Result<GqlCommitConnection> {
        let client = VersioningClient::from_env();
        let _ = Uuid::parse_str(&ontology_id)
            .map_err(|e| async_graphql::Error::new(format!("invalid ontology_id: {e}")))?;
        let branch = match branch_id {
            Some(b) => Some(
                Uuid::parse_str(&b)
                    .map_err(|e| async_graphql::Error::new(format!("invalid branch_id: {e}")))?,
            ),
            None => None,
        };

        let page_result = client
            .list_commits(branch, page, per_page)
            .await
            .map_err(map_error)?;

        Ok(GqlCommitConnection {
            items: page_result
                .items
                .into_iter()
                .map(|c| GqlCommit {
                    id: c.id.to_string(),
                    branch_id: c.branch_id.to_string(),
                    parent_commit_id: c.parent_commit_id.map(|p| p.to_string()),
                    message: c.message,
                    author_id: c.author_id,
                    author_name: c.author_name,
                    total_changes: c.total_changes as i64,
                    created_at: c.created_at.to_rfc3339(),
                })
                .collect(),
            total: page_result.total as i64,
            page: page_result.page as i64,
            per_page: page_result.per_page as i64,
        })
    }

    /// Returns a single branch by ID.
    async fn branch(
        &self,
        _ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Branch ID")] branch_id: String,
    ) -> Result<Option<GqlBranch>> {
        let client = VersioningClient::from_env();
        let ontology_uuid = Uuid::parse_str(&ontology_id)
            .map_err(|e| async_graphql::Error::new(format!("invalid ontology_id: {e}")))?;
        let _ = Uuid::parse_str(&branch_id)
            .map_err(|e| async_graphql::Error::new(format!("invalid branch_id: {e}")))?;

        let branch_page = client
            .list_branches(ontology_uuid, None)
            .await
            .map_err(map_error)?;

        Ok(branch_page
            .items
            .into_iter()
            .find(|b| b.id.to_string() == branch_id)
            .map(branch_into_gql))
    }

    /// Lists branches for an ontology, optionally relative to a reference branch.
    async fn branches(
        &self,
        _ctx: &Context<'_>,
        #[graphql(desc = "Ontology ID")] ontology_id: String,
        #[graphql(desc = "Reference branch for ahead/behind", default)] reference_branch_id: Option<
            String,
        >,
    ) -> Result<GqlBranchConnection> {
        let client = VersioningClient::from_env();
        let ontology_uuid = Uuid::parse_str(&ontology_id)
            .map_err(|e| async_graphql::Error::new(format!("invalid ontology_id: {e}")))?;
        let reference = match reference_branch_id {
            Some(r) => Some(Uuid::parse_str(&r).map_err(|e| {
                async_graphql::Error::new(format!("invalid reference_branch_id: {e}"))
            })?),
            None => None,
        };

        let branch_page = client
            .list_branches(ontology_uuid, reference)
            .await
            .map_err(map_error)?;

        Ok(GqlBranchConnection {
            items: branch_page.items.into_iter().map(branch_into_gql).collect(),
            total: branch_page.total as i64,
        })
    }

    // ── Organization model resolvers (groups / projects / members) -----

    /// Returns all groups from the auth-service.
    async fn groups(&self, ctx: &Context<'_>, _q: Option<String>) -> Result<Vec<GqlGroup>> {
        use crate::AppState;
        let state = ctx
            .data::<Arc<AppState>>()
            .map_err(|e| async_graphql::Error::new(format!("Failed to access app state: {e:?}")))?;
        let groups = state.auth_client.list_groups().await.map_err(map_error)?;
        Ok(groups
            .into_iter()
            .map(|g| GqlGroup {
                id: g.id,
                name: g.name,
                description: g.description,
                parent_group_id: g.parent_id,
                visibility: g.visibility,
                member_count: None,
                project_count: None,
            })
            .collect())
    }

    /// Returns all projects from the auth-service.
    async fn projects(
        &self,
        ctx: &Context<'_>,
        _q: Option<String>,
        _sort_by: Option<String>,
        _sort_dir: Option<String>,
        page: Option<i32>,
        per_page: Option<i32>,
    ) -> Result<Vec<GqlProject>> {
        use crate::AppState;
        let state = ctx
            .data::<Arc<AppState>>()
            .map_err(|e| async_graphql::Error::new(format!("Failed to access app state: {e:?}")))?;
        let projects = state
            .auth_client
            .list_projects(page, per_page)
            .await
            .map_err(map_error)?;
        Ok(projects
            .into_iter()
            .map(|p| GqlProject {
                id: p.id,
                name: p.name,
                description: p.description,
                visibility: p.visibility,
                member_count: None,
                updated_at: None,
            })
            .collect())
    }

    /// Returns members for a given ontology scope.
    async fn members(&self, ctx: &Context<'_>, ontology_id: String) -> Result<Vec<GqlMember>> {
        use crate::AppState;
        let state = ctx
            .data::<Arc<AppState>>()
            .map_err(|e| async_graphql::Error::new(format!("Failed to access app state: {e:?}")))?;
        let scope = format!("ontology/{}", ontology_id);
        let members = state
            .auth_client
            .list_members(&scope)
            .await
            .map_err(map_error)?;
        Ok(members
            .into_iter()
            .map(|m| GqlMember {
                user_id: m.user_id,
                scope: m.scope,
                role: m.role,
                username: None,
                avatar_url: None,
                added_at: None,
            })
            .collect())
    }
}

// ── versioning mapping helpers ────────────────────────────────

pub(super) fn branch_into_gql(b: super::versioning_client::RemoteBranch) -> GqlBranch {
    GqlBranch {
        id: b.id.to_string(),
        name: b.name,
        ontology_id: b.ontology_id.to_string(),
        head_commit_id: b.head_commit_id.map(|c| c.to_string()),
        created_at: b.created_at.to_rfc3339(),
        is_protected: b.is_protected,
        last_commit_message: b.last_commit_message,
        last_commit_author: b.last_commit_author,
        ahead_count: b.ahead_count,
        behind_count: b.behind_count,
    }
}
