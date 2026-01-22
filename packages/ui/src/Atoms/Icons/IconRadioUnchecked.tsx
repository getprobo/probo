import type { IconProps } from "./type";

export function IconRadioUnchecked({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 25 25" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M13 4.875C8.51269 4.875 4.875 8.51269 4.875 13C4.875 17.4873 8.51269 21.125 13 21.125C17.4873 21.125 21.125 17.4873 21.125 13C21.125 8.51269 17.4873 4.875 13 4.875ZM3 13C3 7.47715 7.47715 3 13 3C18.5228 3 23 7.47715 23 13C23 18.5228 18.5228 23 13 23C7.47715 23 3 18.5228 3 13Z" fill="currentColor" />
    </svg>
  );
}
