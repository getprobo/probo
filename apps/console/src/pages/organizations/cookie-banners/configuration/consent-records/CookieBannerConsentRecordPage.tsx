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

import { CaretLeftIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannerConsentRecordPageQuery } from "#/__generated__/core/CookieBannerConsentRecordPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CookieBannerPageHeader } from "../../_components/CookieBannerPageHeader";
import { cookieBannerTrailPath } from "../../_lib/cookieBannerPaths";
import { consentRecordPage } from "../../variants";

import { ConsentRecordActionCard } from "./_components/ConsentRecordActionCard";
import { ConsentRecordCategoriesCard } from "./_components/ConsentRecordCategoriesCard";
import { ConsentRecordLocationCard } from "./_components/ConsentRecordLocationCard";
import { ConsentRecordRequestCard } from "./_components/ConsentRecordRequestCard";
import { ConsentRecordTcfCard } from "./_components/ConsentRecordTcfCard";

export const cookieBannerConsentRecordPageQuery = graphql`
  query CookieBannerConsentRecordPageQuery($consentRecordId: ID!) {
    node(id: $consentRecordId) @required(action: THROW) {
      __typename
      ... on CookieConsentRecord {
        ...ConsentRecordActionCard_cookieConsentRecord
        ...ConsentRecordRequestCard_cookieConsentRecord
        ...ConsentRecordTcfCard_cookieConsentRecord
        ...ConsentRecordCategoriesCard_cookieConsentRecord
        ...ConsentRecordLocationCard_cookieConsentRecord
      }
    }
  }
`;

interface CookieBannerConsentRecordPageProps {
  queryRef: PreloadedQuery<CookieBannerConsentRecordPageQuery>;
}

export default function CookieBannerConsentRecordPage({
  queryRef,
}: CookieBannerConsentRecordPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const organizationId = useOrganizationId();
  const location = useLocation();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const title = t("consentRecordPage.title");
  usePageTitle(title);
  const data = usePreloadedQuery<CookieBannerConsentRecordPageQuery>(
    cookieBannerConsentRecordPageQuery,
    queryRef,
  );

  if (data.node.__typename !== "CookieConsentRecord") {
    throw new Error("invalid type for node");
  }

  if (cookieBannerId == null) {
    throw new Error(":cookieBannerId missing in route params");
  }

  const {
    root,
    header,
    back,
    body,
    column,
    actionCard,
    locationCard,
    requestCard,
    tcfCard,
    categoriesCard,
  } = consentRecordPage();

  return (
    <div className={root()}>
      <div className={header()}>
        <Link
          to={{
            pathname: cookieBannerTrailPath(organizationId, cookieBannerId),
            search: location.search,
          }}
          size={2}
          color="neutral"
          underline={false}
          iconStart={<CaretLeftIcon />}
          className={back()}
        >
          {t("consentRecordsPage.title")}
        </Link>
        <CookieBannerPageHeader
          title={title}
          description={t("consentRecordPage.description")}
        />
      </div>
      <div className={body()}>
        <div className={column()}>
          <ConsentRecordActionCard
            className={actionCard()}
            cookieConsentRecordKey={data.node}
          />
          <ConsentRecordRequestCard
            className={requestCard()}
            cookieConsentRecordKey={data.node}
          />
          <ConsentRecordTcfCard
            className={tcfCard()}
            cookieConsentRecordKey={data.node}
          />
        </div>
        <div className={column()}>
          <ConsentRecordCategoriesCard
            className={categoriesCard()}
            cookieConsentRecordKey={data.node}
          />
          <ConsentRecordLocationCard
            className={locationCard()}
            cookieConsentRecordKey={data.node}
          />
        </div>
      </div>
    </div>
  );
}
