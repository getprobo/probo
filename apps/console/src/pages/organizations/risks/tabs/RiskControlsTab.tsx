import { useTranslate } from "@probo/i18n";
import { Badge, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { graphql, useRefetchableFragment } from "react-relay";
import { useOutletContext } from "react-router";

import type { RiskControlsTabFragment$key } from "#/__generated__/core/RiskControlsTabFragment.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const controlsFragment = graphql`
  fragment RiskControlsTabFragment on Risk
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "ControlOrder", defaultValue: null }
    filter: { type: "ControlFilter", defaultValue: null }
  )
  @refetchable(queryName: "RiskControlsTabControlsQuery") {
    id
    controls(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "RiskControlsTab_controls") {
      edges {
        node {
          id
          sectionTitle
          name
          framework {
            id
            name
          }
        }
      }
    }
  }
`;
export default function RiskControlsTab() {
  const { risk } = useOutletContext<{
    risk: RiskControlsTabFragment$key & { id: string };
  }>();
  const { __ } = useTranslate();
  const [data, refetch] = useRefetchableFragment(controlsFragment, risk);
  const controls = data.controls.edges.map(edge => edge.node);
  const organizationId = useOrganizationId();

  return (
    <SortableTable refetch={refetch}>
      <Thead>
        <Tr>
          <SortableTh field="SECTION_TITLE">{__("Reference")}</SortableTh>
          <Th>{__("Name")}</Th>
        </Tr>
      </Thead>
      <Tbody>
        {controls.length === 0 && (
          <Tr>
            <Td colSpan={2} className="text-center text-txt-secondary">
              {__("No controls linked")}
            </Td>
          </Tr>
        )}
        {controls.map(control => (
          <Tr
            key={control.id}
            to={`/organizations/${organizationId}/frameworks/${control.framework.id}/controls/${control.id}`}
          >
            <Td>
              <span className="inline-flex gap-2 items-center">
                {control.framework.name}
                {" "}
                <Badge size="md">{control.sectionTitle}</Badge>
              </span>
            </Td>
            <Td>{control.name}</Td>
          </Tr>
        ))}
      </Tbody>
    </SortableTable>
  );
}
