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

import { ConfirmDialog } from "@probo/ui/src/v2/Dialog/ConfirmDialog";
import { useCallback, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { type BlockerFunction, useBlocker } from "react-router";

import { TASKS_NAMESPACE } from "./i18n";
import { shouldBlockUnsavedTaskComment } from "./unsavedTaskComment";

export function UnsavedTaskCommentGuard({ pending }: { pending: boolean }) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const shouldBlock = useCallback<BlockerFunction>(
    ({ currentLocation, nextLocation }) => shouldBlockUnsavedTaskComment(
      pending,
      currentLocation.pathname,
      nextLocation.pathname,
    ),
    [pending],
  );
  const blocker = useBlocker(shouldBlock);

  useEffect(() => {
    if (!pending) {
      return;
    }

    function onBeforeUnload(event: BeforeUnloadEvent) {
      event.preventDefault();
      event.returnValue = "";
    }

    window.addEventListener("beforeunload", onBeforeUnload);
    return () => {
      window.removeEventListener("beforeunload", onBeforeUnload);
    };
  }, [pending]);

  useEffect(() => {
    if (!pending && blocker.state === "blocked") {
      blocker.reset();
    }
  }, [blocker, pending]);

  function stay() {
    if (blocker.state === "blocked") {
      blocker.reset();
    }
  }

  function leave() {
    if (blocker.state === "blocked") {
      blocker.proceed();
    }
  }

  return (
    <ConfirmDialog
      open={blocker.state === "blocked"}
      onOpenChange={(open) => {
        if (!open) {
          stay();
        }
      }}
      title={t("comments.unsaved.title")}
      description={t("comments.unsaved.description")}
      cancelLabel={t("comments.unsaved.stay")}
      confirmLabel={t("comments.unsaved.leave")}
      confirmColor="red"
      onConfirm={leave}
    />
  );
}
