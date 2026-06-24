// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { tv } from "tailwind-variants";
import { create } from "zustand";
import { combine } from "zustand/middleware";

type Toast = {
  title: string;
  description: string;
  variant?: "success" | "error" | "warning" | "info";
  id: string;
};

const duration = 3000;

const useToasts = create(
  combine({ toasts: [] as Toast[] }, set => ({
    add: (toast: Pick<Toast, "title" | "description" | "variant">) => {
      const id = Date.now().toString();
      set(state => ({
        toasts: [{ ...toast, id }, ...state.toasts],
      }));
      setTimeout(() => {
        set(state => ({
          toasts: [
            ...state.toasts.filter(toast => toast.id !== id),
          ],
        }));
      }, duration);
    },
    remove: (id: string) =>
      set(state => ({
        toasts: state.toasts.filter(toast => toast.id !== id),
      })),
  })),
);

/**
 * Hook used to trigger toasts
 */
export function useToast() {
  return {
    toast: useToasts().add,
  };
}

/**
 * Toasts displayed at the bottom right of the screen.
 */
export function Toasts() {
  const { toasts, remove } = useToasts();
  return (
    <div className="fixed z-100 bottom-4 right-4 space-y-2 w-85">
      {toasts.map(toast => (
        <div key={toast.id}>
          <Toast {...toast} onClose={() => remove(toast.id)} />
        </div>
      ))}
    </div>
  );
}

const toast = tv({
  slots: {
    wrapper:
            "p-2 border rounded relative animate-in slide-in-from-right w-full cursor-pointer ",
    timer: "absolute bottom-0 left-0 w-full h-1 starting:scale-x-0 scale-x-100 origin-left transition-all ease-linear",
  },
  variants: {
    variant: {
      success: {
        wrapper:
                    "bg-success text-invert text-txt-success border-border-success",
        timer: "bg-border-success",
      },
      error: {
        wrapper:
                    "bg-danger text-invert text-txt-danger border-border-danger",
        timer: "bg-border-danger",
      },
      warning: {
        wrapper:
                    "bg-warning text-invert text-txt-warning border-border-warning",
        timer: "bg-border-warning",
      },
      info: {
        wrapper: "bg-info text-invert text-txt-info border-border-info",
        timer: "bg-border-info",
      },
    },
  },
  defaultVariants: {
    variant: "success",
  },
});

export function Toast({ onClose, ...props }: Toast & { onClose: () => void }) {
  const { wrapper, timer } = toast(props);
  return (
    <div className={wrapper()} onClick={onClose}>
      <div className="font-semibold">{props.title}</div>
      <div className="text-sm opacity-80">{props.description}</div>
      <div
        className={timer()}
        style={{ transitionDuration: `${duration}ms` }}
      />
    </div>
  );
}
