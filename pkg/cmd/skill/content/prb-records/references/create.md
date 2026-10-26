# Required create flags

`--org` is required on every command in this table except `control`.
`prb <command> create --help` is authoritative when a row is stale.

| Command | Required flags besides `--org` |
| --- | --- |
| `framework create` | `--name` |
| `control create` | `--framework`, `--section-title`, `--name`, `--description` |
| `task create` | `--name`, `--priority` |
| `document create` | `--title`, `--document-type`, `--classification` |
| `asset create` | `--name`, `--asset-type`, `--amount`, `--owner`, `--data-types-stored` |
| `datum create` | `--name`, `--data-classification`, `--owner` |
| `risk create` | `--name`, `--category` |
| `processing-activity create` | `--name`, `--role`, `--lawful-basis`, `--special-or-criminal-data`, `--data-protection-impact-assessment`, `--transfer-impact-assessment` |
| `obligation create` | `--area`, `--source`, `--status`, `--type`, `--owner` |
| `rights-request create` | `--data-subject`, `--type`, `--state` |
| `internal-control create` | `--name`, `--category` |
| `business-function create` | `--name`, `--classification`, `--mtd-minutes`, `--rto-minutes`, `--rpo-minutes` |
| `ai-system create` | `--name`, `--status`, `--risk-classification` |
| `audit create` | `--name`, `--framework` (repeat `--framework` for another framework) |
| `finding create` | `--kind`, `--status`, `--priority` |
| `third-party create` | `--name`, `--category` |
| `webhook create` | `--url`, `--event` (repeat `--event` for another event) |

`control list` takes `--framework` and does not take `--org`. Other lists take
`--org`.

`--amount` is an integer of at least 1. Business-function minute flags are
integers. `--international-transfers` defaults to false and can be omitted.
