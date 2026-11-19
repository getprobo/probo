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

import { connectRepoDocument, type RepoCollaborationHandle } from "@probo/collaboration";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

type CollaborationState = {
  attempt: number;
  handle?: RepoCollaborationHandle;
  failed?: boolean;
};

export function useDocumentCollaboration(documentId: string) {
  const { t } = useTranslation("organizations/documents");
  const toast = useToast();
  const addToast = toast.add;
  const attemptRef = useRef(0);
  const [state, setState] = useState<CollaborationState>();

  useEffect(() => {
    const attempt = ++attemptRef.current;
    let activeHandle: RepoCollaborationHandle | undefined;

    async function open() {
      try {
        const handle = await connectRepoDocument(documentId);
        if (attempt !== attemptRef.current) {
          handle.close();
          return;
        }
        activeHandle = handle;
        setState({ attempt, handle });
      } catch (error: unknown) {
        if (attempt !== attemptRef.current) {
          return;
        }
        setState({ attempt, failed: true });
        addToast({
          title: t("details.editor.openError"),
          description: error instanceof Error ? error.message : t("details.editor.openError"),
          type: "error",
        });
      }
    }

    void open();

    return () => {
      attemptRef.current += 1;
      activeHandle?.close();
      setState(current => (current?.attempt === attempt ? undefined : current));
    };
  }, [addToast, documentId, t]);

  return {
    handle: state?.failed ? undefined : state?.handle,
    failed: state?.failed === true,
  };
}
