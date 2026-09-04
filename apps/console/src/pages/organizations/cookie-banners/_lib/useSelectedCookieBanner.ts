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

import { useCallback, useState } from "react";
import { useParams } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";

export interface SelectedCookieBanner {
  id: string;
  name: string;
  tcf: boolean;
}

interface UseSelectedCookieBanner {
  // Banner named by the current route. The URL vouches for this id, so it is
  // worth looking up; anything remembered is not.
  routeId: string | null;
  // Last banner visited in this organization, name included, so leaving a
  // banner page for another privacy page keeps the selection without asking
  // the server who it was again.
  remembered: SelectedCookieBanner | null;
  remember: (banner: SelectedCookieBanner) => void;
}

export function useSelectedCookieBanner(): UseSelectedCookieBanner {
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const [byOrganization, setByOrganization] = useState<
    Record<string, SelectedCookieBanner>
  >({});

  const remember = useCallback((banner: SelectedCookieBanner) => {
    setByOrganization((previous) => {
      const current = previous[organizationId];
      if (current?.id === banner.id && current.name === banner.name && current.tcf === banner.tcf) {
        return previous;
      }
      return { ...previous, [organizationId]: banner };
    });
  }, [organizationId]);

  return {
    routeId: cookieBannerId ?? null,
    remembered: byOrganization[organizationId] ?? null,
    remember,
  };
}
