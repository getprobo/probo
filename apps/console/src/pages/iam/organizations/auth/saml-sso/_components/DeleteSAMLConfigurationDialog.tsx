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
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { DeleteSAMLConfigurationDialog_deleteMutation } from "#/__generated__/iam/DeleteSAMLConfigurationDialog_deleteMutation.graphql";
import type { DeleteSAMLConfigurationDialog_samlConfiguration$key } from "#/__generated__/iam/DeleteSAMLConfigurationDialog_samlConfiguration.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const fragment = graphql`
  fragment DeleteSAMLConfigurationDialog_samlConfiguration on SAMLConfiguration {
    id
    emailDomain
  }
`;

const deleteMutation = graphql`
  mutation DeleteSAMLConfigurationDialog_deleteMutation(
    $input: DeleteSAMLConfigurationInput!
    $connections: [ID!]!
  ) {
    deleteSAMLConfiguration(input: $input) {
      deletedSamlConfigurationId @deleteEdge(connections: $connections)
    }
  }
`;

interface DeleteSAMLConfigurationDialogProps {
  samlConfigurationKey: DeleteSAMLConfigurationDialog_samlConfiguration$key;
  onDeleted: () => void;
}

export function DeleteSAMLConfigurationDialog({
  samlConfigurationKey,
  onDeleted,
}: DeleteSAMLConfigurationDialogProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const config = useFragment(fragment, samlConfigurationKey);
  const [open, setOpen] = useState(false);
  const [deleteSAMLConfiguration, isDeleting]
    = useMutation<DeleteSAMLConfigurationDialog_deleteMutation>(
      deleteMutation,
      {
        successMessage: t("samlConfigurationList.messages.deleted"),
        errorToast: t("samlConfigurationList.errors.delete"),
      },
    );

  function handleDelete() {
    void deleteSAMLConfiguration({
      variables: {
        input: {
          organizationId,
          samlConfigurationId: config.id,
        },
        connections: [
          ConnectionHandler.getConnectionID(
            organizationId,
            "SAMLConfigurationListFragment_samlConfigurations",
          ),
        ],
      },
    }).then(
      () => {
        setOpen(false);
        onDeleted();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        render={(
          <IconButton
            size={1}
            variant="surface"
            color="red"
            aria-label={t("samlConfigurationList.actions.delete")}
          >
            <TrashIcon />
          </IconButton>
        )}
      />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("samlConfigurationList.delete.title")}</DialogTitle>
          <DialogDescription>
            {t("samlConfigurationList.delete.description", { domain: config.emailDomain })}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("samlConfigurationList.actions.cancel")}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color="red"
            loading={isDeleting}
            iconStart={<TrashIcon />}
            onClick={handleDelete}
          >
            {t("samlConfigurationList.actions.delete")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
