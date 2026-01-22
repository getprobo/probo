import type { IconProps } from "./type";

export function IconPin({ size = 24, className }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      className={className}
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 17 16"
    >
      <path
        stroke="currentColor"
        strokeLinejoin="round"
        strokeWidth="1.33"
        d="M10 6.67a1.67 1.67 0 1 1-3.33 0 1.67 1.67 0 0 1 3.33 0Z"
      />
      <path
        stroke="currentColor"
        strokeLinejoin="round"
        strokeWidth="1.33"
        d="M13 6.67c0 2.91-2.6 5.55-3.91 6.71-.44.38-1.07.38-1.51 0-1.32-1.16-3.91-3.8-3.91-6.71a4.67 4.67 0 1 1 9.33 0Z"
      />
    </svg>
  );
}
