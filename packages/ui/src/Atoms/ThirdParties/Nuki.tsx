import type { ComponentProps } from "react";

export function Nuki(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" {...props}>
      <rect width="24" height="24" rx="6" fill="#00BCC5" />
      <path
        fill="#FFFFFF"
        d="M12 5.5a3.75 3.75 0 00-1.85 7.012l-1.16 4.34a.75.75 0 00.724.948h4.572a.75.75 0 00.725-.947l-1.16-4.341A3.75 3.75 0 0012 5.5zm0 1.75a2 2 0 011.05 3.702.875.875 0 00-.386 1.968l.95 3.13h-3.228l.95-3.13a.875.875 0 00-.386-.968A2 2 0 0112 7.25z"
      />
    </svg>
  );
}
