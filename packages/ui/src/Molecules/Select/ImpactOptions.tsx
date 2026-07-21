// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { useTranslation } from "react-i18next";

import { Option } from "../../Atoms/Select/Select";

export function ImpactOptions() {
  const { t } = useTranslation();

  const descriptions = {
    LOW: {
      label: t("ui.risk.severity.low"),
      description: t("ui.impactOptions.low"),
    },
    MEDIUM: {
      label: t("ui.risk.severity.medium"),
      description: t("ui.impactOptions.medium"),
    },
    HIGH: {
      label: t("ui.risk.severity.high"),
      description: t("ui.impactOptions.high"),
    },
    CRITICAL: {
      label: t("ui.risk.severity.critical"),
      description: t("ui.impactOptions.critical"),
    },
  } as const;

  return (
    <>
      {Object.entries(descriptions).map(([key, description]) => (
        <Option
          key={key}
          value={key}
          className="border-b border-border-low"
        >
          <span>
            <span className="text-sm font-bold">
              {description.label}
            </span>
            ,
            {" "}
            <span className="text-sm text-txt-secondary">
              {description.description}
            </span>
          </span>
        </Option>
      ))}
    </>
  );
}
