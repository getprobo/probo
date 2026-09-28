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

interface SAMLConfigurationSsoUrlProps {
  testLoginUrl: string;
}

export function SAMLConfigurationSsoUrl({
  testLoginUrl,
}: SAMLConfigurationSsoUrlProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const { url, urlRow, urlValue } = samlConfigurationListItem();

  async function handleCopyUrl() {
    try {
      await navigator.clipboard.writeText(testLoginUrl);
      toast({
        title: t("samlConfigurationList.messages.copied"),
        description: t("samlConfigurationList.fields.ssoUrl"),
        variant: "success",
      });
    } catch {
      toast({
        title: t("samlConfigurationList.errors.copy"),
        description: t("samlConfigurationList.fields.ssoUrl"),
        variant: "error",
      });
    }
  }

  return (
    <div className={url()}>
      <Text size={1} color="faint">
        {t("samlConfigurationList.fields.ssoUrl")}
      </Text>
      <div className={urlRow()}>
        <Code variant="ghost" size={1} className={urlValue()}>
          {testLoginUrl}
        </Code>
        <IconButton
          size={1}
          variant="ghost"
          color="neutral"
          aria-label={t("samlConfigurationList.actions.copyUrl")}
          onClick={() => {
            void handleCopyUrl();
          }}
        >
          <CopyIcon />
        </IconButton>
      </div>
    </div>
  );
}
