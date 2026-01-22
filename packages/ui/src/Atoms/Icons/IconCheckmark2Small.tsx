import type { IconProps } from "./type";

export function IconCheckmark2Small({ size = 24, className }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" className={className} xmlns="http://www.w3.org/2000/svg">
      <path fillRule="evenodd" clipRule="evenodd" d="M17.6585 7.24744C18.0741 7.61112 18.1163 8.24288 17.7526 8.65852L10.7526 16.6585C10.5703 16.8668 10.3099 16.9902 10.0333 16.9995C9.75666 17.0087 9.4886 16.9028 9.29289 16.7071L6.29289 13.7071C5.90237 13.3166 5.90237 12.6834 6.29289 12.2929C6.68342 11.9024 7.31658 11.9024 7.70711 12.2929L9.95129 14.5371L16.2474 7.34151C16.6111 6.92587 17.2429 6.88375 17.6585 7.24744Z" fill="currentColor" />
    </svg>
  );
}
