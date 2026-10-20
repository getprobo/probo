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
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";
import { useSearchParams } from "react-router";

import type { TasksSettingsDialog_organization$key } from "#/__generated__/core/TasksSettingsDialog_organization.graphql";
import type { TasksSettingsDialogDisconnectMutation } from "#/__generated__/core/TasksSettingsDialogDisconnectMutation.graphql";
import type { TasksSettingsDialogRefetchQuery } from "#/__generated__/core/TasksSettingsDialogRefetchQuery.graphql";
import type { TasksSettingsDialogUpdateLinearSettingsMutation } from "#/__generated__/core/TasksSettingsDialogUpdateLinearSettingsMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { linearInitiateUrl } from "../../_lib/linearInitiateUrl";
import { linearTeamLabel } from "../../_lib/linearTeamLabel";
import { taskListPath } from "../../_lib/taskPath";
import { tasksSettingsDialog } from "../../variants";
import { LinearTeamCombobox } from "../LinearTeamCombobox";

import { isLinearSettingsReturn } from "./linearSettingsReturn";

const linearSyncConnectorName = "tasks";

const tasksSettingsDialogFragment = graphql`
  fragment TasksSettingsDialog_organization on Organization
  @refetchable(queryName: "TasksSettingsDialogRefetchQuery") {
    id
    canInitiateConnector: permission(action: "core:connector:initiate")
    canDeleteConnector: permission(action: "core:connector:delete")
    linearSettings {
      defaultTeam {
        id
        name
        key
      }
      issueImport
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

const updateLinearSettingsMutation = graphql`
  mutation TasksSettingsDialogUpdateLinearSettingsMutation(
    $input: UpdateLinearSettingsInput!
  ) {
    updateLinearSettings(input: $input) {
      organization {
        id
        linearSettings {
          defaultTeam {
            id
            name
            key
          }
          issueImport
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
  const toast = useToast();
  const [searchParams, setSearchParams] = useSearchParams();
  const [open, setOpen] = useState(() => isLinearSettingsReturn(searchParams));
  const handledLinearReturn = useRef(false);
  const [organization, refetch] = useRefetchableFragment<
    TasksSettingsDialogRefetchQuery,
    TasksSettingsDialog_organization$key
  >(
    tasksSettingsDialogFragment,
    organizationKey,
  );
  const {
    sections, section, intro, list, row, identity, heading, defaultTeam, defaultTeamControl, issueImport,
  } = tasksSettingsDialog();
  const linearConnectors = organization.connectors;
  const connectorId = linearConnectors[0]?.id;
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
  const [updateLinearSettings, isSaving] = useMutation<TasksSettingsDialogUpdateLinearSettingsMutation>(
    updateLinearSettingsMutation,
  );
  const linkedTeam = organization.linearSettings.defaultTeam ?? null;
  const canEditDefaultTeam = organization.canInitiateConnector;

  useEffect(() => {
    if (handledLinearReturn.current || !isLinearSettingsReturn(searchParams)) {
      return;
    }

    handledLinearReturn.current = true;

    const callbackError = searchParams.get("error");
    if (callbackError) {
      toast.add({
        title: t("settingsDialog.linear.errors.connect"),
        description: callbackError,
        type: "error",
      });
    }

    setSearchParams((params) => {
      params.delete("connector_id");
      params.delete("provider");
      params.delete("error");
      return params;
    }, { replace: true });
  }, [searchParams, setSearchParams, t, toast]);

  function connect() {
    window.location.assign(
      linearInitiateUrl(organization.id, {
        connectorId,
        name: connected ? undefined : linearSyncConnectorName,
        continuePath: taskListPath(organization.id),
      }),
    );
  }

  async function saveDefaultTeam(teamId: string | null) {
    try {
      await updateLinearSettings({
        variables: {
          input: {
            organizationId: organization.id,
            defaultTeamId: teamId,
          },
        },
      }, {
        successMessage: t("settingsDialog.linear.messages.defaultTeamSaved"),
        errorToast: t("settingsDialog.linear.errors.defaultTeam"),
      });
    } catch {
      // Error toast is already shown by useMutation.
    }
  }

  async function saveIssueImport(status: "ACTIVE" | "INACTIVE" | null) {
    if (status == null || status === organization.linearSettings.issueImport) {
      return;
    }

    try {
      await updateLinearSettings({
        variables: {
          input: {
            organizationId: organization.id,
            issueImport: status,
          },
        },
      }, {
        successMessage: t("settingsDialog.linear.messages.issueImportSaved"),
        errorToast: t("settingsDialog.linear.errors.issueImport"),
      });
    } catch {
      // Error toast is already shown by useMutation.
    }
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
              const organizationRecord = store.get(organization.id);
              const remaining = (organizationRecord?.getLinkedRecords("connectors", connectorArgs) ?? [])
                .filter(record => record != null && record.getDataID() !== deletedId);
              organizationRecord?.setLinkedRecords(remaining, "connectors", connectorArgs);
              if (remaining.length === 0) {
                const settingsRecord = organizationRecord?.getLinkedRecord("linearSettings");
                settingsRecord?.setLinkedRecord(null, "defaultTeam");
                settingsRecord?.setValue("INACTIVE", "issueImport");
              }
              if (typeof deletedId === "string") {
                store.delete(deletedId);
              }
            },
          },
          isLast ? undefined : { successMessage: "" },
        );
      }

      setConfirmDisconnect(false);
    } catch {
      // The deleted connector may have owned the default team. Reload so a
      // later failure does not keep that team in the store.
      refetch({}, { fetchPolicy: "network-only" });
    } finally {
      setIsDisconnectingAll(false);
    }
  }

  const defaultTeamCombobox = (
    <LinearTeamCombobox
      selectedId={linkedTeam?.id ?? null}
      selectedLabel={linkedTeam
        ? linearTeamLabel({ name: linkedTeam.name, key: linkedTeam.key })
        : t("settingsDialog.linear.noDefaultTeam")}
      ariaLabel={t("settingsDialog.linear.defaultTeam")}
      placeholder={t("detailsPage.linear.searchTeam")}
      disabled={
        !canEditDefaultTeam
        || isSaving
        || isDisconnecting
        || isDisconnectingAll
      }
      noneLabel={t("settingsDialog.linear.noDefaultTeam")}
      onClear={() => void saveDefaultTeam(null)}
      onSelect={team => void saveDefaultTeam(team.id)}
    />
  );

  const issueImportSelect = (
    <Select
      value={organization.linearSettings.issueImport}
      disabled={
        linkedTeam == null
        || !canEditDefaultTeam
        || isSaving
        || isDisconnecting
        || isDisconnectingAll
      }
      onValueChange={status => void saveIssueImport(status)}
    >
      <SelectTrigger size={2} aria-label={t("settingsDialog.linear.issueImport")}>
        {(status: "ACTIVE" | "INACTIVE" | null) => (
          status == null
            ? null
            : t(status === "ACTIVE"
                ? "settingsDialog.linear.issueImportStatus.active"
                : "settingsDialog.linear.issueImportStatus.inactive")
        )}
      </SelectTrigger>
      <SelectPopup>
        <SelectItem value="ACTIVE">
          {t("settingsDialog.linear.issueImportStatus.active")}
        </SelectItem>
        <SelectItem value="INACTIVE">
          {t("settingsDialog.linear.issueImportStatus.inactive")}
        </SelectItem>
      </SelectPopup>
    </Select>
  );

  return (
    <>
      <Dialog open={open} onOpenChange={setOpen}>
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
                              {canEditDefaultTeam
                                ? defaultTeamCombobox
                                : (
                                    <Tooltip>
                                      <TooltipTrigger
                                        render={(
                                          <span tabIndex={0} className={defaultTeamControl()}>
                                            {defaultTeamCombobox}
                                          </span>
                                        )}
                                      />
                                      <TooltipPopup>
                                        {t("settingsDialog.linear.defaultTeamReadOnly")}
                                      </TooltipPopup>
                                    </Tooltip>
                                  )}
                            </Field>
                            <Text size={2} color="faint">
                              {t("settingsDialog.linear.defaultTeamDescription")}
                            </Text>
                          </li>
                        )
                      : null}
                    {connected
                      ? (
                          <li className={issueImport()}>
                            <Field label={t("settingsDialog.linear.issueImport")}>
                              {canEditDefaultTeam && linkedTeam != null
                                ? issueImportSelect
                                : (
                                    <Tooltip>
                                      <TooltipTrigger
                                        render={(
                                          <span tabIndex={0} className={defaultTeamControl()}>
                                            {issueImportSelect}
                                          </span>
                                        )}
                                      />
                                      <TooltipPopup>
                                        {linkedTeam == null
                                          ? t("settingsDialog.linear.issueImportNeedsTeam")
                                          : t("settingsDialog.linear.issueImportReadOnly")}
                                      </TooltipPopup>
                                    </Tooltip>
                                  )}
                            </Field>
                            <Text size={2} color="faint">
                              {t("settingsDialog.linear.issueImportDescription")}
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
