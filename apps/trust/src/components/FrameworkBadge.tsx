import { FrameworkLogo } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { FrameworkBadgeFragment$key } from "./__generated__/FrameworkBadgeFragment.graphql";

const frameworkFragment = graphql`
  fragment FrameworkBadgeFragment on Framework {
    id
    name
    lightLogoURL
    darkLogoURL
  }
`;

export function FrameworkBadge(props: { framework: FrameworkBadgeFragment$key }) {
  const framework = useFragment(frameworkFragment, props.framework);

  return (
    <div className="flex flex-col gap-2 items-center w-19">
      <FrameworkLogo
        className="size-19"
        lightLogoURL={framework.lightLogoURL}
        darkLogoURL={framework.darkLogoURL}
        name={framework.name}
      />
      <div className="txt-primary text-sm max-w-19 overflow-hidden min-w-0 whitespace-nowrap text-ellipsis">
        {framework.name}
      </div>
    </div>
  );
}
