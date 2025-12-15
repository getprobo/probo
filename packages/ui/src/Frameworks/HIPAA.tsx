import { useLottieRef } from "./useLottieRef";
import json from "./HIPAA.json";
import { AnimationStyle } from "./AnimationStyle";

export function HIPAA() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
