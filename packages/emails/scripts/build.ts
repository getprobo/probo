import { render } from '@react-email/components';
import { mkdir, readFile, writeFile } from 'fs/promises';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';
import Layout from '../src/Layout';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

function wrapWithGoConditional(html: string, marker: string, condition: string): string {
  const regex = new RegExp(`<span><!--\\s*${marker}_START\\s*--></span\\s*>([\\s\\S]*?)<span><!--\\s*${marker}_END\\s*--></span\\s*>`, 'g');
  return html.replace(regex, `{{if ${condition}}}$1{{end}}`);
}

async function build() {
  const outputDir = join(__dirname, '..', 'dist');
  await mkdir(outputDir, { recursive: true });

  let htmlTemplate = await render(
    Layout({
      subject: '{{.Subject}}',
      header: '{{.Header}}',
      fullName: '{{.FullName}}',
      body: '{{.Body}}',
      buttonText: '{{.ButtonText}}',
      buttonURL: '{{.ButtonURL}}',
      footer: '{{.Footer}}',
    }),
    {
      pretty: true,
    }
  );

  htmlTemplate = wrapWithGoConditional(htmlTemplate, 'CONDITIONAL_BUTTON', '.ButtonURL');
  htmlTemplate = wrapWithGoConditional(htmlTemplate, 'CONDITIONAL_FOOTER', '.Footer');

  const textTemplate = await readFile(join(__dirname, '..', 'assets', 'template.txt'), 'utf-8');

  await writeFile(join(outputDir, 'layout.html'), htmlTemplate);
  await writeFile(join(outputDir, 'layout.txt'), textTemplate);
}

build().catch((err) => {
  console.error('Failed to build email templates:', err);
  process.exit(1);
});
