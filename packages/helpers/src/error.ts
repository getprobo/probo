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

export interface GraphQLError {
  message?: string;
  extensions?: {
    code?: string;
  };
  source?: {
    errors?: Array<{ message: string; extensions?: { code?: string } }>;
  };
}

export function formatError(title: string, error: GraphQLError | GraphQLError[]): string {
  const messages: string[] = [];

  if (Array.isArray(error)) {
    messages.push(...error.map((e) => e.message).filter(Boolean) as string[]);
  } else if (error.source?.errors && Array.isArray(error.source.errors)) {
    messages.push(...error.source.errors.map((e) => e.message).filter(Boolean));
  } else if (error.message) {
    messages.push(error.message);
  }

  if (messages.length === 0) {
    return title;
  }

  const errorList = messages.join(", ");

  return `${title}: ${errorList}${errorList.endsWith('.') ? '' : '.'}`;
}
