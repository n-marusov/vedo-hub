<!-- @hlv:artifact code-frontend implements spec-gui-ticket-001 -->
<!-- @ctx: Attachment upload — file selection with size/count validation -->
<!-- @hlv:sec [INPUT_VALIDATION] — user-supplied files validated for size and count before upload -->

<template>
  <div class="attachment-upload" role="group" aria-labelledby="attachment-title">
    <h3 id="attachment-title" class="attachment-upload__title">
      Attachments ({{ files.length }}/10)
    </h3>

    <div
      class="attachment-upload__dropzone"
      :class="{ 'attachment-upload__dropzone--dragover': isDragOver }"
      @dragover.prevent="isDragOver = true"
      @dragleave.prevent="isDragOver = false"
      @drop.prevent="onDrop"
      @click="triggerInput"
      role="button"
      tabindex="0"
      aria-label="Click or drag files to attach"
      @keydown.enter="triggerInput"
      @keydown.space.prevent="triggerInput"
    >
      <input
        ref="fileInput"
        type="file"
        multiple
        class="sr-only"
        @change="onFileSelect"
        aria-hidden="true"
      />
      <span class="attachment-upload__hint">Click or drag files here (max 10 files, 10 MB each)</span>
    </div>

    <!-- @ctx: file list with remove action -->
    <ul v-if="files.length > 0" class="attachment-upload__list" aria-label="Attached files">
      <li v-for="(file, index) in files" :key="file.name + file.size" class="attachment-upload__item">
        <span class="attachment-upload__filename">{{ file.name }}</span>
        <span class="attachment-upload__size">{{ formatSize(file.size) }}</span>
        <button
          class="attachment-upload__remove"
          @click="removeFile(index)"
          :aria-label="`Remove ${file.name}`"
        >
          ✕
        </button>
      </li>
    </ul>

    <!-- @ctx: error display for attachment limits -->
    <!-- @hlv TICKET-UI-TOO-MANY-FILES -->
    <!-- @hlv TICKET-UI-FILE-TOO-LARGE -->
    <div v-if="error" class="attachment-upload__error" role="alert" aria-live="polite">
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const MAX_FILES = 10
const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10 MB

const props = defineProps<{
  modelValue: File[]
}>()

const emit = defineEmits<{
  'update:modelValue': [files: File[]]
  error: [code: string]
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const isDragOver = ref(false)
const error = ref<string | null>(null)

const files = ref<File[]>([...props.modelValue])

function triggerInput() {
  fileInput.value?.click()
}

function validateAndAdd(newFiles: FileList | File[]) {
  error.value = null
  const fileArray = Array.from(newFiles)

  // @ctx: count limit check
  // @hlv TICKET-UI-TOO-MANY-FILES
  if (files.value.length + fileArray.length > MAX_FILES) {
    error.value = `Cannot attach more than ${MAX_FILES} files. You have ${files.value.length} already.`
    emit('error', 'TICKET-UI-TOO-MANY-FILES')
    return
  }

  const valid: File[] = []
  for (const file of fileArray) {
    // @ctx: size limit check
    // @hlv TICKET-UI-FILE-TOO-LARGE
    if (file.size > MAX_FILE_SIZE) {
      error.value = `File "${file.name}" exceeds 10 MB limit (${formatSize(file.size)}).`
      emit('error', 'TICKET-UI-FILE-TOO-LARGE')
      continue
    }
    valid.push(file)
  }

  files.value = [...files.value, ...valid]
  emit('update:modelValue', files.value)
}

function onFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files) {
    validateAndAdd(input.files)
    input.value = ''
  }
}

function onDrop(event: DragEvent) {
  isDragOver.value = false
  if (event.dataTransfer?.files) {
    validateAndAdd(event.dataTransfer.files)
  }
}

function removeFile(index: number) {
  files.value.splice(index, 1)
  emit('update:modelValue', [...files.value])
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
</script>

<style scoped>
.attachment-upload { display: flex; flex-direction: column; gap: var(--spacing-3); }
.attachment-upload__title { font-size: var(--font-size-md); font-weight: var(--font-weight-semibold); margin: 0; }
.attachment-upload__dropzone {
  border: 2px dashed var(--border-default); border-radius: var(--radius-md);
  padding: var(--spacing-6); text-align: center; cursor: pointer;
  transition: border-color 0.2s, background 0.2s;
}
.attachment-upload__dropzone:hover, .attachment-upload__dropzone--dragover {
  border-color: var(--color-primary); background: var(--surface-hover);
}
.attachment-upload__hint { font-size: var(--font-size-sm); color: var(--text-muted); }
.attachment-upload__list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: var(--spacing-2); }
.attachment-upload__item {
  display: flex; align-items: center; gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3); background: var(--surface-secondary);
  border-radius: var(--radius-sm);
}
.attachment-upload__filename { flex: 1; font-size: var(--font-size-sm); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.attachment-upload__size { font-size: var(--font-size-xs); color: var(--text-muted); }
.attachment-upload__remove {
  background: none; border: none; color: var(--text-muted); cursor: pointer;
  font-size: var(--font-size-md); padding: 0 var(--spacing-1);
}
.attachment-upload__error { font-size: var(--font-size-sm); color: var(--color-error); }
</style>
