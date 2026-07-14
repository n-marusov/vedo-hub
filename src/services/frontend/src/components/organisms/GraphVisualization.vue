<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism GraphVisualization component — interactive 2D graph with zoom, pan, depth controls -->
<template>
  <div class="graph-viz" role="img" :aria-label="'Ontology graph visualization'">
    <div class="graph-viz__toolbar">
      <button class="graph-viz__tool" aria-label="Zoom in" @click="zoomIn">+</button>
      <button class="graph-viz__tool" aria-label="Zoom out" @click="zoomOut">−</button>
      <button class="graph-viz__tool" aria-label="Fit view" @click="fitView">⊡</button>
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

    <div
      class="graph-viz__canvas"
      ref="canvasRef"
      @mousedown="onCanvasMouseDown"
      @mousemove="onCanvasMouseMove"
      @mouseup="onCanvasMouseUp"
      @mouseleave="onCanvasMouseUp"
      @wheel.prevent="onWheel"
      :style="{ cursor: isPanning ? 'grabbing' : 'grab' }"
    >
      <!-- SVG edges layer -->
      <svg class="graph-viz__edges" aria-hidden="true">
        <line
          v-for="edge in visibleEdges"
          :key="`${edge.source}-${edge.target}`"
          :x1="getTransformedX(getNodePosition(edge.source).x)"
          :y1="getTransformedY(getNodePosition(edge.source).y)"
          :x2="getTransformedX(getNodePosition(edge.target).x)"
          :y2="getTransformedY(getNodePosition(edge.target).y)"
          :class="`graph-viz__edge--${edge.type}`"
        />
      </svg>

      <!-- Nodes layer -->
      <div
        v-for="node in visibleNodes"
        :key="node.id"
        :class="['graph-viz__node', `graph-viz__node--${node.type}`, { 'graph-viz__node--selected': node.id === selectedNodeId }]"
        :style="{
          left: `${getTransformedX(node.x)}px`,
          top: `${getTransformedY(node.y)}px`,
          transform: `scale(${zoomLevel})`,
          transformOrigin: 'top left',
        }"
        :tabindex="0"
        role="button"
        :aria-label="`${node.label} (${node.type})`"
        @click="onNodeClick(node)"
        @keydown.enter="onNodeClick(node)"
      >
        <span class="graph-viz__node-icon">{{ getNodeIcon(node.type) }}</span>
        <span class="graph-viz__node-label">{{ node.label }}</span>
      </div>
    </div>

    <!-- Status bar -->
    <div class="graph-viz__status">
      <span>{{ visibleNodes.length }} / {{ nodes.length }} nodes</span>
      <span>{{ visibleEdges.length }} / {{ edges.length }} edges</span>
      <span>Zoom: {{ Math.round(zoomLevel * 100) }}%</span>
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
import { computed, ref } from 'vue'

interface GraphNode {
  id: string
  label: string
  type: 'class' | 'property' | 'individual'
  x: number
  y: number
  depth?: number
}

interface GraphEdge {
  source: string
  target: string
  type: 'subclass_of' | 'object_property' | 'datatype_property'
}

const props = defineProps<{
  nodes: GraphNode[]
  edges: GraphEdge[]
  showTableFallback?: boolean
  maxDepth?: number
}>()

const emit = defineEmits<{
  'node-click': [node: GraphNode]
  'update:maxDepth': [depth: number]
}>()

// ── UI State ────────────────────────────────────────────────────────────────

const layout = ref('dagre')
const canvasRef = ref<HTMLElement | null>(null)
const selectedNodeId = ref<string | null>(null)
const depth = ref(props.maxDepth ?? 3)
const nodeTypeFilter = ref('all')

// ── Zoom & Pan ──────────────────────────────────────────────────────────────

const zoomLevel = ref(1)
const panX = ref(0)
const panY = ref(0)
const isPanning = ref(false)
let lastMouseX = 0
let lastMouseY = 0

const MIN_ZOOM = 0.1
const MAX_ZOOM = 3.0
const ZOOM_STEP = 0.1

function zoomIn() {
  zoomLevel.value = Math.min(MAX_ZOOM, zoomLevel.value + ZOOM_STEP)
}

function zoomOut() {
  zoomLevel.value = Math.max(MIN_ZOOM, zoomLevel.value - ZOOM_STEP)
}

function fitView() {
  zoomLevel.value = 1
  panX.value = 0
  panY.value = 0
}

function onWheel(event: WheelEvent) {
  const delta = event.deltaY > 0 ? -ZOOM_STEP : ZOOM_STEP
  zoomLevel.value = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, zoomLevel.value + delta))
}

function onCanvasMouseDown(event: MouseEvent) {
  isPanning.value = true
  lastMouseX = event.clientX
  lastMouseY = event.clientY
}

function onCanvasMouseMove(event: MouseEvent) {
  if (!isPanning.value) return
  const dx = event.clientX - lastMouseX
  const dy = event.clientY - lastMouseY
  panX.value += dx
  panY.value += dy
  lastMouseX = event.clientX
  lastMouseY = event.clientY
}

function onCanvasMouseUp() {
  isPanning.value = false
}

function getTransformedX(x: number): number {
  return x * zoomLevel.value + panX.value
}

function getTransformedY(y: number): number {
  return y * zoomLevel.value + panY.value
}

// ── Filtering ───────────────────────────────────────────────────────────────

const visibleNodes = computed(() => {
  return props.nodes.filter((n) => {
    // Type filter
    if (nodeTypeFilter.value !== 'all' && n.type !== nodeTypeFilter.value) return false
    // Depth filter
    if (depth.value < 10 && (n.depth ?? 0) > depth.value) return false
    return true
  })
})

const visibleNodeIds = computed(() => new Set(visibleNodes.value.map((n) => n.id)))

const visibleEdges = computed(() => {
  return props.edges.filter((e) => visibleNodeIds.value.has(e.source) && visibleNodeIds.value.has(e.target))
})

// ── Helpers ─────────────────────────────────────────────────────────────────

function getNodeIcon(type: string): string {
  return type === 'class' ? '◆' : type === 'property' ? '◇' : '○'
}

function getNodePosition(id: string): { x: number; y: number } {
  const node = props.nodes.find((n) => n.id === id)
  return node ? { x: node.x + 40, y: node.y + 20 } : { x: 0, y: 0 }
}

function getEdgeCount(nodeId: string): number {
  return props.edges.filter((e) => e.source === nodeId || e.target === nodeId).length
}

function onNodeClick(node: GraphNode) {
  selectedNodeId.value = node.id
  emit('node-click', node)
}
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

.graph-viz__edges {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.graph-viz__node {
  position: absolute;
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--surface-primary);
  border: 2px solid var(--border-default);
  border-radius: var(--radius-lg);
  cursor: pointer;
  font-size: var(--font-size-sm);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
  white-space: nowrap;
  user-select: none;
}

.graph-viz__node:hover {
  border-color: var(--primary);
  box-shadow: var(--shadow-md);
  z-index: 10;
}

.graph-viz__node--selected {
  border-color: var(--primary) !important;
  box-shadow: 0 0 0 2px var(--primary-muted);
  z-index: 11;
}

.graph-viz__node--class { border-left: 3px solid var(--primary); }
.graph-viz__node--property { border-left: 3px solid var(--accent); }
.graph-viz__node--individual { border-left: 3px solid var(--status-info); }

.graph-viz__node-icon {
  font-size: var(--font-size-xs);
}

.graph-viz__node--class .graph-viz__node-icon { color: var(--primary); }
.graph-viz__node--property .graph-viz__node-icon { color: var(--accent); }
.graph-viz__node--individual .graph-viz__node-icon { color: var(--status-info); }

.graph-viz__edge--subclass_of { stroke: var(--primary); stroke-width: 2; }
.graph-viz__edge--object_property { stroke: var(--accent); stroke-width: 2; stroke-dasharray: 5,5; }
.graph-viz__edge--datatype_property { stroke: var(--status-info); stroke-width: 1; stroke-dasharray: 2,2; }

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
