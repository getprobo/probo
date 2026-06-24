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

import { useTranslate } from "@probo/i18n";
import { IconMail } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { MembershipsDropdownInvitingItemFragment$key } from "#/__generated__/iam/MembershipsDropdownInvitingItemFragment.graphql";

const fragment = graphql`
  fragment MembershipsDropdownInvitingItemFragment on Organization {
    name
  }
`;

export function MembershipsDropdownInvitingItem(props: {
  fKey: MembershipsDropdownInvitingItemFragment$key;
}) {
  const { fKey } = props;
  const { __ } = useTranslate();

  const organization = useFragment<MembershipsDropdownInvitingItemFragment$key>(
    fragment,
    fKey,
  );

  return (
    <div
      className="text-txt-primary flex items-center gap-2 p-2 cursor-default"
      title={__("Check your email to accept the invitation")}
    >
      <div className="bg-border-mid text-txt-invert! rounded-full size-6 flex items-center justify-center flex-none">
        <IconMail size={16} />
      </div>
      <span className="flex-1">{organization.name}</span>
    </div>
  );
}
