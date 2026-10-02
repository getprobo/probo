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

import { type ReactNode } from "react";

import type { AddableConnectorListItem_connector$key } from "#/__generated__/core/AddableConnectorListItem_connector.graphql";

import { AddableConnectorListItem } from "./AddableConnectorListItem";
import type { AddableConnectorCard } from "./ResolvedAddableConnectorCard";

interface ResolveAddableConnectorGroupsProps {
  groups: readonly AddableConnectorListItem_connector$key[];
  index: number;
  cards: readonly AddableConnectorCard[];
  normalizedSearch: string;
  organizationId: string;
  connectionId: string;
  children: (cards: readonly AddableConnectorCard[]) => ReactNode;
}

export function ResolveAddableConnectorGroups({
  groups,
  index,
  cards,
  normalizedSearch,
  organizationId,
  connectionId,
  children,
}: ResolveAddableConnectorGroupsProps) {
  if (index >= groups.length) {
    return children(cards);
  }

  const group = groups[index];
  return (
    <AddableConnectorListItem
      connectorKeys={group}
      organizationId={organizationId}
      connectionId={connectionId}
      normalizedSearch={normalizedSearch}
    >
      {card => (
        <ResolveAddableConnectorGroups
          groups={groups}
          index={index + 1}
          cards={card == null ? cards : [...cards, card]}
          normalizedSearch={normalizedSearch}
          organizationId={organizationId}
          connectionId={connectionId}
        >
          {children}
        </ResolveAddableConnectorGroups>
      )}
    </AddableConnectorListItem>
  );
}
