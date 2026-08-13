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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useLazyLoadQuery } from "react-relay";
import { useParams } from "react-router";

import type { CookieBannerSwitcherValueQuery } from "#/__generated__/core/CookieBannerSwitcherValueQuery.graphql";

import { cookieBannerSwitcher } from "./variants";

const cookieBannerSwitcherValueQuery = graphql`
  query CookieBannerSwitcherValueQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        name
      }
    }
  }
`;

export interface CookieBannerSwitcherValueProps {
  fallback: string;
}

/**
 * The name of the banner the URL is on, or the default trigger label.
 *
 * Mounted only when `:cookieBannerId` is present. Other privacy pages keep
 * "Select Banner"; the create form uses "New Banner" in the trigger itself.
 */
export function CookieBannerSwitcherValue({ fallback }: CookieBannerSwitcherValueProps) {
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const data = useLazyLoadQuery<CookieBannerSwitcherValueQuery>(
    cookieBannerSwitcherValueQuery,
    { cookieBannerId: cookieBannerId ?? "" },
    { fetchPolicy: "store-or-network" },
  );
  const slots = cookieBannerSwitcher();
  if (data.node?.__typename !== "CookieBanner") {
    return fallback;
  }
  return (
    <Text size={2} weight="medium" color="neutral" highContrast className={slots.itemName()}>
      {data.node.name}
    </Text>
  );
}

/**
 * A pulse bar the same height as the trigger name, used while the selected
 * banner query suspends. A plain "Select Banner" string inherited the page
 * font and flashed larger than the closed trigger.
 */
export function CookieBannerSwitcherValueSkeleton() {
  const slots = cookieBannerSwitcher();

  return <span className={slots.valueSkeletonName()} aria-hidden />;
}
