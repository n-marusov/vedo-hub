//! GraphQL object types wrapping ontology domain models.
//!
//! Each type mirrors a domain struct from `classes.rs`, `properties.rs`,
//! or `individuals.rs` and delegates to the existing `async_graphql` macros.

use async_graphql::{Enum, SimpleObject};

/// A class in the ontology hierarchy (TBox).
#[derive(SimpleObject)]
#[graphql(name = "Class")]
pub struct GqlClass {
    pub id: String,
    pub label: String,
    pub comment: Option<String>,
    pub parents: Vec<String>,
    pub children: Vec<String>,
}

/// Lightweight class summary used in list responses.
#[derive(SimpleObject)]
#[graphql(name = "ClassSummary")]
pub struct GqlClassSummary {
    pub id: String,
    pub label: String,
    pub comment: Option<String>,
    pub parents: Vec<String>,
}

/// A node in the class hierarchy tree.
#[derive(SimpleObject)]
#[graphql(name = "ClassTreeNode")]
pub struct GqlClassTreeNode {
    pub id: String,
    pub label: String,
    /// Nested children (populated when tree is fetched).
    pub children: Vec<GqlClassTreeNode>,
}

/// Breadcrumb path item from root to a given class.
#[derive(SimpleObject)]
#[graphql(name = "BreadcrumbItem")]
pub struct GqlBreadcrumbItem {
    pub id: String,
    pub label: String,
}

/// Paginated connection wrapper for classes.
#[derive(SimpleObject)]
#[graphql(name = "ClassConnection")]
pub struct GqlClassConnection {
    pub items: Vec<GqlClassSummary>,
    pub total: u64,
    pub page: u64,
    pub per_page: u64,
}

/// The property type discriminator.
#[derive(Enum, Copy, Clone, Eq, PartialEq)]
#[graphql(name = "PropertyType")]
pub enum GqlPropertyType {
    Object,
    Datatype,
}

impl From<crate::properties::PropertyType> for GqlPropertyType {
    fn from(t: crate::properties::PropertyType) -> Self {
        match t {
            crate::properties::PropertyType::Object => GqlPropertyType::Object,
            crate::properties::PropertyType::Datatype => GqlPropertyType::Datatype,
        }
    }
}

impl From<GqlPropertyType> for crate::properties::PropertyType {
    fn from(t: GqlPropertyType) -> Self {
        match t {
            GqlPropertyType::Object => crate::properties::PropertyType::Object,
            GqlPropertyType::Datatype => crate::properties::PropertyType::Datatype,
        }
    }
}

/// ObjectProperty characteristics (functional, transitive, etc.).
#[derive(SimpleObject)]
#[graphql(name = "PropertyCharacteristics")]
pub struct GqlPropertyCharacteristics {
    pub functional: bool,
    pub inverse_functional: bool,
    pub transitive: bool,
    pub symmetric: bool,
}

impl From<crate::properties::PropertyCharacteristics> for GqlPropertyCharacteristics {
    fn from(c: crate::properties::PropertyCharacteristics) -> Self {
        Self {
            functional: c.functional,
            inverse_functional: c.inverse_functional,
            transitive: c.transitive,
            symmetric: c.symmetric,
        }
    }
}

/// An annotation on a property (custom annotation property).
#[derive(SimpleObject)]
#[graphql(name = "Annotation")]
pub struct GqlAnnotation {
    pub property_iri: String,
    pub value: String,
}

impl From<crate::properties::Annotation> for GqlAnnotation {
    fn from(a: crate::properties::Annotation) -> Self {
        Self {
            property_iri: a.property_iri,
            value: a.value,
        }
    }
}

/// A property in the ontology (ObjectProperty or DatatypeProperty).
#[derive(SimpleObject)]
#[graphql(name = "Property")]
pub struct GqlProperty {
    pub id: String,
    pub label: String,
    pub comment: Option<String>,
    pub property_type: GqlPropertyType,
    pub domains: Vec<String>,
    pub ranges: Vec<String>,
    pub xsd_type: Option<String>,
    pub characteristics: GqlPropertyCharacteristics,
    pub annotations: Vec<GqlAnnotation>,
}

/// Lightweight property summary for list responses.
#[derive(SimpleObject)]
#[graphql(name = "PropertySummary")]
pub struct GqlPropertySummary {
    pub id: String,
    pub label: String,
    pub property_type: GqlPropertyType,
    pub xsd_type: Option<String>,
    pub domains: Vec<String>,
}

/// Paginated connection wrapper for properties.
#[derive(SimpleObject)]
#[graphql(name = "PropertyConnection")]
pub struct GqlPropertyConnection {
    pub items: Vec<GqlPropertySummary>,
    pub total: u64,
    pub page: u64,
    pub per_page: u64,
}

/// A literal property value on an individual.
#[derive(SimpleObject)]
#[graphql(name = "LiteralValue")]
pub struct GqlLiteralValue {
    pub property_id: String,
    pub property_label: String,
    pub value: String,
    pub xsd_type: Option<String>,
    pub value_id: Option<String>,
}

/// A reference property value linking to another individual.
#[derive(SimpleObject)]
#[graphql(name = "ReferenceValue")]
pub struct GqlReferenceValue {
    pub property_id: String,
    pub property_label: String,
    pub target_id: String,
    pub target_label: String,
    pub edge_id: Option<String>,
}

/// An OWL individual (ABox instance).
#[derive(SimpleObject)]
#[graphql(name = "Individual")]
pub struct GqlIndividual {
    pub id: String,
    pub label: String,
    pub comment: Option<String>,
    pub class_id: String,
    pub class_label: String,
    pub literal_values: Vec<GqlLiteralValue>,
    pub reference_values: Vec<GqlReferenceValue>,
}

/// Paginated connection wrapper for individuals.
#[derive(SimpleObject)]
#[graphql(name = "IndividualConnection")]
pub struct GqlIndividualConnection {
    pub items: Vec<GqlIndividual>,
    pub total: u64,
    pub page: u64,
    pub per_page: u64,
}

/// A node in the graph neighborhood.
#[derive(SimpleObject)]
#[graphql(name = "GraphNode")]
pub struct GqlGraphNode {
    pub id: String,
    pub label: String,
}

/// An edge in the graph neighborhood.
#[derive(SimpleObject)]
#[graphql(name = "GraphEdge")]
pub struct GqlGraphEdge {
    pub source_id: String,
    pub target_id: String,
    pub property_id: String,
    pub property_label: String,
}

/// Graph neighborhood result (nodes + edges).
#[derive(SimpleObject)]
#[graphql(name = "GraphNeighborhood")]
pub struct GqlGraphNeighborhood {
    pub nodes: Vec<GqlGraphNode>,
    pub edges: Vec<GqlGraphEdge>,
}

/// Result of a delete operation.
#[derive(SimpleObject)]
#[graphql(name = "DeleteResult")]
pub struct GqlDeleteResult {
    pub deleted: bool,
    pub entity_id: String,
    pub warning: Option<String>,
    pub dependent_count: Option<u64>,
}
