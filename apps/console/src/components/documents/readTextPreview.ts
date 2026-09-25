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

export interface TextPreviewResult {
  content: string;
  truncated: boolean;
}

export async function readTextPreview(
  body: ReadableStream<Uint8Array>,
  maxBytes: number,
): Promise<TextPreviewResult> {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let content = "";
  let loadedBytes = 0;
  let truncated = false;

  while (true) {
    const result = await reader.read();
    if (result.done) {
      break;
    }

    const remainingBytes = maxBytes - loadedBytes;
    if (result.value.byteLength > remainingBytes) {
      content += decoder.decode(
        result.value.subarray(0, remainingBytes),
        { stream: true },
      );
      truncated = true;
      await reader.cancel();
      break;
    }

    content += decoder.decode(result.value, { stream: true });
    loadedBytes += result.value.byteLength;
  }

  content += decoder.decode();
  if (truncated && content.endsWith("\uFFFD")) {
    content = content.slice(0, -1);
  }

  return { content, truncated };
}
