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

import type { ErrorCode } from "./types";
import { NetworkError, SignApiError, TimeoutError } from "./errors";

const DEFAULT_TIMEOUT_MS = 10_000;
const MAX_ATTEMPTS = 3;
const BASE_DELAY_MS = 500;

export interface FetchOptions {
  method?: string;
  headers?: Record<string, string>;
  body?: unknown;
  timeout?: number;
  signal?: AbortSignal;
}

interface ApiErrorBody {
  error: {
    code: string;
    message: string;
    request_id?: string;
  };
}

function isRetryable(status: number): boolean {
  return status === 429 || status >= 500;
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function jitteredBackoff(attempt: number): number {
  const base = BASE_DELAY_MS * Math.pow(2, attempt);
  return base + Math.random() * base;
}

async function parseErrorBody(response: Response): Promise<ApiErrorBody["error"]> {
  try {
    const body = (await response.json()) as ApiErrorBody;
    return body.error;
  } catch {
    return {
      code: "server_error",
      message: response.statusText || "unknown error",
    };
  }
}

function throwApiError(status: number, err: ApiErrorBody["error"]): never {
  throw new SignApiError(
    status,
    err.code as ErrorCode,
    err.message,
    err.request_id,
  );
}

async function fetchWithTimeout(
  url: string,
  init: RequestInit,
  timeout: number,
): Promise<Response> {
  const controller = new AbortController();
  const externalSignal = init.signal as AbortSignal | undefined;

  if (externalSignal?.aborted) {
    throw new NetworkError("request aborted", externalSignal.reason);
  }

  const onExternalAbort = (): void => controller.abort(externalSignal!.reason);
  externalSignal?.addEventListener("abort", onExternalAbort, { once: true });

  const timer = setTimeout(() => controller.abort("timeout"), timeout);

  try {
    return await fetch(url, { ...init, signal: controller.signal });
  } catch (err) {
    if (controller.signal.aborted && controller.signal.reason === "timeout") {
      throw new TimeoutError();
    }
    if (externalSignal?.aborted) {
      throw new NetworkError("request aborted", err);
    }
    throw new NetworkError("network request failed", err);
  } finally {
    clearTimeout(timer);
    externalSignal?.removeEventListener("abort", onExternalAbort);
  }
}

export async function fetchJSON<T>(
  url: string,
  options: FetchOptions = {},
): Promise<T> {
  const { method = "GET", headers, body, timeout = DEFAULT_TIMEOUT_MS, signal } = options;

  const init: RequestInit = {
    method,
    mode: "cors",
    credentials: "omit",
    headers: {
      Accept: "application/json",
      "X-SDK-Version": __SDK_VERSION__,
      ...(body !== undefined && { "Content-Type": "application/json" }),
      ...headers,
    },
    ...(body !== undefined && { body: JSON.stringify(body) }),
    ...(signal && { signal }),
  };

  let lastError: Error | undefined;

  for (let attempt = 0; attempt < MAX_ATTEMPTS; attempt++) {
    if (attempt > 0) {
      await delay(jitteredBackoff(attempt - 1));
    }

    let response: Response;
    try {
      response = await fetchWithTimeout(url, init, timeout);
    } catch (err) {
      if (signal?.aborted) throw err;
      if (err instanceof TimeoutError || err instanceof NetworkError) {
        lastError = err;
        continue;
      }
      throw err;
    }

    if (response.ok) {
      if (response.status === 204) return undefined as T;
      return (await response.json()) as T;
    }

    if (!isRetryable(response.status)) {
      const errBody = await parseErrorBody(response);
      throwApiError(response.status, errBody);
    }

    lastError = new SignApiError(
      response.status,
      "server_error",
      `server returned ${response.status}`,
    );
  }

  throw lastError!;
}
