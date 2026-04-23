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

var defaultCategories = []struct {
	Name        string
	Description string
	Kind        coredata.CookieCategoryKind
	Rank        int
}{
	{"Necessary", "Essential cookies required for the website to function properly.", coredata.CookieCategoryKindNecessary, 0},
	{"Analytics", "Cookies that help understand how visitors interact with the website.", coredata.CookieCategoryKindNormal, 1},
	{"Advertising", "Cookies used to deliver relevant advertisements and track campaigns.", coredata.CookieCategoryKindNormal, 2},
	{"Functional", "Cookies that enable enhanced functionality and personalization.", coredata.CookieCategoryKindNormal, 3},
	{"Uncategorised", "Cookies that have not been assigned to a category yet.", coredata.CookieCategoryKindUncategorised, 4},
}

var defaultUIStringsByLanguage = map[string]map[string]string{
	"en": {
		"banner_title":             "Cookie Preferences",
		"banner_description":       "We use cookies to improve your experience and analyze site traffic. {{privacy_policy_link}}",
		"button_accept_all":        "Accept all",
		"button_reject_all":        "Reject all",
		"button_customize":         "Customize",
		"button_save":              "Save preferences",
		"panel_title":              "Customise Preferences",
		"panel_description":        "Choose which cookie categories to allow. {{necessary_category}} cookies are always active as they are needed for the site to work.",
		"label_description":        "Description: {{value}}",
		"label_duration":           "Duration: {{value}}",
		"aria_close":               "Close",
		"aria_show_details":        "Show cookie details",
		"aria_hide_details":        "Hide cookie details",
		"aria_cookie_settings":     "Cookie settings",
		"privacy_policy_link_text": "Privacy Policy",
		"placeholder_text":         "This content requires {{category}} cookies.",
		"placeholder_button":       "Manage cookie preferences",
	},
	"fr": {
		"banner_title":             "Préférences de cookies",
		"banner_description":       "Nous utilisons des cookies pour améliorer votre expérience et analyser le trafic du site. {{privacy_policy_link}}",
		"button_accept_all":        "Tout accepter",
		"button_reject_all":        "Tout refuser",
		"button_customize":         "Personnaliser",
		"button_save":              "Enregistrer les préférences",
		"panel_title":              "Personnaliser les préférences",
		"panel_description":        "Choisissez les catégories de cookies à autoriser. Les cookies {{necessary_category}} sont toujours actifs car ils sont nécessaires au fonctionnement du site.",
		"label_description":        "Description : {{value}}",
		"label_duration":           "Durée : {{value}}",
		"aria_close":               "Fermer",
		"aria_show_details":        "Afficher les détails des cookies",
		"aria_hide_details":        "Masquer les détails des cookies",
		"aria_cookie_settings":     "Paramètres des cookies",
		"privacy_policy_link_text": "Politique de confidentialité",
		"placeholder_text":         "Ce contenu nécessite les cookies {{category}}.",
		"placeholder_button":       "Gérer les préférences de cookies",
	},
	"de": {
		"banner_title":             "Cookie-Einstellungen",
		"banner_description":       "Wir verwenden Cookies, um Ihre Erfahrung zu verbessern und den Website-Verkehr zu analysieren. {{privacy_policy_link}}",
		"button_accept_all":        "Alle akzeptieren",
		"button_reject_all":        "Alle ablehnen",
		"button_customize":         "Anpassen",
		"button_save":              "Einstellungen speichern",
		"panel_title":              "Einstellungen anpassen",
		"panel_description":        "Wählen Sie, welche Cookie-Kategorien zugelassen werden sollen. {{necessary_category}}-Cookies sind immer aktiv, da sie für den Betrieb der Website erforderlich sind.",
		"label_description":        "Beschreibung: {{value}}",
		"label_duration":           "Dauer: {{value}}",
		"aria_close":               "Schließen",
		"aria_show_details":        "Cookie-Details anzeigen",
		"aria_hide_details":        "Cookie-Details ausblenden",
		"aria_cookie_settings":     "Cookie-Einstellungen",
		"privacy_policy_link_text": "Datenschutzrichtlinie",
		"placeholder_text":         "Dieser Inhalt erfordert {{category}}-Cookies.",
		"placeholder_button":       "Cookie-Einstellungen verwalten",
	},
	"es": {
		"banner_title":             "Preferencias de cookies",
		"banner_description":       "Utilizamos cookies para mejorar su experiencia y analizar el tráfico del sitio. {{privacy_policy_link}}",
		"button_accept_all":        "Aceptar todo",
		"button_reject_all":        "Rechazar todo",
		"button_customize":         "Personalizar",
		"button_save":              "Guardar preferencias",
		"panel_title":              "Personalizar preferencias",
		"panel_description":        "Elija qué categorías de cookies permitir. Las cookies {{necessary_category}} siempre están activas ya que son necesarias para el funcionamiento del sitio.",
		"label_description":        "Descripción: {{value}}",
		"label_duration":           "Duración: {{value}}",
		"aria_close":               "Cerrar",
		"aria_show_details":        "Mostrar detalles de cookies",
		"aria_hide_details":        "Ocultar detalles de cookies",
		"aria_cookie_settings":     "Configuración de cookies",
		"privacy_policy_link_text": "Política de privacidad",
		"placeholder_text":         "Este contenido requiere cookies de {{category}}.",
		"placeholder_button":       "Gestionar preferencias de cookies",
	},
}
