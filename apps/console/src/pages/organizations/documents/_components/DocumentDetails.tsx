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

import { ScrollArea } from "@base-ui/react/scroll-area";
import { CaretLeftIcon, SidebarSimpleIcon } from "@phosphor-icons/react";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentDetails_document$key } from "#/__generated__/core/DocumentDetails_document.graphql";

import { useDocumentModes } from "../_lib/useDocumentModes";
import { documentDetailsPage } from "../variants";

import { DocumentApprovalsSection } from "./DocumentApprovalsSection";
import { DocumentControlsSection } from "./DocumentControlsSection";
import { DocumentEditor } from "./DocumentEditor";
import { DocumentModeSelect } from "./DocumentModeSelect";
import { DocumentPropertiesSection } from "./DocumentPropertiesSection";
import { DocumentSignatoriesSection } from "./DocumentSignatoriesSection";
import { DocumentTitleField } from "./DocumentTitleField";

const documentDetailsFragment = graphql`
  fragment DocumentDetails_document on Document {
    ...DocumentTitleField_document
    ...DocumentEditor_document
    ...useDocumentModes_document
    ...DocumentPropertiesSection_document
    ...DocumentControlsSection_document
    ...DocumentApprovalsSection_document
    ...DocumentSignatoriesSection_document
  }
`;

interface DocumentDetailsProps {
  documentKey: DocumentDetails_document$key;
  organizationId: string;
  search: string;
}

export function DocumentDetails({
  documentKey,
  organizationId,
  search,
}: DocumentDetailsProps) {
  const document = useFragment(documentDetailsFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const { mode, selectMode, modes } = useDocumentModes(document);
  const [cardsOpen, setCardsOpen] = useState(true);
  const {
    root, main, backRow, titleRow, title, sideFrame, sideClip, side, viewport, cards, scrollbar, thumb,
  } = documentDetailsPage({ cards: cardsOpen });

  return (
    <div className={root()}>
      <div className={main()}>
        <div className={backRow()}>
          <Link
            to={{
              pathname: `/organizations/${organizationId}/governance/documents`,
              search,
            }}
            size={2}
            color="neutral"
            underline={false}
            iconStart={<CaretLeftIcon />}
          >
            {t("details.back")}
          </Link>
          <IconButton
            size={1}
            variant="ghost"
            color="neutral"
            aria-expanded={cardsOpen}
            aria-label={t(cardsOpen ? "details.cards.hide" : "details.cards.show")}
            onClick={() => setCardsOpen(open => !open)}
          >
            <SidebarSimpleIcon className="size-5! -scale-x-100" />
          </IconButton>
        </div>
        <div className={titleRow()}>
          <div className={title()}>
            <DocumentTitleField documentKey={document} />
          </div>
          <DocumentModeSelect
            mode={mode}
            modes={modes}
            onModeChange={selectMode}
          />
        </div>
        <DocumentEditor documentKey={document} mode={mode} />
      </div>
      <div className={sideFrame()} inert={cardsOpen ? undefined : true}>
        <div className={sideClip()}>
          <ScrollArea.Root className={side()}>
            <ScrollArea.Viewport className={viewport()}>
              <ScrollArea.Content className={cards()}>
                <DocumentPropertiesSection documentKey={document} />
                <DocumentApprovalsSection documentKey={document} />
                <DocumentSignatoriesSection documentKey={document} />
                <DocumentControlsSection documentKey={document} />
              </ScrollArea.Content>
            </ScrollArea.Viewport>
            <ScrollArea.Scrollbar className={scrollbar()} orientation="vertical">
              <ScrollArea.Thumb className={thumb()} />
            </ScrollArea.Scrollbar>
          </ScrollArea.Root>
        </div>
      </div>
    </div>
  );
}
