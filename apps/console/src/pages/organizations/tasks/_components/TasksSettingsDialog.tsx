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

import { GearSixIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { Field } from "@probo/ui/src/v2/form/Field";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TasksSettingsDialog_organization$key } from "#/__generated__/core/TasksSettingsDialog_organization.graphql";
import type { TasksSettingsDialogDisconnectMutation } from "#/__generated__/core/TasksSettingsDialogDisconnectMutation.graphql";
import type { TasksSettingsDialogSetDefaultTeamMutation } from "#/__generated__/core/TasksSettingsDialogSetDefaultTeamMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { linearInitiateUrl } from "../_lib/linearInitiateUrl";
import { taskListPath } from "../_lib/taskPath";
import { tasksSettingsDialog } from "../variants";

import { LinearTeamCombobox, linearTeamLabel } from "./TaskLinearPublishField";

const tasksSettingsDialogFragment = graphql`
  fragment TasksSettingsDialog_organization on Organization {
    id
    canInitiateConnector: permission(action: "core:connector:initiate")
    canDeleteConnector: permission(action: "core:connector:delete")
    linearDefaultTeam {
      id
      name
      key
    }
    connectors(filter: { providers: [LINEAR_SYNC] }) {
      id
    }
  }
`;

const disconnectMutation = graphql`
  mutation TasksSettingsDialogDisconnectMutation($input: DeleteConnectorInput!) {
    deleteConnector(input: $input) {
      deletedConnectorId
    }
  }
`;

const setDefaultTeamMutation = graphql`
  mutation TasksSettingsDialogSetDefaultTeamMutation(
    $input: SetLinearSyncDefaultTeamInput!
  ) {
    setLinearSyncDefaultTeam(input: $input) {
      organization {
        id
        linearDefaultTeam {
          id
          name
          key
        }
      }
    }
  }
`;

interface TasksSettingsDialogProps {
  organizationKey: TasksSettingsDialog_organization$key;
}

export function TasksSettingsDialog({
  organizationKey,
}: TasksSettingsDialogProps) {
  const { t } = useTranslation("organizations/tasks");
  const organization = useFragment(tasksSettingsDialogFragment, organizationKey);
  const { sections, section, intro, list, row, identity, heading, defaultTeam } = tasksSettingsDialog();
  const linearConnectors = organization.connectors;
  const connected = linearConnectors.length > 0;
  const [confirmDisconnect, setConfirmDisconnect] = useState(false);
  const [isDisconnectingAll, setIsDisconnectingAll] = useState(false);
  const [disconnect, isDisconnecting] = useMutation<TasksSettingsDialogDisconnectMutation>(
    disconnectMutation,
    {
      successMessage: t("settingsDialog.linear.messages.disconnected"),
      errorToast: t("settingsDialog.linear.errors.disconnect"),
    },
  );
  const [setDefaultTeam, isSavingTeam] = useMutation<TasksSettingsDialogSetDefaultTeamMutation>(
    setDefaultTeamMutation,
    {
      successMessage: t("settingsDialog.linear.messages.defaultTeamSaved"),
      errorToast: t("settingsDialog.linear.errors.defaultTeam"),
    },
  );
  const selectedTeam = organization.linearDefaultTeam;

  function connect() {
    window.location.assign(
      linearInitiateUrl(organization.id, {
        continuePath: taskListPath(organization.id),
      }),
    );
  }

  function saveDefaultTeam(teamId: string | null) {
    void setDefaultTeam({
      variables: {
        input: {
          organizationId: organization.id,
          teamId,
        },
      },
    });
  }

  async function handleDisconnect() {
    const connectorIds = linearConnectors.map(connector => connector.id);
    if (connectorIds.length === 0) {
      return;
    }

    const connectorArgs = { filter: { providers: ["LINEAR_SYNC"] } };
    setIsDisconnectingAll(true);

    try {
      for (let index = 0; index < connectorIds.length; index++) {
        const connectorId = connectorIds[index];
        const isLast = index === connectorIds.length - 1;

        await disconnect(
          {
            variables: {
              input: { connectorId },
            },
            updater(store) {
              const deletedId = store.getRootField("deleteConnector")?.getValue("deletedConnectorId");
              if (typeof deletedId === "string") {
                store.delete(deletedId);
              }

              const organizationRecord = store.get(organization.id);
              const remaining = (organizationRecord?.getLinkedRecords("connectors", connectorArgs) ?? [])
                .filter(record => record != null && record.getValue("id") !== deletedId);
              organizationRecord?.setLinkedRecords(remaining, "connectors", connectorArgs);
              if (remaining.length === 0) {
                organizationRecord?.setLinkedRecord(null, "linearDefaultTeam");
              }
            },
          },
          isLast ? undefined : { successMessage: "" },
        );
      }

      setConfirmDisconnect(false);
    } catch {
      // Error toast is already shown by useMutation.
    } finally {
      setIsDisconnectingAll(false);
    }
  }

  return (
    <>
      <Dialog>
        <DialogTrigger
          render={(
            <IconButton
              variant="soft"
              color="neutral"
              aria-label={t("settingsDialog.trigger")}
            >
              <GearSixIcon />
            </IconButton>
          )}
        />
        <DialogPopup>
          <DialogHeader>
            <DialogTitle>{t("settingsDialog.title")}</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className={sections()}>
              <section className={section()}>
                <div className={intro()}>
                  <Heading level={3} size={2} weight="medium">
                    {t("settingsDialog.integrations.title")}
                  </Heading>
                  <Text size={2} color="faint">
                    {t("settingsDialog.integrations.description")}
                  </Text>
                </div>
                <Card size={1} variant="surface">
                  <ul className={list()}>
                    <li className={row()}>
                      <div className={identity()}>
                        <div className={heading()}>
                          <Text size={3} weight="medium" highContrast>
                            {t("settingsDialog.linear.name")}
                          </Text>
                          <Badge
                            size={1}
                            variant="soft"
                            color={connected ? "green" : "neutral"}
                          >
                            {connected
                              ? t("settingsDialog.status.connected")
                              : t("settingsDialog.status.notConnected")}
                          </Badge>
                        </div>
                        <Text size={2} color="faint">
                          {t("settingsDialog.linear.description")}
                        </Text>
                      </div>
                      {connected && organization.canDeleteConnector
                        ? (
                            <Button
                              size={1}
                              variant="solid"
                              color="red"
                              disabled={isDisconnecting || isDisconnectingAll}
                              onClick={() => setConfirmDisconnect(true)}
                            >
                              {t("settingsDialog.actions.disconnect")}
                            </Button>
                          )
                        : null}
                      {!connected && organization.canInitiateConnector
                        ? (
                            <Button
                              size={1}
                              variant="solid"
                              color="neutral"
                              highContrast
                              onClick={connect}
                            >
                              {t("settingsDialog.actions.connect")}
                            </Button>
                          )
                        : null}
                    </li>
                    {connected
                      ? (
                          <li className={defaultTeam()}>
                            <Field label={t("settingsDialog.linear.defaultTeam")}>
                              <LinearTeamCombobox
                                selectedId={selectedTeam?.id ?? null}
                                selectedLabel={selectedTeam
                                  ? linearTeamLabel(selectedTeam)
                                  : t("settingsDialog.linear.noDefaultTeam")}
                                placeholder={t("detailsPage.linear.searchTeam")}
                                disabled={
                                  !organization.canInitiateConnector
                                  || isSavingTeam
                                  || isDisconnecting
                                  || isDisconnectingAll
                                }
                                noneLabel={t("settingsDialog.linear.noDefaultTeam")}
                                onClear={() => saveDefaultTeam(null)}
                                onSelect={team => saveDefaultTeam(team.id)}
                              />
                            </Field>
                            <Text size={2} color="faint">
                              {t("settingsDialog.linear.defaultTeamDescription")}
                            </Text>
                          </li>
                        )
                      : null}
                  </ul>
                </Card>
              </section>
            </div>
          </DialogBody>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral">
                  {t("settingsDialog.actions.close")}
                </Button>
              )}
            />
          </DialogFooter>
        </DialogPopup>
      </Dialog>
      <Dialog open={confirmDisconnect} onOpenChange={setConfirmDisconnect}>
        <DialogPopup>
          <DialogHeader>
            <DialogTitle>{t("settingsDialog.linear.disconnect.title")}</DialogTitle>
            <DialogDescription>
              {t("settingsDialog.linear.disconnect.description")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral" disabled={isDisconnecting || isDisconnectingAll}>
                  {t("detailsPage.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="button"
              variant="solid"
              color="red"
              loading={isDisconnecting || isDisconnectingAll}
              onClick={() => void handleDisconnect()}
            >
              {t("settingsDialog.linear.disconnect.confirm")}
            </Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>
    </>
  );
}
