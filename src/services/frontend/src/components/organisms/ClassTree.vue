<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism ClassTree component — hierarchical ontology class browser -->
<template>
  <div class="class-tree" role="tree" :aria-label="'Class hierarchy'">
    <div class="class-tree__header">
      <h3 class="class-tree__title">Classes</h3>
      <SearchInput
        :model-value="filter"
        placeholder="Filter classes..."
        label="Filter classes"
        @update:model-value="filter = $event"
      />
    </div>
    <ul class="class-tree__list" role="group">
      <li v-for="node in filteredNodes" :key="node.id" role="treeitem" :aria-expanded="node.expanded">
        <div
          :class="['class-tree__node', { 'class-tree__node--selected': node.id === selectedId }]"
          :tabindex="0"
          @click="selectNode(node)"
          @keydown.enter="selectNode(node)"
        >
          <button
            v-if="node.children?.length"
            class="class-tree__toggle"
            :aria-label="node.expanded ? 'Collapse' : 'Expand'"
            @click.stop="toggleNode(node)"
          >
            {{ node.expanded ? '▼' : '▶' }}
          </button>
          <span v-else class="class-tree__indent" aria-hidden="true"></span>
          <span class="class-tree__icon" aria-hidden="true">◆</span>
          <span class="class-tree__label">{{ node.label }}</span>
          <span v-if="node.childrenCount" class="class-tree__count">{{ node.childrenCount }}</span>
        </div>
        <ul v-if="node.expanded && node.children" role="group" class="class-tree__children">
          <ClassTreeNode
            v-for="child in node.children"
            :key="child.id"
            :node="child"
            :selected-id="selectedId"
            @select="$emit('select', $event)"
          />
        </ul>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import SearchInput from '../ui-kit/SearchInput.vue'

interface TreeNode {
  id: string
  label: string
  expanded?: boolean
  children?: TreeNode[]
  childrenCount?: number
}

const props = defineProps<{
  nodes: TreeNode[]
  selectedId?: string
}>()

const emit = defineEmits<{
  select: [node: TreeNode]
}>()

const filter = ref('')

const filteredNodes = computed(() => {
  if (!filter.value) return props.nodes
  const q = filter.value.toLowerCase()
  return filterTree(props.nodes, q)
})

function filterTree(nodes: TreeNode[], query: string): TreeNode[] {
  return nodes.reduce<TreeNode[]>((acc, node) => {
    const matches = node.label.toLowerCase().includes(query)
    const filteredChildren = node.children ? filterTree(node.children, query) : []
    if (matches || filteredChildren.length > 0) {
      acc.push({ ...node, children: filteredChildren, expanded: true })
    }
    return acc
  }, [])
}

function selectNode(node: TreeNode) {
  emit('select', node)
}

function toggleNode(node: TreeNode) {
  node.expanded = !node.expanded
}
</script>

<style scoped>
.class-tree {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: auto;
}

.class-tree__header {
  padding: var(--spacing-3);
  border-bottom: 1px solid var(--border-default);
}

.class-tree__title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-2);
}

.class-tree__list {
  flex: 1;
  padding: var(--spacing-2) 0;
}

.class-tree__node {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  padding: var(--spacing-1) var(--spacing-3);
  cursor: pointer;
  border-radius: var(--radius-sm);
}

.class-tree__node:hover {
  background-color: var(--surface-secondary);
}

.class-tree__node--selected {
  background-color: var(--primary-muted);
  color: var(--primary);
}

.class-tree__toggle {
  width: 16px;
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  background: none;
  border: none;
  cursor: pointer;
}

.class-tree__indent { width: 16px; }

.class-tree__icon {
  font-size: var(--font-size-xs);
  color: var(--primary);
}

.class-tree__label {
  flex: 1;
  font-size: var(--font-size-sm);
}

.class-tree__count {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  background: var(--surface-tertiary);
  padding: var(--spacing-1) var(--spacing-2);
  border-radius: var(--radius-full);
}
</style>
