// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { Link } from "@probo/ui/src/v2/Button/Link";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { HomeSection } from "#/components/HomeSection/HomeSection";
import { MailingListUpdateListItem } from "#/components/MailingListUpdateListItem/MailingListUpdateListItem";
import { dotPatternStyle } from "#/components/MediaTile/variants";

import type { RecentUpdatesSection_trustCenter$key } from "./__generated__/RecentUpdatesSection_trustCenter.graphql";

const recentUpdatesSectionFragment = graphql`
  fragment RecentUpdatesSection_trustCenter on TrustCenter {
    updates(first: 5) {
      edges {
        node {
          id
          ...MailingListUpdateListItem_update
        }
      }
    }
  }
`;

interface RecentUpdatesSectionProps {
  trustCenterKey: RecentUpdatesSection_trustCenter$key;
}

// "Recent updates" section: the latest mailing-list updates as a list, with a
// link to the full updates page.
export function RecentUpdatesSection({ trustCenterKey }: RecentUpdatesSectionProps) {
  const { t } = useTranslation();
  const data = useFragment(recentUpdatesSectionFragment, trustCenterKey);
  const updates = data.updates.edges.map(edge => edge.node);

  if (updates.length === 0) {
    return null;
  }

  return (
    <HomeSection
      title={t("home.sections.recentUpdates")}
      action={(
        <Link to="/updates" variant="ghost" color="neutral" size={2}>
          {t("home.recentUpdates.viewAll")}
        </Link>
      )}
    >
      <div className="relative overflow-hidden rounded-5 border border-sand-3 bg-sand-1">
        <div aria-hidden className="pointer-events-none absolute inset-0" style={dotPatternStyle} />
        <div aria-hidden className="pointer-events-none absolute inset-0 bg-linear-to-r from-sand-1/0 to-sand-1 to-[96px]" />
        <div className="relative divide-y divide-sand-a2">
          {updates.map(update => (
            <MailingListUpdateListItem key={update.id} updateKey={update} />
          ))}
        </div>
      </div>
    </HomeSection>
  );
}
