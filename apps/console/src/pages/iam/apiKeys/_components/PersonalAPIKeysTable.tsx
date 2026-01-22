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
