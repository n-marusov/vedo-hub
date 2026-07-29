<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism GraphVisualization component — interactive 2D graph with Vue Flow -->
<template>
  <div class="graph-viz" role="img" :aria-label="'Ontology graph visualization'">
    <div class="graph-viz__toolbar">
      <button class="graph-viz__tool" aria-label="Zoom in" @click="() => zoomIn()">+</button>
        <button class="graph-viz__tool" aria-label="Zoom out" @click="() => zoomOut()">−</button>
        <button class="graph-viz__tool" aria-label="Fit view" @click="() => fitView()">⊡</button>
      <label class="graph-viz__depth-label">
        Depth:
        <select v-model.number="depth" class="graph-viz__depth" aria-label="Graph depth">
          <option :value="1">1</option>
          <option :value="2">2</option>
          <option :value="3">3</option>
          <option :value="5">5</option>
          <option :value="10">10</option>
        </select>
      </label>
      <label class="graph-viz__filter-label">
        Type:
        <select v-model="nodeTypeFilter" class="graph-viz__filter" aria-label="Node type filter">
          <option value="all">All</option>
          <option value="class">Classes</option>
          <option value="property">Properties</option>
          <option value="individual">Individuals</option>
        </select>
      </label>
      <select v-model="layout" class="graph-viz__layout" aria-label="Layout mode">
        <option value="dagre">Dagre</option>
        <option value="grid">Grid</option>
        <option value="circular">Circular</option>
      </select>
    </div>

    <div class="graph-viz__canvas" ref="canvasRef">
      <VueFlow
        :nodes="flowNodes"
        :edges="flowEdges"
        @node-click="onNodeClick"
        :fit-view-on-init="true"
        :min-zoom="0.1"
        :max-zoom="3"
        :default-viewport="{ zoom: 1, x: 0, y: 0 }"
        class="graph-viz__flow"
        aria-label="Ontology graph"
      >
        <template #node-ontology-class="nodeProps">
          <div class="graph-viz__flow-node graph-viz__flow-node--class" :title="nodeProps.data.label">
            <span class="graph-viz__flow-node-icon">◆</span>
            <span class="graph-viz__flow-node-label">{{ nodeProps.data.label }}</span>
          </div>
        </template>
        <template #node-ontology-property="nodeProps">
          <div class="graph-viz__flow-node graph-viz__flow-node--property" :title="nodeProps.data.label">
            <span class="graph-viz__flow-node-icon">◇</span>
            <span class="graph-viz__flow-node-label">{{ nodeProps.data.label }}</span>
          </div>
        </template>
        <template #node-ontology-individual="nodeProps">
          <div class="graph-viz__flow-node graph-viz__flow-node--individual" :title="nodeProps.data.label">
            <span class="graph-viz__flow-node-icon">○</span>
            <span class="graph-viz__flow-node-label">{{ nodeProps.data.label }}</span>
          </div>
        </template>
        <template #edge-custom="edgeProps">
          <CustomEdge
            :id="edgeProps.id"
            :source-x="edgeProps.sourceX"
            :source-y="edgeProps.sourceY"
            :target-x="edgeProps.targetX"
            :target-y="edgeProps.targetY"
            :style="getEdgeStyle(edgeProps.data?.type)"
          />
        </template>

        <Controls
          :show-zoom="true"
          :show-fit-view="true"
          :show-lock="false"
          position="bottom-right"
        />
        <Background :gap="20" variant="dots" />
      </VueFlow>
    </div>

    <!-- Status bar -->
    <div class="graph-viz__status">
      <span>{{ visibleNodes.length }} / {{ nodes.length }} nodes</span>
      <span>{{ visibleEdges.length }} / {{ edges.length }} edges</span>
    </div>

    <!-- Table fallback for accessibility -->
    <table v-if="showTableFallback" class="graph-viz__table" aria-label="Graph data as table">
      <thead>
        <tr>
          <th scope="col">Node</th>
          <th scope="col">Type</th>
          <th scope="col">Connections</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="node in nodes" :key="node.id">
          <td>{{ node.label }}</td>
          <td>{{ node.type }}</td>
          <td>{{ getEdgeCount(node.id) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { Background } from "@vue-flow/background";
import { Controls } from "@vue-flow/controls";
import {
	type Edge as FlowEdge,
	type Node,
	VueFlow,
	useVueFlow,
} from "@vue-flow/core";
import { computed, ref, watch } from "vue";
import CustomEdge from "./CustomEdge.vue";

import "@vue-flow/core/dist/style.css";
import "@vue-flow/core/dist/theme-default.css";
import "@vue-flow/controls/dist/style.css";

interface GraphNode {
	id: string;
	label: string;
	type: "class" | "property" | "individual";
	x: number;
	y: number;
	depth?: number;
}

interface GraphEdge {
	source: string;
	target: string;
	type: "subclass_of" | "object_property" | "datatype_property";
}

const props = defineProps<{
	nodes: GraphNode[];
	edges: GraphEdge[];
	showTableFallback?: boolean;
	maxDepth?: number;
}>();

const emit = defineEmits<{
	"node-click": [node: GraphNode];
	"update:maxDepth": [depth: number];
}>();

// ── UI State ────────────────────────────────────────────────────────────────

const layout = ref("dagre");
const depth = ref(props.maxDepth ?? 3);
const nodeTypeFilter = ref("all");

const { fitView, zoomIn, zoomOut } = useVueFlow();

// ── Node type mapping to Vue Flow node types ──────────────────────────────

function mapType(type: GraphNode["type"]): string {
	if (type === "class") return "ontology-class";
	if (type === "property") return "ontology-property";
	return "ontology-individual";
}

// ── Edge styling ─────────────────────────────────────────────────────────

function getEdgeStyle(edgeType?: string): Partial<CSSStyleDeclaration> {
	switch (edgeType) {
		case "subclass_of":
			return { stroke: "var(--primary, #4f6ef7)", strokeWidth: "2px" };
		case "object_property":
			return {
				stroke: "var(--accent, #e68a2e)",
				strokeWidth: "2px",
				strokeDasharray: "5,5",
			};
		case "datatype_property":
			return {
				stroke: "var(--status-info, #3b82f6)",
				strokeWidth: "1px",
				strokeDasharray: "2,2",
			};
		default:
			return { stroke: "#666", strokeWidth: "1px" };
	}
}

// ── Filtering ───────────────────────────────────────────────────────────────

const visibleNodes = computed(() => {
	return props.nodes.filter((n) => {
		if (nodeTypeFilter.value !== "all" && n.type !== nodeTypeFilter.value)
			return false;
		if (depth.value < 10 && (n.depth ?? 0) > depth.value) return false;
		return true;
	});
});

const visibleNodeIds = computed(
	() => new Set(visibleNodes.value.map((n) => n.id)),
);

const visibleEdges = computed(() => {
	return props.edges.filter(
		(e) =>
			visibleNodeIds.value.has(e.source) && visibleNodeIds.value.has(e.target),
	);
});

// ── Vue Flow conversion ────────────────────────────────────────────────────

const flowNodes = computed<Node[]>(() => {
	return visibleNodes.value.map((n) => ({
		id: n.id,
		type: mapType(n.type),
		position: { x: n.x, y: n.y },
		data: {
			label: n.label,
			type: n.type,
			depth: n.depth,
		},
	}));
});

const flowEdges = computed<FlowEdge[]>(() => {
	return visibleEdges.value.map((e) => ({
		id: `${e.source}-${e.target}`,
		source: e.source,
		target: e.target,
		type: "custom",
		data: { type: e.type },
	}));
});

// ── Helpers ─────────────────────────────────────────────────────────────────

function getEdgeCount(nodeId: string): number {
	return props.edges.filter((e) => e.source === nodeId || e.target === nodeId)
		.length;
}

function onNodeClick(nodeMouseEvent: { node: Node }) {
	const node = nodeMouseEvent.node;
	const original = props.nodes.find((n) => n.id === node.id);
	if (original) {
		emit("node-click", original);
	}
}

// ── Depth watcher ─────────────────────────────────────────────────────────

watch(depth, (val) => {
	emit("update:maxDepth", val);
});
</script>

<style scoped>
.graph-viz {
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;
}

.graph-viz__toolbar {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2);
  background: var(--surface-primary);
  border-bottom: 1px solid var(--border-default);
  flex-wrap: wrap;
}

.graph-viz__tool {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-primary);
  cursor: pointer;
  font-size: var(--font-size-lg);
}

.graph-viz__depth-label,
.graph-viz__filter-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.graph-viz__depth,
.graph-viz__filter {
  padding: var(--spacing-1) var(--spacing-2);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
  background: var(--surface-primary);
}

.graph-viz__layout {
  margin-left: auto;
  padding: var(--spacing-1) var(--spacing-2);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.graph-viz__canvas {
  flex: 1;
  position: relative;
  overflow: hidden;
  background: var(--surface-secondary);
}

.graph-viz__flow {
  width: 100%;
  height: 100%;
}

/* Custom node styles for Vue Flow */
:deep(.graph-viz__flow-node) {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--surface-primary);
  border: 2px solid var(--border-default);
  border-radius: var(--radius-lg);
  cursor: pointer;
  font-size: var(--font-size-sm);
  white-space: nowrap;
  user-select: none;
  transition: border-color 0.15s, box-shadow 0.15s;
  min-width: 80px;
  max-width: 200px;
}

:deep(.graph-viz__flow-node:hover) {
  border-color: var(--primary);
  box-shadow: var(--shadow-md);
  z-index: 10;
}

:deep(.graph-viz__flow-node--class) {
  border-left: 3px solid var(--primary);
}

:deep(.graph-viz__flow-node--property) {
  border-left: 3px solid var(--accent);
}

:deep(.graph-viz__flow-node--individual) {
  border-left: 3px solid var(--status-info);
}

:deep(.graph-viz__flow-node-icon) {
  font-size: var(--font-size-xs);
  flex-shrink: 0;
}

:deep(.graph-viz__flow-node--class .graph-viz__flow-node-icon) {
  color: var(--primary);
}

:deep(.graph-viz__flow-node--property .graph-viz__flow-node-icon) {
  color: var(--accent);
}

:deep(.graph-viz__flow-node--individual .graph-viz__flow-node-icon) {
  color: var(--status-info);
}

:deep(.graph-viz__flow-node-label) {
  overflow: hidden;
  text-overflow: ellipsis;
}

.graph-viz__status {
  display: flex;
  gap: var(--spacing-4);
  padding: var(--spacing-1) var(--spacing-3);
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  background: var(--surface-primary);
  border-top: 1px solid var(--border-default);
}

.graph-viz__table {
  margin-top: var(--spacing-4);
  font-size: var(--font-size-sm);
}
</style>
