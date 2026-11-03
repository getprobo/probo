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

import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ScanCheckpoint_scan$key } from "#/__generated__/core/ScanCheckpoint_scan.graphql";

import { scanSection } from "../../../variants";
import { formatElapsedSeconds } from "../_lib/formatScanTime";

const scanCheckpointFragment = graphql`
  fragment ScanCheckpoint_scan on CookieBannerScan {
    checkpoint {
      visitedTemplates
      queue {
        url
      }
      coveredHosts
      pagesScanned
      elapsedMs
      truncated
    }
  }
`;

interface ScanCheckpointProps {
  scanKey: ScanCheckpoint_scan$key;
}

export function ScanCheckpoint({ scanKey }: ScanCheckpointProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const scan = useFragment(scanCheckpointFragment, scanKey);
  const checkpoint = scan.checkpoint;
  const { root, rows, row, rowBody, list, item } = scanSection();

  if (checkpoint == null) {
    return null;
  }

  const format = new Intl.NumberFormat(i18n.language);
  const facts = [
    {
      label: t("scanPage.checkpoint.elapsed"),
      value: t("scanPage.checkpoint.elapsedValue", {
        seconds: formatElapsedSeconds(i18n.language, checkpoint.elapsedMs),
      }),
    },
    {
      label: t("scanPage.checkpoint.pagesScanned"),
      value: format.format(checkpoint.pagesScanned),
    },
    {
      label: t("scanPage.checkpoint.truncated"),
      value: checkpoint.truncated ? t("scanPage.checkpoint.yes") : t("scanPage.checkpoint.no"),
    },
  ];

  return (
    <section className={root()}>
      <Heading level={2} size={4} weight="medium" highContrast>
        {t("scanPage.checkpoint.title")}
      </Heading>
      <div className={rows()}>
        {facts.map(fact => (
          <div key={fact.label} className={row()}>
            <div className={rowBody()}>
              <Text size={1} color="faint">{fact.label}</Text>
              <Text size={2} highContrast>{fact.value}</Text>
            </div>
          </div>
        ))}
        <CheckpointList
          label={t("scanPage.checkpoint.visited")}
          values={checkpoint.visitedTemplates}
          emptyLabel={t("scanPage.checkpoint.empty")}
          listClassName={list()}
          itemClassName={item()}
          rowClassName={row()}
          rowBodyClassName={rowBody()}
        />
        <CheckpointList
          label={t("scanPage.checkpoint.queue")}
          values={checkpoint.queue.map(entry => entry.url)}
          emptyLabel={t("scanPage.checkpoint.empty")}
          listClassName={list()}
          itemClassName={item()}
          rowClassName={row()}
          rowBodyClassName={rowBody()}
        />
        <CheckpointList
          label={t("scanPage.checkpoint.hosts")}
          values={checkpoint.coveredHosts}
          emptyLabel={t("scanPage.checkpoint.empty")}
          listClassName={list()}
          itemClassName={item()}
          rowClassName={row()}
          rowBodyClassName={rowBody()}
        />
      </div>
    </section>
  );
}

function CheckpointList({
  label,
  values,
  emptyLabel,
  listClassName,
  itemClassName,
  rowClassName,
  rowBodyClassName,
}: {
  label: string;
  values: readonly string[];
  emptyLabel: string;
  listClassName: string;
  itemClassName: string;
  rowClassName: string;
  rowBodyClassName: string;
}) {
  let body: ReactNode;
  if (values.length === 0) {
    body = <Text size={2} highContrast>{emptyLabel}</Text>;
  } else {
    body = (
      <ul className={listClassName}>
        {values.map(value => (
          <li key={value} className={itemClassName}>{value}</li>
        ))}
      </ul>
    );
  }

  return (
    <div className={rowClassName}>
      <div className={rowBodyClassName}>
        <Text size={1} color="faint">{label}</Text>
        {body}
      </div>
    </div>
  );
}
