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

import { ShieldCheckIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { internalControlsPage } from "../variants";

interface InternalControlsEmptyProps {
  filtered: boolean;
  canCreate: boolean;
  create: ReactNode;
  onClear: () => void;
}

export function InternalControlsEmpty(props: InternalControlsEmptyProps) {
  const { filtered, canCreate, create, onClear } = props;
  const { t } = useTranslation("organizations/internal-controls");
  const { empty, emptyActions } = internalControlsPage();

  return (
    <TableRow>
      <TableCell colSpan={6}>
        <div className={empty()}>
          <ShieldCheckIcon size={32} weight="duotone" aria-hidden />
          <Heading level={2} size={4} weight="medium">
            {filtered ? t("list.empty.filteredTitle") : t("list.empty.title")}
          </Heading>
          <Text size={2} color="faint">
            {filtered ? t("list.empty.filteredDescription") : t("list.empty.description")}
          </Text>
          <div className={emptyActions()}>
            {filtered
              ? (
                  <Button variant="soft" color="neutral" onClick={onClear}>
                    {t("list.actions.clearFilters")}
                  </Button>
                )
              : canCreate && create}
          </div>
        </div>
      </TableCell>
    </TableRow>
  );
}
