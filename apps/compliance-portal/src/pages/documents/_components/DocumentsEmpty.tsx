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

import { FileTextIcon } from "@phosphor-icons/react";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/EmptyState/EmptyState";

import { useDocumentTab } from "../_lib/useDocumentTab";

// Empty state for the documents list. When a Public/Private tab is active it
// notes the filter; otherwise it states the trust center publishes no documents.
export function DocumentsEmpty() {
  const { t } = useTranslation("documents");
  const { tab } = useDocumentTab();
  const filtered = tab !== "all";

  return (
    <EmptyState
      icon={<FileTextIcon />}
      title={filtered ? t("empty.filteredTitle") : t("empty.title")}
      description={filtered ? t("empty.filteredDescription") : t("empty.description")}
    />
  );
}
