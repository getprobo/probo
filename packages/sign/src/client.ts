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

import {
  DocumentNotLoadedError,
  IdentityTokenMissingError,
} from "./errors";
import { fetchJSON } from "./http";
import type {
  AcceptanceResult,
  AcceptanceStatusInfo,
  DocumentContent,
  IdentityTokenClaims,
  ProboSignOptions,
  SignClientEvents,
  SignEventHandler,
} from "./types";

const API_BASE = "/api/sign/v1";

interface RawDocumentContent {
  document_version_id: string;
  major: number;
  minor: number;
  content_html: string;
  content_sha256: string;
  pdf_url: string;
  consent_text: string;
  kind: string;
}

interface RawStatusResponse {
  subject_id?: string;
  external_user_id: string;
  status: string;
  accepted_version_id?: string;
  accepted_at?: string;
  current_version_id: string;
  current_version_major: number;
  current_version_minor: number;
}

interface RawBatchStatusResponse {
  [signingConfigId: string]: RawStatusResponse;
}

interface RawBatchAcceptanceResponse {
  acceptances: Array<{
    acceptance_id: string;
    subject_id: string;
    signing_config_id: string;
    document_version_id: string;
    accepted_at: string;
    mode: string;
    status: string;
  }>;
}

function decodeJwtPayload(token: string): IdentityTokenClaims {
  const parts = token.split(".");
  if (parts.length !== 3) {
    throw new Error("invalid identity token format");
  }
  const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
  const json = atob(payload);
  return JSON.parse(json) as IdentityTokenClaims;
}

function normalizeBase(url: string): string {
  if (!url.endsWith("/")) return url + "/";
  return url;
}

function resolveBase(options: ProboSignOptions): string {
  if (options.baseUrl) return normalizeBase(options.baseUrl);
  if (typeof window !== "undefined") {
    return normalizeBase(window.location.origin);
  }
  return "/";
}

type ListenerMap = {
  [K in keyof SignClientEvents]: Array<SignEventHandler<K>>;
};

export class ProboDocuments {
  private readonly publishableKey: string;
  private readonly identityToken: string | undefined;
  private readonly baseUrl: string;
  private readonly docCache = new Map<string, DocumentContent>();
  private readonly listeners: ListenerMap = { accepted: [] };

  constructor(options: ProboSignOptions) {
    this.publishableKey = options.publishableKey;
    this.identityToken = options.identityToken;
    this.baseUrl = resolveBase(options);
  }

  on<K extends keyof SignClientEvents>(
    event: K,
    handler: SignEventHandler<K>,
  ): this {
    (this.listeners[event] as Array<SignEventHandler<K>>).push(handler);
    return this;
  }

  off<K extends keyof SignClientEvents>(
    event: K,
    handler: SignEventHandler<K>,
  ): this {
    const arr = this.listeners[event] as Array<SignEventHandler<K>>;
    const idx = arr.indexOf(handler);
    if (idx !== -1) arr.splice(idx, 1);
    return this;
  }

  private emit<K extends keyof SignClientEvents>(
    event: K,
    data: SignClientEvents[K],
  ): void {
    for (const h of this.listeners[event] as Array<SignEventHandler<K>>) {
      h(data);
    }
  }

  private authHeader(): Record<string, string> {
    return { Authorization: `Bearer ${this.publishableKey}` };
  }

  /**
   * Fetch the current PRESENTABLE version of a document.
   * Result is cached locally so accept() can use it without re-fetching.
   */
  async loadDocument(signingConfigId: string): Promise<DocumentContent> {
    const url = `${this.baseUrl}${API_BASE.slice(1)}/documents/${signingConfigId}/content`;
    const raw = await fetchJSON<RawDocumentContent>(url, {
      headers: this.authHeader(),
    });

    const doc: DocumentContent = {
      documentVersionId: raw.document_version_id,
      major: raw.major,
      minor: raw.minor,
      contentHtml: raw.content_html,
      contentSha256: raw.content_sha256,
      pdfUrl: raw.pdf_url,
      consentText: raw.consent_text,
      kind: raw.kind as DocumentContent["kind"],
    };

    this.docCache.set(signingConfigId, doc);
    return doc;
  }

  /**
   * Check acceptance status for one or more signing configs.
   * Requires an identity token to identify the subject.
   */
  async status(
    signingConfigIds: string[],
  ): Promise<Record<string, AcceptanceStatusInfo>> {
    const claims = this.requireIdentityClaims();
    const url = `${this.baseUrl}${API_BASE.slice(1)}/subjects/${encodeURIComponent(claims.sub)}/status`;
    const raw = await fetchJSON<RawBatchStatusResponse>(url, {
      method: "POST",
      headers: this.authHeader(),
      body: { signing_config_ids: signingConfigIds },
    });

    const result: Record<string, AcceptanceStatusInfo> = {};
    for (const [id, s] of Object.entries(raw)) {
      result[id] = {
        subjectId: s.subject_id,
        externalUserId: s.external_user_id,
        status: s.status as AcceptanceStatusInfo["status"],
        acceptedVersionId: s.accepted_version_id,
        acceptedAt: s.accepted_at,
        currentVersionId: s.current_version_id,
        currentVersionMajor: s.current_version_major,
        currentVersionMinor: s.current_version_minor,
      };
    }
    return result;
  }

  /**
   * Batch-accept one or more signing configs atomically.
   * loadDocument() must have been called for each config beforehand.
   * Requires an identity token.
   */
  async accept(signingConfigIds: string[]): Promise<AcceptanceResult[]> {
    const claims = this.requireIdentityClaims();

    const acceptances = signingConfigIds.map((id) => {
      const doc = this.docCache.get(id);
      if (!doc) throw new DocumentNotLoadedError(id);
      return {
        signing_config_id: id,
        document_version_id: doc.documentVersionId,
        consent_text: doc.consentText,
      };
    });

    const url = `${this.baseUrl}${API_BASE.slice(1)}/acceptances/batch`;
    const raw = await fetchJSON<RawBatchAcceptanceResponse>(url, {
      method: "POST",
      headers: this.authHeader(),
      body: {
        external_user_id: claims.sub,
        email: claims.email,
        ...(claims.name && { full_name: claims.name }),
        ...(claims.account_id && { account_id: claims.account_id }),
        user_agent: typeof navigator !== "undefined" ? navigator.userAgent : "",
        identity_jwt: this.identityToken,
        acceptances,
      },
    });

    const results: AcceptanceResult[] = raw.acceptances.map((a) => ({
      acceptanceId: a.acceptance_id,
      subjectId: a.subject_id,
      signingConfigId: a.signing_config_id,
      documentVersionId: a.document_version_id,
      acceptedAt: a.accepted_at,
      mode: a.mode as AcceptanceResult["mode"],
      status: "pending_seal",
    }));

    this.emit("accepted", results);
    return results;
  }

  private requireIdentityClaims(): IdentityTokenClaims {
    if (!this.identityToken) throw new IdentityTokenMissingError();
    return decodeJwtPayload(this.identityToken);
  }
}

/** Factory function alternative to `new ProboDocuments(...)`. */
export function createSign(options: ProboSignOptions): ProboDocuments {
  return new ProboDocuments(options);
}
