/**
 * @generated SignedSource<<bd4c4d5ded5e4c1ed8ce6d4fe2fbb15c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageAuditListFragment$data = {
  readonly compliancePage: {
    readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_compliancePageFragment">;
  };
  readonly reports: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_reportFragment">;
      };
    }>;
  };
  readonly " $fragmentType": "CompliancePageAuditListFragment";
};
export type CompliancePageAuditListFragment$key = {
  readonly " $data"?: CompliancePageAuditListFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageAuditListFragment",
  "selections": [
    {
      "kind": "RequiredField",
      "field": {
        "alias": "compliancePage",
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "trustCenter",
        "plural": false,
        "selections": [
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageAuditListItem_compliancePageFragment"
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 100
        }
      ],
      "concreteType": "ReportConnection",
      "kind": "LinkedField",
      "name": "reports",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ReportEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Report",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "id",
                  "storageKey": null
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "CompliancePageAuditListItem_reportFragment"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "reports(first:100)"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "68ab2f8338ef800af42336443cd89c80";

export default node;
