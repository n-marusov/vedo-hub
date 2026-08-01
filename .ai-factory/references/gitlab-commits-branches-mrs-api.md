# GitLab Commits, Branches & Merge Requests API Reference

> Source: https://docs.gitlab.com/api/commits.html, https://docs.gitlab.com/api/branches.html, https://docs.gitlab.com/api/merge_requests.html
> Created: 2026-08-01
> Updated: 2026-08-01

## Overview

GitLab REST API v4 for managing Git repository operations: commits, branches, and merge requests. All endpoints use JSON over HTTP with `PRIVATE-TOKEN` or `Authorization: Bearer <token>` authentication. **Base URL:** `https://gitlab.example.com/api/v4`. **Tier:** Free, Premium, Ultimate (some endpoints Premium/Ultimate only). **Project ID** can be numeric `id` or URL-encoded path (`group%2Fproject`).

## Core Concepts

- **Commit**: A snapshot of changes in the repository. Identified by full 40-char SHA or short SHA.
- **Branch**: A named pointer to a commit. `main` is typically the default. Can be protected, merged, or default.
- **Merge Request (MR)**: A request to merge changes from a source branch into a target branch. Identified by `id` (global) and `iid` (project-scoped internal ID).
- **MR IID vs ID**: `iid` is the number you see in the UI (`!42`). `id` is a global unique identifier across all projects. Use `iid` with project-scoped endpoints, `id` for global endpoints.
- **Draft / WIP**: `draft` replaces deprecated `wip` (GitLab 19.0+). Use `draft=true/false` to filter.
- **Merge Status**: `detailed_merge_status` (new) replaces deprecated `merge_status` (GitLab 15.6+). Values include: `mergeable`, `checking`, `conflict`, `need_rebase`, `not_approved`, `ci_must_pass`, `ci_still_running`, `discussions_not_resolved`, `draft_status`, `not_open`, `preparing`, `requested_changes`, `merge_request_blocked`, `merge_time`, `security_policy_violations`, `status_checks_must_pass`, `locked_paths`, `locked_lfs_files`, `title_regex`, etc.
- **Merge User**: `merge_user` replaces deprecated `merged_by` (GitLab 14.7). Use `merge_user` field.
- **Auto-merge**: `auto_merge` replaces deprecated `merge_when_pipeline_succeeds` (GitLab 17.11).

---

## Commits API

### List repository commits

```
GET /projects/:id/repository/commits
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer/string | Yes | Project ID or URL-encoded path |
| `ref_name` | string | No | Branch, tag, or revision range (default: default branch) |
| `path` | string | No | Filter by file path |
| `author` | string | No | Filter by commit author |
| `since` | string (ISO 8601) | No | Commits on or after date (`YYYY-MM-DDTHH:MM:SSZ`) |
| `until` | string (ISO 8601) | No | Commits on or before date |
| `all` | boolean | No | Retrieve every commit (ignores `ref_name`) |
| `first_parent` | boolean | No | Follow only first parent on merge commits |
| `follow` | boolean | No | Follow file renames (default: `true`, only with `path` for single files) |
| `order` | string | No | `default` (reverse chronological) or `topo` |
| `trailers` | boolean | No | Parse and include Git trailers |
| `with_stats` | boolean | No | Include stats (additions, deletions, total) |

Response: array of commits with `id`, `short_id`, `title`, `message`, `author_name`, `author_email`, `authored_date`, `committer_name`, `committer_email`, `committed_date`, `created_at`, `parent_ids`, `web_url`, `trailers`, `extended_trailers`.

**Note:** Pagination headers `x-total` and `x-total-pages` are NOT returned for this endpoint (see GitLab issue 389582).

### Create a commit (batch actions)

```
POST /projects/:id/repository/commits
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer/string | Yes | Project ID |
| `branch` | string | Yes | Target branch. To create a new branch, also provide `start_branch` or `start_sha`. |
| `commit_message` | string | Yes | Commit message |
| `actions[]` | array | No | Array of action hashes (see below) |
| `author_name` | string | No | Override commit author name |
| `author_email` | string | No | Override commit author email |
| `force` | boolean | No | Force-overwrite branch history (requires `start_branch` or `start_sha`) |
| `start_branch` | string | No | Parent branch (mutually exclusive with `start_sha`) |
| `start_sha` | string | No | Parent commit SHA (full 40-char, mutually exclusive with `start_branch`) |
| `start_project` | integer/string | No | Source project for cross-project commits |
| `allow_empty` | boolean | No | Allow empty commit (GitLab 18.8+) |
| `stats` | boolean | No | Include stats (default: `true`) |

**Actions array elements:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action` | string | Yes | `create`, `delete`, `move`, `update`, `chmod` |
| `file_path` | string | Yes | Full file path (`lib/class.rb`) |
| `content` | string | No | File content (required except for `delete`, `chmod`, `move`) |
| `encoding` | string | No | `text` (default) or `base64` |
| `previous_path` | string | No | Original path for `move` action |
| `last_commit_id` | string | No | Last known commit ID (for update/move/delete) |
| `execute_filemode` | boolean | No | Enable/disable execute flag (for `chmod`) |

**Limits:** Max 300 MB per request; >20 MB rate-limited to 3 requests/30 sec.

Returns `201 Created` with commit object including `stats` (additions, deletions, total) and `status`.

### Retrieve a single commit

```
GET /projects/:id/repository/commits/:sha
```

`sha` can be commit hash, branch name, or tag name. Supports `stats` (boolean) parameter. Returns commit object with `last_pipeline` (object) and `stats`.

### List references a commit is pushed to

```
GET /projects/:id/repository/commits/:sha/refs
```

Parameters: `type` = `branch`, `tag`, or `all` (default). Returns array of `{ type, name }`.

### Get commit sequence number

```
GET /projects/:id/repository/commits/:sha/sequence
```

Optional: `first_parent` (boolean). Returns `{ count: <integer> }` — equivalent to `git rev-list --count`.

### Cherry-pick a commit

```
POST /projects/:id/repository/commits/:sha/cherry_pick
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer/string | Yes | Project ID |
| `sha` | string | Yes | Commit SHA to cherry-pick |
| `branch` | string | Yes | Target branch |
| `dry_run` | boolean | No | Test without committing (default: `false`) |
| `message` | string | No | Custom commit message |

Error codes: `empty` (changeset already exists), `conflict` (merge conflict).

### Revert a commit

```
POST /projects/:id/repository/commits/:sha/revert
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer/string | Yes | Project ID |
| `sha` | string | Yes | Commit SHA to revert |
| `branch` | string | Yes | Target branch |
| `dry_run` | boolean | No | Test without committing (default: `false`) |

Error codes: `empty` (already reverted), `conflict`.

### Get commit diff

```
GET /projects/:id/repository/commits/:sha/diff
```

Parameter: `unidiff` (boolean, default `false`). Response: array of `{ diff, old_path, new_path, a_mode, b_mode, new_file, renamed_file, deleted_file, collapsed, too_large }`.

`collapsed` / `too_large` introduced in GitLab 18.4. Subject to diff limits.

### Commit comments

```
GET    /projects/:id/repository/commits/:sha/comments          # List
POST   /projects/:id/repository/commits/:sha/comments          # Create
```

POST parameters: `note` (required), `path` (file path), `line` (line number), `line_type` (`new` or `old`).

### Commit discussions

```
GET /projects/:id/repository/commits/:sha/discussions
```

Returns array of `{ id, individual_note, notes[] }`.

### Commit statuses

```
GET  /projects/:id/repository/commits/:sha/statuses    # List
POST /projects/:id/statuses/:sha                       # Set pipeline status
```

**List parameters:** `ref`, `stage`, `name`, `all` (boolean), `pipeline_id`, `order_by` (`id`/`pipeline_id`), `sort` (`asc`/`desc`).

**Set parameters:** `state` (required: `pending`, `running`, `success`, `failed`, `canceled`, `skipped`), `name`/`context`, `ref`, `target_url`, `description`, `coverage` (float), `pipeline_id`.

Creates or updates a commit status in an `external` stage. Returns `409` if update already in progress.

### List MRs associated with a commit

```
GET /projects/:id/repository/commits/:sha/merge_requests
```

Parameter: `state` = `opened`, `closed`, `locked`, `merged` (GitLab 18.2+). Returns MR objects.

### Get commit signature

```
GET /projects/:id/repository/commits/:sha/signature
```

Returns signature info for signed commits. `404` for unsigned. Response includes `signature_type` (`PGP`, `SSH`, `X509`), `verification_status`, and type-specific key/cert details.

---

## Branches API

### List branches

```
GET /projects/:id/repository/branches
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer/string | Yes | Project ID |
| `search` | string | No | Substring search. Use `^term` for prefix, `term$` for suffix. |
| `regex` | string | No | RE2 regular expression (mutually exclusive with `search`) |

Response: array of `{ name, merged, protected, default, developers_can_push, developers_can_merge, can_push, web_url, commit: {...} }`.

Accessible without authentication for public repositories.

### Get single branch

```
GET /projects/:id/repository/branches/:branch
```

Returns single branch object. `:branch` must be URL-encoded.

### Create branch

```
POST /projects/:id/repository/branches
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer/string | Yes | Project ID |
| `branch` | string | Yes | New branch name |
| `ref` | string | Yes | Source branch name or commit SHA |

Returns `201 Created`.

### Delete branch

```
DELETE /projects/:id/repository/branches/:branch
```

Cannot delete default branch or protected branches. Returns `204 No Content`.

### Delete all merged branches

```
DELETE /projects/:id/repository/merged_branches
```

Deletes all branches merged into the default branch. Protected branches are skipped. Returns `202 Accepted`.

### Protected branches

See dedicated endpoints at:
- `POST   /projects/:id/protected_branches` — protect
- `DELETE /projects/:id/protected_branches/:name` — unprotect

Not covered in detail here; see [Protected Branches API docs](https://docs.gitlab.com/api/protected_branches.html).

---

## Merge Requests API

MRs are identified by `iid` (project-scoped) in most project endpoints, and by `id` (global) in cross-project contexts.

### List merge requests (global)

```
GET /merge_requests
```

Default scope: `created_by_me`. Use `scope=all` for all accessible MRs.

### List project merge requests

```
GET /projects/:id/merge_requests
```

Default scope: `all` (unlike global endpoint which defaults to `created_by_me`).

### List group merge requests

```
GET /groups/:id/merge_requests
```

Includes subgroups. Default `non_archived=true` (only non-archived projects).

### Common filter parameters (all list endpoints)

| Parameter | Type | Description |
|-----------|------|-------------|
| `state` | string | `opened`, `closed`, `locked`, `merged`, `all` |
| `scope` | string | `created_by_me`, `assigned_to_me`, `reviews_for_me`, `all` |
| `author_id` / `author_username` | integer/string | Filter by author (mutually exclusive) |
| `assignee_id` / `assignee_username[]` | integer/string[] | Filter by assignee. `None` = unassigned, `Any` = any assignee |
| `reviewer_id` / `reviewer_username` | integer/string | Filter by reviewer. `None`/`Any` supported |
| `merge_user_id` / `merge_user_username` | integer/string | Filter by merger (GitLab 17.0+) |
| `labels` | string | Comma-separated. `None` = no labels, `Any` = any label |
| `milestone` | string | Milestone title. `None`/`Any` supported |
| `source_branch` / `target_branch` | string | Filter by branch name |
| `search` | string | Search in title/description |
| `in` | string | Search scope: `title`, `description`, or `title,description` (default) |
| `draft` / `wip` | boolean/string | Draft status. `draft` (boolean) preferred over deprecated `wip` (GitLab 19.0+) |
| `created_after` / `created_before` | ISO 8601 | Date range for creation |
| `updated_after` / `updated_before` | ISO 8601 | Date range for updates |
| `merged_after` / `merged_before` | ISO 8601 | Date range for merge |
| `deployed_after` / `deployed_before` | ISO 8601 | Date range for deployment |
| `environment` | string | Filter by deployment environment |
| `order_by` | string | `created_at` (default), `updated_at`, `merged_at` (GitLab 17.2+), `label_priority`, `priority`, `milestone_due`, `popularity`, `title` |
| `sort` | string | `asc` or `desc` (default) |
| `view` | string | `simple` returns `{ iid, web_url, title, description, state }` only |
| `not` | hash | Negation filter: `labels`, `milestone`, `author_id`, `author_username`, `assignee_id`, `assignee_username`, `reviewer_id`, `reviewer_username`, `my_reaction_emoji` |
| `with_labels_details` | boolean | Include label details (name, color, description, text_color) |

Project-specific additional: `iids[]` (integer array) to filter by MR IIDs.
Group-specific additional: `non_archived` (boolean, default `true`), `source_project_id`.

### Get single merge request

```
GET /projects/:id/merge_requests/:merge_request_iid
```

Additional parameters: `include_diverged_commits_count` (boolean), `include_rebase_in_progress` (boolean), `render_html` (boolean).

Returns full MR object including `head_pipeline`, `diff_refs`, `merge_error`, `diverged_commits_count`, `rebase_in_progress`, `first_contribution`, `user.can_merge`, `subscribed`, `changes_count`.

### Key MR response fields

| Field | Type | Description |
|-------|------|-------------|
| `id` / `iid` | integer | Global ID / project-scoped internal ID |
| `project_id` | integer | Target project ID |
| `title` / `description` | string | Title and description (Markdown, limited to 1,048,576 chars) |
| `state` | string | `opened`, `closed`, `merged`, `locked` |
| `draft` | boolean | Draft status (use instead of `work_in_progress`) |
| `source_branch` / `target_branch` | string | Branch names |
| `source_project_id` / `target_project_id` | integer | Project IDs (differ for fork MRs) |
| `sha` | string | HEAD commit SHA of source branch |
| `merge_commit_sha` / `squash_commit_sha` | string | null until merged |
| `detailed_merge_status` | string | Detailed mergeability status (preferred over `merge_status`) |
| `merge_user` | object | Who merged / set auto-merge (use instead of `merged_by`) |
| `merge_after` | dateTime | Timestamp after which MR can be merged (GitLab 17.8+) |
| `merged_at` / `closed_at` / `created_at` / `updated_at` | dateTime | Timestamps |
| `prepared_at` | dateTime | When preparation steps completed (populates once) |
| `allow_collaboration` | boolean | Allow commits from target branch members (forks) |
| `squash` / `squash_on_merge` | boolean | Squash commits on merge |
| `force_remove_source_branch` | boolean | Project setting for source branch deletion |
| `should_remove_source_branch` | boolean | MR-specific source branch deletion flag |
| `has_conflicts` / `blocking_discussions_resolved` | boolean | Conflict and discussion status |
| `labels[]` | array | Label objects |
| `assignees[]` / `reviewers[]` | array | User objects |
| `author` / `closed_by` | object | User objects |
| `milestone` | object | Milestone object (with `id`, `iid`, `title`, `state`, `due_date`, etc.) |
| `references` | object | `{ short, relative, full }` — e.g. `!1`, `project!1`, `group/project!1` |
| `head_pipeline` | object | Pipeline on source branch HEAD (preferred over `pipeline`) |
| `diff_refs` | object | `{ base_sha, head_sha, start_sha }` — populated asynchronously |
| `task_completion_status` | object | `{ count, completed_count }` |
| `time_stats` | object | `{ time_estimate, total_time_spent, human_time_estimate, human_total_time_spent }` |
| `changes_count` | string | Number of changes. `"1000+"` when exceeding display limit. Populated asynchronously. |
| `subscribed` | boolean | Authenticated user subscription status |

### Get MR participants / reviewers

```
GET /projects/:id/merge_requests/:merge_request_iid/participants
GET /projects/:id/merge_requests/:merge_request_iid/reviewers
```

Reviewers response includes `state` field: `unreviewed` or `reviewed`.

### Get MR commits

```
GET /projects/:id/merge_requests/:merge_request_iid/commits
```

Returns array of commit objects (same structure as Commits API).

### MR dependencies (blocks / blockees)

```
GET    /projects/:id/merge_requests/:merge_request_iid/blocks      # List MRs blocking this MR
POST   /projects/:id/merge_requests/:merge_request_iid/blocks      # Create dependency
DELETE /projects/:id/merge_requests/:merge_request_iid/blocks/:block_id  # Remove dependency
GET    /projects/:id/merge_requests/:merge_request_iid/blockees    # List MRs blocked BY this MR
```

Create parameters:
- `blocking_merge_request_iid` (integer, for same-project) OR `blocking_merge_request_id` (integer, global ID)
- `blocking_project_id` (integer/string, for cross-project dependencies)

### MR diffs

```
GET /projects/:id/merge_requests/:merge_request_iid/diffs
```

Parameters: `page`, `per_page` (default 20), `unidiff` (boolean).
Subject to diff limits. Returns `collapsed` / `too_large` flags (GitLab 18.4+).

```
GET /projects/:id/merge_requests/:merge_request_iid/changes        # Deprecated (GitLab 15.7), use /diffs
```

Additional: `access_raw_diffs` (boolean, fetch via Gitaly bypassing DB limits).

### MR raw diffs

```
GET /projects/:id/merge_requests/:merge_request_iid/raw_diffs
```

Returns raw unified diff output. Subject to diff limits.

### MR diff versions

```
GET /projects/:id/merge_requests/:merge_request_iid/versions
GET /projects/:id/merge_requests/:merge_request_iid/versions/:version_id
```

Response: `{ id, head_commit_sha, base_commit_sha, start_commit_sha, created_at, merge_request_id, state, real_size, patch_id_sha, commits[], diffs[] }`.

SHA semantics:
- `base_commit_sha` — merge-base between source and target
- `head_commit_sha` — HEAD of source branch
- `start_commit_sha` — HEAD of target branch when diff was created

### MR pipelines

```
GET  /projects/:id/merge_requests/:merge_request_iid/pipelines
POST /projects/:id/merge_requests/:merge_request_iid/pipelines
```

POST creates a new pipeline (detached or merged results, depending on project settings). Configure `.gitlab-ci.yml` with `only: [merge_requests]`.

### Create a merge request

```
POST /projects/:id/merge_requests
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer/string | Yes | Project ID |
| `source_branch` | string | Yes | Source branch |
| `target_branch` | string | Yes | Target branch |
| `title` | string | Yes | MR title |
| `description` | string | No | MR description (max 1,048,576 chars) |
| `assignee_id` / `assignee_ids` | integer / int[] | No | Assignee(s). `0` or empty = unassign. |
| `reviewer_ids` | int[] | No | Reviewer IDs. `0` or empty = no reviewers. |
| `milestone_id` | integer | No | Milestone global ID (mutually exclusive with `milestone`) |
| `milestone` | string | No | Milestone title, exact match (mutually exclusive with `milestone_id`) |
| `labels` | string | No | Comma-separated. Creates new project labels if they don't exist. |
| `target_project_id` | integer | No | For cross-project MRs |
| `remove_source_branch` | boolean | No | Auto-delete source branch after merge |
| `squash` | boolean | No | Squash commits on merge (project settings may override) |
| `allow_collaboration` | boolean | No | Allow commits from target branch members |
| `merge_after` | string | No | Date after which merge is allowed (GitLab 17.8+) |

### Update a merge request

```
PUT /projects/:id/merge_requests/:merge_request_iid
```

Must include at least one non-required attribute. Additional parameters beyond create:

| Parameter | Type | Description |
|-----------|------|-------------|
| `add_labels` | string | Comma-separated labels to add |
| `remove_labels` | string | Comma-separated labels to remove |
| `state_event` | string | `close` or `reopen` |
| `discussion_locked` | boolean | Lock/unlock discussions |
| `target_branch` | string | Change target branch |

### Delete a merge request

```
DELETE /projects/:id/merge_requests/:merge_request_iid
```

Only administrators and project owners can delete MRs.

### Merge a merge request

```
PUT /projects/:id/merge_requests/:merge_request_iid/merge
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer/string | Yes | Project ID |
| `merge_request_iid` | integer | Yes | MR IID |
| `sha` | string | Conditional | HEAD SHA of source branch. Required if project/group setting is enabled (GitLab 19.2+). |
| `auto_merge` | boolean | No | Merge when checks pass (replaces deprecated `merge_when_pipeline_succeeds`) |
| `merge_commit_message` | string | No | Custom merge commit message |
| `should_remove_source_branch` | boolean | No | Delete source branch after merge |
| `squash` | boolean | No | Squash commits |
| `squash_commit_message` | string | No | Custom squash commit message |

Error responses:

| Status | Message | Reason |
|--------|---------|--------|
| `400` | `SHA must be provided when merging` | SHA required by project settings but not provided |
| `401` | `401 Unauthorized` | No permission to merge |
| `405` | `405 Method Not Allowed` | MR cannot merge |
| `409` | `SHA does not match HEAD of source branch` | Provided SHA doesn't match |
| `422` | `Branch cannot be merged` | Merge failed |

**Note (GitLab 19.1+):** On projects with merge trains enabled, `auto_merge` routes the MR to the merge train instead of merging directly.

### Merge to default merge ref path

```
GET /projects/:id/merge_requests/:merge_request_iid/merge_ref
```

Merges into `refs/merge-requests/:iid/merge` without changing the target branch. Returns `{ commit_id }`.

### Cancel auto-merge

```
POST /projects/:id/merge_requests/:merge_request_iid/cancel_merge_when_pipeline_succeeds
```

Returns `201` on success (or if already merged), `406` if MR is closed. On merge trains, also removes from train.

### Rebase a merge request

```
PUT /projects/:id/merge_requests/:merge_request_iid/rebase
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `skip_ci` | boolean | No | Skip creating CI pipeline |

Returns `202` on successful enqueue: `{ "rebase_in_progress": true }`.

Error responses: `403` (no push permission, branch doesn't exist, or protected from force push), `409` (failed to enqueue).

Poll the GET MR endpoint with `include_rebase_in_progress=true` to check status. On failure: `{ "rebase_in_progress": false, "merge_error": "Rebase failed. Please rebase locally" }`.

### MR issues

```
GET /projects/:id/merge_requests/:merge_request_iid/closes_issues      # Issues closed on merge
GET /projects/:id/merge_requests/:merge_request_iid/related_issues     # Issues referenced in MR
```

### MR subscription

```
POST /projects/:id/merge_requests/:merge_request_iid/subscribe        # Subscribe
POST /projects/:id/merge_requests/:merge_request_iid/unsubscribe      # Unsubscribe
```

Returns `304 Not Modified` if already subscribed/unsubscribed.

### MR to-do item

```
POST /projects/:id/merge_requests/:merge_request_iid/todo
```

Returns `304 Not Modified` if to-do already exists.

### MR time tracking

```
POST /projects/:id/merge_requests/:merge_request_iid/time_estimate      # Set estimate
POST /projects/:id/merge_requests/:merge_request_iid/reset_time_estimate # Reset to 0
POST /projects/:id/merge_requests/:merge_request_iid/add_spent_time     # Add spent time
POST /projects/:id/merge_requests/:merge_request_iid/reset_spent_time   # Reset spent to 0
GET  /projects/:id/merge_requests/:merge_request_iid/time_stats         # Get stats
```

Parameter: `duration` in human format (e.g. `3h30m`). `add_spent_time` also supports `summary`.

---

## Usage Patterns

### List commits with stats for a branch since a date

```bash
curl --header "PRIVATE-TOKEN: $TOKEN" \
  "https://gitlab.example.com/api/v4/projects/5/repository/commits?ref_name=main&since=2024-01-01T00:00:00Z&with_stats=true"
```

### Create a branch

```bash
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
  "https://gitlab.example.com/api/v4/projects/5/repository/branches?branch=feature-x&ref=main"
```

### Batch commit: create and update files

```bash
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
  --header "Content-Type: application/json" \
  --data '{
    "branch": "main",
    "commit_message": "Add and update files",
    "actions": [
      {"action": "create", "file_path": "src/new.go", "content": "package src\n"},
      {"action": "update", "file_path": "README.md", "content": "# Updated\n"}
    ]
  }' \
  "https://gitlab.example.com/api/v4/projects/5/repository/commits"
```

### Cherry-pick with dry run

```bash
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
  --form "branch=main" --form "dry_run=true" \
  "https://gitlab.example.com/api/v4/projects/5/repository/commits/abc123def/cherry_pick"
```

### Search open MRs by label and branch

```bash
curl --header "PRIVATE-TOKEN: $TOKEN" \
  "https://gitlab.example.com/api/v4/projects/5/merge_requests?state=opened&labels=bug&target_branch=main"
```

### Create an MR with assignees and reviewers

```bash
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
  --header "Content-Type: application/json" \
  --data '{
    "source_branch": "feature-x",
    "target_branch": "main",
    "title": "Implement feature X",
    "assignee_ids": [5, 8],
    "reviewer_ids": [12],
    "labels": "feature,backend",
    "remove_source_branch": true,
    "squash": true
  }' \
  "https://gitlab.example.com/api/v4/projects/5/merge_requests"
```

### Merge with SHA verification

```bash
curl --request PUT --header "PRIVATE-TOKEN: $TOKEN" \
  "https://gitlab.example.com/api/v4/projects/5/merge_requests/42/merge?sha=abc123def456&should_remove_source_branch=true"
```

### Auto-merge (merge when pipeline succeeds)

```bash
curl --request PUT --header "PRIVATE-TOKEN: $TOKEN" \
  --header "Content-Type: application/json" \
  --data '{"auto_merge": true}' \
  "https://gitlab.example.com/api/v4/projects/5/merge_requests/42/merge"
```

### Depend MR on another MR

```bash
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
  "https://gitlab.example.com/api/v4/projects/5/merge_requests/42/blocks?blocking_merge_request_iid=41"
```

### Cross-project MR dependency

```bash
curl --request POST --header "PRIVATE-TOKEN: $TOKEN" \
  "https://gitlab.example.com/api/v4/projects/5/merge_requests/42/blocks?blocking_merge_request_iid=10&blocking_project_id=8"
```

### Delete all merged branches

```bash
curl --request DELETE --header "PRIVATE-TOKEN: $TOKEN" \
  "https://gitlab.example.com/api/v4/projects/5/repository/merged_branches"
```

---

## Pagination

Standard GitLab pagination via `page` (default 1) and `per_page` (default 20, max 100).

**Important:** Commits API does NOT return `x-total` and `x-total-pages` headers. For other endpoints, these headers are available.

## Best Practices

1. **Use `detailed_merge_status` over `merge_status`** — `merge_status` is deprecated since GitLab 15.6 and doesn't cover all states.
2. **Use `draft` (boolean) over `wip` (string)** — `wip` deprecated in GitLab 19.0.
3. **Use `merge_user` over `merged_by`** — `merged_by` deprecated in GitLab 14.7.
4. **Use `auto_merge` over `merge_when_pipeline_succeeds`** — deprecated in GitLab 17.11.
5. **Use MR `iid` for project-scoped operations** — Most project endpoints expect `merge_request_iid`, not global `id`.
6. **Use `head_pipeline` over `pipeline`** — More complete information about the pipeline on the source branch HEAD.
7. **Provide `sha` when merging** — Ensures the merge target matches what was reviewed. Some instances require it (GitLab 19.2+).
8. **Check `rebase_in_progress` after enqueuing a rebase** — Poll the GET MR endpoint with `include_rebase_in_progress=true`.
9. **Use `not` parameter for exclusion queries** — e.g., `?not[labels]=wontfix` to exclude labeled MRs.
10. **For batch commits, watch request size** — >20 MB is rate-limited (3 req/30 sec); >300 MB is rejected.
11. **Handle async population of `diff_refs` and `changes_count`** — These fields are empty immediately after MR creation.
12. **Use `search` and `in` parameters efficiently** — Combine `search=term&in=title` for targeted queries.

## Common Pitfalls

- **`merge_status` returns `can_be_merged` even when MR is already merged** — The field is cached. Use `detailed_merge_status` which returns `not_open` for merged MRs.
- **`merge_status` / `has_conflicts` may be stale** — Use `with_merge_status_recheck=true` to trigger async recalculation.
- **Deleting a branch doesn't completely erase data** — Some information persists for history and recovery.
- **MR `id` vs `iid` confusion** — Project endpoints use `iid`, cross-project references need global `id`.
- **`x-total` header missing from Commits API** — Cannot rely on total count for commit listings.
- **`diff_refs` and `changes_count` empty for new MRs** — Populate asynchronously after creation.
- **Force-pushing branch via commits API** — When `force=true` without `start_branch`/`start_sha`, it defaults to the branch itself, having no effect.
- **Creating an MR doesn't auto-populate pipelines** — Use `POST /merge_requests/:iid/pipelines` if needed.

## Deprecated Attributes

| Deprecated | Replacement | Since |
|------------|------------|-------|
| `merge_status` | `detailed_merge_status` | 15.6 |
| `merged_by` | `merge_user` | 14.7 |
| `merge_when_pipeline_succeeds` | `auto_merge` | 17.11 |
| `wip` | `draft` | 19.0 |
| `work_in_progress` | `draft` | 19.0 |
| `approvals_before_merge` | Merge request approvals API | 16.0 |
| `allow_maintainer_to_push` | `allow_collaboration` | — |
| `reference` | `references` | 12.7 |
| `assignee` (singular) | `assignees` (array) | — |
| `pipeline` | `head_pipeline` | — |
| `/changes` endpoint | `/diffs` endpoint | 15.7 |
