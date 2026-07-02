// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Table, Tbody, Th, Thead, Tr } from "@probo/ui";

import type { PersonalAPIKeyListFragment$data } from "#/__generated__/iam/PersonalAPIKeyListFragment.graphql";

import { PersonalAPIKeyRow } from "./PersonalAPIKeyRow";

export function PersonalAPIKeysTable(props: {
  edges: PersonalAPIKeyListFragment$data["personalAPIKeys"]["edges"];
  connectionId: string;
}) {
  const { edges, connectionId } = props;
  const { __ } = useTranslate();

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>{__("Name")}</Th>
          <Th>{__("Last used")}</Th>
          <Th>{__("Created")}</Th>
          <Th>{__("Expires")}</Th>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {edges.map(({ node }) => (
          <PersonalAPIKeyRow
            key={node.id}
            fKey={node}
            connectionId={connectionId}
          />
        ))}
      </Tbody>
    </Table>
  );
}
