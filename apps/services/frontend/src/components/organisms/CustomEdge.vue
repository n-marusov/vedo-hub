<!-- @hlv:artifact code-frontend implements spec-gui-ow-001 -->
<!-- @ctx: Molecule CustomEdge — Vue Flow custom edge with styled path -->
<template>
  <path
    :d="edgePath"
    class="custom-edge"
    :style="computedStyle"
    fill="none"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  id: string
  sourceX: number
  sourceY: number
  targetX: number
  targetY: number
  style?: Partial<CSSStyleDeclaration>
}>()

const edgePath = computed(() => {
  const dx = props.targetX - props.sourceX
  const cx = dx * 0.5
  return `M${props.sourceX},${props.sourceY} C${props.sourceX + cx},${props.sourceY} ${props.targetX - cx},${props.targetY} ${props.targetX},${props.targetY}`
})

const computedStyle = computed(() => ({
  stroke: (props.style as Record<string, string>)?.stroke ?? '#666',
  strokeWidth: (props.style as Record<string, string>)?.strokeWidth ?? '1px',
  strokeDasharray: (props.style as Record<string, string>)?.strokeDasharray ?? 'none'
}))
</script>

<style scoped>
.custom-edge {
  pointer-events: none;
}
</style>
