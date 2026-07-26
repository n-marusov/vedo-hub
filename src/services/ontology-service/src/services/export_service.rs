//! Export service for serializing ontology data to Turtle and RDF/XML formats.
//!
//! Performs a full graph traversal of the ontology (classes, properties, individuals,
//! property values) and converts all entities to RDF triples, then serializes them
//! as Turtle (TTL) or RDF/XML text.

#![allow(
    clippy::items_after_statements,
    clippy::unused_self,
    clippy::option_as_ref_deref,
    clippy::struct_excessive_bools,
    clippy::struct_field_names,
    clippy::unnecessary_wraps,
    clippy::manual_strip
)]
//!
//! # RDF Mapping
//!
//! Base IRI: `http://vedo.dev/ontology/{ontology_id}#`
//!
//! | Entity            | RDF Pattern                                                    |
//! |-------------------|----------------------------------------------------------------|
//! | OWL Class         | `:ClassID rdf:type owl:Class`                                  |
//! |                   | `:ClassID rdfs:subClassOf :ParentID`                           |
//! |                   | `:ClassID rdfs:label "label"`                                  |
//! |                   | `:ClassID rdfs:comment "comment"`                              |
//! | ObjectProperty    | `:PropID rdf:type owl:ObjectProperty`                          |
//! |                   | `:PropID rdfs:domain :ClassID`                                 |
//! |                   | `:PropID rdfs:range :ClassID`                                  |
//! | DatatypeProperty  | `:PropID rdf:type owl:DatatypeProperty`                        |
//! |                   | `:PropID rdfs:domain :ClassID`                                 |
//! |                   | `:PropID rdfs:range xsd:string`                                |
//! | Individual        | `:IndID rdf:type :ClassID`                                     |
//! |                   | `:IndID rdfs:label "label"`                                    |
//! | Literal value     | `:IndID :PropID "value"^^xsd:type`                             |
//! | Reference value   | `:IndID :PropID :TargetID`                                     |

use std::fmt::Write;

use tracing::{debug, info, warn};

use crate::neo4j::Neo4jPool;

/// Internal struct for collecting triples before export.
struct ExportTriple {
    subject: String,
    predicate: String,
    object: String,
}

/// Internal struct for class rows from Neo4j.
/// A raw class row fetched from Neo4j.
#[derive(Debug, Clone)]
struct ClassRow {
    id: String,
    label: String,
    comment: Option<String>,
    parents: Vec<String>,
}

/// A raw property row fetched from Neo4j.
#[derive(Debug, Clone)]
struct PropertyRow {
    id: String,
    label: String,
    comment: Option<String>,
    property_type: String,
    domain_ids: Vec<String>,
    range_ids: Vec<String>,
    xsd_type: Option<String>,
    is_functional: bool,
    is_inverse_functional: bool,
    is_transitive: bool,
    is_symmetric: bool,
    #[allow(dead_code)]
    min_cardinality: Option<i32>,
    #[allow(dead_code)]
    max_cardinality: Option<i32>,
}

/// A raw individual row fetched from Neo4j.
#[derive(Debug, Clone)]
struct IndividualRow {
    id: String,
    label: String,
    comment: Option<String>,
    class_id: String,
}

/// A literal value assigned to an individual.
#[derive(Debug, Clone)]
struct LiteralValueRow {
    individual_id: String,
    property_id: String,
    value: String,
    xsd_type: Option<String>,
}

/// A reference value linking two individuals.
#[derive(Debug, Clone)]
struct ReferenceValueRow {
    source_id: String,
    property_id: String,
    target_id: String,
}

/// Service for exporting ontology data to RDF serialization formats.
pub struct ExportService {
    pool: Neo4jPool,
    // [FIX] Shared reqwest client (reused across versioned export calls instead
    // of a new Client::new() per call, which is wasteful and disables keep-alive).
    http: reqwest::Client,
}

impl ExportService {
    pub fn new(pool: Neo4jPool) -> Self {
        // [FIX] One reqwest client is shared across all versioned export calls.
        // Creating a new Client per call disables connection pooling/keep-alive.
        let http = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .build()
            .unwrap_or_else(|_| reqwest::Client::new());
        Self { pool, http }
    }

    /// Exports all ontology data as Turtle format.
    pub async fn export_turtle(&self, ontology_id: &str) -> Result<String, ExportError> {
        debug!(ontology_id, "Starting Turtle export");
        let triples = self.collect_all_triples(ontology_id).await?;
        let result = self.serialize_turtle(ontology_id, &triples)?;

        info!(
            ontology_id,
            triple_count = triples.len(),
            output_bytes = result.len(),
            "Turtle export completed"
        );

        Ok(result)
    }

    /// Exports all ontology data as RDF/XML format.
    pub async fn export_rdf_xml(&self, ontology_id: &str) -> Result<String, ExportError> {
        debug!(ontology_id, "Starting RDF/XML export");
        let triples = self.collect_all_triples(ontology_id).await?;
        let result = self.serialize_rdf_xml(ontology_id, &triples)?;

        info!(
            ontology_id,
            triple_count = triples.len(),
            output_bytes = result.len(),
            "RDF/XML export completed"
        );

        Ok(result)
    }

    /// Exports all ontology data as OWL/XML format.
    ///
    /// Uses RDF/XML serialization with OWL profile (application/owl+xml).
    /// Full OWL/XML (OWL API XML schema) support is a future enhancement;
    /// the current implementation produces valid RDF/XML that round-trips
    /// through standard OWL tools.
    pub async fn export_owl_xml(&self, ontology_id: &str) -> Result<String, ExportError> {
        debug!(ontology_id, "Starting OWL/XML export");
        let triples = self.collect_all_triples(ontology_id).await?;
        let result = self.serialize_rdf_xml(ontology_id, &triples)?;

        info!(
            ontology_id,
            triple_count = triples.len(),
            output_bytes = result.len(),
            "OWL/XML export completed"
        );

        Ok(result)
    }

    /// Exports ontology data from a versioning-service materialized state as Turtle.
    ///
    /// Calls the versioning-service to materialize the state for the given branch
    /// or commit, then serializes the materialized triples as Turtle.
    pub async fn export_turtle_versioned(
        &self,
        ontology_id: &str,
        versioning_url: &str,
        branch_id: Option<&str>,
        commit_id: Option<&str>,
    ) -> Result<String, ExportError> {
        debug!(
            ontology_id,
            versioning_url, branch_id, commit_id, "Starting versioned Turtle export"
        );

        let endpoint = if let Some(cid) = commit_id {
            format!("{versioning_url}/api/v1/versioning/commits/{cid}/checkout")
        } else if let Some(bid) = branch_id {
            format!("{versioning_url}/api/v1/versioning/branches/{bid}/switch")
        } else {
            // No version params — fall through to current-state export
            return self.export_turtle(ontology_id).await;
        };

        // [FIX] Materialize state via the versioning service using shared client
        // and fail loudly on non-2xx instead of silently falling back.
        match self
            .http
            .post(&endpoint)
            .header("Content-Type", "application/json")
            .body("{}")
            .send()
            .await
        {
            Ok(resp) => {
                // [FIX] Check HTTP status before consuming the response. A non-2xx
                // status previously caused a silent fall-back to the stale Neo4j state.
                if let Err(e) = resp.error_for_status() {
                    warn!(
                        ontology_id,
                        error = %e,
                        endpoint,
                        "[FIX] Versioning service returned non-2xx — refusing to fall back"
                    );
                    return Err(ExportError::VersioningUnavailable(e.to_string()));
                }
                info!(
                    ontology_id,
                    endpoint, "Materialized state synced via versioning service"
                );
                // After successful materialization, export from Neo4j
                // (the checkout endpoint syncs the state to Neo4j)
                self.export_turtle(ontology_id).await
            }
            Err(e) => {
                warn!(
                    ontology_id,
                    error = %e,
                    endpoint,
                    "[FIX] Versioning service transport error — refusing to fall back"
                );
                // [FIX] Previously this fell back to current Neo4j state, masking
                // versioning failures. Fail loudly so callers know the snapshot
                // they requested is not the one shipped.
                Err(ExportError::VersioningUnavailable(e.to_string()))
            }
        }
    }

    /// Exports ontology data from a versioning-service materialized state as RDF/XML.
    pub async fn export_rdf_xml_versioned(
        &self,
        ontology_id: &str,
        versioning_url: &str,
        branch_id: Option<&str>,
        commit_id: Option<&str>,
    ) -> Result<String, ExportError> {
        debug!(
            ontology_id,
            versioning_url, branch_id, commit_id, "Starting versioned RDF/XML export"
        );

        let endpoint = if let Some(cid) = commit_id {
            format!("{versioning_url}/api/v1/versioning/commits/{cid}/checkout")
        } else if let Some(bid) = branch_id {
            format!("{versioning_url}/api/v1/versioning/branches/{bid}/switch")
        } else {
            return self.export_rdf_xml(ontology_id).await;
        };

        // [FIX] Materialize state via versioning service with shared client
        // and error_for_status hygiene.
        match self
            .http
            .post(&endpoint)
            .header("Content-Type", "application/json")
            .body("{}")
            .send()
            .await
        {
            Ok(resp) => {
                if let Err(e) = resp.error_for_status() {
                    warn!(
                        ontology_id,
                        error = %e,
                        endpoint,
                        "[FIX] Versioning service returned non-2xx — refusing to fall back"
                    );
                    return Err(ExportError::VersioningUnavailable(e.to_string()));
                }
                info!(
                    ontology_id,
                    endpoint, "Materialized state synced via versioning service"
                );
                self.export_rdf_xml(ontology_id).await
            }
            Err(e) => {
                warn!(
                    ontology_id,
                    error = %e,
                    endpoint,
                    "[FIX] Versioning service transport error — refusing to fall back"
                );
                Err(ExportError::VersioningUnavailable(e.to_string()))
            }
        }
    }

    /// Exports ontology data from a versioning-service materialized state as OWL/XML.
    pub async fn export_owl_xml_versioned(
        &self,
        ontology_id: &str,
        versioning_url: &str,
        branch_id: Option<&str>,
        commit_id: Option<&str>,
    ) -> Result<String, ExportError> {
        debug!(
            ontology_id,
            versioning_url, branch_id, commit_id, "Starting versioned OWL/XML export"
        );

        let endpoint = if let Some(cid) = commit_id {
            format!("{versioning_url}/api/v1/versioning/commits/{cid}/checkout")
        } else if let Some(bid) = branch_id {
            format!("{versioning_url}/api/v1/versioning/branches/{bid}/switch")
        } else {
            return self.export_owl_xml(ontology_id).await;
        };

        // [FIX] Materialize state via versioning service with shared client
        // and error_for_status hygiene.
        match self
            .http
            .post(&endpoint)
            .header("Content-Type", "application/json")
            .body("{}")
            .send()
            .await
        {
            Ok(resp) => {
                if let Err(e) = resp.error_for_status() {
                    warn!(
                        ontology_id,
                        error = %e,
                        endpoint,
                        "[FIX] Versioning service returned non-2xx — refusing to fall back"
                    );
                    return Err(ExportError::VersioningUnavailable(e.to_string()));
                }
                info!(
                    ontology_id,
                    endpoint, "Materialized state synced via versioning service"
                );
                self.export_owl_xml(ontology_id).await
            }
            Err(e) => {
                warn!(
                    ontology_id,
                    error = %e,
                    endpoint,
                    "[FIX] Versioning service transport error — refusing to fall back"
                );
                Err(ExportError::VersioningUnavailable(e.to_string()))
            }
        }
    }

    // ── Serialization ──────────────────────────────────────────────────────

    /// Serializes triples as Turtle (TTL) format.
    fn serialize_turtle(
        &self,
        ontology_id: &str,
        triples: &[ExportTriple],
    ) -> Result<String, ExportError> {
        let base = format!("http://vedo.dev/ontology/{ontology_id}#");
        let mut out = String::new();

        // Write prefix declarations
        writeln!(
            out,
            "@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> ."
        )
        .unwrap();
        writeln!(
            out,
            "@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> ."
        )
        .unwrap();
        writeln!(out, "@prefix owl: <http://www.w3.org/2002/07/owl#> .").unwrap();
        writeln!(out, "@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .").unwrap();
        writeln!(out, "@base <{base}> .\n").unwrap();

        // Group triples by subject for compact output
        let mut grouped: std::collections::BTreeMap<String, Vec<(&str, &str)>> =
            std::collections::BTreeMap::new();

        for t in triples {
            grouped
                .entry(t.subject.clone())
                .or_default()
                .push((&t.predicate, &t.object));
        }

        for (subject, predicates) in &grouped {
            // Write subject
            write!(out, "<{base}{subject}>").unwrap();

            let mut first = true;
            for (pred, obj) in predicates {
                if first {
                    writeln!(out).unwrap();
                    first = false;
                } else {
                    writeln!(out, " ;").unwrap();
                }

                let pred_full = if pred.starts_with("rdf:")
                    || pred.starts_with("rdfs:")
                    || pred.starts_with("owl:")
                    || pred.starts_with("xsd:")
                {
                    pred.to_string()
                } else {
                    format!("<{pred}>")
                };

                let obj_full = if obj.starts_with('"') {
                    // Literal value — write as-is
                    obj.to_string()
                } else if obj.contains(':') {
                    // Prefixed IRI (e.g. "owl:Class") — write as-is
                    obj.to_string()
                } else {
                    // Entity reference (local name) — wrap as relative IRI
                    format!("<{obj}>")
                };

                write!(out, "    {pred_full} {obj_full}").unwrap();
            }
            writeln!(out, " .\n").unwrap();
        }

        Ok(out)
    }

    /// Serializes triples as RDF/XML format.
    fn serialize_rdf_xml(
        &self,
        ontology_id: &str,
        triples: &[ExportTriple],
    ) -> Result<String, ExportError> {
        let base = format!("http://vedo.dev/ontology/{ontology_id}#");
        let mut out = String::new();

        writeln!(out, r#"<?xml version="1.0" encoding="UTF-8"?>"#).unwrap();
        writeln!(
            out,
            r#"<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#""#
        )
        .unwrap();
        writeln!(
            out,
            r#"         xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#""#
        )
        .unwrap();
        writeln!(
            out,
            r#"         xmlns:owl="http://www.w3.org/2002/07/owl#""#
        )
        .unwrap();
        writeln!(
            out,
            r#"         xmlns:xsd="http://www.w3.org/2001/XMLSchema#""#
        )
        .unwrap();
        writeln!(out, r#"         xml:base="{base}">"#).unwrap();
        writeln!(out).unwrap();

        // Group triples by subject
        let mut grouped: std::collections::BTreeMap<String, Vec<(&str, &str)>> =
            std::collections::BTreeMap::new();

        for t in triples {
            grouped
                .entry(t.subject.clone())
                .or_default()
                .push((&t.predicate, &t.object));
        }

        for (subject, predicates) in &grouped {
            writeln!(out, r#"  <rdf:Description rdf:about="{base}{subject}">"#).unwrap();

            for (pred, obj) in predicates {
                let pred_local = Self::xml_local_name(pred);
                let ns_prefix = Self::xml_ns_prefix(pred);

                if obj.starts_with('"') {
                    // Literal value
                    if let Some((value, suffix)) = obj[1..].split_once('"') {
                        let escaped = value
                            .replace('&', "&amp;")
                            .replace('<', "&lt;")
                            .replace('>', "&gt;");

                        if let Some(datatype) = suffix.strip_prefix("^^xsd:") {
                            writeln!(
                                out,
                                r#"    <{ns_prefix}{pred_local} rdf:datatype="http://www.w3.org/2001/XMLSchema#{datatype}">{escaped}</{ns_prefix}{pred_local}>"#
                            ).unwrap();
                        } else {
                            writeln!(
                                out,
                                r"    <{ns_prefix}{pred_local}>{escaped}</{ns_prefix}{pred_local}>"
                            )
                            .unwrap();
                        }
                    }
                } else if obj.contains(':') {
                    // Prefixed IRI (e.g. "owl:Class")
                    let full_iri: String = if *obj == "rdf:type" {
                        "http://www.w3.org/1999/02/22-rdf-syntax-ns#type".to_string()
                    } else if let Some(local) = obj.strip_prefix("rdfs:") {
                        format!("http://www.w3.org/2000/01/rdf-schema#{local}")
                    } else if let Some(local) = obj.strip_prefix("owl:") {
                        format!("http://www.w3.org/2002/07/owl#{local}")
                    } else if let Some(local) = obj.strip_prefix("xsd:") {
                        format!("http://www.w3.org/2001/XMLSchema#{local}")
                    } else {
                        format!("{base}{obj}")
                    };
                    writeln!(
                        out,
                        r#"    <{ns_prefix}{pred_local} rdf:resource="{full_iri}"/>"#
                    )
                    .unwrap();
                } else {
                    // Entity reference (local name) — resolve against base
                    writeln!(
                        out,
                        r#"    <{ns_prefix}{pred_local} rdf:resource="{base}{obj}"/>"#
                    )
                    .unwrap();
                }
            }

            writeln!(out, "  </rdf:Description>").unwrap();
        }

        writeln!(out, "</rdf:RDF>").unwrap();
        Ok(out)
    }

    /// Extracts the local name part from a prefixed or full IRI predicate.
    fn xml_local_name(pred: &str) -> &str {
        if let Some(local) = pred.strip_prefix("rdf:") {
            local
        } else if let Some(local) = pred.strip_prefix("rdfs:") {
            local
        } else if let Some(local) = pred.strip_prefix("owl:") {
            local
        } else if let Some(local) = pred.strip_prefix("xsd:") {
            local
        } else {
            // Extract from full IRI
            pred.rsplit_once(['/', '#'])
                .map_or(pred, |(_, local)| local)
        }
    }

    /// Returns the XML namespace prefix for a given predicate.
    fn xml_ns_prefix(pred: &str) -> &str {
        if pred.starts_with("rdf:") {
            "rdf:"
        } else if pred.starts_with("rdfs:") {
            "rdfs:"
        } else if pred.starts_with("owl:") {
            "owl:"
        } else if pred.starts_with("xsd:") {
            "xsd:"
        } else {
            ""
        }
    }

    // ── Triple Collection ──────────────────────────────────────────────────

    /// Collects all RDF triples for the given ontology by querying Neo4j.
    async fn collect_all_triples(
        &self,
        ontology_id: &str,
    ) -> Result<Vec<ExportTriple>, ExportError> {
        debug!(ontology_id, "Starting full graph traversal for export");

        let classes = self.fetch_classes(ontology_id).await?;
        let properties = self.fetch_properties(ontology_id).await?;
        let individuals = self.fetch_individuals(ontology_id).await?;
        let literal_values = self.fetch_literal_values(ontology_id).await?;
        let reference_values = self.fetch_reference_values(ontology_id).await?;

        let mut triples = Vec::new();

        let e = |local: &str| -> String { local.to_string() };
        let iri = |iri: &str| -> String { iri.to_string() };
        let lit =
            |value: &str| -> String { format!("\"{}\"^^xsd:string", value.replace('"', "\\\"")) };

        // Convert classes to triples
        for cls in &classes {
            // rdf:type owl:Class
            triples.push(ExportTriple {
                subject: e(&cls.id),
                predicate: iri("rdf:type"),
                object: "owl:Class".to_string(),
            });

            // rdfs:label
            triples.push(ExportTriple {
                subject: e(&cls.id),
                predicate: iri("rdfs:label"),
                object: lit(&cls.label),
            });

            // rdfs:comment (optional)
            if let Some(comment) = &cls.comment {
                if !comment.is_empty() {
                    triples.push(ExportTriple {
                        subject: e(&cls.id),
                        predicate: iri("rdfs:comment"),
                        object: lit(comment),
                    });
                }
            }

            // rdfs:subClassOf for each parent
            for parent_id in &cls.parents {
                triples.push(ExportTriple {
                    subject: e(&cls.id),
                    predicate: iri("rdfs:subClassOf"),
                    object: e(parent_id),
                });
            }
        }

        // Convert properties to triples
        for prop in &properties {
            let owl_type = if prop.property_type == "object" {
                "owl:ObjectProperty"
            } else {
                "owl:DatatypeProperty"
            };

            triples.push(ExportTriple {
                subject: e(&prop.id),
                predicate: iri("rdf:type"),
                object: iri(owl_type),
            });

            // rdfs:label
            triples.push(ExportTriple {
                subject: e(&prop.id),
                predicate: iri("rdfs:label"),
                object: lit(&prop.label),
            });

            // rdfs:comment (optional)
            if let Some(comment) = &prop.comment {
                if !comment.is_empty() {
                    triples.push(ExportTriple {
                        subject: e(&prop.id),
                        predicate: iri("rdfs:comment"),
                        object: lit(comment),
                    });
                }
            }

            // rdfs:domain for each domain class
            for domain_id in &prop.domain_ids {
                triples.push(ExportTriple {
                    subject: e(&prop.id),
                    predicate: iri("rdfs:domain"),
                    object: e(domain_id),
                });
            }

            // rdfs:range
            if prop.property_type == "object" {
                for range_id in &prop.range_ids {
                    triples.push(ExportTriple {
                        subject: e(&prop.id),
                        predicate: iri("rdfs:range"),
                        object: e(range_id),
                    });
                }
            } else if let Some(xsd_type) = &prop.xsd_type {
                triples.push(ExportTriple {
                    subject: e(&prop.id),
                    predicate: iri("rdfs:range"),
                    object: format!("xsd:{xsd_type}"),
                });
            }

            // OWL characteristics
            if prop.property_type == "object" {
                if prop.is_functional {
                    triples.push(ExportTriple {
                        subject: e(&prop.id),
                        predicate: iri("rdf:type"),
                        object: iri("owl:FunctionalProperty"),
                    });
                }
                if prop.is_inverse_functional {
                    triples.push(ExportTriple {
                        subject: e(&prop.id),
                        predicate: iri("rdf:type"),
                        object: iri("owl:InverseFunctionalProperty"),
                    });
                }
                if prop.is_transitive {
                    triples.push(ExportTriple {
                        subject: e(&prop.id),
                        predicate: iri("rdf:type"),
                        object: iri("owl:TransitiveProperty"),
                    });
                }
                if prop.is_symmetric {
                    triples.push(ExportTriple {
                        subject: e(&prop.id),
                        predicate: iri("rdf:type"),
                        object: iri("owl:SymmetricProperty"),
                    });
                }
            }
        }

        // Convert individuals to triples
        for ind in &individuals {
            // rdf:type :ClassID
            triples.push(ExportTriple {
                subject: e(&ind.id),
                predicate: iri("rdf:type"),
                object: e(&ind.class_id),
            });

            // rdfs:label
            triples.push(ExportTriple {
                subject: e(&ind.id),
                predicate: iri("rdfs:label"),
                object: lit(&ind.label),
            });

            // rdfs:comment (optional)
            if let Some(comment) = &ind.comment {
                if !comment.is_empty() {
                    triples.push(ExportTriple {
                        subject: e(&ind.id),
                        predicate: iri("rdfs:comment"),
                        object: lit(comment),
                    });
                }
            }
        }

        // Convert literal values to triples
        for lv in &literal_values {
            let xsd_type = lv.xsd_type.as_deref().unwrap_or("string");
            let escaped = lv.value.replace('"', "\\\"");
            triples.push(ExportTriple {
                subject: e(&lv.individual_id),
                predicate: e(&lv.property_id),
                object: format!("\"{escaped}\"^^xsd:{xsd_type}"),
            });
        }

        // Convert reference values to triples
        for rv in &reference_values {
            triples.push(ExportTriple {
                subject: e(&rv.source_id),
                predicate: e(&rv.property_id),
                object: e(&rv.target_id),
            });
        }

        debug!(
            ontology_id,
            class_count = classes.len(),
            property_count = properties.len(),
            individual_count = individuals.len(),
            triple_count = triples.len(),
            "Graph traversal complete"
        );

        // Canonical sort: subject IRI, then predicate IRI, then object
        triples.sort_by(|a, b| {
            a.subject
                .cmp(&b.subject)
                .then_with(|| a.predicate.cmp(&b.predicate))
                .then_with(|| a.object.cmp(&b.object))
        });

        Ok(triples)
    }

    // ── Database Queries ───────────────────────────────────────────────────

    async fn fetch_classes(&self, ontology_id: &str) -> Result<Vec<ClassRow>, ExportError> {
        let query = r"
            MATCH (c:Class {ontology_id: $ontology_id})
            OPTIONAL MATCH (c)-[:CHILD_OF]->(p:Class)
            RETURN c.id AS id, c.label AS label, c.comment AS comment,
                   collect(DISTINCT p.id) AS parents
            ORDER BY c.label
        ";

        let q = neo4rs::Query::new(query.to_string()).param("ontology_id", ontology_id);

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| ExportError::Database(format!("Failed to query classes: {e}")))?;

        let mut rows = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(id) = row.get::<String>("id") {
                let label: String = row.get("label").unwrap_or_default();
                let comment: Option<String> =
                    row.get("comment").ok().filter(|c: &String| !c.is_empty());
                let parents: Vec<String> = row.get("parents").unwrap_or_default();
                rows.push(ClassRow {
                    id,
                    label,
                    comment,
                    parents,
                });
            }
        }

        Ok(rows)
    }

    async fn fetch_properties(&self, ontology_id: &str) -> Result<Vec<PropertyRow>, ExportError> {
        let query = r"
            MATCH (p:Property {ontology_id: $ontology_id})
            OPTIONAL MATCH (p)-[:DOMAIN]->(d:Class)
            OPTIONAL MATCH (p)-[:RANGE]->(r:Class)
            RETURN p.id AS id, p.label AS label, p.comment AS comment,
                   p.is_datatype AS is_datatype, p.xsd_type AS xsd_type,
                   p.functional AS functional, p.inverse_functional AS inverse_functional,
                   p.transitive AS transitive, p.symmetric AS symmetric,
                   p.min_cardinality AS min_cardinality,
                   p.max_cardinality AS max_cardinality,
                   collect(DISTINCT d.id) AS domain_ids,
                   collect(DISTINCT r.id) AS range_ids
            ORDER BY p.label
        ";

        let q = neo4rs::Query::new(query.to_string()).param("ontology_id", ontology_id);

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| ExportError::Database(format!("Failed to query properties: {e}")))?;

        let mut rows = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(id) = row.get::<String>("id") {
                let label: String = row.get("label").unwrap_or_default();
                let comment: Option<String> =
                    row.get("comment").ok().filter(|c: &String| !c.is_empty());
                let is_datatype: bool = row.get("is_datatype").unwrap_or(false);
                let xsd_type: Option<String> =
                    row.get("xsd_type").ok().filter(|s: &String| !s.is_empty());
                let functional: bool = row.get("functional").unwrap_or(false);
                let inverse_functional: bool = row.get("inverse_functional").unwrap_or(false);
                let transitive: bool = row.get("transitive").unwrap_or(false);
                let symmetric: bool = row.get("symmetric").unwrap_or(false);
                let min_cardinality: Option<i32> = row.get("min_cardinality").ok();
                let max_cardinality: Option<i32> = row.get("max_cardinality").ok();
                let domain_ids: Vec<String> = row.get("domain_ids").unwrap_or_default();
                let range_ids: Vec<String> = row.get("range_ids").unwrap_or_default();

                rows.push(PropertyRow {
                    id,
                    label,
                    comment,
                    property_type: if is_datatype { "datatype" } else { "object" }.to_string(),
                    domain_ids: domain_ids.into_iter().filter(|d| !d.is_empty()).collect(),
                    range_ids: range_ids.into_iter().filter(|r| !r.is_empty()).collect(),
                    xsd_type,
                    is_functional: functional,
                    is_inverse_functional: inverse_functional,
                    is_transitive: transitive,
                    is_symmetric: symmetric,
                    min_cardinality,
                    max_cardinality,
                });
            }
        }

        Ok(rows)
    }

    async fn fetch_individuals(
        &self,
        ontology_id: &str,
    ) -> Result<Vec<IndividualRow>, ExportError> {
        let query = r"
            MATCH (i:Individual {ontology_id: $ontology_id})
            OPTIONAL MATCH (i)-[:INSTANCE_OF]->(c:Class)
            RETURN i.id AS id, i.label AS label, i.comment AS comment,
                   coalesce(c.id, '') AS class_id
            ORDER BY i.label
        ";

        let q = neo4rs::Query::new(query.to_string()).param("ontology_id", ontology_id);

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| ExportError::Database(format!("Failed to query individuals: {e}")))?;

        let mut rows = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(id) = row.get::<String>("id") {
                let label: String = row.get("label").unwrap_or_default();
                let comment: Option<String> =
                    row.get("comment").ok().filter(|c: &String| !c.is_empty());
                let class_id: String = row.get("class_id").unwrap_or_default();
                rows.push(IndividualRow {
                    id,
                    label,
                    comment,
                    class_id,
                });
            }
        }

        Ok(rows)
    }

    async fn fetch_literal_values(
        &self,
        ontology_id: &str,
    ) -> Result<Vec<LiteralValueRow>, ExportError> {
        let query = r"
            MATCH (i:Individual {ontology_id: $ontology_id})
                  -[:HAS_VALUE]->(lv:LiteralValue)
            RETURN i.id AS individual_id, lv.property_id AS property_id,
                   lv.value AS value, lv.xsd_type AS xsd_type
        ";

        let q = neo4rs::Query::new(query.to_string()).param("ontology_id", ontology_id);

        let mut result =
            self.pool.graph().execute(q).await.map_err(|e| {
                ExportError::Database(format!("Failed to query literal values: {e}"))
            })?;

        let mut rows = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(individual_id) = row.get::<String>("individual_id") {
                let property_id: String = row.get("property_id").unwrap_or_default();
                let value: String = row.get("value").unwrap_or_default();
                let xsd_type: Option<String> =
                    row.get("xsd_type").ok().filter(|s: &String| !s.is_empty());
                rows.push(LiteralValueRow {
                    individual_id,
                    property_id,
                    value,
                    xsd_type,
                });
            }
        }

        Ok(rows)
    }

    async fn fetch_reference_values(
        &self,
        ontology_id: &str,
    ) -> Result<Vec<ReferenceValueRow>, ExportError> {
        let query = r"
            MATCH (source:Individual {ontology_id: $ontology_id})
                  -[r:HAS_REF]->(target:Individual)
            RETURN source.id AS source_id, r.property_id AS property_id,
                   target.id AS target_id
        ";

        let q = neo4rs::Query::new(query.to_string()).param("ontology_id", ontology_id);

        let mut result =
            self.pool.graph().execute(q).await.map_err(|e| {
                ExportError::Database(format!("Failed to query reference values: {e}"))
            })?;

        let mut rows = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(source_id) = row.get::<String>("source_id") {
                let property_id: String = row.get("property_id").unwrap_or_default();
                let target_id: String = row.get("target_id").unwrap_or_default();
                rows.push(ReferenceValueRow {
                    source_id,
                    property_id,
                    target_id,
                });
            }
        }

        Ok(rows)
    }
}

/// Errors that can occur during export.
#[derive(Debug, thiserror::Error)]
pub enum ExportError {
    #[error("Database error: {0}")]
    Database(String),

    #[error("Serialization error: {0}")]
    Serialization(String),

    #[error("Neo4j database is not configured")]
    Neo4jNotConfigured,

    #[error("Unsupported export format: {0}")]
    UnsupportedFormat(String),

    /// [FIX] Returned when the versioning-service call to materialize state
    /// fails (non-2xx or transport error). Prevents the exporter from
    /// silently falling back to a stale Neo4j state and shipping a snapshot
    /// that doesn't match the requested version.
    #[error("Versioning service unavailable: {0}")]
    VersioningUnavailable(String),
}
