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

import { FileTextIcon, PlusIcon, UploadSimpleIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation, useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { FindingsPageQuery } from "#/__generated__/core/FindingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { AuditFilterSelect } from "./_components/AuditFilterSelect";
import { FindingList } from "./_components/FindingList/FindingList";
import { FindingsOwnerFilter } from "./_components/FindingsOwnerFilter";
import { findingsPage } from "./_components/variants";
import {
  findingKindKey,
  findingKinds,
  findingPriorities,
  findingStatuses,
  isFindingKind,
  isFindingPriority,
  isFindingStatus,
} from "./_lib/findingOptions";
import { findingDocumentPath, findingNewPath } from "./_lib/findingPath";
import { useFindingsListFilters } from "./_lib/useFindingsListFilters";
import { PublishFindingListDialog } from "./dialogs/PublishFindingListDialog";

export const findingsPageQuery = graphql`
  query FindingsPageQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
    $kind: FindingKind
    $status: FindingStatus
    $priority: FindingPriority
    $ownerId: ID
    $auditId: ID
  ) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateFinding: permission(action: "core:finding:create")
        canPublishFindings: permission(action: "core:finding:publish")
        findingsDocument {
          id
          defaultApprovers {
            id
          }
        }
        findings(
          first: $first
          after: $after
          last: $last
          before: $before
          orderBy: { field: CREATED_AT, direction: DESC }
          filter: {
            kind: $kind
            status: $status
            priority: $priority
            ownerId: $ownerId
            auditId: $auditId
          }
        ) {
          totalCount
        }
        ...FindingList_organization
          @arguments(
            first: $first
            after: $after
            last: $last
            before: $before
            kind: $kind
            status: $status
            priority: $priority
            ownerId: $ownerId
            auditId: $auditId
          )
      }
    }
  }
`;

interface FindingsPageProps {
  queryRef: PreloadedQuery<FindingsPageQuery>;
}

export function FindingsPage({ queryRef }: FindingsPageProps) {
  const { t } = useTranslation("organizations/findings");
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const location = useLocation();
  const filters = useFindingsListFilters();
  const { organization } = usePreloadedQuery<FindingsPageQuery>(findingsPageQuery, queryRef);
  const queriedCount = organization.__typename === "Organization"
    ? organization.findings?.totalCount ?? 0
    : 0;
  const [totalCount, setTotalCount] = useState(queriedCount);
  const [seenQueryCount, setSeenQueryCount] = useState(queriedCount);
  if (queriedCount !== seenQueryCount) {
    setSeenQueryCount(queriedCount);
    setTotalCount(queriedCount);
  }

  usePageTitle(t("list.title"));

  const handleTotalCount = useCallback((count: number) => {
    setTotalCount(count);
  }, []);

  if (organization.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  const canCreate = organization.canCreateFinding ?? false;
  const documentId = organization.findingsDocument?.id;
  const defaultApproverIds = organization.findingsDocument?.defaultApprovers.map(approver => approver.id);
  const {
    root,
    header,
    intro,
    title,
    actions,
    results,
    tableFilters,
    filter,
  } = findingsPage();

  const create = canCreate
    ? (
        <ButtonLink
          to={{ pathname: findingNewPath(organizationId), search: location.search }}
          variant="solid"
          iconStart={<PlusIcon />}
        >
          {t("list.actions.new")}
        </ButtonLink>
      )
    : null;

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <div className={title()}>
            <Heading level={1} size={6} weight="medium" highContrast>
              {t("list.title")}
            </Heading>
            <Text size={2} color="faint">{totalCount}</Text>
          </div>
          <Text size={2} color="faint">{t("list.description")}</Text>
        </div>
        <div className={actions()}>
          {documentId && (
            <ButtonLink
              to={findingDocumentPath(organizationId, documentId)}
              variant="soft"
              color="neutral"
              iconStart={<FileTextIcon />}
            >
              {t("list.actions.document")}
            </ButtonLink>
          )}
          {organization.canPublishFindings && (
            <PublishFindingListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={(publishedDocumentId) => {
                void navigate(findingDocumentPath(organizationId, publishedDocumentId));
              }}
            >
              <Button variant="soft" color="neutral" iconStart={<UploadSimpleIcon />}>
                {t("list.actions.publish")}
              </Button>
            </PublishFindingListDialog>
          )}
          {create}
        </div>
      </div>
      <div className={results()}>
        <div className={tableFilters()}>
          <div className={filter()}>
            <Select<string | null>
              value={filters.kind}
              onValueChange={(value) => {
                filters.setKind(value != null && isFindingKind(value) ? value : null);
              }}
            >
              <SelectTrigger size={2} placeholder={t("list.filters.allKinds")} aria-label={t("list.columns.kind")}>
                {(value: string | null) => value == null
                  ? t("list.filters.allKinds")
                  : t(`list.kinds.${findingKindKey(value)}`)}
              </SelectTrigger>
              <SelectPopup align="end">
                <SelectItem value={null}>{t("list.filters.allKinds")}</SelectItem>
                {findingKinds.map(kind => (
                  <SelectItem key={kind} value={kind}>
                    {t(`list.kinds.${findingKindKey(kind)}`)}
                  </SelectItem>
                ))}
              </SelectPopup>
            </Select>
          </div>
          <div className={filter()}>
            <Select<string | null>
              value={filters.status}
              onValueChange={(value) => {
                filters.setStatus(value != null && isFindingStatus(value) ? value : null);
              }}
            >
              <SelectTrigger size={2} placeholder={t("list.filters.allStatuses")} aria-label={t("list.columns.status")}>
                {(value: string | null) => value == null
                  ? t("list.filters.allStatuses")
                  : t(`list.statuses.${value.toLowerCase()}`)}
              </SelectTrigger>
              <SelectPopup align="end">
                <SelectItem value={null}>{t("list.filters.allStatuses")}</SelectItem>
                {findingStatuses.map(status => (
                  <SelectItem key={status} value={status}>
                    {t(`list.statuses.${status.toLowerCase()}`)}
                  </SelectItem>
                ))}
              </SelectPopup>
            </Select>
          </div>
          <div className={filter()}>
            <Select<string | null>
              value={filters.priority}
              onValueChange={(value) => {
                filters.setPriority(value != null && isFindingPriority(value) ? value : null);
              }}
            >
              <SelectTrigger size={2} placeholder={t("list.filters.allPriorities")} aria-label={t("list.columns.priority")}>
                {(value: string | null) => value == null
                  ? t("list.filters.allPriorities")
                  : t(`list.priorities.${value.toLowerCase()}`)}
              </SelectTrigger>
              <SelectPopup align="end">
                <SelectItem value={null}>{t("list.filters.allPriorities")}</SelectItem>
                {findingPriorities.map(priority => (
                  <SelectItem key={priority} value={priority}>
                    {t(`list.priorities.${priority.toLowerCase()}`)}
                  </SelectItem>
                ))}
              </SelectPopup>
            </Select>
          </div>
          <div className={filter()}>
            <FindingsOwnerFilter />
          </div>
          <div className={filter()}>
            <AuditFilterSelect
              organizationId={organizationId}
              value={filters.auditId}
              onChange={filters.setAuditId}
            />
          </div>
        </div>
        <FindingList
          organizationKey={organization}
          loadedVariables={queryRef.variables}
          create={create}
          onTotalCount={handleTotalCount}
        />
      </div>
    </div>
  );
}
