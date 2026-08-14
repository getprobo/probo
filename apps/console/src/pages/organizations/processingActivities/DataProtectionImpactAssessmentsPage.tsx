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

import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  IconPageTextLine,
  IconUpload,
  PageHeader,
  Table,
  Tbody,
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
import { Link, useNavigate } from "react-router";

import type { DataProtectionImpactAssessmentsPageFragment$key } from "#/__generated__/core/DataProtectionImpactAssessmentsPageFragment.graphql";
import type { ProcessingActivityGraphDPIAListQuery } from "#/__generated__/core/ProcessingActivityGraphDPIAListQuery.graphql";
import { dataProtectionImpactAssessmentsQuery } from "#/hooks/graph/ProcessingActivityGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DataProtectionImpactAssessmentListItem } from "./_components/DataProtectionImpactAssessmentListItem";
import { PublishDataProtectionImpactAssessmentListDialog } from "./dialogs/PublishDataProtectionImpactAssessmentListDialog";

interface DataProtectionImpactAssessmentsPageProps {
  queryRef: PreloadedQuery<ProcessingActivityGraphDPIAListQuery>;
}

const dataProtectionImpactAssessmentsPageFragment = graphql`
  fragment DataProtectionImpactAssessmentsPageFragment on Organization
  @refetchable(queryName: "DataProtectionImpactAssessmentsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
  ) {
    id
    dataProtectionImpactAssessments(first: $first, after: $after)
      @connection(
        key: "DataProtectionImpactAssessmentsPage_dataProtectionImpactAssessments"
      ) {
      edges {
        node {
          id
          ...DataProtectionImpactAssessmentListItem_dataProtectionImpactAssessment
        }
      }
    }
  }
`;

export default function DataProtectionImpactAssessmentsPage({
  queryRef,
}: DataProtectionImpactAssessmentsPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  usePageTitle(t("dataProtectionImpactAssessmentsPage.title"));

  const organization = usePreloadedQuery<ProcessingActivityGraphDPIAListQuery>(
    dataProtectionImpactAssessmentsQuery,
    queryRef,
  );

  const dpiaDocument = organization.node.dataProtectionImpactAssessmentsDocument;
  const dpiaDefaultApproverIds = (dpiaDocument?.defaultApprovers ?? []).map(
    a => a.id,
  );

  const goToDocument = (documentId: string) => {
    void navigate(
      `/organizations/${organizationId}/governance/documents/${documentId}`,
    );
  };

  const {
    data: dpiaData,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    ProcessingActivityGraphDPIAListQuery,
    DataProtectionImpactAssessmentsPageFragment$key
  >(dataProtectionImpactAssessmentsPageFragment, organization.node);

  const dpias
    = dpiaData.dataProtectionImpactAssessments?.edges?.map(edge => edge.node)
      ?? [];

  const canPublishDPIA
    = organization.node.canPublishDataProtectionImpactAssessments;

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("dataProtectionImpactAssessmentsPage.title")}
        description={t("dataProtectionImpactAssessmentsPage.description")}
      />

      <div className="flex justify-end gap-2">
        {dpiaDocument?.id && (
          <Button variant="secondary" asChild>
            <Link
              to={`/organizations/${organizationId}/governance/documents/${dpiaDocument.id}`}
            >
              <IconPageTextLine size={16} />
              {t("dataProtectionImpactAssessmentsPage.actions.document")}
            </Link>
          </Button>
        )}
        {canPublishDPIA && (
          <PublishDataProtectionImpactAssessmentListDialog
            organizationId={organizationId}
            defaultApproverIds={dpiaDefaultApproverIds}
            onPublished={goToDocument}
          >
            <Button variant="secondary" icon={IconUpload}>
              {t("dataProtectionImpactAssessmentsPage.actions.publish")}
            </Button>
          </PublishDataProtectionImpactAssessmentListDialog>
        )}
      </div>

      {dpias.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>
                      {t("dataProtectionImpactAssessmentsPage.columns.processingActivity")}
                    </Th>
                    <Th>
                      {t("dataProtectionImpactAssessmentsPage.columns.description")}
                    </Th>
                    <Th>
                      {t("dataProtectionImpactAssessmentsPage.columns.potentialRisk")}
                    </Th>
                    <Th>
                      {t("dataProtectionImpactAssessmentsPage.columns.residualRisk")}
                    </Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {dpias.map(dpia => (
                    <DataProtectionImpactAssessmentListItem
                      key={dpia.id}
                      dataProtectionImpactAssessmentKey={dpia}
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
                      ? t("dataProtectionImpactAssessmentsPage.actions.loading")
                      : t("dataProtectionImpactAssessmentsPage.actions.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {t("dataProtectionImpactAssessmentsPage.empty.title")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {t("dataProtectionImpactAssessmentsPage.empty.description")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
