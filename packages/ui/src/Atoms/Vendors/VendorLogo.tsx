// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { ComponentProps, FC } from "react";

import { Brex } from "./Brex";
import { Cloudflare } from "./Cloudflare";
import { DocuSign } from "./DocuSign";
import { Figma } from "./Figma";
import { GitHub } from "./GitHub";
import { Google } from "./Google";
import { HubSpot } from "./HubSpot";
import { Intercom } from "./Intercom";
import { Linear } from "./Linear";
import { Microsoft } from "./Microsoft";
import { Notion } from "./Notion";
import { OnePassword } from "./OnePassword";
import { OpenAI } from "./OpenAI";
import { Resend } from "./Resend";
import { Sentry } from "./Sentry";
import { Slack } from "./Slack";
import { Supabase } from "./Supabase";
import { Tally } from "./Tally";

const vendors: Record<string, FC<ComponentProps<"svg">>> = {
  BREX: Brex,
  CLOUDFLARE: Cloudflare,
  DOCUSIGN: DocuSign,
  FIGMA: Figma,
  GITHUB: GitHub,
  GOOGLE: Google,
  GOOGLE_WORKSPACE: Google,
  HUBSPOT: HubSpot,
  INTERCOM: Intercom,
  LINEAR: Linear,
  MICROSOFT: Microsoft,
  NOTION: Notion,
  ONE_PASSWORD: OnePassword,
  ONEPASSWORD: OnePassword,
  OPENAI: OpenAI,
  RESEND: Resend,
  SENTRY: Sentry,
  SLACK: Slack,
  SUPABASE: Supabase,
  TALLY: Tally,
};

type VendorLogoProps = ComponentProps<"svg"> & {
  /** The vendor/brand name (case-insensitive, supports enum values like GOOGLE_WORKSPACE). */
  vendor: string;
  /** When true, renders the SVG in monochrome, adapting to the current theme. */
  tint?: boolean;
};

export function VendorLogo({ vendor, tint, ...props }: VendorLogoProps) {
  const Component = vendors[vendor.toUpperCase()];
  if (!Component) return null;

  if (tint) {
    return (
      <Component
        {...props}
        className={["grayscale brightness-0 dark:invert", props.className].filter(Boolean).join(" ")}
      />
    );
  }

  return <Component {...props} />;
}
