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

import { type ReactNode, useCallback, useMemo, useState } from "react";

import { SignInDialog } from "#/components/auth/SignInDialog";
import {
  type OpenSignInOptions,
  SignInDialogContextProvider,
} from "#/lib/auth/signInDialogContext";

interface SignInDialogProviderProps {
  children: ReactNode;
}

// Owns the single sign-in dialog instance and exposes `openSignIn` so any
// descendant (top bar, resource rows, …) can prompt authentication and pass the
// URL to return to afterwards.
export function SignInDialogProvider({ children }: SignInDialogProviderProps) {
  const [open, setOpen] = useState(false);
  const [continueTo, setContinueTo] = useState(() => window.location.href);

  const openSignIn = useCallback((options?: OpenSignInOptions) => {
    setContinueTo(options?.continueTo ?? window.location.href);
    setOpen(true);
  }, []);

  const value = useMemo(() => ({ openSignIn }), [openSignIn]);

  return (
    <SignInDialogContextProvider value={value}>
      {children}
      <SignInDialog open={open} onOpenChange={setOpen} continueTo={continueTo} />
    </SignInDialogContextProvider>
  );
}
