import type { IconProps } from "./type";

export function IconInProgress({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M12 3.875C7.51269 3.875 3.875 7.51269 3.875 12C3.875 16.4873 7.51269 20.125 12 20.125C16.4873 20.125 20.125 16.4873 20.125 12C20.125 7.51269 16.4873 3.875 12 3.875ZM2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12Z" fill="currentColor" />
      <path d="M12 5.74985V18.2499C15.4513 18.2499 18.25 15.4511 18.25 11.9999C18.25 8.5486 15.4513 5.74985 12 5.74985Z" fill="currentColor" />
    </svg>
  );
}
