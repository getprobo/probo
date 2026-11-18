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

INSERT INTO cookie_banner_translations (
    id,
    tenant_id,
    organization_id,
    cookie_banner_id,
    language,
    translations,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(cb.tenant_id), 86),
    cb.tenant_id,
    cb.organization_id,
    cb.id,
    'nl',
    '{
        "banner_title": "Cookievoorkeuren",
        "banner_description": "We gebruiken cookies om je ervaring te verbeteren en het websiteverkeer te analyseren. {{cookie_policy_link}}",
        "button_accept_all": "Alles accepteren",
        "button_reject_all": "Alles weigeren",
        "button_customize": "Aanpassen",
        "button_save": "Voorkeuren opslaan",
        "panel_title": "Voorkeuren aanpassen",
        "panel_description": "Kies welke cookiecategorieën je wilt toestaan. {{necessary_category}} cookies zijn altijd actief omdat ze nodig zijn om de website te laten werken.",
        "aria_close": "Sluiten",
        "aria_show_details": "Cookiedetails weergeven",
        "aria_hide_details": "Cookiedetails verbergen",
        "aria_cookie_settings": "Cookie-instellingen",
        "privacy_policy_link_text": "Privacybeleid",
        "cookie_policy_link_text": "Cookiebeleid",
        "placeholder_text": "Deze inhoud vereist {{category}} cookies.",
        "placeholder_button": "Cookievoorkeuren beheren",
        "banner_title_opt_out": "Cookiemelding",
        "banner_description_opt_out": "We gebruiken cookies en vergelijkbare technologieën. Je kunt niet-essentiële cookies weigeren. {{cookie_policy_link}}",
        "button_acknowledge": "OK",
        "button_opt_out": "Verkoop of deel mijn persoonlijke gegevens niet",
        "banner_title_notice": "Cookiemelding",
        "banner_description_notice": "Deze website gebruikt cookies om je ervaring te verbeteren. {{cookie_policy_link}}",
        "button_dismiss": "Begrepen"
    }'::jsonb,
    NOW(),
    NOW()
FROM cookie_banners cb
ON CONFLICT (cookie_banner_id, language)
DO NOTHING;
