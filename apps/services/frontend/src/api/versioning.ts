// REST API client for versioning service (commits, branches, tags, diffs).
//
// After GraphQL tightening: versioning reads migrated from GraphQL
// (commits, branch, branches, tags, compareRevisions) to REST.

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

export interface CommitSummary {
  id: string
  branchId: string
  parentCommitId: string | null
  message: string
  authorId: string
  authorName: string
  totalChanges: number
  createdAt: string
}

export interface BranchInfo {
  id: string
  name: string
  ontologyId: string
  headCommitId: string | null
  createdAt: string
  isProtected: boolean
  lastCommitMessage: string | null
  lastCommitAuthor: string | null
  aheadCount: number
  behindCount: number
}

export interface TagInfo {
  id: string
  name: string
  commitId: string
  message: string
  authorName: string
  createdAt: string
}

export interface RevisionChange {
  entityId: string
  entityType: string
  entityLabel: string
  changeType: string
  field: string
  oldValue: string | null
  newValue: string | null
}

export interface CompareRevisionsResult {
  additions: number
  deletions: number
  changes: RevisionChange[]
}

// ── Commit History ─────────────────────────────────────────────────────────────────

export async function listCommits(
  ontologyId: string,
  branchId?: string,
  page = 0,
  perPage = 20
): Promise<{ items: CommitSummary[]; total: number; page: number; perPage: number }> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'versioning.commits.list.request',
      ontologyId,
      branchId,
      page,
      perPage,
      ts: new Date().toISOString()
    })
  )

  try {
    const params = new URLSearchParams()
    if (branchId) params.set('branch_id', branchId)
    params.set('page', String(page))
    params.set('per_page', String(perPage))
    const { data } = await api.get(`/versioning/commits?${params}`)
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'versioning.commits.list.success',
        ontologyId,
        total: data.total,
        ts: new Date().toISOString()
      })
    )
    return data
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? 'Failed to list commits'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'versioning.commits.list.failed',
        ontologyId,
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}

// ── Branches ────────────────────────────────────────────────────────────────────────

export async function listBranches(
  ontologyId: string,
  referenceBranchId?: string
): Promise<{ items: BranchInfo[]; total: number }> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'versioning.branches.list.request',
      ontologyId,
      referenceBranchId,
      ts: new Date().toISOString()
    })
  )

  try {
    const params = new URLSearchParams()
    if (referenceBranchId) params.set('reference_branch_id', referenceBranchId)
    const { data } = await api.get(`/versioning/branches?${params}`)
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'versioning.branches.list.success',
        ontologyId,
        total: data.total,
        ts: new Date().toISOString()
      })
    )
    return data
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? 'Failed to list branches'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'versioning.branches.list.failed',
        ontologyId,
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}

export async function getBranch(branchId: string): Promise<BranchInfo> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'versioning.branch.get.request',
      branchId,
      ts: new Date().toISOString()
    })
  )

  try {
    const { data } = await api.get(`/versioning/branches/${branchId}`)
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'versioning.branch.get.success',
        branchId,
        ts: new Date().toISOString()
      })
    )
    return data
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? 'Failed to get branch'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'versioning.branch.get.failed',
        branchId,
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}

// ── Tags ────────────────────────────────────────────────────────────────────────────

export async function listTags(ontologyId: string): Promise<TagInfo[]> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'versioning.tags.list.request',
      ontologyId,
      ts: new Date().toISOString()
    })
  )

  try {
    const { data } = await api.get(`/versioning/tags?ontology_id=${ontologyId}`)
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'versioning.tags.list.success',
        ontologyId,
        count: data.length,
        ts: new Date().toISOString()
      })
    )
    return data
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? 'Failed to list tags'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'versioning.tags.list.failed',
        ontologyId,
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}

// ── Compare Revisions ──────────────────────────────────────────────────────────────

export async function compareRevisions(
  ontologyId: string,
  fromRevision: string,
  toRevision: string
): Promise<CompareRevisionsResult> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'versioning.compare.request',
      ontologyId,
      fromRevision,
      toRevision,
      ts: new Date().toISOString()
    })
  )

  try {
    const { data } = await api.get(`/versioning/commits/${toRevision}/delta?from=${fromRevision}`)
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'versioning.compare.success',
        ontologyId,
        additions: data.additions,
        deletions: data.deletions,
        ts: new Date().toISOString()
      })
    )
    return data
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? 'Failed to compare revisions'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'versioning.compare.failed',
        ontologyId,
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}
