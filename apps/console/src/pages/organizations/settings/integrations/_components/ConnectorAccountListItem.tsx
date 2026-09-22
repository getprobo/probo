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

import { graphql, useFragment } from "react-relay";

import type { ConnectorAccountListItem_account$key } from "#/__generated__/core/ConnectorAccountListItem_account.graphql";

import { integrationSection } from "../variants";

// connectionStatus is omitted: the field probes the connector, not this row,
// so every account would show the same live answer.
const connectorAccountListItemFragment = graphql`
  fragment ConnectorAccountListItem_account on ConnectorAccount {
    name
    externalAccountId
  }
`;

interface ConnectorAccountListItemProps {
  accountKey: ConnectorAccountListItem_account$key;
}

export function ConnectorAccountListItem({
  accountKey,
}: ConnectorAccountListItemProps) {
  const account = useFragment(connectorAccountListItemFragment, accountKey);
  const { item, content, name, description } = integrationSection();

  return (
    <li className={item()}>
      <div className={content()}>
        <span className={name()}>{account.name}</span>
        <span className={description()}>{account.externalAccountId}</span>
      </div>
    </li>
  );
}
