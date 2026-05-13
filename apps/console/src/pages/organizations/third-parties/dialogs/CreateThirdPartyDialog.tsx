// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import {
  Combobox,
  ComboboxItem,
  Dialog,
  DialogContent,
  DialogFooter,
  IconPlusLarge,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode, Suspense, useCallback, useState } from "react";
import { useQueryLoader } from "react-relay";
import { useDebounceCallback } from "usehooks-ts";

import type { CommonThirdPartyComboboxQuery } from "#/__generated__/core/CommonThirdPartyComboboxQuery.graphql";
import type { CreateThirdPartyInput } from "#/__generated__/core/ThirdPartyGraphCreateMutation.graphql";
import { useCreateThirdPartyMutation } from "#/hooks/graph/ThirdPartyGraph";

import {
  commonThirdPartiesQuery,
  CommonThirdPartyCombobox,
} from "./CommonThirdPartyCombobox";

type Props = {
  children: ReactNode;
  organizationId: string;
  connection: string;
};

export function CreateThirdPartyDialog({
  children,
  organizationId,
  connection,
}: Props) {
  const { __ } = useTranslate();
  const [createThirdParty] = useCreateThirdPartyMutation();
  const dialogRef = useDialogRef();
  const [searchQuery, setSearchQuery] = useState("");
  const [queryRef, loadQuery]
    = useQueryLoader<CommonThirdPartyComboboxQuery>(commonThirdPartiesQuery);

  const onSelect = async (thirdParty: Omit<CreateThirdPartyInput, "organizationId"> | string) => {
    const input
      = typeof thirdParty === "string"
        ? {
            organizationId,
            name: thirdParty,
            category: null,
          }
        : {
            ...thirdParty,
            organizationId,
          };
    await createThirdParty({
      variables: {
        input,
        connections: [connection],
      },
      onSuccess: () => {
        dialogRef.current?.close();
      },
    });
  };

  const debouncedLoadQuery = useDebounceCallback(
    useCallback(
      (name: string) => {
        loadQuery({ name });
      },
      [loadQuery],
    ),
    500,
  );

  const handleSearch = (name: string) => {
    setSearchQuery(name);
    const trimmed = name.trim();
    if (trimmed.length >= 2) {
      debouncedLoadQuery(trimmed);
    }
  };

  return (
    <Dialog ref={dialogRef} trigger={children} title={__("Add a third party")}>
      <DialogContent className="p-6">
        <Combobox onSearch={handleSearch} placeholder={__("Type third party's name")}>
          {searchQuery.trim().length >= 2 && queryRef && (
            <Suspense>
              <CommonThirdPartyCombobox
                queryRef={queryRef}
                onSelect={thirdPartyRef => void onSelect(thirdPartyRef)}
              />
            </Suspense>
          )}
          {searchQuery.trim().length >= 2 && (
            <ComboboxItem onClick={() => void onSelect(searchQuery.trim())}>
              <IconPlusLarge size={20} />
              {__("Create a new third party")}
              {" "}
              :
              {searchQuery}
            </ComboboxItem>
          )}
        </Combobox>
      </DialogContent>
      <DialogFooter />
    </Dialog>
  );
}
