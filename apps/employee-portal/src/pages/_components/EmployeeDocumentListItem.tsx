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

import { CaretRightIcon } from "@phosphor-icons/react";
import { parseDate } from "@probo/helpers";
import { button } from "@probo/ui/src/v2/Button/variants";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { EmployeeDocumentListItem_document$key } from "#/__generated__/core/EmployeeDocumentListItem_document.graphql";

import { employeeDocumentListItem } from "./variants";

const employeeDocumentListItemFragment = graphql`
  fragment EmployeeDocumentListItem_document on EmployeeDocument @throwOnFieldError {
    title
    updatedAt
    lastVersion: versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
      edges {
        node {
          documentType
          classification
        }
      }
    }
  }
`;

type EmployeeDocumentListItemProps = {
  documentKey: EmployeeDocumentListItem_document$key;
  to: string;
  // Optional status chrome (e.g. approved / rejected on history rows).
  badge?: ReactNode;
} & (
  | { trailing: "action"; actionLabel: string }
  | { trailing: "chevron" }
);

export function EmployeeDocumentListItem(props: EmployeeDocumentListItemProps) {
  const { documentKey, to, trailing, badge } = props;
  const { t, i18n } = useTranslation();
  const slots = employeeDocumentListItem();
  const document = useFragment(employeeDocumentListItemFragment, documentKey);
  const lastVersion = document.lastVersion.edges[0]?.node;
  const updatedAt = new Intl.DateTimeFormat(i18n.language, {
    dateStyle: "medium",
  }).format(parseDate(document.updatedAt));

  return (
    <TableRow align="center">
      <TableCell className={slots.cell()} style={{ padding: 0 }}>
        <Link to={to} className={slots.link()}>
          <Text size={2} weight="medium" highContrast className={slots.title()}>
            {document.title}
          </Text>
          <div className={slots.meta()}>
            <span className={slots.timestamp()}>
              <Text size={1} color="current" className={slots.timestampLabel()}>
                {t("documents.lastUpdated")}
              </Text>
              <Text size={1} color="current" className={slots.timestampValue()}>
                {updatedAt}
              </Text>
            </span>
            {lastVersion == null
              ? null
              : (
                  <Text size={1} color="current" className={slots.chip()}>
                    {t(`documents.classifications.${lastVersion.classification.toLowerCase()}`)}
                  </Text>
                )}
            {lastVersion == null
              ? null
              : (
                  <Text size={1} color="current" className={slots.chip()}>
                    {t(`documents.types.${lastVersion.documentType.toLowerCase()}`)}
                  </Text>
                )}
            {badge === undefined ? null : badge}
            {trailing === "action"
              ? (
                  <span
                    className={button({
                      size: 2,
                      variant: "soft",
                      color: "neutral",
                      highContrast: true,
                    })}
                  >
                    {props.actionLabel}
                  </span>
                )
              : <CaretRightIcon className={slots.chevron()} />}
          </div>
        </Link>
      </TableCell>
    </TableRow>
  );
}
