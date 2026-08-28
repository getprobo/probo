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

import { AppleLogoIcon, DownloadSimpleIcon, WindowsLogoIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { AGENT_INSTALL_URL } from "#/pages/devices/_lib/registerDeviceSteps";

import { RegisterDeviceCard } from "./RegisterDeviceCard";
import { downloadList } from "./variants";

export interface DownloadStepProps {
  onContinue: () => void;
}

export function DownloadStep({ onContinue }: DownloadStepProps) {
  const { t } = useTranslation("devices");
  const slots = downloadList();

  return (
    <RegisterDeviceCard
      icon={<DownloadSimpleIcon />}
      title={t("download.title")}
      description={t("download.description")}
      action={(
        <Button size={3} variant="solid" color="neutral" highContrast onClick={onContinue}>
          {t("download.continue")}
        </Button>
      )}
    >
      <ButtonAnchor
        href={AGENT_INSTALL_URL}
        target="_blank"
        rel="noopener noreferrer"
        size={3}
        variant="solid"
        color="neutral"
        highContrast
        iconStart={<AppleLogoIcon />}
      >
        {t("download.featured")}
      </ButtonAnchor>
      <List className="w-full">
        <DownloadRow
          title={t("download.macosIntel.title")}
          meta={t("download.macosIntel.meta")}
          icon={<AppleLogoIcon />}
          className={slots.item()}
          metaClassName={slots.meta()}
        />
        <DownloadRow
          title={t("download.windows.title")}
          meta={t("download.windows.meta")}
          icon={<WindowsLogoIcon />}
          className={slots.item()}
          metaClassName={slots.meta()}
        />
      </List>
    </RegisterDeviceCard>
  );
}

function DownloadRow({
  title,
  meta,
  icon,
  className,
  metaClassName,
}: {
  title: string;
  meta: string;
  icon: ReactNode;
  className: string;
  metaClassName: string;
}) {
  const { t } = useTranslation("devices");

  return (
    <ListItem className={className}>
      <Text size={2} weight="medium" highContrast className="min-w-0 flex-1">
        {title}
      </Text>
      <Text size={1} color="neutral" className={metaClassName}>
        {meta}
      </Text>
      <ButtonAnchor
        href={AGENT_INSTALL_URL}
        target="_blank"
        rel="noopener noreferrer"
        size={2}
        variant="soft"
        color="neutral"
        iconStart={icon}
      >
        {t("download.download")}
      </ButtonAnchor>
    </ListItem>
  );
}
