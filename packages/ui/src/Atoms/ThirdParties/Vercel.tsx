import type { ComponentProps } from "react";

export function Vercel(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <path d="M24 22.525H0l12-21.05 12 21.05z" fill="#000000" />
    </svg>
  );
}
