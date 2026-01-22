import type { IconProps } from "./type";

export function IconPageTextSolid({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M4 6C4 3.79086 5.79086 2 8 2H16C18.2091 2 20 3.79086 20 6V18C20 20.2091 18.2091 22 16 22H8C5.79086 22 4 20.2091 4 18V6ZM9 6C8.44772 6 8 6.44772 8 7C8 7.55228 8.44772 8 9 8H15C15.5523 8 16 7.55228 16 7C16 6.44772 15.5523 6 15 6H9ZM9 10C8.44772 10 8 10.4477 8 11C8 11.5523 8.44772 12 9 12H15C15.5523 12 16 11.5523 16 11C16 10.4477 15.5523 10 15 10H9ZM9 14C8.44772 14 8 14.4477 8 15C8 15.5523 8.44772 16 9 16H11C11.5523 16 12 15.5523 12 15C12 14.4477 11.5523 14 11 14H9Z" fill="currentColor" />
    </svg>
  );
}
