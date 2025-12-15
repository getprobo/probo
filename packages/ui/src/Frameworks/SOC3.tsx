import { useLottieRef } from "./useLottieRef";
import json from "./SOC3.json";
import { AnimationStyle } from "./AnimationStyle";

export function SOC3() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
