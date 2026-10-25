# Workspace authorization

Authorization is evaluated in the service layer from the caller's active workspace membership. Suspended and removed users have no access. The first release has no project-level role overrides; every project inherits its workspace policy.

Workspace-visible projects are available according to the matrix below. Private projects and their tasks are restricted to owners and admins. Archived workspaces and projects remain readable when explicitly requested, but reject new mutations. Removing a member preserves existing reporter and assignee IDs for historical attribution while immediately revoking future access.

| Action | Owner | Admin | Member | Viewer |
|---|:---:|:---:|:---:|:---:|
| Read workspace and projects | Yes | Yes | Yes | Yes |
| Update workspace | Yes | Yes | No | No |
| Archive workspace | Yes | No | No | No |
| Transfer ownership | Yes | No | No | No |
| Create, update, archive projects | Yes | Yes | No | No |
| Read members | Yes | Yes | Yes | Yes |
| Invite, change, suspend, remove members | Yes | Yes* | No | No |
| Read audit log | Yes | Yes | No | No |
| Read tasks | Yes | Yes | Yes | Yes |
| Create and mutate tasks | Yes | Yes | Yes | No |

\* Admins cannot alter the owner, assign the owner role, or transfer ownership.
