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

import { Button } from "../Button/Button";

import { Toaster } from "./Toaster";

export default {
  title: "v2/Toaster",
  component: Toaster,
};

const types = ["neutral", "success", "error", "warning", "info"] as const;

const titles: Record<(typeof types)[number], string> = {
  neutral: "Saved",
  success: "Access requested",
  error: "Something went wrong",
  warning: "Session expiring",
  info: "New update available",
};

const descriptions: Record<(typeof types)[number], string> = {
  neutral: "Your changes have been saved.",
  success: "We'll email you once it's approved.",
  error: "Please try again in a moment.",
  warning: "Your session expires in 5 minutes.",
  info: "A subprocessor list was published.",
};

function ToasterDemo({ withDescription = false }: { withDescription?: boolean }) {
  const toast = Toast.useToastManager();

  return (
    <div className="flex flex-wrap gap-2">
      {types.map(type => (
        <Button
          key={type}
          variant="soft"
          color="neutral"
          highContrast
          onClick={() =>
            toast.add({
              title: titles[type],
              description: withDescription ? descriptions[type] : undefined,
              type,
            })}
        >
          {type}
        </Button>
      ))}
      <Toaster />
    </div>
  );
}

// Title-only toasts (the common case — most app mutations toast a single line).
// The Toaster reads Base UI's toast manager, so it must be wrapped in a
// `Toast.Provider` (mounted once at the app root in real usage).
export function Default() {
  return (
    <Toast.Provider>
      <ToasterDemo />
    </Toast.Provider>
  );
}

// Toasts carrying a supporting description line under the title.
export function WithDescription() {
  return (
    <Toast.Provider>
      <ToasterDemo withDescription />
    </Toast.Provider>
  );
}
