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

export function ScimOnboardingIllustration() {
  return (
    <svg
      viewBox="0 0 320 180"
      className="h-full w-full max-w-sm text-accent-9"
      role="img"
      aria-hidden
    >
      <rect x="24" y="48" width="72" height="84" rx="8" className="fill-sand-4 stroke-sand-8" strokeWidth="2" />
      <text x="60" y="98" textAnchor="middle" className="fill-sand-11 text-[11px] font-medium">IdP</text>
      <path
        d="M108 90 H148"
        className="stroke-accent-9 motion-safe:animate-pulse"
        strokeWidth="3"
        strokeLinecap="round"
        fill="none"
      />
      <polygon points="148,84 160,90 148,96" className="fill-accent-9" />
      <rect x="168" y="40" width="128" height="100" rx="12" className="fill-accent-3 stroke-accent-8" strokeWidth="2" />
      <text x="232" y="78" textAnchor="middle" className="fill-accent-11 text-[11px] font-medium">Probo</text>
      <circle cx="200" cy="108" r="8" className="fill-green-9 motion-safe:animate-pulse" />
      <circle cx="232" cy="118" r="8" className="fill-green-9 motion-safe:animate-pulse [animation-delay:200ms]" />
      <circle cx="264" cy="108" r="8" className="fill-green-9 motion-safe:animate-pulse [animation-delay:400ms]" />
    </svg>
  );
}

export function AccessReviewOnboardingIllustration() {
  return (
    <svg viewBox="0 0 320 180" className="h-full w-full max-w-sm" aria-hidden>
      <rect x="32" y="56" width="56" height="56" rx="8" className="fill-sand-4 stroke-sand-8" strokeWidth="2" />
      <rect x="132" y="36" width="56" height="56" rx="8" className="fill-sand-4 stroke-sand-8" strokeWidth="2" />
      <rect x="232" y="56" width="56" height="56" rx="8" className="fill-sand-4 stroke-sand-8" strokeWidth="2" />
      <path d="M88 84 L132 64" className="stroke-accent-9 motion-safe:animate-pulse" strokeWidth="2" fill="none" />
      <path d="M188 64 L232 84" className="stroke-accent-9 motion-safe:animate-pulse [animation-delay:150ms]" strokeWidth="2" fill="none" />
      <rect x="112" y="112" width="96" height="48" rx="10" className="fill-accent-3 stroke-accent-9" strokeWidth="2" />
      <text x="160" y="142" textAnchor="middle" className="fill-accent-11 text-[10px] font-medium">Review</text>
    </svg>
  );
}

export function AgentOnboardingIllustration() {
  return (
    <svg viewBox="0 0 320 180" className="h-full w-full max-w-sm" aria-hidden>
      <rect x="100" y="40" width="120" height="80" rx="8" className="fill-sand-3 stroke-sand-8" strokeWidth="2" />
      <rect x="108" y="48" width="104" height="56" rx="4" className="fill-sand-1" />
      <rect x="140" y="128" width="40" height="8" rx="4" className="fill-sand-6" />
      <circle cx="160" cy="100" r="16" className="fill-green-9 motion-safe:animate-pulse" />
      <path d="M154 100 L158 104 L166 94" className="stroke-white" strokeWidth="2" fill="none" strokeLinecap="round" />
    </svg>
  );
}

export function McpOnboardingIllustration() {
  return (
    <svg viewBox="0 0 320 180" className="h-full w-full max-w-sm" aria-hidden>
      <rect x="40" y="50" width="80" height="80" rx="8" className="fill-sand-4 stroke-sand-8" strokeWidth="2" />
      <text x="80" y="96" textAnchor="middle" className="fill-sand-11 text-[10px]">IDE</text>
      <path d="M128 90 H192" className="stroke-accent-9 motion-safe:animate-pulse" strokeWidth="3" fill="none" />
      <rect x="200" y="50" width="80" height="80" rx="8" className="fill-accent-3 stroke-accent-9" strokeWidth="2" />
      <text x="240" y="96" textAnchor="middle" className="fill-accent-11 text-[10px]">MCP</text>
    </svg>
  );
}

export function CongratsOnboardingIllustration() {
  return (
    <svg viewBox="0 0 320 180" className="h-full w-full max-w-sm" aria-hidden>
      <circle cx="160" cy="90" r="48" className="fill-green-3 stroke-green-9" strokeWidth="3" />
      <path
        d="M136 90 L152 106 L184 74"
        className="stroke-green-11 motion-safe:animate-in motion-safe:zoom-in"
        strokeWidth="4"
        fill="none"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
