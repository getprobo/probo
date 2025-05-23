/*
Generates Icon components from SVG files
 */

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const svgFiles = fs
    .readdirSync(path.join(__dirname))
    .filter((file) => file.endsWith(".svg"));

function camelCase(s: string) {
    return s.replace(/-([a-z0-9])/g, (g) => g[1].toUpperCase());
}

function capitalize(s: string) {
    return s.charAt(0).toUpperCase() + s.slice(1);
}

const exports = [];

for (const svgFile of svgFiles) {
    const iconName = svgFile.replace(".svg", "");
    const componentName = capitalize(camelCase(iconName));
    let svgContent = fs.readFileSync(path.join(__dirname, svgFile), "utf-8");

    svgContent = svgContent.replace(/width="[^"]*"/, "width={size}");
    svgContent = svgContent.replace(/height="[^"]*"/, "height={size}");
    svgContent = svgContent.replace(/fill="[^"]*"/g, 'fill="currentColor"');
    svgContent = svgContent.replace(/xmlns=/g, "className={className} xmlns=");
    svgContent = svgContent.replace(/([a-z]+-[a-z]+)="/g, function (_, attr) {
        return `${camelCase(attr)}="`;
    });

    const componentContent = `import type {IconProps} from "./type.ts";

export function Icon${componentName}({size = 24, className}: IconProps) {
  return ${svgContent}
}`;

    fs.writeFileSync(
        path.join(__dirname, `Icon${componentName}.tsx`),
        componentContent,
    );
    exports.push(
        `export { Icon${componentName} } from "./Icon${componentName}.tsx"`,
    );
}

// Write exports in the index.ts file
fs.writeFileSync(
    path.join(__dirname, "index.tsx"),
    `${exports.join(";\n")};\n`,
);
