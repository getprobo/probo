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

import { Toast } from "@base-ui/react/toast";
import { XIcon } from "@phosphor-icons/react";

import { toaster } from "./variants";

type ToastType = "success" | "error" | "neutral";

function resolveType(type: string | undefined): ToastType {
  return type === "success" || type === "error" ? type : "neutral";
}

// Renders the active toasts from the Base UI toast manager. Mount once at the
// app root, inside a `<Toast.Provider>`.
export function Toaster() {
  const { toasts } = Toast.useToastManager();
  const slots = toaster();

  return (
    <Toast.Portal>
      <Toast.Viewport className={slots.viewport()}>
        {toasts.map((toast) => {
          const variant = { type: resolveType(toast.type) };

          return (
            <Toast.Root key={toast.id} toast={toast} className={slots.toast(variant)}>
              <div className={slots.content()}>
                <Toast.Title className={slots.title(variant)} />
                <Toast.Description className={slots.description()} />
              </div>
              <Toast.Close className={slots.close()} aria-label="Close">
                <XIcon />
              </Toast.Close>
            </Toast.Root>
          );
        })}
      </Toast.Viewport>
    </Toast.Portal>
  );
}
