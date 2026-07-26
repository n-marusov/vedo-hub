// @m4 — AnnotationDialog vitest spec
// Validates: REQ-USR.UI.gui-implementation
import { mountWithProviders, waitForQuery } from '@/__tests__/setup/mock-providers'
import AnnotationDialog from '@/components/ontology/AnnotationDialog.vue'
import { describe, expect, it } from 'vitest'
import { nextTick } from 'vue'

const TeleportStub = { template: '<div><slot /></div>' }

describe('AnnotationDialog', () => {
  it('should render when open is true', async () => {
    const wrapper = mountWithProviders(AnnotationDialog, {
      props: { open: true },
      global: { stubs: { Teleport: TeleportStub } }
    })
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.dialog-overlay').exists()).toBe(true)
    expect(wrapper.text()).toContain('Annotation')
  })

  it('should show property selector and value input', async () => {
    const wrapper = mountWithProviders(AnnotationDialog, {
      props: { open: true },
      global: { stubs: { Teleport: TeleportStub } }
    })
    await nextTick()
    expect(wrapper.find('.form-select').exists()).toBe(true)
    expect(wrapper.findAll('.form-input').length).toBeGreaterThanOrEqual(1)
  })

  it('should emit saved on successful save', async () => {
    const wrapper = mountWithProviders(AnnotationDialog, {
      props: { open: true },
      global: { stubs: { Teleport: TeleportStub } }
    })
    await nextTick()
    const select = wrapper.find('.form-select')
    await select.setValue('rdfs:label')
    const inputs = wrapper.findAll('.form-input')
    await inputs[0]?.setValue('My Label')
    const saveBtn = wrapper.find('.btn--primary')
    await saveBtn.trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 600))
    expect(wrapper.emitted('saved')).toBeTruthy()
  })

  it('should emit close on cancel', async () => {
    const wrapper = mountWithProviders(AnnotationDialog, {
      props: { open: true },
      global: { stubs: { Teleport: TeleportStub } }
    })
    await nextTick()
    const cancelBtn = wrapper.findAll('button').filter((b) => b.text().includes('Cancel'))
    expect(cancelBtn.length).toBeGreaterThanOrEqual(1)
    await cancelBtn[0]?.trigger('click')
    expect(wrapper.emitted('close')).toBeTruthy()
  })
})
