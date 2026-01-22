import type { IconProps } from "./type";

export function IconPageTextLine({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M4 6C4 3.79087 5.79087 2 8 2H16C18.2092 2 20 3.79088 20 6V18C20 20.2092 18.2092 22 16 22H8C5.79088 22 4 20.2092 4 18V6ZM8 4C6.89543 4 6 4.89543 6 6V18C6 19.1046 6.89542 20 8 20H16C17.1046 20 18 19.1046 18 18V6C18 4.89542 17.1046 4 16 4H8ZM8 7C8 6.44772 8.44772 6 9 6H15C15.5523 6 16 6.44772 16 7C16 7.55228 15.5523 8 15 8H9C8.44772 8 8 7.55228 8 7ZM8 11C8 10.4477 8.44772 10 9 10H15C15.5523 10 16 10.4477 16 11C16 11.5523 15.5523 12 15 12H9C8.44772 12 8 11.5523 8 11ZM8 15C8 14.4477 8.44772 14 9 14H11C11.5523 14 12 14.4477 12 15C12 15.5523 11.5523 16 11 16H9C8.44772 16 8 15.5523 8 15Z" fill="currentColor" />
    </svg>
  );
}
