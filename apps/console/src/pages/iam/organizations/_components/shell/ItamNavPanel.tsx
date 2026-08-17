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
import { graphql, useFragment } from "react-relay";

import type { ItamNavPanel_organization$key } from "#/__generated__/iam/ItamNavPanel_organization.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navHref } from "#/pages/iam/organizations/_lib/navigation";

import { NavPanelItem } from "./NavPanelItem";
import type { NavPanelBodyProps } from "./navPanels";

const itamNavPanelFragment = graphql`
  fragment ItamNavPanel_organization on Organization {
    canListDevices: permission(action: "itam:device:list")
  }
`;

export function ItamNavPanel({ organizationKey, group }: NavPanelBodyProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const organization = useFragment<ItamNavPanel_organization$key>(
    itamNavPanelFragment,
    organizationKey,
  );

  if (!organization.canListDevices) {
    return null;
  }

  return (
    <NavPanelItem
      label={t("nav.devices")}
      to={navHref(organizationId, group, "devices")}
    />
  );
}
