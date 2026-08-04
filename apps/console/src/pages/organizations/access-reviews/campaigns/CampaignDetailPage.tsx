// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { formatError } from "@probo/helpers";
import { useList } from "@probo/hooks";
import {
  Breadcrumb,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconPlusLarge,
  IconTrashCan,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery, useRelayEnvironment } from "react-relay";
import { useNavigate } from "react-router";
import { ConnectionHandler, fetchQuery, graphql } from "relay-runtime";

import type { AccessReviewEntryDecision, CampaignDetailPageBulkDecisionMutation } from "#/__generated__/core/CampaignDetailPageBulkDecisionMutation.graphql";
import type { AccessReviewEntryFlag, CampaignDetailPageBulkFlagMutation } from "#/__generated__/core/CampaignDetailPageBulkFlagMutation.graphql";
import type { CampaignDetailPageCloseMutation } from "#/__generated__/core/CampaignDetailPageCloseMutation.graphql";
import type { CampaignDetailPageDeleteMutation } from "#/__generated__/core/CampaignDetailPageDeleteMutation.graphql";
import type { CampaignDetailPageQuery } from "#/__generated__/core/CampaignDetailPageQuery.graphql";
import type { CampaignDetailPageStartMutation } from "#/__generated__/core/CampaignDetailPageStartMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AddCampaignSourceDialog } from "../dialogs/AddCampaignSourceDialog";

import { AccessEntriesSelectionBar } from "./_components/AccessEntriesSelectionBar";
import { AccessEntriesToolbar } from "./_components/AccessEntriesToolbar";
import { AccessEntryListItem } from "./_components/AccessEntryListItem";
import { AccessEntrySection } from "./_components/AccessEntrySection";
import { accessEntriesLayout, accessEntryList } from "./_components/variants";

const startCampaignMutation = graphql`
  mutation CampaignDetailPageStartMutation(
    $input: StartAccessReviewCampaignInput!
  ) {
    startAccessReviewCampaign(input: $input) {
      accessReviewCampaign {
        id
        status
        startedAt
      }
    }
  }
`;

const closeCampaignMutation = graphql`
  mutation CampaignDetailPageCloseMutation(
    $input: CloseAccessReviewCampaignInput!
  ) {
    closeAccessReviewCampaign(input: $input) {
      accessReviewCampaign {
        id
        status
        completedAt
      }
    }
  }
`;

const deleteCampaignMutation = graphql`
  mutation CampaignDetailPageDeleteMutation(
    $input: DeleteAccessReviewCampaignInput!
    $connections: [ID!]!
  ) {
    deleteAccessReviewCampaign(input: $input) {
      deletedAccessReviewCampaignId @deleteEdge(connections: $connections)
    }
  }
`;

const bulkDecisionMutation = graphql`
  mutation CampaignDetailPageBulkDecisionMutation(
    $input: RecordAccessReviewEntryDecisionsInput!
  ) {
    recordAccessReviewEntryDecisions(input: $input) {
      accessReviewEntries {
        id
        decision
        decisionNote
      }
    }
  }
`;

const bulkFlagMutation = graphql`
  mutation CampaignDetailPageBulkFlagMutation(
    $input: FlagAccessReviewEntryInput!
  ) {
    flagAccessReviewEntry(input: $input) {
      accessReviewEntry {
        id
        flags
        flagReasons
      }
    }
  }
`;

export const campaignDetailPageQuery = graphql`
  query CampaignDetailPageQuery($campaignId: ID!) {
    node(id: $campaignId) {
      __typename
      ... on AccessReviewCampaign {
        id
        name
        status
        pendingEntries: entries(first: 0, filter: { decision: PENDING }) {
          totalCount
        }
        canDelete: permission(action: "access-review:campaign:delete")
        sources {
          id
          source {
            id
            connector {
              provider
            }
          }
          name
          fetchAttempts(first: 1) {
            edges {
              node {
                status
                error
              }
            }
          }
          entries(first: 500) {
            edges {
              node {
                id
                email
                fullName
                isAdmin
                mfaStatus
                ...AccessEntryListItem_entry
              }
            }
            pageInfo {
              hasNextPage
            }
          }
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<CampaignDetailPageQuery>;
};

type CampaignSource = NonNullable<
  Extract<
    CampaignDetailPageQuery["response"]["node"],
    { readonly __typename: "AccessReviewCampaign" }
  >["sources"]
>[number];

type CampaignEntryEdge = CampaignSource["entries"]["edges"][number];

type EntryFilters = {
  email: string;
  connectorIds: ReadonlyArray<string>;
  mfa: ReadonlyArray<string>;
  admin: ReadonlyArray<string>;
};

function entryMatchesFilters(
  edge: CampaignEntryEdge,
  sourceId: string,
  filters: EntryFilters,
): boolean {
  if (filters.connectorIds.length > 0 && !filters.connectorIds.includes(sourceId)) {
    return false;
  }

  const query = filters.email.trim().toLowerCase();
  if (query) {
    const email = edge.node.email.toLowerCase();
    const fullName = edge.node.fullName.toLowerCase();
    if (!email.includes(query) && !fullName.includes(query)) {
      return false;
    }
  }

  if (filters.mfa.length > 0 && !filters.mfa.includes(edge.node.mfaStatus)) {
    return false;
  }

  if (filters.admin.length > 0) {
    const adminValue = edge.node.isAdmin ? "YES" : "NO";
    if (!filters.admin.includes(adminValue)) {
      return false;
    }
  }

  return true;
}

export default function CampaignDetailPage({ queryRef }: Props) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const environment = useRelayEnvironment();
  const data = usePreloadedQuery<CampaignDetailPageQuery>(campaignDetailPageQuery, queryRef);
  const { toast } = useToast();
  const confirm = useConfirm();

  if (data.node.__typename !== "AccessReviewCampaign") {
    throw new Error("Campaign not found");
  }

  const campaign = data.node;
  const isInProgress = campaign.status === "IN_PROGRESS";
  const isDraft = campaign.status === "DRAFT";
  const isPendingActions = campaign.status === "PENDING_ACTIONS";
  const canDelete = campaign.canDelete && !isInProgress;

  const campaignIdRef = useRef(campaign.id);
  const selectionAnchorRef = useRef<string | null>(null);
  const { list: selection, toggle, clear: clearSelectionList, reset } = useList<string>([]);
  const clear = () => {
    selectionAnchorRef.current = null;
    clearSelectionList();
  };
  const [emailFilter, setEmailFilter] = useState("");
  const [connectorFilter, setConnectorFilter] = useState<string[]>([]);
  const [mfaFilter, setMfaFilter] = useState<string[]>([]);
  const [adminFilter, setAdminFilter] = useState<string[]>([]);

  const [bulkDecision, setBulkDecision] = useState<AccessReviewEntryDecision | null>(null);
  const [bulkPendingDecision, setBulkPendingDecision] = useState<AccessReviewEntryDecision | null>(null);
  const [bulkNote, setBulkNote] = useState("");
  const bulkNoteRef = useDialogRef();
  const [bulkFlagSelection, setBulkFlagSelection] = useState<AccessReviewEntryFlag[]>([]);
  const [bulkFlagsDirty, setBulkFlagsDirty] = useState(false);

  const [startCampaign, isStarting]
    = useMutation<CampaignDetailPageStartMutation>(startCampaignMutation);
  const [closeCampaign, isClosing]
    = useMutation<CampaignDetailPageCloseMutation>(closeCampaignMutation);
  const [deleteCampaign, isDeleting]
    = useMutation<CampaignDetailPageDeleteMutation>(deleteCampaignMutation);
  const [bulkDecide, isBulkDeciding]
    = useMutation<CampaignDetailPageBulkDecisionMutation>(bulkDecisionMutation);
  const [bulkFlag, isBulkFlagging]
    = useMutation<CampaignDetailPageBulkFlagMutation>(bulkFlagMutation);
  const isBulkSubmitting = isBulkDeciding || isBulkFlagging;

  // Reset selection/bulk UI when navigating between campaigns (same page instance).
  const [prevCampaignId, setPrevCampaignId] = useState(campaign.id);
  if (campaign.id !== prevCampaignId) {
    setPrevCampaignId(campaign.id);
    clearSelectionList();
    setBulkDecision(null);
    setBulkPendingDecision(null);
    setBulkNote("");
    setBulkFlagSelection([]);
    setBulkFlagsDirty(false);
  }

  // Layout effect so async bulk callbacks see the new id before paint / network replies.
  useLayoutEffect(() => {
    campaignIdRef.current = campaign.id;
    selectionAnchorRef.current = null;
    // Close the note dialog so Confirm cannot submit against a cleared pending decision.
    bulkNoteRef.current?.close();
  }, [campaign.id, bulkNoteRef]);

  useEffect(() => {
    if (!isInProgress) return;
    const interval = setInterval(() => {
      if (document.hidden) return;
      fetchQuery<CampaignDetailPageQuery>(
        environment,
        campaignDetailPageQuery,
        { campaignId: campaignIdRef.current },
        { fetchPolicy: "network-only" },
      ).subscribe({});
    }, 3000);
    return () => clearInterval(interval);
  }, [isInProgress, environment]);

  const existingCampaignSourceIds = useMemo(
    () => campaign.sources.flatMap(s => s.source?.id ? [s.source.id] : []),
    [campaign.sources],
  );

  const filters = useMemo<EntryFilters>(() => ({
    email: emailFilter,
    connectorIds: connectorFilter,
    mfa: mfaFilter,
    admin: adminFilter,
  }), [emailFilter, connectorFilter, mfaFilter, adminFilter]);

  const hasActiveFilters = emailFilter.trim() !== ""
    || connectorFilter.length > 0
    || mfaFilter.length > 0
    || adminFilter.length > 0;

  const filteredSources = useMemo(() => {
    return campaign.sources
      .map(source => ({
        source,
        entries: (source.entries?.edges ?? []).filter(edge =>
          entryMatchesFilters(edge, source.id, filters),
        ),
      }))
      .filter(({ source, entries }) => {
        if (filters.connectorIds.length > 0 && !filters.connectorIds.includes(source.id)) {
          return false;
        }
        if (hasActiveFilters && entries.length === 0) {
          return false;
        }
        return true;
      });
  }, [campaign.sources, filters, hasActiveFilters]);

  const connectorOptions = useMemo(
    () => campaign.sources.map(source => ({
      value: source.id,
      label: source.name,
    })),
    [campaign.sources],
  );

  const canComplete = campaign.pendingEntries.totalCount === 0;

  const handleStart = () => {
    startCampaign({
      variables: { input: { accessReviewCampaignId: campaign.id } },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("campaignDetailPage.messages.error"),
            description: formatError(t("campaignDetailPage.errors.start"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("campaignDetailPage.messages.success"),
          description: t("campaignDetailPage.messages.started"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: t("campaignDetailPage.messages.error"),
          description: formatError(t("campaignDetailPage.errors.start"), error),
          variant: "error",
        });
      },
    });
  };

  const handleDelete = () => {
    const connections = [
      ConnectionHandler.getConnectionID(
        organizationId,
        "AccessReviewCampaignsTab_accessReviewCampaigns",
      ),
    ];
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteCampaign({
            variables: {
              input: { accessReviewCampaignId: campaign.id },
              connections,
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: t("campaignDetailPage.messages.error"),
                  description: formatError(t("campaignDetailPage.errors.delete"), errors),
                  variant: "error",
                });
                resolve();
                return;
              }
              toast({
                title: t("campaignDetailPage.messages.success"),
                description: t("campaignDetailPage.messages.deleted"),
                variant: "success",
              });
              resolve();
              void navigate(`/organizations/${organizationId}/access-reviews`);
            },
            onError(error) {
              toast({
                title: t("campaignDetailPage.messages.error"),
                description: formatError(t("campaignDetailPage.errors.delete"), error),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("campaignDetailPage.deleteConfirmation", { name: campaign.name }),
        label: t("campaignDetailPage.actions.delete"),
        variant: "danger",
      },
    );
  };

  const handleComplete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          closeCampaign({
            variables: { input: { accessReviewCampaignId: campaign.id } },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: t("campaignDetailPage.messages.error"),
                  description: formatError(t("campaignDetailPage.errors.complete"), errors),
                  variant: "error",
                });
                resolve();
                return;
              }
              toast({
                title: t("campaignDetailPage.messages.success"),
                description: t("campaignDetailPage.messages.completed"),
                variant: "success",
              });
              resolve();
            },
            onError(error) {
              toast({
                title: t("campaignDetailPage.messages.error"),
                description: formatError(t("campaignDetailPage.errors.complete"), error),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("campaignDetailPage.completeConfirmation"),
        label: t("campaignDetailPage.actions.complete"),
        variant: "primary",
      },
    );
  };

  const applyBulkFlags = (onDone?: () => void) => {
    if (!bulkFlagsDirty) {
      onDone?.();
      return;
    }

    const campaignIdAtStart = campaignIdRef.current;
    let errorCount = 0;
    let completedCount = 0;
    const total = selection.length;
    const flags = bulkFlagSelection;

    if (total === 0) {
      setBulkFlagSelection([]);
      setBulkFlagsDirty(false);
      onDone?.();
      return;
    }

    const isStale = () => campaignIdRef.current !== campaignIdAtStart;

    for (const entryId of selection) {
      bulkFlag({
        variables: {
          input: {
            accessReviewEntryId: entryId,
            flags,
          },
        },
        onCompleted(_, errors) {
          if (isStale()) {
            return;
          }
          if (errors?.length) {
            errorCount++;
          }
          completedCount++;
          if (completedCount === total) {
            if (errorCount > 0) {
              toast({
                title: t("campaignDetailPage.messages.error"),
                description: t("campaignDetailPage.errors.updateFlags", { count: errorCount }),
                variant: "error",
              });
            } else {
              toast({
                title: t("campaignDetailPage.messages.success"),
                description: t("campaignDetailPage.messages.flagsUpdated"),
                variant: "success",
              });
            }
            setBulkFlagSelection([]);
            setBulkFlagsDirty(false);
            onDone?.();
          }
        },
        onError() {
          if (isStale()) {
            return;
          }
          errorCount++;
          completedCount++;
          if (completedCount === total) {
            toast({
              title: t("campaignDetailPage.messages.error"),
              description: t("campaignDetailPage.errors.updateFlags", { count: errorCount }),
              variant: "error",
            });
            setBulkFlagSelection([]);
            setBulkFlagsDirty(false);
            onDone?.();
          }
        },
      });
    }
  };

  const applyBulkDecision = (
    decision: AccessReviewEntryDecision,
    decisionNote?: string,
    onDone?: () => void,
  ) => {
    const BATCH_SIZE = 100;
    const campaignIdAtStart = campaignIdRef.current;
    const decisions = selection.map(id => ({
      accessReviewEntryId: id,
      decision,
      decisionNote: decisionNote || null,
    }));

    if (decisions.length === 0) {
      onDone?.();
      return;
    }

    const batches: typeof decisions[] = [];
    for (let i = 0; i < decisions.length; i += BATCH_SIZE) {
      batches.push(decisions.slice(i, i + BATCH_SIZE));
    }

    let failedEntryCount = 0;
    let completedCount = 0;
    const total = batches.length;

    const isStale = () => campaignIdRef.current !== campaignIdAtStart;

    const finish = () => {
      if (failedEntryCount > 0) {
        toast({
          title: t("campaignDetailPage.messages.error"),
          description: t("campaignDetailPage.errors.recordDecisions", { count: failedEntryCount }),
          variant: "error",
        });
        // Keep selection and bulk decision state so the user can retry failures
        // (total or partial). Only clear on full success.
        return;
      }
      toast({
        title: t("campaignDetailPage.messages.success"),
        description: t("campaignDetailPage.messages.decisionsRecorded"),
        variant: "success",
      });
      setBulkDecision(null);
      setBulkPendingDecision(null);
      setBulkNote("");
      onDone?.();
    };

    for (const batch of batches) {
      bulkDecide({
        variables: {
          input: { decisions: batch },
        },
        onCompleted(_, errors) {
          if (isStale()) {
            return;
          }
          if (errors?.length) {
            failedEntryCount += batch.length;
          }
          completedCount++;
          if (completedCount === total) {
            finish();
          }
        },
        onError() {
          if (isStale()) {
            return;
          }
          failedEntryCount += batch.length;
          completedCount++;
          if (completedCount === total) {
            finish();
          }
        },
      });
    }
  };

  const handleSubmit = () => {
    if (bulkDecision !== null && bulkDecision !== "APPROVED") {
      setBulkPendingDecision(bulkDecision);
      setBulkNote("");
      bulkNoteRef.current?.open();
      return;
    }

    const finish = () => clear();

    if (bulkDecision === "APPROVED") {
      applyBulkDecision("APPROVED", undefined, () => {
        applyBulkFlags(finish);
      });
      return;
    }

    applyBulkFlags(finish);
  };

  const handleBulkFlagSelectionChange = (flags: AccessReviewEntryFlag[]) => {
    setBulkFlagSelection(flags);
    setBulkFlagsDirty(true);
  };

  const allFilteredEntryIds = useMemo(
    () => filteredSources.flatMap(({ entries }) => entries.map(edge => edge.node.id)),
    [filteredSources],
  );

  const handleSelectedChange = (entryId: string, { shiftKey }: { shiftKey: boolean }) => {
    if (shiftKey && selectionAnchorRef.current) {
      const start = allFilteredEntryIds.indexOf(selectionAnchorRef.current);
      const end = allFilteredEntryIds.indexOf(entryId);
      if (start !== -1 && end !== -1) {
        const from = Math.min(start, end);
        const to = Math.max(start, end);
        const rangeIds = allFilteredEntryIds.slice(from, to + 1);
        const next = new Set(selection);
        for (const id of rangeIds) {
          next.add(id);
        }
        reset([...next]);
        return;
      }
    }

    toggle(entryId);
    selectionAnchorRef.current = entryId;
  };

  const { page, results } = accessEntriesLayout();
  const { root: listRoot } = accessEntryList();

  return (
    <div className={page()}>
      <Breadcrumb
        items={[
          {
            label: t("campaignDetailPage.breadcrumb"),
            to: `/organizations/${organizationId}/access-reviews`,
          },
          { label: campaign.name },
        ]}
      />

      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-semibold">
          {campaign.name}
          <span className="font-normal text-txt-tertiary">
            {` (${t(`campaignDetailPage.status.${campaign.status.toLowerCase()}`)})`}
          </span>
        </h1>
        <div className="ml-auto flex items-center gap-2">
          {isPendingActions && (
            <Button onClick={handleComplete} disabled={!canComplete || isClosing}>
              {isClosing
                ? t("campaignDetailPage.actions.completing")
                : t("campaignDetailPage.actions.completeCampaign")}
            </Button>
          )}
          {canDelete && (
            <Button
              icon={IconTrashCan}
              variant="danger"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting
                ? t("campaignDetailPage.actions.deleting")
                : t("campaignDetailPage.actions.delete")}
            </Button>
          )}
        </div>
      </div>

      {isDraft && (
        <div className="flex items-center justify-end gap-2">
          <AddCampaignSourceDialog
            organizationId={organizationId}
            campaignId={campaign.id}
            existingCampaignSourceIds={existingCampaignSourceIds}
          >
            <Button icon={IconPlusLarge} variant="secondary">
              {t("campaignDetailPage.actions.addSource")}
            </Button>
          </AddCampaignSourceDialog>
          {campaign.sources.length > 0 && (
            <Button onClick={handleStart} disabled={isStarting}>
              {isStarting
                ? t("campaignDetailPage.actions.starting")
                : t("campaignDetailPage.actions.startCampaign")}
            </Button>
          )}
        </div>
      )}

      {campaign.sources.length > 0 && (
        <AccessEntriesToolbar
          emailFilter={emailFilter}
          onEmailFilterChange={setEmailFilter}
          connectorOptions={connectorOptions}
          connectorFilter={connectorFilter}
          onConnectorFilterChange={setConnectorFilter}
          mfaFilter={mfaFilter}
          onMfaFilterChange={setMfaFilter}
          adminFilter={adminFilter}
          onAdminFilterChange={setAdminFilter}
        />
      )}

      <div className={results()}>
        {filteredSources.map(({ source, entries }) => {
          const latestAttempt = source.fetchAttempts.edges[0]?.node;
          const fetchStatus = latestAttempt?.status;
          const fetchError = fetchStatus === "FAILED"
            ? latestAttempt.error
            : null;
          const isFetchInProgress = fetchStatus === "QUEUED" || fetchStatus === "FETCHING";
          const statusMessage = entries.length === 0 && isFetchInProgress && fetchStatus
            ? t(`campaignDetailPage.fetchStatus.${fetchStatus.toLowerCase()}`)
            : null;
          const hasNextPage = source.entries.pageInfo.hasNextPage;
          const truncatedCount = hasNextPage ? source.entries.edges.length : null;

          return (
            <AccessEntrySection
              key={source.id}
              title={source.name}
              count={entries.length}
              provider={source.source?.connector?.provider}
              error={fetchError}
              statusMessage={statusMessage}
              truncatedCount={truncatedCount}
            >
              {entries.length === 0
                ? (
                    statusMessage || fetchError
                      ? null
                      : (
                          <div className="rounded-[10px] border border-border-low bg-level-1 px-4 py-8 text-center text-sm text-txt-tertiary">
                            {t("campaignDetailPage.emptyEntries")}
                          </div>
                        )
                  )
                : (
                    <ul className={listRoot()}>
                      {entries.map(edge => (
                        <AccessEntryListItem
                          key={edge.node.id}
                          entryKey={edge.node}
                          isPendingActions={isPendingActions}
                          selected={selection.includes(edge.node.id)}
                          onSelectedChange={event => handleSelectedChange(edge.node.id, event)}
                        />
                      ))}
                    </ul>
                  )}
            </AccessEntrySection>
          );
        })}

        {campaign.sources.length === 0 && (
          <Card padded>
            <div className="py-8 text-center">
              <p className="text-txt-tertiary">
                {t("campaignDetailPage.emptySources")}
              </p>
            </div>
          </Card>
        )}

        {campaign.sources.length > 0 && filteredSources.length === 0 && (
          <Card padded>
            <div className="py-8 text-center">
              <p className="text-txt-tertiary">
                {t("campaignDetailPage.emptyFiltered")}
              </p>
            </div>
          </Card>
        )}
      </div>

      {isPendingActions && (
        <AccessEntriesSelectionBar
          selectedCount={selection.length}
          selectedIds={selection}
          allEntryIds={allFilteredEntryIds}
          onClear={() => {
            setBulkDecision(null);
            setBulkFlagSelection([]);
            setBulkFlagsDirty(false);
            clear();
          }}
          onSelectAll={() => reset(allFilteredEntryIds)}
          bulkDecision={bulkDecision}
          onBulkDecisionChange={setBulkDecision}
          bulkFlagSelection={bulkFlagSelection}
          onBulkFlagSelectionChange={handleBulkFlagSelectionChange}
          bulkFlagsDirty={bulkFlagsDirty}
          isSubmitting={isBulkSubmitting}
          onSubmit={handleSubmit}
        />
      )}

      <Dialog ref={bulkNoteRef} title={t("campaignDetailPage.note.title")}>
        <DialogContent padded className="space-y-4">
          <p className="text-sm text-txt-secondary">
            {t("campaignDetailPage.note.description")}
          </p>
          <Field
            label={t("campaignDetailPage.note.label")}
            type="textarea"
            value={bulkNote}
            onValueChange={setBulkNote}
          />
        </DialogContent>
        <DialogFooter>
          <Button
            disabled={!bulkNote.trim() || isBulkSubmitting}
            onClick={() => {
              if (!bulkPendingDecision) {
                return;
              }
              applyBulkDecision(bulkPendingDecision, bulkNote, () => {
                applyBulkFlags(() => {
                  bulkNoteRef.current?.close();
                  clear();
                });
              });
            }}
          >
            {t("campaignDetailPage.actions.confirm")}
          </Button>
        </DialogFooter>
      </Dialog>
    </div>
  );
}
