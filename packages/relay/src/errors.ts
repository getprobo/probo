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

export class UnAuthenticatedError extends Error {
  constructor(message?: string) {
    super(message || "UNAUTHENTICATED");
    this.name = "UnAuthenticatedError";
    Object.setPrototypeOf(this, UnAuthenticatedError.prototype);
  }
}

export class FullNameRequiredError extends Error {
  constructor(message?: string) {
    super(message || "FULL_NAME_REQUIRED");
    this.name = "FullNameRequiredError";
    Object.setPrototypeOf(this, FullNameRequiredError.prototype);
  }
}

export class NDASignatureRequiredError extends Error {
  constructor(message?: string) {
    super(message || "NDA_SIGNATURE_REQUIRED");
    this.name = "NDASignatureRequiredError";
    Object.setPrototypeOf(this, NDASignatureRequiredError.prototype);
  }
}

export class InternalServerError extends Error {
  constructor() {
    super("INTERNAL_SERVER_ERROR");
    this.name = "InternalServerError";
    Object.setPrototypeOf(this, InternalServerError.prototype);
  }
}

export class AssumptionRequiredError extends Error {
  constructor(message?: string) {
    super(message ?? "ASSUMPTION_REQUIRED");
    this.name = "AssumptionRequiredError";
    Object.setPrototypeOf(this, AssumptionRequiredError.prototype);
  }
}

export class ForbiddenError extends Error {
  constructor(message?: string) {
    super(message || "FORBIDDEN");
    this.name = "ForbiddenError";
    Object.setPrototypeOf(this, ForbiddenError.prototype);
  }
}
