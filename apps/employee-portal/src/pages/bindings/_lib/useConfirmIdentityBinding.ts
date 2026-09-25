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

import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { useConfirmIdentityBindingMutation } from "#/__generated__/core/useConfirmIdentityBindingMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const confirmIdentityBindingMutation = graphql`
  mutation useConfirmIdentityBindingMutation(
    $input: ConfirmProbotIdentityBindingInput!
  ) {
    confirmProbotIdentityBinding(input: $input) {
      probotIdentityBinding {
        id
      }
    }
  }
`;

export function useConfirmIdentityBinding() {
  const { t } = useTranslation("bindings");

  return useMutation<useConfirmIdentityBindingMutation>(
    confirmIdentityBindingMutation,
    {
      successMessage: t("bind.messages.linked"),
      errorToast: t("bind.errors.linkFailed"),
    },
  );
}
