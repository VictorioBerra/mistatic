You can use sqlite3 to see pb_data/data.db if you wanna see real state while working/debugging issues.

The backend is pocketbase, there are a lot of things you may be tempted to hand-roll tat are uilt in.

Som examples, PB does:

- Event Hooks
- Routing
- Database functions that can run all kinds of SQL statements, including raw queries.
- Cron jobs (scheduling tasks at specific intervals)
- File storage (handling file uploads and storage)
- User authentication and management (handling user login, registration, and sessions)
- Records have their own idiomatic way of being accessed and manipulated within PocketBase. https://pocketbase.io/docs/go-records/ load this.
- Migrations
- Sending emails

One of the most important things to keep in mind is API rules. Be very careful with these, they get converted to SQL in opinionated and unexpected ways. They trigger joins in ways that may not be obvious, which can impact performance and correctness.

## PocketBase API rule gotchas (sourced from reading PB v0.39.9 source, not the docs)

**`@collection.X` is a cross-join, not an EXISTS.** `@collection.foo.field = value` adds `LEFT JOIN foo ON 1=1` (no ON clause) to the main query, then filters via WHERE. Every row in the primary table is initially multiplied by every row in `foo`. PocketBase automatically adds `DISTINCT` to collapse duplicates, but the intermediate row explosion is real — keep referenced collections small.

**Multiple conditions on the same `@collection.X` ARE correlated.** `@collection.permissions.name ?= 'x' && @collection.permissions.roles.users.id ?= @request.auth.id` shares one JOIN alias for `permissions` — both conditions apply to the same permission row. You do NOT get two independent EXISTS checks.

**Multi-hop relation traversal works but chains `json_each`.** `@collection.permissions.roles.users.id` expands as: `json_each(permissions.roles)` → `roles` → `json_each(roles.users)`. Each multi-value hop adds two more JOINs to the main query. Semantically correct; potentially slow for large datasets.

**`?=` (any-match) skips the MultiMatchSubquery path entirely.** The "match-all vs match-any" distinction matters: plain `=` on a multi-value relation requires ALL items to satisfy the condition; `?=` requires AT LEAST ONE. Always use `?=` when traversing multi-value relations unless you truly mean match-all.

**The `listRule` of a referenced `@collection.X` is also enforced.** If `foo` has a non-empty `listRule`, PocketBase evaluates it as an additional `AND` on the main query. This is a security feature but can cause silent filtering that looks like missing data.

**Two different `@collection` references are always independent.** `@collection.A.x = 1 && @collection.B.y = 2` — A and B are separate cross-joins with no correlation between them. Only conditions on the *same* `@collection.collectionName` share a join.

**Permission rule pattern used in this codebase:**
```
@request.auth.id != '' && @collection.permissions.name ?= 'perm:name' && @collection.permissions.roles.users.id ?= @request.auth.id
```
This correctly gates access to users who hold a specific permission via the roles → permissions → users chain. Verified against PB source.

Use Golang for all PB hooks. Use JS for migrations.

In /frontend, always use pb.send() to call our backend, never window fetch.

## Svelte 5 / Frontend rules

**Never nest `<button>` inside `<button>`.** The browser repairs this silently, breaking Svelte's DOM assumptions and causing a hard build error. Always check that interactive elements are siblings, not ancestors.

**Always associate `<label>` with its control** using matching `for`/`id` attributes. A label wrapping an input in a sibling `<div>` is not implicitly associated — the `for`/`id` pair is required.

**`$state(prop)` and `const x = data.x` only capture the initial value.** In Svelte 5 runes mode this triggers a `state_referenced_locally` warning. Use `untrack(() => value)` (imported from `svelte`) when you intentionally want the initial value only (e.g. initializing editable form state from a prop). Use `$derived(value)` when you want the value to stay in sync.

**`autofocus` triggers an a11y warning.** Suppress intentional uses with `<!-- svelte-ignore a11y_autofocus -->` on the line before the element.

**Drag-and-drop divs need an ARIA role.** A `<div>` with `ondragover`/`ondrop` must have `role="group"` (or similar) to satisfy `a11y_no_static_element_interactions`.
