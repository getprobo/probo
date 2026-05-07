-- Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
--
-- Permission to use, copy, modify, and/or distribute this software for any
-- purpose with or without fee is hereby granted, provided that the above
-- copyright notice and this permission notice appear in all copies.
--
-- THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
-- REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
-- AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
-- INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
-- LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
-- OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
-- PERFORMANCE OF THIS SOFTWARE.

UPDATE cookie_banner_translations
SET translations = translations || '{
    "banner_title_opt_out": "Cookie Notice",
    "banner_description_opt_out": "We use cookies and similar technologies. You can opt out of non-essential cookies. {{cookie_policy_link}}",
    "button_acknowledge": "OK",
    "button_opt_out": "Do Not Sell or Share My Personal Information",
    "banner_title_notice": "Cookie Notice",
    "banner_description_notice": "This site uses cookies to enhance your experience. {{cookie_policy_link}}",
    "button_dismiss": "Got it"
}'::jsonb,
    updated_at = NOW()
WHERE language = 'en'
  AND NOT translations ? 'banner_title_opt_out';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "banner_title_opt_out": "Avis sur les cookies",
    "banner_description_opt_out": "Nous utilisons des cookies et technologies similaires. Vous pouvez refuser les cookies non essentiels. {{cookie_policy_link}}",
    "button_acknowledge": "OK",
    "button_opt_out": "Ne pas vendre ni partager mes informations personnelles",
    "banner_title_notice": "Avis sur les cookies",
    "banner_description_notice": "Ce site utilise des cookies pour améliorer votre expérience. {{cookie_policy_link}}",
    "button_dismiss": "Compris"
}'::jsonb,
    updated_at = NOW()
WHERE language = 'fr'
  AND NOT translations ? 'banner_title_opt_out';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "banner_title_opt_out": "Cookie-Hinweis",
    "banner_description_opt_out": "Wir verwenden Cookies und ähnliche Technologien. Sie können nicht wesentliche Cookies ablehnen. {{cookie_policy_link}}",
    "button_acknowledge": "OK",
    "button_opt_out": "Meine persönlichen Daten nicht verkaufen oder weitergeben",
    "banner_title_notice": "Cookie-Hinweis",
    "banner_description_notice": "Diese Website verwendet Cookies, um Ihre Erfahrung zu verbessern. {{cookie_policy_link}}",
    "button_dismiss": "Verstanden"
}'::jsonb,
    updated_at = NOW()
WHERE language = 'de'
  AND NOT translations ? 'banner_title_opt_out';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "banner_title_opt_out": "Aviso de cookies",
    "banner_description_opt_out": "Utilizamos cookies y tecnologías similares. Puede optar por no recibir cookies no esenciales. {{cookie_policy_link}}",
    "button_acknowledge": "OK",
    "button_opt_out": "No vender ni compartir mi información personal",
    "banner_title_notice": "Aviso de cookies",
    "banner_description_notice": "Este sitio utiliza cookies para mejorar su experiencia. {{cookie_policy_link}}",
    "button_dismiss": "Entendido"
}'::jsonb,
    updated_at = NOW()
WHERE language = 'es'
  AND NOT translations ? 'banner_title_opt_out';
