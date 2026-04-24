// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package cookiebanner

import "go.probo.inc/probo/pkg/coredata"

var SupportedLanguages = []string{"en", "fr", "de", "es"}

var defaultCategories = []struct {
	Name            string
	Slug            string
	Description     string
	Kind            coredata.CookieCategoryKind
	Rank            int
	GCMConsentTypes []string
	PostHogConsent  bool
}{
	{
		Name:            "Necessary",
		Slug:            "necessary",
		Description:     "Essential cookies required for the website to function properly.",
		Kind:            coredata.CookieCategoryKindNecessary,
		Rank:            0,
		GCMConsentTypes: []string{"security_storage"},
		PostHogConsent:  false,
	},
	{
		Name:            "Analytics",
		Slug:            "analytics",
		Description:     "Cookies that help understand how visitors interact with the website.",
		Kind:            coredata.CookieCategoryKindNormal,
		Rank:            1,
		GCMConsentTypes: []string{"analytics_storage"},
		PostHogConsent:  true,
	},
	{
		Name:            "Advertising",
		Slug:            "advertising",
		Description:     "Cookies used to deliver relevant advertisements and track campaigns.",
		Kind:            coredata.CookieCategoryKindNormal,
		Rank:            2,
		GCMConsentTypes: []string{"ad_storage", "ad_user_data", "ad_personalization"},
		PostHogConsent:  false,
	},
	{
		Name:            "Functional",
		Slug:            "functional",
		Description:     "Cookies that enable enhanced functionality and personalization.",
		Kind:            coredata.CookieCategoryKindNormal,
		Rank:            3,
		GCMConsentTypes: []string{"functionality_storage", "personalization_storage"},
		PostHogConsent:  false,
	},
	{
		Name:            "Uncategorised",
		Slug:            "uncategorised",
		Description:     "Cookies that have not been assigned to a category yet.",
		Kind:            coredata.CookieCategoryKindUncategorised,
		Rank:            4,
		GCMConsentTypes: nil,
		PostHogConsent:  false,
	},
}

var defaultCategoryTranslationsByLanguage = map[string]map[string]struct {
	Name        string
	Description string
}{
	"fr": {
		"necessary":     {"Nécessaires", "Cookies essentiels au bon fonctionnement du site web."},
		"analytics":     {"Analytiques", "Cookies qui aident à comprendre comment les visiteurs interagissent avec le site web."},
		"advertising":   {"Publicitaires", "Cookies utilisés pour diffuser des publicités pertinentes et suivre les campagnes."},
		"functional":    {"Fonctionnels", "Cookies qui permettent des fonctionnalités améliorées et la personnalisation."},
		"uncategorised": {"Non classés", "Cookies qui n'ont pas encore été assignés à une catégorie."},
	},
	"de": {
		"necessary":     {"Notwendige", "Wesentliche Cookies, die für das ordnungsgemäße Funktionieren der Website erforderlich sind."},
		"analytics":     {"Analytische", "Cookies, die helfen zu verstehen, wie Besucher mit der Website interagieren."},
		"advertising":   {"Werbe", "Cookies, die verwendet werden, um relevante Werbung zu liefern und Kampagnen zu verfolgen."},
		"functional":    {"Funktionale", "Cookies, die erweiterte Funktionalität und Personalisierung ermöglichen."},
		"uncategorised": {"Nicht kategorisiert", "Cookies, die noch keiner Kategorie zugeordnet wurden."},
	},
	"es": {
		"necessary":     {"Necesarias", "Cookies esenciales necesarias para el correcto funcionamiento del sitio web."},
		"analytics":     {"Analíticas", "Cookies que ayudan a entender cómo los visitantes interactúan con el sitio web."},
		"advertising":   {"Publicitarias", "Cookies utilizadas para mostrar anuncios relevantes y rastrear campañas."},
		"functional":    {"Funcionales", "Cookies que permiten funcionalidades mejoradas y personalización."},
		"uncategorised": {"Sin categoría", "Cookies que aún no han sido asignadas a una categoría."},
	},
}

var defaultUIStringsByLanguage = map[string]map[string]string{
	"en": {
		"banner_title":             "Cookie Preferences",
		"banner_description":       "We use cookies to improve your experience and analyze site traffic. {{cookie_policy_link}}",
		"button_accept_all":        "Accept all",
		"button_reject_all":        "Reject all",
		"button_customize":         "Customize",
		"button_save":              "Save preferences",
		"panel_title":              "Customise Preferences",
		"panel_description":        "Choose which cookie categories to allow. {{necessary_category}} cookies are always active as they are needed for the site to work.",
		"aria_close":               "Close",
		"aria_show_details":        "Show cookie details",
		"aria_hide_details":        "Hide cookie details",
		"aria_cookie_settings":     "Cookie settings",
		"privacy_policy_link_text": "Privacy Policy",
		"cookie_policy_link_text":  "Cookie Policy",
		"placeholder_text":         "This content requires {{category}} cookies.",
		"placeholder_button":       "Manage cookie preferences",
	},
	"fr": {
		"banner_title":             "Préférences de cookies",
		"banner_description":       "Nous utilisons des cookies pour améliorer votre expérience et analyser le trafic du site. {{cookie_policy_link}}",
		"button_accept_all":        "Tout accepter",
		"button_reject_all":        "Tout refuser",
		"button_customize":         "Personnaliser",
		"button_save":              "Enregistrer les préférences",
		"panel_title":              "Personnaliser les préférences",
		"panel_description":        "Choisissez les catégories de cookies à autoriser. Les cookies {{necessary_category}} sont toujours actifs car ils sont nécessaires au fonctionnement du site.",
		"aria_close":               "Fermer",
		"aria_show_details":        "Afficher les détails des cookies",
		"aria_hide_details":        "Masquer les détails des cookies",
		"aria_cookie_settings":     "Paramètres des cookies",
		"privacy_policy_link_text": "Politique de confidentialité",
		"cookie_policy_link_text":  "Politique relative aux cookies",
		"placeholder_text":         "Ce contenu nécessite les cookies {{category}}.",
		"placeholder_button":       "Gérer les préférences de cookies",
	},
	"de": {
		"banner_title":             "Cookie-Einstellungen",
		"banner_description":       "Wir verwenden Cookies, um Ihre Erfahrung zu verbessern und den Website-Verkehr zu analysieren. {{cookie_policy_link}}",
		"button_accept_all":        "Alle akzeptieren",
		"button_reject_all":        "Alle ablehnen",
		"button_customize":         "Anpassen",
		"button_save":              "Einstellungen speichern",
		"panel_title":              "Einstellungen anpassen",
		"panel_description":        "Wählen Sie, welche Cookie-Kategorien zugelassen werden sollen. {{necessary_category}}-Cookies sind immer aktiv, da sie für den Betrieb der Website erforderlich sind.",
		"aria_close":               "Schließen",
		"aria_show_details":        "Cookie-Details anzeigen",
		"aria_hide_details":        "Cookie-Details ausblenden",
		"aria_cookie_settings":     "Cookie-Einstellungen",
		"privacy_policy_link_text": "Datenschutzrichtlinie",
		"cookie_policy_link_text":  "Cookie-Richtlinie",
		"placeholder_text":         "Dieser Inhalt erfordert {{category}}-Cookies.",
		"placeholder_button":       "Cookie-Einstellungen verwalten",
	},
	"es": {
		"banner_title":             "Preferencias de cookies",
		"banner_description":       "Utilizamos cookies para mejorar su experiencia y analizar el tráfico del sitio. {{cookie_policy_link}}",
		"button_accept_all":        "Aceptar todo",
		"button_reject_all":        "Rechazar todo",
		"button_customize":         "Personalizar",
		"button_save":              "Guardar preferencias",
		"panel_title":              "Personalizar preferencias",
		"panel_description":        "Elija qué categorías de cookies permitir. Las cookies {{necessary_category}} siempre están activas ya que son necesarias para el funcionamiento del sitio.",
		"aria_close":               "Cerrar",
		"aria_show_details":        "Mostrar detalles de cookies",
		"aria_hide_details":        "Ocultar detalles de cookies",
		"aria_cookie_settings":     "Configuración de cookies",
		"privacy_policy_link_text": "Política de privacidad",
		"cookie_policy_link_text":  "Política de cookies",
		"placeholder_text":         "Este contenido requiere cookies de {{category}}.",
		"placeholder_button":       "Gestionar preferencias de cookies",
	},
}
