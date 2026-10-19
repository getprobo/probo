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

import { IconFlagCA, IconFlagEU, IconFlagUS } from "@probo/ui";
import type { ComponentType } from "react";

import { connectorProviderVisibleName, ovhcloudRegionLabel } from "../_lib/connectorProviderLabel";

const ovhcloudFlags: Record<string, ComponentType<{ size?: number }>> = {
  OVHCLOUD: IconFlagEU,
  OVHCLOUD_US: IconFlagUS,
  OVHCLOUD_CA: IconFlagCA,
};

interface ConnectorProviderNameProps {
  provider: string;
  name: string;
  flagSize?: number;
}

export function ConnectorProviderName({
  provider,
  name,
  flagSize = 14,
}: ConnectorProviderNameProps) {
  return (
    <span className="inline-flex min-w-0 items-center gap-1.5">
      <span className="truncate">{connectorProviderVisibleName(provider, name)}</span>
      <ConnectorRegionFlag provider={provider} size={flagSize} />
    </span>
  );
}

export function ConnectorRegionFlag({
  provider,
  size = 14,
}: {
  provider: string;
  size?: number;
}) {
  const Flag = ovhcloudFlags[provider];
  const label = ovhcloudRegionLabel(provider);
  if (Flag == null || label == null) {
    return null;
  }

  return (
    <span
      className="inline-flex shrink-0 overflow-hidden rounded-[2px] ring-1 ring-black/15"
      role="img"
      aria-label={label}
    >
      <span aria-hidden="true" className="inline-flex">
        <Flag size={size} />
      </span>
    </span>
  );
}
