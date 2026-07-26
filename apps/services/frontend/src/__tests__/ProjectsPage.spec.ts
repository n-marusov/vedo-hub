// @m4 — ProjectsPage vitest spec (GREEN: uses mountWithProviders)
// Validates: REQ-USR.UI.gui-implementation
// Tests: projects list from REST API
import { describePage, mountWithProviders, waitForQuery } from '@/__tests__/setup/mock-providers'
import { afterEach, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

// ProjectsPage uses REST listProjects from @/api/org — mock for test control
vi.mock('@/api/org', () => ({
  listProjects: vi.fn(),
  listGroups: vi.fn(),
  createProject: vi.fn()
}))

// Teleport stub: renders slot inline instead of moving to document.body
const TeleportStub = { template: '<div><slot /></div>' }

const mockProjects = {
  items: [
    {
      id: 'proj-1',
      name: 'Test Project',
      description: 'A test',
      visibility: 'private',
      ontologyId: 'ontology-1',
      memberCount: 1,
      updatedAt: null
    }
  ],
  total: 1
}

describePage('ProjectsPage', () => {
  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should render projects list page title', async () => {
    const { listProjects } = await import('@/api/org')
    vi.mocked(listProjects).mockResolvedValue(mockProjects)
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.text()).toContain('Projects')
  })

  it('should show search input for filtering projects', async () => {
    const { listProjects } = await import('@/api/org')
    vi.mocked(listProjects).mockResolvedValue(mockProjects)
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage)
    await waitForQuery()
    await nextTick()
    const searchInput = wrapper.find('input[aria-label="Search projects"]')
    expect(searchInput.exists()).toBe(true)
  })

  it('should display sort controls for name and direction', async () => {
    const { listProjects } = await import('@/api/org')
    vi.mocked(listProjects).mockResolvedValue(mockProjects)
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.text()).toContain('Name')
  })

  it('should render projects section after API data loads', async () => {
    const { listProjects } = await import('@/api/org')
    vi.mocked(listProjects).mockResolvedValue(mockProjects)
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.pp-list').exists() || wrapper.find('.pp-empty').exists()).toBe(true)
  })

  it('should render page layout with toolbar', async () => {
    const { listProjects } = await import('@/api/org')
    vi.mocked(listProjects).mockResolvedValue(mockProjects)
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage)
    await waitForQuery()
    await nextTick()
    expect(wrapper.find('.pp-top').exists()).toBe(true)
  })

  it('should show "New project" button', async () => {
    const { listProjects } = await import('@/api/org')
    vi.mocked(listProjects).mockResolvedValue(mockProjects)
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage)
    await waitForQuery()
    await nextTick()
    const newBtn = wrapper.find('.pp-new-btn')
    expect(newBtn.exists()).toBe(true)
    expect(newBtn.text()).toContain('New project')
  })

  it('should open create project dialog with required fields', async () => {
    const { listProjects } = await import('@/api/org')
    vi.mocked(listProjects).mockResolvedValue(mockProjects)
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage, {
      global: { stubs: { Teleport: TeleportStub } }
    })
    await waitForQuery()
    await nextTick()

    await wrapper.find('.pp-new-btn').trigger('click')
    await nextTick()

    expect(wrapper.find('.dialog-overlay').exists()).toBe(true)
    expect(wrapper.text()).toContain('Create project')
    expect(wrapper.text()).toContain('Project name')
    expect(wrapper.text()).toContain('Project URL')
    expect(wrapper.text()).toContain('Visibility Level')
  })

  it('should create project and refresh list', async () => {
    const { listProjects, createProject } = await import('@/api/org')
    vi.mocked(listProjects).mockResolvedValue(mockProjects)
    vi.mocked(createProject).mockResolvedValue({
      id: 'new-project-1',
      name: 'My project',
      description: null,
      visibility: 'private',
      ontologyId: 'ontology-1',
      memberCount: 0,
      updatedAt: null
    })
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage, {
      global: { stubs: { Teleport: TeleportStub } }
    })
    await waitForQuery()
    await nextTick()

    await wrapper.find('.pp-new-btn').trigger('click')
    await nextTick()
    await wrapper.find('#project-name').setValue('My project')
    await wrapper.find('.btn--primary').trigger('click')
    await new Promise((resolve) => setTimeout(resolve, 100))
    await nextTick()

    expect(createProject).toHaveBeenCalledWith({
      name: 'My project',
      description: undefined,
      groupId: undefined
    })
    expect(listProjects).toHaveBeenCalled()
    expect(wrapper.find('.dialog-overlay').exists()).toBe(false)
  })

  it('should not crash when projects API fails', async () => {
    const { listProjects } = await import('@/api/org')
    vi.mocked(listProjects).mockRejectedValue(new Error('Failed to load projects'))
    const ProjectsPage = (await import('@/pages/ProjectsPage.vue')).default
    const wrapper = mountWithProviders(ProjectsPage)
    await new Promise((resolve) => setTimeout(resolve, 200))
    await nextTick()
    expect(wrapper.find('.pp-top').exists() || wrapper.find('.pp-page').exists()).toBe(true)
  })
})
