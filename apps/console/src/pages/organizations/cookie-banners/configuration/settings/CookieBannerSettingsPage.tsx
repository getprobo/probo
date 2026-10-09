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

import { TrashIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerSettingsPageQuery } from "#/__generated__/core/CookieBannerSettingsPageQuery.graphql";

import { CookieBannerPageHeader } from "../../_components/CookieBannerPageHeader";
import { cookieBannerPage } from "../../variants";
import { DeleteCookieBannerDialog } from "../_components/DeleteCookieBannerDialog";

import { BannerSettingsForm } from "./_components/BannerSettingsForm";
import { CategoryList } from "./_components/CategoryList";
import { ThemeSection } from "./_components/ThemeSection";

export const cookieBannerSettingsPageQuery = graphql`
  query CookieBannerSettingsPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        id
        name
        canDelete: permission(action: "core:cookie-banner:delete")
        ...BannerSettingsForm_cookieBanner
        ...CategoryList_cookieBanner
        ...ThemeSection_cookieBanner
      }
    }
  }
`;

interface CookieBannerSettingsPageProps {
  queryRef: PreloadedQuery<CookieBannerSettingsPageQuery>;
}

export function CookieBannerSettingsPage({
  queryRef,
}: CookieBannerSettingsPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const title = t("settingsPage.title");
  usePageTitle(title);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const data = usePreloadedQuery<CookieBannerSettingsPageQuery>(
    cookieBannerSettingsPageQuery,
    queryRef,
  );

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;

  return (
    <div className={cookieBannerPage()}>
      <CookieBannerPageHeader
        title={title}
        description={t("settingsPage.description")}
        actions={banner.canDelete
          ? (
              <Button
                size={2}
                variant="solid"
                color="red"
                iconStart={<TrashIcon />}
                onClick={() => setDeleteOpen(true)}
              >
                {t("configLayout.actions.delete")}
              </Button>
            )
          : undefined}
      />
      <BannerSettingsForm cookieBannerKey={banner} />
      <CategoryList cookieBannerKey={banner} />
      <ThemeSection cookieBannerKey={banner} />
      {banner.canDelete && (
        <DeleteCookieBannerDialog
          cookieBannerId={banner.id}
          name={banner.name}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
        />
      )}
    </div>
  );
}
