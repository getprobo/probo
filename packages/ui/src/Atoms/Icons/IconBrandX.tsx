// X (formerly Twitter) logo — trademark of X Corp.
// Used here solely for brand identification (nominative fair use).
// SVG path from Simple Icons (https://simpleicons.org), CC0 license.
import type { IconProps } from "./type";

type Props = IconProps & { colored?: boolean };

export function IconBrandX({ size = 24, className, colored = false }: Props) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill={colored ? "#000000" : "currentColor"} className={className} xmlns="http://www.w3.org/2000/svg">
      <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-4.714-6.231-5.401 6.231H2.746l7.73-8.835L1.254 2.25H8.08l4.713 6.097zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
    </svg>
  );
}
