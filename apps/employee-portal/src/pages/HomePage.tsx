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

import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { HomePageQuery } from "#/__generated__/core/HomePageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { ApprovalDashboardCard } from "#/pages/_components/ApprovalDashboardCard";
import { DeviceCard } from "#/pages/_components/DeviceCard";
import { GetStartedCard } from "#/pages/_components/GetStartedCard";
import { SignatureDashboardCard } from "#/pages/_components/SignatureDashboardCard";
import { SlackCard } from "#/pages/_components/SlackCard";
import { useViewerFirstName } from "#/pages/iam/_lib/ViewerIdentityContext";

export const homePageQuery = graphql`
  query HomePageQuery($organizationId: ID!) @throwOnFieldError {
    viewer @required(action: THROW) {
      pendingSignatures: signableDocuments(
        organizationId: $organizationId
        first: 1
        filter: { signed: false }
      ) {
        totalCount
      }
      completedSignatures: signableDocuments(
        organizationId: $organizationId
        filter: { signed: true }
      ) {
        totalCount
      }
      pendingApprovals: approvableDocuments(
        organizationId: $organizationId
        first: 1
        filter: { approvalStates: [PENDING] }
      ) {
        totalCount
      }
      approvedDocuments: approvableDocuments(
        organizationId: $organizationId
        filter: { approvalStates: [APPROVED] }
      ) {
        totalCount
      }
      ...GetStartedCard_viewer @arguments(organizationId: $organizationId)
      ...SignatureDashboardCard_viewer @arguments(organizationId: $organizationId)
      ...ApprovalDashboardCard_viewer @arguments(organizationId: $organizationId)
      ...DeviceCard_viewer @arguments(organizationId: $organizationId)
      ...SlackCard_viewer
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...DeviceCard_organization
        ...SlackCard_organization
      }
    }
  }
`;

interface HomePageProps {
  queryRef: PreloadedQuery<HomePageQuery>;
}

export function HomePage({ queryRef }: HomePageProps) {
  const { t } = useTranslation();
  const firstName = useViewerFirstName();
  const { viewer, organization } = usePreloadedQuery<HomePageQuery>(
    homePageQuery,
    queryRef,
  );

  if (organization == null || organization.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  const showGetStarted
    = (viewer.pendingSignatures.totalCount > 0 || viewer.pendingApprovals.totalCount > 0)
      && viewer.completedSignatures.totalCount === 0
      && viewer.approvedDocuments.totalCount === 0;

  const welcome = firstName === ""
    ? t("homePage.welcomeFallback")
    : t("homePage.welcome", { name: firstName });

  return (
    <main className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32">
      <div className="flex flex-col gap-4">
        <Text size={2} weight="medium" color="neutral">
          {t("homePage.breadcrumb")}
        </Text>
        <Heading level={1} size={7} weight="medium" highContrast>
          {welcome}
        </Heading>
      </div>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {showGetStarted && (
          <div className="md:row-span-2">
            <GetStartedCard viewerKey={viewer} />
          </div>
        )}
        <SignatureDashboardCard
          viewerKey={viewer}
          wash={!showGetStarted && viewer.pendingSignatures.totalCount > 0}
        />
        <ApprovalDashboardCard
          viewerKey={viewer}
          wash={!showGetStarted && viewer.pendingApprovals.totalCount > 0}
        />
        <DeviceCard
          viewerKey={viewer}
          organizationKey={organization}
        />
        <SlackCard
          viewerKey={viewer}
          organizationKey={organization}
        />
      </div>
    </main>
  );
}
