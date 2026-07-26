<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism ClassTree component — hierarchical ontology class browser with drag-and-drop -->
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
      <div v-if="lastUndo" class="class-tree__undo">
        <button class="class-tree__undo-btn" @click="undoDrop" title="Undo last move">
          ↩ Undo
        </button>
      </div>
    </div>
    <ul class="class-tree__list" role="group">
      <li
        v-for="node in filteredNodes"
        :key="node.id"
        role="treeitem"
        :aria-expanded="node.expanded"
        :draggable="!readonly"
        @dragstart="onDragStart($event, node)"
        @dragover.prevent="onDragOver($event, node)"
        @dragleave="onDragLeave($event, node)"
        @drop.prevent="onDrop($event, node)"
        :class="{ 'class-tree__drop-target': dropTargetId === node.id }"
      >
        <div
          :class="['class-tree__node', { 'class-tree__node--selected': node.id === selectedId, 'class-tree__node--dragging': draggingId === node.id }]"
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
          <li
            v-for="child in node.children"
            :key="child.id"
            role="treeitem"
            :aria-expanded="child.expanded"
            :draggable="!readonly"
            @dragstart="onDragStart($event, child)"
            @dragover.prevent="onDragOver($event, child)"
            @dragleave="onDragLeave($event, child)"
            @drop.prevent="onDrop($event, child)"
            :class="{ 'class-tree__drop-target': dropTargetId === child.id }"
          >
            <div
              :class="['class-tree__node', { 'class-tree__node--selected': child.id === selectedId, 'class-tree__node--dragging': draggingId === child.id }]"
              :tabindex="0"
              @click="selectNode(child)"
              @keydown.enter="selectNode(child)"
            >
              <button
                v-if="child.children?.length"
                class="class-tree__toggle"
                :aria-label="child.expanded ? 'Collapse' : 'Expand'"
                @click.stop="toggleNode(child)"
              >
                {{ child.expanded ? '▼' : '▶' }}
              </button>
              <span v-else class="class-tree__indent" aria-hidden="true"></span>
              <span class="class-tree__icon" aria-hidden="true">◆</span>
              <span class="class-tree__label">{{ child.label }}</span>
              <span v-if="child.childrenCount" class="class-tree__count">{{ child.childrenCount }}</span>
            </div>
            <ul v-if="child.expanded && child.children" role="group" class="class-tree__children">
              <li
                v-for="grandchild in child.children"
                :key="grandchild.id"
                role="treeitem"
                :draggable="!readonly"
                @dragstart="onDragStart($event, grandchild)"
                @dragover.prevent="onDragOver($event, grandchild)"
                @dragleave="onDragLeave($event, grandchild)"
                @drop.prevent="onDrop($event, grandchild)"
              >
                <div
                  :class="['class-tree__node', { 'class-tree__node--selected': grandchild.id === selectedId, 'class-tree__node--dragging': draggingId === grandchild.id }]"
                  :tabindex="0"
                  @click="selectNode(grandchild)"
                  @keydown.enter="selectNode(grandchild)"
                >
                  <span class="class-tree__icon" aria-hidden="true">◆</span>
                  <span class="class-tree__label">{{ grandchild.label }}</span>
                </div>
              </li>
            </ul>
          </li>
        </ul>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import SearchInput from "../ui-kit/SearchInput.vue";

interface TreeNode {
	id: string;
	label: string;
	expanded?: boolean;
	children?: TreeNode[];
	childrenCount?: number;
}

interface UndoState {
	sourceId: string;
	targetParentId: string | null;
	oldParentId: string | null;
	sourceNode: TreeNode;
}

const props = defineProps<{
	nodes: TreeNode[];
	selectedId?: string;
	readonly?: boolean;
}>();

const emit = defineEmits<{
	select: [node: TreeNode];
	"move-class": [sourceId: string, newParentId: string | null];
}>();

const filter = ref("");
const draggingId = ref<string | null>(null);
const dropTargetId = ref<string | null>(null);
const lastUndo = ref<UndoState | null>(null);

const filteredNodes = computed(() => {
	if (!filter.value) return props.nodes;
	const q = filter.value.toLowerCase();
	return filterTree(props.nodes, q);
});

function filterTree(nodes: TreeNode[], query: string): TreeNode[] {
	return nodes.reduce<TreeNode[]>((acc, node) => {
		const matches = node.label.toLowerCase().includes(query);
		const filteredChildren = node.children
			? filterTree(node.children, query)
			: [];
		if (matches || filteredChildren.length > 0) {
			acc.push({ ...node, children: filteredChildren, expanded: true });
		}
		return acc;
	}, []);
}

function selectNode(node: TreeNode) {
	emit("select", node);
}

function toggleNode(node: TreeNode) {
	node.expanded = !node.expanded;
}

// ── Drag-and-drop ──────────────────────────────────────────────────────────

function onDragStart(event: DragEvent, node: TreeNode) {
	if (props.readonly) return;
	draggingId.value = node.id;
	event.dataTransfer?.setData("text/plain", node.id);
	if (event.dataTransfer) event.dataTransfer.effectAllowed = "move";
}

function onDragOver(event: DragEvent, node: TreeNode) {
	if (props.readonly || draggingId.value === node.id) return;
	// Prevent dropping on self or own descendants (cycle prevention)
	const dragId = draggingId.value;
	if (!dragId || isDescendant(node, dragId)) return;
	dropTargetId.value = node.id;
	if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
}

function onDragLeave(_event: DragEvent, _node: TreeNode) {
	dropTargetId.value = null;
}

function onDrop(_event: DragEvent, targetNode: TreeNode) {
	dropTargetId.value = null;
	const sourceId = draggingId.value;
	draggingId.value = null;

	if (!sourceId || props.readonly) return;

	// Prevent cycle: cannot drop on self or own descendant
	if (sourceId === targetNode.id || isDescendant(targetNode, sourceId)) return;

	emit("move-class", sourceId, targetNode.id);
	lastUndo.value = {
		sourceId,
		targetParentId: targetNode.id,
		oldParentId: findParentId(props.nodes, sourceId),
		sourceNode: findNode(props.nodes, sourceId) ?? { id: sourceId, label: "" },
	};
}

function undoDrop() {
	if (!lastUndo.value) return;
	// Re-emit the move with the old parent (null = root level)
	emit("move-class", lastUndo.value.sourceId, lastUndo.value.oldParentId);
	lastUndo.value = null;
}

/// Checks if `targetId` is a descendant of `node` (cycle detection).
function isDescendant(node: TreeNode, targetId: string): boolean {
	if (!node.children) return false;
	for (const child of node.children) {
		if (child.id === targetId) return true;
		if (isDescendant(child, targetId)) return true;
	}
	return false;
}

/// Finds the parent ID of a node by searching the tree.
function findParentId(nodes: TreeNode[], targetId: string): string | null {
	for (const node of nodes) {
		if (node.children?.some((c) => c.id === targetId)) return node.id;
		if (node.children) {
			const found = findParentId(node.children, targetId);
			if (found) return found;
		}
	}
	return null;
}

/// Finds a node by ID anywhere in the tree.
function findNode(nodes: TreeNode[], targetId: string): TreeNode | null {
	for (const node of nodes) {
		if (node.id === targetId) return node;
		if (node.children) {
			const found = findNode(node.children, targetId);
			if (found) return found;
		}
	}
	return null;
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
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.class-tree__title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
  margin: 0;
}

.class-tree__undo {
  display: flex;
  justify-content: flex-end;
}

.class-tree__undo-btn {
  font-size: var(--font-size-xs);
  color: var(--primary);
  background: none;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  padding: var(--spacing-1) var(--spacing-2);
  cursor: pointer;
}

.class-tree__list {
  flex: 1;
  padding: var(--spacing-2) 0;
  list-style: none;
  margin: 0;
}

.class-tree__children {
  list-style: none;
  margin: 0;
  padding-left: var(--spacing-4);
}

.class-tree__node {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  padding: var(--spacing-1) var(--spacing-3);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: background-color var(--transition-fast);
}

.class-tree__node:hover {
  background-color: var(--surface-secondary);
}

.class-tree__node--selected {
  background-color: var(--primary-muted);
  color: var(--primary);
}

.class-tree__node--dragging {
  opacity: 0.5;
}

.class-tree__drop-target > .class-tree__node {
  border: 2px dashed var(--primary);
  background-color: var(--primary-muted);
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
