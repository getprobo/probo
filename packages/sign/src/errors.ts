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

/** Error returned by the data plane API. Carries a typed error code. */
export class SignApiError extends Error {
  readonly status: number;
  readonly code: ErrorCode;
  readonly requestId?: string;

  constructor(status: number, code: ErrorCode, message: string, requestId?: string) {
    super(message);
    this.name = "SignApiError";
    this.status = status;
    this.code = code;
    this.requestId = requestId;
  }
}

export class NetworkError extends Error {
  readonly cause?: unknown;

  constructor(message: string, cause?: unknown) {
    super(message);
    this.name = "NetworkError";
    this.cause = cause;
  }
}

export class TimeoutError extends NetworkError {
  constructor() {
    super("request timed out");
    this.name = "TimeoutError";
  }
}

export class IdentityTokenMissingError extends Error {
  constructor() {
    super(
      "An identity token is required to record acceptance. " +
        "Pass identityToken to createSign() or ProboDocuments.",
    );
    this.name = "IdentityTokenMissingError";
  }
}

export class DocumentNotLoadedError extends Error {
  constructor(signingConfigId: string) {
    super(
      `Document '${signingConfigId}' has not been loaded. ` +
        "Call loadDocument() before accept().",
    );
    this.name = "DocumentNotLoadedError";
  }
}
