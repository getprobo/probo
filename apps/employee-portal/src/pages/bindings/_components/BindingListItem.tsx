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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { BindingListItem_binding$key } from "#/__generated__/core/BindingListItem_binding.graphql";

import { bindPreviewValue } from "../_lib/bindPreview";

import { UnlinkBindingDialog } from "./UnlinkBindingDialog";
import { bindingListItem } from "./variants";

const bindingListItemFragment = graphql`
  fragment BindingListItem_binding on ProbotIdentityBinding @throwOnFieldError {
    id
    provider
    externalTenantId
    externalUserId
    externalTenantName
    externalUserName
  }
`;

export interface BindingListItemProps {
  bindingKey: BindingListItem_binding$key;
  onDeleted: () => void;
}

export function BindingListItem({ bindingKey, onDeleted }: BindingListItemProps) {
  const { t } = useTranslation("bindings");
  const binding = useFragment(bindingListItemFragment, bindingKey);
  const slots = bindingListItem();
  const workspace = bindPreviewValue(
    binding.externalTenantName,
    binding.externalTenantId,
  );
  const account = bindPreviewValue(
    binding.externalUserName,
    binding.externalUserId,
  );

  return (
    <TableRow align="center">
      <TableRowHeaderCell className={slots.accountCell()} style={{ padding: 0 }}>
        <div className={slots.row()}>
          <Text size={2} weight="medium" highContrast className={slots.title()}>
            {t(`list.providers.${binding.provider}`, {
              defaultValue: binding.provider,
            })}
          </Text>
          <Text size={1} color="current" className={slots.meta()}>
            {t("list.connectedDescription", { account, workspace })}
          </Text>
        </div>
      </TableRowHeaderCell>
      <TableCell className={slots.actionCell()} justify="end">
        <UnlinkBindingDialog bindingId={binding.id} onDeleted={onDeleted}>
          <Button size={2} variant="soft" color="red">
            {t("list.actions.unlink")}
          </Button>
        </UnlinkBindingDialog>
      </TableCell>
    </TableRow>
  );
}
