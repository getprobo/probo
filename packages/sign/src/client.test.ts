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

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createSign, ProboDocuments } from "./client";
import {
  DocumentNotLoadedError,
  IdentityTokenMissingError,
  SignApiError,
} from "./errors";

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

const PK = "pk_live_testkey123";
const BASE = "https://app.example.com";

/** Minimal identity JWT (unsigned — the browser only base64-decodes it). */
function makeIdentityToken(overrides: Record<string, unknown> = {}): string {
  const payload = {
    sub: "u_1234",
    email: "alice@example.com",
    name: "Alice Smith",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 600,
    ...overrides,
  };
  const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }))
    .replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
  const body = btoa(JSON.stringify(payload))
    .replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
  return `${header}.${body}.fakesig`;
}

const DOC_RESPONSE = {
  document_version_id: "dv_abc",
  major: 2,
  minor: 0,
  content_html: "<p>Terms</p>",
  content_sha256: "deadbeef",
  pdf_url: "https://cdn.example.com/doc.pdf",
  consent_text: "By checking this box you agree.",
  kind: "TERMS_OF_SERVICE",
};

const ACCEPTANCE_RESPONSE = {
  acceptances: [
    {
      acceptance_id: "acc_xyz",
      subject_id: "sub_1",
      signing_config_id: PK,
      document_version_id: "dv_abc",
      accepted_at: "2026-11-20T10:00:00Z",
      mode: "LIVE",
      status: "pending_seal",
    },
  ],
};

const STATUS_RESPONSE: Record<string, unknown> = {
  [PK]: {
    subject_id: "sub_1",
    external_user_id: "u_1234",
    status: "REQUIRED",
    current_version_id: "dv_abc",
    current_version_major: 2,
    current_version_minor: 0,
  },
};

function mockFetch(handler: (url: string, init: RequestInit) => Promise<Response>): void {
  vi.stubGlobal("fetch", handler);
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function errorResponse(code: string, message: string, status = 400): Response {
  return jsonResponse({ error: { code, message, request_id: "req_1" } }, status);
}

// ---------------------------------------------------------------------------
// createSign
// ---------------------------------------------------------------------------

describe("createSign", () => {
  it("returns a ProboDocuments instance", () => {
    expect(createSign({ publishableKey: PK, baseUrl: BASE })).toBeInstanceOf(ProboDocuments);
  });
});

// ---------------------------------------------------------------------------
// loadDocument
// ---------------------------------------------------------------------------

describe("ProboDocuments.loadDocument", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("fetches document content and returns camelCase shape", async () => {
    mockFetch(async () => jsonResponse(DOC_RESPONSE));
    const client = new ProboDocuments({ publishableKey: PK, baseUrl: BASE });
    const doc = await client.loadDocument(PK);

    expect(doc.documentVersionId).toBe("dv_abc");
    expect(doc.contentHtml).toBe("<p>Terms</p>");
    expect(doc.pdfUrl).toBe("https://cdn.example.com/doc.pdf");
    expect(doc.consentText).toBe("By checking this box you agree.");
    expect(doc.kind).toBe("TERMS_OF_SERVICE");
  });

  it("sends Authorization header with the publishable key", async () => {
    let capturedHeaders: Record<string, string> = {};
    mockFetch(async (_, init) => {
      capturedHeaders = Object.fromEntries(
        new Headers(init.headers as HeadersInit).entries(),
      );
      return jsonResponse(DOC_RESPONSE);
    });

    const client = new ProboDocuments({ publishableKey: PK, baseUrl: BASE });
    await client.loadDocument(PK);
    expect(capturedHeaders["authorization"]).toBe(`Bearer ${PK}`);
  });

  it("throws SignApiError on 404", async () => {
    mockFetch(async () => errorResponse("document_not_presentable", "No PDF yet", 404));
    const client = new ProboDocuments({ publishableKey: PK, baseUrl: BASE });
    await expect(client.loadDocument(PK)).rejects.toBeInstanceOf(SignApiError);
  });
});

// ---------------------------------------------------------------------------
// status
// ---------------------------------------------------------------------------

describe("ProboDocuments.status", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("returns a map of status info keyed by signing config ID", async () => {
    mockFetch(async () => jsonResponse(STATUS_RESPONSE));
    const client = new ProboDocuments({
      publishableKey: PK,
      identityToken: makeIdentityToken(),
      baseUrl: BASE,
    });

    const result = await client.status([PK]);
    expect(result[PK].status).toBe("REQUIRED");
    expect(result[PK].externalUserId).toBe("u_1234");
  });

  it("throws IdentityTokenMissingError when no token provided", async () => {
    const client = new ProboDocuments({ publishableKey: PK, baseUrl: BASE });
    await expect(client.status([PK])).rejects.toBeInstanceOf(IdentityTokenMissingError);
  });

  it("includes external_user_id from the token in the request URL", async () => {
    let capturedUrl = "";
    mockFetch(async (url) => {
      capturedUrl = url;
      return jsonResponse(STATUS_RESPONSE);
    });

    const client = new ProboDocuments({
      publishableKey: PK,
      identityToken: makeIdentityToken({ sub: "u_9999" }),
      baseUrl: BASE,
    });

    await client.status([PK]);
    expect(capturedUrl).toContain("u_9999");
  });
});

// ---------------------------------------------------------------------------
// accept
// ---------------------------------------------------------------------------

describe("ProboDocuments.accept", () => {
  beforeEach(() => {
    vi.stubGlobal("navigator", { userAgent: "TestBrowser/1.0" });
  });
  afterEach(() => vi.unstubAllGlobals());

  it("throws IdentityTokenMissingError when no token provided", async () => {
    const client = new ProboDocuments({ publishableKey: PK, baseUrl: BASE });
    await expect(client.accept([PK])).rejects.toBeInstanceOf(IdentityTokenMissingError);
  });

  it("throws DocumentNotLoadedError when loadDocument was not called first", async () => {
    const client = new ProboDocuments({
      publishableKey: PK,
      identityToken: makeIdentityToken(),
      baseUrl: BASE,
    });
    await expect(client.accept([PK])).rejects.toBeInstanceOf(DocumentNotLoadedError);
  });

  it("posts batch acceptance and returns AcceptanceResult[]", async () => {
    let callCount = 0;
    mockFetch(async (url) => {
      callCount++;
      if (url.includes("/content")) return jsonResponse(DOC_RESPONSE);
      return jsonResponse(ACCEPTANCE_RESPONSE, 201);
    });

    const client = new ProboDocuments({
      publishableKey: PK,
      identityToken: makeIdentityToken(),
      baseUrl: BASE,
    });

    await client.loadDocument(PK);
    const results = await client.accept([PK]);

    expect(callCount).toBe(2);
    expect(results).toHaveLength(1);
    expect(results[0].acceptanceId).toBe("acc_xyz");
    expect(results[0].status).toBe("pending_seal");
  });

  it("fires 'accepted' event after successful accept", async () => {
    mockFetch(async (url) => {
      if (url.includes("/content")) return jsonResponse(DOC_RESPONSE);
      return jsonResponse(ACCEPTANCE_RESPONSE, 201);
    });

    const client = new ProboDocuments({
      publishableKey: PK,
      identityToken: makeIdentityToken(),
      baseUrl: BASE,
    });

    const events: unknown[] = [];
    client.on("accepted", (r) => events.push(r));

    await client.loadDocument(PK);
    await client.accept([PK]);

    expect(events).toHaveLength(1);
  });

  it("includes sub and email from identity token in request body", async () => {
    let body: Record<string, unknown> = {};
    mockFetch(async (url, init) => {
      if (url.includes("/content")) return jsonResponse(DOC_RESPONSE);
      body = JSON.parse(init.body as string) as Record<string, unknown>;
      return jsonResponse(ACCEPTANCE_RESPONSE, 201);
    });

    const client = new ProboDocuments({
      publishableKey: PK,
      identityToken: makeIdentityToken({ sub: "u_xyz", email: "x@test.com" }),
      baseUrl: BASE,
    });

    await client.loadDocument(PK);
    await client.accept([PK]);

    expect(body["external_user_id"]).toBe("u_xyz");
    expect(body["email"]).toBe("x@test.com");
  });

  it("supports off() to remove an event listener", async () => {
    mockFetch(async (url) => {
      if (url.includes("/content")) return jsonResponse(DOC_RESPONSE);
      return jsonResponse(ACCEPTANCE_RESPONSE, 201);
    });

    const client = new ProboDocuments({
      publishableKey: PK,
      identityToken: makeIdentityToken(),
      baseUrl: BASE,
    });

    const calls: number[] = [];
    const handler = (): void => void calls.push(1);
    client.on("accepted", handler);
    client.off("accepted", handler);

    await client.loadDocument(PK);
    await client.accept([PK]);

    expect(calls).toHaveLength(0);
  });
});
