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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Fragment } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AddSourceSection_organization$key } from "#/__generated__/core/AddSourceSection_organization.graphql";
import { MarketplaceEntryCard } from "#/pages/organizations/settings/integrations/_components/MarketplaceEntryCard";

import { CsvSourceCard } from "./CsvSourceCard";
import type { AddableConnectorCard } from "./ResolvedAddableConnectorCard";
import { sourcesPage } from "./variants";

const addSourceSectionFragment = graphql`
  fragment AddSourceSection_organization on Organization {
    ...MarketplaceEntryCard_organization
  }
`;

interface AddSourceSectionProps {
  cards: readonly AddableConnectorCard[];
  showCSV: boolean;
  organizationKey: AddSourceSection_organization$key;
}

export function AddSourceSection({
  cards,
  showCSV,
  organizationKey,
}: AddSourceSectionProps) {
  const { t } = useTranslation();
  const organization = useFragment(addSourceSectionFragment, organizationKey);
  const { section, sectionTitle, grid, empty } = sourcesPage();
  const hasAddable = cards.length > 0 || showCSV;
  const availableCount = cards.length + (showCSV ? 1 : 0);

  return (
    <section className={section()}>
      <div className={sectionTitle()}>
        <Heading level={2} size={3} weight="medium">
          {t("accessReviewSourcesPage.sections.addSource")}
        </Heading>
        <Text size={2} color="faint">{availableCount}</Text>
      </div>
      {!hasAddable
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {t("accessReviewSourcesPage.emptyAccounts")}
                </Text>
              </div>
            </Card>
          )
        : (
            <div className={grid()}>
              <MarketplaceEntryCard organizationKey={organization} />
              {cards.map(({ provider, card }) => (
                <Fragment key={provider}>{card}</Fragment>
              ))}
              {showCSV && (
                <CsvSourceCard />
              )}
            </div>
          )}
    </section>
  );
}
