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

import type { ComponentProps, FC } from "react";

import { Anthropic } from "./Anthropic";
import { Apollo } from "./Apollo";
import { Asana } from "./Asana";
import { BetterStack } from "./BetterStack";
import { Bitbucket } from "./Bitbucket";
import { Brevo } from "./Brevo";
import { Brex } from "./Brex";
import { Clerk } from "./Clerk";
import { ClickHouse } from "./ClickHouse";
import { ClickUp } from "./ClickUp";
import { Cloudflare } from "./Cloudflare";
import { Crisp } from "./Crisp";
import { Cursor } from "./Cursor";
import { Datadog } from "./Datadog";
import { Deepgram } from "./Deepgram";
import { DocuSign } from "./DocuSign";
import { Dotfile } from "./Dotfile";
import { Figma } from "./Figma";
import { GitHub } from "./GitHub";
import { GitLab } from "./GitLab";
import { Google } from "./Google";
import { GoogleAnalytics } from "./GoogleAnalytics";
import { Grafana } from "./Grafana";
import { Heroku } from "./Heroku";
import { HubSpot } from "./HubSpot";
import { IncidentIO } from "./IncidentIO";
import { Intercom } from "./Intercom";
import { Langfuse } from "./Langfuse";
import { Linear } from "./Linear";
import { Mercury } from "./Mercury";
import { Metabase } from "./Metabase";
import { Microsoft } from "./Microsoft";
import { Monday } from "./Monday";
import { Neon } from "./Neon";
import { Netlify } from "./Netlify";
import { Notion } from "./Notion";
import { Okta } from "./Okta";
import { OnePassword } from "./OnePassword";
import { OpenAI } from "./OpenAI";
import { OpenRouter } from "./OpenRouter";
import { PagerDuty } from "./PagerDuty";
import { PostHog } from "./PostHog";
import { Pylon } from "./Pylon";
import { Qovery } from "./Qovery";
import { Railway } from "./Railway";
import { Render } from "./Render";
import { Resend } from "./Resend";
import { Scaleway } from "./Scaleway";
import { Segment } from "./Segment";
import { SendGrid } from "./SendGrid";
import { Sentry } from "./Sentry";
import { SigNoz } from "./SigNoz";
import { Slack } from "./Slack";
import { Square } from "./Square";
import { Supabase } from "./Supabase";
import { Tailscale } from "./Tailscale";
import { Tally } from "./Tally";
import { Vercel } from "./Vercel";
import { Yousign } from "./Yousign";
import { Zendesk } from "./Zendesk";

const thirdParties: Record<string, FC<ComponentProps<"svg">>> = {
  ANTHROPIC: Anthropic,
  APOLLO: Apollo,
  ASANA: Asana,
  BETTER_STACK: BetterStack,
  BITBUCKET: Bitbucket,
  BREVO: Brevo,
  BREX: Brex,
  CLERK: Clerk,
  CLICKHOUSE: ClickHouse,
  CLICKUP: ClickUp,
  CLOUDFLARE: Cloudflare,
  CRISP: Crisp,
  CURSOR: Cursor,
  DATADOG: Datadog,
  DEEPGRAM: Deepgram,
  DOCUSIGN: DocuSign,
  DOTFILE: Dotfile,
  FIGMA: Figma,
  GITHUB: GitHub,
  GITLAB: GitLab,
  GOOGLE: Google,
  GOOGLE_ANALYTICS: GoogleAnalytics,
  GOOGLE_WORKSPACE: Google,
  GRAFANA: Grafana,
  HEROKU: Heroku,
  HUBSPOT: HubSpot,
  INCIDENT_IO: IncidentIO,
  INTERCOM: Intercom,
  LANGFUSE: Langfuse,
  LINEAR: Linear,
  MERCURY: Mercury,
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
  OPENROUTER: OpenRouter,
  PAGERDUTY: PagerDuty,
  POSTHOG: PostHog,
  PYLON: Pylon,
  QOVERY: Qovery,
  RAILWAY: Railway,
  RENDER: Render,
  RESEND: Resend,
  SCALEWAY: Scaleway,
  SEGMENT: Segment,
  SENDGRID: SendGrid,
  SENTRY: Sentry,
  SIGNOZ: SigNoz,
  SLACK: Slack,
  SQUARE: Square,
  SUPABASE: Supabase,
  TAILSCALE: Tailscale,
  TALLY: Tally,
  VERCEL: Vercel,
  YOUSIGN: Yousign,
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
