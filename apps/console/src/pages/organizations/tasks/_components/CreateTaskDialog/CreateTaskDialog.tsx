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

import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { type ReactElement, Suspense, useState } from "react";
import { useQueryLoader } from "react-relay";

import type { CreateTaskDialogLoadedQuery } from "#/__generated__/core/CreateTaskDialogLoadedQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateTaskDialogFallback } from "./CreateTaskDialogFallback";
import { CreateTaskDialogLoaded, linearConnectionQuery } from "./CreateTaskDialogLoaded";
import { CreateTaskDialogSkeleton } from "./CreateTaskDialogSkeleton";
import { type DialogOpenChangeDetails, isFloatingDismiss } from "./isFloatingDismiss";
import { useCreateTaskDialogForm } from "./useCreateTaskDialogForm";

export interface CreateTaskDialogProps {
  connectionId?: string;
  internalControlId?: string;
  children: ReactElement;
  onCompleted?: () => void;
}

export function CreateTaskDialog({
  connectionId,
  internalControlId,
  children,
  onCompleted,
}: CreateTaskDialogProps) {
  const organizationId = useOrganizationId();
  const [open, setOpen] = useState(false);
  const [session, setSession] = useState(0);
  const [queryRef, loadQuery] = useQueryLoader<CreateTaskDialogLoadedQuery>(
    linearConnectionQuery,
  );

  function close() {
    setOpen(false);
  }

  const form = useCreateTaskDialogForm({
    connectionId,
    internalControlId,
    onCompleted,
    onClose: close,
  });

  function handleOpenChange(next: boolean, details: DialogOpenChangeDetails) {
    if (!next && isFloatingDismiss(details)) {
      details.cancel();
      return;
    }

    if (next) {
      form.reset();
      loadQuery({ organizationId }, { fetchPolicy: "store-and-network" });
      setSession(current => current + 1);
    }
    setOpen(next);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={children} />
      <DialogPopup lockScroll>
        {queryRef == null
          ? (open ? <CreateTaskDialogSkeleton /> : null)
          : (
              <ErrorBoundary
                key={session}
                fallback={(_error, reset) => (
                  <CreateTaskDialogFallback
                    form={form}
                    onRetry={() => {
                      loadQuery(
                        { organizationId },
                        { fetchPolicy: "network-only" },
                      );
                      reset();
                    }}
                  />
                )}
              >
                <Suspense fallback={<CreateTaskDialogSkeleton />}>
                  <CreateTaskDialogLoaded queryRef={queryRef} form={form} />
                </Suspense>
              </ErrorBoundary>
            )}
      </DialogPopup>
    </Dialog>
  );
}
