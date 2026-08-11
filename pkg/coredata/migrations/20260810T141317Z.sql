-- Copyright (c) 2026 Probo Inc <hello@probo.com>.
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

-- Backfill the neutral opt-out button label for existing banners. The stored
-- `button_opt_out` is the statutory California phrase (11 CCR § 7015), which
-- misdescribes the choice in the non-US jurisdictions that now share the
-- opt-out presentation. New banners already receive this key from
-- defaultUIStringsByLanguage.

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Reject non-essential cookies"}'::jsonb,
    updated_at = NOW()
WHERE language = 'en'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Refuser les cookies non essentiels"}'::jsonb,
    updated_at = NOW()
WHERE language = 'fr'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Nicht wesentliche Cookies ablehnen"}'::jsonb,
    updated_at = NOW()
WHERE language = 'de'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Rechazar cookies no esenciales"}'::jsonb,
    updated_at = NOW()
WHERE language = 'es'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Tolak cookie yang tidak penting"}'::jsonb,
    updated_at = NOW()
WHERE language = 'id'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Rifiuta i cookie non essenziali"}'::jsonb,
    updated_at = NOW()
WHERE language = 'it'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "不要なクッキーを拒否する"}'::jsonb,
    updated_at = NOW()
WHERE language = 'ja'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "비필수 쿠키 거부"}'::jsonb,
    updated_at = NOW()
WHERE language = 'ko'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Niet-essentiële cookies weigeren"}'::jsonb,
    updated_at = NOW()
WHERE language = 'nl'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Odrzuć nieistotne pliki cookie"}'::jsonb,
    updated_at = NOW()
WHERE language = 'pl'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Recusar cookies não essenciais"}'::jsonb,
    updated_at = NOW()
WHERE language = 'pt'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Zorunlu olmayan çerezleri reddet"}'::jsonb,
    updated_at = NOW()
WHERE language = 'tr'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "Відхилити неістотні файли cookie"}'::jsonb,
    updated_at = NOW()
WHERE language = 'uk'
  AND NOT translations ? 'button_opt_out_generic';

UPDATE cookie_banner_translations
SET translations = translations || '{"button_opt_out_generic": "拒绝非必要 Cookie"}'::jsonb,
    updated_at = NOW()
WHERE language = 'zh'
  AND NOT translations ? 'button_opt_out_generic';

-- Languages Probo does not ship are added by integrators, so there is no
-- localized neutral label to seed. Reuse their own reject-all wording rather
-- than injecting English, which would read worse than the phrase it replaces.

UPDATE cookie_banner_translations
SET translations = translations || jsonb_build_object(
        'button_opt_out_generic', translations ->> 'button_reject_all'
    ),
    updated_at = NOW()
WHERE NOT translations ? 'button_opt_out_generic'
  AND COALESCE(translations ->> 'button_reject_all', '') <> '';
