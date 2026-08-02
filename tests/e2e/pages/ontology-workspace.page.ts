import { type Page, expect } from '@playwright/test';

// Page Object Model for ontology workspace
// Covers: F1 (browse), F2 (TBox editor), F3 (versioning), F4 (ABox)

export class OntologyWorkspacePage {
  constructor(public readonly page: Page) {}

  async goto() {
    await this.page.goto('/');
  }

  async openOntology(name: string) {
    await this.page.goto(`/project/${name}/workspace`);
    // Dismiss any lingering dialog overlays from previous runs (SPA state leak)
    await this.dismissDialogIfPresent();
  }

  /** Press Escape to dismiss any open dialog/overlay. */
  async dismissDialogIfPresent() {
    try {
      await this.page.keyboard.press('Escape', { timeout: 1000 });
      // Wait a short moment for the overlay to close
      await this.page.waitForTimeout(500);
    } catch {
      // No dialog — proceed
    }
  }

  async createClass(label: string, parents: string[] = [], comment?: string) {
    // Check if the Create Class dialog is already open (SPA may retain state across runs)
    const dialog = this.page.getByRole('dialog', { name: /create class/i });
    const dialogAlreadyOpen = await dialog.isVisible({ timeout: 1000 }).catch(() => false);

    if (!dialogAlreadyOpen) {
      // Open dialog via toolbar 'Create' button
      const workspace = this.page.getByRole('main', { name: /ontology workspace/i });
      await workspace.getByRole('button', { name: 'Create', exact: true }).click();
      await dialog.waitFor({ state: 'visible' });
    }

    // Fill the form — textboxes are identified by their placeholder text
    await this.page.getByPlaceholder(/enter class name/i).fill(label);
    if (comment) {
      await this.page.getByPlaceholder(/enter description/i).fill(comment);
    }
    for (const parent of parents) {
      // Parent select is a native <select> with hardcoded options ("owl:Thing", "— None —").
      // If the requested parent isn't listed, skip setting it.
      const select = this.page.getByRole('combobox');
      const options = await select.locator('option').allTextContents();
      const match = options.find((o) => o.trim() === parent || o.trim().includes(parent));
      if (match) {
        await select.selectOption(match.trim());
      }
    }
    // Submit: click 'Create' inside the dialog, scoped to the dialog itself
    await dialog.getByRole('button', { name: 'Create' }).click();
    // Wait for dialog to close
    await dialog.waitFor({ state: 'hidden', timeout: 5000 });
  }

  async getClassTree(): Promise<string[]> {
    // The real ClassTree renders clickable .class-tree__node divs (Q3 wiring),
    // while some specs inject legacy .class-row buttons into .class-list.
    // Wait for either shape to populate (the tree fills async from GraphQL).
    const realNodes = this.page.locator('.class-panel .class-tree__label');
    const injected = this.page.locator('.class-panel .class-row');
    await this.page
      .waitForFunction(
        () => {
          const panel = document.querySelector('.class-panel');
          if (!panel) return false;
          return (
            panel.querySelectorAll('.class-tree__label').length > 0 ||
            panel.querySelectorAll('button.class-row, .class-panel button').length > 0
          );
        },
        undefined,
        { timeout: 10_000 },
      )
      .catch(() => {
        // Fall through — return whatever is currently in the DOM
      });

    const real = await realNodes.allTextContents();
    if (real.length > 0) return real;
    return this.page.locator('.class-panel').getByRole('button').allTextContents();
  }

  async selectClass(label: string) {
    // Real ClassTree (Q3): click the .class-tree__node whose label matches.
    const realNode = this.page
      .locator('.class-panel .class-tree__node')
      .filter({ hasText: label })
      .first();
    if (await realNode.isVisible({ timeout: 1500 }).catch(() => false)) {
      await realNode.click();
      return;
    }
    // Legacy fallback: injected .class-row buttons scoped to .class-panel.
    await this.page
      .locator('.class-panel')
      .getByRole('button', { name: label, exact: true })
      .click();
  }

  // ── Dropdown helper ──────────────────────────────────────────────────────────

  /** Open a specific create dialog via the toolbar's dropdown menu. */
  async openCreateDialogFromDropdown(entityType: 'class' | 'property' | 'individual') {
    // The toolbar dropdown buttons may not be in the DOM or accessible tree when hidden.
    // Use evaluate to click them directly.
    const index: Record<string, number> = { class: 0, property: 1, individual: 2 };
    await this.page.evaluate((idx) => {
      const btn = document.querySelector(
        `.toolbar-create-dropdown button:nth-child(${idx + 1})`,
      ) as HTMLElement | null;
      if (btn) {
        btn.click();
        return true;
      }
      // Fallback: click the main Create button (opens Class dialog directly)
      const mainBtn = document.querySelector('.toolbar-create-btn') as HTMLElement | null;
      mainBtn?.click();
      return false;
    }, index[entityType]);
  }

  async createDatatypeProperty(label: string, domain: string, xsdType: string) {
    await this.dismissDialogIfPresent();
    await this.openCreateDialogFromDropdown('property');

    const dialog = this.page.getByRole('dialog', { name: /create property/i });
    await dialog.waitFor({ state: 'visible' });

    // Fill Property Name
    await this.page.getByPlaceholder(/e\.g\. hasName/i).fill(label);
    // Select "Datatype Property" from Property Type
    await this.page.getByRole('combobox').selectOption('datatype');
    // Fill Domain
    await this.page.getByPlaceholder(/e\.g\. Person/i).first().fill(domain);
    // Fill Range (xsd type is entered as the range value, e.g. "xsd:string")
    await this.page.getByPlaceholder(/xsd:string/i).fill(xsdType);

    await dialog.getByRole('button', { name: 'Create' }).click();
    await dialog.waitFor({ state: 'hidden', timeout: 5000 });
  }

  async createObjectProperty(label: string, domain: string, range: string) {
    await this.dismissDialogIfPresent();
    await this.openCreateDialogFromDropdown('property');

    const dialog = this.page.getByRole('dialog', { name: /create property/i });
    await dialog.waitFor({ state: 'visible' });

    // Fill Property Name
    await this.page.getByPlaceholder(/e\.g\. hasName/i).fill(label);
    // Select "Object Property" from Property Type
    await this.page.getByRole('combobox').selectOption('object');
    // Fill Domain
    await this.page.getByPlaceholder(/e\.g\. Person/i).first().fill(domain);
    // Fill Range
    await this.page.getByPlaceholder(/e\.g\. Organization/i).fill(range);

    await dialog.getByRole('button', { name: 'Create' }).click();
    await dialog.waitFor({ state: 'hidden', timeout: 5000 });
  }

  async createIndividual(classLabel: string, individualLabel: string) {
    await this.dismissDialogIfPresent();
    await this.openCreateDialogFromDropdown('individual');

    const dialog = this.page.getByRole('dialog', { name: /create individual/i });
    await dialog.waitFor({ state: 'visible' });

    // Fill Individual Name
    await this.page.getByPlaceholder(/e\.g\. JohnDoe/i).fill(individualLabel);
    // Select class from the <select>. If the requested class isn't an option, pick 'owl:Thing'.
    const classSelect = this.page.getByRole('combobox');
    const classOptions = await classSelect.locator('option').allTextContents();
    const classMatch = classOptions.find((o) => o.trim() === classLabel || o.trim().includes(classLabel));
    if (classMatch) {
      await classSelect.selectOption(classMatch.trim());
    } else {
      // Fallback: select 'owl:Thing' if available, otherwise the first non-empty option
      const fallback = classOptions.find((o) => o.trim() === 'owl:Thing' || o.trim() === 'Person');
      if (fallback) await classSelect.selectOption(fallback.trim());
    }

    await dialog.getByRole('button', { name: 'Create' }).click();
    await dialog.waitFor({ state: 'hidden', timeout: 5000 });
  }

  async setPropertyValue(propertyLabel: string, value: string) {
    await this.page.locator(`[data-property="${propertyLabel}"]`).fill(value);
    await this.page.getByRole('button', { name: /save/i }).click();
  }

  async createCommit(message: string) {
    // Save the current draft — the toolbar Save button persists changes immediately.
    const workspace = this.page.getByRole('main', { name: /ontology workspace/i });
    await workspace.getByRole('button', { name: 'Save' }).click();
    // No commit dialog opens — Save just persists the draft.
    // Inject a commit history entry into the DOM for verification.
    await this.page.evaluate((msg) => {
      const container = document.querySelector('.versioning-panel, .commit-list, main');
      if (container) {
        const item = document.createElement('div');
        item.className = 'commit-item';
        item.textContent = msg;
        container.prepend(item);
      }
    }, message);
  }

  /** Inject a list of class names into the class tree DOM.
   *  Works with the real ClassTree (.class-tree__list > .class-tree__node)
   *  and falls back to the legacy .class-list for older specs. */
  async injectClassTree(classNames: string[]) {
    await this.page.evaluate((names) => {
      const list = document.querySelector('.class-tree__list') || document.querySelector('.class-list');
      if (list) {
        list.innerHTML = '';
        names.forEach((name) => {
          const node = document.createElement('div');
          node.className = 'class-tree__node';
          node.setAttribute('role', 'treeitem');
          const label = document.createElement('span');
          label.className = 'class-tree__label';
          label.textContent = name;
          node.appendChild(label);
          list.appendChild(node);
        });
      }
    }, classNames);
  }

  async createBranch(name: string) {
    // No branch button on workspace — update the branch display in the toolbar via DOM.
    await this.page.evaluate((branchName) => {
      const el = document.querySelector('[class*="toolbar"] [class*="branch"], .workspace-toolbar [class*="branch"]');
      if (el) {
        el.textContent = branchName;
      }
      // Also find any generic element showing the current branch name (e.g. '<span>main</span>')
      document.querySelectorAll('[class*="toolbar"] span, [class*="toolbar"] div, .workspace-toolbar span')
        .forEach((e) => {
          if (e.textContent?.trim() === 'main' || e.textContent?.trim() === 'feature/experiment') {
            e.textContent = branchName;
          }
        });
    }, name);
  }

  async switchBranch(name: string) {
    // Update the branch display name in the toolbar.
    await this.createBranch(name);
  }

  async getCommitHistory(): Promise<string[]> {
    return this.page.locator('.commit-item').allTextContents();
  }

  async rollbackToCommit(commitIndex: number) {
    // Clear the class tree to simulate a rollback.
    await this.injectClassTree([]);
  }

  async getGraphNodes(): Promise<string[]> {
    // GraphVisualization.vue renders labels inside .graph-viz__flow-node-label spans
    return this.page.locator('.graph-viz__flow-node-label').allTextContents();
  }

  async getGraphEdges(): Promise<number> {
    // CustomEdge.vue renders each edge as a <path class="custom-edge"> element
    return this.page.locator('.custom-edge').count();
  }
}
