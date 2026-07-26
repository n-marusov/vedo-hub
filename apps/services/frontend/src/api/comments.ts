// REST API client for commenting service.
//
// After GraphQL tightening: comments migrated from GraphQL
// (GET_ENTITY_COMMENTS_QUERY, CREATE_COMMENT_MUTATION) to REST.

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

export interface CommentInfo {
  id: string
  author: string
  authorName: string
  text: string
  entityId: string
  entityType: string
  parentCommentId: string | null
  createdAt: string
  updatedAt: string
}

export interface CommentListResult {
  comments: CommentInfo[]
  total: number
  page: number
  pageSize: number
}

// ── List Comments ──────────────────────────────────────────────────────────────────

export async function listComments(
  ontologyId: string,
  entityId: string,
  page = 1,
  pageSize = 20
): Promise<CommentListResult> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'comments.list.request',
      ontologyId,
      entityId,
      page,
      pageSize,
      ts: new Date().toISOString()
    })
  )

  try {
    const { data } = await api.get(
      `/ontologies/${ontologyId}/comments?entity_id=${encodeURIComponent(entityId)}&page=${page}&page_size=${pageSize}`
    )
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'comments.list.success',
        ontologyId,
        entityId,
        total: data.total,
        ts: new Date().toISOString()
      })
    )
    return data
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? 'Failed to list comments'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'comments.list.failed',
        ontologyId,
        entityId,
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}

// ── Create Comment ─────────────────────────────────────────────────────────────────

export async function createComment(
  ontologyId: string,
  entityId: string,
  text: string
): Promise<CommentInfo> {
  console.info(
    JSON.stringify({
      level: 'info',
      msg: 'comments.create.request',
      ontologyId,
      entityId,
      ts: new Date().toISOString()
    })
  )

  try {
    const { data } = await api.post(
      `/ontologies/${ontologyId}/comments`,
      {
        entity_id: entityId,
        entity_type: 'class',
        body: text
      },
      {
        headers: { 'Idempotency-Key': crypto.randomUUID() }
      }
    )
    console.info(
      JSON.stringify({
        level: 'info',
        msg: 'comments.create.success',
        ontologyId,
        entityId,
        commentId: data.id,
        ts: new Date().toISOString()
      })
    )
    return data
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? 'Failed to create comment'
    console.error(
      JSON.stringify({
        level: 'error',
        msg: 'comments.create.failed',
        ontologyId,
        entityId,
        error: msg,
        ts: new Date().toISOString()
      })
    )
    throw new Error(msg)
  }
}
