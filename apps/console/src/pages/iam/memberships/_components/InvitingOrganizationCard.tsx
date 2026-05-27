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

import { useTranslate } from "@probo/i18n";
import { Badge, Card, IconMail } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { InvitingOrganizationCardFragment$key } from "#/__generated__/iam/InvitingOrganizationCardFragment.graphql";

const fragment = graphql`
  fragment InvitingOrganizationCardFragment on Organization {
    name
  }
`;

interface InvitingOrganizationCardProps {
  fKey: InvitingOrganizationCardFragment$key;
}

export function InvitingOrganizationCard(props: InvitingOrganizationCardProps) {
  const { fKey } = props;
  const { __ } = useTranslate();

  const organization = useFragment<InvitingOrganizationCardFragment$key>(
    fragment,
    fKey,
  );

  return (
    <Card padded className="w-full">
      <div className="flex items-center justify-between">
        <h2 className="font-semibold text-xl">{organization.name}</h2>
        <Badge variant="neutral" className="flex items-center gap-1">
          <IconMail size={14} />
          {__("Check your email")}
        </Badge>
      </div>
    </Card>
  );
}
