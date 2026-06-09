// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
