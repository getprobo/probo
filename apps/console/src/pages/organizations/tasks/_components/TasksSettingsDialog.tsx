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
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TasksSettingsDialog_organization$key } from "#/__generated__/core/TasksSettingsDialog_organization.graphql";

import { linearInitiateUrl } from "../_lib/linearInitiateUrl";
import { taskListPath } from "../_lib/taskPath";
import { tasksSettingsDialog } from "../variants";

const tasksSettingsDialogFragment = graphql`
  fragment TasksSettingsDialog_organization on Organization {
    id
    canInitiateConnector: permission(action: "core:connector:initiate")
    connectors(filter: { providers: [LINEAR_SYNC] }) {
      id
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
  const { sections, section, intro, list, row, identity, heading } = tasksSettingsDialog();
  const connectorId = organization.connectors[0]?.id;
  const connected = connectorId != null;

  function connect() {
    window.location.assign(
      linearInitiateUrl(organization.id, {
        connectorId,
        continuePath: taskListPath(organization.id),
      }),
    );
  }

  return (
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
                    {organization.canInitiateConnector
                      ? (
                          <Button
                            size={1}
                            variant={connected ? "soft" : "solid"}
                            color="neutral"
                            highContrast={!connected}
                            onClick={connect}
                          >
                            {connected
                              ? t("settingsDialog.actions.reconnect")
                              : t("settingsDialog.actions.connect")}
                          </Button>
                        )
                      : null}
                  </li>
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
  );
}
