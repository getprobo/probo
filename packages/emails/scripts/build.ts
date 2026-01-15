import { render } from "@react-email/components";
import { copyFile, mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import * as React from "react";

import ConfirmEmail from "../src/ConfirmEmail";
import DocumentExport from "../src/DocumentExport";
import DocumentSigning from "../src/DocumentSigning";
import FrameworkExport from "../src/FrameworkExport";
import Invitation from "../src/Invitation";
import PasswordReset from "../src/PasswordReset";
import TrustCenterAccess from "../src/TrustCenterAccess";
import TrustCenterDocumentAccessRejected from "../src/TrustCenterDocumentAccessRejected";
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
