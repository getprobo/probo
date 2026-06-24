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

import { FrameworkLogo } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { FrameworkBadgeFragment$key } from "./__generated__/FrameworkBadgeFragment.graphql";

const frameworkFragment = graphql`
  fragment FrameworkBadgeFragment on Framework {
    # eslint-disable-next-line relay/unused-fields
    id
    name
    lightLogo {
      downloadUrl
    }
    darkLogo {
      downloadUrl
    }
  }
`;

export function FrameworkBadge(props: { framework: FrameworkBadgeFragment$key }) {
  const framework = useFragment(frameworkFragment, props.framework);

  return (
    <div className="flex flex-col gap-2 items-center w-19">
      <FrameworkLogo
        className="size-19"
        lightLogoURL={framework.lightLogo?.downloadUrl}
        darkLogoURL={framework.darkLogo?.downloadUrl}
        name={framework.name}
      />
      <div className="txt-primary text-xs max-w-19 min-w-0 text-center">
        {framework.name}
      </div>
    </div>
  );
}
