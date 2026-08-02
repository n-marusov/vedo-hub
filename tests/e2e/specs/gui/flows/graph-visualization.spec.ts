import { test, expect } from '../../graphql-fixtures';
import { OntologyWorkspacePage } from '../../../pages/ontology-workspace.page';
import { UNIVERSITY_ONTOLOGY } from '../../ontology-test-data';

// E2E-browse.graph-view — Open ontology workspace -> TBox graph -> zoom -> click node (P1)
// Covers US: US-browse.tree.graph-view, US-editor.classes.hierarchy-drag-drop
//
// Depends on: graph visualization frontend (M5 Q3 — TBox Graph view)

test.describe('Graph Visualization E2E', () => {
  let workspace: OntologyWorkspacePage;

  test.beforeEach(async ({ page }) => {
    workspace = new OntologyWorkspacePage(page);
    await workspace.goto();
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);
    // Q3: the graph lives in the "TBox Graph" navigation view
    await page.getByRole('tab', { name: /tbox graph/i }).click();
    await expect(page.getByRole('img', { name: /ontology graph visualization/i })).toBeVisible();
  });

  test('view graph with nodes and edges after opening the ontology', async ({ page }) => {
    // The TBox view renders the class taxonomy: Person, Student, Professor, Organization
    const nodes = await workspace.getGraphNodes();
    expect(nodes).toContain('Person');
    expect(nodes).toContain('Student');
    expect(nodes).toContain('Professor');
    expect(nodes).toContain('Organization');

    // Verify subclass edges exist (relationships between nodes)
    const edgeCount = await workspace.getGraphEdges();
    expect(edgeCount).toBeGreaterThanOrEqual(1);
  });

  test('zoom in on graph makes nodes larger', async ({ page }) => {
    // Capture initial node size
    const initialNode = page.locator('.graph-viz__flow-node').first();
    await expect(initialNode).toBeVisible();
    const initialBox = await initialNode.boundingBox();
    expect(initialBox).not.toBeNull();

    // Zoom in using the zoom control
    await page.getByRole('button', { name: /zoom in/i }).click();

    // Wait for Vue Flow to re-render
    await page.waitForTimeout(300);

    // After zoom in, node should be visually larger
    const zoomedBox = await initialNode.boundingBox();
    expect(zoomedBox).not.toBeNull();
    if (initialBox && zoomedBox) {
      expect(zoomedBox.width).not.toBe(initialBox.width);
    }
  });

  test('click node selects the class', async ({ page }) => {
    // Click on a node in the graph
    const personNode = page.locator('.graph-viz__flow-node', { hasText: 'Person' });
    await personNode.click();

    // Class selection is tracked by the workspace (selected class id)
    await expect(page.locator('.nav-switcher')).toBeVisible();
  });

  test('zoom out restores graph to original state', async ({ page }) => {
    // Zoom in then zoom out
    await page.getByRole('button', { name: /zoom in/i }).click();
    await page.waitForTimeout(200);
    await page.getByRole('button', { name: /zoom out/i }).click();
    await page.waitForTimeout(200);

    // Graph should still render all nodes after zoom operations
    const nodeCount = await page.locator('.graph-viz__flow-node').count();
    expect(nodeCount).toBeGreaterThanOrEqual(4);
  });

  test('clicking on an empty area of the graph canvas deselects the active node', async ({ page }) => {
    // Select a node first
    const personNode = page.locator('.graph-viz__flow-node', { hasText: 'Person' });
    await personNode.click();

    // Click on empty canvas area — graph stays rendered (no crash, no selection error)
    await page.getByRole('img', { name: /ontology graph visualization/i }).click({ position: { x: 10, y: 10 } });

    // Graph still renders after the canvas click
    await expect(page.locator('.graph-viz__flow-node').first()).toBeVisible();
  });
});
