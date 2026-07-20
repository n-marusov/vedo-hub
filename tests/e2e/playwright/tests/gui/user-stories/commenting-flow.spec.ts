import { test, expect } from '../../fixtures';
import { OntologyWorkspacePage } from '../../../pages/ontology-workspace.page';
import { UNIVERSITY_ONTOLOGY } from '../../ontology-test-data';

// E2E-commenting.flow — Add comment to class -> view in feed -> reply -> verify notification (P1)
// Covers US: US-team.comments.feed, US-team.comments.mention, US-team.comments.visibility
//
// Depends on: commenting-service (Task 10.1), frontend commenting UI (Task 11.9)

// @skip — commenting-service and API-level comment flows exist, but the current
// frontend workspace only exposes navigation/count placeholders; no entity comment
// panel/input/reply UI (`comments-panel`, `comment-input`, `comment-feed`) is wired.
// Roadmap: M5 still lists entity/project comments as an MVP gap; M10 expands this
// to full threaded collaboration/review workflows.
// AI-agent note: wire comments UI to the existing commenting service before unskipping.
test.describe.skip('Commenting Flow E2E', () => {
  let workspace: OntologyWorkspacePage;

  test.beforeEach(async ({ page }) => {
    workspace = new OntologyWorkspacePage(page);
    await workspace.goto();
  });

  test('add comment to class, view in feed, reply, verify notification', async ({ page }) => {
    // Open an existing ontology
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);

    // Select a class to comment on
    await workspace.selectClass('Person');
    await expect(page.locator('.class-tree-item.selected')).toContainText('Person');

    // Open the comments panel
    await page.getByRole('button', { name: /comments/i }).click();
    await expect(page.locator('.comments-panel')).toBeVisible();

    // Add a comment to the selected class
    const commentText = 'Need to add address property';
    await page.locator('.comment-input').fill(commentText);
    await page.getByRole('button', { name: /submit comment/i }).click();

    // Verify the comment appears in the comment feed
    await expect(page.locator('.comment-feed')).toContainText(commentText);
    await expect(page.locator('.comment-item')).toContainText(commentText);

    // Verify the author is shown
    await expect(page.locator('.comment-author')).toHaveCount(1);
    await expect(page.locator('.comment-author')).toContainText('test-user');

    // Add a reply to the comment
    const replyText = 'Agreed, adding address field';
    await page.locator('.comment-item').first().hover();
    await page.getByRole('button', { name: /reply/i }).click();
    await page.locator('.reply-input').fill(replyText);
    await page.getByRole('button', { name: /submit reply/i }).click();

    // Verify the reply is displayed under the original comment
    await expect(page.locator('.comment-replies')).toContainText(replyText);
    const replies = await page.locator('.comment-reply').allTextContents();
    expect(replies.some(r => r.includes(replyText))).toBeTruthy();

    // Open the ontology-wide activity feed
    await page.getByRole('button', { name: /activity feed/i }).click();
    await expect(page.locator('.activity-feed')).toBeVisible();

    // Verify both comment and reply appear in the feed in chronological order
    const feedItems = await page.locator('.feed-item').allTextContents();
    const feedText = feedItems.join(' ');
    expect(feedText).toContain(commentText);
    expect(feedText).toContain(replyText);

    // Verify comment timestamps are present
    const timestamps = await page.locator('.comment-timestamp').allTextContents();
    expect(timestamps.length).toBeGreaterThanOrEqual(2);
  });

  test('comments are scoped to the selected ontology entity', async ({ page }) => {
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);

    // Select different entities and verify comment scoping
    await workspace.selectClass('Organization');
    await page.getByRole('button', { name: /comments/i }).click();
    await page.locator('.comment-input').fill('Comment for Organization');
    await page.getByRole('button', { name: /submit comment/i }).click();

    await workspace.selectClass('Person');
    await expect(page.locator('.comment-feed')).not.toContainText('Comment for Organization');
  });

  test('empty comment validation', async ({ page }) => {
    await workspace.openOntology(UNIVERSITY_ONTOLOGY.name);
    await workspace.selectClass('Person');
    await page.getByRole('button', { name: /comments/i }).click();

    // Try submitting an empty comment
    await page.getByRole('button', { name: /submit comment/i }).click();
    await expect(page.locator('.validation-error')).toBeVisible();
    await expect(page.locator('.validation-error')).toContainText('cannot be empty');
  });
});
