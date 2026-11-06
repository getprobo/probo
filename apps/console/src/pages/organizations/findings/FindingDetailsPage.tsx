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

import { CaretLeftIcon, TrashIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { TaskDetails } from "@probo/tasks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { useLocation, useNavigate } from "react-router";

import type { FindingDetailsPage_finding$key } from "#/__generated__/core/FindingDetailsPage_finding.graphql";
import type { FindingDetailsPageQuery } from "#/__generated__/core/FindingDetailsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { FindingAuditsSection } from "./_components/FindingAuditsSection";
import { FindingDeleteDialog } from "./_components/FindingList/FindingDeleteDialog";
import { FindingNarrativeSection } from "./_components/FindingNarrativeSection";
import { FindingPropertiesSection } from "./_components/FindingPropertiesSection";
import { FindingTasksSection } from "./_components/FindingTasksSection";
import { FindingTitleField } from "./_components/FindingTitleField";
import { findingsPage } from "./_components/variants";
import { useUpdateFinding } from "./_lib/useUpdateFinding";

export const findingDetailsPageFragment = graphql`
  fragment FindingDetailsPage_finding on Finding {
    id
    title
    referenceId
    canUpdate: permission(action: "core:finding:update")
    canDelete: permission(action: "core:finding:delete")
    organization {
      id
    }
    ...FindingPropertiesSection_finding
    ...FindingNarrativeSection_finding
    ...FindingTasksSection_finding
    ...FindingAuditsSection_finding
  }
`;

export const findingDetailsPageQuery = graphql`
  query FindingDetailsPageQuery($findingId: ID!) {
    node(id: $findingId) @required(action: THROW) {
      __typename
      ... on Finding {
        ...FindingDetailsPage_finding
      }
    }
  }
`;

interface FindingDetailsPageProps {
  queryRef: PreloadedQuery<FindingDetailsPageQuery>;
}

export function FindingDetailsPage({ queryRef }: FindingDetailsPageProps) {
  const data = usePreloadedQuery<FindingDetailsPageQuery>(findingDetailsPageQuery, queryRef);
  const { t } = useTranslation("organizations/findings");
  const { back, side } = findingsPage();
  const navigate = useNavigate();
  const location = useLocation();
  const organizationId = useOrganizationId();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const findingKey = data.node.__typename === "Finding" ? data.node : null;
  const finding = useFragment<FindingDetailsPage_finding$key>(
    findingDetailsPageFragment,
    findingKey,
  );
  const [updateFinding] = useUpdateFinding();
  usePageTitle(finding?.title ?? "");

  if (!finding || finding.organization?.id !== organizationId) {
    throw new NotFoundError(t("details.notFound"));
  }

  const backTarget = {
    pathname: `/organizations/${organizationId}/governance/findings`,
    search: location.search,
  };

  return (
    <>
      <TaskDetails
        back={(
          <Link
            to={backTarget}
            size={2}
            color="neutral"
            underline={false}
            iconStart={<CaretLeftIcon />}
            className={back()}
          >
            {t("details.back")}
          </Link>
        )}
        name={(
          <FindingTitleField
            title={finding.title}
            canUpdate={finding.canUpdate}
            onSave={title => updateFinding({
              variables: { input: { id: finding.id, title } },
            })}
          />
        )}
        actions={finding.canDelete
          ? (
              <Button color="red" iconStart={<TrashIcon />} onClick={() => setDeleteOpen(true)}>
                {t("list.actions.delete")}
              </Button>
            )
          : undefined}
        engagement={<FindingNarrativeSection findingKey={finding} />}
        properties={(
          <div className={side()}>
            <FindingPropertiesSection findingKey={finding} />
            <FindingTasksSection findingKey={finding} />
            <FindingAuditsSection findingKey={finding} />
          </div>
        )}
      />
      {finding.canDelete && (
        <FindingDeleteDialog
          findingId={finding.id}
          referenceId={finding.referenceId}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
          onDeleted={() => {
            void navigate(backTarget);
          }}
        />
      )}
    </>
  );
}
