import type { IconProps } from "./type";

export function IconPriority({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path d="M4.82812 15H4.17188C2.90335 15 1.875 15.9402 1.875 17.1V18.9C1.875 20.0598 2.90335 21 4.17188 21H4.82812C6.09665 21 7.125 20.0598 7.125 18.9V17.1C7.125 15.9402 6.09665 15 4.82812 15Z" fill="currentColor" />
      <path d="M12.3281 9H11.6719C10.4033 9 9.375 10.0447 9.375 11.3333V18.6667C9.375 19.9553 10.4033 21 11.6719 21H12.3281C13.5967 21 14.625 19.9553 14.625 18.6667V11.3333C14.625 10.0447 13.5967 9 12.3281 9Z" fill="currentColor" />
      <path d="M19.8281 3H19.1719C17.9033 3 16.875 4.00736 16.875 5.25V18.75C16.875 19.9926 17.9033 21 19.1719 21H19.8281C21.0967 21 22.125 19.9926 22.125 18.75V5.25C22.125 4.00736 21.0967 3 19.8281 3Z" fill="currentColor" />
    </svg>
  );
}
