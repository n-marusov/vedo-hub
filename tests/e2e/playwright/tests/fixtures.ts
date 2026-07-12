import { test as base } from '@playwright/test';

type SeededUser = {
  username: string;
  role: 'owner' | 'editor' | 'viewer';
};

export const test = base.extend<{ seededUsers: SeededUser[] }>({
  seededUsers: async ({}, use) => {
    await use([
      { username: 'owner_seed', role: 'owner' },
      { username: 'editor_seed', role: 'editor' },
      { username: 'viewer_seed', role: 'viewer' }
    ]);
  }
});

export { expect } from '@playwright/test';
