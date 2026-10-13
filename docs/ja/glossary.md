<!--
Copyright (c) 2026 MICHISUU Inc.
SPDX-License-Identifier: MIT
-->

# 日本語化 用語集・表記ルール

`ja-JP.json` の翻訳とレビューで使う訳語と表記のルールです。ISMS・Pマークの実務で通じる言葉を優先します。JIS Q 27001 / JIS Q 15001 などの規格本文は引用しません。

## 表記ルール

| 項目 | ルール | 例 |
| --- | --- | --- |
| 文体 | 説明文・メッセージは「です・ます」調 | 「ドキュメントを公開しました。」 |
| ボタン・メニュー | 体言止め、または動詞の終止形を短く | 「保存」「削除」「リスクを追加」 |
| 見出し・ラベル | 体言止め | 「リスク一覧」「担当者」 |
| プレースホルダー | 「〜を入力」「〜を選択」 | 「名前を入力」 |
| 確認ダイアログ | 「〜してもよろしいですか？」 | 「このリスクを削除してもよろしいですか？」 |
| 成功メッセージ | 「〜しました」 | 「保存しました」 |
| 失敗メッセージ | 「〜できませんでした」 | 「ドキュメントを削除できませんでした」 |
| 進行中 | 「〜中…」 | 「読み込み中…」 |
| 空状態 | 「〜はありません」 | 「リスクはありません」 |
| 句読点 | 「、」「。」を使う。ボタン・ラベル・見出しには句点を付けない | |
| 疑問符・感嘆符 | 全角「？」「！」 | |
| 括弧 | 全角「（）」。ただし英数字だけを囲むときは半角 `()` でもよい | 「適用宣言書（SoA）」 |
| 英数字 | 半角。英数字と日本語の間にスペースを入れない | 「SAML SSOを設定」「{{count}}件のリスク」 |
| 長音符 | 語末の長音を付ける（JIS Z 8301 慣例に合わせない） | 「ユーザー」「サーバー」「フォルダー」「プロバイダー」 |
| 省略記号 | 原文の `…` はそのまま `…` | 「検索…」 |
| 複数形 | 日本語は `_other` のみ使われる。`_one` などのキーも残し、同じ訳を入れる | |

## 翻訳時に保持するもの（変更禁止）

- 補間変数 `{{name}}`、`{{count}}` など。順序は日本語の語順に合わせて入れ替えてよい
- ネスト参照 `$t(...)`
- `<0>…</0>`、`<1/>`、`<strong>…</strong>` などのタグと番号
- 改行 `\n`、URL、メールアドレス、コード（`curl ...` 等）、HTTP ヘッダー名
- 製品名・サービス名・規格名：Probo、Slack、Google Workspace、Microsoft 365、Entra ID、Okta、GitHub、Linear、PostHog、AWS、n8n、SAML、SCIM、OIDC、OAuth、SSO、API、CSV、PDF、JSON、URL、DNS、TLS、GDPR、HIPAA、CCPA、SOC 2、ISO 27001、ISO 42001、TCF、IAB
- 国名は日本で一般的な名称（外務省の表記に準じる）

## 略語の扱い

GRC 分野の略語（TPRM、CMP、DPIA、TIA、SoA、MTD、RTO、RPO、GPAI など）は、専門家以外には伝わりにくいため日本語にします。表示する場所によって書き方を変えます。

| 場所 | 書き方 | 例 |
| --- | --- | --- |
| ナビ・ページ見出し・独立した見出し | 日本語（略称） | 「データ保護影響評価（DPIA）」 |
| ボタン・メッセージ・説明文・空状態 | 日本語のみ | 「データ保護影響評価を作成」 |
| 表の列見出し | 日本語のみ | 「目標復旧時間」 |

IT 管理者が設定画面で使う技術規格の名前（SSO、SAML、SCIM、IdP、OAuth、TCF、IAB など）は、そのまま残します。設定項目名やベンダーのドキュメントと照合できるようにするためです。

| 英語 | 日本語 |
| --- | --- |
| TPRM | 取引先リスク管理 |
| CMP | Cookie同意管理（TCF の「CMP ID」はそのまま） |
| IAM | ID・アクセス管理 |
| MFA | 多要素認証 |
| MTD / RTO / RPO | 最大許容停止時間 / 目標復旧時間 / 目標復旧時点 |
| GPAI | 汎用AI |
| NDA | 秘密保持契約 |
| GVL | ベンダーリスト（GVL） |
| GPC | GPC（拒否シグナル） |

## 用語集

### ガバナンス・ISMS 全般

| 英語 | 日本語 | 備考 |
| --- | --- | --- |
| Governance | ガバナンス | |
| Compliance | コンプライアンス | |
| Framework | フレームワーク | ISO 27001 などの規格・基準 |
| Control (framework control) | 管理策 | 規格側の要求項目 |
| Internal control | 内部統制 | 組織が実施する統制 |
| Measure | 対策 | |
| Measure implementation | 対策の実施状況 | |
| Statement of Applicability (SoA) | 適用宣言書 | 略称は使わない |
| Applicability statement | 適用可否 | |
| Applicable / Not applicable | 適用 / 適用除外 | 状態の「Not Applicable」は「対象外」 |
| Justification | 根拠 | 適用除外の理由など |
| Obligation | 義務 | 法令・契約上の義務 |
| Compliant / Partially compliant / Non-compliant | 遵守 / 一部遵守 / 未遵守 | 義務の状態。監査の「不適合」と区別する |
| Policy | 方針 | 文書種別。「ポリシー」は不可 |
| Procedure | 手順 | |
| Plan | 計画 | |
| Register | 台帳 | |
| Record | 記録 | |
| Report | 報告書 | 画面上の「レポート」機能はレポート |
| Template | テンプレート | |
| Context | 組織の状況 | ISMS の 4 章に相当するページ |
| Business function | 業務機能 | |
| Stakeholder | 利害関係者 | |
| Maturity level | 成熟度 | |
| Owner | オーナー | リスクの場合は「リスクオーナー」 |
| Approver | 承認者 | |
| Approval | 承認 | |
| Signature / Sign | 署名 / 署名する | |
| Publish / Published | 公開 / 公開済み | |
| Draft | 下書き | |
| Version | 版 | 「バージョン」は製品・ソフトウェアの版に限る |
| Classification | 機密区分 | |
| Public / Internal / Confidential / Secret | 公開 / 社内限 / 機密 / 極秘 | 文書の機密区分 |

### リスク

| 英語 | 日本語 | 備考 |
| --- | --- | --- |
| Risk | リスク | |
| Risk assessment | リスクアセスメント | |
| Risk analysis | リスク分析 | |
| Risk owner | リスクオーナー | |
| Risk treatment | リスク対応 | 「リスク処理」は不可 |
| Treatment plan | リスク対応計画 | |
| Mitigate / Accept / Avoid / Transfer | 低減 / 受容 / 回避 / 移転 | 対応方針 |
| Inherent risk | 固有リスク | |
| Residual risk | 残留リスク | |
| Likelihood | 発生可能性 | |
| Impact | 影響度 | |
| Severity | 深刻度 | |
| Threat | 脅威 | |
| Vulnerability | 脆弱性 | |
| Scenario | シナリオ | |
| Boundary | 境界 | リスク分析の図 |
| Asset | 資産 | 情報資産 |
| Datum / Data | データ | |

### 監査・是正

| 英語 | 日本語 | 備考 |
| --- | --- | --- |
| Audit | 監査 | |
| Auditor | 監査人 | |
| To book (audit state) | 日程調整待ち | |
| Finding | 指摘事項 | |
| Nonconformity | 不適合 | |
| Major / Minor nonconformity | 重大な不適合 / 軽微な不適合 | |
| Observation | 観察事項 | |
| Opportunity for improvement | 改善の機会 | |
| Corrective action | 是正処置 | 「是正措置」は不可 |
| Root cause | 根本原因 | |
| Evidence | 証跡 | 「エビデンス」は不可 |
| Certificate | 認証書 | 認証機関の発行物。TLS の場合は「証明書」 |
| Compliance report | 準拠報告書 | SOC 2 報告書など |
| Access review | アクセスレビュー | |
| Campaign | キャンペーン | アクセスレビューの実施単位 |
| Entry | エントリー | |
| Task | タスク | |
| Due date | 期限 | |
| Assignee | 担当者 | |

### 個人情報・プライバシー

| 英語 | 日本語 | 備考 |
| --- | --- | --- |
| Privacy | プライバシー | |
| Personal data | 個人情報 | |
| Processing activity | 処理活動 | |
| Record of processing activities | 処理活動の記録 | |
| Data subject | 本人 | 「データ主体」は不可 |
| Rights request | 本人からの請求 | 開示等の請求 |
| Data Protection Impact Assessment (DPIA) | データ保護影響評価 | 略称は使わない |
| Transfer Impact Assessment (TIA) | 移転影響評価 | 略称は使わない |
| Lawful basis | 法的根拠 | |
| Retention period | 保存期間 | |
| Consent | 同意 | |
| Controller / Processor | 管理者 / 処理者 | |
| Subprocessor | サブプロセッサ | compliance-portal の既訳に合わせる |
| Data Privacy Agreement (DPA) | データ処理契約（DPA） | |
| Business Associate Agreement (BAA) | 事業提携契約（BAA） | HIPAA の契約 |
| Cookie banner | Cookieバナー | |
| Tracker | トラッカー | |
| Visitor | 訪問者 | |

### 取引先管理

| 英語 | 日本語 | 備考 |
| --- | --- | --- |
| Third party | 取引先 | 委託先を含む外部の組織 |
| Third-party risk management (TPRM) | 取引先リスク管理 | 略称は使わない |
| Vendor | ベンダー | |
| Service | サービス | |
| Supply chain | サプライチェーン | |
| Assurance | 保証 | |
| Agreement | 契約 | |
| Vetting | 審査 | |
| Contact | 担当者連絡先 | |

### IT 資産・アカウント

| 英語 | 日本語 | 備考 |
| --- | --- | --- |
| IT Asset Management (ITAM) | IT資産管理 | |
| Device | 端末 | |
| Enroll / Enrollment | 登録 | 端末登録 |
| Agent | エージェント | |
| AI system | AIシステム | |
| Provider / Deployer (AI) | 提供者 / 導入者 | EU AI 法の役割 |
| Posture | セキュリティ状態 | 端末のセキュリティ状態 |
| Contractor | 業務委託 | 人の区分。「契約社員」は不可 |
| User | ユーザー | |
| Member / Membership | メンバー / メンバーシップ | |
| Organization | 組織 | |
| Workspace | ワークスペース | |
| Role | ロール | |
| Invitation | 招待 | |
| Sign in / Sign out / Sign up | ログイン / ログアウト / 新規登録 | compliance-portal の既訳「サインアウト」と揃える場合は別途統一 |
| Password | パスワード | |
| Magic link | マジックリンク | |
| Provisioning | プロビジョニング | |
| Connector / Integration | コネクター / 連携 | |
| Link / Unlink (records) | 紐付け / 紐付けを解除 | レコード同士の関連付け。URL の意味のときは「リンク」 |
| Trail (consent records) | 同意履歴 | Cookieバナーの同意記録 |
| Source | ソース | |
| Webhook | Webhook | |
| Token | トークン | |
| Scope | スコープ | |
| Secret | シークレット | 認証情報の意味。機密区分の Secret は「極秘」 |
| Audit log | 監査ログ | |
| Settings | 設定 | |
| Compliance portal | コンプライアンスポータル | |
| Employee portal | 従業員ポータル | |

### 状態

| 英語 | 日本語 |
| --- | --- |
| Not started | 未着手 |
| In progress | 対応中 |
| Implemented | 実施済み |
| Not implemented | 未実施 |
| Not applicable | 対象外 |
| Unknown | 不明 |
| Pending | 保留中 |
| Pending approval | 承認待ち |
| Approved / Rejected | 承認済み / 却下 |
| Active / Inactive | 有効 / 無効 |
| Open / Closed | 未完了 / 完了 |
| Done | 完了 |
| Connected / Disconnected | 接続済み / 未接続 |
| Low / Medium / High / Critical | 低 / 中 / 高 / 重大 |

## 出典とライセンス

この用語集は本リポジトリの MIT ライセンスで公開します。訳語は実務の慣用に基づいて独自に選定したもので、規格本文からの転載ではありません。
