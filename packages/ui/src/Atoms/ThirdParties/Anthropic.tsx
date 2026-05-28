import type { ComponentProps } from "react";

export function Anthropic(props: ComponentProps<"svg">) {
  return (
    <svg
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M17.3041 3.541h-3.6718l6.696 16.918H24Zm-10.6082 0L0 20.459h3.7442l1.3693-3.5528h7.0052l1.3693 3.5528h3.7442L10.5363 3.541Zm-.3712 10.2231 2.2914-5.9456 2.2914 5.9456Z"
        fill="#D97757"
      />
    </svg>
  );
}
