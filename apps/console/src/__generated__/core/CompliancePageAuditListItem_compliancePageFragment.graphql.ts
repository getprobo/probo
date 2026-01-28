/**
 * @generated SignedSource<<9a44b4f71be46d5e02c3f1c8587d6f32>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageAuditListItem_compliancePageFragment$data = {
  readonly canUpdate: boolean;
  readonly " $fragmentType": "CompliancePageAuditListItem_compliancePageFragment";
};
export type CompliancePageAuditListItem_compliancePageFragment$key = {
  readonly " $data"?: CompliancePageAuditListItem_compliancePageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_compliancePageFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageAuditListItem_compliancePageFragment",
  "selections": [
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:update\")"
    }
  ],
  "type": "TrustCenter",
  "abstractKey": null
};

(node as any).hash = "4744b831ee9ad55304d2159e8b2ebfd5";

export default node;
