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

import { parseDate } from "@probo/helpers";
import {
  Avatar,
  DropdownItem,
  IconCheckmark1,
  IconClock,
  IconLock,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { MembershipsDropdownMenuItem_organizationFragment$key } from "#/__generated__/iam/MembershipsDropdownMenuItem_organizationFragment.graphql";
import type { MembershipsDropdownMenuItemFragment$key } from "#/__generated__/iam/MembershipsDropdownMenuItemFragment.graphql";

const fragment = graphql`
  fragment MembershipsDropdownMenuItemFragment on Membership {
    id
    lastSession {
      id
      expiresAt
    }
  }
`;

const organizationFragment = graphql`
  fragment MembershipsDropdownMenuItem_organizationFragment on Organization {
    id
    name
    logo {
      downloadUrl
    }
  }
`;

export function MembershipsDropdownMenuItem(props: {
  fKey: MembershipsDropdownMenuItemFragment$key;
  organizationFragmentRef: MembershipsDropdownMenuItem_organizationFragment$key;
}) {
  const { fKey, organizationFragmentRef } = props;

  const { id, lastSession }
    = useFragment<MembershipsDropdownMenuItemFragment$key>(fragment, fKey);
  const organization
    = useFragment<MembershipsDropdownMenuItem_organizationFragment$key>(organizationFragment, organizationFragmentRef);

  const isAssuming = !!lastSession;
  const isExpired
    = lastSession && parseDate(lastSession.expiresAt) < new Date();

  return (
    <DropdownItem key={id} asChild>
      <Link to={`/organizations/${organization.id}`}>
        <Avatar name={organization.name} src={organization.logo?.downloadUrl} />
        <span className="flex-1">{organization.name}</span>
        {isAssuming && (
          <IconCheckmark1 size={16} className="text-green-600" />
        )}
        {isExpired && <IconClock size={16} className="text-orange-600" />}
        {!lastSession && <IconLock size={16} className="text-gray-400" />}
      </Link>
    </DropdownItem>
  );
}
