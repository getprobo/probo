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

import { formatError } from "@probo/helpers";
import {
  Combobox,
  ComboboxItem,
  Dialog,
  DialogContent,
  DialogFooter,
  IconPlusLarge,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, Suspense, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { CommonThirdPartyComboboxQuery } from "#/__generated__/core/CommonThirdPartyComboboxQuery.graphql";
import type {
  CreateThirdPartyDialogCreateMutation,
  CreateThirdPartyInput,
} from "#/__generated__/core/CreateThirdPartyDialogCreateMutation.graphql";

import { commonThirdPartiesQuery, CommonThirdPartyCombobox } from "./CommonThirdPartyCombobox";

const createThirdPartyMutation = graphql`
  mutation CreateThirdPartyDialogCreateMutation(
    $input: CreateThirdPartyInput!
    $connections: [ID!]!
  ) {
    createThirdParty(input: $input) {
      thirdPartyEdge @prependEdge(connections: $connections) {
        node {
          id
          canDelete: permission(action: "core:thirdParty:delete")
          ...ThirdPartyRow_thirdParty
        }
      }
    }
  }
`;

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
  const { t } = useTranslation();
  const { toast } = useToast();
  const [createThirdParty] = useMutation<CreateThirdPartyDialogCreateMutation>(createThirdPartyMutation);
  const dialogRef = useDialogRef();
  const [searchQuery, setSearchQuery] = useState("");
  const [queryRef, loadQuery]
    = useQueryLoader<CommonThirdPartyComboboxQuery>(commonThirdPartiesQuery);

  const onSelect = (thirdParty: Omit<CreateThirdPartyInput, "organizationId"> | string) => {
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
    createThirdParty({
      variables: {
        input,
        connections: [connection],
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: t("createThirdPartyDialog.messages.error"),
            description: formatError(t("createThirdPartyDialog.errors.create"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("createThirdPartyDialog.messages.success"),
          description: t("createThirdPartyDialog.messages.created"),
          variant: "success",
        });
        dialogRef.current?.close();
      },
      onError(error) {
        toast({
          title: t("createThirdPartyDialog.messages.error"),
          description: formatError(t("createThirdPartyDialog.errors.create"), error),
          variant: "error",
        });
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
    <Dialog ref={dialogRef} trigger={children} title={t("createThirdPartyDialog.title")}>
      <DialogContent className="p-6">
        <Combobox onSearch={handleSearch} placeholder={t("createThirdPartyDialog.searchPlaceholder")}>
          {searchQuery.trim().length >= 2 && queryRef && (
            <Suspense>
              <CommonThirdPartyCombobox
                queryRef={queryRef}
                excludeNames={new Set()}
                onSelect={onSelect}
              />
            </Suspense>
          )}
          {searchQuery.trim().length >= 2 && (
            <ComboboxItem onClick={() => onSelect(searchQuery.trim())}>
              <IconPlusLarge size={20} />
              {t("createThirdPartyDialog.createNew", { name: searchQuery })}
            </ComboboxItem>
          )}
        </Combobox>
      </DialogContent>
      <DialogFooter />
    </Dialog>
  );
}
