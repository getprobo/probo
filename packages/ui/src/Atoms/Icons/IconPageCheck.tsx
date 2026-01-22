import type { IconProps } from "./type";

export function IconPageCheck({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M8 4C6.89543 4 6 4.89543 6 6V18C6 19.1046 6.89542 20 8 20H9C9.55228 20 10 20.4477 10 21C10 21.5523 9.55228 22 9 22H8C5.79088 22 4 20.2092 4 18V6C4 3.79087 5.79087 2 8 2H16C18.2092 2 20 3.79088 20 6V10C20 10.5523 19.5523 11 19 11C18.4477 11 18 10.5523 18 10V6C18 4.89542 17.1046 4 16 4H8Z" fill="currentColor" />
      <path fillRule="evenodd" clipRule="evenodd" d="M21.1402 14.2318C21.5645 14.5853 21.6218 15.2159 21.2682 15.6402L16.2682 21.6404C16.0884 21.8562 15.826 21.9864 15.5453 21.9992C15.2646 22.0119 14.9915 21.906 14.7929 21.7073L12.7929 19.7071C12.4024 19.3165 12.4024 18.6834 12.7929 18.2929C13.1835 17.9024 13.8166 17.9024 14.2071 18.2929L15.4328 19.5187L19.7318 14.3598C20.0853 13.9355 20.7159 13.8782 21.1402 14.2318Z" fill="currentColor" />
    </svg>
  );
}
