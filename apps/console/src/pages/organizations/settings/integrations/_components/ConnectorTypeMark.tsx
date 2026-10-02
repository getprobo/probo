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

import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { useTranslation } from "react-i18next";

import type { ConnectorProtocol } from "#/__generated__/core/ConnectorProviderListItem_provider.graphql";

import { connectMethodFromConnector } from "../_lib/connectMethods";

import { ConnectorMethodIcon } from "./ConnectorMethodIcon";

interface ConnectorTypeMarkProps {
  protocol: ConnectorProtocol;
  canReconnect: boolean;
}

export function ConnectorTypeMark({
  protocol,
  canReconnect,
}: ConnectorTypeMarkProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const method = connectMethodFromConnector(protocol, canReconnect);
  const label = t(`marketplacePage.methods.${method}`);

  return (
    <Tooltip>
      <TooltipTrigger
        aria-label={label}
        className="pointer-events-auto text-inherit"
      >
        <ConnectorMethodIcon method={method} />
      </TooltipTrigger>
      <TooltipPopup>{label}</TooltipPopup>
    </Tooltip>
  );
}
