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

import type { ProcessingActivityGraphTIAListQuery } from "#/__generated__/core/ProcessingActivityGraphTIAListQuery.graphql";
import type { TransferImpactAssessmentsPageFragment$key } from "#/__generated__/core/TransferImpactAssessmentsPageFragment.graphql";
import { transferImpactAssessmentsQuery } from "#/hooks/graph/ProcessingActivityGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { TransferImpactAssessmentListItem } from "./_components/TransferImpactAssessmentListItem";
import { PublishTransferImpactAssessmentListDialog } from "./dialogs/PublishTransferImpactAssessmentListDialog";

interface TransferImpactAssessmentsPageProps {
  queryRef: PreloadedQuery<ProcessingActivityGraphTIAListQuery>;
}

const transferImpactAssessmentsPageFragment = graphql`
  fragment TransferImpactAssessmentsPageFragment on Organization
  @refetchable(queryName: "TransferImpactAssessmentsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
  ) {
    id
    transferImpactAssessments(first: $first, after: $after)
      @connection(
        key: "TransferImpactAssessmentsPage_transferImpactAssessments"
      ) {
      edges {
        node {
          id
          ...TransferImpactAssessmentListItem_transferImpactAssessment
        }
      }
    }
  }
`;

export default function TransferImpactAssessmentsPage({
  queryRef,
}: TransferImpactAssessmentsPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  usePageTitle(t("nav.transferImpactAssessments"));

  const organization = usePreloadedQuery<ProcessingActivityGraphTIAListQuery>(
    transferImpactAssessmentsQuery,
    queryRef,
  );

  const tiaDocument = organization.node.transferImpactAssessmentsDocument;
  const tiaDefaultApproverIds = (tiaDocument?.defaultApprovers ?? []).map(
    a => a.id,
  );

  const goToDocument = (documentId: string) => {
    void navigate(
      `/organizations/${organizationId}/governance/documents/${documentId}`,
    );
  };

  const {
    data: tiaData,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    ProcessingActivityGraphTIAListQuery,
    TransferImpactAssessmentsPageFragment$key
  >(transferImpactAssessmentsPageFragment, organization.node);

  const tias
    = tiaData.transferImpactAssessments?.edges?.map(edge => edge.node) ?? [];

  const canPublishTIA
    = organization.node.canPublishTransferImpactAssessments;

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("nav.transferImpactAssessments")}
        description={t("transferImpactAssessmentsPage.description")}
      />

      <div className="flex justify-end gap-2">
        {tiaDocument?.id && (
          <Button variant="secondary" asChild>
            <Link
              to={`/organizations/${organizationId}/governance/documents/${tiaDocument.id}`}
            >
              <IconPageTextLine size={16} />
              {t("transferImpactAssessmentsPage.actions.document")}
            </Link>
          </Button>
        )}
        {canPublishTIA && (
          <PublishTransferImpactAssessmentListDialog
            organizationId={organizationId}
            defaultApproverIds={tiaDefaultApproverIds}
            onPublished={goToDocument}
          >
            <Button variant="secondary" icon={IconUpload}>
              {t("transferImpactAssessmentsPage.actions.publish")}
            </Button>
          </PublishTransferImpactAssessmentListDialog>
        )}
      </div>

      {tias.length > 0
        ? (
            <Card>
              <Table>
                <Thead>
                  <Tr>
                    <Th>
                      {t("transferImpactAssessmentsPage.columns.processingActivity")}
                    </Th>
                    <Th>
                      {t("transferImpactAssessmentsPage.columns.dataSubjects")}
                    </Th>
                    <Th>
                      {t("transferImpactAssessmentsPage.columns.transfer")}
                    </Th>
                    <Th>
                      {t("transferImpactAssessmentsPage.columns.localLawRisk")}
                    </Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {tias.map(tia => (
                    <TransferImpactAssessmentListItem
                      key={tia.id}
                      transferImpactAssessmentKey={tia}
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
                      ? t("transferImpactAssessmentsPage.actions.loading")
                      : t("transferImpactAssessmentsPage.actions.loadMore")}
                  </Button>
                </div>
              )}
            </Card>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {t("transferImpactAssessmentsPage.empty.title")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {t("transferImpactAssessmentsPage.empty.description")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}
