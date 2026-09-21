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
import { graphql, useFragment } from "react-relay";
import { useNavigate } from "react-router";

import type { ConnectorDeleteDialog_connector$key } from "#/__generated__/core/ConnectorDeleteDialog_connector.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { integrationListPath } from "../_lib/integrationPath";
import { useDeleteConnector } from "../_lib/useDeleteConnector";

const connectorDeleteDialogFragment = graphql`
  fragment ConnectorDeleteDialog_connector on Connector {
    id
    displayName
  }
`;

interface ConnectorDeleteDialogProps {
  connectorKey: ConnectorDeleteDialog_connector$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ConnectorDeleteDialog({
  connectorKey,
  open,
  onOpenChange,
}: ConnectorDeleteDialogProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const connector = useFragment(connectorDeleteDialogFragment, connectorKey);
  const [deleteConnector, isDeleting] = useDeleteConnector();

  function handleDelete() {
    void deleteConnector(connector.id).then(
      () => {
        onOpenChange(false);
        void navigate(integrationListPath(organizationId));
      },
      () => {
        // The refusal names the feature still holding the credential, and
        // useMutation has already shown it. Staying open would hide it behind
        // the dialog.
        onOpenChange(false);
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("detailsPage.delete.title")}</DialogTitle>
          <DialogDescription>
            {t("detailsPage.delete.description", {
              name: connector.displayName,
            })}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("detailsPage.actions.cancel")}
              </Button>
            )}
          />
          <Button
            type="button"
            variant="solid"
            color="red"
            iconStart={<TrashIcon />}
            loading={isDeleting}
            onClick={handleDelete}
          >
            {t("detailsPage.delete.confirm")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
