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

import { CopyIcon } from "@phosphor-icons/react";
import { useToast } from "@probo/ui";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { samlConfigurationListItem } from "../variants";

const TXT_HOST = "@";
const TXT_TYPE = "TXT";

interface SAMLConfigurationDnsRecordProps {
  domainVerificationToken: string;
}

export function SAMLConfigurationDnsRecord({
  domainVerificationToken,
}: SAMLConfigurationDnsRecordProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const { record, recordHeader, recordField, recordValue, code } = samlConfigurationListItem();
  const value = `probo-verification=${domainVerificationToken}`;

  async function copyToClipboard(text: string, description: string) {
    try {
      await navigator.clipboard.writeText(text);
      toast({
        title: t("samlConfigurationList.messages.copied"),
        description,
        variant: "success",
      });
    } catch {
      toast({
        title: t("samlConfigurationList.errors.copy"),
        description,
        variant: "error",
      });
    }
  }

  return (
    <div className={record()}>
      <div className={recordHeader()}>
        <Text size={2} weight="medium">{TXT_TYPE}</Text>
      </div>
      <div className={recordField()}>
        <Text size={1} color="faint">{t("samlConfigurationList.dns.name")}</Text>
        <div className={recordValue()}>
          <Code size={1} className={code()}>{TXT_HOST}</Code>
          <IconButton
            size={1}
            variant="soft"
            color="neutral"
            aria-label={t("samlConfigurationList.dns.copyName")}
            onClick={() => {
              void copyToClipboard(TXT_HOST, t("samlConfigurationList.dns.name"));
            }}
          >
            <CopyIcon />
          </IconButton>
        </div>
      </div>
      <div className={recordField()}>
        <Text size={1} color="faint">{t("samlConfigurationList.dns.value")}</Text>
        <div className={recordValue()}>
          <Code size={1} className={code()}>{value}</Code>
          <IconButton
            size={1}
            variant="soft"
            color="neutral"
            aria-label={t("samlConfigurationList.dns.copyValue")}
            onClick={() => {
              void copyToClipboard(value, t("samlConfigurationList.dns.value"));
            }}
          >
            <CopyIcon />
          </IconButton>
        </div>
      </div>
      <Text size={1} color="faint">
        {t("samlConfigurationList.dns.note")}
      </Text>
    </div>
  );
}
