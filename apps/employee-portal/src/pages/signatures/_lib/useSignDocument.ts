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

import { graphql } from "react-relay";

import type { useSignDocumentMutation } from "#/__generated__/core/useSignDocumentMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const signDocumentMutation = graphql`
  mutation useSignDocumentMutation($input: SignDocumentInput!) {
    signDocument(input: $input) {
      documentVersionSignature {
        id
        state
      }
    }
  }
`;

// Signs the given document version and marks the employee document signed in
// the Relay store so the request panel updates in place.
export function useSignDocument(documentId: string) {
  const [signDocument, isSigning] = useMutation<useSignDocumentMutation>(
    signDocumentMutation,
  );

  return [
    (documentVersionId: string) => signDocument({
      variables: { input: { documentVersionId } },
      updater: (store) => {
        store.get(documentId)?.setValue(true, "signed");
        store.get(documentVersionId)?.setValue(true, "signed");
      },
    }),
    isSigning,
  ] as const;
}
