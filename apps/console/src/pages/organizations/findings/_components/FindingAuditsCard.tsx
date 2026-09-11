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
import {
  Button,
  Card,
  IconChevronDown,
  IconPlusLarge,
  Spinner,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useToast,
} from "@probo/ui";
import { Suspense, useEffect, useTransition } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { useParams } from "react-router";

import type { FindingAuditsCard_finding$key } from "#/__generated__/core/FindingAuditsCard_finding.graphql";
import type { FindingAuditsCardLinkMutation } from "#/__generated__/core/FindingAuditsCardLinkMutation.graphql";
import type { FindingAuditsCardPaginationQuery } from "#/__generated__/core/FindingAuditsCardPaginationQuery.graphql";
import type { FindingAuditsCardQuery } from "#/__generated__/core/FindingAuditsCardQuery.graphql";
import type { FindingAuditsCardUnlinkMutation } from "#/__generated__/core/FindingAuditsCardUnlinkMutation.graphql";
import { LinkedAuditsDialog } from "#/components/audits/LinkedAuditsDialog";

import { FindingAuditListItem } from "./FindingAuditListItem";

const findingAuditsCardQuery = graphql`
  query FindingAuditsCardQuery($findingId: ID!) {
    node(id: $findingId) {
      __typename
      ... on Finding {
        ...FindingAuditsCard_finding
      }
    }
  }
`;

const findingAuditsCardFragment = graphql`
  fragment FindingAuditsCard_finding on Finding
  @refetchable(queryName: "FindingAuditsCardPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    id
    canLinkAudit: permission(action: "core:finding:create-audit-mapping")
    canUnlinkAudit: permission(action: "core:finding:delete-audit-mapping")
    audits(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
    )
    @connection(key: "FindingAuditsCard_audits", filters: [])
    @required(action: THROW) {
      __id
      edges {
        node {
          id
        }
        ...FindingAuditListItem_auditEdge
      }
    }
  }
`;

const linkAuditMutation = graphql`
  mutation FindingAuditsCardLinkMutation($input: CreateFindingAuditMappingInput!) {
    createFindingAuditMapping(input: $input) {
      auditEdge {
        node {
          id
        }
        ...FindingAuditListItem_auditEdge
      }
    }
  }
`;

const unlinkAuditMutation = graphql`
  mutation FindingAuditsCardUnlinkMutation(
    $input: DeleteFindingAuditMappingInput!
    $connections: [ID!]!
  ) {
    deleteFindingAuditMapping(input: $input) {
      deletedAuditId @deleteEdge(connections: $connections)
    }
  }
`;

interface FindingAuditsCardContentProps {
  queryRef: PreloadedQuery<FindingAuditsCardQuery>;
}

interface FindingAuditsCardDataProps {
  findingKey: FindingAuditsCard_finding$key;
}

export function FindingAuditsCard() {
  return (
    <ErrorBoundary fallback={null}>
      <FindingAuditsCardQueryLoader />
    </ErrorBoundary>
  );
}

function FindingAuditsCardQueryLoader() {
  const { findingId } = useParams<{ findingId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<FindingAuditsCardQuery>(
    findingAuditsCardQuery,
  );

  useEffect(() => {
    if (findingId) {
      loadQuery({ findingId });
    }
  }, [findingId, loadQuery]);

  if (!queryRef || queryRef.variables.findingId !== findingId) {
    return (
      <Card padded>
        <Spinner centered />
      </Card>
    );
  }

  return (
    <Suspense
      fallback={(
        <Card padded>
          <Spinner centered />
        </Card>
      )}
    >
      <FindingAuditsCardContent queryRef={queryRef} />
    </Suspense>
  );
}

function FindingAuditsCardContent({
  queryRef,
}: FindingAuditsCardContentProps) {
  const query = usePreloadedQuery<FindingAuditsCardQuery>(
    findingAuditsCardQuery,
    queryRef,
  );
  if (query.node?.__typename !== "Finding") {
    return null;
  }

  return <FindingAuditsCardData findingKey={query.node} />;
}

function FindingAuditsCardData({
  findingKey,
}: FindingAuditsCardDataProps) {
  const {
    data: finding,
    hasNext,
    isLoadingNext,
    loadNext,
    refetch,
  } = usePaginationFragment<
    FindingAuditsCardPaginationQuery,
    FindingAuditsCard_finding$key
  >(findingAuditsCardFragment, findingKey);
  const [linkAudit, isLinking] = useMutation<FindingAuditsCardLinkMutation>(
    linkAuditMutation,
  );
  const [unlinkAudit, isUnlinking] = useMutation<FindingAuditsCardUnlinkMutation>(
    unlinkAuditMutation,
  );
  const [isRefreshing, startTransition] = useTransition();
  const { t } = useTranslation();
  const { toast } = useToast();

  const connectionId = finding.audits.__id;
  const auditEdges = finding.audits.edges;
  const linkedAudits = auditEdges.map(edge => edge.node);

  function onLink(auditId: string, referenceId?: string) {
    if (!referenceId) {
      return;
    }

    linkAudit({
      variables: {
        input: {
          findingId: finding.id,
          auditId,
          referenceId,
        },
      },
      onCompleted() {
        startTransition(() => {
          refetch({}, { fetchPolicy: "network-only" });
        });
      },
      onError(error) {
        toast({
          title: t("findingDetails.errors.title"),
          description: formatError(
            t("findingDetails.errors.linkAudit"),
            error,
          ),
          variant: "error",
        });
      },
    });
  }

  function onUnlink(auditId: string) {
    unlinkAudit({
      variables: {
        input: {
          findingId: finding.id,
          auditId,
        },
        connections: [connectionId],
      },
      onError(error) {
        toast({
          title: t("findingDetails.errors.title"),
          description: formatError(
            t("findingDetails.errors.unlinkAudit"),
            error,
          ),
          variant: "error",
        });
      },
    });
  }

  const columnCount = finding.canUnlinkAudit ? 4 : 3;
  const isMutating = isLinking || isUnlinking || isRefreshing;

  return (
    <Card padded className="space-y-[10px]">
      <div className="flex justify-between">
        <div className="text-lg font-semibold">
          {t("findingDetails.audits.title")}
        </div>
        {finding.canLinkAudit && (
          <LinkedAuditsDialog
            disabled={isMutating}
            linkedAudits={linkedAudits}
            referenceIdRequired
            onLink={onLink}
            onUnlink={finding.canUnlinkAudit ? onUnlink : undefined}
          >
            <Button
              variant="tertiary"
              icon={IconPlusLarge}
              disabled={hasNext || isLoadingNext}
            >
              {t("findingDetails.audits.link")}
            </Button>
          </LinkedAuditsDialog>
        )}
      </div>
      <Table className="bg-invert">
        <Thead>
          <Tr>
            <Th>{t("linkedAuditsCard.columns.name")}</Th>
            <Th>{t("findingDetails.audits.referenceId")}</Th>
            <Th>{t("linkedAuditsCard.columns.state")}</Th>
            {finding.canUnlinkAudit && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {auditEdges.length === 0 && (
            <Tr>
              <Td
                colSpan={columnCount}
                className="text-center text-txt-secondary"
              >
                {t("findingDetails.audits.empty")}
              </Td>
            </Tr>
          )}
          {auditEdges.map(edge => (
            <FindingAuditListItem
              key={edge.node.id}
              auditEdgeKey={edge}
              canUnlink={finding.canUnlinkAudit}
              disabled={isMutating}
              onUnlink={onUnlink}
            />
          ))}
        </Tbody>
      </Table>
      {hasNext && (
        <Button
          className="mx-auto"
          variant="tertiary"
          icon={IconChevronDown}
          disabled={isLoadingNext}
          onClick={() => loadNext(100)}
        >
          {isLoadingNext
            ? t("findingDetails.audits.loading")
            : t("findingDetails.audits.loadMore")}
        </Button>
      )}
    </Card>
  );
}
