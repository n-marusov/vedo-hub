## Test Cases: Ontology Core Coverage

---

### Group A — TBox/ABox Editor (ontology-service)

#### TC-A01: Create class successfully

**Priority:** High
**Type:** Positive

**Precondition:** Neo4j is accessible, test ontology exists

**Steps:**
1. Send POST to `/api/v1/ontologies/{oid}/classes` with valid class JSON
2. Verify HTTP 200 response
3. Query Neo4j directly: `MATCH (c:Class {id: "Person"}) RETURN c`
4. Verify node exists with correct label and properties

**Expected result:**
Class node created in Neo4j with correct id, label, and ontology_id. Response includes the full class object.

**Test data:**
```json
{
  "id": "Person",
  "label": "Person",
  "comment": "A person entity"
}
```

**Validates:** `REQ-USR.UI.tbox-editor`, `REQ-FUN.API.class-hierarchy-accuracy`

---

#### TC-A02: Create class with duplicate IRI returns error

**Priority:** High
**Type:** Negative

**Precondition:** Class "Person" already exists in ontology

**Steps:**
1. Send POST to `/api/v1/ontologies/{oid}/classes` with id "Person" again
2. Verify HTTP 409 Conflict
3. Verify error code contains "IRI_COLLISION" or equivalent

**Expected result:**
Duplicate IRI rejected with meaningful error message. No duplicate node created in Neo4j.

**Test data:**
```json
{"id": "Person", "label": "Person Duplicate"}
```

**Validates:** `REQ-FUN.API.class-hierarchy-accuracy`, `REQ-FUN.API.owl-unique-iri`

---

#### TC-A03: Create class hierarchy with subClassOf

**Priority:** High
**Type:** Positive

**Precondition:** Test ontology exists

**Steps:**
1. Create class "Person"
2. Create class "Student" with `subClassOf: ["Person"]`
3. Query Neo4j: `MATCH (s:Student)-[:SUB_CLASS_OF]->(p:Person) RETURN s, p`
4. Verify relationship exists

**Expected result:**
Subclass relationship correctly stored in Neo4j. Class tree displays Student under Person.

**Test data:**
```json
{"id": "Student", "label": "Student", "subClassOf": ["Person"]}
```

**Validates:** `REQ-FUN.API.class-hierarchy-accuracy`

---

#### TC-A04: Detect cyclic class hierarchy

**Priority:** High
**Type:** Negative

**Precondition:** Classes A and B exist with A subClassOf B

**Steps:**
1. Send POST to update B with `subClassOf: ["A"]`
2. Verify HTTP 400 or 422
3. Verify error code "CYCLE_DETECTED" or equivalent
4. Query Neo4j to verify no cycle was created

**Expected result:**
Cyclic hierarchy rejected. Error message identifies the cycle path (A → B → A).

**Validates:** `REQ-FUN.API.no-cyclic-hierarchy`, `REQ-FUN.API.pre-save-validation`

---

#### TC-A05: Edit existing class properties

**Priority:** High
**Type:** Positive

**Precondition:** Class "Person" exists with label "Person"

**Steps:**
1. Send PUT to `/api/v1/ontologies/{oid}/classes/Person` with updated label
2. Verify HTTP 200
3. Query Neo4j for the updated class
4. Verify label changed to "Human"

**Expected result:**
Class label updated. Other properties unchanged. Versioning draft reflects the change.

**Test data:**
```json
{"label": "Human", "comment": "A human entity"}
```

**Validates:** `REQ-USR.UI.tbox-editor`

---

#### TC-A06: Delete class with existing subclasses

**Priority:** High
**Type:** Negative

**Precondition:** Class "Person" has subclass "Student"

**Steps:**
1. Send DELETE to `/api/v1/ontologies/{oid}/classes/Person`
2. Verify HTTP 409 or warning about existing subclasses
3. Verify class Person still exists in Neo4j

**Expected result:**
Deletion blocked when subclasses exist. Error message references dependent subclasses.

**Validates:** `REQ-USR.UI.tbox-editor`

---

#### TC-A07: Create datatype property with domain/range

**Priority:** High
**Type:** Positive

**Precondition:** Test ontology exists

**Steps:**
1. Send POST to `/api/v1/ontologies/{oid}/properties` with datatype property
2. Verify HTTP 200
3. Query Neo4j for the property node
4. Verify domain, range, and datatype constraints stored correctly

**Expected result:**
Property created with correct domain (class), range (xsd:string), and datatype constraints.

**Test data:**
```json
{
  "id": "hasName",
  "label": "has name",
  "domain": ["Person"],
  "range": "xsd:string",
  "type": "DatatypeProperty"
}
```

**Validates:** `REQ-FUN.API.properties-accuracy`, `REQ-USR.UI.tbox-editor`

---

#### TC-A08: Create object property with domain/range

**Priority:** High
**Type:** Positive

**Precondition:** Classes "Person" and "Organization" exist

**Steps:**
1. Send POST to create object property "worksFor" with domain Person, range Organization
2. Verify HTTP 200
3. Query Neo4j for property node
4. Verify domain and range associations

**Expected result:**
Object property created linking Person → Organization.

**Validates:** `REQ-USR.UI.tbox-editor`, `REQ-FUN.API.properties-accuracy`

---

#### TC-A09: Create individual with property values

**Priority:** High
**Type:** Positive

**Precondition:** Class "Person" and property "hasName" exist

**Steps:**
1. Send POST to `/api/v1/ontologies/{oid}/individuals` with individual JSON
2. Verify HTTP 200
3. Query Neo4j: `MATCH (i:Individual {id: "alice"}) RETURN i`
4. Verify individual exists with correct class and property values

**Expected result:**
Individual created, linked to class Person, with property values stored.

**Test data:**
```json
{
  "id": "alice",
  "type": ["Person"],
  "properties": {"hasName": "Alice"}
}
```

**Validates:** `REQ-USR.UI.abox-editor`

---

#### TC-A10: Validate class against SHACL shape

**Priority:** Medium
**Type:** Positive

**Precondition:** SHACL shape exists requiring Person to have "hasName" property

**Steps:**
1. Create class "Person" without "hasName" property
2. Send POST to validate against SHACL shape
3. Verify violation report indicates missing required property

**Expected result:**
SHACL violation detected and reported with shape ID, severity, and path.

**Validates:** `REQ-FUN.API.pre-save-validation`, `REQ-USR.UI.tbox-editor`

---

### Group B — Versioning (versioning-service)

#### TC-B01: Create commit with changes

**Priority:** High
**Type:** Positive

**Precondition:** Ontology has staged changes (class created, not committed)

**Steps:**
1. Send POST to `/api/v1/ontologies/{oid}/commits` with commit message
2. Verify HTTP 200
3. Query commit history to verify new commit appears
4. Verify commit contains expected changeset (class added)

**Expected result:**
Commit created with SHA, message, timestamp, author. Changeset includes added class.

**Test data:**
```json
{"message": "Add Person class", "author": "alice"}
```

**Validates:** `REQ-FUN.DATA.versioning`

---

#### TC-B02: Branch operations — create, switch, list

**Priority:** High
**Type:** Positive

**Precondition:** Ontology exists on main branch

**Steps:**
1. Create branch "feature/new-classes" from main
2. Switch to "feature/new-classes"
3. Create a class on the new branch
4. Switch back to main
5. Verify class does NOT exist on main (branch isolation)
6. List branches → verify both "main" and "feature/new-classes" appear

**Expected result:**
Branch isolation works — changes on one branch don't affect another.
Branch list includes all branches.

**Validates:** `REQ-FUN.DATA.versioning`

---

#### TC-B03: Merge branches

**Priority:** High
**Type:** Positive

**Precondition:** "feature/new-classes" branch has 1 commit ahead of main

**Steps:**
1. Create merge request from "feature/new-classes" → main
2. Approve and merge
3. Switch to main
4. Verify the class from the feature branch now exists on main
5. Verify merge commit appears in history

**Expected result:**
Changes merged correctly. Merge commit shows both parents.

**Validates:** `REQ-FUN.DATA.versioning`

---

#### TC-B04: Merge conflict detection

**Priority:** Medium
**Type:** Negative

**Precondition:** Class "Person" edited differently on two branches

**Steps:**
1. On branch A: change Person label to "Human"
2. On branch B: change Person label to "Individual"
3. Attempt to merge B into A
4. Verify conflict detected
5. Verify conflict details show the conflicting property and values

**Expected result:**
Merge blocked with conflict report. User can resolve by choosing one value.

**Validates:** `REQ-FUN.DATA.versioning`

---

#### TC-B05: Diff between commits

**Priority:** High
**Type:** Positive

**Precondition:** Two commits exist with different changes

**Steps:**
1. GET `/api/v1/ontologies/{oid}/diff?from={sha1}&to={sha2}`
2. Verify HTTP 200
3. Verify diff contains added/removed/modified entities
4. Verify diff format is readable (added → green, removed → red)

**Expected result:**
Diff shows all changes between two commits: classes added, properties modified, individuals removed.

**Validates:** `REQ-FUN.DATA.versioning`

---

#### TC-B06: Rollback to previous commit

**Priority:** Medium
**Type:** Positive

**Precondition:** Ontology has 3 commits, current state has class "Project"

**Steps:**
1. Rollback to commit #2 (before "Project" was added)
2. Verify class "Project" no longer exists in Neo4j
3. Verify commit history shows rollback as a new commit

**Expected result:**
Ontology state restored to target commit. Rollback creates a revert commit in history.

**Validates:** `REQ-FUN.DATA.versioning`

---

### Group C — Import/Export

#### TC-C01: Export ontology as canonical Turtle

**Priority:** High
**Type:** Positive

**Precondition:** Ontology with 3 classes + 2 properties + 1 individual exists

**Steps:**
1. Send GET to `/api/v1/ontologies/{oid}/export?format=turtle`
2. Verify HTTP 200
3. Verify Content-Type is `text/turtle`
4. Verify Turtle syntax is valid (parse with `riot --validate`)
5. Verify all entities appear in the Turtle output

**Expected result:**
Valid canonical Turtle file containing all ontology entities with proper namespaces.

**Validates:** `REQ-USR.UI.tbox-editor`, `REQ-FUN.API.integration`

---

#### TC-C02: Export → Import round-trip

**Priority:** Medium
**Type:** Positive

**Precondition:** Ontology A with known entity count exists

**Steps:**
1. Export ontology A as Turtle → save to file
2. Create new empty ontology B
3. Import Turtle file into ontology B
4. Compare entity counts between A and B
5. Compare entity structure (class hierarchy, property domains/ranges)

**Expected result:**
Round-trip preserves all entities, relationships, and annotations.
Entity count in B matches A exactly.

**Validates:** `REQ-FUN.API.integration`, `REQ-FUN.DATA.ontology-identifier-standard`

---

#### TC-C03: Publish ontology snapshot

**Priority:** Medium
**Type:** Positive

**Precondition:** Ontology exists with committed changes

**Steps:**
1. POST to `/api/v1/ontologies/{oid}/publish` with snapshot config
2. Verify HTTP 200
3. Verify published snapshot appears in public browse API
4. Verify unauthenticated GET to public URL returns ontology

**Expected result:**
Snapshot published with version tag. Public API serves the ontology in read-only mode.

**Validates:** `REQ-FUN.API.integration`, `REQ-USR.UI.tbox-editor`

---

#### TC-C04: Import XLSX with valid structure

**Priority:** Medium
**Type:** Positive

**Precondition:** XLSX file with classes in sheet 1, properties in sheet 2

**Steps:**
1. POST file to `/api/v1/ontologies/{oid}/import/xlsx`
2. Verify HTTP 200
3. Verify classes from sheet 1 appear in ontology
4. Verify properties from sheet 2 appear in ontology

**Expected result:**
XLSX imported successfully. Entity count matches file content.

**Validates:** `REQ-FUN.API.excel-import-block`, `REQ-FUN.DATA.ontology-identifier-standard`

---

### Group D — REST API Integration

#### TC-D01: TBox CRUD via API Gateway

**Priority:** Medium
**Type:** Positive

**Precondition:** API Gateway running, ontology service accessible

**Steps:**
1. POST `/api/v1/ontologies/{oid}/classes` → create ClassA
2. GET `/api/v1/ontologies/{oid}/classes/ClassA` → verify response
3. PUT `/api/v1/ontologies/{oid}/classes/ClassA` → update label
4. DELETE `/api/v1/ontologies/{oid}/classes/ClassA` → remove
5. Verify 404 on subsequent GET

**Expected result:**
Full CRUD lifecycle works through API Gateway. Each operation returns appropriate HTTP status.

**Validates:** `REQ-FUN.API.protocol-stack`, `REQ-USR.UI.tbox-editor`

---

#### TC-D02: Versioning via REST endpoints

**Priority:** Medium
**Type:** Positive

**Precondition:** Ontology exists with at least one committed change

**Steps:**
1. GET `/api/v1/ontologies/{oid}/commits` → list commits
2. GET `/api/v1/ontologies/{oid}/branches` → list branches
3. POST `/api/v1/ontologies/{oid}/branches` → create branch
4. Verify all endpoints return 200 with valid JSON

**Expected result:**
Versioning API responds correctly. JSON schemas match expected contract.

**Validates:** `REQ-FUN.API.rest-versioning`, `REQ-FUN.DATA.versioning`

---

### TC-D03: REST versioning backward compatibility

**Priority:** Medium
**Type:** Negative

**Precondition:** API Gateway routes exist for v1 endpoints

**Steps:**
1. GET `/api/v1/ontologies/{oid}/commits` with Accept header for v1
2. Verify response format matches v1 schema
3. Verify no breaking changes introduced

**Expected result:**
Backward-compatible REST API. Existing clients continue to work.

**Validates:** `REQ-FUN.API.rest-versioning`
