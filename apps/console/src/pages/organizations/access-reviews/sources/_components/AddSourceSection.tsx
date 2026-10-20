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
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Fragment } from "react";
import { useTranslation } from "react-i18next";

import { CsvSourceCard } from "./CsvSourceCard";
import type { AddableConnectorCard } from "./ResolvedAddableConnectorCard";
import { sourcesPage } from "./variants";

interface AddSourceSectionProps {
  cards: readonly AddableConnectorCard[];
  showCSV: boolean;
  isSearching: boolean;
}

export function AddSourceSection({
  cards,
  showCSV,
  isSearching,
}: AddSourceSectionProps) {
  const { t } = useTranslation();
  const { section, grid, empty } = sourcesPage();
  const hasAddable = cards.length > 0 || showCSV;

  return (
    <section className={section()}>
      {!hasAddable
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {isSearching
                    ? t("accessReviewSourcesPage.emptyAvailableSearch")
                    : t("accessReviewSourcesPage.emptyAccounts")}
                </Text>
              </div>
            </Card>
          )
        : (
            <div className={grid()}>
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
