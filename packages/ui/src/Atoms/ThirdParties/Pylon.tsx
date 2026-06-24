import type { ComponentProps } from "react";

export function Pylon(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path
        fill="#5B0EFF"
        fillRule="evenodd"
        d="M5 3h7a6 6 0 0 1 0 12H9v6H5Zm4 3v6h3a3 3 0 0 0 0-6Z"
      />
    </svg>
  );
}
