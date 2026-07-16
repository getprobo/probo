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
import type { GraphQLTaggedNode, MutationParameters } from "relay-runtime";

import { useMutation } from "#/lib/relay/useMutation";

import { openExportedFile } from "./openExportedFile";

// Shared "export then open" behavior for the document/file/report list items:
// commit the resource's export mutation, open the returned base64 payload, and
// let failures surface through the mutation notifier's toast. Each caller
// supplies its typed mutation and a selector for the payload string.
export function useExportAndOpen<T extends MutationParameters>(
  mutation: GraphQLTaggedNode,
  selectData: (response: T["response"]) => string,
): readonly [(variables: T["variables"]) => void, boolean] {
  const [commit, isExporting] = useMutation<T>(mutation);

  const open = useCallback(
    (variables: T["variables"]) => {
      commit({
        variables,
        onCompleted: response => openExportedFile(selectData(response)),
      }).catch(() => {
        // The mutation failure is already surfaced through a toast.
      });
    },
    [commit, selectData],
  );

  return [open, isExporting];
}
