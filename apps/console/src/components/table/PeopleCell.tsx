import { GraphQLCell } from "/components/table/GraphQLCell.tsx";
import type { PeopleGraphQuery } from "/hooks/graph/__generated__/PeopleGraphQuery.graphql.ts";
import { peopleQuery } from "/hooks/graph/PeopleGraph.ts";
import { Avatar } from "@probo/ui";

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
      items={(data) =>
        data.organization?.peoples?.edges.map((edge) => edge.node) ?? []
      }
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
