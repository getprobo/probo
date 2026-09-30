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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DisconnectSCIMProviderDialog_deleteMutation } from "#/__generated__/iam/DisconnectSCIMProviderDialog_deleteMutation.graphql";
import type { DisconnectSCIMProviderDialog_scimConfiguration$key } from "#/__generated__/iam/DisconnectSCIMProviderDialog_scimConfiguration.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import type { SCIMProviderCopy } from "../_lib/scimProvider";

const fragment = graphql`
  fragment DisconnectSCIMProviderDialog_scimConfiguration on SCIMConfiguration {
    id
  }
`;

const deleteMutation = graphql`
  mutation DisconnectSCIMProviderDialog_deleteMutation(
    $input: DeleteSCIMConfigurationInput!
  ) {
    deleteSCIMConfiguration(input: $input) {
      deletedScimConfigurationId @deleteRecord
    }
  }
`;

export interface DisconnectSCIMProviderDialogProps {
  scimConfigurationKey: DisconnectSCIMProviderDialog_scimConfiguration$key;
  copy: SCIMProviderCopy;
}

export function DisconnectSCIMProviderDialog({
  scimConfigurationKey,
  copy,
}: DisconnectSCIMProviderDialogProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const config = useFragment(fragment, scimConfigurationKey);
  const [open, setOpen] = useState(false);
  const [deleteSCIMConfiguration, isDeleting]
    = useMutation<DisconnectSCIMProviderDialog_deleteMutation>(
      deleteMutation,
      {
        successMessage: t(`${copy}.messages.disconnected`),
        errorToast: t(`${copy}.errors.disconnect`),
      },
    );

  function handleDisconnect() {
    void deleteSCIMConfiguration({
      variables: {
        input: {
          organizationId,
          scimConfigurationId: config.id,
        },
      },
    }).then(() => {
      setOpen(false);
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        render={(
          <Button size={2} variant="solid" color="red">
            {t(`${copy}.actions.disconnect`)}
          </Button>
        )}
      />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t(`${copy}.disconnect.title`)}</DialogTitle>
          <DialogDescription>
            {t(`${copy}.disconnect.description`)}
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <Text size={2} color="red" highContrast>
            {t(`${copy}.disconnect.warning`)}
          </Text>
        </DialogBody>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("common.actions.close")}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color="red"
            loading={isDeleting}
            onClick={handleDisconnect}
          >
            {t(`${copy}.actions.disconnect`)}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
