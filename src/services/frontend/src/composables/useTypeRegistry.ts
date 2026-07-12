// @hlv:artifact code-frontend implements spec-gui-ow-001
// @ctx: Ontology UI type registry — maps graph node types to UI representations

import { computed, ref } from 'vue'

export type GraphNodeType = 'class' | 'property' | 'individual'
export type GraphEdgeType = 'subclass_of' | 'object_property' | 'datatype_property'

interface NodeTypeConfig {
  type: GraphNodeType
  label: string
  color: string
  icon: string
}

interface EdgeTypeConfig {
  type: GraphEdgeType
  label: string
  color: string
  style: 'solid' | 'dashed' | 'dotted'
}

const nodeTypes = ref<NodeTypeConfig[]>([
  { type: 'class', label: 'Class', color: '#0f766e', icon: '◆' },
  { type: 'property', label: 'Property', color: '#7c3aed', icon: '◇' },
  { type: 'individual', label: 'Individual', color: '#2563eb', icon: '○' }
])

const edgeTypes = ref<EdgeTypeConfig[]>([
  { type: 'subclass_of', label: 'Subclass of', color: '#0f766e', style: 'solid' },
  { type: 'object_property', label: 'Object property', color: '#7c3aed', style: 'dashed' },
  { type: 'datatype_property', label: 'Datatype property', color: '#2563eb', style: 'dotted' }
])

export function useTypeRegistry() {
  const getNodeConfig = computed(() => {
    return (type: GraphNodeType) => nodeTypes.value.find((t) => t.type === type)
  })

  const getEdgeConfig = computed(() => {
    return (type: GraphEdgeType) => edgeTypes.value.find((t) => t.type === type)
  })

  function registerNodeType(config: NodeTypeConfig) {
    const existing = nodeTypes.value.findIndex((t) => t.type === config.type)
    if (existing >= 0) {
      nodeTypes.value[existing] = config
    } else {
      nodeTypes.value.push(config)
    }
  }

  function registerEdgeType(config: EdgeTypeConfig) {
    const existing = edgeTypes.value.findIndex((t) => t.type === config.type)
    if (existing >= 0) {
      edgeTypes.value[existing] = config
    } else {
      edgeTypes.value.push(config)
    }
  }

  return {
    nodeTypes,
    edgeTypes,
    getNodeConfig,
    getEdgeConfig,
    registerNodeType,
    registerEdgeType
  }
}
