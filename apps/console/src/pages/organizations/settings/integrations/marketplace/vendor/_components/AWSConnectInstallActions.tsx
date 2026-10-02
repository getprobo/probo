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

import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { AWSConnectInstallActions_install$key } from "#/__generated__/core/AWSConnectInstallActions_install.graphql";

import { TerraformInstallButton } from "./TerraformInstallButton";

const awsConnectInstallActionsFragment = graphql`
  fragment AWSConnectInstallActions_install on AWSConnectorSetup {
    terraformSnippet
    cloudFormationQuickCreateURL
  }
`;

interface AWSConnectInstallActionsProps {
  installKey: AWSConnectInstallActions_install$key;
}

export function AWSConnectInstallActions({
  installKey,
}: AWSConnectInstallActionsProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const setup = useFragment(awsConnectInstallActionsFragment, installKey);

  return (
    <div className="flex shrink-0 flex-wrap justify-end gap-2">
      {setup.cloudFormationQuickCreateURL != null && (
        <ButtonAnchor href={setup.cloudFormationQuickCreateURL} target="_blank" rel="noreferrer" variant="soft">
          {t("marketplacePage.workloadIdentity.actions.installViaCloudFormation")}
        </ButtonAnchor>
      )}
      <TerraformInstallButton snippet={setup.terraformSnippet} />
    </div>
  );
}
