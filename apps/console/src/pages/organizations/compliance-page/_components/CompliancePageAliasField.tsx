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

import { useTranslate } from "@probo/i18n";
import { Field } from "@probo/ui";
import { useCallback, useState } from "react";
import { graphql } from "relay-runtime";

import type { CompliancePageAliasField_removeResourceAliasMutation } from "#/__generated__/core/CompliancePageAliasField_removeResourceAliasMutation.graphql";
import type { CompliancePageAliasField_setResourceAliasMutation } from "#/__generated__/core/CompliancePageAliasField_setResourceAliasMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

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

  const [setResourceAlias, isSettingAlias] = useMutation<CompliancePageAliasField_setResourceAliasMutation>(
    setResourceAliasMutation,
    {
      successMessage: __("Alias updated successfully."),
      errorToast: __("Failed to update alias"),
    },
  );
  const [removeResourceAlias, isRemovingAlias] = useMutation<CompliancePageAliasField_removeResourceAliasMutation>(
    removeResourceAliasMutation,
    {
      successMessage: __("Alias removed successfully."),
      errorToast: __("Failed to remove alias"),
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
      // useMutation already shows an error toast.
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
