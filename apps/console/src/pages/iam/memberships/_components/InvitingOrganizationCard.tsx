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

import { Badge, Card, IconMail } from "@probo/ui";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();

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
          {t("invitingOrganizationCard.checkEmail")}
        </Badge>
      </div>
    </Card>
  );
}
