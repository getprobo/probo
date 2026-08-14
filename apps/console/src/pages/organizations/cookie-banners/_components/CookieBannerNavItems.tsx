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

export interface CookieBannerNavItemsProps {
  cookieBannerId: string;
}

/**
 * Configure, Discovery, and Trail for the selected banner.
 *
 * Hidden until a banner is selected (URL or newest default): these paths
 * need an id, and the create page is not one of the three sections.
 */
export function CookieBannerNavItems({ cookieBannerId }: CookieBannerNavItemsProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const prefix = `/organizations/${organizationId}/privacy/cookie-banners/${cookieBannerId}`;

  return (
    <>
      <NavPanelItem label={t("nav.cookieBannersConfigure")} to={`${prefix}/configure`} />
      <NavPanelItem label={t("nav.cookieBannersDiscovery")} to={`${prefix}/discovery`} />
      <NavPanelItem label={t("nav.cookieBannersTrail")} to={`${prefix}/trail`} />
    </>
  );
}
