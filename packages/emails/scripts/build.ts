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

import { render } from "react-email";
import { copyFile, mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import * as React from "react";

import MailingListUpdates from "../src/MailingListUpdates";
import ConfirmEmail from "../src/ConfirmEmail";
import DocumentExport from "../src/DocumentExport";
import DocumentApproval from "../src/DocumentApproval";
import DocumentSigning from "../src/DocumentSigning";
import FrameworkExport from "../src/FrameworkExport";
import Invitation from "../src/Invitation";
import PasswordReset from "../src/PasswordReset";
import TrustCenterAccess from "../src/TrustCenterAccess";
import TrustCenterDocumentAccessRejected from "../src/TrustCenterDocumentAccessRejected";
import ElectronicSignatureCertificate from "../src/ElectronicSignatureCertificate";
import MailingListSubscription from "../src/MailingListSubscription";
import MailingListUnsubscription from "../src/MailingListUnsubscription";
import MagicLink from "../src/MagicLink";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

type TemplateConfig = {
  name: string;
  render: () => React.ReactElement;
};

const templates: TemplateConfig[] = [
  {
    name: "confirm-email",
    render: () => ConfirmEmail(),
  },
  {
    name: "password-reset",
    render: () => PasswordReset(),
  },
  {
    name: "invitation",
    render: () => Invitation(),
  },
  {
    name: "document-approval",
    render: () => DocumentApproval(),
  },
  {
    name: "document-signing",
    render: () => DocumentSigning(),
  },
  {
    name: "document-export",
    render: () => DocumentExport(),
  },
  {
    name: "framework-export",
    render: () => FrameworkExport(),
  },
  {
    name: "trust-center-access",
    render: () => TrustCenterAccess(),
  },
  {
    name: "trust-center-document-access-rejected",
    render: () => TrustCenterDocumentAccessRejected(),
  },
  {
    name: "magic-link",
    render: () => MagicLink(),
  },
  {
    name: "electronic-signature-certificate",
    render: () => ElectronicSignatureCertificate(),
  },
  {
    name: "mailing-list-subscription",
    render: () => MailingListSubscription(),
  },
  {
    name: "mailing-list-unsubscription",
    render: () => MailingListUnsubscription(),
  },
  {
    name: "mailing-list-updates",
    render: () => MailingListUpdates(),
  },
];

async function build() {
  const outputDir = join(__dirname, "..", "dist");
  const templatesDir = join(__dirname, "..", "templates");
  await mkdir(outputDir, { recursive: true });

  for (const template of templates) {
    const html = await render(template.render(), { pretty: true });

    const htmlPath = join(outputDir, `${template.name}.html.tmpl`);
    const txtSrcPath = join(templatesDir, `${template.name}.txt`);
    const txtDstPath = join(outputDir, `${template.name}.txt.tmpl`);

    await writeFile(htmlPath, html);
    await copyFile(txtSrcPath, txtDstPath);
  }
}

build().catch((err) => {
  console.error("Failed to build email templates:", err);
  process.exit(1);
});
