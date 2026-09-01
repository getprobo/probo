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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { SlackLogo } from "@probo/ui/src/v2/SlackLogo/SlackLogo";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { bindPreviewValue } from "../_lib/bindPreview";

import { bindConfirmCard } from "./variants";

export interface BindConfirmCardProps {
  externalTenantId: string;
  externalUserId: string;
  externalTenantName: string;
  externalUserName: string;
  isConfirming: boolean;
  onConfirm: () => void;
}

export function BindConfirmCard({
  externalTenantId,
  externalUserId,
  externalTenantName,
  externalUserName,
  isConfirming,
  onConfirm,
}: BindConfirmCardProps) {
  const { t } = useTranslation("bindings");
  const slots = bindConfirmCard();
  const tenant = bindPreviewValue(externalTenantName, externalTenantId);
  const account = bindPreviewValue(externalUserName, externalUserId);
  const tenantIsId = externalTenantName === "";
  const accountIsId = externalUserName === "";

  return (
    <Card variant="soft" size={3} padding="none" className={slots.frame()}>
      <div className={slots.header()}>
        <span aria-hidden className={slots.icon()}>
          <SlackLogo />
        </span>
        <Heading level={2} size={6} weight="medium" highContrast>
          {t("bind.title")}
        </Heading>
        <Text size={2} color="neutral">
          {t("bind.description")}
        </Text>
      </div>
      <div className={slots.body()}>
        <div className={slots.fields()}>
          {tenant !== "" && (
            <div className={slots.field()}>
              <Text size={1} weight="medium" color="neutral">
                {t("bind.fields.tenant")}
              </Text>
              <Text
                size={2}
                weight="medium"
                highContrast
                className={slots.value({ class: tenantIsId ? "font-mono" : undefined })}
              >
                {tenant}
              </Text>
            </div>
          )}
          <div className={slots.field()}>
            <Text size={1} weight="medium" color="neutral">
              {t("bind.fields.account")}
            </Text>
            <Text
              size={2}
              weight="medium"
              highContrast
              className={slots.value({ class: accountIsId ? "font-mono" : undefined })}
            >
              {account}
            </Text>
          </div>
        </div>
        <Button
          className={slots.action()}
          size={3}
          variant="solid"
          color="neutral"
          highContrast
          loading={isConfirming}
          onClick={onConfirm}
        >
          {t("bind.actions.confirm")}
        </Button>
      </div>
    </Card>
  );
}
