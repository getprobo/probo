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

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ProboDocuments } from "../client";
import type {
  AcceptanceResult,
  AcceptanceStatus,
  AcceptanceStatusInfo,
  DocumentContent,
  ProboSignOptions,
} from "../types";

// Re-export everything from the headless entry for tree-shake-safe imports.
export { ProboDocuments, createSign } from "../client";
export {
  DocumentNotLoadedError,
  IdentityTokenMissingError,
  NetworkError,
  SignApiError,
  TimeoutError,
} from "../errors";
export type {
  AcceptanceResult,
  AcceptanceStatus,
  AcceptanceStatusInfo,
  DocumentContent,
  ErrorCode,
  IdentityTokenClaims,
  ProboSignOptions,
  SignClientEvents,
  SignEventHandler,
  SigningConfigKind,
  SigningConfigMode,
} from "../types";

// ---------------------------------------------------------------------------
// useSign hook
// ---------------------------------------------------------------------------

export interface UseSignState {
  statuses: Record<string, AcceptanceStatusInfo>;
  loading: boolean;
  error: Error | null;
  accept: (signingConfigIds: string[]) => Promise<AcceptanceResult[]>;
}

/**
 * Hook that checks acceptance status for one or more signing configs and
 * returns a stable accept() function.
 */
export function useSign(
  signingConfigIds: string[],
  options: ProboSignOptions,
): UseSignState {
  const clientRef = useRef<ProboDocuments | null>(null);
  const [statuses, setStatuses] = useState<Record<string, AcceptanceStatusInfo>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const idsKey = signingConfigIds.join(",");
  const optionsRef = useRef(options);
  optionsRef.current = options;

  useEffect(() => {
    if (!clientRef.current) {
      clientRef.current = new ProboDocuments(optionsRef.current);
    }
  }, []);

  useEffect(() => {
    if (!signingConfigIds.length || !options.identityToken) return;

    const client = new ProboDocuments(options);
    clientRef.current = client;

    setLoading(true);
    setError(null);

    client
      .status(signingConfigIds)
      .then((s) => setStatuses(s))
      .catch((err: unknown) => setError(err instanceof Error ? err : new Error(String(err))))
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [idsKey, options.publishableKey, options.identityToken]);

  const accept = useCallback(
    async (ids: string[]): Promise<AcceptanceResult[]> => {
      if (!clientRef.current) {
        clientRef.current = new ProboDocuments(optionsRef.current);
      }
      return clientRef.current.accept(ids);
    },
    [],
  );

  return useMemo(
    () => ({ statuses, loading, error, accept }),
    [statuses, loading, error, accept],
  );
}

// ---------------------------------------------------------------------------
// SignCheckbox component
// ---------------------------------------------------------------------------

export interface SignCheckboxProps {
  /** Publishable key (pk_live_... or pk_test_...) for this signing config. */
  publishableKey: string;
  /** Identity token minted by your backend. Required to call accept(). */
  identityToken?: string;
  /** Override the API base URL. Defaults to the current origin. */
  baseUrl?: string;
  /** Override the label text. Defaults to the document's consent_text. */
  label?: string;
  /** Called with acceptance results after the user checks the box. */
  onAccepted?: (results: AcceptanceResult[]) => void;
  /** Called when an error occurs. */
  onError?: (error: Error) => void;
  /** Additional class name for the wrapper element. */
  className?: string;
}

type CheckboxState = "loading" | "ready" | "submitting" | "accepted" | "error";

/**
 * A self-contained consent checkbox that loads the document, renders the
 * consent_text label, and records acceptance when the user checks the box.
 */
export function SignCheckbox({
  publishableKey,
  identityToken,
  baseUrl,
  label,
  onAccepted,
  onError,
  className,
}: SignCheckboxProps): React.ReactElement {
  const client = useMemo(
    () => new ProboDocuments({ publishableKey, identityToken, baseUrl }),
    [publishableKey, identityToken, baseUrl],
  );

  const [doc, setDoc] = useState<DocumentContent | null>(null);
  const [checked, setChecked] = useState(false);
  const [state, setState] = useState<CheckboxState>("loading");

  useEffect(() => {
    setState("loading");
    client
      .loadDocument(publishableKey)
      .then((d) => {
        setDoc(d);
        setState("ready");
      })
      .catch((err: unknown) => {
        setState("error");
        onError?.(err instanceof Error ? err : new Error(String(err)));
      });
  }, [client, publishableKey, onError]);

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>): void => {
      const next = e.target.checked;
      setChecked(next);
      if (!next) return;

      setState("submitting");
      client
        .accept([publishableKey])
        .then((results) => {
          setState("accepted");
          onAccepted?.(results);
        })
        .catch((err: unknown) => {
          setChecked(false);
          setState("ready");
          onError?.(err instanceof Error ? err : new Error(String(err)));
        });
    },
    [client, publishableKey, onAccepted, onError],
  );

  const isDisabled = state === "loading" || state === "submitting";
  const displayLabel = label ?? doc?.consentText ?? "Loading…";
  const checkboxId = useMemo(
    () => `probo-sign-${Math.random().toString(36).slice(2)}`,
    [],
  );

  return (
    <div
      data-probo-sign="checkbox"
      style={{ display: "flex", alignItems: "flex-start", gap: 8 }}
      className={className}
    >
      <input
        id={checkboxId}
        type="checkbox"
        checked={checked}
        disabled={isDisabled}
        onChange={handleChange}
        style={{ marginTop: 3, flexShrink: 0, cursor: isDisabled ? "default" : "pointer" }}
      />
      <label
        htmlFor={checkboxId}
        style={{ lineHeight: 1.4, cursor: isDisabled ? "default" : "pointer" }}
      >
        {displayLabel}
        {doc?.pdfUrl && (
          <>
            {" "}
            <a
              href={doc.pdfUrl}
              target="_blank"
              rel="noopener noreferrer"
              style={{ whiteSpace: "nowrap" }}
              onClick={(e) => e.stopPropagation()}
            >
              View document
            </a>
          </>
        )}
      </label>
    </div>
  );
}

// ---------------------------------------------------------------------------
// AcceptanceStatus helpers
// ---------------------------------------------------------------------------

export function isAccepted(status: AcceptanceStatus): boolean {
  return status === "ACCEPTED";
}

export function requiresAcceptance(status: AcceptanceStatus): boolean {
  return status === "REQUIRED" || status === "REACCEPTANCE_REQUIRED";
}
