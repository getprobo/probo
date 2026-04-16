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

import {
  ApiError,
  BadRequestError,
  InternalServerError,
  NetworkError,
  NotFoundError,
  TimeoutError,
} from "./errors";

const DEFAULT_TIMEOUT_MS = 5_000;
const MAX_ATTEMPTS = 3;
const BASE_DELAY_MS = 500;

export interface RequestOptions {
  method?: string;
  headers?: Record<string, string>;
  body?: unknown;
  timeout?: number;
  signal?: AbortSignal;
}

interface ApiErrorBody {
  error: string;
  message: string;
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

function throwApiError(status: number, body: ApiErrorBody): never {
  switch (status) {
    case 400:
      throw new BadRequestError(body.message);
    case 404:
      throw new NotFoundError(body.message);
    case 500:
      throw new InternalServerError(body.message);
    default:
      throw new ApiError(status, body.error, body.message);
  }
}

async function parseErrorBody(response: Response): Promise<ApiErrorBody> {
  try {
    return (await response.json()) as ApiErrorBody;
  } catch {
    return {
      error: "unknown",
      message: response.statusText || "unknown error",
    };
  }
}

async function fetchWithTimeout(
  url: URL | string,
  init: RequestInit,
  timeout: number,
): Promise<Response> {
  const controller = new AbortController();
  const externalSignal = init.signal;

  if (externalSignal?.aborted) {
    throw new NetworkError("request aborted", externalSignal.reason);
  }

  const onExternalAbort = () => controller.abort(externalSignal!.reason);
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
  url: URL | string,
  options: RequestOptions = {},
): Promise<T> {
  const { method = "GET", headers, body, timeout = DEFAULT_TIMEOUT_MS, signal } = options;

  const init: RequestInit = {
    method,
    mode: "cors",
    credentials: "omit",
    headers: {
      Accept: "application/json",
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
      if (signal?.aborted) {
        throw err;
      }
      if (err instanceof TimeoutError || err instanceof NetworkError) {
        lastError = err;
        continue;
      }
      throw err;
    }

    if (response.ok) {
      if (response.status === 204) {
        return undefined as T;
      }
      return (await response.json()) as T;
    }

    if (!isRetryable(response.status)) {
      const body = await parseErrorBody(response);
      throwApiError(response.status, body);
    }

    lastError = new ApiError(
      response.status,
      "server_error",
      `server returned ${response.status}`,
    );
  }

  throw lastError!;
}
