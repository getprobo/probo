/**
 * @generated SignedSource<<24c6b253b4339703abe0106728cb4a6f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageDocumentListItem_compliancePageFragment$data = {
  readonly canUpdate: boolean;
  readonly " $fragmentType": "CompliancePageDocumentListItem_compliancePageFragment";
};
export type CompliancePageDocumentListItem_compliancePageFragment$key = {
  readonly " $data"?: CompliancePageDocumentListItem_compliancePageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDocumentListItem_compliancePageFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageDocumentListItem_compliancePageFragment",
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

(node as any).hash = "7c884fb02ab6613707d02aa44c547ac7";

export default node;
