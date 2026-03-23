// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { WidgetStrings } from "./types";

const translations: Record<string, WidgetStrings> = {
  en: {
    customize: "Customize",
    cookiePreferences: "Cookie Preferences",
    back: "Back",
    rejectAll: "Reject all",
    acceptAll: "Accept all",
    savePreferences: "Save my preferences",
    privacyPolicy: "Privacy Policy",
    cookiePreferencesTooltip: "Cookie preferences",
    contextualBlockedMessage:
      'This content requires "{categoryName}" cookies to be displayed.',
    contextualAllowButton: "Allow and show content",
  },
  fr: {
    customize: "Personnaliser",
    cookiePreferences: "Préférences de cookies",
    back: "Retour",
    rejectAll: "Tout refuser",
    acceptAll: "Tout accepter",
    savePreferences: "Enregistrer mes préférences",
    privacyPolicy: "Politique de confidentialité",
    cookiePreferencesTooltip: "Préférences de cookies",
    contextualBlockedMessage:
      'Ce contenu nécessite les cookies « {categoryName} » pour être affiché.',
    contextualAllowButton: "Autoriser et afficher le contenu",
  },
  de: {
    customize: "Anpassen",
    cookiePreferences: "Cookie-Einstellungen",
    back: "Zurück",
    rejectAll: "Alle ablehnen",
    acceptAll: "Alle akzeptieren",
    savePreferences: "Meine Einstellungen speichern",
    privacyPolicy: "Datenschutzerklärung",
    cookiePreferencesTooltip: "Cookie-Einstellungen",
    contextualBlockedMessage:
      'Dieser Inhalt erfordert „{categoryName}"-Cookies, um angezeigt zu werden.',
    contextualAllowButton: "Zulassen und Inhalt anzeigen",
  },
  es: {
    customize: "Personalizar",
    cookiePreferences: "Preferencias de cookies",
    back: "Volver",
    rejectAll: "Rechazar todo",
    acceptAll: "Aceptar todo",
    savePreferences: "Guardar mis preferencias",
    privacyPolicy: "Política de privacidad",
    cookiePreferencesTooltip: "Preferencias de cookies",
    contextualBlockedMessage:
      'Este contenido requiere cookies de "{categoryName}" para mostrarse.',
    contextualAllowButton: "Permitir y mostrar contenido",
  },
};

export function getStrings(lang: string): WidgetStrings {
  const exact = translations[lang];
  if (exact) return exact;

  const prefix = lang.split("-")[0];
  const prefixMatch = translations[prefix];
  if (prefixMatch) return prefixMatch;

  return translations["en"];
}
