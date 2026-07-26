// REST API client for ontology entity CRUD operations.
//
// Per ADR-DES.API.rest-graphql-mutation-boundary.md, entity writes (classes,
// properties, individuals) are REST-only. GraphQL mutations for create/update
// delete operations are forbidden because:
//   1. They would bypass Idempotency-Key header (REQ-FUN.API.write-idempotency)
//   2. They would skip the API Gateway auth middleware / audit surface
//   3. They would bypass the Circuit BreakerMiddleware DoS defenses
//
// All write endpoints already exist in `src/services/api-gateway/routes.go`
// under `/api/v1/ontologies/{id}/...` — this file wires the Vue frontend to them.

import axios from 'axios'

const ONTOLOGY_BASE = '/api/v1'

const api = axios.create({
  baseURL: ONTOLOGY_BASE,
  headers: { 'X-Requested-With': 'XMLHttpRequest' }
})

// Inject JWT token from localStorage (same pattern used by ai.ts and sparql.ts).
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('vedo-jwt-token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// ── Class CRUD ────────────────────────────────────────────────────────────────

export interface CreateClassRequest {
  ontologyId: string
  label: string
  parentId?: string
  description?: string
  annotations?: Array<{ propertyIri: string; value: string }>
}

export interface ClassResult {
  id: string
  label: string
  comment: string | null
  parents: string[]
  children: string[]
}

export async function createClass(req: CreateClassRequest): Promise<ClassResult> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'ontology.class.create.request',
      ontologyId: req.ontologyId,
      label: req.label,
      parentId: req.parentId,
      ts: new Date().toISOString()
    })
  )

  try {
    const response = await api.post<ClassResult>(
      `/ontologies/${req.ontologyId}/classes`,
      {
        label: req.label,
        parentId: req.parentId,
        description: req.description,
        annotations: req.annotations
      },
      {
        headers: { 'Idempotency-Key': crypto.randomUUID() }
      }
    )

    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'ontology.class.create.success',
        ontologyId: req.ontologyId,
        classId: response.data.id,
        label: response.data.label,
        ts: new Date().toISOString()
      })
    )

    return response.data
  } catch (err) {
    const axiosErr = err as {
      response?: { data?: { error?: { message?: string } } }
      message?: string
    }
    const serverMessage = axiosErr.response?.data?.error?.message
    const message = serverMessage ?? axiosErr.message ?? 'Failed to create class'

    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'ontology.class.create.failed',
        ontologyId: req.ontologyId,
        label: req.label,
        error: message,
        ts: new Date().toISOString()
      })
    )

    throw new Error(message)
  }
}

// ── Property CRUD ─────────────────────────────────────────────────────────────

export interface CreatePropertyRequest {
  ontologyId: string
  label: string
  propertyType: string
  domain?: string
  range?: string
  description?: string
}

export interface PropertyResult {
  id: string
  label: string
  propertyType: string
  domains: string[]
  ranges: string[]
}

export async function createProperty(req: CreatePropertyRequest): Promise<PropertyResult> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'ontology.property.create.request',
      ontologyId: req.ontologyId,
      label: req.label,
      propertyType: req.propertyType,
      ts: new Date().toISOString()
    })
  )

  try {
    const response = await api.post<PropertyResult>(
      `/ontologies/${req.ontologyId}/properties`,
      {
        label: req.label,
        propertyType: req.propertyType,
        domain: req.domain,
        range: req.range,
        description: req.description
      },
      {
        headers: { 'Idempotency-Key': crypto.randomUUID() }
      }
    )

    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'ontology.property.create.success',
        ontologyId: req.ontologyId,
        propertyId: response.data.id,
        label: response.data.label,
        ts: new Date().toISOString()
      })
    )

    return response.data
  } catch (err) {
    const axiosErr = err as {
      response?: { data?: { error?: { message?: string } } }
      message?: string
    }
    const serverMessage = axiosErr.response?.data?.error?.message
    const message = serverMessage ?? axiosErr.message ?? 'Failed to create property'

    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'ontology.property.create.failed',
        ontologyId: req.ontologyId,
        label: req.label,
        error: message,
        ts: new Date().toISOString()
      })
    )

    throw new Error(message)
  }
}

// ── Individual CRUD ────────────────────────────────────────────────────────────

export interface PropertyValueInput {
  propertyId: string
  value: string
  xsdType?: string
}

export interface CreateIndividualRequest {
  ontologyId: string
  label: string
  classId: string
  propertyValues?: PropertyValueInput[]
}

export interface IndividualResult {
  id: string
  label: string
  classId: string
  classLabel: string
}

export async function createIndividual(req: CreateIndividualRequest): Promise<IndividualResult> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'ontology.individual.create.request',
      ontologyId: req.ontologyId,
      label: req.label,
      classId: req.classId,
      ts: new Date().toISOString()
    })
  )

  try {
    const response = await api.post<IndividualResult>(
      `/ontologies/${req.ontologyId}/individuals`,
      {
        label: req.label,
        classId: req.classId,
        propertyValues: req.propertyValues
      },
      {
        headers: { 'Idempotency-Key': crypto.randomUUID() }
      }
    )

    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'ontology.individual.create.success',
        ontologyId: req.ontologyId,
        individualId: response.data.id,
        label: response.data.label,
        ts: new Date().toISOString()
      })
    )

    return response.data
  } catch (err) {
    const axiosErr = err as {
      response?: { data?: { error?: { message?: string } } }
      message?: string
    }
    const serverMessage = axiosErr.response?.data?.error?.message
    const message = serverMessage ?? axiosErr.message ?? 'Failed to create individual'

    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'ontology.individual.create.failed',
        ontologyId: req.ontologyId,
        label: req.label,
        error: message,
        ts: new Date().toISOString()
      })
    )

    throw new Error(message)
  }
}
