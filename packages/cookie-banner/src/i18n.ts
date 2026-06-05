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

export interface BannerTexts {
  [key: string]: string;
}

function normalizeLocale(locale: string): string {
  return locale.split("-")[0].toLowerCase();
}

export function detectLanguage(explicit?: string): string {
  if (explicit) return normalizeLocale(explicit);

  if (typeof document !== "undefined" && document.documentElement) {
    const htmlLang = document.documentElement.lang;
    if (htmlLang) return normalizeLocale(htmlLang);
  }

  if (typeof navigator !== "undefined" && navigator.language) {
    return normalizeLocale(navigator.language);
  }

  return "";
}

export function interpolate(
  template: string,
  vars: Record<string, string>,
): string {
  return template.replace(/\{\{(\w+)\}\}/g, (_, key) => vars[key] ?? "");
}

const COOKIE_DETAIL_LABELS: Record<string, Record<string, string>> = {
  en: { label_type: "Type: {{value}}", label_description: "Description: {{value}}", label_duration: "Duration: {{value}}" },
  de: { label_type: "Typ: {{value}}", label_description: "Beschreibung: {{value}}", label_duration: "Dauer: {{value}}" },
  es: { label_type: "Tipo: {{value}}", label_description: "Descripción: {{value}}", label_duration: "Duración: {{value}}" },
  fr: { label_type: "Type : {{value}}", label_description: "Description : {{value}}", label_duration: "Durée : {{value}}" },
  id: { label_type: "Tipe: {{value}}", label_description: "Deskripsi: {{value}}", label_duration: "Durasi: {{value}}" },
  it: { label_type: "Tipo: {{value}}", label_description: "Descrizione: {{value}}", label_duration: "Durata: {{value}}" },
  ja: { label_type: "タイプ：{{value}}", label_description: "説明：{{value}}", label_duration: "期間：{{value}}" },
  ko: { label_type: "유형: {{value}}", label_description: "설명: {{value}}", label_duration: "기간: {{value}}" },
  pl: { label_type: "Typ: {{value}}", label_description: "Opis: {{value}}", label_duration: "Czas trwania: {{value}}" },
  pt: { label_type: "Tipo: {{value}}", label_description: "Descrição: {{value}}", label_duration: "Duração: {{value}}" },
  tr: { label_type: "Tür: {{value}}", label_description: "Açıklama: {{value}}", label_duration: "Süre: {{value}}" },
  uk: { label_type: "Тип: {{value}}", label_description: "Опис: {{value}}", label_duration: "Тривалість: {{value}}" },
  zh: { label_type: "类型：{{value}}", label_description: "描述：{{value}}", label_duration: "时长：{{value}}" },
};

export function getCookieDetailLabels(lang: string): Record<string, string> {
  return COOKIE_DETAIL_LABELS[lang] ?? COOKIE_DETAIL_LABELS.en;
}

// Tracker type names are Web platform API names (proper nouns), so they are
// not translated; only the surrounding "Type:" label is localized.
const TRACKER_TYPE_LABELS: Record<string, string> = {
  COOKIE: "Cookie",
  LOCAL_STORAGE: "Local storage",
  SESSION_STORAGE: "Session storage",
  INDEXED_DB: "IndexedDB",
  CACHE_STORAGE: "Cache storage",
};

export function getTrackerTypeLabel(type: string): string {
  return TRACKER_TYPE_LABELS[type] ?? type;
}

const GPC_LABELS: Record<string, string> = {
  en: "Opt-Out Preference Signal Honored",
  de: "Opt-Out-Präferenzsignal beachtet",
  es: "Señal de exclusión respetada",
  fr: "Signal de préférence de désinscription respecté",
  id: "Sinyal Preferensi Opt-Out Dihormati",
  it: "Segnale di preferenza di rinuncia rispettato",
  ja: "オプトアウト設定シグナルが有効です",
  ko: "옵트아웃 기본 설정 신호가 적용되었습니다",
  pl: "Sygnał preferencji rezygnacji uhonorowany",
  pt: "Sinal de preferência de exclusão respeitado",
  tr: "Çıkış Tercihi Sinyali Onurlandırıldı",
  uk: "Сигнал переваги відмови враховано",
  zh: "退出偏好信号已生效",
};

export function getGpcLabel(lang: string): string {
  return GPC_LABELS[lang] ?? GPC_LABELS.en;
}
