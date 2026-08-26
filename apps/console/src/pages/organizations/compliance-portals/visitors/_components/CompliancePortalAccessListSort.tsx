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

import { ArrowsDownUpIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { useTranslation } from "react-i18next";

import { type AccessListSort, useAccessListSort } from "../_lib/useAccessListSort";
import { accessSection } from "../variants";

const sortOptions: AccessListSort[] = ["requests", "joined"];

function isAccessListSort(value: string): value is AccessListSort {
  return value === "requests" || value === "joined";
}

export function CompliancePortalAccessListSort() {
  const { t } = useTranslation("organizations/compliance-portals");
  const { sort, setSort } = useAccessListSort();
  const { sort: sortClass } = accessSection();

  return (
    <div className={sortClass()}>
      <Dropdown>
        <DropdownTrigger
          render={(
            <Button
              variant="ghost"
              color="neutral"
              size={2}
              iconStart={<ArrowsDownUpIcon />}
              aria-label={t("accessList.sort.label")}
            >
              {t(`accessList.sort.${sort}`)}
            </Button>
          )}
        />
        <DropdownPopup align="end">
          <DropdownRadioGroup
            value={sort}
            onValueChange={(value: string) => {
              if (isAccessListSort(value)) {
                setSort(value);
              }
            }}
          >
            {sortOptions.map(option => (
              <DropdownRadioItem key={option} value={option}>
                {t(`accessList.sort.${option}`)}
              </DropdownRadioItem>
            ))}
          </DropdownRadioGroup>
        </DropdownPopup>
      </Dropdown>
    </div>
  );
}
