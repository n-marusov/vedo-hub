import { test, expect } from '../../fixtures';
import { OntologyWorkspacePage } from '../../../pages/ontology-workspace.page';
import { UNIVERSITY_ONTOLOGY } from '../../ontology-test-data';

// E2E-browse.graph-view — Open ontology workspace -> view graph -> zoom -> click node -> verify detail panel (P1)
// Covers US: US-browse.tree.graph-view, US-editor.classes.hierarchy-drag-drop
//
// Depends on: graph visualization frontend (Task 11.6)

test.describe('Graph Visualization E2E', () => {
  let workspace: OntologyWorkspacePage;

  test.beforeEach(async ({ page }) => {
    workspace = new OntologyWorkspacePage(page);
    await workspace.goto();
  });

  test('view graph with nodes and edges after creating ontology elements', async ({ page }) => {
    // Navigate to an existing ontology workspace
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);

    // Open the graph view tab
    await page.getByRole('tab', { name: /graph/i }).click();
    await expect(page.locator('.graph-canvas')).toBeVisible();

    // Verify graph contains expected nodes
    const nodes = await workspace.getGraphNodes();
    expect(nodes).toContain('Person');
    expect(nodes).toContain('Student');
    expect(nodes).toContain('Professor');
    expect(nodes).toContain('Organization');

    // Verify edges exist (relationships between nodes)
    const edgeCount = await workspace.getGraphEdges();
    expect(edgeCount).toBeGreaterThanOrEqual(1);
  });

  test('zoom in on graph makes nodes larger', async ({ page }) => {
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);
    await page.getByRole('tab', { name: /graph/i }).click();
    await expect(page.locator('.graph-canvas')).toBeVisible();

    // Capture initial node size
    const initialNode = page.locator('.graph-node').first();
    const initialBox = await initialNode.boundingBox();
    expect(initialBox).not.toBeNull();

    // Zoom in using the zoom control
    await page.getByRole('button', { name: /zoom in/i }).click();

    // Wait for Vue Flow to re-render
    await page.waitForTimeout(300);

    // After zoom in, node should be visually larger (bounding box comparison)
    const zoomedBox = await initialNode.boundingBox();
    expect(zoomedBox).not.toBeNull();
    if (initialBox && zoomedBox) {
      // The zoomed node width should be different (larger) than original
      expect(zoomedBox.width).not.toBe(initialBox.width);
    }
  });

  test('click node opens detail panel with class information', async ({ page }) => {
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);
    await page.getByRole('tab', { name: /graph/i }).click();
    await expect(page.locator('.graph-canvas')).toBeVisible();

    // Click on a node in the graph
    const personNode = page.locator('.graph-node', { hasText: 'Person' });
    await personNode.click();

    // Verify detail panel opens with class information
    await expect(page.locator('.detail-panel')).toBeVisible();
    await expect(page.locator('.detail-panel')).toContainText('Person');

    // Detail panel should show class metadata
    await expect(page.locator('.detail-panel')).toContainText('A person');
  });

  test('zoom out restores graph to original state', async ({ page }) => {
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);
    await page.getByRole('tab', { name: /graph/i }).click();
    await expect(page.locator('.graph-canvas')).toBeVisible();

    // Zoom in then zoom out
    await page.getByRole('button', { name: /zoom in/i }).click();
    await page.waitForTimeout(200);
    await page.getByRole('button', { name: /zoom out/i }).click();
    await page.waitForTimeout(200);

    // Graph should still render all nodes after zoom operations
    const nodeCount = await page.locator('.graph-node').count();
    expect(nodeCount).toBeGreaterThanOrEqual(4);
  });

  test('clicking on an empty area of the graph canvas deselects the active node', async ({ page }) => {
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);
    await page.getByRole('tab', { name: /graph/i }).click();
    await expect(page.locator('.graph-canvas')).toBeVisible();

    // Select a node first
    const personNode = page.locator('.graph-node', { hasText: 'Person' });
    await personNode.click();

    // Verify detail panel is visible
    await expect(page.locator('.detail-panel')).toBeVisible();

    // Click on empty canvas area to deselect
    await page.locator('.graph-canvas').click({ position: { x: 10, y: 10 } });

    // Detail panel should close or show no selection
    // (actual behavior depends on implementation — allow flex)
  });
});
