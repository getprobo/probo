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

import { graphql, useFragment } from "react-relay";
import { useParams } from "react-router";

import type { ThirdPartySwitcherListItem_thirdParty$key } from "#/__generated__/core/ThirdPartySwitcherListItem_thirdParty.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NavPanelSwitcherListItem } from "#/pages/organizations/_components/NavPanelSwitcherListItem";

import { ThirdPartySwitcherAvatar } from "./ThirdPartySwitcherAvatar";

const thirdPartySwitcherListItemFragment = graphql`
  fragment ThirdPartySwitcherListItem_thirdParty on ThirdParty {
    id
    name
    websiteUrl
  }
`;

export interface ThirdPartySwitcherListItemProps {
  thirdPartyKey: ThirdPartySwitcherListItem_thirdParty$key;
}

/**
 * One third party in the TPRM-panel switcher.
 */
export function ThirdPartySwitcherListItem({ thirdPartyKey }: ThirdPartySwitcherListItemProps) {
  const organizationId = useOrganizationId();
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  const thirdParty = useFragment(thirdPartySwitcherListItemFragment, thirdPartyKey);

  return (
    <NavPanelSwitcherListItem
      to={`/organizations/${organizationId}/tprm/third-parties/${thirdParty.id}/overview`}
      name={thirdParty.name}
      detail={thirdParty.websiteUrl ?? undefined}
      leading={<ThirdPartySwitcherAvatar name={thirdParty.name} websiteUrl={thirdParty.websiteUrl} />}
      selected={thirdPartyId === thirdParty.id}
    />
  );
}
