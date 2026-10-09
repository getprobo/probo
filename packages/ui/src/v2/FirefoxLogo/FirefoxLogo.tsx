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
import { useId } from "react";

export type FirefoxLogoProps = Omit<ComponentProps<"svg">, "width" | "height"> & {
  size?: string | number;
};

// Official Firefox mark. Fills are Firefox's brand palette and do not
// follow currentColor. Size it with `size`, matching Phosphor icons.
export function FirefoxLogo({ size = "1em", ...props }: FirefoxLogoProps) {
  const uid = useId();

  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 51500 51500"
      width={size}
      height={size}
      role="img"
      aria-label="Firefox"
      {...props}
    >
      <radialGradient id={`${uid}b`} cx="87.4%" cy="-12.9%" r="128%" gradientTransform="matrix(.8 0 0 1 .18 .13)">
        <stop offset=".13" stopColor="#ffbd4f" />
        <stop offset=".28" stopColor="#ff980e" />
        <stop offset=".47" stopColor="#ff3750" />
        <stop offset=".78" stopColor="#eb0878" />
        <stop offset=".86" stopColor="#e50080" />
      </radialGradient>
      <radialGradient id={`${uid}d`} cx="49%" cy="40%" r="128%" gradientTransform="matrix(.82 0 0 1 .09 0)">
        <stop offset=".3" stopColor="#960e18" />
        <stop offset=".35" stopColor="#b11927" stopOpacity=".74" />
        <stop offset=".43" stopColor="#db293d" stopOpacity=".34" />
        <stop offset=".5" stopColor="#f5334b" stopOpacity=".1" />
        <stop offset=".53" stopColor="#ff3750" stopOpacity="0" />
      </radialGradient>
      <radialGradient id={`${uid}e`} cx="48%" cy="-12%" r="140%">
        <stop offset=".13" stopColor="#fff44f" />
        <stop offset=".53" stopColor="#ff980e" />
      </radialGradient>
      <radialGradient id={`${uid}f`} cx="22.76%" cy="110.11%" r="100%">
        <stop offset=".35" stopColor="#3a8ee6" />
        <stop offset=".67" stopColor="#9059ff" />
        <stop offset="1" stopColor="#c139e6" />
      </radialGradient>
      <radialGradient id={`${uid}h`} cx="52%" cy="33%" r="59%" gradientTransform="scale(.9 1)">
        <stop offset=".21" stopColor="#9059ff" stopOpacity="0" />
        <stop offset=".97" stopColor="#6e008b" stopOpacity=".6" />
      </radialGradient>
      <radialGradient id={`${uid}i`} cx="210%" cy="-100%" r="290%">
        <stop offset=".1" stopColor="#ffe226" />
        <stop offset=".79" stopColor="#ff7139" />
      </radialGradient>
      <radialGradient id={`${uid}j`} cx="84%" cy="-41%" r="180%">
        <stop offset=".11" stopColor="#fff44f" />
        <stop offset=".46" stopColor="#ff980e" />
        <stop offset=".72" stopColor="#ff3647" />
        <stop offset=".9" stopColor="#e31587" />
      </radialGradient>
      <radialGradient id={`${uid}k`} cx="16.1%" cy="-18.6%" r="348.8%" gradientTransform="matrix(.10453 .46743 -.99452 .04913 -.05 -.26)">
        <stop offset="0" stopColor="#fff44f" />
        <stop offset=".3" stopColor="#ff980e" />
        <stop offset=".57" stopColor="#ff3647" />
        <stop offset=".74" stopColor="#e31587" />
      </radialGradient>
      <radialGradient id={`${uid}l`} cx="18.9%" cy="-42.5%" r="238.4%">
        <stop offset=".14" stopColor="#fff44f" />
        <stop offset=".48" stopColor="#ff980e" />
        <stop offset=".66" stopColor="#ff3647" />
        <stop offset=".9" stopColor="#e31587" />
      </radialGradient>
      <radialGradient id={`${uid}m`} cx="159.3%" cy="-44.72%" r="313.1%">
        <stop offset=".09" stopColor="#fff44f" />
        <stop offset=".63" stopColor="#ff980e" />
      </radialGradient>
      <linearGradient id={`${uid}a`} x1="87.25%" x2="9.4%" y1="15.5%" y2="93.1%">
        <stop offset=".05" stopColor="#fff44f" />
        <stop offset=".37" stopColor="#ff980e" />
        <stop offset=".53" stopColor="#ff3647" />
        <stop offset=".7" stopColor="#e31587" />
      </linearGradient>
      <linearGradient id={`${uid}n`} x1="80%" x2="18%" y1="14%" y2="84%">
        <stop offset=".17" stopColor="#fff44f" stopOpacity=".8" />
        <stop offset=".6" stopColor="#fff44f" stopOpacity="0" />
      </linearGradient>
      <path id={`${uid}c`} d="M47870 16735c-1044-2512-3160-5224-4820-6082 1352 2650 2134 5310 2433 7294 0-6 2 5 4 22l4 26c2268 6147 1032 12398-748 16218-2754 5910-9420 11967-19857 11670-11276-318-21210-8683-23064-19643-338-1728 0-2605 170-4008-207 1080-286 1394-390 3315l-2 123c0 13270 10760 24030 24032 24030 11887 0 21756-8630 23690-19963l110-927c477-4120-53-8453-1560-12075z" />
      <path id={`${uid}g`} d="M25677 21050c-40 598-2150 2660-2890 2660-6834 0-7943 4133-7943 4133 303 3480 2726 6348 5660 7865 134 70 270 130 405 193a13277 13277 0 00706 289 10674 10674 0 003127 603c11978 562 14300-14320 5655-18640 2213-385 4510 505 5794 1407-2100-3672-6025-6150-10530-6150-285 0-564 24-844 43a12025 12025 0 00-6614 2549c366 310 780 724 1650 1583 1630 1606 5813 3270 5822 3465z" />
      <path fill={`url(#${uid}a)`} d="M47870 16735c-1044-2512-3160-5224-4820-6082 1352 2650 2134 5310 2433 7294l5 40c-2718-6773-7325-9505-11088-15452l-566-920a7372 7372 0 01-265-497 4370 4370 0 01-359-950 63 63 0 00-55-65 82 82 0 00-45 0l-12 7-17 10 10-14c-6037 3536-8085 10076-8274 13350a12025 12025 0 00-6614 2548 7136 7136 0 00-622-470 11134 11134 0 01-68-5873c-2468 1124-4390 2900-5785 4470h-10c-953-1206-886-5187-832-6018-10-52-710 363-802 425a17507 17507 0 00-2349 2012 21048 21048 0 00-2244 2692l-1 3v-3a20284 20284 0 00-3225 7280l-32 160a39700 39700 0 00-237 1500l-5 52a22907 22907 0 00-390 3316l-1 120c0 13270 10760 24030 24032 24030 11887 0 21756-8630 23690-19963l110-927c477-4120-53-8453-1560-12075zM20170 35545c113 53 220 112 334 164l16 10a12620 12620 0 01-350-174zm5506-14493zm19813-3060l-3-23 4 26z" />
      <use fill={`url(#${uid}b)`} href={`#${uid}c`} />
      <use fill={`url(#${uid}d)`} href={`#${uid}c`} />
      <path fill={`url(#${uid}e)`} d="M36192 19560l150 110a13070 13070 0 00-2231-2911C26640 9290 32150 563 33080 120l10-13c-6037 3535-8085 10076-8273 13348 280-20 560-43 844-43 4505 0 8430 2477 10530 6150z" />
      <use fill={`url(#${uid}f)`} href={`#${uid}g`} />
      <use fill={`url(#${uid}h)`} href={`#${uid}g`} />
      <path fill={`url(#${uid}i)`} d="M17083 15204a24404 24404 0 01498 330 11134 11134 0 01-67-5874c-2470 1125-4390 2900-5785 4470 115-3 3600-66 5354 1074z" />
      <path fill={`url(#${uid}j)`} d="M1822 26240c1855 10960 11788 19325 23063 19644 10437 296 17104-5762 19858-11670 1780-3820 3016-10070 748-16218v-2l-4-24c-2-17-4-28-4-22l5 40c853 5566-1980 10958-6405 14604l-13 30c-8625 7023-16878 4237-18550 3097a14410 14410 0 01-350-174c-5028-2403-7105-6984-6660-10913-4245 0-5693-3580-5693-3580s3812-2718 8836-355c4653 2190 9023 355 9023 354-10-195-4192-1860-5822-3465-872-860-1285-1272-1652-1583a7136 7136 0 00-622-470 28293 28293 0 00-498-330c-1753-1140-5240-1076-5355-1073h-10c-953-1207-886-5188-832-6020-10-50-710 363-802 426a17507 17507 0 00-2349 2012 21048 21048 0 00-2244 2692l-1 3v-3a20284 20284 0 00-3225 7280c-10 52-865 3784-444 5720z" />
      <path fill={`url(#${uid}k)`} d="M34110 16760a13070 13070 0 012231 2910l360 296c5450 5020 2594 12120 2380 12626 4426-3646 7258-9038 6405-14604-2716-6774-7323-9506-11086-15453l-566-920a7372 7372 0 01-265-497 4370 4370 0 01-359-950 63 63 0 00-55-65 82 82 0 00-45 0l-12 7-17 10c-930 443-6440 9170 1030 16640z" />
      <path fill={`url(#${uid}l)`} d="M36702 19965a4743 4743 0 00-360-295l-150-110c-1283-900-3580-1792-5794-1407 8644 4322 6323 19203-5655 18640a10674 10674 0 01-3127-603 13451 13451 0 01-706-289 9064 9064 0 01-405-193l16 10c1670 1140 9924 3925 18550-3097l13-30c213-506 3068-7606-2380-12626z" />
      <path fill={`url(#${uid}m)`} d="M14844 27844s1110-4133 7943-4133c740 0 2850-2062 2890-2660s-4370 1836-9023-354c-5024-2363-8836 354-8836 354s1448 3580 5693 3580c-445 3930 1632 8510 6660 10913 113 53 218 112 334 164-2935-1517-5358-4384-5660-7865z" />
      <path fill={`url(#${uid}n)`} d="M47870 16735c-1044-2512-3160-5224-4820-6082 1352 2650 2134 5310 2433 7294l5 40c-2718-6773-7325-9505-11088-15452l-566-920a7372 7372 0 01-265-497 4370 4370 0 01-359-950 63 63 0 00-55-65 82 82 0 00-45 0l-12 7-17 10 10-14c-6037 3536-8085 10076-8274 13350 280-20 560-43 845-43 4505 0 8430 2477 10530 6148-1284-900-3580-1792-5795-1407 8644 4322 6323 19203-5655 18640a10674 10674 0 01-3127-603 13451 13451 0 01-706-289 9064 9064 0 01-405-193l17 10a14410 14410 0 01-350-174c112 53 218 112 333 164-2935-1517-5358-4384-5660-7865 0 0 1108-4133 7942-4133 740 0 2850-2062 2890-2660-10-195-4190-1860-5822-3465-870-860-1285-1272-1650-1583a7136 7136 0 00-623-470 11134 11134 0 01-67-5873c-2470 1124-4390 2900-5785 4470h-10c-953-1207-886-5187-832-6020-10-50-710 363-802 426a17507 17507 0 00-2349 2012 21048 21048 0 00-2243 2692l-1 3v-3a20284 20284 0 00-3225 7280l-32 160a39787 39787 0 00-277 1515c-2 18 2-17 0 0a27956 27956 0 00-355 3353l-3 122c0 13270 10760 24030 24032 24030 11887 0 21756-8630 23690-19963l110-927c477-4120-53-8453-1560-12075zm-2384 1234l4 26v-2l-4-24z" />
    </svg>
  );
}
