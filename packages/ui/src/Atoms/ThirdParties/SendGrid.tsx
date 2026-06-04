import type { ComponentProps } from "react";

export function SendGrid(props: ComponentProps<"svg">) {
  return (
    <svg
      viewBox="0 0 256 256"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M256 0v170.667h-85.333v85.329H0V170.666H0V85.331h85.333V0H256Z"
        fill="#9DD6E3"
      />
      <polygon fill="#3F72AB" points="0 255.996 85.335 255.996 85.335 170.663 0 170.663" />
      <polygon fill="#00A9D1" points="170.667 170.667 256 170.667 256 85.331 170.667 85.331" />
      <polygon fill="#00A9D1" points="85.333 85.333 170.667 85.333 170.667 0 85.333 0" />
      <polygon fill="#2191C4" points="85.333 170.665 170.667 170.665 170.667 85.331 85.333 85.331" />
      <polygon fill="#3F72AB" points="170.667 85.333 256 85.333 256 0 170.667 0" />
    </svg>
  );
}
