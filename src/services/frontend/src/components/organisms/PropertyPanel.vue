<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Organism PropertyPanel component — displays properties of selected ontology entity -->
<template>
  <div class="property-panel" role="complementary" :aria-label="'Entity properties'">
    <div v-if="!entity" class="property-panel__empty">
      <p>Select an entity to view its properties</p>
    </div>

    <template v-else>
      <div class="property-panel__header">
        <h3 class="property-panel__title">{{ entity.label }}</h3>
        <Badge :text="entity.type" variant="info" />
      </div>

      <div class="property-panel__section">
        <h4 class="property-panel__section-title">General</h4>
        <dl class="property-panel__list">
          <div class="property-panel__item">
            <dt>ID</dt>
            <dd class="property-panel__value">{{ entity.id }}</dd>
          </div>
          <div class="property-panel__item">
            <dt>Type</dt>
            <dd class="property-panel__value">{{ entity.type }}</dd>
          </div>
          <div v-if="entity.description" class="property-panel__item">
            <dt>Description</dt>
            <dd class="property-panel__value">{{ entity.description }}</dd>
          </div>
        </dl>
      </div>

      <div v-if="entity.properties?.length" class="property-panel__section">
        <h4 class="property-panel__section-title">Properties</h4>
        <table class="property-panel__table">
          <thead>
            <tr>
              <th scope="col">Name</th>
              <th scope="col">Type</th>
              <th scope="col">Value</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="prop in entity.properties" :key="prop.name">
              <td>{{ prop.name }}</td>
              <td>{{ prop.type }}</td>
              <td>{{ prop.value }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="entity.relations?.length" class="property-panel__section">
        <h4 class="property-panel__section-title">Relations</h4>
        <ul class="property-panel__relations">
          <li v-for="rel in entity.relations" :key="rel.target">
            {{ rel.type }} → {{ rel.target }}
          </li>
        </ul>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import Badge from '../ui-kit/Badge.vue'

interface EntityProperty {
  name: string
  type: string
  value: string
}

interface EntityRelation {
  type: string
  target: string
}

interface Entity {
  id: string
  label: string
  type: string
  description?: string
  properties?: EntityProperty[]
  relations?: EntityRelation[]
}

defineProps<{
  entity?: Entity | null
}>()
</script>

<style scoped>
.property-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: auto;
  background: var(--surface-primary);
}

.property-panel__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  font-size: var(--font-size-sm);
}

.property-panel__header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--border-default);
}

.property-panel__title {
  flex: 1;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.property-panel__section {
  padding: var(--spacing-4);
  border-bottom: 1px solid var(--border-default);
}

.property-panel__section-title {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--spacing-2);
}

.property-panel__list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.property-panel__item dt {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
}

.property-panel__value {
  font-size: var(--font-size-sm);
  word-break: break-all;
}

.property-panel__table {
  font-size: var(--font-size-sm);
}

.property-panel__table th,
.property-panel__table td {
  padding: var(--spacing-1) var(--spacing-2);
  text-align: left;
  border-bottom: 1px solid var(--border-default);
}

.property-panel__relations {
  font-size: var(--font-size-sm);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}
</style>
