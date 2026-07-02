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

import type { PropsWithChildren } from "react";
import { Outlet } from "react-router";

import { Skeleton } from "../Atoms/Skeleton/Skeleton";

export function CenteredLayout({ children }: PropsWithChildren) {
  return (
    <div className="grid place-items-center min-h-[calc(100vh-48px-(2*48px))] text-txt-primary bg-level-0">
      <div className="w-full max-w-2xl flex flex-col items-center">
        {children ?? <Outlet />}
      </div>
    </div>
  );
}

export function CenteredLayoutSkeleton() {
  return (
    <CenteredLayout>
      <div className="w-full max-w-2xl flex flex-col items-center space-y-6">
        <Skeleton className="w-77 h-9" />
        <Skeleton className="w-full h-20" />
      </div>
    </CenteredLayout>
  );
}
