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

-- Backfill Dutch Privacy Choices panel copy omitted from 20260804T084255Z.

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Uw privacykeuzes",
    "privacy_choices_intro": "De Californische wet geeft je het recht om te bepalen hoe wij je persoonlijke gegevens gebruiken en delen. Gebruik de onderstaande opties om die rechten uit te oefenen. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Verkoop of delen van persoonlijke gegevens weigeren",
    "privacy_choices_sale_description": "Je hebt het recht om je af te melden voor de verkoop of het delen van je persoonlijke gegevens. Als je Verkoop of deel mijn persoonlijke gegevens niet kiest, stoppen wij met het verkopen of delen van je persoonlijke gegevens voor deze browser.",
    "privacy_choices_spi_title": "Gebruik van gevoelige persoonlijke gegevens beperken",
    "privacy_choices_spi_description": "Je hebt het recht om het gebruik en de openbaarmaking van je gevoelige persoonlijke gegevens te beperken. Wij gebruiken of openbaren geen gevoelige persoonlijke gegevens voor andere doeleinden dan die zijn toegestaan door de Californische wet."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'nl'
  AND NOT translations ? 'privacy_choices_title';
