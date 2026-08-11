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

-- Backfill Privacy Choices panel copy (11 CCR § 7015) for existing banners.
-- New banners already receive these keys from defaultUIStringsByLanguage.

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Your Privacy Choices",
    "privacy_choices_intro": "California law gives you the right to control how we use and share your personal information. Use the options below to exercise those rights. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Opt out of the sale or sharing of personal information",
    "privacy_choices_sale_description": "You have the right to opt out of the sale or sharing of your personal information. Choosing Do Not Sell or Share My Personal Information will stop us from selling or sharing your personal information for this browser.",
    "privacy_choices_spi_title": "Limit the use of sensitive personal information",
    "privacy_choices_spi_description": "You have the right to limit the use and disclosure of your sensitive personal information. We do not use or disclose sensitive personal information for purposes other than those permitted by California law."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'en'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Vos choix en matière de confidentialité",
    "privacy_choices_intro": "La loi californienne vous donne le droit de contrôler la façon dont nous utilisons et partageons vos informations personnelles. Utilisez les options ci-dessous pour exercer ces droits. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Refuser la vente ou le partage des informations personnelles",
    "privacy_choices_sale_description": "Vous avez le droit de refuser la vente ou le partage de vos informations personnelles. Choisir Ne pas vendre ni partager mes informations personnelles arrêtera la vente ou le partage de vos informations personnelles pour ce navigateur.",
    "privacy_choices_spi_title": "Limiter l''utilisation des informations personnelles sensibles",
    "privacy_choices_spi_description": "Vous avez le droit de limiter l''utilisation et la divulgation de vos informations personnelles sensibles. Nous n''utilisons ni ne divulguons d''informations personnelles sensibles à des fins autres que celles autorisées par la loi californienne."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'fr'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Ihre Datenschutzoptionen",
    "privacy_choices_intro": "Das kalifornische Recht gibt Ihnen das Recht zu kontrollieren, wie wir Ihre personenbezogenen Daten verwenden und weitergeben. Nutzen Sie die folgenden Optionen, um diese Rechte auszuüben. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Verkauf oder Weitergabe personenbezogener Daten ablehnen",
    "privacy_choices_sale_description": "Sie haben das Recht, dem Verkauf oder der Weitergabe Ihrer personenbezogenen Daten zu widersprechen. Wenn Sie „Meine persönlichen Daten nicht verkaufen oder weitergeben“ wählen, stellen wir den Verkauf oder die Weitergabe Ihrer personenbezogenen Daten für diesen Browser ein.",
    "privacy_choices_spi_title": "Verwendung sensibler personenbezogener Daten einschränken",
    "privacy_choices_spi_description": "Sie haben das Recht, die Verwendung und Offenlegung Ihrer sensiblen personenbezogenen Daten einzuschränken. Wir verwenden oder offenbaren sensible personenbezogene Daten nicht für Zwecke, die über die nach kalifornischem Recht zulässigen hinausgehen."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'de'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Sus opciones de privacidad",
    "privacy_choices_intro": "La ley de California le otorga el derecho a controlar cómo usamos y compartimos su información personal. Use las opciones a continuación para ejercer esos derechos. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Exclusión de la venta o el intercambio de información personal",
    "privacy_choices_sale_description": "Tiene derecho a optar por no participar en la venta o el intercambio de su información personal. Elegir No vender ni compartir mi información personal detendrá la venta o el intercambio de su información personal en este navegador.",
    "privacy_choices_spi_title": "Limitar el uso de información personal sensible",
    "privacy_choices_spi_description": "Tiene derecho a limitar el uso y la divulgación de su información personal sensible. No usamos ni divulgamos información personal sensible para fines distintos de los permitidos por la ley de California."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'es'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Pilihan Privasi Anda",
    "privacy_choices_intro": "Hukum California memberi Anda hak untuk mengontrol bagaimana kami menggunakan dan membagikan informasi pribadi Anda. Gunakan opsi di bawah untuk menggunakan hak tersebut. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Tolak penjualan atau pembagian informasi pribadi",
    "privacy_choices_sale_description": "Anda berhak menolak penjualan atau pembagian informasi pribadi Anda. Memilih Jangan Jual atau Bagikan Informasi Pribadi Saya akan menghentikan penjualan atau pembagian informasi pribadi Anda untuk browser ini.",
    "privacy_choices_spi_title": "Batasi penggunaan informasi pribadi sensitif",
    "privacy_choices_spi_description": "Anda berhak membatasi penggunaan dan pengungkapan informasi pribadi sensitif Anda. Kami tidak menggunakan atau mengungkapkan informasi pribadi sensitif untuk tujuan selain yang diizinkan oleh hukum California."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'id'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Le tue scelte sulla privacy",
    "privacy_choices_intro": "La legge della California ti dà il diritto di controllare come utilizziamo e condividiamo le tue informazioni personali. Usa le opzioni qui sotto per esercitare tali diritti. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Rinuncia alla vendita o condivisione delle informazioni personali",
    "privacy_choices_sale_description": "Hai il diritto di rinunciare alla vendita o alla condivisione delle tue informazioni personali. Scegliendo Non vendere né condividere le mie informazioni personali interromperemo la vendita o la condivisione delle tue informazioni personali per questo browser.",
    "privacy_choices_spi_title": "Limita l''uso delle informazioni personali sensibili",
    "privacy_choices_spi_description": "Hai il diritto di limitare l''uso e la divulgazione delle tue informazioni personali sensibili. Non utilizziamo né divulghiamo informazioni personali sensibili per scopi diversi da quelli consentiti dalla legge della California."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'it'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "プライバシーに関する選択",
    "privacy_choices_intro": "カリフォルニア州法により、お客様は個人情報の利用および共有方法を管理する権利があります。以下のオプションを使用してこれらの権利を行使してください。{{privacy_policy_link}}",
    "privacy_choices_sale_title": "個人情報の販売または共有のオプトアウト",
    "privacy_choices_sale_description": "お客様には、個人情報の販売または共有を拒否する権利があります。「個人情報の販売・共有を拒否する」を選択すると、このブラウザでの個人情報の販売または共有を停止します。",
    "privacy_choices_spi_title": "機微な個人情報の利用の制限",
    "privacy_choices_spi_description": "お客様には、機微な個人情報の利用および開示を制限する権利があります。当社は、カリフォルニア州法で認められる目的以外で機微な個人情報を利用または開示しません。"
}'::jsonb,
    updated_at = NOW()
WHERE language = 'ja'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "개인정보 선택 사항",
    "privacy_choices_intro": "캘리포니아 법은 저희가 개인정보를 사용하고 공유하는 방식을 통제할 권리를 부여합니다. 아래 옵션을 사용하여 해당 권리를 행사하세요. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "개인정보 판매 또는 공유 거부",
    "privacy_choices_sale_description": "귀하는 개인정보의 판매 또는 공유를 거부할 권리가 있습니다. 내 개인정보 판매 또는 공유 금지를 선택하면 이 브라우저에서 개인정보 판매 또는 공유를 중단합니다.",
    "privacy_choices_spi_title": "민감한 개인정보 사용 제한",
    "privacy_choices_spi_description": "귀하는 민감한 개인정보의 사용 및 공개를 제한할 권리가 있습니다. 저희는 캘리포니아 법이 허용하는 목적 외로는 민감한 개인정보를 사용하거나 공개하지 않습니다."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'ko'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Twoje wybory dotyczące prywatności",
    "privacy_choices_intro": "Prawo Kalifornii daje Ci prawo do kontrolowania, w jaki sposób wykorzystujemy i udostępniamy Twoje dane osobowe. Skorzystaj z poniższych opcji, aby skorzystać z tych praw. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Rezygnacja ze sprzedaży lub udostępniania danych osobowych",
    "privacy_choices_sale_description": "Masz prawo zrezygnować ze sprzedaży lub udostępniania swoich danych osobowych. Wybór Nie sprzedawaj ani nie udostępniaj moich danych osobowych zatrzyma sprzedaż lub udostępnianie Twoich danych osobowych w tej przeglądarce.",
    "privacy_choices_spi_title": "Ograniczenie wykorzystywania wrażliwych danych osobowych",
    "privacy_choices_spi_description": "Masz prawo ograniczyć wykorzystywanie i ujawnianie swoich wrażliwych danych osobowych. Nie wykorzystujemy ani nie ujawniamy wrażliwych danych osobowych w celach innych niż dozwolone przez prawo Kalifornii."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'pl'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Suas escolhas de privacidade",
    "privacy_choices_intro": "A lei da Califórnia dá a você o direito de controlar como usamos e compartilhamos suas informações pessoais. Use as opções abaixo para exercer esses direitos. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Recusar a venda ou o compartilhamento de informações pessoais",
    "privacy_choices_sale_description": "Você tem o direito de recusar a venda ou o compartilhamento de suas informações pessoais. Escolher Não vender nem compartilhar minhas informações pessoais interromperá a venda ou o compartilhamento de suas informações pessoais neste navegador.",
    "privacy_choices_spi_title": "Limitar o uso de informações pessoais sensíveis",
    "privacy_choices_spi_description": "Você tem o direito de limitar o uso e a divulgação de suas informações pessoais sensíveis. Não usamos nem divulgamos informações pessoais sensíveis para fins além dos permitidos pela lei da Califórnia."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'pt'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Gizlilik tercihleriniz",
    "privacy_choices_intro": "Kaliforniya yasası, kişisel bilgilerinizi nasıl kullandığımızı ve paylaştığımızı kontrol etme hakkı tanır. Bu hakları kullanmak için aşağıdaki seçenekleri kullanın. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Kişisel bilgilerin satışından veya paylaşımından vazgeçme",
    "privacy_choices_sale_description": "Kişisel bilgilerinizin satışından veya paylaşımından vazgeçme hakkına sahipsiniz. Kişisel Bilgilerimi Satma veya Paylaşma''yı seçmek, bu tarayıcı için kişisel bilgilerinizin satışını veya paylaşımını durdurur.",
    "privacy_choices_spi_title": "Hassas kişisel bilgilerin kullanımını sınırlama",
    "privacy_choices_spi_description": "Hassas kişisel bilgilerinizin kullanımını ve ifşasını sınırlama hakkına sahipsiniz. Hassas kişisel bilgileri Kaliforniya yasasının izin verdiği amaçlar dışında kullanmayız veya ifşa etmeyiz."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'tr'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "Ваш вибір щодо конфіденційності",
    "privacy_choices_intro": "Закон Каліфорнії дає вам право контролювати, як ми використовуємо та поширюємо ваші персональні дані. Скористайтеся наведеними нижче параметрами, щоб реалізувати ці права. {{privacy_policy_link}}",
    "privacy_choices_sale_title": "Відмова від продажу або поширення персональних даних",
    "privacy_choices_sale_description": "Ви маєте право відмовитися від продажу або поширення ваших персональних даних. Вибір «Не продавати та не ділитися моїми персональними даними» зупинить продаж або поширення ваших персональних даних для цього браузера.",
    "privacy_choices_spi_title": "Обмеження використання конфіденційних персональних даних",
    "privacy_choices_spi_description": "Ви маєте право обмежити використання та розкриття ваших конфіденційних персональних даних. Ми не використовуємо та не розкриваємо конфіденційні персональні дані для цілей, не дозволених законом Каліфорнії."
}'::jsonb,
    updated_at = NOW()
WHERE language = 'uk'
  AND NOT translations ? 'privacy_choices_title';

UPDATE cookie_banner_translations
SET translations = translations || '{
    "privacy_choices_title": "您的隐私选择",
    "privacy_choices_intro": "加州法律赋予您控制我们如何使用和共享您的个人信息的权利。请使用以下选项行使这些权利。{{privacy_policy_link}}",
    "privacy_choices_sale_title": "选择退出出售或共享个人信息",
    "privacy_choices_sale_description": "您有权选择退出出售或共享您的个人信息。选择“不要出售或共享我的个人信息”将停止我们在此浏览器中出售或共享您的个人信息。",
    "privacy_choices_spi_title": "限制敏感个人信息的使用",
    "privacy_choices_spi_description": "您有权限制敏感个人信息的使用和披露。我们不会将敏感个人信息用于加州法律允许范围以外的目的。"
}'::jsonb,
    updated_at = NOW()
WHERE language = 'zh'
  AND NOT translations ? 'privacy_choices_title';
