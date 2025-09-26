import { availableFrameworks } from "@probo/helpers";
import { Avatar } from "../../Atoms/Avatar/Avatar.tsx";

const availableLogos = new Map(
    availableFrameworks.map((framework) => [framework.name, framework.logo]),
);

export function FrameworkLogo({
    name,
    alt,
    className,
}: {
    name: string;
    alt?: string;
    className?: string;
}) {
    const logo = availableLogos.get(name);
    return logo ? (
        <img src={logo} alt={alt} className={className} />
    ) : (
        <Avatar name={name} size="l" className={className} />
    );
}
