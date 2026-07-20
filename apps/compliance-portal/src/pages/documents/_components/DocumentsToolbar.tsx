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

import { Tabs } from "@probo/ui/src/v2/Tabs/Tabs";
import { TabsIndicator } from "@probo/ui/src/v2/Tabs/TabsIndicator";
import { TabsList } from "@probo/ui/src/v2/Tabs/TabsList";
import { TabsTab } from "@probo/ui/src/v2/Tabs/TabsTab";
import { useTranslation } from "react-i18next";

import type { DocumentTab } from "../_lib/useDocumentTab";
import { DOCUMENT_TABS, useDocumentTab } from "../_lib/useDocumentTab";

// Access filter for the documents page: All / Public / Private tabs. Writes the
// active tab to the URL; the page reacts and refetches the matching slice.
export function DocumentsToolbar() {
  const { t } = useTranslation("documents");
  const { tab, setTab } = useDocumentTab();

  return (
    <Tabs value={tab} onValueChange={value => setTab(value as DocumentTab)}>
      <TabsList className="max-w-full overflow-x-auto">
        {DOCUMENT_TABS.map(value => (
          <TabsTab key={value} value={value}>
            {t(`tabs.${value}`)}
          </TabsTab>
        ))}
        <TabsIndicator />
      </TabsList>
    </Tabs>
  );
}
