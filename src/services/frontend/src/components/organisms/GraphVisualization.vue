<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism GraphVisualization component — interactive 2D graph using Vue Flow placeholder -->
<template>
  <div class="graph-viz" role="img" :aria-label="'Ontology graph visualization'">
    <div class="graph-viz__toolbar">
      <button class="graph-viz__tool" aria-label="Zoom in" @click="zoomIn">+</button>
      <button class="graph-viz__tool" aria-label="Zoom out" @click="zoomOut">−</button>
      <button class="graph-viz__tool" aria-label="Fit view" @click="fitView">⊡</button>
      <select v-model="layout" class="graph-viz__layout" aria-label="Layout mode">
        <option value="dagre">Dagre</option>
        <option value="grid">Grid</option>
        <option value="circular">Circular</option>
      </select>
    </div>

    <div class="graph-viz__canvas" ref="canvasRef">
      <!-- Vue Flow placeholder — graph nodes rendered as positioned divs -->
      <div
        v-for="node in nodes"
        :key="node.id"
        :class="['graph-viz__node', `graph-viz__node--${node.type}`]"
        :style="{ left: `${node.x}px`, top: `${node.y}px` }"
        :tabindex="0"
        role="button"
        :aria-label="`${node.label} (${node.type})`"
        @click="$emit('node-click', node)"
        @keydown.enter="$emit('node-click', node)"
      >
        <span class="graph-viz__node-icon">{{ getNodeIcon(node.type) }}</span>
        <span class="graph-viz__node-label">{{ node.label }}</span>
      </div>

      <!-- SVG edges -->
      <svg class="graph-viz__edges" aria-hidden="true">
        <line
          v-for="edge in edges"
          :key="`${edge.source}-${edge.target}`"
          :x1="getNodePosition(edge.source).x"
          :y1="getNodePosition(edge.source).y"
          :x2="getNodePosition(edge.target).x"
          :y2="getNodePosition(edge.target).y"
          :class="`graph-viz__edge--${edge.type}`"
        />
      </svg>
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
import { ref } from 'vue'

interface GraphNode {
  id: string
  label: string
  type: 'class' | 'property' | 'individual'
  x: number
  y: number
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
}>()

defineEmits<{
  'node-click': [node: GraphNode]
}>()

const layout = ref('dagre')
const canvasRef = ref<HTMLElement | null>(null)

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

function zoomIn() {
  /* Vue Flow zoom */
}
function zoomOut() {
  /* Vue Flow zoom */
}
function fitView() {
  /* Vue Flow fit */
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
  overflow: auto;
  background: var(--surface-secondary);
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
  transition: all var(--transition-fast);
}

.graph-viz__node:hover {
  border-color: var(--primary);
  box-shadow: var(--shadow-md);
}

.graph-viz__node--class { border-left: 3px solid var(--primary); }
.graph-viz__node--property { border-left: 3px solid var(--accent); }
.graph-viz__node--individual { border-left: 3px solid var(--status-info); }

.graph-viz__edges {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.graph-viz__edge--subclass_of { stroke: var(--primary); stroke-width: 2; }
.graph-viz__edge--object_property { stroke: var(--accent); stroke-width: 2; stroke-dasharray: 5,5; }
.graph-viz__edge--datatype_property { stroke: var(--status-info); stroke-width: 1; stroke-dasharray: 2,2; }

.graph-viz__table {
  margin-top: var(--spacing-4);
  font-size: var(--font-size-sm);
}
</style>
