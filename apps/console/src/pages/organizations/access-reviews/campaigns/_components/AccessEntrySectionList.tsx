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

import { memo } from "react";
import { graphql, readInlineData } from "relay-runtime";

import type { AccessEntrySectionList_entry$key } from "#/__generated__/core/AccessEntrySectionList_entry.graphql";

import { AccessEntryListItem } from "./AccessEntryListItem";
import { accessEntryList } from "./variants";

const accessEntrySectionListFragment = graphql`
  fragment AccessEntrySectionList_entry on AccessReviewEntry @inline {
    id
    ...AccessEntryListItem_entry
  }
`;

interface AccessEntrySectionListProps {
  entryKeys: ReadonlyArray<AccessEntrySectionList_entry$key>;
  isPendingActions: boolean;
  selectedIds: ReadonlySet<string>;
  onSelectedChange: (
    entryId: string,
    event: { shiftKey: boolean },
  ) => void;
}

function AccessEntrySectionListComponent({
  entryKeys,
  isPendingActions,
  selectedIds,
  onSelectedChange,
}: AccessEntrySectionListProps) {
  const { root } = accessEntryList();

  return (
    <ul className={root()}>
      {entryKeys.map((entryKey) => {
        const entry = readInlineData(
          accessEntrySectionListFragment,
          entryKey,
        );

        return (
          <AccessEntryListItem
            key={entry.id}
            entryKey={entry}
            isPendingActions={isPendingActions}
            selected={selectedIds.has(entry.id)}
            onSelectedChange={onSelectedChange}
          />
        );
      })}
    </ul>
  );
}

export const AccessEntrySectionList = memo(AccessEntrySectionListComponent);
