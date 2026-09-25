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
import { useParams } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NavPanelGroup } from "#/pages/iam/organizations/_components/shell/NavPanelGroup";
import { NavPanelItem } from "#/pages/iam/organizations/_components/shell/NavPanelItem";

import {
  THIRD_PARTY_SECTION_GROUPS,
  THIRD_PARTY_SECTIONS,
  thirdPartyHref,
} from "../_lib/thirdPartySections";

export function ThirdPartyNavItems() {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();

  if (thirdPartyId == null) {
    return null;
  }

  const ungrouped = THIRD_PARTY_SECTIONS.filter(section => section.group == null);

  return (
    <>
      {ungrouped.map(section => (
        <NavPanelItem
          key={section.id}
          label={t(section.labelKey)}
          to={thirdPartyHref(organizationId, thirdPartyId, section)}
        />
      ))}
      {THIRD_PARTY_SECTION_GROUPS.map((group) => {
        const items = THIRD_PARTY_SECTIONS.filter(section => section.group === group.id);
        return (
          <NavPanelGroup key={group.id} label={t(group.labelKey)}>
            {items.map(section => (
              <NavPanelItem
                key={section.id}
                label={t(section.labelKey)}
                to={thirdPartyHref(organizationId, thirdPartyId, section)}
              />
            ))}
          </NavPanelGroup>
        );
      })}
    </>
  );
}
