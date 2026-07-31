//! GraphQL query resolvers that delegate to repository methods.
//!
//! Each resolver extracts the Neo4j pool from the axum application state,
//! creates the appropriate repository, and converts domain types to
//! GraphQL output types.

use async_graphql::{Context, Object, Result};
use std::sync::Arc;

use crate::classes::{self, ClassRepository, ListClassesParams, OwlClassSummary};
use crate::individuals::{IndividualRepository, IndividualSummary, ListIndividualsParams};
use crate::properties::{ListPropertiesParams, PropertyRepository, PropertySummary};
use crate::AppState;

#[allow(clippy::wildcard_imports)]
use super::types::*;

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
/// Converts an `OwlClass` into a `GqlClass`.
fn gql_class_from(cls: classes::OwlClass) -> GqlClass {
    GqlClass {
        id: cls.id,
        label: cls.label,
        comment: cls.comment,
        entity_type: GqlEntityType::Class,
        parents: cls.parents,
        children: cls.children,
        // Flags are not persisted in the domain model yet (follow-up: store
        // in Neo4j + expose via REST). Default to false per schema contract.
        is_abstract: false,
        is_deprecated: false,
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
        entity_type: GqlEntityType::Individual,
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
        entity_type: GqlEntityType::Individual,
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
        entity_type: GqlEntityType::Property,
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
        // Convert ClassTreeNode → GqlClassTreeNode recursively
        fn convert(node: classes::ClassTreeNode) -> GqlClassTreeNode {
            GqlClassTreeNode {
                id: node.id,
                label: node.label,
                children: node.children.into_iter().map(convert).collect(),
            }
        }

        let pool = pool_from_ctx(ctx)?;
        let repo = ClassRepository::new(pool);
        let tree = repo
            .get_descendants_tree(&ontology_id, &class_id, max_depth)
            .await
            .map_err(map_error)?;
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
            GqlPropertyType::Annotation => "annotation".to_string(),
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
}

// ── Converter Tests ────────────────────────────────────────────────────────────
//
// The converters populate the `Entity` interface discriminator and the
// Class boolean flags per REQ-USR.UI.graph-navigation.md. These tests pin
// the values so the contract stays stable.

#[cfg(test)]
mod tests {
    use super::*;
    use crate::classes::OwlClass;
    use crate::individuals::{IndividualSummary, LiteralValue, ReferenceValue};
    use crate::properties::{
        Annotation as DomainAnnotation, Property, PropertyCharacteristics, PropertyType,
    };

    #[test]
    fn gql_class_sets_entity_discriminator_and_flags() {
        let cls = OwlClass {
            id: "Person".to_string(),
            label: "Person".to_string(),
            comment: Some("A person".to_string()),
            parents: vec!["Thing".to_string()],
            children: vec![],
        };

        let gql = gql_class_from(cls);

        assert_eq!(gql.entity_type, GqlEntityType::Class);
        assert!(!gql.is_abstract, "isAbstract must default to false");
        assert!(!gql.is_deprecated, "isDeprecated must default to false");
        assert_eq!(gql.id, "Person");
        assert_eq!(gql.parents, vec!["Thing".to_string()]);
    }

    #[test]
    fn gql_property_sets_entity_discriminator() {
        let prop = Property {
            id: "hasName".to_string(),
            label: "has name".to_string(),
            comment: None,
            property_type: PropertyType::Annotation,
            domains: vec!["Person".to_string()],
            ranges: vec![],
            xsd_type: None,
            characteristics: PropertyCharacteristics::default(),
            annotations: vec![DomainAnnotation {
                property_iri: "skos:definition".to_string(),
                value: "Full name".to_string(),
            }],
        };

        let gql = gql_property_from_full(prop);

        assert_eq!(gql.entity_type, GqlEntityType::Property);
        assert_eq!(gql.property_type, GqlPropertyType::Annotation);
        assert_eq!(gql.id, "hasName");
    }

    #[test]
    fn gql_individual_from_summary_sets_entity_discriminator() {
        let ind = IndividualSummary {
            id: "alice".to_string(),
            label: "Alice".to_string(),
            comment: None,
            class_id: "Person".to_string(),
            class_label: "Person".to_string(),
        };

        let gql = gql_individual_from_summary(ind);

        assert_eq!(gql.entity_type, GqlEntityType::Individual);
        assert_eq!(gql.class_id, "Person");
    }

    #[test]
    fn gql_individual_from_full_sets_entity_discriminator() {
        let ind = crate::individuals::Individual {
            id: "alice".to_string(),
            label: "Alice".to_string(),
            comment: None,
            class_id: "Person".to_string(),
            class_label: "Person".to_string(),
            literal_values: vec![LiteralValue {
                property_id: "age".to_string(),
                property_label: "age".to_string(),
                value: "30".to_string(),
                xsd_type: Some("integer".to_string()),
                value_id: None,
            }],
            reference_values: vec![ReferenceValue {
                property_id: "worksFor".to_string(),
                property_label: "works for".to_string(),
                target_id: "acme".to_string(),
                target_label: "Acme".to_string(),
                edge_id: None,
            }],
        };

        let gql = gql_individual_from_full(ind);

        assert_eq!(gql.entity_type, GqlEntityType::Individual);
        assert_eq!(gql.literal_values.len(), 1);
        assert_eq!(gql.reference_values.len(), 1);
    }

    #[test]
    fn gql_property_type_annotation_round_trips() {
        let domain: crate::properties::PropertyType = GqlPropertyType::Annotation.into();
        assert_eq!(domain, PropertyType::Annotation);

        let gql: GqlPropertyType = PropertyType::Annotation.into();
        assert_eq!(gql, GqlPropertyType::Annotation);
    }
}
