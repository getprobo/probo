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

import { Link } from "@probo/ui/src/v2/Button/Link";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { InlineError } from "@probo/ui/src/v2/InlineError/InlineError";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { HomeSection } from "#/components/HomeSection/HomeSection";
import { MailingListUpdateListItem } from "#/components/MailingListUpdateListItem/MailingListUpdateListItem";
import { dotPatternStyle } from "#/components/MediaTile/variants";

import type { RecentUpdatesSection_trustCenter$key } from "./__generated__/RecentUpdatesSection_trustCenter.graphql";

// @throwOnFieldError surfaces a field error at the read below so the section
// ErrorBoundary contains it. See contrib/claude/error-handling.md.
const recentUpdatesSectionFragment = graphql`
  fragment RecentUpdatesSection_trustCenter on TrustCenter @throwOnFieldError {
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
// link to the full updates page. A load failure degrades to an inline error.
export function RecentUpdatesSection({ trustCenterKey }: RecentUpdatesSectionProps) {
  const { t } = useTranslation();

  return (
    <ErrorBoundary
      fallback={(_, reset) => (
        <HomeSection title={t("home.sections.recentUpdates")}>
          <InlineError
            message={t("errors.inline.message")}
            retryLabel={t("errors.inline.retry")}
            onRetry={reset}
          />
        </HomeSection>
      )}
    >
      <RecentUpdatesSectionContent trustCenterKey={trustCenterKey} />
    </ErrorBoundary>
  );
}

function RecentUpdatesSectionContent({ trustCenterKey }: RecentUpdatesSectionProps) {
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
            // A single failing row degrades to a compact horizontal inline error.
            <ErrorBoundary
              key={update.id}
              fallback={(_, reset) => (
                <div className="px-6 py-4">
                  <InlineError
                    layout="horizontal"
                    message={t("errors.inline.message")}
                    retryLabel={t("errors.inline.retry")}
                    onRetry={reset}
                  />
                </div>
              )}
            >
              <MailingListUpdateListItem updateKey={update} />
            </ErrorBoundary>
          ))}
        </div>
      </div>
    </HomeSection>
  );
}
