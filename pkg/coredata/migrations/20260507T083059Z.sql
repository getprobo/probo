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
