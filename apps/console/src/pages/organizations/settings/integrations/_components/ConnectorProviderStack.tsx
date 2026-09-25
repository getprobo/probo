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

import { CaretDownIcon } from "@phosphor-icons/react";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import type { AccessReviewSourceProviderListItem_provider$key } from "#/__generated__/core/AccessReviewSourceProviderListItem_provider.graphql";
import type { ConnectorListItem_connector$key } from "#/__generated__/core/ConnectorListItem_connector.graphql";

import { connectorProviderStack } from "../variants";

import { ConnectorListItem } from "./ConnectorListItem";

const maxPeek = 2;
const peekStepXPx = 10;
const peekStepYPx = 10;

export interface ConnectorProviderSlide extends ConnectorListItem_connector$key {
  readonly id: string;
}

interface ConnectorProviderStackProps {
  connectors: readonly ConnectorProviderSlide[];
  providerKey?: AccessReviewSourceProviderListItem_provider$key;
  organizationId: string;
  canConnect: boolean;
}

export function ConnectorProviderStack({
  connectors,
  providerKey,
  organizationId,
  canConnect,
}: ConnectorProviderStackProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const [activeIndex, setActiveIndex] = useState(0);
  const { root, deck, slide, page, next } = connectorProviderStack();
  const ids = connectors.map(connector => connector.id).join("\0");
  const peek = Math.min(connectors.length - 1, maxPeek);
  const visibleCount = peek + 1;

  useEffect(() => {
    setActiveIndex(0);
  }, [ids]);

  return (
    <div className={root()}>
      <div
        style={{
          paddingRight: peek * peekStepXPx,
          paddingBottom: peek * peekStepYPx,
        }}
      >
        <div className={deck()}>
          {Array.from({ length: visibleCount }, (_, depth) => {
            const index = (activeIndex + depth) % connectors.length;
            const connector = connectors[index];
            if (connector == null) {
              return null;
            }

            return (
              <div
                key={connector.id}
                className={slide({ className: "transition-transform duration-300 ease-out" })}
                style={{
                  transform: `translate(${depth * peekStepXPx}px, ${depth * peekStepYPx}px)`,
                  zIndex: visibleCount - depth,
                }}
                inert={depth !== 0}
              >
                <ConnectorListItem
                  connectorKey={connector}
                  providerKey={providerKey}
                  organizationId={organizationId}
                  canConnect={canConnect}
                />
                {depth === 0 && (
                  <Text size={1} color="faint" className={page()}>
                    {`${activeIndex + 1}/${connectors.length}`}
                  </Text>
                )}
              </div>
            );
          })}
        </div>
      </div>
      <div className={next()}>
        <IconButton
          variant="surface"
          color="neutral"
          size={2}
          aria-label={t("listPage.actions.nextConnector", {
            current: activeIndex + 1,
            total: connectors.length,
          })}
          onClick={() => {
            setActiveIndex(current => (current + 1) % connectors.length);
          }}
        >
          <CaretDownIcon />
        </IconButton>
      </div>
    </div>
  );
}
