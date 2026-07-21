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

import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import { useMutation } from "#/lib/relay/useMutation";

import type { useUpdateLocaleMutation } from "./__generated__/useUpdateLocaleMutation.graphql";
import type { UrlLocale } from "./locale";

const updateLocaleMutation = graphql`
  mutation useUpdateLocaleMutation($input: UpdateLocaleInput!) {
    updateLocale(input: $input) {
      identity {
        id
        locale
      }
    }
  }
`;

// Persists the viewer's preferred UI locale on their identity.
export function useUpdateLocale() {
  const { t } = useTranslation();
  const [commit, isUpdating] = useMutation<useUpdateLocaleMutation>(updateLocaleMutation, {
    errorToast: t("locale.updateFailed"),
  });

  const updateLocale = useCallback(async (locale: UrlLocale) => {
    await commit({
      variables: { input: { locale } },
    });
  }, [commit]);

  return [updateLocale, isUpdating] as const;
}
