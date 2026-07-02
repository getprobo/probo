import type { ComponentProps } from "react";

export function Pylon(props: ComponentProps<"svg">) {
  return (
    <svg
      viewBox="0 0 614 614"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <circle cx="306" cy="306" r="277" fill="none" stroke="#5B0EFF" strokeWidth="60" />
      <clipPath id="pylon-logo-clip">
        <circle cx="306" cy="306" r="247" />
      </clipPath>
      <g clipPath="url(#pylon-logo-clip)" fill="#5B0EFF">
        <rect x="279" y="0" width="56" height="614" />
        <rect x="0" y="155" width="335" height="60" />
        <rect x="0" y="277" width="335" height="60" />
        <rect x="0" y="399" width="335" height="60" />
      </g>
    </svg>
  );
}
