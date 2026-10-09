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

import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NavPanelItem } from "#/pages/iam/organizations/_components/shell/NavPanelItem";

import {
  cookieBannerInstallPath,
  cookieBannerPath,
  cookieBannerResourcesPath,
  cookieBannerSettingsPath,
  cookieBannerTCFPath,
  cookieBannerTrackersPath,
  cookieBannerTrailPath,
  cookieBannerTranslationsPath,
} from "../_lib/cookieBannerPaths";

export interface CookieBannerNavItemsProps {
  cookieBannerId: string;
  tcf: boolean;
}

export function CookieBannerNavItems({ cookieBannerId, tcf }: CookieBannerNavItemsProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const consentRecordsPrefix = `${cookieBannerPath(organizationId, cookieBannerId)}/consent-records`;

  return (
    <>
      <NavPanelItem
        label={t("nav.cookieBannersSettings")}
        to={cookieBannerSettingsPath(organizationId, cookieBannerId)}
      />
      <NavPanelItem
        label={t("nav.cookieBannersInstall")}
        to={cookieBannerInstallPath(organizationId, cookieBannerId)}
      />
      <NavPanelItem
        label={t("nav.cookieBannersTranslations")}
        to={cookieBannerTranslationsPath(organizationId, cookieBannerId)}
      />
      <NavPanelItem
        label={t("nav.cookieBannersTrackers")}
        to={cookieBannerTrackersPath(organizationId, cookieBannerId)}
      />
      <NavPanelItem
        label={t("nav.cookieBannersResources")}
        to={cookieBannerResourcesPath(organizationId, cookieBannerId)}
      />
      <NavPanelItem
        label={t("nav.cookieBannersTrail")}
        to={cookieBannerTrailPath(organizationId, cookieBannerId)}
        alsoActiveFor={[consentRecordsPrefix]}
      />
      {tcf && (
        <NavPanelItem
          label={t("nav.cookieBannersTCF")}
          to={cookieBannerTCFPath(organizationId, cookieBannerId)}
        />
      )}
    </>
  );
}
