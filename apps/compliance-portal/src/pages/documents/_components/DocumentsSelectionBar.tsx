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

import { LockSimpleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import type { DocumentSelectionEntry } from "../_lib/DocumentSelectionContext";
import { useDocumentSelection } from "../_lib/DocumentSelectionContext";
import { useBulkRequestAccess } from "../_lib/useBulkRequestAccess";
import { selectionBar } from "../variants";

interface DocumentsSelectionBarProps {
  // Every selectable row on the page, used to resolve the current selection into
  // concrete entries (kind + lock state) and to power "Select all".
  entries: DocumentSelectionEntry[];
}

// Bottom action bar shown while rows are selected: the selection count, clear /
// select-all shortcuts, and a bulk "Request Access" that acts only on the
// selected rows still locked.
export function DocumentsSelectionBar({ entries }: DocumentsSelectionBarProps) {
  const { t } = useTranslation("documents");
  const { selectedIds, selectAll, clear } = useDocumentSelection();
  const { requestAccess, isRequesting } = useBulkRequestAccess(clear);

  if (selectedIds.size === 0) {
    return null;
  }

  const lockedSelected = entries.filter(entry => selectedIds.has(entry.id) && entry.locked);
  const lockedCount = lockedSelected.length;

  const handleRequestAccess = () => {
    if (lockedCount === 0) {
      return;
    }
    requestAccess(lockedSelected.map(entry => ({ id: entry.id, kind: entry.kind })));
  };

  const { bar, inner, actions } = selectionBar();

  return (
    <div className={bar()}>
      <div className={inner()}>
        <Text size={2} weight="medium" color="neutral" highContrast>
          {t("selection.count", { count: selectedIds.size })}
        </Text>
        <div className={actions()}>
          <Button variant="ghost" color="neutral" onClick={clear}>
            {t("selection.clear")}
          </Button>
          <Button
            variant="ghost"
            color="neutral"
            onClick={() => selectAll(entries.filter(entry => entry.locked).map(entry => entry.id))}
          >
            {t("selection.selectAll")}
          </Button>
          <Button
            variant="solid"
            color="neutral"
            highContrast
            iconStart={<LockSimpleIcon />}
            loading={isRequesting}
            disabled={lockedCount === 0}
            onClick={handleRequestAccess}
          >
            {t("selection.requestAccess", { count: lockedCount })}
          </Button>
        </div>
      </div>
    </div>
  );
}
