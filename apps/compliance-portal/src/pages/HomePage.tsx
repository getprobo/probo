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

import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import { ComplianceFrameworksSection } from "#/components/ComplianceFrameworks/ComplianceFrameworksSection";
import { Hero } from "#/components/Hero/Hero";
import { TrustCenterContactInfo } from "#/components/Hero/TrustCenterContactInfo";
import { RecentUpdatesSection } from "#/components/RecentUpdates/RecentUpdatesSection";
import { SecurityCommitmentsSection } from "#/components/SecurityCommitments/SecurityCommitmentsSection";
import { TrustedBySection } from "#/components/TrustedBy/TrustedBySection";

import type { HomePageQuery } from "./__generated__/HomePageQuery.graphql";

export const homePageQuery = graphql`
  query HomePageQuery @throwOnFieldError {
    currentTrustCenter @required(action: THROW) {
      organization {
        name
      }
      ...TrustCenterContactInfo_trustCenter
      ...ComplianceFrameworksSection_trustCenter
      ...SecurityCommitmentsSection_trustCenter
      ...TrustedBySection_trustCenter
      ...RecentUpdatesSection_trustCenter
    }
  }
`;

interface HomePageProps {
  queryRef: PreloadedQuery<HomePageQuery>;
}

export function HomePage({ queryRef }: HomePageProps) {
  const { t } = useTranslation();
  const data = usePreloadedQuery<HomePageQuery>(homePageQuery, queryRef);
  const { currentTrustCenter } = data;
  const { organization } = currentTrustCenter;

  return (
    <>
      <Hero
        title={t("home.heroTitle", { name: organization.name })}
        description={t("home.heroDescription")}
      >
        <TrustCenterContactInfo trustCenterKey={currentTrustCenter} />
      </Hero>
      <div className="flex w-full flex-col items-center px-8 max-md:px-4">
        <div className="flex w-full max-w-5xl flex-col">
          <ComplianceFrameworksSection trustCenterKey={currentTrustCenter} />
          <SecurityCommitmentsSection trustCenterKey={currentTrustCenter} />
          <TrustedBySection trustCenterKey={currentTrustCenter} />
          <RecentUpdatesSection trustCenterKey={currentTrustCenter} />
        </div>
      </div>
    </>
  );
}
