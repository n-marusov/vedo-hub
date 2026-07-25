// Draft state composable — preserves unsaved changes across navigation.
// After GraphQL tightening: draft persistence migrated from UPDATE_DRAFT_MUTATION
// (GraphQL) to PUT /api/v1/ontologies/{id}/draft (REST).

import { computed, ref } from "vue";
import axios from "axios";

const api = axios.create({
  baseURL: "/api/v1",
  headers: { "X-Requested-With": "XMLHttpRequest" },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("vedo-jwt-token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

interface DraftChange {
  field: string;
  oldValue: unknown;
  newValue: unknown;
  timestamp: number;
}

interface DraftState {
  ontologyId: string | null;
  changes: Map<string, DraftChange>;
  hasUnsavedChanges: boolean;
}

const state = ref<DraftState>({
  ontologyId: null,
  changes: new Map(),
  hasUnsavedChanges: false,
});

export function useDraftState() {
  const hasUnsavedChanges = computed(() => state.value.hasUnsavedChanges);
  const changeCount = computed(() => state.value.changes.size);

  function setOntologyContext(ontologyId: string) {
    state.value.ontologyId = ontologyId;
  }

  function trackChange(field: string, oldValue: unknown, newValue: unknown) {
    if (!state.value.ontologyId) {
      throw new Error(
        "GUI_STORE_NOT_INITIALIZED: Draft state accessed before ontology context",
      );
    }

    state.value.changes.set(field, {
      field,
      oldValue,
      newValue,
      timestamp: Date.now(),
    });
    state.value.hasUnsavedChanges = true;
  }

  function getChange(field: string): DraftChange | undefined {
    return state.value.changes.get(field);
  }

  function getAllChanges(): DraftChange[] {
    return Array.from(state.value.changes.values());
  }

  /** Persist draft via REST (replaces Apollo UPDATE_DRAFT_MUTATION). */
  async function saveDraft(): Promise<boolean> {
    if (!state.value.ontologyId || state.value.changes.size === 0) {
      return false;
    }

    try {
      const changes = getAllChanges();

      console.info(
        JSON.stringify({
          level: "info",
          msg: "draft.save.request",
          ontologyId: state.value.ontologyId,
          changeCount: changes.length,
          ts: new Date().toISOString(),
        }),
      );

      await api.put(
        `/ontologies/${state.value.ontologyId}/draft`,
        { changes: { fields: changes } },
        {
          headers: { "Idempotency-Key": crypto.randomUUID() },
        },
      );

      state.value.changes.clear();
      state.value.hasUnsavedChanges = false;

      console.info(
        JSON.stringify({
          level: "info",
          msg: "draft.save.success",
          ontologyId: state.value.ontologyId,
          ts: new Date().toISOString(),
        }),
      );

      return true;
    } catch (err: any) {
      const msg =
        err.response?.data?.error?.message ?? err.message ?? "Failed to save draft";
      console.error(
        JSON.stringify({
          level: "error",
          msg: "draft.save.failed",
          ontologyId: state.value.ontologyId,
          error: msg,
          ts: new Date().toISOString(),
        }),
      );
      return false;
    }
  }

  function discardChanges() {
    state.value.changes.clear();
    state.value.hasUnsavedChanges = false;
  }

  function reset() {
    state.value.ontologyId = null;
    state.value.changes.clear();
    state.value.hasUnsavedChanges = false;
  }

  return {
    hasUnsavedChanges,
    changeCount,
    setOntologyContext,
    trackChange,
    getChange,
    getAllChanges,
    saveDraft,
    discardChanges,
    reset,
  };
}
