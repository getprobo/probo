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

import { GearIcon, XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Drawer } from "@probo/ui/src/v2/Drawer/Drawer";
import { DrawerBody } from "@probo/ui/src/v2/Drawer/DrawerBody";
import { DrawerClose } from "@probo/ui/src/v2/Drawer/DrawerClose";
import { DrawerDescription } from "@probo/ui/src/v2/Drawer/DrawerDescription";
import { DrawerHeader } from "@probo/ui/src/v2/Drawer/DrawerHeader";
import { DrawerPopup } from "@probo/ui/src/v2/Drawer/DrawerPopup";
import { DrawerTitle } from "@probo/ui/src/v2/Drawer/DrawerTitle";
import { DrawerTrigger } from "@probo/ui/src/v2/Drawer/DrawerTrigger";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SCIMProviderSettingsDrawer_scimBridge$key } from "#/__generated__/iam/SCIMProviderSettingsDrawer_scimBridge.graphql";
import type { SCIMProviderSettingsDrawer_updateMutation } from "#/__generated__/iam/SCIMProviderSettingsDrawer_updateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  isExcludedUserEmail,
  type SCIMProviderCopy,
} from "../_lib/scimProvider";
import { scimProviderCard } from "../variants";

const fragment = graphql`
  fragment SCIMProviderSettingsDrawer_scimBridge on SCIMBridge {
    id
    excludedUserNames
  }
`;

const updateMutation = graphql`
  mutation SCIMProviderSettingsDrawer_updateMutation(
    $input: UpdateSCIMBridgeInput!
  ) {
    updateSCIMBridge(input: $input) {
      scimBridge {
        id
        excludedUserNames
      }
    }
  }
`;

export interface SCIMProviderSettingsDrawerProps {
  scimBridgeKey: SCIMProviderSettingsDrawer_scimBridge$key;
  copy: SCIMProviderCopy;
}

export function SCIMProviderSettingsDrawer({
  scimBridgeKey,
  copy,
}: SCIMProviderSettingsDrawerProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const bridge = useFragment(fragment, scimBridgeKey);
  const [open, setOpen] = useState(false);
  const [newUser, setNewUser] = useState("");
  const [emailError, setEmailError] = useState<string | null>(null);
  const {
    settingsFields,
    addRow,
    addField,
    userList,
    userRow,
    userEmpty,
  } = scimProviderCard();
  const excludedUserNames = [...bridge.excludedUserNames];
  const [updateSCIMBridge, isUpdating]
    = useMutation<SCIMProviderSettingsDrawer_updateMutation>(
      updateMutation,
      {
        successMessage: t(`${copy}.messages.excludedUsersUpdated`),
        errorToast: t(`${copy}.errors.update`),
      },
    );

  function saveExcludedUserNames(next: string[]) {
    return updateSCIMBridge({
      variables: {
        input: {
          organizationId,
          scimBridgeId: bridge.id,
          excludedUserNames: next,
        },
      },
    });
  }

  function handleAddUser() {
    const user = newUser.trim().toLowerCase();
    if (user === "" || excludedUserNames.includes(user)) {
      return;
    }
    if (!isExcludedUserEmail(user)) {
      setEmailError(t(`${copy}.settings.invalidEmail`));
      return;
    }
    void saveExcludedUserNames([...excludedUserNames, user]).then(() => {
      setNewUser("");
      setEmailError(null);
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <Drawer open={open} onOpenChange={setOpen} swipeDirection="right">
      <DrawerTrigger
        render={(
          <Button
            size={2}
            variant="surface"
            color="neutral"
            iconStart={<GearIcon />}
          >
            {t(`${copy}.actions.settings`)}
          </Button>
        )}
      />
      <DrawerPopup side="right" size={2}>
        <DrawerHeader>
          <DrawerTitle>{t(`${copy}.settings.title`)}</DrawerTitle>
          <DrawerClose
            render={(
              <IconButton
                size={1}
                variant="ghost"
                color="neutral"
                aria-label={t("common.actions.close")}
              >
                <XIcon />
              </IconButton>
            )}
          />
        </DrawerHeader>
        <DrawerDescription>
          {t(`${copy}.settings.excludedUserNamesDescription`)}
        </DrawerDescription>
        <DrawerBody>
          <div className={settingsFields()}>
            <div className={addRow()}>
              <Field
                label={t(`${copy}.settings.excludedUserNames`)}
                className={addField()}
                error={emailError ?? undefined}
              >
                <TextField
                  type="email"
                  size={2}
                  value={newUser}
                  placeholder="user@example.com"
                  onValueChange={(value) => {
                    setNewUser(value);
                    setEmailError(null);
                  }}
                  onKeyDown={(event) => {
                    if (event.key !== "Enter") {
                      return;
                    }
                    event.preventDefault();
                    if (isUpdating) {
                      return;
                    }
                    handleAddUser();
                  }}
                />
              </Field>
              <Button
                size={2}
                variant="soft"
                color="neutral"
                disabled={isUpdating}
                onClick={handleAddUser}
              >
                {t(`${copy}.actions.add`)}
              </Button>
            </div>
            {excludedUserNames.length === 0
              ? (
                  <Text size={2} color="faint" className={userEmpty()}>
                    {t(`${copy}.settings.noExcludedUserNames`)}
                  </Text>
                )
              : (
                  <div className={userList()}>
                    {excludedUserNames.map(user => (
                      <div key={user} className={userRow()}>
                        <Text size={2} highContrast>
                          {user}
                        </Text>
                        <IconButton
                          size={1}
                          variant="ghost"
                          color="neutral"
                          disabled={isUpdating}
                          aria-label={t(`${copy}.actions.remove`)}
                          onClick={() => {
                            void saveExcludedUserNames(
                              excludedUserNames.filter(name => name !== user),
                            ).catch(() => {
                              // Error toast is already shown by useMutation.
                            });
                          }}
                        >
                          <XIcon />
                        </IconButton>
                      </div>
                    ))}
                  </div>
                )}
          </div>
        </DrawerBody>
      </DrawerPopup>
    </Drawer>
  );
}
