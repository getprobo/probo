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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { useNavigate } from "react-router";

export default function MagicLinkExpiredPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();

  usePageTitle(__("Link Expired"));

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{__("Link Expired")}</h1>
        <p className="text-txt-tertiary">
          {__(
            "This magic link has expired. Magic links are only valid for 15 minutes. Please request a new one.",
          )}
        </p>
      </div>
      <div className="flex justify-center">
        <Button
          className="w-xs h-10"
          onClick={() => void navigate("/connect")}
        >
          {__("Request a new link")}
        </Button>
      </div>
    </div>
  );
}
