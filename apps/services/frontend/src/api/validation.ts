// REST API client for SHACL validation.
//
// After GraphQL tightening: validation migrated from GraphQL
// (RUN_VALIDATION_MUTATION) to REST POST /api/v1/ontologies/{id}/validate.

import axios from "axios";

const BASE = "/api/v1";

const api = axios.create({
  baseURL: BASE,
  headers: { "X-Requested-With": "XMLHttpRequest" },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("vedo-jwt-token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ── Types ──────────────────────────────────────────────────────────────────────────

export interface ValidationViolation {
  path: string;
  message: string;
  severity: string;
  node: string;
}

export interface ValidationReport {
  status: string;
  violations: ValidationViolation[];
  validatedAt: string;
}

// ── Run Validation ────────────────────────────────────────────────────────────────

export async function runValidation(ontologyId: string): Promise<ValidationReport> {
  console.info(JSON.stringify({
    level: "info", msg: "validation.run.request",
    ontologyId, ts: new Date().toISOString(),
  }));

  try {
    const { data } = await api.post(
      `/ontologies/${ontologyId}/validate`,
      {},
      {
        headers: { "Idempotency-Key": crypto.randomUUID() },
      },
    );
    console.info(JSON.stringify({
      level: "info", msg: "validation.run.success",
      ontologyId, status: data.status, violations: data.violations?.length ?? 0,
      ts: new Date().toISOString(),
    }));
    return data;
  } catch (err: any) {
    const msg = err.response?.data?.error?.message ?? err.message ?? "Validation failed";
    console.error(JSON.stringify({
      level: "error", msg: "validation.run.failed",
      ontologyId, error: msg, ts: new Date().toISOString(),
    }));
    throw new Error(msg);
  }
}
