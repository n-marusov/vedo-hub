// Test data fixtures for ontology E2E tests
// References: E2E-editor.workflow.full-cycle, E2E-versioning.branches.switch-rollback, E2E-query.sparql.execute

export interface OntologyClass {
  id: string;
  label: string;
  comment?: string;
  parents?: string[];
}

export interface OntologyProperty {
  id: string;
  label: string;
  propertyType: 'object' | 'datatype';
  domain: string[];
  range: string[];
  characteristics?: string[];
  xsdType?: string;
}

export interface OntologyIndividual {
  id: string;
  label: string;
  classId: string;
  propertyValues: Record<string, string>;
}

export interface OntologyData {
  name: string;
  classes: OntologyClass[];
  properties: OntologyProperty[];
  individuals: OntologyIndividual[];
}

export const UNIVERSITY_ONTOLOGY: OntologyData = {
  name: 'University',
  classes: [
    { id: 'Person', label: 'Person', comment: 'A person' },
    { id: 'Student', label: 'Student', parents: ['Person'] },
    { id: 'Professor', label: 'Professor', parents: ['Person'] },
    { id: 'Organization', label: 'Organization' },
  ],
  properties: [
    {
      id: 'name', label: 'name',
      propertyType: 'datatype',
      domain: ['Person'],
      range: ['string'],
      xsdType: 'string',
    },
    {
      id: 'advises', label: 'advises',
      propertyType: 'object',
      domain: ['Professor'],
      range: ['Student'],
    },
  ],
  individuals: [
    {
      id: 'John', label: 'John',
      classId: 'Professor',
      propertyValues: { name: 'John' },
    },
    {
      id: 'Alice', label: 'Alice',
      classId: 'Student',
      propertyValues: { name: 'Alice' },
    },
  ],
};
