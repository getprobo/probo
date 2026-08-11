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

-- Clean cookie-banner UI keys: generic wording under button_opt_out, CCPA
-- statutory CTA under button_opt_out_ccpa, Privacy Choices under *_ccpa.
-- Also backfill missing us/ca description overlays for older banners.

DO $migrate$
DECLARE
  shipped text[] := ARRAY[
    'Do Not Sell or Share My Personal Information',
    'Jangan Jual atau Bagikan Informasi Pribadi Saya',
    'Kişisel Bilgilerimi Satma veya Paylaşma',
    'Meine persönlichen Daten nicht verkaufen oder weitergeben',
    'Ne pas vendre ni partager mes informations personnelles',
    'Nie sprzedawaj ani nie udostępniaj moich danych osobowych',
    'No vender ni compartir mi información personal',
    'Non vendere né condividere le mie informazioni personali',
    'Não vender nem compartilhar minhas informações pessoais',
    'Verkoop of deel mijn persoonlijke gegevens niet',
    'Не продавати та не ділитися моїми персональними даними',
    '不要出售或共享我的个人信息',
    '個人情報の販売・共有を拒否する',
    '내 개인정보 판매 또는 공유 금지'
  ];
  r RECORD;
  t jsonb;
  defaults jsonb;
  opt_out text;
  generic text;
  is_shipped boolean;
BEGIN
  FOR r IN SELECT id, language, translations FROM cookie_banner_translations
  LOOP
    t := r.translations;
    IF t IS NULL THEN
      CONTINUE;
    END IF;

    defaults := CASE r.language
    WHEN 'de' THEN '{"button_opt_out": "Nicht wesentliche Cookies ablehnen", "button_opt_out_ccpa": "Meine persönlichen Daten nicht verkaufen oder weitergeben", "button_reject_all": "Alle ablehnen", "banner_description_us_opt_out": "Wir verwenden Cookies und ähnliche Technologien für Analysen und Werbung. Nach den geltenden US-Bundesstaats-Datenschutzgesetzen können Sie dem Verkauf oder der Weitergabe Ihrer personenbezogenen Daten und gezielter Werbung widersprechen. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Wir verwenden Cookies und ähnliche Technologien. Nach den geltenden kanadischen Datenschutzgesetzen können Sie nicht wesentliche Cookies jederzeit ablehnen. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Ihre Datenschutzoptionen", "privacy_choices_intro_ccpa": "Das kalifornische Recht gibt Ihnen das Recht zu kontrollieren, wie wir Ihre personenbezogenen Daten verwenden und weitergeben. Nutzen Sie die folgenden Optionen, um diese Rechte auszuüben. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Verkauf oder Weitergabe personenbezogener Daten ablehnen", "privacy_choices_sale_description_ccpa": "Sie haben das Recht, dem Verkauf oder der Weitergabe Ihrer personenbezogenen Daten zu widersprechen. Wenn Sie „Meine persönlichen Daten nicht verkaufen oder weitergeben“ wählen, stellen wir den Verkauf oder die Weitergabe Ihrer personenbezogenen Daten für diesen Browser ein.", "privacy_choices_spi_title_ccpa": "Verwendung sensibler personenbezogener Daten einschränken", "privacy_choices_spi_description_ccpa": "Sie haben das Recht, die Verwendung und Offenlegung Ihrer sensiblen personenbezogenen Daten einzuschränken. Wir verwenden oder offenbaren sensible personenbezogene Daten nicht für Zwecke, die über die nach kalifornischem Recht zulässigen hinausgehen."}'::jsonb
    WHEN 'en' THEN '{"button_opt_out": "Reject non-essential cookies", "button_opt_out_ccpa": "Do Not Sell or Share My Personal Information", "button_reject_all": "Reject all", "banner_description_us_opt_out": "We use cookies and similar technologies for analytics and advertising. Under applicable US state privacy laws, you may opt out of the sale or sharing of your personal information and targeted advertising. {{cookie_policy_link}}", "banner_description_ca_opt_out": "We use cookies and similar technologies. Under applicable Canadian privacy law, you may opt out of non-essential cookies at any time. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Your Privacy Choices", "privacy_choices_intro_ccpa": "California law gives you the right to control how we use and share your personal information. Use the options below to exercise those rights. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Opt out of the sale or sharing of personal information", "privacy_choices_sale_description_ccpa": "You have the right to opt out of the sale or sharing of your personal information. Choosing Do Not Sell or Share My Personal Information will stop us from selling or sharing your personal information for this browser.", "privacy_choices_spi_title_ccpa": "Limit the use of sensitive personal information", "privacy_choices_spi_description_ccpa": "You have the right to limit the use and disclosure of your sensitive personal information. We do not use or disclose sensitive personal information for purposes other than those permitted by California law."}'::jsonb
    WHEN 'es' THEN '{"button_opt_out": "Rechazar cookies no esenciales", "button_opt_out_ccpa": "No vender ni compartir mi información personal", "button_reject_all": "Rechazar todo", "banner_description_us_opt_out": "Utilizamos cookies y tecnologías similares para análisis y publicidad. Según las leyes estatales de privacidad de EE. UU. aplicables, puede optar por no participar en la venta o el intercambio de su información personal y en la publicidad dirigida. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Utilizamos cookies y tecnologías similares. Según las leyes canadenses de privacidad aplicables, puede optar por no usar cookies no esenciales en cualquier momento. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Sus opciones de privacidad", "privacy_choices_intro_ccpa": "La ley de California le otorga el derecho a controlar cómo usamos y compartimos su información personal. Use las opciones a continuación para ejercer esos derechos. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Exclusión de la venta o el intercambio de información personal", "privacy_choices_sale_description_ccpa": "Tiene derecho a optar por no participar en la venta o el intercambio de su información personal. Elegir No vender ni compartir mi información personal detendrá la venta o el intercambio de su información personal en este navegador.", "privacy_choices_spi_title_ccpa": "Limitar el uso de información personal sensible", "privacy_choices_spi_description_ccpa": "Tiene derecho a limitar el uso y la divulgación de su información personal sensible. No usamos ni divulgamos información personal sensible para fines distintos de los permitidos por la ley de California."}'::jsonb
    WHEN 'fr' THEN '{"button_opt_out": "Refuser les cookies non essentiels", "button_opt_out_ccpa": "Ne pas vendre ni partager mes informations personnelles", "button_reject_all": "Tout refuser", "banner_description_us_opt_out": "Nous utilisons des cookies et technologies similaires pour l''analyse et la publicité. En vertu des lois étatiques américaines sur la confidentialité applicables, vous pouvez refuser la vente ou le partage de vos informations personnelles et la publicité ciblée. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Nous utilisons des cookies et technologies similaires. En vertu des lois canadiennes sur la confidentialité applicables, vous pouvez refuser les cookies non essentiels à tout moment. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Vos choix en matière de confidentialité", "privacy_choices_intro_ccpa": "La loi californienne vous donne le droit de contrôler la façon dont nous utilisons et partageons vos informations personnelles. Utilisez les options ci-dessous pour exercer ces droits. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Refuser la vente ou le partage des informations personnelles", "privacy_choices_sale_description_ccpa": "Vous avez le droit de refuser la vente ou le partage de vos informations personnelles. Choisir Ne pas vendre ni partager mes informations personnelles arrêtera la vente ou le partage de vos informations personnelles pour ce navigateur.", "privacy_choices_spi_title_ccpa": "Limiter l''utilisation des informations personnelles sensibles", "privacy_choices_spi_description_ccpa": "Vous avez le droit de limiter l''utilisation et la divulgation de vos informations personnelles sensibles. Nous n''utilisons ni ne divulguons d''informations personnelles sensibles à des fins autres que celles autorisées par la loi californienne."}'::jsonb
    WHEN 'id' THEN '{"button_opt_out": "Tolak cookie yang tidak penting", "button_opt_out_ccpa": "Jangan Jual atau Bagikan Informasi Pribadi Saya", "button_reject_all": "Tolak semua", "banner_description_us_opt_out": "Kami menggunakan cookie dan teknologi serupa untuk analitik dan periklanan. Berdasarkan undang-undang privasi negara bagian AS yang berlaku, Anda dapat menolak penjualan atau pembagian informasi pribadi Anda serta iklan yang ditargetkan. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Kami menggunakan cookie dan teknologi serupa. Berdasarkan undang-undang privasi Kanada yang berlaku, Anda dapat menolak cookie yang tidak penting kapan saja. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Pilihan Privasi Anda", "privacy_choices_intro_ccpa": "Hukum California memberi Anda hak untuk mengontrol bagaimana kami menggunakan dan membagikan informasi pribadi Anda. Gunakan opsi di bawah untuk menggunakan hak tersebut. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Tolak penjualan atau pembagian informasi pribadi", "privacy_choices_sale_description_ccpa": "Anda berhak menolak penjualan atau pembagian informasi pribadi Anda. Memilih Jangan Jual atau Bagikan Informasi Pribadi Saya akan menghentikan penjualan atau pembagian informasi pribadi Anda untuk browser ini.", "privacy_choices_spi_title_ccpa": "Batasi penggunaan informasi pribadi sensitif", "privacy_choices_spi_description_ccpa": "Anda berhak membatasi penggunaan dan pengungkapan informasi pribadi sensitif Anda. Kami tidak menggunakan atau mengungkapkan informasi pribadi sensitif untuk tujuan selain yang diizinkan oleh hukum California."}'::jsonb
    WHEN 'it' THEN '{"button_opt_out": "Rifiuta i cookie non essenziali", "button_opt_out_ccpa": "Non vendere né condividere le mie informazioni personali", "button_reject_all": "Rifiuta tutto", "banner_description_us_opt_out": "Utilizziamo cookie e tecnologie simili per analisi e pubblicità. Ai sensi delle leggi statali statunitensi sulla privacy applicabili, puoi rinunciare alla vendita o alla condivisione delle tue informazioni personali e alla pubblicità mirata. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Utilizziamo cookie e tecnologie simili. Ai sensi delle leggi canadesi sulla privacy applicabili, puoi rinunciare ai cookie non essenziali in qualsiasi momento. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Le tue scelte sulla privacy", "privacy_choices_intro_ccpa": "La legge della California ti dà il diritto di controllare come utilizziamo e condividiamo le tue informazioni personali. Usa le opzioni qui sotto per esercitare tali diritti. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Rinuncia alla vendita o condivisione delle informazioni personali", "privacy_choices_sale_description_ccpa": "Hai il diritto di rinunciare alla vendita o alla condivisione delle tue informazioni personali. Scegliendo Non vendere né condividere le mie informazioni personali interromperemo la vendita o la condivisione delle tue informazioni personali per questo browser.", "privacy_choices_spi_title_ccpa": "Limita l''uso delle informazioni personali sensibili", "privacy_choices_spi_description_ccpa": "Hai il diritto di limitare l''uso e la divulgazione delle tue informazioni personali sensibili. Non utilizziamo né divulghiamo informazioni personali sensibili per scopi diversi da quelli consentiti dalla legge della California."}'::jsonb
    WHEN 'ja' THEN '{"button_opt_out": "不要なクッキーを拒否する", "button_opt_out_ccpa": "個人情報の販売・共有を拒否する", "button_reject_all": "すべて拒否", "banner_description_us_opt_out": "当社は分析および広告のためにクッキーおよび類似技術を使用しています。適用される米国州プライバシー法に基づき、個人情報の販売または共有およびターゲット広告をオプトアウトできます。{{cookie_policy_link}}", "banner_description_ca_opt_out": "当社はクッキーおよび類似技術を使用しています。適用されるカナダのプライバシー法に基づき、不要なクッキーをいつでもオプトアウトできます。{{cookie_policy_link}}", "privacy_choices_title_ccpa": "プライバシーに関する選択", "privacy_choices_intro_ccpa": "カリフォルニア州法により、お客様は個人情報の利用および共有方法を管理する権利があります。以下のオプションを使用してこれらの権利を行使してください。{{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "個人情報の販売または共有のオプトアウト", "privacy_choices_sale_description_ccpa": "お客様には、個人情報の販売または共有を拒否する権利があります。「個人情報の販売・共有を拒否する」を選択すると、このブラウザでの個人情報の販売または共有を停止します。", "privacy_choices_spi_title_ccpa": "機微な個人情報の利用の制限", "privacy_choices_spi_description_ccpa": "お客様には、機微な個人情報の利用および開示を制限する権利があります。当社は、カリフォルニア州法で認められる目的以外で機微な個人情報を利用または開示しません。"}'::jsonb
    WHEN 'ko' THEN '{"button_opt_out": "비필수 쿠키 거부", "button_opt_out_ccpa": "내 개인정보 판매 또는 공유 금지", "button_reject_all": "모두 거부", "banner_description_us_opt_out": "저희는 분석 및 광고를 위해 쿠키 및 유사 기술을 사용합니다. 적용되는 미국 주 개인정보 보호법에 따라 개인정보의 판매 또는 공유 및 맞춤형 광고를 거부할 수 있습니다. {{cookie_policy_link}}", "banner_description_ca_opt_out": "저희는 쿠키 및 유사 기술을 사용합니다. 적용되는 캐나다 개인정보 보호법에 따라 언제든지 필수적이지 않은 쿠키를 거부할 수 있습니다. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "개인정보 선택 사항", "privacy_choices_intro_ccpa": "캘리포니아 법은 저희가 개인정보를 사용하고 공유하는 방식을 통제할 권리를 부여합니다. 아래 옵션을 사용하여 해당 권리를 행사하세요. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "개인정보 판매 또는 공유 거부", "privacy_choices_sale_description_ccpa": "귀하는 개인정보의 판매 또는 공유를 거부할 권리가 있습니다. 내 개인정보 판매 또는 공유 금지를 선택하면 이 브라우저에서 개인정보 판매 또는 공유를 중단합니다.", "privacy_choices_spi_title_ccpa": "민감한 개인정보 사용 제한", "privacy_choices_spi_description_ccpa": "귀하는 민감한 개인정보의 사용 및 공개를 제한할 권리가 있습니다. 저희는 캘리포니아 법이 허용하는 목적 외로는 민감한 개인정보를 사용하거나 공개하지 않습니다."}'::jsonb
    WHEN 'nl' THEN '{"button_opt_out": "Niet-essentiële cookies weigeren", "button_opt_out_ccpa": "Verkoop of deel mijn persoonlijke gegevens niet", "button_reject_all": "Alles weigeren", "banner_description_us_opt_out": "Wij gebruiken cookies en vergelijkbare technologieën voor analyse en advertenties. Op grond van toepasselijke Amerikaanse staatsprivacywetten kunt u zich afmelden voor de verkoop of het delen van uw persoonlijke gegevens en gerichte advertenties. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Wij gebruiken cookies en vergelijkbare technologieën. Op grond van toepasselijke Canadese privacywetgeving kunt u zich op elk moment afmelden voor niet-essentiële cookies. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Uw privacykeuzes", "privacy_choices_intro_ccpa": "De Californische wet geeft je het recht om te bepalen hoe wij je persoonlijke gegevens gebruiken en delen. Gebruik de onderstaande opties om die rechten uit te oefenen. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Verkoop of delen van persoonlijke gegevens weigeren", "privacy_choices_sale_description_ccpa": "Je hebt het recht om je af te melden voor de verkoop of het delen van je persoonlijke gegevens. Als je Verkoop of deel mijn persoonlijke gegevens niet kiest, stoppen wij met het verkopen of delen van je persoonlijke gegevens voor deze browser.", "privacy_choices_spi_title_ccpa": "Gebruik van gevoelige persoonlijke gegevens beperken", "privacy_choices_spi_description_ccpa": "Je hebt het recht om het gebruik en de openbaarmaking van je gevoelige persoonlijke gegevens te beperken. Wij gebruiken of openbaren geen gevoelige persoonlijke gegevens voor andere doeleinden dan die zijn toegestaan door de Californische wet."}'::jsonb
    WHEN 'pl' THEN '{"button_opt_out": "Odrzuć nieistotne pliki cookie", "button_opt_out_ccpa": "Nie sprzedawaj ani nie udostępniaj moich danych osobowych", "button_reject_all": "Odrzuć wszystkie", "banner_description_us_opt_out": "Używamy plików cookie i podobnych technologii do analityki i reklam. Zgodnie z obowiązującymi amerykańskimi przepisami o prywatności stanowej możesz zrezygnować ze sprzedaży lub udostępniania swoich danych osobowych oraz reklam ukierunkowanych. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Używamy plików cookie i podobnych technologii. Zgodnie z obowiązującymi kanadyjskimi przepisami o prywatności możesz w każdej chwili zrezygnować z nieistotnych plików cookie. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Twoje wybory dotyczące prywatności", "privacy_choices_intro_ccpa": "Prawo Kalifornii daje Ci prawo do kontrolowania, w jaki sposób wykorzystujemy i udostępniamy Twoje dane osobowe. Skorzystaj z poniższych opcji, aby skorzystać z tych praw. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Rezygnacja ze sprzedaży lub udostępniania danych osobowych", "privacy_choices_sale_description_ccpa": "Masz prawo zrezygnować ze sprzedaży lub udostępniania swoich danych osobowych. Wybór Nie sprzedawaj ani nie udostępniaj moich danych osobowych zatrzyma sprzedaż lub udostępnianie Twoich danych osobowych w tej przeglądarce.", "privacy_choices_spi_title_ccpa": "Ograniczenie wykorzystywania wrażliwych danych osobowych", "privacy_choices_spi_description_ccpa": "Masz prawo ograniczyć wykorzystywanie i ujawnianie swoich wrażliwych danych osobowych. Nie wykorzystujemy ani nie ujawniamy wrażliwych danych osobowych w celach innych niż dozwolone przez prawo Kalifornii."}'::jsonb
    WHEN 'pt' THEN '{"button_opt_out": "Recusar cookies não essenciais", "button_opt_out_ccpa": "Não vender nem compartilhar minhas informações pessoais", "button_reject_all": "Rejeitar tudo", "banner_description_us_opt_out": "Utilizamos cookies e tecnologias similares para análise e publicidade. De acordo com as leis estaduais de privacidade dos EUA aplicáveis, você pode recusar a venda ou o compartilhamento de suas informações pessoais e a publicidade direcionada. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Utilizamos cookies e tecnologias similares. De acordo com as leis canadenses de privacidade aplicáveis, você pode recusar cookies não essenciais a qualquer momento. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Suas escolhas de privacidade", "privacy_choices_intro_ccpa": "A lei da Califórnia dá a você o direito de controlar como usamos e compartilhamos suas informações pessoais. Use as opções abaixo para exercer esses direitos. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Recusar a venda ou o compartilhamento de informações pessoais", "privacy_choices_sale_description_ccpa": "Você tem o direito de recusar a venda ou o compartilhamento de suas informações pessoais. Escolher Não vender nem compartilhar minhas informações pessoais interromperá a venda ou o compartilhamento de suas informações pessoais neste navegador.", "privacy_choices_spi_title_ccpa": "Limitar o uso de informações pessoais sensíveis", "privacy_choices_spi_description_ccpa": "Você tem o direito de limitar o uso e a divulgação de suas informações pessoais sensíveis. Não usamos nem divulgamos informações pessoais sensíveis para fins além dos permitidos pela lei da Califórnia."}'::jsonb
    WHEN 'tr' THEN '{"button_opt_out": "Zorunlu olmayan çerezleri reddet", "button_opt_out_ccpa": "Kişisel Bilgilerimi Satma veya Paylaşma", "button_reject_all": "Tümünü reddet", "banner_description_us_opt_out": "Analitik ve reklamcılık için çerezler ve benzer teknolojiler kullanıyoruz. Geçerli ABD eyalet gizlilik yasaları kapsamında kişisel bilgilerinizin satışından veya paylaşımından ve hedefli reklamcılıktan vazgeçebilirsiniz. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Çerezler ve benzer teknolojiler kullanıyoruz. Geçerli Kanada gizlilik yasaları kapsamında gerekli olmayan çerezlerden istediğiniz zaman vazgeçebilirsiniz. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Gizlilik tercihleriniz", "privacy_choices_intro_ccpa": "Kaliforniya yasası, kişisel bilgilerinizi nasıl kullandığımızı ve paylaştığımızı kontrol etme hakkı tanır. Bu hakları kullanmak için aşağıdaki seçenekleri kullanın. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Kişisel bilgilerin satışından veya paylaşımından vazgeçme", "privacy_choices_sale_description_ccpa": "Kişisel bilgilerinizin satışından veya paylaşımından vazgeçme hakkına sahipsiniz. Kişisel Bilgilerimi Satma veya Paylaşma''yı seçmek, bu tarayıcı için kişisel bilgilerinizin satışını veya paylaşımını durdurur.", "privacy_choices_spi_title_ccpa": "Hassas kişisel bilgilerin kullanımını sınırlama", "privacy_choices_spi_description_ccpa": "Hassas kişisel bilgilerinizin kullanımını ve ifşasını sınırlama hakkına sahipsiniz. Hassas kişisel bilgileri Kaliforniya yasasının izin verdiği amaçlar dışında kullanmayız veya ifşa etmeyiz."}'::jsonb
    WHEN 'uk' THEN '{"button_opt_out": "Відхилити неістотні файли cookie", "button_opt_out_ccpa": "Не продавати та не ділитися моїми персональними даними", "button_reject_all": "Відхилити всі", "banner_description_us_opt_out": "Ми використовуємо файли cookie та подібні технології для аналітики та реклами. Відповідно до чинних законів штатів США про конфіденційність ви можете відмовитися від продажу або поширення ваших персональних даних і таргетованої реклами. {{cookie_policy_link}}", "banner_description_ca_opt_out": "Ми використовуємо файли cookie та подібні технології. Відповідно до чинного канадського законодавства про конфіденційність ви можете відмовитися від неосновних файлів cookie у будь-який час. {{cookie_policy_link}}", "privacy_choices_title_ccpa": "Ваш вибір щодо конфіденційності", "privacy_choices_intro_ccpa": "Закон Каліфорнії дає вам право контролювати, як ми використовуємо та поширюємо ваші персональні дані. Скористайтеся наведеними нижче параметрами, щоб реалізувати ці права. {{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "Відмова від продажу або поширення персональних даних", "privacy_choices_sale_description_ccpa": "Ви маєте право відмовитися від продажу або поширення ваших персональних даних. Вибір «Не продавати та не ділитися моїми персональними даними» зупинить продаж або поширення ваших персональних даних для цього браузера.", "privacy_choices_spi_title_ccpa": "Обмеження використання конфіденційних персональних даних", "privacy_choices_spi_description_ccpa": "Ви маєте право обмежити використання та розкриття ваших конфіденційних персональних даних. Ми не використовуємо та не розкриваємо конфіденційні персональні дані для цілей, не дозволених законом Каліфорнії."}'::jsonb
    WHEN 'zh' THEN '{"button_opt_out": "拒绝非必要 Cookie", "button_opt_out_ccpa": "不要出售或共享我的个人信息", "button_reject_all": "全部拒绝", "banner_description_us_opt_out": "我们使用 Cookie 和类似技术进行分析和广告。根据适用的美国州隐私法，您可以选择退出出售或共享您的个人信息以及定向广告。{{cookie_policy_link}}", "banner_description_ca_opt_out": "我们使用 Cookie 和类似技术。根据适用的加拿大隐私法，您可以随时选择退出非必要 Cookie。{{cookie_policy_link}}", "privacy_choices_title_ccpa": "您的隐私选择", "privacy_choices_intro_ccpa": "加州法律赋予您控制我们如何使用和共享您的个人信息的权利。请使用以下选项行使这些权利。{{privacy_policy_link}}", "privacy_choices_sale_title_ccpa": "选择退出出售或共享个人信息", "privacy_choices_sale_description_ccpa": "您有权选择退出出售或共享您的个人信息。选择“不要出售或共享我的个人信息”将停止我们在此浏览器中出售或共享您的个人信息。", "privacy_choices_spi_title_ccpa": "限制敏感个人信息的使用", "privacy_choices_spi_description_ccpa": "您有权限制敏感个人信息的使用和披露。我们不会将敏感个人信息用于加州法律允许范围以外的目的。"}'::jsonb
      ELSE '{}'::jsonb
    END;

    -- Rename Privacy Choices keys to *_ccpa when the old key is present.

    IF t ? 'privacy_choices_title' AND NOT (t ? 'privacy_choices_title_ccpa') THEN
      t := t || jsonb_build_object('privacy_choices_title_ccpa', t->'privacy_choices_title');
    END IF;
    t := t - 'privacy_choices_title';

    IF t ? 'privacy_choices_intro' AND NOT (t ? 'privacy_choices_intro_ccpa') THEN
      t := t || jsonb_build_object('privacy_choices_intro_ccpa', t->'privacy_choices_intro');
    END IF;
    t := t - 'privacy_choices_intro';

    IF t ? 'privacy_choices_sale_title' AND NOT (t ? 'privacy_choices_sale_title_ccpa') THEN
      t := t || jsonb_build_object('privacy_choices_sale_title_ccpa', t->'privacy_choices_sale_title');
    END IF;
    t := t - 'privacy_choices_sale_title';

    IF t ? 'privacy_choices_sale_description' AND NOT (t ? 'privacy_choices_sale_description_ccpa') THEN
      t := t || jsonb_build_object('privacy_choices_sale_description_ccpa', t->'privacy_choices_sale_description');
    END IF;
    t := t - 'privacy_choices_sale_description';

    IF t ? 'privacy_choices_spi_title' AND NOT (t ? 'privacy_choices_spi_title_ccpa') THEN
      t := t || jsonb_build_object('privacy_choices_spi_title_ccpa', t->'privacy_choices_spi_title');
    END IF;
    t := t - 'privacy_choices_spi_title';

    IF t ? 'privacy_choices_spi_description' AND NOT (t ? 'privacy_choices_spi_description_ccpa') THEN
      t := t || jsonb_build_object('privacy_choices_spi_description_ccpa', t->'privacy_choices_spi_description');
    END IF;
    t := t - 'privacy_choices_spi_description';

    opt_out := t->>'button_opt_out';
    generic := t->>'button_opt_out_generic';
    is_shipped := opt_out IS NOT NULL AND opt_out = ANY (shipped);

    -- Move shipped DNSMPI CTA into button_opt_out_ccpa.
    IF is_shipped AND (NOT (t ? 'button_opt_out_ccpa') OR COALESCE(t->>'button_opt_out_ccpa', '') = '') THEN
      t := t || jsonb_build_object('button_opt_out_ccpa', to_jsonb(opt_out));
    END IF;

    -- When the stored CTA is still Probo's California default, replace it
    -- with the neutral label. Integrator-customized button_opt_out is left
    -- alone; button_opt_out_ccpa is backfilled from defaults below.
    IF is_shipped THEN
      IF generic IS NOT NULL AND generic <> '' THEN
        t := t || jsonb_build_object('button_opt_out', to_jsonb(generic));
      ELSIF defaults ? 'button_opt_out' THEN
        t := t || jsonb_build_object('button_opt_out', defaults->'button_opt_out');
      ELSIF t ? 'button_reject_all' THEN
        t := t || jsonb_build_object('button_opt_out', t->'button_reject_all');
      END IF;
    END IF;

    t := t - 'button_opt_out_generic';

    -- Backfill missing jurisdiction / CCPA keys from language defaults.
    IF defaults <> '{}'::jsonb THEN
      t := defaults || t;
    END IF;

    UPDATE cookie_banner_translations
    SET translations = t,
        updated_at = NOW()
    WHERE id = r.id;
  END LOOP;
END;
$migrate$;
