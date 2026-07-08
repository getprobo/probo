// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package membershipprofile

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

// The mapping CTE pairs each manual profile with the one SCIM profile it
// should be merged into: same normalized name and both real IAM members (have
// a role). Creation order does not matter: the manual profile may predate or
// postdate the SCIM profile.

// previewSQL lists the profiles that would be merged.
const previewSQL = `
WITH manual AS (
    SELECT p.id, p.identity_id, p.full_name, p.created_at, lower(btrim(p.full_name)) AS norm_name
    FROM iam_membership_profiles p
    WHERE p.organization_id = @org_id AND p.source = 'MANUAL'
      AND EXISTS (
          SELECT 1 FROM iam_memberships mem
          WHERE mem.identity_id     = p.identity_id
            AND mem.organization_id = p.organization_id
      )
),
scim AS (
    SELECT p.id, p.created_at, lower(btrim(p.full_name)) AS norm_name
    FROM iam_membership_profiles p
    WHERE p.organization_id = @org_id AND p.source = 'SCIM'
      AND EXISTS (
          SELECT 1 FROM iam_memberships mem
          WHERE mem.identity_id     = p.identity_id
            AND mem.organization_id = p.organization_id
      )
),
scim_unique AS (
    SELECT norm_name FROM scim GROUP BY norm_name HAVING COUNT(*) = 1
),
mapping AS (
    SELECT m.id AS manual_id, m.identity_id AS manual_identity_id, m.full_name, s.id AS scim_id
    FROM manual m
    JOIN scim s        ON s.norm_name = m.norm_name
    JOIN scim_unique u ON u.norm_name = m.norm_name
)
SELECT
    m.full_name       AS full_name,
    i.email_address   AS email,
    m.manual_id       AS manual_id,
    m.scim_id         AS scim_id
FROM mapping m
JOIN identities i ON i.id = m.manual_identity_id
ORDER BY m.full_name`

// mergeSQL repoints every reference from the manual id to the SCIM id in one
// statement. Nothing is ever deleted. For the three slots that must stay unique
// per parent (document_default_approvers, document_version_approval_decisions
// and document_version_signatures) a repoint could land a second row on the
// SCIM profile; there we repoint only the single top row per (parent, SCIM
// profile) and leave any manual row that would create a duplicate untouched
// (still pointing at the manual profile). Every other reference is a plain
// repoint. The whole thing runs in one statement, so it is all-or-nothing.
const mergeSQL = `
WITH manual AS (
    SELECT p.id, p.created_at, lower(btrim(p.full_name)) AS norm_name
    FROM iam_membership_profiles p
    WHERE p.organization_id = @org_id AND p.source = 'MANUAL'
      AND EXISTS (
          SELECT 1 FROM iam_memberships mem
          WHERE mem.identity_id     = p.identity_id
            AND mem.organization_id = p.organization_id
      )
),
scim AS (
    SELECT p.id, p.created_at, lower(btrim(p.full_name)) AS norm_name
    FROM iam_membership_profiles p
    WHERE p.organization_id = @org_id AND p.source = 'SCIM'
      AND EXISTS (
          SELECT 1 FROM iam_memberships mem
          WHERE mem.identity_id     = p.identity_id
            AND mem.organization_id = p.organization_id
      )
),
scim_unique AS (
    SELECT norm_name FROM scim GROUP BY norm_name HAVING COUNT(*) = 1
),
mapping AS (
    SELECT m.id AS manual_id, s.id AS scim_id
    FROM manual m
    JOIN scim s        ON s.norm_name = m.norm_name
    JOIN scim_unique u ON u.norm_name = m.norm_name
),
u_assets AS (
    UPDATE assets t SET owner_profile_id = m.scim_id
    FROM mapping m WHERE t.owner_profile_id = m.manual_id RETURNING 1
),
u_data AS (
    UPDATE data t SET owner_profile_id = m.scim_id
    FROM mapping m WHERE t.owner_profile_id = m.manual_id RETURNING 1
),
u_obligations AS (
    UPDATE obligations t SET owner_profile_id = m.scim_id
    FROM mapping m WHERE t.owner_profile_id = m.manual_id RETURNING 1
),
u_risks AS (
    UPDATE risks t SET owner_profile_id = m.scim_id
    FROM mapping m WHERE t.owner_profile_id = m.manual_id RETURNING 1
),
u_soa AS (
    UPDATE statements_of_applicability t SET owner_profile_id = m.scim_id
    FROM mapping m WHERE t.owner_profile_id = m.manual_id RETURNING 1
),
u_tasks AS (
    UPDATE tasks t SET assigned_to_profile_id = m.scim_id
    FROM mapping m WHERE t.assigned_to_profile_id = m.manual_id RETURNING 1
),
u_third_parties AS (
    -- both owner columns in one UPDATE so a row is never modified twice
    UPDATE third_parties t SET
        business_owner_profile_id = COALESCE(mb.scim_id, t.business_owner_profile_id),
        security_owner_profile_id = COALESCE(ms.scim_id, t.security_owner_profile_id)
    FROM third_parties src
    LEFT JOIN mapping mb ON mb.manual_id = src.business_owner_profile_id
    LEFT JOIN mapping ms ON ms.manual_id = src.security_owner_profile_id
    WHERE src.id = t.id
      AND (mb.scim_id IS NOT NULL OR ms.scim_id IS NOT NULL)
    RETURNING 1
),
u_processing AS (
    UPDATE processing_activities t SET dpo_profile_id = m.scim_id
    FROM mapping m WHERE t.dpo_profile_id = m.manual_id RETURNING 1
),
u_findings AS (
    UPDATE findings t SET owner_id = m.scim_id
    FROM mapping m WHERE t.owner_id = m.manual_id RETURNING 1
),
-- document_default_approvers: unique per (document_id, approver). Rank every
-- row in each (document_id, post-merge profile) group with the SCIM-side row
-- first. Only the top manual row is repointed; any manual row that would land
-- on an already-taken slot (rn > 1) is left untouched (never moved, never
-- deleted), so no duplicate is created.
dda_ranked AS (
    SELECT t.document_id,
           t.approver_profile_id,
           (m.manual_id IS NOT NULL) AS is_manual,
           row_number() OVER (
               PARTITION BY t.document_id, COALESCE(m.scim_id, t.approver_profile_id)
               ORDER BY (m.manual_id IS NOT NULL), t.created_at, t.approver_profile_id
           ) AS rn
    FROM document_default_approvers t
    LEFT JOIN mapping m ON m.manual_id = t.approver_profile_id
    WHERE t.organization_id = @org_id
),
u_default_approvers AS (
    UPDATE document_default_approvers t SET approver_profile_id = m.scim_id
    FROM mapping m, dda_ranked r
    WHERE t.document_id = r.document_id AND t.approver_profile_id = r.approver_profile_id
      AND r.is_manual AND r.rn = 1
      AND t.approver_profile_id = m.manual_id
    RETURNING 1
),
-- document_version_approval_decisions: unique per (quorum_id, approver).
dvad_ranked AS (
    SELECT t.id,
           (m.manual_id IS NOT NULL) AS is_manual,
           row_number() OVER (
               PARTITION BY t.quorum_id, COALESCE(m.scim_id, t.approver_id)
               ORDER BY (m.manual_id IS NOT NULL), t.created_at, t.id
           ) AS rn
    FROM document_version_approval_decisions t
    LEFT JOIN mapping m ON m.manual_id = t.approver_id
    WHERE t.organization_id = @org_id
),
u_approval_decisions AS (
    UPDATE document_version_approval_decisions t SET approver_id = m.scim_id
    FROM mapping m, dvad_ranked r
    WHERE t.id = r.id AND r.is_manual AND r.rn = 1
      AND t.approver_id = m.manual_id
    RETURNING 1
),
-- document_version_signatures: one signature per (document, major, signatory).
-- No DB constraint enforces it, but we treat it as unique: only the top manual
-- signature (preferring SIGNED then most recent) is repointed; any that would
-- duplicate an already-present signature is left untouched.
sig_ranked AS (
    SELECT t.id,
           (m.manual_id IS NOT NULL) AS is_manual,
           row_number() OVER (
               PARTITION BY dv.document_id, dv.major, COALESCE(m.scim_id, t.signed_by_profile_id)
               ORDER BY (m.manual_id IS NOT NULL),
                        CASE t.state WHEN 'SIGNED' THEN 0 ELSE 1 END,
                        t.created_at DESC,
                        t.id DESC
           ) AS rn
    FROM document_version_signatures t
    JOIN document_versions dv ON dv.id = t.document_version_id
    LEFT JOIN mapping m ON m.manual_id = t.signed_by_profile_id
    WHERE t.organization_id = @org_id
),
u_signatures AS (
    UPDATE document_version_signatures t SET signed_by_profile_id = m.scim_id
    FROM mapping m, sig_ranked r
    WHERE t.id = r.id AND r.is_manual AND r.rn = 1
      AND t.signed_by_profile_id = m.manual_id
    RETURNING 1
)
SELECT 'assets.owner_profile_id'                        AS target, (SELECT count(*) FROM u_assets)  AS moved, 0::bigint AS skipped
UNION ALL SELECT 'data.owner_profile_id',                          (SELECT count(*) FROM u_data),                0
UNION ALL SELECT 'obligations.owner_profile_id',                   (SELECT count(*) FROM u_obligations),         0
UNION ALL SELECT 'risks.owner_profile_id',                         (SELECT count(*) FROM u_risks),               0
UNION ALL SELECT 'statements_of_applicability.owner_profile_id',   (SELECT count(*) FROM u_soa),                 0
UNION ALL SELECT 'tasks.assigned_to_profile_id',                   (SELECT count(*) FROM u_tasks),               0
UNION ALL SELECT 'third_parties.business/security_owner',          (SELECT count(*) FROM u_third_parties),       0
UNION ALL SELECT 'processing_activities.dpo_profile_id',           (SELECT count(*) FROM u_processing),          0
UNION ALL SELECT 'findings.owner_id',                              (SELECT count(*) FROM u_findings),            0
UNION ALL SELECT 'document_default_approvers.approver_profile_id', (SELECT count(*) FROM u_default_approvers),   (SELECT count(*) FROM dda_ranked WHERE is_manual AND rn > 1)
UNION ALL SELECT 'document_version_approval_decisions.approver_id',(SELECT count(*) FROM u_approval_decisions),  (SELECT count(*) FROM dvad_ranked WHERE is_manual AND rn > 1)
UNION ALL SELECT 'document_version_signatures.signed_by_profile_id',(SELECT count(*) FROM u_signatures),         (SELECT count(*) FROM sig_ranked WHERE is_manual AND rn > 1)`

// previewCountsSQL is the read-only twin of mergeSQL: instead of writing, it
// reports, per target, how many rows would be repointed onto the SCIM id
// (moved) and how many manual-side rows would be left untouched because moving
// them would create a duplicate (skipped). The three unique-slot tables reuse
// the exact same ranking as mergeSQL so the numbers match the apply step.
const previewCountsSQL = `
WITH manual AS (
    SELECT p.id, lower(btrim(p.full_name)) AS norm_name
    FROM iam_membership_profiles p
    WHERE p.organization_id = @org_id AND p.source = 'MANUAL'
      AND EXISTS (
          SELECT 1 FROM iam_memberships mem
          WHERE mem.identity_id     = p.identity_id
            AND mem.organization_id = p.organization_id
      )
),
scim AS (
    SELECT p.id, lower(btrim(p.full_name)) AS norm_name
    FROM iam_membership_profiles p
    WHERE p.organization_id = @org_id AND p.source = 'SCIM'
      AND EXISTS (
          SELECT 1 FROM iam_memberships mem
          WHERE mem.identity_id     = p.identity_id
            AND mem.organization_id = p.organization_id
      )
),
scim_unique AS (
    SELECT norm_name FROM scim GROUP BY norm_name HAVING COUNT(*) = 1
),
mapping AS (
    SELECT m.id AS manual_id, s.id AS scim_id
    FROM manual m
    JOIN scim s        ON s.norm_name = m.norm_name
    JOIN scim_unique u ON u.norm_name = m.norm_name
),
dda_ranked AS (
    SELECT (m.manual_id IS NOT NULL) AS is_manual,
           row_number() OVER (
               PARTITION BY t.document_id, COALESCE(m.scim_id, t.approver_profile_id)
               ORDER BY (m.manual_id IS NOT NULL), t.created_at, t.approver_profile_id
           ) AS rn
    FROM document_default_approvers t
    LEFT JOIN mapping m ON m.manual_id = t.approver_profile_id
    WHERE t.organization_id = @org_id
),
dvad_ranked AS (
    SELECT (m.manual_id IS NOT NULL) AS is_manual,
           row_number() OVER (
               PARTITION BY t.quorum_id, COALESCE(m.scim_id, t.approver_id)
               ORDER BY (m.manual_id IS NOT NULL), t.created_at, t.id
           ) AS rn
    FROM document_version_approval_decisions t
    LEFT JOIN mapping m ON m.manual_id = t.approver_id
    WHERE t.organization_id = @org_id
),
sig_ranked AS (
    SELECT (m.manual_id IS NOT NULL) AS is_manual,
           row_number() OVER (
               PARTITION BY dv.document_id, dv.major, COALESCE(m.scim_id, t.signed_by_profile_id)
               ORDER BY (m.manual_id IS NOT NULL),
                        CASE t.state WHEN 'SIGNED' THEN 0 ELSE 1 END,
                        t.created_at DESC,
                        t.id DESC
           ) AS rn
    FROM document_version_signatures t
    JOIN document_versions dv ON dv.id = t.document_version_id
    LEFT JOIN mapping m ON m.manual_id = t.signed_by_profile_id
    WHERE t.organization_id = @org_id
)
SELECT 'assets.owner_profile_id'                         AS target,
       (SELECT count(*) FROM assets t                        JOIN mapping m ON t.owner_profile_id       = m.manual_id) AS moved,
       0::bigint AS skipped
UNION ALL SELECT 'data.owner_profile_id',
       (SELECT count(*) FROM data t                          JOIN mapping m ON t.owner_profile_id       = m.manual_id), 0
UNION ALL SELECT 'obligations.owner_profile_id',
       (SELECT count(*) FROM obligations t                   JOIN mapping m ON t.owner_profile_id       = m.manual_id), 0
UNION ALL SELECT 'risks.owner_profile_id',
       (SELECT count(*) FROM risks t                         JOIN mapping m ON t.owner_profile_id       = m.manual_id), 0
UNION ALL SELECT 'statements_of_applicability.owner_profile_id',
       (SELECT count(*) FROM statements_of_applicability t   JOIN mapping m ON t.owner_profile_id       = m.manual_id), 0
UNION ALL SELECT 'tasks.assigned_to_profile_id',
       (SELECT count(*) FROM tasks t                         JOIN mapping m ON t.assigned_to_profile_id = m.manual_id), 0
UNION ALL SELECT 'third_parties.business/security_owner',
       (SELECT count(*) FROM third_parties t
        WHERE t.business_owner_profile_id IN (SELECT manual_id FROM mapping)
           OR t.security_owner_profile_id IN (SELECT manual_id FROM mapping)), 0
UNION ALL SELECT 'processing_activities.dpo_profile_id',
       (SELECT count(*) FROM processing_activities t         JOIN mapping m ON t.dpo_profile_id         = m.manual_id), 0
UNION ALL SELECT 'findings.owner_id',
       (SELECT count(*) FROM findings t                      JOIN mapping m ON t.owner_id               = m.manual_id), 0
UNION ALL SELECT 'document_default_approvers.approver_profile_id',
       (SELECT count(*) FROM dda_ranked WHERE is_manual AND rn = 1),
       (SELECT count(*) FROM dda_ranked WHERE is_manual AND rn > 1)
UNION ALL SELECT 'document_version_approval_decisions.approver_id',
       (SELECT count(*) FROM dvad_ranked WHERE is_manual AND rn = 1),
       (SELECT count(*) FROM dvad_ranked WHERE is_manual AND rn > 1)
UNION ALL SELECT 'document_version_signatures.signed_by_profile_id',
       (SELECT count(*) FROM sig_ranked WHERE is_manual AND rn = 1),
       (SELECT count(*) FROM sig_ranked WHERE is_manual AND rn > 1)`

// collisionsSQL lists the unique slots where a repoint would land two rows on
// the same SCIM profile, i.e. a "double". It projects each child row onto its
// post-merge profile (COALESCE(scim_id, current)) and keeps the groups that end
// up with more than one row and that the merge itself creates (bool_or(mapped)).
// These are exactly the slots where mergeSQL leaves the manual-side row
// untouched; this query surfaces them (with names/scopes) for the operator.
const collisionsSQL = `
WITH manual AS (
    SELECT p.id, lower(btrim(p.full_name)) AS norm_name, p.created_at
    FROM iam_membership_profiles p
    WHERE p.organization_id = @org_id AND p.source = 'MANUAL'
      AND EXISTS (
          SELECT 1 FROM iam_memberships mem
          WHERE mem.identity_id     = p.identity_id
            AND mem.organization_id = p.organization_id
      )
),
scim AS (
    SELECT p.id, lower(btrim(p.full_name)) AS norm_name, p.created_at
    FROM iam_membership_profiles p
    WHERE p.organization_id = @org_id AND p.source = 'SCIM'
      AND EXISTS (
          SELECT 1 FROM iam_memberships mem
          WHERE mem.identity_id     = p.identity_id
            AND mem.organization_id = p.organization_id
      )
),
scim_unique AS (
    SELECT norm_name FROM scim GROUP BY norm_name HAVING COUNT(*) = 1
),
mapping AS (
    SELECT m.id AS manual_id, s.id AS scim_id
    FROM manual m
    JOIN scim s        ON s.norm_name = m.norm_name
    JOIN scim_unique u ON u.norm_name = m.norm_name
),
approvers AS (
    SELECT t.document_id::text AS scope,
           COALESCE(m.scim_id, t.approver_profile_id) AS eff_profile,
           (m.manual_id IS NOT NULL) AS mapped
    FROM document_default_approvers t
    LEFT JOIN mapping m ON m.manual_id = t.approver_profile_id
    WHERE t.organization_id = @org_id
),
decisions AS (
    SELECT t.quorum_id::text AS scope,
           COALESCE(m.scim_id, t.approver_id) AS eff_profile,
           (m.manual_id IS NOT NULL) AS mapped
    FROM document_version_approval_decisions t
    LEFT JOIN mapping m ON m.manual_id = t.approver_id
    WHERE t.organization_id = @org_id
),
signatures AS (
    SELECT dv.document_id || ':' || dv.major::text AS scope,
           COALESCE(m.scim_id, t.signed_by_profile_id) AS eff_profile,
           (m.manual_id IS NOT NULL) AS mapped
    FROM document_version_signatures t
    JOIN document_versions dv ON dv.id = t.document_version_id
    LEFT JOIN mapping m ON m.manual_id = t.signed_by_profile_id
    WHERE t.organization_id = @org_id
),
collisions AS (
    SELECT 'document_default_approvers (document_id, approver)' AS target,
           scope, eff_profile, count(*) AS dup
    FROM approvers GROUP BY scope, eff_profile HAVING count(*) > 1 AND bool_or(mapped)
    UNION ALL
    SELECT 'document_version_approval_decisions (quorum_id, approver)',
           scope, eff_profile, count(*)
    FROM decisions GROUP BY scope, eff_profile HAVING count(*) > 1 AND bool_or(mapped)
    UNION ALL
    SELECT 'document_version_signatures (document, major, signatory)',
           scope, eff_profile, count(*)
    FROM signatures GROUP BY scope, eff_profile HAVING count(*) > 1 AND bool_or(mapped)
)
SELECT c.target       AS target,
       c.scope        AS scope,
       c.eff_profile  AS scim_id,
       p.full_name    AS full_name,
       c.dup          AS dup_count
FROM collisions c
JOIN iam_membership_profiles p ON p.id = c.eff_profile
ORDER BY c.target, c.scope`

type (
	mappingRow struct {
		FullName string `db:"full_name"`
		Email    string `db:"email"`
		ManualID string `db:"manual_id"`
		ScimID   string `db:"scim_id"`
	}

	resultRow struct {
		Target  string `db:"target"`
		Moved   int64  `db:"moved"`
		Skipped int64  `db:"skipped"`
	}

	collisionRow struct {
		Target   string `db:"target"`
		Scope    string `db:"scope"`
		ScimID   string `db:"scim_id"`
		FullName string `db:"full_name"`
		DupCount int64  `db:"dup_count"`
	}
)

func newCmdMergeManualIntoSCIM(f *cmdutil.Factory) *cobra.Command {
	var (
		flagDryRun bool
		flagYes    bool
	)

	cmd := &cobra.Command{
		Use:   "merge-manual-into-scim <organization-gid>",
		Short: "Repoint all references from MANUAL profiles onto the matching SCIM profile",
		Long: "Destructive, organization-scoped operator action. For the given organization, " +
			"every MANUAL membership profile whose normalized full_name matches exactly one " +
			"SCIM profile has all of its references (ownership, task assignment, signatures, " +
			"approvals, etc.) repointed onto that SCIM profile. Both profiles must be actual " +
			"IAM members (their identity has a role in the org, not a compliance-page-only " +
			"people record); creation order does not matter (the manual profile may predate " +
			"or postdate the SCIM profile). References are repointed in a single statement; " +
			"nothing is ever deleted. For the slots that must stay unique per parent (default " +
			"approvers, approval decisions and signatures), if a repoint would put the SCIM " +
			"profile on the slot twice the manual-side row is left untouched instead of moved, " +
			"so no duplicate is created. With --dry-run the mapping, the per-target " +
			"move/skip counts and the duplicate slots are printed without writing.",
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the manual -> scim mapping without writing")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation and apply the merge")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		orgID, err := gid.ParseGID(args[0])
		if err != nil {
			return fmt.Errorf("invalid organization GID %q: %w", args[0], err)
		}

		ctx := cmd.Context()
		out := f.IOStreams.Out

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		orgArgs := pgx.StrictNamedArgs{"org_id": orgID.String()}

		var (
			orgName    string
			mappings   []mappingRow
			moves      []resultRow
			collisions []collisionRow
			manualN    int
		)

		if err := pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				if err := conn.QueryRow(
					ctx,
					`SELECT name FROM organizations WHERE id = @org_id`,
					orgArgs,
				).Scan(&orgName); err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						return fmt.Errorf("organization %s not found", orgID)
					}

					return fmt.Errorf("cannot load organization: %w", err)
				}

				if err := conn.QueryRow(
					ctx,
					`SELECT count(*)
FROM iam_membership_profiles p
WHERE p.organization_id = @org_id
  AND p.source = 'MANUAL'
  AND EXISTS (
      SELECT 1 FROM iam_memberships mem
      WHERE mem.identity_id = p.identity_id
        AND mem.organization_id = p.organization_id
  )`,
					orgArgs,
				).Scan(&manualN); err != nil {
					return fmt.Errorf("cannot count manual profiles: %w", err)
				}

				rows, err := conn.Query(ctx, previewSQL, orgArgs)
				if err != nil {
					return fmt.Errorf("cannot compute merge mapping: %w", err)
				}
				defer rows.Close()

				mappings, err = pgx.CollectRows(rows, pgx.RowToStructByName[mappingRow])
				if err != nil {
					return fmt.Errorf("cannot collect merge mapping: %w", err)
				}

				moveRows, err := conn.Query(ctx, previewCountsSQL, orgArgs)
				if err != nil {
					return fmt.Errorf("cannot compute move counts: %w", err)
				}
				defer moveRows.Close()

				moves, err = pgx.CollectRows(moveRows, pgx.RowToStructByName[resultRow])
				if err != nil {
					return fmt.Errorf("cannot collect move counts: %w", err)
				}

				collisionRows, err := conn.Query(ctx, collisionsSQL, orgArgs)
				if err != nil {
					return fmt.Errorf("cannot detect merge collisions: %w", err)
				}
				defer collisionRows.Close()

				collisions, err = pgx.CollectRows(collisionRows, pgx.RowToStructByName[collisionRow])
				if err != nil {
					return fmt.Errorf("cannot collect merge collisions: %w", err)
				}

				return nil
			},
		); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(out, "Organization: %s (%s)\n", orgName, orgID)

		if len(mappings) == 0 {
			_, _ = fmt.Fprintf(out, "No manual profiles with a unique matching SCIM profile. Nothing to merge.\n")
			return nil
		}

		preview := clicmdutil.NewTable("FULL NAME", "EMAIL", "MANUAL PROFILE", "SCIM PROFILE")
		for _, m := range mappings {
			preview.Row(m.FullName, m.Email, m.ManualID, m.ScimID)
		}

		_, _ = fmt.Fprintln(out, preview.Render())
		_, _ = fmt.Fprintf(
			out,
			"%d of %d manual profile(s) will be merged (%d skipped: no unique SCIM match).\n",
			len(mappings),
			manualN,
			manualN-len(mappings),
		)

		var totalMoves, totalSkipped int64

		movesTable := clicmdutil.NewTable("TARGET", "ROWS TO MOVE", "DUPLICATES SKIPPED")
		for _, m := range moves {
			movesTable.Row(m.Target, fmt.Sprintf("%d", m.Moved), fmt.Sprintf("%d", m.Skipped))
			totalMoves += m.Moved
			totalSkipped += m.Skipped
		}

		_, _ = fmt.Fprintln(out, movesTable.Render())
		_, _ = fmt.Fprintf(
			out,
			"%d reference(s) will be repointed onto SCIM profiles; %d manual-side row(s) will be left as-is to avoid duplicates.\n",
			totalMoves,
			totalSkipped,
		)

		if len(collisions) > 0 {
			collisionTable := clicmdutil.NewTable("SLOT", "SCOPE", "SCIM PROFILE", "FULL NAME", "ROWS")
			for _, c := range collisions {
				collisionTable.Row(c.Target, c.Scope, c.ScimID, c.FullName, fmt.Sprintf("%d", c.DupCount))
			}

			_, _ = fmt.Fprintf(out, "\n%d slot(s) would end up with the SCIM profile twice, so the manual-side row is a duplicate:\n", len(collisions))
			_, _ = fmt.Fprintln(out, collisionTable.Render())
			_, _ = fmt.Fprintf(out, "In each of these the manual-side row is left untouched (not moved, not deleted) "+
				"to avoid the duplicate (counted under DUPLICATES SKIPPED above).\n")
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(out, "Dry run: no changes written.\n")
			return nil
		}

		if !flagYes {
			return fmt.Errorf("about to merge %d manual profile(s) in organization %s; pass --yes to proceed or --dry-run to preview", len(mappings), orgID)
		}

		var results []resultRow

		if err := pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				rows, err := conn.Query(ctx, mergeSQL, orgArgs)
				if err != nil {
					return mergeFailure(err)
				}
				defer rows.Close()

				results, err = pgx.CollectRows(rows, pgx.RowToStructByName[resultRow])
				if err != nil {
					return mergeFailure(err)
				}

				return nil
			},
		); err != nil {
			return err
		}

		var movedTotal, skippedTotal int64

		summary := clicmdutil.NewTable("TARGET", "MOVED", "SKIPPED")
		for _, r := range results {
			summary.Row(r.Target, fmt.Sprintf("%d", r.Moved), fmt.Sprintf("%d", r.Skipped))
			movedTotal += r.Moved
			skippedTotal += r.Skipped
		}

		_, _ = fmt.Fprintln(out, summary.Render())
		_, _ = fmt.Fprintf(
			out,
			"Merged %d manual profile(s) into their SCIM counterparts: %d reference(s) repointed, %d manual-side row(s) left as-is to avoid duplicates.\n",
			len(mappings),
			movedTotal,
			skippedTotal,
		)

		return nil
	}

	return cmd
}

// mergeFailure surfaces a violated constraint if one is somehow hit. The merge
// skips repoints that would duplicate a known unique slot, so this is a safety
// net for unexpected constraints; it names the constraint/table to investigate.
func mergeFailure(err error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.ConstraintName != "" {
		return fmt.Errorf(
			"cannot apply merge: constraint %q on %q violated: %s (reconcile the conflicting manual/SCIM pair, then retry): %w",
			pgErr.ConstraintName,
			pgErr.TableName,
			pgErr.Detail,
			err,
		)
	}

	return fmt.Errorf("cannot apply merge: %w", err)
}
