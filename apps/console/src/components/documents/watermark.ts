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

import { z } from "zod";

export const maxWatermarkTextLength = 64;

const textEncoder = new TextEncoder();

export const exportSchema = z.object({
  withWatermark: z.boolean(),
  watermarkText: z.string().optional(),
  withSignatures: z.boolean(),
}).refine(
  data => !data.withWatermark || !!data.watermarkText?.trim(),
  {
    message: "Please enter watermark text",
    path: ["watermarkText"],
  },
).refine(
  data => !data.withWatermark
    || getWatermarkTextLength(data.watermarkText ?? "") <= maxWatermarkTextLength,
  {
    message: `Watermark text must not exceed ${maxWatermarkTextLength} bytes`,
    path: ["watermarkText"],
  },
);

export type ExportFormData = z.infer<typeof exportSchema>;

export function getWatermarkTextLength(watermarkText: string): number {
  return textEncoder.encode(watermarkText).length;
}

export function truncateWatermarkText(watermarkText: string): string {
  let truncated = "";

  for (const character of watermarkText) {
    const candidate = truncated + character;
    if (getWatermarkTextLength(candidate) > maxWatermarkTextLength) {
      break;
    }

    truncated = candidate;
  }

  return truncated;
}
