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
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { SharedDocumentListItem_document$key } from "#/__generated__/core/SharedDocumentListItem_document.graphql";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

const NS = "documents";

const documentFragment = graphql`
  fragment SharedDocumentListItem_document on EmployeePortalDocument @throwOnFieldError {
    id
    title
    documentType
  }
`;

export interface SharedDocumentListItemProps {
  documentKey: SharedDocumentListItem_document$key;
}

export function SharedDocumentListItem({ documentKey }: SharedDocumentListItemProps) {
  const { t } = useTranslation(NS);
  const { t: tApp } = useTranslation();
  const { employeePortalId } = useEmployeePortalRoute();
  const document = useFragment(documentFragment, documentKey);
  const title = document.title;

  return (
    <ListItem className="relative cursor-pointer hover:bg-sand-2">
      <ListItemContent>
        <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
          {title}
        </Text>
        <Text size={1} color="gold" className="truncate">
          {tApp(`documents.types.${document.documentType.toLowerCase()}`)}
        </Text>
      </ListItemContent>
      <CaretRightIcon aria-hidden className="size-4 text-sand-11" />
      <Link
        to={`/${employeePortalId}/documents/${document.id}`}
        aria-label={t("shared.view", { title })}
        className="absolute inset-0"
      />
    </ListItem>
  );
}
