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

import { graphql, useLazyLoadQuery } from "react-relay";
import { useParams } from "react-router";

import type { ThirdPartySwitcherValueQuery } from "#/__generated__/core/ThirdPartySwitcherValueQuery.graphql";
import { NavPanelSwitcherValue } from "#/pages/organizations/_components/NavPanelSwitcher";

import { ThirdPartySwitcherAvatar } from "./ThirdPartySwitcherAvatar";

const thirdPartySwitcherValueQuery = graphql`
  query ThirdPartySwitcherValueQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        name
        websiteUrl
      }
    }
  }
`;

export interface ThirdPartySwitcherValueProps {
  fallback: string;
}

export function ThirdPartySwitcherValue({ fallback }: ThirdPartySwitcherValueProps) {
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  const data = useLazyLoadQuery<ThirdPartySwitcherValueQuery>(
    thirdPartySwitcherValueQuery,
    { thirdPartyId: thirdPartyId ?? "" },
    { fetchPolicy: "store-or-network" },
  );
  if (data.node?.__typename !== "ThirdParty") {
    return fallback;
  }
  return (
    <>
      <ThirdPartySwitcherAvatar name={data.node.name} websiteUrl={data.node.websiteUrl} />
      <NavPanelSwitcherValue>{data.node.name}</NavPanelSwitcherValue>
    </>
  );
}
