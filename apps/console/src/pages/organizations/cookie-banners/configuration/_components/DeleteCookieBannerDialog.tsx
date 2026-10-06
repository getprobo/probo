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

import { TrashIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { useTranslation } from "react-i18next";
import { useRelayEnvironment } from "react-relay";
import { useNavigate } from "react-router";
import { ConnectionHandler, fetchQuery, graphql } from "relay-runtime";

import type { DeleteCookieBannerDialogMutation } from "#/__generated__/core/DeleteCookieBannerDialogMutation.graphql";
import type { DeleteCookieBannerDialogRemainingQuery } from "#/__generated__/core/DeleteCookieBannerDialogRemainingQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  cookieBannerConfigurePath,
  cookieBannersNewPath,
} from "../../_lib/cookieBannerPaths";

const deleteMutation = graphql`
  mutation DeleteCookieBannerDialogMutation(
    $input: DeleteCookieBannerInput!
    $connections: [ID!]!
  ) {
    deleteCookieBanner(input: $input) {
      deletedCookieBannerId @deleteEdge(connections: $connections)
    }
  }
`;

const remainingBannersQuery = graphql`
  query DeleteCookieBannerDialogRemainingQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        cookieBanners(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
            }
          }
        }
      }
    }
  }
`;

interface DeleteCookieBannerDialogProps {
  cookieBannerId: string;
  name: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DeleteCookieBannerDialog({
  cookieBannerId,
  name,
  open,
  onOpenChange,
}: DeleteCookieBannerDialogProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const navigate = useNavigate();
  const environment = useRelayEnvironment();
  const organizationId = useOrganizationId();
  const [deleteCookieBanner, isDeleting] = useMutation<DeleteCookieBannerDialogMutation>(
    deleteMutation,
    {
      successMessage: t("configLayout.messages.deleted"),
      errorToast: t("configLayout.errors.delete"),
    },
  );
  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    "CookieBannerSwitcherMenu_cookieBanners",
  );

  function handleDelete() {
    void deleteCookieBanner({
      variables: {
        input: { cookieBannerId },
        connections: [connectionId],
      },
    }).then(
      async () => {
        onOpenChange(false);
        const fallbackPath = cookieBannersNewPath(organizationId);
        try {
          const data = await fetchQuery<DeleteCookieBannerDialogRemainingQuery>(
            environment,
            remainingBannersQuery,
            { organizationId },
            { fetchPolicy: "network-only" },
          ).toPromise();
          const remainingId = data?.organization?.__typename === "Organization"
            ? data.organization.cookieBanners?.edges[0]?.node.id
            : undefined;
          void navigate(
            remainingId != null
              ? cookieBannerConfigurePath(organizationId, remainingId)
              : fallbackPath,
          );
        } catch {
          void navigate(fallbackPath);
        }
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("configLayout.dialogs.deleteTitle")}</DialogTitle>
          <DialogDescription>
            {t("configLayout.deleteConfirmation", { name })}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("configLayout.actions.cancel")}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color="red"
            iconStart={<TrashIcon />}
            loading={isDeleting}
            onClick={handleDelete}
          >
            {t("configLayout.actions.delete")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
