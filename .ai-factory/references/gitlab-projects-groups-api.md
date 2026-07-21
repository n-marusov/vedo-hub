# GitLab Projects & Groups API Reference

> Source: https://docs.gitlab.com/api/groups.html, https://docs.gitlab.com/api/projects.html
> Created: 2026-07-21
> Updated: 2026-07-21

## Overview

GitLab REST API v4 for managing Projects and Groups — the two core organizational units. Projects contain repositories, issues, merge requests, and CI/CD pipelines. Groups organize projects and users with hierarchical nesting (subgroups). Both APIs use JSON over HTTP with `PRIVATE-TOKEN` or `Authorization: Bearer <token>` authentication.

**Base URL:** `https://gitlab.example.com/api/v4`

**Tier:** Free, Premium, Ultimate (some endpoints Premium/Ultimate only)

**Project ID** can be numeric `id` or URL-encoded path (`group%2Fproject`).

**Group ID** can be numeric `id` or URL-encoded path (`group%2Fsubgroup`).

## Core Concepts

- **Project**: Central hub for code, issues, MRs, CI/CD. Owned by a user or group namespace.
- **Group**: Organizational container for projects and users. Supports nesting via subgroups.
- **Namespace**: Either a user personal namespace or a group. Projects belong to a namespace.
- **Visibility**: `private` (members only), `internal` (all authenticated), `public` (everyone).
- **Access Levels**: `5` Minimal, `10` Guest, `15` Planner, `20` Reporter, `25` Security Manager, `30` Developer, `40` Maintainer, `50` Owner.
- **Feature Visibility**: Each project feature (issues, MRs, wiki, builds, etc.) has independent access level: `disabled`, `private`, `enabled`, or `public` (pages only).

## Groups API

### Retrieve a group

```
GET /groups/:id
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer or string | Yes | Group ID or URL-encoded path |
| `with_custom_attributes` | boolean | No | Include custom attributes (admin only) |
| `with_projects` | boolean | No | Include project details (default `true`, deprecated in API v5) |

Returns group object with `id`, `name`, `path`, `full_path`, `parent_id`, `visibility`, `avatar_url`, `web_url`, `created_at`, `request_access_enabled`, `default_branch_protection_defaults`, `shared_with_groups`, etc.

Premium/Ultimate additions: `shared_runners_minutes_limit`, `extra_shared_runners_minutes_limit`, `marked_for_deletion_on`, `membership_lock`, `wiki_access_level`, `duo_features_enabled`, `experiment_features_enabled`.

AI settings (GitLab 19.2+): `ai_settings` object with `duo_agent_platform_enabled`, `ai_catalog_restricted_to_group_hierarchy`, `prompt_injection_protection_level`, etc.

### List all groups

```
GET /groups
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `all_available` | boolean | No | Return all accessible groups (admin default `true`) |
| `owned` | boolean | No | Only groups owned by current user |
| `min_access_level` | integer | No | Minimum access level filter |
| `search` | string | No | Search by name or path |
| `order_by` | string | No | `name` (default), `path`, `id`, `similarity` |
| `sort` | string | No | `asc` (default) or `desc` |
| `top_level_only` | boolean | No | Exclude subgroups |
| `visibility` | string | No | `public`, `internal`, `private` |
| `statistics` | boolean | No | Include storage stats (admin only) |
| `active` | boolean | No | Exclude archived and marked-for-deletion groups |
| `archived` | boolean | No | Only archived groups (GitLab 18.2+) |
| `marked_for_deletion_on` | date | No | Filter by deletion date (Premium+) |
| `repository_storage` | string | No | Filter by storage (admin, Premium+) |

Default pagination: 20 results. Max per page: 100 (`?per_page=100`).

### List group projects

```
GET /groups/:id/projects
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `active` | boolean | No | Only active projects (GitLab 18.8+) |
| `archived` | boolean | No | Filter by archived status |
| `include_subgroups` | boolean | No | Include subgroup projects |
| `with_shared` | boolean | No | Include shared projects (default `true`) |
| `min_access_level` | integer | No | Minimum access level |
| `order_by` | string | No | `id`, `name`, `path`, `created_at`, `updated_at`, `similarity`, `star_count`, `last_activity_at` |
| `search` | string | No | Search projects |

To distinguish group-owned vs shared projects, check `namespace` field — shared projects have a different namespace.

### List subgroups

```
GET /groups/:id/subgroups
```

Same filtering as List groups. Returns direct child subgroups only.

### List descendant groups

```
GET /groups/:id/descendant_groups
```

Returns all nested subgroups recursively.

### List shared groups

```
GET /groups/:id/groups/shared
```

Groups where this group has been invited.

### List invited groups

```
GET /groups/:id/invited_groups
```

Groups invited to this group (rate-limited: 60 req/min).

### Create a group

```
POST /groups
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Group name |
| `path` | string | Yes | Group URL path |
| `description` | string | No | Group description |
| `visibility` | string | No | `private`, `internal`, `public` |
| `parent_id` | integer | No | Parent group ID for nesting |
| `default_branch` | string | No | Default branch for group projects (GitLab 16.11+) |
| `default_branch_protection` | integer | No | Deprecated. Use `default_branch_protection_defaults` |
| `default_branch_protection_defaults` | hash | No | See branch protection options below |
| `project_creation_level` | string | No | `noone`, `maintainer`, `developer`, `administrator` |
| `subgroup_creation_level` | string | No | `owner`, `maintainer` |
| `require_two_factor_authentication` | boolean | No | Enforce 2FA |
| `two_factor_grace_period` | integer | No | Hours before 2FA enforced |
| `share_with_group_lock` | boolean | No | Prevent sharing projects outside group |
| `request_access_enabled` | boolean | No | Allow access requests |
| `lfs_enabled` | boolean | No | Enable LFS |
| `emails_enabled` | boolean | No | Enable email notifications |
| `mentions_disabled` | boolean | No | Disable group mentions |
| `auto_devops_enabled` | boolean | No | Default Auto DevOps |
| `enabled_git_access_protocol` | string | No | `ssh`, `http`, `all` (GitLab 16.9+) |
| `organization_id` | integer | No | Organization ID |
| `wiki_access_level` | string | No | `disabled`, `private`, `enabled` (Premium+) |
| `membership_lock` | boolean | No | Prevent adding users to projects (Premium+) |
| `duo_availability` | string | No | `default_on`, `default_off`, `never_on` |
| `experiment_features_enabled` | boolean | No | Enable experiment features |
| `ai_settings_attributes` | hash | No | AI/Duo Agent Platform settings |

**Note:** On GitLab.com, top-level groups cannot be created via API — use GitLab UI.

### Create a subgroup

Same as Create a group, but with `parent_id` set to the parent group.

### Update a group

```
PUT /groups/:id
```

Same parameters as Create, plus:
- `prevent_sharing_groups_outside_hierarchy` (top-level groups only)
- `shared_runners_setting`: `enabled`, `disabled_and_overridable`, `disabled_and_unoverridable`
- `prevent_forking_outside_group` (Premium+)
- `unique_project_download_limit` and related params (Ultimate+)
- `ip_restriction_ranges` (Premium+)
- `allowed_email_domains_list` (Premium+ GitLab 17.4+)
- `allow_personal_snippets` (GitLab 18.9+)
- `max_artifacts_size` (GitLab 19.1+)
- `ai_audit_events_storage_enabled` (GitLab 19.2+)
- `web_based_commit_signing_enabled` (GitLab.com only)

### Delete / Schedule deletion / Restore a group

```
DELETE /groups/:id                              # Schedule for deletion (30-day retention on GitLab.com)
DELETE /groups/:id  (permanently_remove=true)    # Immediately delete a sub-group scheduled for deletion
POST /groups/:id/restore                        # Restore a group marked for deletion
```

### Archive / Unarchive a group (GitLab 18.9+)

```
POST /groups/:id/archive
POST /groups/:id/unarchive
```

Prerequisites: Administrator or Owner role. Returns `422` if already in that state.

### Transfer a group

```
POST /groups/:id/transfer?group_id=<new_parent_id>
```

Without `group_id`, transforms a subgroup into a top-level group.

### Invite groups

```
POST /groups/:id/share
```

| Parameter | Type | Required |
|-----------|------|----------|
| `group_id` | integer | Yes |
| `group_access` | integer | Yes (access level) |
| `expires_at` | date | No |

```
DELETE /groups/:id/share/:group_id   # Remove invitation
```

### Group avatar

```
GET  /groups/:id/avatar              # Download avatar
PUT  /groups/:id  (avatar=@file.png) # Upload avatar (multipart/form-data)
PUT  /groups/:id  (avatar=)          # Remove avatar
```

### LDAP sync (Premium+, Self-Managed)

```
POST /groups/:id/ldap_sync
```

### Group transfer locations

```
GET /groups/:id/transfer_locations?search=<name>
```

## Projects API

### Retrieve a project

```
GET /projects/:id
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer or string | Yes | Project ID or URL-encoded path |
| `license` | boolean | No | Include license data |
| `statistics` | boolean | No | Include storage stats (Reporter+) |
| `with_custom_attributes` | boolean | No | Custom attributes (admin only) |

Returns full project object with ~100+ fields including: `id`, `name`, `path_with_namespace`, `default_branch`, `visibility`, `namespace` (object with `id`, `name`, `path`, `kind`, `full_path`, `parent_id`), all `*_access_level` feature flags, CI/CD settings, merge settings, statistics, permissions, license, `_links`, etc.

### List all projects

```
GET /projects
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `archived` | boolean | No | Filter by archived status |
| `membership` | boolean | No | Only projects user is member of |
| `owned` | boolean | No | Only projects owned by user |
| `starred` | boolean | No | Only starred projects |
| `search` | string | No | Search by name/path/description (case-insensitive, AND terms with `+`) |
| `simple` | boolean | No | Return limited fields |
| `visibility` | string | No | `public`, `internal`, `private` |
| `order_by` | string | No | `id`, `name`, `path`, `created_at`, `updated_at`, `star_count`, `last_activity_at`, `similarity` |
| `sort` | string | No | `asc` or `desc` (default `desc`) |
| `min_access_level` | integer | No | Minimum access level |
| `topic` | string | No | Comma-separated topics |
| `with_issues_enabled` | boolean | No | Filter by issues enabled |
| `with_merge_requests_enabled` | boolean | No | Filter by MRs enabled |
| `with_programming_language` | string | No | Filter by language |
| `updated_after` | datetime | No | ISO 8601. Requires `order_by=updated_at` |
| `updated_before` | datetime | No | ISO 8601. Requires `order_by=updated_at` |
| `last_activity_after` | datetime | No | ISO 8601 |
| `last_activity_before` | datetime | No | ISO 8601 |
| `active` | boolean | No | Exclude archived and marked-for-deletion |
| `imported` | boolean | No | Only imported projects |
| `id_after` / `id_before` | integer | No | ID range filtering |
| `repository_checksum_failed` | boolean | No | Premium+ admin only |
| `wiki_checksum_failed` | boolean | No | Premium+ admin only |

Supports offset-based (up to 50k) and keyset-based pagination (50k+).

### List user personal projects

```
GET /users/:user_id/projects
```

Returns only projects in the user's personal namespace.

### List contributed projects

```
GET /users/:user_id/contributed_projects
```

Projects the user contributed to in the past year.

### Create a project

```
POST /projects
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes* | Project name (*if `path` not provided) |
| `path` | string | Yes* | Repository path (*if `name` not provided) |
| `namespace_id` | integer | No | Group/subgroup ID (default: personal namespace) |
| `description` | string | No | Project description |
| `visibility` | string | No | `private`, `internal`, `public` |
| `default_branch` | string | No | Default branch name (requires `initialize_with_readme=true`) |
| `initialize_with_readme` | boolean | No | Create with empty README (default `false`) |
| `import_url` | string | No | Import from URL (mutually exclusive with `initialize_with_readme`) |
| `template_name` | string | No | Built-in or custom template name |
| `template_project_id` | integer | No | Custom template project ID (Premium+) |
| `use_custom_template` | boolean | No | Use custom template (Premium+) |
| `topics` | array | No | Project topics |
| `merge_method` | string | No | `merge`, `rebase_merge`, `ff` |
| `squash_option` | string | No | `never`, `always`, `default_on`, `default_off` |
| `shared_runners_enabled` | boolean | No | Enable instance runners |
| `container_registry_access_level` | string | No | `disabled`, `private`, `enabled` |
| `package_registry_access_level` | string | No | `disabled`, `private`, `enabled` |
| `lfs_enabled` | boolean | No | Enable LFS |
| `request_access_enabled` | boolean | No | Allow access requests |
| `autoclose_referenced_issues` | boolean | No | Auto-close referenced issues |
| `remove_source_branch_after_merge` | boolean | No | Delete source branch after merge |
| `only_allow_merge_if_pipeline_succeeds` | boolean | No | Pipelines must succeed |
| `only_allow_merge_if_all_discussions_are_resolved` | boolean | No | All discussions resolved |
| `mirror` | boolean | No | Enable pull mirroring (Premium+) |

### Update a project

```
PUT /projects/:id
```

Same parameters as Create, plus:
- `ci_default_git_depth` — shallow clone depth
- `ci_forward_deployment_enabled` — prevent outdated deployment jobs
- `ci_separated_caches` — separate caches by branch protection
- `ci_restrict_pipeline_cancellation_role` — `developer`, `maintainer`, `no_one`
- `ci_pipeline_variables_minimum_override_role` — `owner`, `maintainer`, `developer`, `no_one_allowed`
- `ci_config_path` — CI/CD config file path
- `build_timeout` — job timeout in seconds
- `auto_devops_enabled` / `auto_devops_deploy_strategy`
- `max_artifacts_size` — max artifact size in MB
- `merge_pipelines_enabled` / `merge_trains_enabled`
- `issues_template` / `merge_requests_template` (Premium+)
- `mr_default_title_template` (GitLab 19.0+)
- `web_based_commit_signing_enabled` (GitLab.com only)
- `secret_push_protection_enabled` via Security Settings API

### Archive / Unarchive a project

```
POST /projects/:id/archive
POST /projects/:id/unarchive
```

Prerequisites: Administrator or Owner role.

### Delete a project

```
DELETE /projects/:id
DELETE /projects/:id?permanently_remove=true&full_path=<path>  # Immediate delete (Self-Managed only)
```

On GitLab.com: 30-day retention after marking for deletion.

### Restore a project

```
POST /projects/:id/restore
```

### Transfer a project

```
PUT /projects/:id/transfer?namespace=<group_id_or_path>
```

List transfer locations: `GET /projects/:id/transfer_locations`

### Project members

```
GET /projects/:id/users              # List all members
POST /projects/:id/import_project_members/:project_id  # Import members from another project
```

### Project sharing

```
POST /projects/:id/share             # Share with a group
DELETE /projects/:id/share/:group_id # Unshare from a group
```

| Parameter | Type | Required |
|-----------|------|----------|
| `group_id` | integer | Yes |
| `group_access` | integer | Yes (access level) |
| `expires_at` | date | No |

### Project languages

```
GET /projects/:id/languages
```

Returns percentage breakdown by language.

### Project housekeeping

```
POST /projects/:id/housekeeping
```

Optional `task` parameter: `prune` or `eager`.

### Project avatar

```
PUT  /projects/:id  (avatar=@file.png)  # Upload (max 200KB, 192x192 ideal)
GET  /projects/:id/avatar               # Download
PUT  /projects/:id  (avatar=)           # Remove
```

Accepted formats: `.bmp`, `.gif`, `.ico`, `.jpeg`, `.png`, `.tiff`.

## Project Feature Visibility Levels

Each feature has independent access control. Values: `disabled`, `private` (members only), `enabled` (everyone with access), `public` (everyone, pages only).

| Feature Attribute | Controls |
|-------------------|----------|
| `analytics_access_level` | Analytics |
| `builds_access_level` | Pipelines/CI |
| `container_registry_access_level` | Container registry |
| `environments_access_level` | Environments |
| `feature_flags_access_level` | Feature flags |
| `forking_access_level` | Forking |
| `infrastructure_access_level` | Infrastructure |
| `issues_access_level` | Issues |
| `merge_requests_access_level` | Merge requests |
| `model_experiments_access_level` | ML experiments |
| `model_registry_access_level` | ML registry |
| `monitor_access_level` | APM |
| `pages_access_level` | GitLab Pages |
| `releases_access_level` | Releases |
| `repository_access_level` | Repository |
| `requirements_access_level` | Requirements |
| `security_and_compliance_access_level` | Security & compliance |
| `snippets_access_level` | Snippets |
| `wiki_access_level` | Wiki |
| `package_registry_access_level` | Package registry |

## Access Levels Reference

| Value | Role |
|-------|------|
| `5` | Minimal access |
| `10` | Guest |
| `15` | Planner |
| `20` | Reporter |
| `25` | Security Manager |
| `30` | Developer |
| `40` | Maintainer |
| `50` | Owner |

## Default Branch Protection (Groups)

| Value | Description |
|-------|-------------|
| `0` | No protection — devs can push, force push, delete |
| `1` | Partial — devs can push new commits |
| `2` | Full — only maintainers can push |
| `3` | Protected against pushes — maintainers push/force push/accept MR; devs accept MR |
| `4` | Full after initial push — devs can push to empty repo; maintainers push and accept MR |

`default_branch_protection_defaults` hash options: `allowed_to_push` (array of access levels), `allow_force_push` (boolean), `allowed_to_merge` (array), `developer_can_initial_push` (boolean), `code_owner_approval_required` (boolean).

## Shared Runners Setting (Groups)

| Value | Description |
|-------|-------------|
| `enabled` | Instance runners on for all subgroups/projects |
| `disabled_and_overridable` | Off, but subgroups can override |
| `disabled_and_unoverridable` | Off, subgroups cannot override |

## Usage Patterns

### Create project in a group

```bash
curl --request POST \
  --header "PRIVATE-TOKEN: <token>" \
  --header "Content-Type: application/json" \
  --data '{
    "name": "my-project",
    "path": "my-project",
    "namespace_id": 42,
    "initialize_with_readme": true,
    "visibility": "internal"
  }' \
  --url "https://gitlab.example.com/api/v4/projects"
```

### List all projects in a group (with subgroups)

```bash
curl --header "PRIVATE-TOKEN: <token>" \
  "https://gitlab.example.com/api/v4/groups/my-group/projects?include_subgroups=true&per_page=100"
```

### Search projects

```bash
curl --header "PRIVATE-TOKEN: <token>" \
  "https://gitlab.example.com/api/v4/projects?search=my+project&membership=true"
```

### Update project feature visibility

```bash
curl --request PUT \
  --header "PRIVATE-TOKEN: <token>" \
  --header "Content-Type: application/json" \
  --data '{"issues_access_level": "disabled", "wiki_access_level": "disabled"}' \
  --url "https://gitlab.example.com/api/v4/projects/123"
```

### Share project with group

```bash
curl --request POST \
  --header "PRIVATE-TOKEN: <token>" \
  --header "Content-Type: application/json" \
  --data '{"group_id": 45, "group_access": 30}' \
  --url "https://gitlab.example.com/api/v4/projects/123/share"
```

### Create subgroup

```bash
curl --request POST \
  --header "PRIVATE-TOKEN: <token>" \
  --header "Content-Type: application/json" \
  --data '{"name": "Backend", "path": "backend", "parent_id": 42}' \
  --url "https://gitlab.example.com/api/v4/groups"
```

### Transfer project between groups

```bash
curl --request PUT \
  --header "PRIVATE-TOKEN: <token>" \
  "https://gitlab.example.com/api/v4/projects/123/transfer?namespace=45"
```

## Pagination

- Default: 20 results per page
- Max per page: 100 (`?per_page=100`)
- Offset-based: up to 50,000 projects
- Keyset-based: beyond 50,000 (use `id_after`/`id_before`)
- Response headers: `X-Page`, `X-Per-Page`, `X-Total`, `X-Total-Pages`, `Link`

## Best Practices

1. **Use URL-encoded paths** for project/group identifiers — more readable and stable than numeric IDs.
2. **Use `simple=true`** on list endpoints when you only need basic fields — significantly reduces response size.
3. **Use `include_subgroups=true`** when querying group projects to get a complete view.
4. **Prefer `*_access_level` over boolean fields** — old boolean fields (`issues_enabled`, `wiki_enabled`) are deprecated.
5. **Use `topics` instead of `tag_list`** — `tag_list` is deprecated since GitLab 14.0.
6. **Use `marked_for_deletion_on` instead of `marked_for_deletion_at`** — the latter is deprecated.
7. **Use `package_registry_access_level`** instead of deprecated `packages_enabled`.
8. **Handle pagination** — always check `Link` header or `X-Total-Pages` for complete results.
9. **Respect rate limits** — `invited_groups` endpoint is limited to 60 req/min per user/IP.
10. **Check response codes** — `202 Accepted` for scheduled deletions, `204 No Content` for successful deletes.

## Common Pitfalls

- **`with_projects=true`** on group retrieve is deprecated and capped at 100 projects. Use `GET /groups/:id/projects` instead.
- **`default_branch` on project creation** requires `initialize_with_readme=true`.
- **`import_url` and `initialize_with_readme=true`** are mutually exclusive.
- **`package_registry_access_level: "disabled"`** may not fully disable package registry on its own — also set `packages_enabled: false` in the same request (GitLab 17.10+ workaround).
- **`last_activity_at`** updates at most once per hour for performance reasons.
- **`approvals_before_merge`** is deprecated since GitLab 16.0 — use Merge Request Approvals API.
- **Top-level group creation** is not possible via API on GitLab.com — use the UI.
- **`service_desk_enabled`** is a computed field — it reads `false` when incoming email is not configured, even if the project setting is on.
- **`restrict_user_defined_variables`** is overridden by `ci_pipeline_variables_minimum_override_role` when both are set.

## Deprecated Attributes

| Deprecated | Replacement | Since |
|------------|-------------|-------|
| `tag_list` | `topics` | GitLab 14.0 |
| `marked_for_deletion_at` | `marked_for_deletion_on` | — |
| `approvals_before_merge` | Merge Request Approvals API | GitLab 16.0 |
| `packages_enabled` | `package_registry_access_level` | GitLab 17.10 |
| `container_registry_enabled` | `container_registry_access_level` | — |
| `issues_enabled` | `issues_access_level` | — |
| `jobs_enabled` | `builds_access_level` | — |
| `merge_requests_enabled` | `merge_requests_access_level` | — |
| `snippets_enabled` | `snippets_access_level` | — |
| `wiki_enabled` | `wiki_access_level` | — |
| `emails_disabled` | `emails_enabled` | — |
| `public_builds` | `public_jobs` | — |
| `restrict_user_defined_variables` | `ci_pipeline_variables_minimum_override_role` | GitLab 17.7 |
| `default_branch_protection` (group) | `default_branch_protection_defaults` | GitLab 17.0 |

## Version Notes

- **GitLab 17.0**: `default_branch_protection_defaults` replaces `default_branch_protection` on groups.
- **GitLab 17.4**: `allowed_email_domains_list` for groups (Premium+).
- **GitLab 17.10**: `package_registry_access_level` replaces `packages_enabled`.
- **GitLab 18.0**: Group archive/unarchive GA. Group deletion moved to Free tier.
- **GitLab 18.2**: `archived` filter on `GET /groups`.
- **GitLab 18.5**: `allow_personal_snippets` for groups.
- **GitLab 18.7**: AI settings `minimum_access_level_*` params.
- **GitLab 18.8**: `active` filter on group projects.
- **GitLab 18.9**: Group archive/unarchive feature flag removed.
- **GitLab 19.0**: `mr_default_title_template` GA. `built_in_project_templates_enabled` for groups.
- **GitLab 19.1**: `max_artifacts_size` on groups. `web_based_commit_signing_enabled` on projects (GitLab.com).
- **GitLab 19.2**: `ai_settings` object on group response. `ai_audit_events_storage_enabled`.
