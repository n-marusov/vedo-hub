/* eslint-disable */
import * as types from './graphql';
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 * Learn more about it here: https://the-guild.dev/graphql/codegen/plugins/presets/preset-client#reducing-bundle-size
 */
type Documents = {
    "\n  fragment ClassSummaryFields on ClassSummary {\n    id\n    label\n    comment\n    parents\n  }\n": typeof types.ClassSummaryFieldsFragmentDoc,
    "\n  fragment ClassFields on Class {\n    id\n    label\n    comment\n    parents\n    children\n  }\n": typeof types.ClassFieldsFragmentDoc,
    "\n  fragment PropertyFields on Property {\n    id\n    label\n    comment\n    propertyType\n    domains\n    ranges\n    xsdType\n    characteristics {\n      functional\n      inverseFunctional\n      transitive\n      symmetric\n    }\n    annotations {\n      propertyIri\n      value\n    }\n  }\n": typeof types.PropertyFieldsFragmentDoc,
    "\n  fragment IndividualFields on Individual {\n    id\n    label\n    comment\n    classId\n    classLabel\n    literalValues {\n      propertyId\n      propertyLabel\n      value\n      xsdType\n      valueId\n    }\n    referenceValues {\n      propertyId\n      propertyLabel\n      targetId\n      targetLabel\n      edgeId\n    }\n  }\n": typeof types.IndividualFieldsFragmentDoc,
    "\n  query GetClass($ontologyId: String!, $classId: String!) {\n    class(ontologyId: $ontologyId, classId: $classId) {\n      ...ClassFields\n    }\n  }\n  \n": typeof types.GetClassDocument,
    "\n  query ListClasses($ontologyId: String!, $q: String, $page: Int, $perPage: Int) {\n    classes(ontologyId: $ontologyId, q: $q, page: $page, perPage: $perPage) {\n      items {\n        ...ClassSummaryFields\n      }\n      total\n      page\n      perPage\n    }\n  }\n  \n": typeof types.ListClassesDocument,
    "\n  query ClassTree($ontologyId: String!) {\n    classTree(ontologyId: $ontologyId) {\n      id\n      label\n      children {\n        id\n        label\n        children {\n          id\n          label\n        }\n      }\n    }\n  }\n": typeof types.ClassTreeDocument,
    "\n  query ClassAncestors($ontologyId: String!, $classId: String!) {\n    classAncestors(ontologyId: $ontologyId, classId: $classId) {\n      id\n      label\n    }\n  }\n": typeof types.ClassAncestorsDocument,
    "\n  query ClassDescendants($ontologyId: String!, $classId: String!, $maxDepth: Int) {\n    classDescendants(\n      ontologyId: $ontologyId\n      classId: $classId\n      maxDepth: $maxDepth\n    ) {\n      id\n      label\n      children {\n        id\n        label\n      }\n    }\n  }\n": typeof types.ClassDescendantsDocument,
    "\n  query GraphNeighborhood(\n    $ontologyId: String!\n    $classId: String!\n    $depth: Int\n  ) {\n    graphNeighborhood(\n      ontologyId: $ontologyId\n      classId: $classId\n      depth: $depth\n    ) {\n      nodes {\n        id\n        label\n      }\n      edges {\n        sourceId\n        targetId\n        propertyId\n        propertyLabel\n      }\n    }\n  }\n": typeof types.GraphNeighborhoodDocument,
    "\n  query AutocompleteClasses($ontologyId: String!, $q: String!, $limit: Int) {\n    autocompleteClasses(ontologyId: $ontologyId, q: $q, limit: $limit) {\n      ...ClassSummaryFields\n    }\n  }\n  \n": typeof types.AutocompleteClassesDocument,
    "\n  query GetProperty($ontologyId: String!, $propertyId: String!) {\n    property(ontologyId: $ontologyId, propertyId: $propertyId) {\n      ...PropertyFields\n    }\n  }\n  \n": typeof types.GetPropertyDocument,
    "\n  query ListProperties(\n    $ontologyId: String!\n    $q: String\n    $propertyType: PropertyType\n    $page: Int\n    $perPage: Int\n  ) {\n    properties(\n      ontologyId: $ontologyId\n      q: $q\n      propertyType: $propertyType\n      page: $page\n      perPage: $perPage\n    ) {\n      items {\n        id\n        label\n        propertyType\n        xsdType\n        domains\n      }\n      total\n      page\n      perPage\n    }\n  }\n": typeof types.ListPropertiesDocument,
    "\n  query GetIndividual($ontologyId: String!, $individualId: String!) {\n    individual(ontologyId: $ontologyId, individualId: $individualId) {\n      ...IndividualFields\n    }\n  }\n  \n": typeof types.GetIndividualDocument,
    "\n  query ListIndividuals(\n    $ontologyId: String!\n    $classId: String!\n    $q: String\n    $page: Int\n    $perPage: Int\n  ) {\n    individuals(\n      ontologyId: $ontologyId\n      classId: $classId\n      q: $q\n      page: $page\n      perPage: $perPage\n    ) {\n      items {\n        ...IndividualFields\n      }\n      total\n      page\n      perPage\n    }\n  }\n  \n": typeof types.ListIndividualsDocument,
};
const documents: Documents = {
    "\n  fragment ClassSummaryFields on ClassSummary {\n    id\n    label\n    comment\n    parents\n  }\n": types.ClassSummaryFieldsFragmentDoc,
    "\n  fragment ClassFields on Class {\n    id\n    label\n    comment\n    parents\n    children\n  }\n": types.ClassFieldsFragmentDoc,
    "\n  fragment PropertyFields on Property {\n    id\n    label\n    comment\n    propertyType\n    domains\n    ranges\n    xsdType\n    characteristics {\n      functional\n      inverseFunctional\n      transitive\n      symmetric\n    }\n    annotations {\n      propertyIri\n      value\n    }\n  }\n": types.PropertyFieldsFragmentDoc,
    "\n  fragment IndividualFields on Individual {\n    id\n    label\n    comment\n    classId\n    classLabel\n    literalValues {\n      propertyId\n      propertyLabel\n      value\n      xsdType\n      valueId\n    }\n    referenceValues {\n      propertyId\n      propertyLabel\n      targetId\n      targetLabel\n      edgeId\n    }\n  }\n": types.IndividualFieldsFragmentDoc,
    "\n  query GetClass($ontologyId: String!, $classId: String!) {\n    class(ontologyId: $ontologyId, classId: $classId) {\n      ...ClassFields\n    }\n  }\n  \n": types.GetClassDocument,
    "\n  query ListClasses($ontologyId: String!, $q: String, $page: Int, $perPage: Int) {\n    classes(ontologyId: $ontologyId, q: $q, page: $page, perPage: $perPage) {\n      items {\n        ...ClassSummaryFields\n      }\n      total\n      page\n      perPage\n    }\n  }\n  \n": types.ListClassesDocument,
    "\n  query ClassTree($ontologyId: String!) {\n    classTree(ontologyId: $ontologyId) {\n      id\n      label\n      children {\n        id\n        label\n        children {\n          id\n          label\n        }\n      }\n    }\n  }\n": types.ClassTreeDocument,
    "\n  query ClassAncestors($ontologyId: String!, $classId: String!) {\n    classAncestors(ontologyId: $ontologyId, classId: $classId) {\n      id\n      label\n    }\n  }\n": types.ClassAncestorsDocument,
    "\n  query ClassDescendants($ontologyId: String!, $classId: String!, $maxDepth: Int) {\n    classDescendants(\n      ontologyId: $ontologyId\n      classId: $classId\n      maxDepth: $maxDepth\n    ) {\n      id\n      label\n      children {\n        id\n        label\n      }\n    }\n  }\n": types.ClassDescendantsDocument,
    "\n  query GraphNeighborhood(\n    $ontologyId: String!\n    $classId: String!\n    $depth: Int\n  ) {\n    graphNeighborhood(\n      ontologyId: $ontologyId\n      classId: $classId\n      depth: $depth\n    ) {\n      nodes {\n        id\n        label\n      }\n      edges {\n        sourceId\n        targetId\n        propertyId\n        propertyLabel\n      }\n    }\n  }\n": types.GraphNeighborhoodDocument,
    "\n  query AutocompleteClasses($ontologyId: String!, $q: String!, $limit: Int) {\n    autocompleteClasses(ontologyId: $ontologyId, q: $q, limit: $limit) {\n      ...ClassSummaryFields\n    }\n  }\n  \n": types.AutocompleteClassesDocument,
    "\n  query GetProperty($ontologyId: String!, $propertyId: String!) {\n    property(ontologyId: $ontologyId, propertyId: $propertyId) {\n      ...PropertyFields\n    }\n  }\n  \n": types.GetPropertyDocument,
    "\n  query ListProperties(\n    $ontologyId: String!\n    $q: String\n    $propertyType: PropertyType\n    $page: Int\n    $perPage: Int\n  ) {\n    properties(\n      ontologyId: $ontologyId\n      q: $q\n      propertyType: $propertyType\n      page: $page\n      perPage: $perPage\n    ) {\n      items {\n        id\n        label\n        propertyType\n        xsdType\n        domains\n      }\n      total\n      page\n      perPage\n    }\n  }\n": types.ListPropertiesDocument,
    "\n  query GetIndividual($ontologyId: String!, $individualId: String!) {\n    individual(ontologyId: $ontologyId, individualId: $individualId) {\n      ...IndividualFields\n    }\n  }\n  \n": types.GetIndividualDocument,
    "\n  query ListIndividuals(\n    $ontologyId: String!\n    $classId: String!\n    $q: String\n    $page: Int\n    $perPage: Int\n  ) {\n    individuals(\n      ontologyId: $ontologyId\n      classId: $classId\n      q: $q\n      page: $page\n      perPage: $perPage\n    ) {\n      items {\n        ...IndividualFields\n      }\n      total\n      page\n      perPage\n    }\n  }\n  \n": types.ListIndividualsDocument,
};

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = graphql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function graphql(source: string): unknown;

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  fragment ClassSummaryFields on ClassSummary {\n    id\n    label\n    comment\n    parents\n  }\n"): (typeof documents)["\n  fragment ClassSummaryFields on ClassSummary {\n    id\n    label\n    comment\n    parents\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  fragment ClassFields on Class {\n    id\n    label\n    comment\n    parents\n    children\n  }\n"): (typeof documents)["\n  fragment ClassFields on Class {\n    id\n    label\n    comment\n    parents\n    children\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  fragment PropertyFields on Property {\n    id\n    label\n    comment\n    propertyType\n    domains\n    ranges\n    xsdType\n    characteristics {\n      functional\n      inverseFunctional\n      transitive\n      symmetric\n    }\n    annotations {\n      propertyIri\n      value\n    }\n  }\n"): (typeof documents)["\n  fragment PropertyFields on Property {\n    id\n    label\n    comment\n    propertyType\n    domains\n    ranges\n    xsdType\n    characteristics {\n      functional\n      inverseFunctional\n      transitive\n      symmetric\n    }\n    annotations {\n      propertyIri\n      value\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  fragment IndividualFields on Individual {\n    id\n    label\n    comment\n    classId\n    classLabel\n    literalValues {\n      propertyId\n      propertyLabel\n      value\n      xsdType\n      valueId\n    }\n    referenceValues {\n      propertyId\n      propertyLabel\n      targetId\n      targetLabel\n      edgeId\n    }\n  }\n"): (typeof documents)["\n  fragment IndividualFields on Individual {\n    id\n    label\n    comment\n    classId\n    classLabel\n    literalValues {\n      propertyId\n      propertyLabel\n      value\n      xsdType\n      valueId\n    }\n    referenceValues {\n      propertyId\n      propertyLabel\n      targetId\n      targetLabel\n      edgeId\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GetClass($ontologyId: String!, $classId: String!) {\n    class(ontologyId: $ontologyId, classId: $classId) {\n      ...ClassFields\n    }\n  }\n  \n"): (typeof documents)["\n  query GetClass($ontologyId: String!, $classId: String!) {\n    class(ontologyId: $ontologyId, classId: $classId) {\n      ...ClassFields\n    }\n  }\n  \n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ListClasses($ontologyId: String!, $q: String, $page: Int, $perPage: Int) {\n    classes(ontologyId: $ontologyId, q: $q, page: $page, perPage: $perPage) {\n      items {\n        ...ClassSummaryFields\n      }\n      total\n      page\n      perPage\n    }\n  }\n  \n"): (typeof documents)["\n  query ListClasses($ontologyId: String!, $q: String, $page: Int, $perPage: Int) {\n    classes(ontologyId: $ontologyId, q: $q, page: $page, perPage: $perPage) {\n      items {\n        ...ClassSummaryFields\n      }\n      total\n      page\n      perPage\n    }\n  }\n  \n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ClassTree($ontologyId: String!) {\n    classTree(ontologyId: $ontologyId) {\n      id\n      label\n      children {\n        id\n        label\n        children {\n          id\n          label\n        }\n      }\n    }\n  }\n"): (typeof documents)["\n  query ClassTree($ontologyId: String!) {\n    classTree(ontologyId: $ontologyId) {\n      id\n      label\n      children {\n        id\n        label\n        children {\n          id\n          label\n        }\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ClassAncestors($ontologyId: String!, $classId: String!) {\n    classAncestors(ontologyId: $ontologyId, classId: $classId) {\n      id\n      label\n    }\n  }\n"): (typeof documents)["\n  query ClassAncestors($ontologyId: String!, $classId: String!) {\n    classAncestors(ontologyId: $ontologyId, classId: $classId) {\n      id\n      label\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ClassDescendants($ontologyId: String!, $classId: String!, $maxDepth: Int) {\n    classDescendants(\n      ontologyId: $ontologyId\n      classId: $classId\n      maxDepth: $maxDepth\n    ) {\n      id\n      label\n      children {\n        id\n        label\n      }\n    }\n  }\n"): (typeof documents)["\n  query ClassDescendants($ontologyId: String!, $classId: String!, $maxDepth: Int) {\n    classDescendants(\n      ontologyId: $ontologyId\n      classId: $classId\n      maxDepth: $maxDepth\n    ) {\n      id\n      label\n      children {\n        id\n        label\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GraphNeighborhood(\n    $ontologyId: String!\n    $classId: String!\n    $depth: Int\n  ) {\n    graphNeighborhood(\n      ontologyId: $ontologyId\n      classId: $classId\n      depth: $depth\n    ) {\n      nodes {\n        id\n        label\n      }\n      edges {\n        sourceId\n        targetId\n        propertyId\n        propertyLabel\n      }\n    }\n  }\n"): (typeof documents)["\n  query GraphNeighborhood(\n    $ontologyId: String!\n    $classId: String!\n    $depth: Int\n  ) {\n    graphNeighborhood(\n      ontologyId: $ontologyId\n      classId: $classId\n      depth: $depth\n    ) {\n      nodes {\n        id\n        label\n      }\n      edges {\n        sourceId\n        targetId\n        propertyId\n        propertyLabel\n      }\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query AutocompleteClasses($ontologyId: String!, $q: String!, $limit: Int) {\n    autocompleteClasses(ontologyId: $ontologyId, q: $q, limit: $limit) {\n      ...ClassSummaryFields\n    }\n  }\n  \n"): (typeof documents)["\n  query AutocompleteClasses($ontologyId: String!, $q: String!, $limit: Int) {\n    autocompleteClasses(ontologyId: $ontologyId, q: $q, limit: $limit) {\n      ...ClassSummaryFields\n    }\n  }\n  \n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GetProperty($ontologyId: String!, $propertyId: String!) {\n    property(ontologyId: $ontologyId, propertyId: $propertyId) {\n      ...PropertyFields\n    }\n  }\n  \n"): (typeof documents)["\n  query GetProperty($ontologyId: String!, $propertyId: String!) {\n    property(ontologyId: $ontologyId, propertyId: $propertyId) {\n      ...PropertyFields\n    }\n  }\n  \n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ListProperties(\n    $ontologyId: String!\n    $q: String\n    $propertyType: PropertyType\n    $page: Int\n    $perPage: Int\n  ) {\n    properties(\n      ontologyId: $ontologyId\n      q: $q\n      propertyType: $propertyType\n      page: $page\n      perPage: $perPage\n    ) {\n      items {\n        id\n        label\n        propertyType\n        xsdType\n        domains\n      }\n      total\n      page\n      perPage\n    }\n  }\n"): (typeof documents)["\n  query ListProperties(\n    $ontologyId: String!\n    $q: String\n    $propertyType: PropertyType\n    $page: Int\n    $perPage: Int\n  ) {\n    properties(\n      ontologyId: $ontologyId\n      q: $q\n      propertyType: $propertyType\n      page: $page\n      perPage: $perPage\n    ) {\n      items {\n        id\n        label\n        propertyType\n        xsdType\n        domains\n      }\n      total\n      page\n      perPage\n    }\n  }\n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query GetIndividual($ontologyId: String!, $individualId: String!) {\n    individual(ontologyId: $ontologyId, individualId: $individualId) {\n      ...IndividualFields\n    }\n  }\n  \n"): (typeof documents)["\n  query GetIndividual($ontologyId: String!, $individualId: String!) {\n    individual(ontologyId: $ontologyId, individualId: $individualId) {\n      ...IndividualFields\n    }\n  }\n  \n"];
/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "\n  query ListIndividuals(\n    $ontologyId: String!\n    $classId: String!\n    $q: String\n    $page: Int\n    $perPage: Int\n  ) {\n    individuals(\n      ontologyId: $ontologyId\n      classId: $classId\n      q: $q\n      page: $page\n      perPage: $perPage\n    ) {\n      items {\n        ...IndividualFields\n      }\n      total\n      page\n      perPage\n    }\n  }\n  \n"): (typeof documents)["\n  query ListIndividuals(\n    $ontologyId: String!\n    $classId: String!\n    $q: String\n    $page: Int\n    $perPage: Int\n  ) {\n    individuals(\n      ontologyId: $ontologyId\n      classId: $classId\n      q: $q\n      page: $page\n      perPage: $perPage\n    ) {\n      items {\n        ...IndividualFields\n      }\n      total\n      page\n      perPage\n    }\n  }\n  \n"];

export function graphql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;