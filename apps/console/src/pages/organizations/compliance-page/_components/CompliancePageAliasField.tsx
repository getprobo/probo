// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { Field } from "@probo/ui";
import { useCallback, useState } from "react";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const setResourceAliasMutation = graphql`
  mutation CompliancePageAliasField_setResourceAliasMutation(
    $input: SetResourceAliasInput!
  ) {
    setResourceAlias(input: $input) {
      resourceAlias {
        resourceId
        alias
      }
    }
  }
`;

const removeResourceAliasMutation = graphql`
  mutation CompliancePageAliasField_removeResourceAliasMutation(
    $input: RemoveResourceAliasInput!
  ) {
    removeResourceAlias(input: $input) {
      deletedResourceId
    }
  }
`;

export function CompliancePageAliasField(props: {
  resourceId: string;
  alias: string | null | undefined;
  canSetAlias: boolean;
  canRemoveAlias: boolean;
}) {
  const { resourceId, alias, canSetAlias, canRemoveAlias } = props;

  const { __ } = useTranslate();
  const [value, setValue] = useState(alias ?? "");
  const [prevAlias, setPrevAlias] = useState(alias);

  if (alias !== prevAlias) {
    setPrevAlias(alias);
    setValue(alias ?? "");
  }

  const [setResourceAlias, isSettingAlias] = useMutationWithToasts(
    setResourceAliasMutation,
    {
      successMessage: __("Alias updated successfully."),
      errorMessage: __("Failed to update alias"),
    },
  );
  const [removeResourceAlias, isRemovingAlias] = useMutationWithToasts(
    removeResourceAliasMutation,
    {
      successMessage: __("Alias removed successfully."),
      errorMessage: __("Failed to remove alias"),
    },
  );

  const handleBlur = useCallback(async () => {
    const trimmed = value.trim();
    const current = alias ?? "";

    if (trimmed === current) {
      return;
    }

    try {
      if (trimmed === "") {
        if (current !== "" && canRemoveAlias) {
          await removeResourceAlias({
            variables: {
              input: {
                resourceId,
              },
            },
          });
        }

        return;
      }

      if (!canSetAlias) {
        setValue(current);
        return;
      }

      await setResourceAlias({
        variables: {
          input: {
            resourceId,
            alias: trimmed,
          },
        },
      });
    } catch {
      // useMutationWithToasts already shows an error toast.
    }
  }, [alias, canRemoveAlias, canSetAlias, removeResourceAlias, resourceId, setResourceAlias, value]);

  const canEdit = canSetAlias || (canRemoveAlias && (alias ?? "") !== "");

  return (
    <Field
      type="text"
      value={value}
      onChange={event => setValue(event.target.value)}
      onBlur={() => void handleBlur()}
      disabled={!canEdit || isSettingAlias || isRemovingAlias}
      placeholder={__("e.g. privacy-policy")}
      className="min-w-[160px]"
    />
  );
}
