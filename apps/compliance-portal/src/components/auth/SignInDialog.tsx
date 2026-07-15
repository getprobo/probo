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
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { useTranslation } from "react-i18next";

import { LoginForm } from "./LoginForm";

interface SignInDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  // Absolute URL to return to after authentication.
  continueTo: string;
}

// The reusable "Login Dialog" from the design: a modal sign-in gate composed of
// the kit Dialog plus the magic-link / SSO login form.
export function SignInDialog({ open, onOpenChange, continueTo }: SignInDialogProps) {
  const { t } = useTranslation();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("auth.signIn.title")}</DialogTitle>
          <DialogDescription>{t("auth.signIn.description")}</DialogDescription>
        </DialogHeader>
        <LoginForm continueTo={continueTo} onCancel={() => onOpenChange(false)} />
      </DialogPopup>
    </Dialog>
  );
}
