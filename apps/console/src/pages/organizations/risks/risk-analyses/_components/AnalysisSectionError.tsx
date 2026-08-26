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

import { Button, Card } from "@probo/ui";
import { type ReactNode } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { useTranslation } from "react-i18next";

interface AnalysisSectionErrorProps {
  children: ReactNode;
}

export function AnalysisSectionError({ children }: AnalysisSectionErrorProps) {
  const { t } = useTranslation();

  return (
    <ErrorBoundary
      fallbackRender={({ resetErrorBoundary }) => (
        <Card padded>
          <div className="flex items-center justify-between gap-4">
            <p className="text-sm text-txt-secondary">{t("common.error")}</p>
            <Button variant="secondary" onClick={resetErrorBoundary}>
              {t("common.actions.tryAgain")}
            </Button>
          </div>
        </Card>
      )}
    >
      {children}
    </ErrorBoundary>
  );
}
