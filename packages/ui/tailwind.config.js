/** @type {import('tailwindcss').Config} */
module.exports = {
  theme: {
    extend: {
      typography: () => ({
        neutral: {
          css: {
            "--tw-prose-body": "var(--color-txt-secondary)",
            "--tw-prose-headings": "var(--color-txt-primary)",
            "--tw-prose-lead": "var(--color-txt-secondary)",
            "--tw-prose-links": "var(--color-txt-secondary)",
            "--tw-prose-bold": "var(--color-txt-secondary)",
            "--tw-prose-counters": "var(--color-txt-secondary)",
            "--tw-prose-bullets": "var(--color-txt-secondary)",
            "--tw-prose-hr": "var(--color-txt-secondary)",
            "--tw-prose-quotes": "var(--color-txt-secondary)",
            "--tw-prose-quote-borders": "var(--color-txt-primary)",
            "--tw-prose-captions": "var(--color-txt-secondary)",
            "--tw-prose-code": "var(--color-txt-primary)",
            "--tw-prose-pre-code": "var(--color-txt-primary)",
            "--tw-prose-pre-bg": "transparent",
            "--tw-prose-th-borders": "var(--color-txt-secondary)",
            "--tw-prose-td-borders": "var(--color-txt-secondary)",
          },
        },
      }),
    },
  },
};
