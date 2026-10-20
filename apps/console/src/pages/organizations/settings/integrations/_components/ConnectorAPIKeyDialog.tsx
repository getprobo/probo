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

import { Form } from "@base-ui/react/form";
import { KeyIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorAPIKeyDialog_connector$key } from "#/__generated__/core/ConnectorAPIKeyDialog_connector.graphql";
import type { ConnectorAPIKeyDialogUpdateMutation } from "#/__generated__/core/ConnectorAPIKeyDialogUpdateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const connectorAPIKeyDialogFragment = graphql`
  fragment ConnectorAPIKeyDialog_connector on Connector {
    id
    canUpdate: permission(action: "core:connector:update")
    canReplaceAPIKey
    apiKeyFormat {
      pattern
      example
    }
  }
`;

const updateAPIKeyConnectorMutation = graphql`
  mutation ConnectorAPIKeyDialogUpdateMutation($input: UpdateAPIKeyConnectorInput!) {
    updateAPIKeyConnector(input: $input) {
      connector {
        id
        connectionStatus
      }
    }
  }
`;

interface ConnectorAPIKeyDialogProps {
  connectorKey: ConnectorAPIKeyDialog_connector$key;
}

export function ConnectorAPIKeyDialog({
  connectorKey,
}: ConnectorAPIKeyDialogProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const connector = useFragment(connectorAPIKeyDialogFragment, connectorKey);
  const [open, setOpen] = useState(false);
  const [apiKey, setApiKey] = useState("");
  const [error, setError] = useState<string | undefined>();
  const [updateAPIKey, isUpdating] = useMutation<ConnectorAPIKeyDialogUpdateMutation>(
    updateAPIKeyConnectorMutation,
  );

  if (!connector.canUpdate || !connector.canReplaceAPIKey) {
    return null;
  }

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setApiKey("");
      setError(undefined);
    }
  }

  function rejectsShape(value: string) {
    const pattern = connector.apiKeyFormat?.pattern;
    if (pattern == null || pattern === "") {
      return false;
    }

    // MatchString allows text after the match. An HTML pattern would not.
    try {
      return !new RegExp(pattern).test(value);
    } catch {
      return false;
    }
  }

  function shapeError(value: string) {
    const trimmed = value.trim();
    if (trimmed === "" || !rejectsShape(trimmed)) {
      return undefined;
    }

    return t("detailsPage.apiKey.invalid", {
      example: connector.apiKeyFormat?.example,
    });
  }

  async function save() {
    const trimmed = apiKey.trim();
    const nextError = shapeError(trimmed);
    if (trimmed === "" || nextError != null) {
      setError(nextError);
      return;
    }
    setError(undefined);

    try {
      await updateAPIKey({
        variables: {
          input: {
            connectorId: connector.id,
            apiKey: trimmed,
          },
        },
      }, {
        errorToast: t("detailsPage.apiKey.saveFailed"),
        successMessage: t("detailsPage.apiKey.saved"),
      });
      handleOpenChange(false);
    } catch {
      // useMutation has already shown the refusal.
    }
  }

  const label = t("detailsPage.apiKey.update");
  const canSave = apiKey.trim() !== "";

  return (
    <>
      <Tooltip>
        <TooltipTrigger
          render={(
            <IconButton
              variant="ghost"
              color="neutral"
              size={1}
              aria-label={label}
              onClick={() => setOpen(true)}
            >
              <KeyIcon />
            </IconButton>
          )}
        />
        <TooltipPopup>{label}</TooltipPopup>
      </Tooltip>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogPopup>
          <DialogHeader>
            <DialogTitle>{t("detailsPage.apiKey.title")}</DialogTitle>
            <DialogDescription>
              {t("detailsPage.apiKey.description")}
            </DialogDescription>
          </DialogHeader>
          <Form
            className="flex flex-col gap-4"
            onFormSubmit={() => {
              void save();
            }}
          >
            <DialogBody>
              <Field required label={t("marketplacePage.fields.apiKey")} error={error}>
                <TextField
                  name="apiKey"
                  type="password"
                  required
                  autoComplete="off"
                  value={apiKey}
                  placeholder={connector.apiKeyFormat?.example}
                  disabled={isUpdating}
                  onBlur={() => {
                    setError(shapeError(apiKey));
                  }}
                  onValueChange={(value) => {
                    setApiKey(value);
                    setError(undefined);
                  }}
                />
              </Field>
            </DialogBody>
            <DialogFooter>
              <DialogClose
                render={(
                  <Button variant="soft" color="neutral" disabled={isUpdating}>
                    {t("detailsPage.actions.cancel")}
                  </Button>
                )}
              />
              <Button type="submit" variant="solid" loading={isUpdating} disabled={!canSave}>
                {t("detailsPage.apiKey.save")}
              </Button>
            </DialogFooter>
          </Form>
        </DialogPopup>
      </Dialog>
    </>
  );
}
