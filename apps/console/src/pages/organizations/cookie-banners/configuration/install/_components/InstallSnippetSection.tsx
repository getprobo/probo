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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Trans, useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { InstallSnippetSection_cookieBanner$key } from "#/__generated__/core/InstallSnippetSection_cookieBanner.graphql";

import { cookieBannerInstallSection } from "../../../variants";
import { cookieBannerEmbedSnippet } from "../_lib/embedSnippet";

const fragment = graphql`
  fragment InstallSnippetSection_cookieBanner on CookieBanner {
    id
    capabilities {
      tcf
    }
  }
`;

interface InstallSnippetSectionProps {
  cookieBannerKey: InstallSnippetSection_cookieBanner$key;
}

export function InstallSnippetSection({
  cookieBannerKey,
}: InstallSnippetSectionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const { root, intro, snippetWrap, snippet, snippetCopy } = cookieBannerInstallSection();
  const banner = useFragment(fragment, cookieBannerKey);
  const baseUrl = new URL("/api/cookie-banner/v1", window.location.origin).href;
  const code = cookieBannerEmbedSnippet({
    bannerId: banner.id,
    baseUrl,
    tcf: banner.capabilities.tcf,
  });

  function handleCopy() {
    void navigator.clipboard.writeText(code).then(
      () => {
        toast({
          title: t("codeSnippets.messages.copiedTitle"),
          description: t("codeSnippets.messages.copied"),
          variant: "success",
        });
      },
      () => {
        toast({
          title: t("codeSnippets.errors.title"),
          description: t("codeSnippets.errors.copy"),
          variant: "error",
        });
      },
    );
  }

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("installPage.title")}
        </Heading>
        <Text size={2} color="faint">
          {t("installPage.description")}
        </Text>
        {banner.capabilities.tcf && (
          <Text size={2} color="faint">
            {t("codeSnippets.tcfNote")}
          </Text>
        )}
      </div>
      <Card size={2} variant="soft">
        <div className={snippetWrap()}>
          <IconButton
            variant="surface"
            color="neutral"
            size={1}
            className={snippetCopy()}
            aria-label={t("codeSnippets.actions.copy")}
            onClick={handleCopy}
          >
            <CopyIcon />
          </IconButton>
          <pre className={snippet()}>
            <code>{code}</code>
          </pre>
        </div>
      </Card>
      <Text size={2} color="faint">
        <Trans
          ns="organizations/cookie-banners"
          i18nKey="codeSnippets.documentation"
          components={{
            link: (
              <Anchor
                href="https://www.probo.com/docs/product/cookie-banner/javascript-sdk"
                target="_blank"
                rel="noopener noreferrer"
              />
            ),
          }}
        />
      </Text>
    </section>
  );
}
