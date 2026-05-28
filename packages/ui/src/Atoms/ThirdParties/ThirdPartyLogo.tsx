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

import { Anthropic } from "./Anthropic";
import { Asana } from "./Asana";
import { Bitbucket } from "./Bitbucket";
import { Brex } from "./Brex";
import { ClickUp } from "./ClickUp";
import { Cloudflare } from "./Cloudflare";
import { Cursor } from "./Cursor";
import { DocuSign } from "./DocuSign";
import { Figma } from "./Figma";
import { GitHub } from "./GitHub";
import { GitLab } from "./GitLab";
import { Google } from "./Google";
import { Heroku } from "./Heroku";
import { HubSpot } from "./HubSpot";
import { Intercom } from "./Intercom";
import { Linear } from "./Linear";
import { Microsoft } from "./Microsoft";
import { Monday } from "./Monday";
import { Netlify } from "./Netlify";
import { Notion } from "./Notion";
import { OnePassword } from "./OnePassword";
import { OpenAI } from "./OpenAI";
import { PagerDuty } from "./PagerDuty";
import { Resend } from "./Resend";
import { Sentry } from "./Sentry";
import { Slack } from "./Slack";
import { Supabase } from "./Supabase";
import { Tailscale } from "./Tailscale";
import { Tally } from "./Tally";
import { Vercel } from "./Vercel";

const thirdParties: Record<string, FC<ComponentProps<"svg">>> = {
  ANTHROPIC: Anthropic,
  ASANA: Asana,
  BITBUCKET: Bitbucket,
  BREX: Brex,
  CLICKUP: ClickUp,
  CLOUDFLARE: Cloudflare,
  CURSOR: Cursor,
  DOCUSIGN: DocuSign,
  FIGMA: Figma,
  GITHUB: GitHub,
  GITLAB: GitLab,
  GOOGLE: Google,
  GOOGLE_WORKSPACE: Google,
  HEROKU: Heroku,
  HUBSPOT: HubSpot,
  INTERCOM: Intercom,
  LINEAR: Linear,
  MICROSOFT: Microsoft,
  MICROSOFT_365: Microsoft,
  MONDAY: Monday,
  NETLIFY: Netlify,
  NOTION: Notion,
  ONE_PASSWORD: OnePassword,
  ONEPASSWORD: OnePassword,
  OPENAI: OpenAI,
  PAGERDUTY: PagerDuty,
  RESEND: Resend,
  SENTRY: Sentry,
  SLACK: Slack,
  SUPABASE: Supabase,
  TAILSCALE: Tailscale,
  TALLY: Tally,
  VERCEL: Vercel,
};

type ThirdPartyLogoProps = ComponentProps<"svg"> & {
  /** The thirdParty/brand name (case-insensitive, supports enum values like GOOGLE_WORKSPACE). */
  thirdParty: string;
  /** When true, renders the SVG in monochrome, adapting to the current theme. */
  tint?: boolean;
};

export function ThirdPartyLogo({ thirdParty, tint, ...props }: ThirdPartyLogoProps) {
  const Component = thirdParties[thirdParty.toUpperCase()];
  if (!Component) return null;

  if (tint) {
    return (
      <Component
        {...props}
        className={["grayscale brightness-0 dark:invert", props.className]
          .filter(Boolean)
          .join(" ")}
      />
    );
  }

  return <Component {...props} />;
}
