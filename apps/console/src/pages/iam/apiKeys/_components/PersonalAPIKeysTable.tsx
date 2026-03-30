// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
