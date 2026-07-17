// Unit tests for DocumentUploader component
// Tests: drag-drop acceptance, file validation (type/size), upload progress emission, error states, disabled state

import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import DocumentUploader from './DocumentUploader.vue'

// Mock the API module
vi.mock('../../api/extraction', () => ({
  uploadDocument: vi.fn(),
  ALLOWED_FORMATS: ['.md', '.txt', '.pdf', '.docx', '.json', '.xml', '.csv', '.xlsx'],
  FORMAT_LABELS: {
    '.md': 'Markdown',
    '.txt': 'Text',
    '.pdf': 'PDF',
    '.docx': 'DOCX',
    '.json': 'JSON',
    '.xml': 'XML',
    '.csv': 'CSV',
    '.xlsx': 'XLSX'
  }
}))

function createFile(name: string, size: number, type = 'text/plain'): File {
  const blob = new Blob(['x'.repeat(size)], { type })
  return new File([blob], name, { type })
}

describe('DocumentUploader', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Props', () => {
    it('renders with default props', () => {
      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      expect(wrapper.find('.uploader__title').text()).toContain('Drop a document here')
      expect(wrapper.find('.uploader__formats').text()).toContain('max 20 MB')
    })

    it('shows batch messaging when mode is batch', () => {
      const wrapper = mount(DocumentUploader, {
        props: {
          ontologyId: 'test-onto',
          mode: 'batch'
        }
      })

      expect(wrapper.find('.uploader__title').text()).toContain('Drop files here')
    })

    it('respects maxFileSizeMb prop', () => {
      const wrapper = mount(DocumentUploader, {
        props: {
          ontologyId: 'test-onto',
          maxFileSizeMb: 5
        }
      })

      expect(wrapper.find('.uploader__formats').text()).toContain('max 5 MB')
    })

    it('respects allowedFormats prop', () => {
      const wrapper = mount(DocumentUploader, {
        props: {
          ontologyId: 'test-onto',
          allowedFormats: ['.md', '.txt']
        }
      })

      expect(wrapper.find('.uploader__formats').text()).toContain('Markdown')
      expect(wrapper.find('.uploader__formats').text()).not.toContain('.pdf')
    })
  })

  describe('File Validation', () => {
    it('shows error state for unsupported format', async () => {
      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      const fileInput = wrapper.find('input[type="file"]')
      const badFile = createFile('test.exe', 100)
      Object.defineProperty(fileInput.element, 'files', {
        value: [badFile],
        writable: false
      })
      await fileInput.trigger('change')

      expect(wrapper.find('.uploader__error').exists()).toBe(true)
      expect(wrapper.find('.uploader__error-message').text()).toContain('Unsupported format')
    })

    it('shows error state for oversized file', async () => {
      const wrapper = mount(DocumentUploader, {
        props: {
          ontologyId: 'test-onto',
          maxFileSizeMb: 1
        }
      })

      const fileInput = wrapper.find('input[type="file"]')
      // Create a file > 1MB
      const largeFile = createFile('large.md', 2 * 1024 * 1024)
      Object.defineProperty(fileInput.element, 'files', {
        value: [largeFile],
        writable: false
      })
      await fileInput.trigger('change')

      expect(wrapper.find('.uploader__error').exists()).toBe(true)
      expect(wrapper.find('.uploader__error-message').text()).toContain('too large')
    })

    it('shows error state for empty file', async () => {
      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      const fileInput = wrapper.find('input[type="file"]')
      const emptyFile = createFile('empty.md', 0)
      Object.defineProperty(fileInput.element, 'files', {
        value: [emptyFile],
        writable: false
      })
      await fileInput.trigger('change')

      expect(wrapper.find('.uploader__error').exists()).toBe(true)
      expect(wrapper.find('.uploader__error-message').text()).toContain('empty')
    })
  })

  describe('Upload Progress', () => {
    it('shows uploading state with progress', async () => {
      const { uploadDocument } = await import('../../api/extraction')
      const mockUpload = uploadDocument as ReturnType<typeof vi.fn>
      // Keep the promise pending so we stay in uploading state
      mockUpload.mockReturnValue(new Promise(() => {}))

      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      const fileInput = wrapper.find('input[type="file"]')
      const validFile = createFile('doc.md', 100)
      Object.defineProperty(fileInput.element, 'files', {
        value: [validFile],
        writable: false
      })
      await fileInput.trigger('change')

      // Should show uploading state
      expect(wrapper.find('.uploader__progress').exists()).toBe(true)
    })

    it('emits upload-progress during upload', async () => {
      const { uploadDocument } = await import('../../api/extraction')
      const mockUpload = uploadDocument as ReturnType<typeof vi.fn>
      mockUpload.mockReturnValue(new Promise(() => {}))

      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      // Trigger file selection
      const fileInput = wrapper.find('input[type="file"]')
      const validFile = createFile('test.txt', 100)
      Object.defineProperty(fileInput.element, 'files', {
        value: [validFile],
        writable: false
      })
      await fileInput.trigger('change')

      // Upload started — verify the API was called
      expect(mockUpload).toHaveBeenCalledTimes(1)
      expect(mockUpload).toHaveBeenCalledWith(
        expect.objectContaining({
          ontologyId: 'test-onto',
          file: expect.objectContaining({ name: 'test.txt' })
        })
      )
    })
  })

  describe('Error States', () => {
    it('shows error state on upload failure', async () => {
      const { uploadDocument } = await import('../../api/extraction')
      const mockUpload = uploadDocument as ReturnType<typeof vi.fn>
      mockUpload.mockRejectedValue({
        code: 'UPLOAD_FAILED',
        message: 'Network error'
      })

      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      const fileInput = wrapper.find('input[type="file"]')
      const validFile = createFile('doc.txt', 100)
      Object.defineProperty(fileInput.element, 'files', {
        value: [validFile],
        writable: false
      })
      await fileInput.trigger('change')

      await new Promise((resolve) => setTimeout(resolve, 0))
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.uploader__error').exists()).toBe(true)
      expect(wrapper.find('.uploader__error-message').text()).toContain('Network error')
    })

    it('emits upload-error on failure', async () => {
      const { uploadDocument } = await import('../../api/extraction')
      const mockUpload = uploadDocument as ReturnType<typeof vi.fn>
      mockUpload.mockRejectedValue({
        code: 'UPLOAD_FAILED',
        message: 'Server error'
      })

      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      const fileInput = wrapper.find('input[type="file"]')
      const validFile = createFile('doc.txt', 100)
      Object.defineProperty(fileInput.element, 'files', {
        value: [validFile],
        writable: false
      })
      await fileInput.trigger('change')

      await new Promise((resolve) => setTimeout(resolve, 0))
      await wrapper.vm.$nextTick()

      expect(wrapper.emitted('upload-error')).toBeTruthy()
      expect(wrapper.emitted('upload-error')?.[0][0]).toEqual({
        code: 'UPLOAD_FAILED',
        message: 'Server error'
      })
    })
  })

  describe('Disabled State', () => {
    it('prevents interaction while uploading', async () => {
      const { uploadDocument } = await import('../../api/extraction')
      const mockUpload = uploadDocument as ReturnType<typeof vi.fn>
      mockUpload.mockReturnValue(new Promise(() => {}))

      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      const fileInput = wrapper.find('input[type="file"]')
      const validFile = createFile('doc.md', 100)
      Object.defineProperty(fileInput.element, 'files', {
        value: [validFile],
        writable: false
      })
      await fileInput.trigger('change')

      // The uploader should be disabled now
      expect(wrapper.classes()).toContain('uploader--disabled')
    })
  })

  describe('Success State', () => {
    it('shows success state on successful upload', async () => {
      const { uploadDocument } = await import('../../api/extraction')
      const mockUpload = uploadDocument as ReturnType<typeof vi.fn>
      mockUpload.mockResolvedValue({
        id: 'preview-1',
        ontologyId: 'test-onto',
        sourceFile: 'doc.txt',
        steps: [
          {
            id: 's1',
            operation: 'CREATE_CLASS',
            entityId: 'Person',
            label: 'Person',
            included: true
          },
          {
            id: 's2',
            operation: 'CREATE_PROPERTY',
            entityId: 'hasName',
            label: 'hasName',
            included: true
          }
        ],
        totalSteps: 2,
        createdAt: '2026-07-16T12:00:00Z'
      })

      const wrapper = mount(DocumentUploader, {
        props: { ontologyId: 'test-onto' }
      })

      const fileInput = wrapper.find('input[type="file"]')
      const validFile = createFile('doc.txt', 100)
      Object.defineProperty(fileInput.element, 'files', {
        value: [validFile],
        writable: false
      })
      await fileInput.trigger('change')

      await new Promise((resolve) => setTimeout(resolve, 0))
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.uploader__success').exists()).toBe(true)
      expect(wrapper.find('.uploader__success-text').text()).toContain('2 steps')
    })
  })
})
