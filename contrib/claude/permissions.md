# Permission-gated UI

Authorization is enforced on the **server** (see [`authorization.md`](authorization.md)). The frontend never decides what a user is *allowed* to do — it asks the API and renders accordingly. The API exposes per-record permissions through the `permission(action:)` field, which a component selects as a boolean alias (`canUpdate`, `canDelete`, …) on the node it renders.

This guide covers how to **consume** those booleans. It does not grant access; hiding a button is a UX nicety, not a security control — the mutation is still authorized server-side.

## Related guides

| Topic | Guide |
|-------|--------|
| Server-side IAM policies and actions | [`contrib/claude/authorization.md`](authorization.md) |
| Fragments, colocated data | [`contrib/claude/relay.md`](relay.md) |
| Component shape and props | [`contrib/claude/react-components.md`](react-components.md) |

## Select permissions in the fragment that needs them

A component that renders an action selects the matching permission **in its own fragment**, aliased to a `can…` boolean. Keep the action string identical to the IAM action it guards.

```tsx
const documentListItemFragment = graphql`
  fragment DocumentListItem_document on Document {
    id
    title
    canUpdate: permission(action: "core:document:update")
    canDelete: permission(action: "core:document:delete")
  }
`;
```

Colocate the permission with the action it gates — never drill a `canDelete` boolean down as a prop from a parent (the same data-as-props rule as everywhere else; see [`react-components.md`](react-components.md#props-are-for-configuration-and-composition-not-data)).

## Gate the action on the boolean

Read the boolean via `useFragment` and gate the control. Default to **hiding** an action the user cannot perform; **disable** (with an explanatory tooltip) only when the action's *absence* would be confusing.

```tsx
export function DocumentListItem({ documentKey }: DocumentListItemProps) {
  const document = useFragment(documentListItemFragment, documentKey);
  return (
    <Tr>
      <Td>{document.title}</Td>
      <Td>
        {document.canUpdate && <EditDocumentDialog documentKey={document} />}
        {document.canDelete && <DeleteDocumentButton documentId={document.id} />}
      </Td>
    </Tr>
  );
}
```

### Hide vs. disable

```text
// Hide — the user has no business with this action (most cases)
{canDelete && <DeleteButton … />}

// Disable — the action is expected to be there, but is currently unavailable;
// pair with a tooltip explaining why
<Button disabled={!canPublish} title={!canPublish ? t("noPublishPermission") : undefined}>
  {t("publish")}
</Button>
```

## Bulk / toolbar actions

For list toolbars, derive the aggregate from the items and hide the bulk control when no row qualifies.

```tsx
const canDeleteAny = documents.some(({ canDelete }) => canDelete);

{canDeleteAny && <BulkDeleteButton ids={selection} />}
```

## Don't

```text
// Bad — client-side role check standing in for a server permission
if (currentUser.role === "ADMIN") { showDelete(); }

// Bad — drilling a permission boolean as a prop instead of selecting it where used
<DocumentListItem canDelete={doc.canDelete} />

// Bad — gating on a hand-rolled action string that drifts from the IAM action
permission(action: "document_delete")   // must match "core:document:delete"

// Bad — treating a hidden button as the security boundary
// (the mutation must still be authorized server-side; UI gating is UX only)
```

## Why server-derived, not role-based

Roles are coarse and change; resource-level permissions answer the exact question the UI asks ("can *this* user act on *this* record?"). Selecting `permission(action:)` keeps the frontend in lockstep with the IAM policies in [`authorization.md`](authorization.md) without re-encoding any authorization logic in the client.
