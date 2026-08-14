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

import { Button, Card, Slack, useConfirm } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EmployeeBindingListItem_binding$key } from "#/__generated__/core/EmployeeBindingListItem_binding.graphql";
import type { EmployeeBindingListItemDeleteMutation } from "#/__generated__/core/EmployeeBindingListItemDeleteMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const employeeBindingListItemFragment = graphql`
  fragment EmployeeBindingListItem_binding on ProbotIdentityBinding {
    id
    provider
    externalTenantId
    externalUserId
    externalTenantName
    externalUserName
  }
`;

const deleteProbotIdentityBindingMutation = graphql`
  mutation EmployeeBindingListItemDeleteMutation(
    $input: DeleteProbotIdentityBindingInput!
  ) {
    deleteProbotIdentityBinding(input: $input) {
      probotIdentityBindingId
      viewer {
        probotIdentityBindings {
          id
          ...EmployeeBindingListItem_binding
        }
      }
    }
  }
`;

interface EmployeeBindingListItemProps {
  bindingKey: EmployeeBindingListItem_binding$key;
  onDeleted: () => void;
}

export function EmployeeBindingListItem({
  bindingKey,
  onDeleted,
}: EmployeeBindingListItemProps) {
  const { t } = useTranslation();
  const confirm = useConfirm();
  const binding = useFragment(employeeBindingListItemFragment, bindingKey);
  const [deleteBinding, isDeleting]
    = useMutation<EmployeeBindingListItemDeleteMutation>(
      deleteProbotIdentityBindingMutation,
      {
        successMessage: t("employeeBindingsPage.messages.unlinked"),
        errorToast: t("employeeBindingsPage.errors.unlinkFailed"),
      },
    );

  const handleUnlink = () => {
    confirm(
      async () => {
        await deleteBinding({
          variables: { input: { id: binding.id } },
        });
        onDeleted();
      },
      {
        message: t("employeeBindingsPage.confirmations.unlink"),
      },
    );
  };

  return (
    <Card padded>
      <div className="flex items-center gap-3">
        <div className="h-10 w-10 flex items-center justify-center bg-subtle rounded">
          <Slack className="h-6 w-6" />
        </div>
        <div className="mr-auto min-w-0">
          <h3 className="text-base font-semibold">
            {t(`employeeBindingsPage.providers.${binding.provider}`, {
              defaultValue: binding.provider,
            })}
          </h3>
          <p className="text-sm text-txt-tertiary break-all">
            {t("employeeBindingsPage.connectedDescription", {
              workspace:
                binding.externalTenantName || binding.externalTenantId,
              account: binding.externalUserName || binding.externalUserId,
            })}
          </p>
        </div>
        <Button
          variant="danger"
          disabled={isDeleting}
          onClick={handleUnlink}
        >
          {t("employeeBindingsPage.actions.unlink")}
        </Button>
      </div>
    </Card>
  );
}
