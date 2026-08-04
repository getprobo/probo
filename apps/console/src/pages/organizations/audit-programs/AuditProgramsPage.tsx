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

import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { AuditProgramGraphListQuery } from "#/__generated__/core/AuditProgramGraphListQuery.graphql";
import type {
  AuditProgramsPageFragment$data,
  AuditProgramsPageFragment$key,
} from "#/__generated__/core/AuditProgramsPageFragment.graphql";
import { SortableTable } from "#/components/SortableTable";
import {
  auditProgramsQuery,
  useDeleteAuditProgram,
} from "#/hooks/graph/AuditProgramGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { CreateAuditProgramDialog } from "./dialogs/CreateAuditProgramDialog";

const paginatedAuditProgramsFragment = graphql`
  fragment AuditProgramsPageFragment on Organization
  @refetchable(queryName: "AuditProgramsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    orderBy: { type: "AuditProgramOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    auditPrograms(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
    ) @connection(key: "AuditProgramsPage_auditPrograms") {
      __id
      edges {
        node {
          id
          name
          validFrom
          validUntil
          framework {
            id
            name
          }
          canUpdate: permission(action: "core:audit-program:update")
          canDelete: permission(action: "core:audit-program:delete")
        }
      }
    }
  }
`;

type AuditProgramEntry = NodeOf<
  AuditProgramsPageFragment$data["auditPrograms"]
>;

type Props = {
  queryRef: PreloadedQuery<AuditProgramGraphListQuery>;
};

export default function AuditProgramsPage(props: Props) {
  const { t } = useTranslation("organizations/audit-programs");
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery<AuditProgramGraphListQuery>(
    auditProgramsQuery,
    props.queryRef,
  );
  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(
    paginatedAuditProgramsFragment,
    data.node as AuditProgramsPageFragment$key,
  );
  const programs
    = pagination.data.auditPrograms?.edges?.map(edge => edge.node) ?? [];
  const connectionId = pagination.data.auditPrograms.__id;

  usePageTitle(t("auditProgramsPage.title"));

  const hasAnyAction = programs.some(
    program => program.canDelete || program.canUpdate,
  );

  const canCreateAuditProgram = data.node.canCreateAuditProgram;

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("auditProgramsPage.title")}
        description={t("auditProgramsPage.description")}
      >
        {canCreateAuditProgram && (
          <CreateAuditProgramDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>
              {t("auditProgramsPage.actions.add")}
            </Button>
          </CreateAuditProgramDialog>
        )}
      </PageHeader>
      <SortableTable {...pagination} pageSize={10}>
        <Thead>
          <Tr>
            <Th>{t("auditProgramsPage.columns.name")}</Th>
            <Th>{t("auditProgramsPage.columns.framework")}</Th>
            <Th>{t("auditProgramsPage.columns.validFrom")}</Th>
            <Th>{t("auditProgramsPage.columns.validUntil")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {programs.map(entry => (
            <AuditProgramRow
              key={entry.id}
              entry={entry}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function AuditProgramRow({
  entry,
  connectionId,
  hasAnyAction,
}: {
  entry: AuditProgramEntry;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { i18n, t } = useTranslation("organizations/audit-programs");
  const deleteAuditProgram = useDeleteAuditProgram(
    { id: entry.id, name: entry.name },
    connectionId,
  );

  return (
    <Tr
      to={`/organizations/${organizationId}/audits/programs/${entry.id}`}
    >
      <Td>{entry.name}</Td>
      <Td>
        {entry.framework?.name ?? t("auditProgramsPage.row.unknownFramework")}
      </Td>
      <Td>
        {dateFormat(i18n.language, entry.validFrom)
          || t("auditProgramsPage.row.notSet")}
      </Td>
      <Td>
        {dateFormat(i18n.language, entry.validUntil)
          || t("auditProgramsPage.row.notSet")}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {entry.canDelete && (
              <DropdownItem
                onClick={deleteAuditProgram}
                variant="danger"
                icon={IconTrashCan}
              >
                {t("auditProgramsPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
