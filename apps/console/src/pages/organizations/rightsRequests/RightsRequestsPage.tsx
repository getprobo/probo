// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import {
  getRightsRequestStateVariant,
  promisifyMutation,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  Card,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { RightsRequestGraphDeleteMutation } from "#/__generated__/core/RightsRequestGraphDeleteMutation.graphql";
import type { RightsRequestGraphListQuery } from "#/__generated__/core/RightsRequestGraphListQuery.graphql";
import type {
  RightsRequestsPageFragment$data,
  RightsRequestsPageFragment$key,
} from "#/__generated__/core/RightsRequestsPageFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import {
  deleteRightsRequestMutation,
  RightsRequestsConnectionKey,
  rightsRequestsQuery,
} from "../../../hooks/graph/RightsRequestGraph";

import { CreateRightsRequestDialog } from "./dialogs/CreateRightsRequestDialog";

interface RightsRequestsPageProps {
  queryRef: PreloadedQuery<RightsRequestGraphListQuery>;
}

const rightsRequestsPageFragment = graphql`
    fragment RightsRequestsPageFragment on Organization
    @refetchable(queryName: "RightsRequestsPageRefetchQuery")
    @argumentDefinitions(
        first: { type: "Int", defaultValue: 10 }
        after: { type: "CursorKey" }
    ) {
        id
        rightsRequests(first: $first, after: $after)
            @connection(key: "RightsRequestsPage_rightsRequests") {
            edges {
                node {
                    id
                    requestType
                    requestState
                    dataSubject
                    contact
                    deadline

                    canDelete: permission(action: "core:rights-request:delete")
                    canUpdate: permission(action: "core:rights-request:update")
                }
            }
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }
`;

export default function RightsRequestsPage({
  queryRef,
}: RightsRequestsPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();

  usePageTitle(t("rightsRequestsPage.title"));

  const organization = usePreloadedQuery<RightsRequestGraphListQuery>(
    rightsRequestsQuery,
    queryRef,
  );

  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    RightsRequestGraphListQuery,
    RightsRequestsPageFragment$key
  >(rightsRequestsPageFragment, organization.node);

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    RightsRequestsConnectionKey,
  );
  const requests
    = data?.rightsRequests?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction = requests.some(
    ({ canUpdate, canDelete }) => canUpdate || canDelete,
  );

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("rightsRequestsPage.title")}
        description={t("rightsRequestsPage.description")}
      >
        {organization.node.canCreateRightsRequest && (
          <CreateRightsRequestDialog
            organizationId={organizationId}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>
              {t("rightsRequestsPage.actions.add")}
            </Button>
          </CreateRightsRequestDialog>
        )}
      </PageHeader>

      {requests.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("rightsRequestsPage.columns.type")}</Th>
                    <Th>{t("rightsRequestsPage.columns.state")}</Th>
                    <Th>{t("rightsRequestsPage.columns.dataSubject")}</Th>
                    <Th>{t("rightsRequestsPage.columns.contact")}</Th>
                    <Th>{t("rightsRequestsPage.columns.deadline")}</Th>
                    {hasAnyAction && <Th>{t("rightsRequestsPage.columns.actions")}</Th>}
                  </Tr>
                </Thead>
                <Tbody>
                  {requests.map(request => (
                    <RequestRow
                      key={request.id}
                      request={request}
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
                    disabled={isLoadingNext}
                  >
                    {isLoadingNext
                      ? t("rightsRequestsPage.actions.loading")
                      : t("rightsRequestsPage.actions.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {t("rightsRequestsPage.empty.title")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {t("rightsRequestsPage.empty.description")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}

function RequestRow({
  request,
  connectionId,
  hasAnyAction,
}: {
  request: NodeOf<
    NonNullable<RightsRequestsPageFragment$data["rightsRequests"]>
  >;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { i18n, t } = useTranslation();
  const [deleteRequest] = useMutation<RightsRequestGraphDeleteMutation>(deleteRightsRequestMutation);
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        promisifyMutation(deleteRequest)({
          variables: {
            input: {
              rightsRequestId: request.id,
            },
            connections: [connectionId],
          },
        }),
      {
        message: t("rightsRequestsPage.deleteConfirmation"),
      },
    );
  };

  const detailsUrl = `/organizations/${organizationId}/rights-requests/${request.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>
        <Badge variant="neutral">
          {t(`rightsRequestsPage.types.${request.requestType.toLowerCase()}`)}
        </Badge>
      </Td>
      <Td>
        <Badge
          variant={getRightsRequestStateVariant(request.requestState)}
        >
          {t(`rightsRequestsPage.states.${request.requestState.toLowerCase()}`)}
        </Badge>
      </Td>
      <Td>{request.dataSubject || "-"}</Td>
      <Td>{request.contact || "-"}</Td>
      <Td>
        {request.deadline
          ? (
              <time dateTime={request.deadline}>
                {dateFormat(i18n.language, request.deadline)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">
                {t("rightsRequestsPage.noDeadline")}
              </span>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {request.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {t("rightsRequestsPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
