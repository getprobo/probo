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
  getObligationStatusVariant,
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
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  usePageTitle(t("obligationsPage.title"));

  const organization = usePreloadedQuery<ObligationGraphListQuery>(obligationsQuery, queryRef);
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
        title={t("obligationsPage.title")}
        description={t("obligationsPage.description")}
      >
        <div className="flex gap-2">
          {organization.node.obligationsDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${organization.node.obligationsDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {t("obligationsPage.actions.document")}
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
                {t("obligationsPage.actions.publish")}
              </Button>
            </PublishObligationListDialog>
          )}
          {organization.node.canCreateObligation && (
            <CreateObligationDialog
              organizationId={organizationId}
              connection={connectionId}
            >
              <Button icon={IconPlusLarge}>{t("obligationsPage.actions.add")}</Button>
            </CreateObligationDialog>
          )}
        </div>
      </PageHeader>

      {obligations.length === 0
        ? (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {t("obligationsPage.empty.title")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {t("obligationsPage.empty.description")}
                </p>
              </div>
            </Card>
          )
        : (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("obligationsPage.columns.area")}</Th>
                    <Th>{t("obligationsPage.columns.source")}</Th>
                    <Th>{t("obligationsPage.columns.status")}</Th>
                    <Th>{t("obligationsPage.columns.owner")}</Th>
                    <Th>{t("obligationsPage.columns.dueDate")}</Th>
                    {hasAnyAction && <Th>{t("obligationsPage.columns.actions")}</Th>}
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
                    {t("obligationsPage.actions.loadMore")}
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
  const { i18n, t } = useTranslation();
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
        message: t("obligationsPage.deleteConfirmation"),
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
          {t(`obligationsPage.statuses.${(obligation.status || "NON_COMPLIANT").toLowerCase()}`)}
        </Badge>
      </Td>
      <Td>{obligation.owner?.fullName || "-"}</Td>
      <Td>
        {obligation.dueDate
          ? (
              <time dateTime={obligation.dueDate}>
                {dateFormat(i18n.language, obligation.dueDate)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">{t("obligationsPage.noDueDate")}</span>
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
                {t("obligationsPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
