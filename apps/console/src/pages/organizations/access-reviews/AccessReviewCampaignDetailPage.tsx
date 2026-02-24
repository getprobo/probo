import { formatDate, formatError, type GraphQLError, promisifyMutation } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Breadcrumb,
  Button,
  Card,
  PageHeader,
  Table,
  Tabs,
  TabItem,
  Tbody,
  Th,
  Thead,
  Tr,
  useToast,
} from "@probo/ui";
import { useEffect, useRef, useState } from "react";
import {
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { AccessReviewCampaignDetailPageCancelMutation } from "#/__generated__/core/AccessReviewCampaignDetailPageCancelMutation.graphql";
import type { AccessReviewCampaignDetailPageCloseMutation } from "#/__generated__/core/AccessReviewCampaignDetailPageCloseMutation.graphql";
import type { AccessReviewCampaignDetailPageEntriesFragment$key } from "#/__generated__/core/AccessReviewCampaignDetailPageEntriesFragment.graphql";
import type { AccessReviewCampaignDetailPageEntriesPaginationQuery } from "#/__generated__/core/AccessReviewCampaignDetailPageEntriesPaginationQuery.graphql";
import type { AccessReviewCampaignDetailPageExportMutation } from "#/__generated__/core/AccessReviewCampaignDetailPageExportMutation.graphql";
import type { AccessReviewCampaignDetailPageQuery } from "#/__generated__/core/AccessReviewCampaignDetailPageQuery.graphql";
import type { AccessReviewCampaignDetailPageRetryStartMutation } from "#/__generated__/core/AccessReviewCampaignDetailPageRetryStartMutation.graphql";
import type { AccessReviewCampaignDetailPageStartMutation } from "#/__generated__/core/AccessReviewCampaignDetailPageStartMutation.graphql";
import type { AccessReviewCampaignDetailPageValidateMutation } from "#/__generated__/core/AccessReviewCampaignDetailPageValidateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AccessEntryRow } from "./_components/AccessEntryRow";

export const accessReviewCampaignDetailPageQuery = graphql`
  query AccessReviewCampaignDetailPageQuery($campaignId: ID!) {
    node(id: $campaignId) {
      ... on AccessReviewCampaign {
        id
        name
        status
        startedAt
        completedAt
        frameworkControls
        createdAt
        updatedAt
        canStart: permission(action: "core:access-review-campaign:start")
        canClose: permission(action: "core:access-review-campaign:close")
        canCancel: permission(action: "core:access-review-campaign:cancel")
        scopeSources {
          id
          name
          fetchStatus
          fetchedAccountsCount
          attemptCount
          lastError
          fetchStartedAt
          fetchCompletedAt
        }
        ...AccessReviewCampaignDetailPageEntriesFragment
      }
    }
  }
`;

const entriesPaginatedFragment = graphql`
  fragment AccessReviewCampaignDetailPageEntriesFragment on AccessReviewCampaign
  @refetchable(queryName: "AccessReviewCampaignDetailPageEntriesPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "AccessEntryOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    accessSourceId: { type: "ID", defaultValue: null }
  ) {
    entries(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      accessSourceId: $accessSourceId
    ) @connection(key: "AccessReviewCampaignDetailPage_entries") {
      __id
      edges {
        node {
          id
          ...AccessEntryRowFragment
        }
      }
    }
  }
`;

const startMutation = graphql`
  mutation AccessReviewCampaignDetailPageStartMutation(
    $input: StartAccessReviewCampaignInput!
  ) {
    startAccessReviewCampaign(input: $input) {
      accessReviewCampaign {
        id
        status
        startedAt
        scopeSources {
          id
          name
          fetchStatus
          fetchedAccountsCount
          attemptCount
          lastError
          fetchStartedAt
          fetchCompletedAt
        }
      }
    }
  }
`;

const closeMutation = graphql`
  mutation AccessReviewCampaignDetailPageCloseMutation(
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

const retryStartMutation = graphql`
  mutation AccessReviewCampaignDetailPageRetryStartMutation(
    $input: RetryStartAccessReviewCampaignInput!
  ) {
    retryStartAccessReviewCampaign(input: $input) {
      accessReviewCampaign {
        id
        status
        startedAt
        scopeSources {
          id
          name
          fetchStatus
          fetchedAccountsCount
          attemptCount
          lastError
          fetchStartedAt
          fetchCompletedAt
        }
      }
    }
  }
`;

const validateMutation = graphql`
  mutation AccessReviewCampaignDetailPageValidateMutation(
    $input: ValidateAccessReviewCampaignInput!
  ) {
    validateAccessReviewCampaign(input: $input) {
      accessReviewCampaign {
        id
        status
      }
    }
  }
`;

const exportMutation = graphql`
  mutation AccessReviewCampaignDetailPageExportMutation(
    $input: ExportCampaignEvidenceInput!
  ) {
    exportCampaignEvidence(input: $input) {
      checksumSha256
      payload
    }
  }
`;

const cancelMutation = graphql`
  mutation AccessReviewCampaignDetailPageCancelMutation(
    $input: CancelAccessReviewCampaignInput!
  ) {
    cancelAccessReviewCampaign(input: $input) {
      accessReviewCampaign {
        id
        status
      }
    }
  }
`;

function statusBadgeVariant(status: string) {
  switch (status) {
    case "DRAFT":
      return "neutral" as const;
    case "IN_PROGRESS":
      return "info" as const;
    case "PENDING_ACTIONS":
      return "warning" as const;
    case "FAILED":
      return "danger" as const;
    case "COMPLETED":
      return "success" as const;
    case "CANCELLED":
      return "danger" as const;
    default:
      return "neutral" as const;
  }
}

function sourceFetchBadgeVariant(status: string) {
  switch (status) {
    case "QUEUED":
      return "neutral" as const;
    case "FETCHING":
      return "info" as const;
    case "SUCCESS":
      return "success" as const;
    case "FAILED":
      return "danger" as const;
    default:
      return "neutral" as const;
  }
}

type Props = {
  queryRef: PreloadedQuery<AccessReviewCampaignDetailPageQuery>;
};

export default function AccessReviewCampaignDetailPage(props: Props) {
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(accessReviewCampaignDetailPageQuery, props.queryRef);
  const campaign = data.node;
  const { __ } = useTranslate();
  const { toast } = useToast();

  if (!campaign?.id) {
    throw new Error("Cannot load access review campaign detail page");
  }

  usePageTitle(campaign.name || __("Campaign"));

  const scopeSources = campaign.scopeSources ?? [];
  const [selectedSourceId, setSelectedSourceId] = useState<string | null>(
    scopeSources.length > 0 ? scopeSources[0].id : null
  );

  const {
    data: { entries },
    loadNext,
    hasNext,
    isLoadingNext,
    refetch,
  } = usePaginationFragment<
    AccessReviewCampaignDetailPageEntriesPaginationQuery,
    AccessReviewCampaignDetailPageEntriesFragment$key
  >(entriesPaginatedFragment, campaign);

  const [startCampaign, isStarting] = useMutation<AccessReviewCampaignDetailPageStartMutation>(startMutation);
  const [retryStartCampaign, isRetrying] = useMutation<AccessReviewCampaignDetailPageRetryStartMutation>(retryStartMutation);
  const [validateCampaign, isValidating] = useMutation<AccessReviewCampaignDetailPageValidateMutation>(validateMutation);
  const [exportEvidence, isExporting] = useMutation<AccessReviewCampaignDetailPageExportMutation>(exportMutation);
  const [closeCampaign, isClosing] = useMutation<AccessReviewCampaignDetailPageCloseMutation>(closeMutation);
  const [cancelCampaign, isCancelling] = useMutation<AccessReviewCampaignDetailPageCancelMutation>(cancelMutation);
  const [completedSourceIds, setCompletedSourceIds] = useState<Set<string>>(new Set());

  // On initial mount, if there are scope sources, refetch entries
  // filtered by the first source (initial query loads unfiltered).
  const initialRefetchDone = useRef(false);
  useEffect(() => {
    if (!initialRefetchDone.current && selectedSourceId) {
      initialRefetchDone.current = true;
      refetch({ accessSourceId: selectedSourceId });
    }
  }, [selectedSourceId, refetch]);

  useEffect(() => {
    if (!scopeSources.some(source => source.fetchStatus === "QUEUED" || source.fetchStatus === "FETCHING")) {
      return undefined;
    }

    const intervalId = window.setInterval(() => {
      refetch({ accessSourceId: selectedSourceId });
    }, 5000);

    return () => window.clearInterval(intervalId);
  }, [scopeSources, refetch, selectedSourceId]);

  const handleSourceChange = (sourceId: string) => {
    setSelectedSourceId(sourceId);
    refetch({ accessSourceId: sourceId });
  };

  const currentSourceIndex = scopeSources.findIndex((source) => source.id === selectedSourceId);
  const hasPrevSource = currentSourceIndex > 0;
  const hasNextSource = currentSourceIndex >= 0 && currentSourceIndex < scopeSources.length - 1;
  const hasFetchingSources = scopeSources.some(
    source => source.fetchStatus === "QUEUED" || source.fetchStatus === "FETCHING"
  );

  const goToSource = (sourceId: string) => {
    setSelectedSourceId(sourceId);
    refetch({ accessSourceId: sourceId });
  };

  const markCurrentSourceReviewed = () => {
    if (!selectedSourceId) {
      return;
    }
    setCompletedSourceIds((prev) => {
      const next = new Set(prev);
      next.add(selectedSourceId);
      return next;
    });
  };

  const handleStart = () => {
    promisifyMutation(startCampaign)({
      variables: { input: { accessReviewCampaignId: campaign.id! } },
    })
      .then((response) => {
        const sources = response.startAccessReviewCampaign?.accessReviewCampaign?.scopeSources;
        if (sources && sources.length > 0) {
          const firstSourceId = sources[0].id;
          setSelectedSourceId(firstSourceId);
          refetch({ accessSourceId: firstSourceId });
        }
      })
      .catch((error) => {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to start campaign"), error as GraphQLError),
          variant: "error",
        });
      });
  };

  const handleClose = () => {
    promisifyMutation(closeCampaign)({
      variables: { input: { accessReviewCampaignId: campaign.id! } },
    }).catch((error) => {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to close campaign"), error as GraphQLError),
        variant: "error",
      });
    });
  };

  const handleRetryStart = () => {
    promisifyMutation(retryStartCampaign)({
      variables: { input: { accessReviewCampaignId: campaign.id! } },
    })
      .then((response) => {
        const sources = response.retryStartAccessReviewCampaign?.accessReviewCampaign?.scopeSources;
        if (sources && sources.length > 0) {
          goToSource(sources[0].id);
        }
      })
      .catch((error) => {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to retry campaign start"), error as GraphQLError),
          variant: "error",
        });
      });
  };

  const handleValidate = () => {
    const note = window.prompt(__("Add validation note (optional)")) ?? "";
    promisifyMutation(validateCampaign)({
      variables: { input: { accessReviewCampaignId: campaign.id!, note } },
    }).catch((error) => {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to validate campaign"), error as GraphQLError),
        variant: "error",
      });
    });
  };

  const handleExportEvidence = () => {
    promisifyMutation(exportEvidence)({
      variables: { input: { accessReviewCampaignId: campaign.id! } },
    }).then((response) => {
      const checksum = response.exportCampaignEvidence?.checksumSha256;
      if (checksum) {
        toast({
          title: __("Evidence exported"),
          description: `${__("Checksum")}: ${checksum}`,
          variant: "success",
        });
      }
    }).catch((error) => {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to export campaign evidence"), error as GraphQLError),
        variant: "error",
      });
    });
  };

  const handleCancel = () => {
    promisifyMutation(cancelCampaign)({
      variables: { input: { accessReviewCampaignId: campaign.id! } },
    }).catch((error) => {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to cancel campaign"), error as GraphQLError),
        variant: "error",
      });
    });
  };

  const listUrl = `/organizations/${organizationId}/access-reviews`;

  const isDraft = campaign.status === "DRAFT";
  const isFailed = campaign.status === "FAILED";
  const isInProgress = campaign.status === "IN_PROGRESS" || campaign.status === "PENDING_ACTIONS";
  const isTerminal = campaign.status === "COMPLETED" || campaign.status === "CANCELLED";
  const allSourcesReviewed = scopeSources.length > 0 && scopeSources.every((source) => completedSourceIds.has(source.id));

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Access Reviews"),
            to: listUrl,
          },
          {
            label: campaign.name || __("Campaign detail"),
          },
        ]}
      />

      <PageHeader
        title={(
          <div className="flex items-center gap-3">
            <span>{campaign.name}</span>
            <Badge variant={statusBadgeVariant(campaign.status!)}>
              {campaign.status}
            </Badge>
          </div>
        )}
      >
        {isDraft && campaign.canStart && (
          <Button
            onClick={handleStart}
            disabled={isStarting}
          >
            {__("Start Campaign")}
          </Button>
        )}
        {isFailed && campaign.canStart && (
          <Button
            onClick={handleRetryStart}
            disabled={isRetrying}
          >
            {__("Retry Start")}
          </Button>
        )}
        {isInProgress && (
          <Button
            variant="secondary"
            onClick={handleExportEvidence}
            disabled={isExporting}
          >
            {__("Export Evidence")}
          </Button>
        )}
        {isInProgress && (
          <Button
            variant="secondary"
            onClick={handleValidate}
            disabled={isValidating || !allSourcesReviewed}
          >
            {__("Validate Campaign")}
          </Button>
        )}
        {isInProgress && campaign.canClose && (
          <Button
            onClick={handleClose}
            disabled={isClosing}
          >
            {__("Close Campaign")}
          </Button>
        )}
        {!isTerminal && campaign.canCancel && (
          <Button
            variant="secondary"
            onClick={handleCancel}
            disabled={isCancelling}
          >
            {__("Cancel Campaign")}
          </Button>
        )}
      </PageHeader>

      <div className="space-y-6">
        <div className="space-y-4">
          <h2 className="text-base font-medium">{__("Details")}</h2>
          <Card className="space-y-4" padded>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Started at")}
                </div>
                <div className="text-sm text-txt-primary">
                  {campaign.startedAt ? formatDate(campaign.startedAt) : "-"}
                </div>
              </div>
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Completed at")}
                </div>
                <div className="text-sm text-txt-primary">
                  {campaign.completedAt ? formatDate(campaign.completedAt) : "-"}
                </div>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Created at")}
                </div>
                <div className="text-sm text-txt-primary">
                  {formatDate(campaign.createdAt)}
                </div>
              </div>
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Updated at")}
                </div>
                <div className="text-sm text-txt-primary">
                  {formatDate(campaign.updatedAt)}
                </div>
              </div>
            </div>
            {campaign.frameworkControls && campaign.frameworkControls.length > 0 && (
              <div>
                <div className="text-xs text-txt-tertiary font-semibold mb-1">
                  {__("Framework Controls")}
                </div>
                <div className="flex flex-wrap gap-1">
                  {campaign.frameworkControls.map((control) => (
                    <Badge key={control} variant="neutral" size="sm">
                      {control}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </Card>
        </div>

        <div className="space-y-4">
          <h2 className="text-base font-medium">{__("Access Entries")}</h2>
          {scopeSources.length > 0 && (
            <Card padded className="space-y-3">
              <div className="text-sm text-txt-secondary">
                {hasFetchingSources
                  ? __("Fetching source access in background. Progress updates automatically.")
                  : __("Review each source one-by-one before validating.")}
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="secondary"
                  disabled={!hasPrevSource}
                  onClick={() => {
                    if (currentSourceIndex > 0) {
                      goToSource(scopeSources[currentSourceIndex - 1].id);
                    }
                  }}
                >
                  {__("Previous Source")}
                </Button>
                <Button
                  variant="secondary"
                  disabled={!hasNextSource}
                  onClick={() => {
                    if (currentSourceIndex >= 0 && currentSourceIndex < scopeSources.length - 1) {
                      goToSource(scopeSources[currentSourceIndex + 1].id);
                    }
                  }}
                >
                  {__("Next Source")}
                </Button>
                <Button
                  onClick={markCurrentSourceReviewed}
                  disabled={!selectedSourceId}
                >
                  {__("Mark Source Reviewed")}
                </Button>
              </div>
              <div className="text-sm">
                {`${__("Reviewed sources")}: ${completedSourceIds.size} / ${scopeSources.length}`}
              </div>
            </Card>
          )}

          {scopeSources.length > 0
            ? (
                <div className="space-y-4">
                  <Tabs>
                    {scopeSources.map((source) => (
                      <TabItem
                        key={source.id}
                        active={selectedSourceId === source.id}
                        onClick={() => handleSourceChange(source.id)}
                      >
                        <span className="inline-flex items-center gap-2">
                          <span>{source.name}</span>
                          <Badge variant={sourceFetchBadgeVariant(source.fetchStatus)} size="sm">
                            {source.fetchStatus}
                          </Badge>
                        </span>
                      </TabItem>
                    ))}
                  </Tabs>

                  {selectedSourceId && (
                    <Card padded>
                      {scopeSources
                        .filter(source => source.id === selectedSourceId)
                        .map(source => (
                          <div key={source.id} className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                                {__("Fetch status")}
                              </div>
                              <div>{source.fetchStatus}</div>
                            </div>
                            <div>
                              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                                {__("Fetched accounts")}
                              </div>
                              <div>{source.fetchedAccountsCount}</div>
                            </div>
                            <div>
                              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                                {__("Attempt count")}
                              </div>
                              <div>{source.attemptCount}</div>
                            </div>
                            <div>
                              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                                {__("Fetch started at")}
                              </div>
                              <div>{source.fetchStartedAt ? formatDate(source.fetchStartedAt) : "-"}</div>
                            </div>
                            <div>
                              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                                {__("Fetch completed at")}
                              </div>
                              <div>{source.fetchCompletedAt ? formatDate(source.fetchCompletedAt) : "-"}</div>
                            </div>
                            <div>
                              <div className="text-xs text-txt-tertiary font-semibold mb-1">
                                {__("Last fetch error")}
                              </div>
                              <div>{source.lastError || "-"}</div>
                            </div>
                          </div>
                        ))}
                    </Card>
                  )}

                  {entries && entries.edges.length > 0
                    ? (
                        <Card>
                          <Table>
                            <Thead>
                              <Tr>
                                <Th>{__("Name")}</Th>
                                <Th>{__("Email")}</Th>
                                <Th>{__("Role")}</Th>
                                <Th>{__("Incremental")}</Th>
                                <Th>{__("Flag")}</Th>
                                <Th>{__("Decision")}</Th>
                                <Th>{__("Decision note")}</Th>
                                <Th className="w-12"></Th>
                              </Tr>
                            </Thead>
                            <Tbody>
                              {entries.edges.map(edge => (
                                <AccessEntryRow
                                  key={edge.node.id}
                                  fKey={edge.node}
                                />
                              ))}
                            </Tbody>
                          </Table>

                          {hasNext && (
                            <div className="p-4 border-t">
                              <Button
                                variant="secondary"
                                onClick={() => loadNext(50)}
                                disabled={isLoadingNext}
                              >
                                {isLoadingNext
                                  ? __("Loading...")
                                  : __("Load more")}
                              </Button>
                            </div>
                          )}
                        </Card>
                      )
                    : (
                        <Card padded>
                          <div className="text-center py-8">
                            <p className="text-txt-tertiary">
                              {__("No access entries found for this source.")}
                            </p>
                          </div>
                        </Card>
                      )}
                </div>
              )
            : (
                <Card padded>
                  <div className="text-center py-8">
                    <p className="text-txt-tertiary">
                      {isDraft
                        ? __("Start the campaign to snapshot access entries from your sources.")
                        : __("No access entries found for this campaign.")}
                    </p>
                  </div>
                </Card>
              )}
        </div>
      </div>
    </div>
  );
}
