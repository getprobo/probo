import { graphql, useFragment } from "react-relay";
import { type ControlList_List$key } from "./__generated__/ControlList_List.graphql";
import { NavLink, useParams } from "react-router";
import { cn } from "@/lib/utils";

const maxControlNameLength = 80;

export const controlListFragment = graphql`
  fragment ControlList_List on Framework {
    controls(first: 100, orderBy: { field: CREATED_AT, direction: ASC })
      @connection(key: "FrameworkView_controls") {
      edges {
        node {
          id
          referenceId
          name
        }
      }
    }
  }
`;

interface ControlListProps {
  fragmentKey: ControlList_List$key;
  className?: string;
}

export function ControlList(props: ControlListProps) {
  const { organizationId, frameworkId, controlId } = useParams<{
    organizationId: string;
    frameworkId: string;
    controlId?: string;
  }>();
  const { className, fragmentKey } = props;
  const { controls } = useFragment<ControlList_List$key>(
    controlListFragment,
    fragmentKey,
  );

  if (controls.edges.length === 0) {
    return "No controls available for this framework";
  }

  return (
    <aside
      className={cn(
        "shrink-0 w-70 border-r h-full overflow-x-hidden overflow-y-auto [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-level-0 [&::-webkit-scrollbar-thumb]:bg-highlight-bg [&::-webkit-scrollbar-thumb]:rounded-xs",
        className,
      )}
    >
      {controls.edges.map(({ node: control }, i) => (
        <NavLink
          key={control.id}
          to={`/organizations/${organizationId}/frameworks/${frameworkId}/controls/${control.id}`}
          className={({ isActive }) =>
            cn(
              "block pl-8 pr-4 py-4 border-b hover:bg-h-subtle-bg",
              (isActive || (i === 0 && !controlId)) && "bg-subtle-bg",
            )
          }
        >
          {({ isActive }) => {
            return (
              <>
                <div
                  className={cn(
                    "inline-block font-mono text-sm px-1 py-0.25 rounded-sm bg-highlight-bg border font-semibold",
                    isActive && "font-bold bg-active-bg border-mid-b",
                  )}
                >
                  {control.referenceId}
                </div>
                <div
                  className={cn(
                    "text-sm leading-none break-words mt-2",
                    isActive && "font-medium",
                  )}
                >
                  {control.name.length > maxControlNameLength
                    ? `${control.name.slice(0, maxControlNameLength - 2)}…`
                    : control.name}
                </div>
              </>
            );
          }}
        </NavLink>
      ))}
    </aside>
  );
}
