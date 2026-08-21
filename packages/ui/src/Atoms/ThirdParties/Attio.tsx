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

import type { ComponentProps } from "react";

export function Attio(props: ComponentProps<"svg">) {
  return (
    <svg
      viewBox="0 0 32 28"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M31.669 19.1747 28.9957 14.8963l-.0159-.0258-.2108-.3362c-.3978-.6384-1.086-1.0203-1.8379-1.0223l-4.3062-.014-.3004.4814-5.1456 8.2346-.2844.4555 2.1561 3.445c.3978.6405 1.086 1.0223 1.8438 1.0223h6.0347c.7439 0 1.448-.3918 1.8419-1.0203l.2128-.3401.01-.014 2.6772-4.2844c.4396-.7001.4396-1.6051 0-2.3033h-.002Zm-.8155 1.7922-2.6772 4.2843-.0378.0518a.362.362 0 0 1-.2686.1193.3577.3577 0 0 1-.3102-.173l-2.6773-4.2844a1.1946 1.1946 0 0 1-.0795-.1512 1.2864 1.2864 0 0 1-.0577-.1571 1.2428 1.2428 0 0 1 0-.6604c.0298-.1054.0756-.2108.1352-.3063l2.6733-4.2804.006-.0099a.3788.3788 0 0 1 .2128-.1532l.0716-.0139.0298-.002a.357.357 0 0 1 .3103.175l2.6733 4.2784c.2446.3899.2446.8911 0 1.281h-.004Z"
        fill="#2E3238"
      />
      <path
        d="M23.7582 8.8234c.4376-.7021.4376-1.6051 0-2.3033L21.085 2.2417l-.2228-.36C20.4624 1.2432 19.7742.8613 19.0184.8613h-6.0347c-.7539 0-1.4421.3819-1.8439 1.0224L.3334 19.1783A2.176 2.176 0 0 0-.0007 20.33c0 .4058.1154.8055.3321 1.1496l2.898 4.6405c.3999.6404 1.0881 1.0203 1.842 1.0203h6.0346c.7578 0 1.446-.3819 1.8439-1.0223l.2207-.3501v-.004l.004-.008 2.1541-3.445 6.3848-10.2177 2.0408-3.2679.004-.002Zm-.6305-1.1516c0 .2208-.0616.4435-.1869.6404L12.3551 25.2548a.357.357 0 0 1-.3103.1711.3576.3576 0 0 1-.3102-.1711l-2.6753-4.2863c-.2427-.3879-.2427-.8871 0-1.279L19.6449 2.7509a.357.357 0 0 1 .3103-.173.3578.3578 0 0 1 .3123.175l2.6733 4.2784c.1253.1969.1869.4197.1869.6405Z"
        fill="#2E3238"
      />
    </svg>
  );
}
