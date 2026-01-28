/**
 * @generated SignedSource<<0bf1290a489f4685a6057509b9eaed9c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageAuditListFragment$data = {
  readonly audits: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_auditFragment">;
      };
    }>;
  };
  readonly compliancePage: {
    readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_compliancePageFragment">;
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
          "value": 1000
        }
      ],
      "concreteType": "AuditConnection",
      "kind": "LinkedField",
      "name": "audits",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "AuditEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Audit",
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
                  "name": "CompliancePageAuditListItem_auditFragment"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "audits(first:1000)"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "d307c5c23d18357bc09c1115989dd734";

export default node;
