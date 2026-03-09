import { IconBrandFacebook } from "./IconBrandFacebook";
import { IconBrandLinkedin } from "./IconBrandLinkedin";
import { IconBrandX } from "./IconBrandX";
import { IconGlobe } from "./IconGlobe";
import type { IconProps } from "./type";

export type SocialIconProps = IconProps & { colored?: boolean };

export function SocialIcon({ socialName, ...props }: { socialName: string | null } & SocialIconProps) {
  switch (socialName) {
    case "LinkedIn": return <IconBrandLinkedin {...props} />;
    case "X": return <IconBrandX {...props} />;
    case "Facebook": return <IconBrandFacebook {...props} />;
    default: return <IconGlobe {...props} />;
  }
}
