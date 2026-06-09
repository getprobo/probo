// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { BetterStack } from "./BetterStack";
import { Bitbucket } from "./Bitbucket";
import { Brex } from "./Brex";
import { Clerk } from "./Clerk";
import { ClickUp } from "./ClickUp";
import { Cloudflare } from "./Cloudflare";
import { Cursor } from "./Cursor";
import { Datadog } from "./Datadog";
import { DocuSign } from "./DocuSign";
import { Figma } from "./Figma";
import { GitHub } from "./GitHub";
import { GitLab } from "./GitLab";
import { Google } from "./Google";
import { Grafana } from "./Grafana";
import { Heroku } from "./Heroku";
import { HubSpot } from "./HubSpot";
import { Intercom } from "./Intercom";
import { Linear } from "./Linear";
import { Metabase } from "./Metabase";
import { Microsoft } from "./Microsoft";
import { Monday } from "./Monday";
import { Neon } from "./Neon";
import { Netlify } from "./Netlify";
import { Notion } from "./Notion";
import { Okta } from "./Okta";
import { OnePassword } from "./OnePassword";
import { OpenAI } from "./OpenAI";
import { PagerDuty } from "./PagerDuty";
import { PostHog } from "./PostHog";
import { Qovery } from "./Qovery";
import { Render } from "./Render";
import { Resend } from "./Resend";
import { SendGrid } from "./SendGrid";
import { Sentry } from "./Sentry";
import { SigNoz } from "./SigNoz";
import { Slack } from "./Slack";
import { Supabase } from "./Supabase";
import { Tailscale } from "./Tailscale";
import { Tally } from "./Tally";
import { Vercel } from "./Vercel";
import { Zendesk } from "./Zendesk";

const thirdParties: Record<string, FC<ComponentProps<"svg">>> = {
  ANTHROPIC: Anthropic,
  ASANA: Asana,
  BETTER_STACK: BetterStack,
  BITBUCKET: Bitbucket,
  BREX: Brex,
  CLERK: Clerk,
  CLICKUP: ClickUp,
  CLOUDFLARE: Cloudflare,
  CURSOR: Cursor,
  DATADOG: Datadog,
  DOCUSIGN: DocuSign,
  FIGMA: Figma,
  GITHUB: GitHub,
  GITLAB: GitLab,
  GOOGLE: Google,
  GOOGLE_WORKSPACE: Google,
  GRAFANA: Grafana,
  HEROKU: Heroku,
  HUBSPOT: HubSpot,
  INTERCOM: Intercom,
  LINEAR: Linear,
  METABASE: Metabase,
  MICROSOFT: Microsoft,
  MICROSOFT_365: Microsoft,
  MONDAY: Monday,
  NEON: Neon,
  NETLIFY: Netlify,
  NOTION: Notion,
  OKTA: Okta,
  ONE_PASSWORD: OnePassword,
  ONEPASSWORD: OnePassword,
  OPENAI: OpenAI,
  PAGERDUTY: PagerDuty,
  POSTHOG: PostHog,
  QOVERY: Qovery,
  RENDER: Render,
  RESEND: Resend,
  SENDGRID: SendGrid,
  SENTRY: Sentry,
  SIGNOZ: SigNoz,
  SLACK: Slack,
  SUPABASE: Supabase,
  TAILSCALE: Tailscale,
  TALLY: Tally,
  VERCEL: Vercel,
  ZENDESK: Zendesk,
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
