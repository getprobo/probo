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
import { graphql, useFragment } from "react-relay";

import type { AddableConnectorListItem_connector$key } from "#/__generated__/core/AddableConnectorListItem_connector.graphql";

import { ConnectorAccountPages } from "./AddableConnectorAccounts";
import {
  type AddableConnectorCard,
  ResolvedAddableConnectorCard,
} from "./ResolvedAddableConnectorCard";

const fragment = graphql`
  fragment AddableConnectorListItem_connector on Connector @relay(plural: true) {
    ...ResolvedAddableConnectorCard_connector
    ...AddableConnectorAccounts_connector
  }
`;

interface AddableConnectorListItemProps {
  connectorKeys: AddableConnectorListItem_connector$key;
  normalizedSearch: string;
  selectedProviders: ReadonlySet<string>;
  onSelectedChange: (provider: string, selected: boolean) => void;
  children: (card: AddableConnectorCard | null) => ReactNode;
}

export function AddableConnectorListItem({
  connectorKeys,
  normalizedSearch,
  selectedProviders,
  onSelectedChange,
  children,
}: AddableConnectorListItemProps) {
  const connectors = useFragment(fragment, connectorKeys);

  return (
    <ConnectorAccountPages connectorKeys={connectors}>
      {pages => (
        <ResolvedAddableConnectorCard
          pages={pages}
          connectorKeys={connectors}
          normalizedSearch={normalizedSearch}
          selectedProviders={selectedProviders}
          onSelectedChange={onSelectedChange}
        >
          {children}
        </ResolvedAddableConnectorCard>
      )}
    </ConnectorAccountPages>
  );
}
