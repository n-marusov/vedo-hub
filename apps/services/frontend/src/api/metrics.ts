// REST API client for metrics service (ontology KPIs and trends).
//
// After GraphQL tightening: metrics migrated from GraphQL
// (ONTOLOGY_METRICS_QUERY) to REST via API Gateway proxy.

import axios from 'axios'

const BASE = '/api/v1'

const api = axios.create({
  baseURL: BASE,
  headers: { 'X-Requested-With': 'XMLHttpRequest' }
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('vedo-jwt-token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// ── Types ──────────────────────────────────────────────────────────────────────────

export interface MetricsCounters {
  classCount: number
  propertyCount: number
  individualCount: number
  axiomCount: number
  commentCount: number
  mergeRequestCount: number
}

export interface MetricsTrendPoint {
  date: string
  classCount: number
  propertyCount: number
  individualCount: number
}

export interface OntologyMetrics {
  counters: MetricsCounters
  trends: MetricsTrendPoint[]
}

// ── Get Ontology Metrics ───────────────────────────────────────────────────────────

export async function getOntologyMetrics(ontologyId: string): Promise<OntologyMetrics> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'metrics.ontology.get.request',
      ontologyId,
      ts: new Date().toISOString()
    })
  )

  try {
    const { data } = await api.get(`/metrics/ontologies/${ontologyId}`)
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'metrics.ontology.get.success',
        ontologyId,
        ts: new Date().toISOString()
      })
    )
    return data
  } catch (err: any) {
    const msg =
      err.response?.data?.error?.message ?? err.message ?? 'Failed to get ontology metrics'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'metrics.ontology.get.failed',
        ontologyId,
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}

// ── List All Metrics ───────────────────────────────────────────────────────────────

export async function listAllMetrics(): Promise<OntologyMetrics[]> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'metrics.list.request',
      ts: new Date().toISOString()
    })
  )

  try {
    const { data } = await api.get('/metrics/ontologies')
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'metrics.list.success',
        count: data?.length ?? 0,
        ts: new Date().toISOString()
      })
    )
    return data ?? []
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? 'Failed to list metrics'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'metrics.list.failed',
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}
