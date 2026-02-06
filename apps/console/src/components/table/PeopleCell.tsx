import { Avatar } from "@probo/ui";

import type { PeopleGraphQuery } from "#/__generated__/core/PeopleGraphQuery.graphql";
import { GraphQLCell } from "#/components/table/GraphQLCell";
import { peopleQuery } from "#/hooks/graph/PeopleGraph";

type Props = {
  name: string;
  defaultValue?: { fullName: string; id: string };
  organizationId: string;
};

export function PeopleCell(props: Props) {
  return (
    <GraphQLCell<PeopleGraphQuery, { fullName: string }>
      name={props.name}
      query={peopleQuery}
      variables={{
        organizationId: props.organizationId,
        filter: { excludeContractEnded: true },
      }}
      items={data =>
        data.organization?.profiles?.edges.map(edge => edge.node) ?? []}
      itemRenderer={({ item }) => (
        <div className="flex gap-2 whitespace-nowrap items-center text-xs">
          <Avatar name={item.fullName} />
          {item.fullName}
        </div>
      )}
      defaultValue={props.defaultValue}
    />
  );
}
