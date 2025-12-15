import { useLottieRef } from "./useLottieRef";
import json from "./SOC2_TYPE1.json";
import { AnimationStyle } from "./AnimationStyle";

export function SOC2_TYPE1() {
    const ref = useLottieRef(json);

    return (
        <div ref={ref}>
            <AnimationStyle />
        </div>
    );
}
