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
import {
  useCallback,
  useDeferredValue,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useTransition,
} from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery, useRelayEnvironment } from "react-relay";
import { useNavigate } from "react-router";
import { ConnectionHandler, fetchQuery, graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { AccessReviewEntryDecision, CampaignDetailPageBulkDecisionMutation } from "#/__generated__/core/CampaignDetailPageBulkDecisionMutation.graphql";
import type { AccessReviewEntryFlag, CampaignDetailPageBulkFlagMutation } from "#/__generated__/core/CampaignDetailPageBulkFlagMutation.graphql";
import type { CampaignDetailPageCloseMutation } from "#/__generated__/core/CampaignDetailPageCloseMutation.graphql";
import type { CampaignDetailPageDeleteMutation } from "#/__generated__/core/CampaignDetailPageDeleteMutation.graphql";
import type { CampaignDetailPageQuery } from "#/__generated__/core/CampaignDetailPageQuery.graphql";
import type { CampaignDetailPageStartMutation } from "#/__generated__/core/CampaignDetailPageStartMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { isCampaignDeletableStatus } from "../_components/accessReviewHelpers";
import { AddCampaignSourceDialog } from "../dialogs/AddCampaignSourceDialog";

import { AccessEntriesSelectionBar } from "./_components/AccessEntriesSelectionBar";
import { AccessEntriesToolbar } from "./_components/AccessEntriesToolbar";
import type { EntryFilters, SourceMatches } from "./_components/AccessEntrySourceSection";
import { AccessEntrySourceSection } from "./_components/AccessEntrySourceSection";
import { accessEntriesLayout } from "./_components/variants";

const SEARCH_DEBOUNCE_MS = 300;

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
          }
          name
          ...AccessEntrySourceSection_source
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<CampaignDetailPageQuery>;
};

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
  const canDelete
    = campaign.canDelete && isCampaignDeletableStatus(campaign.status);

  const campaignIdRef = useRef(campaign.id);
  // Bumped on every campaign navigation so returning to a prior campaign cannot
  // revive in-flight bulk callbacks from an earlier visit.
  const bulkGenerationRef = useRef(0);
  const selectionAnchorRef = useRef<string | null>(null);
  const { list: selection, toggle, clear: clearSelectionList, reset } = useList<string>([]);
  const clear = () => {
    selectionAnchorRef.current = null;
    clearSelectionList();
  };
  const [emailFilter, setEmailFilter] = useState("");
  const [appliedEmailFilter, setAppliedEmailFilter] = useState("");
  const [connectorFilter, setConnectorFilter] = useState<string[]>([]);
  const [mfaFilter, setMfaFilter] = useState<string[]>([]);
  const [authMethodFilter, setAuthMethodFilter] = useState<string[]>([]);
  const [adminFilter, setAdminFilter] = useState<string[]>([]);
  const [sourceMatches, setSourceMatches] = useState<Record<string, SourceMatches>>({});
  // Polling replaces every connection with its first page, so it has to stop
  // once the reviewer has scrolled past it.
  const [hasLoadedMore, setHasLoadedMore] = useState(false);
  const [, startTransition] = useTransition();

  const debouncedApplyEmailFilter = useDebounceCallback(
    setAppliedEmailFilter,
    SEARCH_DEBOUNCE_MS,
  );
  const handleEmailFilterChange = useCallback(
    (value: string) => {
      setEmailFilter(value);
      debouncedApplyEmailFilter(value);
    },
    [debouncedApplyEmailFilter],
  );
  const handleConnectorFilterChange = useCallback(
    (value: string[]) => {
      startTransition(() => setConnectorFilter(value));
    },
    [],
  );
  const handleMfaFilterChange = useCallback(
    (value: string[]) => {
      startTransition(() => setMfaFilter(value));
    },
    [],
  );
  const handleAuthMethodFilterChange = useCallback(
    (value: string[]) => {
      startTransition(() => setAuthMethodFilter(value));
    },
    [],
  );
  const handleAdminFilterChange = useCallback(
    (value: string[]) => {
      startTransition(() => setAdminFilter(value));
    },
    [],
  );

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
    setSourceMatches({});
    setHasLoadedMore(false);
  }

  // Layout effect so async bulk callbacks see the new generation before paint / network replies.
  useLayoutEffect(() => {
    campaignIdRef.current = campaign.id;
    bulkGenerationRef.current += 1;
    selectionAnchorRef.current = null;
    // Close the note dialog so Confirm cannot submit against a cleared pending decision.
    bulkNoteRef.current?.close();
  }, [campaign.id, bulkNoteRef]);

  useEffect(() => {
    if (!isInProgress || hasLoadedMore) return;
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
  }, [isInProgress, hasLoadedMore, environment]);

  const existingCampaignSourceIds = useMemo(
    () => campaign.sources.flatMap(s => s.source?.id ? [s.source.id] : []),
    [campaign.sources],
  );

  const filters = useMemo<EntryFilters>(() => ({
    email: appliedEmailFilter.trim().toLowerCase(),
    connectorIds: connectorFilter,
    mfa: mfaFilter,
    authMethod: authMethodFilter,
    admin: adminFilter,
  }), [appliedEmailFilter, connectorFilter, mfaFilter, authMethodFilter, adminFilter]);
  const deferredFilters = useDeferredValue(filters);

  const hasActiveFilters = deferredFilters.email !== ""
    || deferredFilters.connectorIds.length > 0
    || deferredFilters.mfa.length > 0
    || deferredFilters.authMethod.length > 0
    || deferredFilters.admin.length > 0;

  // Each source section paginates on its own and reports the entries it shows,
  // so the page can still offer select-all and shift-range across connectors.
  const handleMatchesChange = useCallback(
    (sourceId: string, matches: SourceMatches) => {
      setSourceMatches((previous) => {
        const current = previous[sourceId];
        if (
          current
          && current.hasNext === matches.hasNext
          && current.entryIds.length === matches.entryIds.length
          && current.entryIds.every((id, index) => id === matches.entryIds[index])
        ) {
          return previous;
        }
        return { ...previous, [sourceId]: matches };
      });
    },
    [],
  );

  const handleLoadNext = useCallback(() => setHasLoadedMore(true), []);

  const allFilteredEntryIds = useMemo(
    () => campaign.sources.flatMap(
      source => sourceMatches[source.id]?.entryIds ?? [],
    ),
    [campaign.sources, sourceMatches],
  );
  // A source that has not reported yet counts as unfinished, so the empty state
  // cannot flash before the sections have applied the filters.
  const mayHaveMoreMatches = campaign.sources.some(
    source => sourceMatches[source.id]?.hasNext ?? true,
  );

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

  type BulkFlagsResult = {
    succeededIds: string[];
    failedIds: string[];
  };

  const applyBulkFlags = (
    entryIds?: ReadonlyArray<string>,
    onDone?: (result: BulkFlagsResult) => void,
  ) => {
    if (!bulkFlagsDirty) {
      onDone?.({ succeededIds: [], failedIds: [] });
      return;
    }

    const generationAtStart = bulkGenerationRef.current;
    const ids = entryIds ?? selection;
    const succeededIds: string[] = [];
    const failedIds: string[] = [];
    let completedCount = 0;
    const total = ids.length;
    const flags = bulkFlagSelection;

    if (total === 0) {
      setBulkFlagSelection([]);
      setBulkFlagsDirty(false);
      onDone?.({ succeededIds: [], failedIds: [] });
      return;
    }

    const isStale = () => bulkGenerationRef.current !== generationAtStart;

    const finish = () => {
      if (failedIds.length > 0) {
        toast({
          title: t("campaignDetailPage.messages.error"),
          description: t("campaignDetailPage.errors.updateFlags", { count: failedIds.length }),
          variant: "error",
        });
        // Keep flag selection/dirty so the user can retry failed entries.
        onDone?.({ succeededIds, failedIds });
        return;
      }

      toast({
        title: t("campaignDetailPage.messages.success"),
        description: t("campaignDetailPage.messages.flagsUpdated"),
        variant: "success",
      });
      setBulkFlagSelection([]);
      setBulkFlagsDirty(false);
      onDone?.({ succeededIds, failedIds });
    };

    for (const entryId of ids) {
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
            failedIds.push(entryId);
          } else {
            succeededIds.push(entryId);
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
          failedIds.push(entryId);
          completedCount++;
          if (completedCount === total) {
            finish();
          }
        },
      });
    }
  };

  type BulkDecisionResult = {
    succeededIds: string[];
    failedIds: string[];
  };

  const applyBulkDecision = (
    decision: AccessReviewEntryDecision,
    decisionNote?: string,
    onDone?: (result: BulkDecisionResult) => void,
  ) => {
    const BATCH_SIZE = 100;
    const generationAtStart = bulkGenerationRef.current;
    const decisions = selection.map(id => ({
      accessReviewEntryId: id,
      decision,
      decisionNote: decisionNote || null,
    }));

    if (decisions.length === 0) {
      onDone?.({ succeededIds: [], failedIds: [] });
      return;
    }

    const batches: typeof decisions[] = [];
    for (let i = 0; i < decisions.length; i += BATCH_SIZE) {
      batches.push(decisions.slice(i, i + BATCH_SIZE));
    }

    const succeededIds: string[] = [];
    const failedIds: string[] = [];
    let completedCount = 0;
    const total = batches.length;

    const isStale = () => bulkGenerationRef.current !== generationAtStart;

    const finish = () => {
      if (failedIds.length === decisions.length) {
        toast({
          title: t("campaignDetailPage.messages.error"),
          description: t("campaignDetailPage.errors.recordDecisions", { count: failedIds.length }),
          variant: "error",
        });
        // Full failure — keep selection and bulk state for retry; do not chain flags.
        return;
      }

      if (failedIds.length > 0) {
        toast({
          title: t("campaignDetailPage.messages.error"),
          description: t("campaignDetailPage.errors.recordDecisions", { count: failedIds.length }),
          variant: "error",
        });
        // Partial failure — keep bulk decision state for retrying failed entries;
        // caller applies flags to successes and narrows selection to failures.
        onDone?.({ succeededIds, failedIds });
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
      onDone?.({ succeededIds, failedIds });
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
          const batchIds = batch.map(d => d.accessReviewEntryId);
          if (errors?.length) {
            failedIds.push(...batchIds);
          } else {
            succeededIds.push(...batchIds);
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
          failedIds.push(...batch.map(d => d.accessReviewEntryId));
          completedCount++;
          if (completedCount === total) {
            finish();
          }
        },
      });
    }
  };

  const continueAfterBulkDecision = (
    result: BulkDecisionResult,
    onFullyDone: () => void,
  ) => {
    if (result.failedIds.length > 0) {
      const decisionFailedIds = result.failedIds;
      // Keep decision failures selected for retry; flag the successes.
      reset(decisionFailedIds);
      applyBulkFlags(result.succeededIds, (flagResult) => {
        if (flagResult.failedIds.length > 0) {
          // Also keep entries whose flags failed so both recovery paths remain.
          reset([...decisionFailedIds, ...flagResult.failedIds]);
        }
      });
      return;
    }
    applyBulkFlags(result.succeededIds, (flagResult) => {
      if (flagResult.failedIds.length > 0) {
        reset(flagResult.failedIds);
        return;
      }
      onFullyDone();
    });
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
      applyBulkDecision("APPROVED", undefined, (result) => {
        continueAfterBulkDecision(result, finish);
      });
      return;
    }

    applyBulkFlags(undefined, (flagResult) => {
      if (flagResult.failedIds.length > 0) {
        reset(flagResult.failedIds);
        return;
      }
      finish();
    });
  };

  const handleBulkFlagSelectionChange = (flags: AccessReviewEntryFlag[]) => {
    setBulkFlagSelection(flags);
    setBulkFlagsDirty(true);
  };

  const selectedIdSet = useMemo(() => new Set(selection), [selection]);

  const handleSelectedChange = useCallback(
    (entryId: string, { shiftKey }: { shiftKey: boolean }) => {
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
    },
    [allFilteredEntryIds, reset, selection, toggle],
  );

  const { page, results } = accessEntriesLayout();

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
          onEmailFilterChange={handleEmailFilterChange}
          connectorOptions={connectorOptions}
          connectorFilter={connectorFilter}
          onConnectorFilterChange={handleConnectorFilterChange}
          mfaFilter={mfaFilter}
          onMfaFilterChange={handleMfaFilterChange}
          authMethodFilter={authMethodFilter}
          onAuthMethodFilterChange={handleAuthMethodFilterChange}
          adminFilter={adminFilter}
          onAdminFilterChange={handleAdminFilterChange}
        />
      )}

      <div className={results()}>
        {campaign.sources.map(source => (
          <AccessEntrySourceSection
            key={source.id}
            sourceKey={source}
            filters={deferredFilters}
            isPendingActions={isPendingActions}
            selectedIds={selectedIdSet}
            onSelectedChange={handleSelectedChange}
            onMatchesChange={handleMatchesChange}
            onLoadNext={handleLoadNext}
          />
        ))}

        {campaign.sources.length === 0 && (
          <Card padded>
            <div className="py-8 text-center">
              <p className="text-txt-tertiary">
                {t("campaignDetailPage.emptySources")}
              </p>
            </div>
          </Card>
        )}

        {campaign.sources.length > 0
          && hasActiveFilters
          && allFilteredEntryIds.length === 0
          && !mayHaveMoreMatches && (
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
              applyBulkDecision(bulkPendingDecision, bulkNote, (result) => {
                continueAfterBulkDecision(result, () => {
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
