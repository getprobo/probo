import { useLottieRef } from "./useLottieRef";
import json from "./GDPR.json";
import { AnimationStyle } from "./AnimationStyle";

export function GDPR() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
