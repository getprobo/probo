# @probo/n8n-nodes-probo

n8n community node package for the [Probo](https://www.probo.com) compliance platform. Automate compliance workflows — manage controls, documents, risks, vendors, cookie banners, and more — over the Probo GraphQL API.

This package provides two nodes:

| Node | Type | Description |
|------|------|-------------|
| **Probo** | Action | Read and write Probo resources (tasks, documents, controls, organizations, and 30+ other resources) |
| **Probo Trigger** | Trigger | Start a workflow when Probo webhook events occur (document published, user created, obligation updated, and more) |

## Requirements

- A self-hosted n8n instance with [community nodes enabled](https://docs.n8n.io/hosting/configuration/configuration-examples/community-nodes/)
- A Probo account with an API key
- n8n 1.0+ (uses the community node package format)

## Installation

Install the package from npm on your self-hosted n8n instance. Only users with the **Owner** or **Admin** role can install community nodes.

### GUI installation (recommended)

1. In n8n, go to **Settings → Community Nodes**.
2. Click **Install**.
3. Enter the npm package name:

   ```
   @probo/n8n-nodes-probo
   ```

   To pin a specific version, append it (for example `@probo/n8n-nodes-probo@0.199.0`).

4. Accept the community node risk notice and click **Install**.
5. Restart n8n if the new nodes do not appear in the node palette immediately.

See the [n8n GUI installation guide](https://docs.n8n.io/integrations/community-nodes/installation-and-management/gui-installation/) for details.

### Manual installation

If you run n8n in Docker or queue mode, you can install the package manually:

```bash
mkdir -p ~/.n8n/nodes
cd ~/.n8n/nodes
npm install @probo/n8n-nodes-probo
```

Restart n8n after installation. See the [manual installation guide](https://docs.n8n.io/integrations/community-nodes/installation/manual-install/) for upgrade and downgrade steps.

## Credentials

All Probo nodes use the **Probo API** credential type. The node sends your API key as a `Bearer` token on every request.

### Configure credentials in n8n

1. Add a **Probo** or **Probo Trigger** node to a workflow.
2. Open the **Credential** dropdown and select **Create New Credential**.
3. Fill in the fields:

   | Field | Default | Description |
   |-------|---------|-------------|
   | **Probo Server** | `https://us.probo.com` | Base URL of your Probo instance. Use `https://eu.probo.com` for the EU region, or your own URL when self-hosting. |
   | **API Key** | — | A Probo API key with access to the organizations you automate against. |

4. Click **Test** to verify connectivity. n8n calls the Probo GraphQL API and checks that the key is valid.
5. Click **Save**. The credential is shared across all Probo nodes in your instance.

### Get an API key

1. Sign in to the Probo console.
2. Go to **Settings → API Keys**.
3. Click **Create API Key**.
4. Copy the key immediately — it is shown only once.

For self-hosted Probo, set **Probo Server** to your instance URL (for example `https://probo.example.com`). The node talks to `/api/console/v1/graphql` on that host.

More detail: [Probo n8n authentication docs](https://www.probo.com/docs/api/n8n/authentication).

## Workflow example: notify Slack when a document is published

This workflow listens for Probo document events and posts a message to Slack.

```
Probo Trigger  →  Slack
(document      (post message
 published)     with document name)
```

### Steps

1. **Create credentials** as described above and save them as `Probo API`.

2. **Add a Probo Trigger node**
   - **Credential:** Probo API
   - **Organization ID:** your Probo organization GID (for example `gid://probo/Organization/…`)
   - **Events:** `Document Version Published`
   - **Verify Signature:** enabled (recommended)

3. **Add a Slack node** (or any notification node) connected to the trigger output.
   - Map fields from the webhook payload, for example:
     - **Text:** `A document was published: {{ $json.data.documentVersion.document.name }}`

4. **Activate the workflow.** n8n registers a webhook subscription in Probo. When a document version is published, Probo delivers the event and the Slack message is sent.

### Update events carry the previous state

For `*:updated` events, the payload includes an `updatedFrom` object next to `data`, holding a full snapshot of the entity as it was before the update. This lets a workflow react to what actually changed — for example, only notify when a user's role changes:

- **Condition:** `{{ $json.data.membership.role !== $json.updatedFrom.membership.role }}`
- **Text:** `Role changed from {{ $json.updatedFrom.membership.role }} to {{ $json.data.membership.role }}`

`updatedFrom` is present only on update events; it is absent for created, deleted, and other lifecycle events.

### Alternative: list open tasks on a schedule

Use the **Probo** action node without a trigger:

1. Add a **Schedule Trigger** node (for example, every weekday at 9:00).
2. Add a **Probo** node:
   - **Resource:** Task
   - **Operation:** Get Many
   - **Organization ID:** your organization GID
   - **Return All:** enabled (or set a **Limit**)
3. Add a downstream node (Slack, email, or spreadsheet) to process `$json` task records.

This pattern works well for daily compliance standups or overdue-task digests.

## Resources

The **Probo** node exposes operations across the platform, including:

Access Review, Asset, Audit, Audit Log, Control, Cookie Banner, Cookie Category, Cookie Consent Record, Data, Document, DPIA, Evidence, Finding, Framework, Measure, Obligation, Organization, Processing Activity, Risk, Task, Third Party, Trust Center, User, Vendor, and more.

Use the **Execute** resource to run custom GraphQL queries or mutations when a dedicated operation is not available.

## Links

- [Probo documentation](https://www.probo.com/docs)
- [Probo n8n authentication](https://www.probo.com/docs/api/n8n/authentication)
- [Package on npm](https://www.npmjs.com/package/@probo/n8n-nodes-probo)
- [Source code](https://github.com/getprobo/probo/tree/main/packages/n8n-node)
- [Report an issue](https://github.com/getprobo/probo/issues)

## License

MIT
