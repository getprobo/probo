// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { FrameworkLogo } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { FrameworkBadgeFragment$key } from "./__generated__/FrameworkBadgeFragment.graphql";

const frameworkFragment = graphql`
  fragment FrameworkBadgeFragment on Framework {
    # eslint-disable-next-line relay/unused-fields
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
      <div className="txt-primary text-xs max-w-19 min-w-0 text-center">
        {framework.name}
      </div>
    </div>
  );
}
