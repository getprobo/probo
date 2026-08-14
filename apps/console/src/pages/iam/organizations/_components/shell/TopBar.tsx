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

import { Logo } from "@probo/ui";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { TopBar_organization$key } from "#/__generated__/iam/TopBar_organization.graphql";

import { OrganizationSwitcher } from "./OrganizationSwitcher";
import { topBar } from "./variants";
import { ViewerMembershipMenu } from "./ViewerMembershipMenu";

const topBarFragment = graphql`
  fragment TopBar_organization on Organization {
    ...OrganizationSwitcher_organization
    ...ViewerMembershipMenu_organization
  }
`;

export interface TopBarProps {
  organizationKey: TopBar_organization$key;
}

export function TopBar({ organizationKey }: TopBarProps) {
  const organization = useFragment(topBarFragment, organizationKey);

  const slots = topBar();

  return (
    <header className={slots.bar()}>
      <Link to="/" className={slots.brand()}>
        <Logo className={slots.logo()} />
      </Link>
      <span className={slots.separator()} aria-hidden>/</span>
      <OrganizationSwitcher organizationKey={organization} />
      <div className={slots.trailing()}>
        <ViewerMembershipMenu organizationKey={organization} />
      </div>
    </header>
  );
}
