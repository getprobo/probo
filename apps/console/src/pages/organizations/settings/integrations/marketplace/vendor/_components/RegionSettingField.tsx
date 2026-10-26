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

import { IconFlagCA, IconFlagEU, IconFlagUS, IconGlobe } from "@probo/ui";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";

const REGIONS = [
  { value: "EU", labelKey: "marketplacePage.fields.regionEurope" },
  { value: "US", labelKey: "marketplacePage.fields.regionUnitedStates" },
  { value: "CA", labelKey: "marketplacePage.fields.regionCanada" },
  { value: "ENT", labelKey: "marketplacePage.fields.regionEnterprise" },
] as const;

type Region = (typeof REGIONS)[number]["value"];

interface RegionSettingFieldProps {
  required: boolean;
  value: string | null;
  onValueChange: (value: string) => void;
}

export function RegionSettingField({ required, value, onValueChange }: RegionSettingFieldProps) {
  const { t } = useTranslation("organizations/settings/integrations");

  return (
    <Field required={required} label={t("marketplacePage.fields.region")}>
      <Select
        value={value}
        onValueChange={(next: string | null) => onValueChange(next ?? "")}
      >
        <SelectTrigger placeholder={t("marketplacePage.fields.regionPlaceholder")}>
          {(selected: string | null) => {
            const region = REGIONS.find(entry => entry.value === selected);
            return region ? <RegionChoice region={region.value} label={t(region.labelKey)} /> : null;
          }}
        </SelectTrigger>
        <SelectPopup>
          {REGIONS.map(entry => (
            <SelectItem key={entry.value} value={entry.value}>
              <RegionChoice region={entry.value} label={t(entry.labelKey)} />
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </Field>
  );
}

function RegionChoice({ region, label }: { region: Region; label: string }) {
  return (
    <span className="inline-flex min-w-0 items-center gap-2">
      <RegionFlag region={region} />
      <span className="truncate">{label}</span>
    </span>
  );
}

const REGION_FLAGS = {
  EU: IconFlagEU,
  US: IconFlagUS,
  CA: IconFlagCA,
  ENT: IconGlobe,
} as const;

function RegionFlag({ region }: { region: Region }) {
  const Icon = REGION_FLAGS[region];

  return (
    <span className="inline-flex shrink-0 overflow-hidden rounded-1 ring-1 ring-sand-a6 ring-inset" aria-hidden>
      <Icon size={14} />
    </span>
  );
}
