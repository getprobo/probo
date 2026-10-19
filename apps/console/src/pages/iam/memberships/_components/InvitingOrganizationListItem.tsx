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

import { formatError, type GraphQLError } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { InvitingOrganizationListItem_organization$key } from "#/__generated__/iam/InvitingOrganizationListItem_organization.graphql";
import type { InvitingOrganizationListItemAcceptMutation } from "#/__generated__/iam/InvitingOrganizationListItemAcceptMutation.graphql";

const invitingOrganizationListItemFragment = graphql`
  fragment InvitingOrganizationListItem_organization on Organization {
    id
    name
  }
`;

const acceptInvitationMutation = graphql`
  mutation InvitingOrganizationListItemAcceptMutation(
    $input: AcceptInvitationInput!
  ) {
    acceptInvitation(input: $input) {
      profile {
        id
      }
    }
  }
`;

interface InvitingOrganizationListItemProps {
  organizationKey: InvitingOrganizationListItem_organization$key;
  onAccepted: () => void;
}

export function InvitingOrganizationListItem({
  organizationKey,
  onAccepted,
}: InvitingOrganizationListItemProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const organization = useFragment(
    invitingOrganizationListItemFragment,
    organizationKey,
  );
  const [acceptInvitation, isAccepting]
    = useMutation<InvitingOrganizationListItemAcceptMutation>(acceptInvitationMutation);

  const handleAccept = () => {
    acceptInvitation({
      variables: {
        input: { organizationId: organization.id },
      },
      onCompleted: (_response, errors: GraphQLError[] | null) => {
        if (errors) {
          for (const err of errors) {
            if (err.extensions?.code === "ACTIVATION_NOT_REQUIRED") {
              onAccepted();
              return;
            }
          }

          toast.add({
            title: t("invitingOrganizationCard.errors.joinFailed"),
            description: formatError(t("invitingOrganizationCard.errors.joinFailed"), errors),
            type: "error",
          });
          return;
        }

        onAccepted();
      },
      onError: (e) => {
        toast.add({
          title: t("invitingOrganizationCard.errors.joinFailed"),
          description: e.message,
          type: "error",
        });
      },
    });
  };

  return (
    <ListItem>
      <ListItemContent>
        <Text size={2} weight="medium" color="neutral" highContrast>
          {organization.name}
        </Text>
      </ListItemContent>
      <Button
        type="button"
        size={1}
        variant="soft"
        color="neutral"
        loading={isAccepting}
        onClick={handleAccept}
      >
        {t("invitingOrganizationCard.actions.join")}
      </Button>
    </ListItem>
  );
}
