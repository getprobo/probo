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

export const resourceTypeBadges = {
  SCRIPT: { color: "sky" as const, labelKey: "script", variant: "soft" as const },
  IFRAME: { color: "amber" as const, labelKey: "iframe", variant: "soft" as const },
  IMAGE: { color: "neutral" as const, labelKey: "image", variant: "soft" as const },
  STYLESHEET: { color: "indigo" as const, labelKey: "stylesheet", variant: "soft" as const },
  FONT: { color: "neutral" as const, labelKey: "font", variant: "outline" as const },
  BEACON: { color: "red" as const, labelKey: "beacon", variant: "soft" as const },
  FETCH: { color: "green" as const, labelKey: "fetch", variant: "soft" as const },
  MEDIA: { color: "neutral" as const, labelKey: "media", variant: "soft" as const },
  SERVICE_WORKER: { color: "amber" as const, labelKey: "serviceWorker", variant: "soft" as const },
};
