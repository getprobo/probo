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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { useTranslation } from "react-i18next";

import { useCopyValue } from "../_lib/useCopyValue";

interface TerraformInstallButtonProps {
  snippet: string;
}

export function TerraformInstallButton({ snippet }: TerraformInstallButtonProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const copyValue = useCopyValue();

  return (
    <Button
      type="button"
      variant="soft"
      className="shrink-0"
      onClick={() => copyValue(
        snippet,
        t("marketplacePage.workloadIdentity.messages.copiedTerraform"),
        t("marketplacePage.workloadIdentity.messages.copyFailed"),
      )}
    >
      {t("marketplacePage.workloadIdentity.actions.installViaTerraform")}
    </Button>
  );
}
