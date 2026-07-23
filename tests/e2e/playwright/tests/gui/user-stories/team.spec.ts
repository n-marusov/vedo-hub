// Validates: US-team.reviews.approve
// Validates: US-projects.fork
import { test, expect } from '@playwright/test';

test.describe.skip('Team & Projects — collaboration and project management', () => {
  test('US-team.reviews.approve: approve ontology merge request', async ({ page }) => {
    // TODO: Open MR → review changes → approve → verify MR status
  });

  test('US-projects.fork: fork an ontology project', async ({ page }) => {
    // TODO: Open project → fork → verify new fork appears in user projects
  });
});
