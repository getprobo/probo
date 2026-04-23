// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import {
  formatDate,
  getObligationStatusLabel,
  getObligationStatusVariant,
  promisifyMutation,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  DropdownItem,
  IconPageTextLine,
  IconPlusLarge,
  IconTrashCan,
  IconUpload,
  PageHeader,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import {
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";

import type { ObligationGraphDeleteMutation } from "#/__generated__/core/ObligationGraphDeleteMutation.graphql";
import type { ObligationGraphListQuery } from "#/__generated__/core/ObligationGraphListQuery.graphql";
import type {
  ObligationsPageFragment$data,
  ObligationsPageFragment$key,
} from "#/__generated__/core/ObligationsPageFragment.graphql";
import type { ObligationsPageRefetchQuery } from "#/__generated__/core/ObligationsPageRefetchQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  deleteObligationMutation,
  obligationsQuery,
} from "../../../hooks/graph/ObligationGraph";

import { CreateObligationDialog } from "./dialogs/CreateObligationDialog";
import { PublishObligationListDialog } from "./dialogs/PublishObligationListDialog";

type Obligation
  = ObligationsPageFragment$data["obligations"]["edges"][number]["node"];

interface ObligationsPageProps {
  queryRef: PreloadedQuery<ObligationGraphListQuery>;
}

const obligationsPageFragment = graphql`
  fragment ObligationsPageFragment on Organization
  @refetchable(queryName: "ObligationsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey" }
  ) {
    id
    obligations(
      first: $first
      after: $after
    ) @connection(key: "ObligationsPage_obligations") {
      __id
      edges {
        node {
          id
          area
          source
          status
          dueDate
          owner {
            id
            fullName
          }
          canUpdate: permission(action: "core:obligation:update")
          canDelete: permission(action: "core:obligation:delete")
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

export default function ObligationsPage({ queryRef }: ObligationsPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  usePageTitle(__("Obligations"));

  const organization = usePreloadedQuery(obligationsQuery, queryRef);
  const defaultApproverIds = (organization.node.obligationsDocument?.defaultApprovers ?? []).map(a => a.id);

  const {
    data: obligationsData,
    loadNext,
    hasNext,
  } = usePaginationFragment<ObligationsPageRefetchQuery, ObligationsPageFragment$key>(
    obligationsPageFragment,
    organization.node as ObligationsPageFragment$key,
  );

  const connectionId = obligationsData?.obligations?.__id || "";
  const obligations: Obligation[]
    = obligationsData?.obligations?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction
    = obligations.some(({ canUpdate, canDelete }) => canDelete || canUpdate);

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Obligations")}
        description={__("Manage your organization's obligations.")}
      >
        <div className="flex gap-2">
          {organization.node.obligationsDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${organization.node.obligationsDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {__("Document")}
              </Link>
            </Button>
          )}
          {organization.node.canPublishObligations && (
            <PublishObligationListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={(documentId) => {
                void navigate(
                  `/organizations/${organizationId}/documents/${documentId}`,
                );
              }}
            >
              <Button variant="secondary" icon={IconUpload}>
                {__("Publish")}
              </Button>
            </PublishObligationListDialog>
          )}
          {organization.node.canCreateObligation && (
            <CreateObligationDialog
              organizationId={organizationId}
              connection={connectionId}
            >
              <Button icon={IconPlusLarge}>{__("Add obligation")}</Button>
            </CreateObligationDialog>
          )}
        </div>
      </PageHeader>

      {obligations.length === 0
        ? (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {__("No obligations yet")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__("Create your first obligation to get started.")}
                </p>
              </div>
            </Card>
          )
        : (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Area")}</Th>
                    <Th>{__("Source")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Owner")}</Th>
                    <Th>{__("Due Date")}</Th>
                    {hasAnyAction && <Th>{__("Actions")}</Th>}
                  </Tr>
                </Thead>
                <Tbody>
                  {obligations.map(obligation => (
                    <ObligationRow
                      key={obligation.id}
                      obligation={obligation}
                      connectionId={connectionId}
                      hasAnyAction={hasAnyAction}
                    />
                  ))}
                </Tbody>
              </Table>

              {hasNext && (
                <div className="p-4 border-t">
                  <Button
                    variant="secondary"
                    onClick={() => loadNext(10)}
                    disabled={!hasNext}
                  >
                    {__("Load more")}
                  </Button>
                </div>
              )}
            </Card>
          )}
    </div>
  );
}

function ObligationRow({
  obligation,
  connectionId,
  hasAnyAction,
}: {
  obligation: Obligation;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const [deleteObligation] = useMutation<ObligationGraphDeleteMutation>(deleteObligationMutation);
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteObligation)({
          variables: {
            input: {
              obligationId: obligation.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: __(
          "This will permanently delete this obligation. This action cannot be undone.",
        ),
      },
    );
  };

  const detailsUrl = `/organizations/${organizationId}/obligations/${obligation.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>{obligation.area || "-"}</Td>
      <Td>{obligation.source || "-"}</Td>
      <Td>
        <Badge
          variant={getObligationStatusVariant(
            obligation.status || "NON_COMPLIANT",
          )}
        >
          {getObligationStatusLabel(obligation.status || "NON_COMPLIANT")}
        </Badge>
      </Td>
      <Td>{obligation.owner?.fullName || "-"}</Td>
      <Td>
        {obligation.dueDate
          ? (
              <time dateTime={obligation.dueDate}>
                {formatDate(obligation.dueDate)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">{__("No due date")}</span>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {obligation.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {__("Delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
