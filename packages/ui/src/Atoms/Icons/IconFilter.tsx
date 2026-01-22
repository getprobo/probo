import type { IconProps } from "./type";

export function IconFilter({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M6 12C6 11.3373 6.53726 10.8 7.2 10.8H16.8C17.4627 10.8 18 11.3373 18 12C18 12.6627 17.4627 13.2 16.8 13.2H7.2C6.53726 13.2 6 12.6627 6 12Z" fill="currentColor" />
      <path fillRule="evenodd" clipRule="evenodd" d="M2.4 6C2.4 5.33726 2.93726 4.8 3.6 4.8H20.4C21.0627 4.8 21.6 5.33726 21.6 6C21.6 6.66274 21.0627 7.2 20.4 7.2H3.6C2.93726 7.2 2.4 6.66274 2.4 6Z" fill="currentColor" />
      <path fillRule="evenodd" clipRule="evenodd" d="M9.6 18C9.6 17.3373 10.1373 16.8 10.8 16.8H13.2C13.8627 16.8 14.4 17.3373 14.4 18C14.4 18.6627 13.8627 19.2 13.2 19.2H10.8C10.1373 19.2 9.6 18.6627 9.6 18Z" fill="currentColor" />
    </svg>
  );
}
