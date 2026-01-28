/**
 * @generated SignedSource<<b40dce1506005cda2d2be9bd6aeee6dd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageFileListItem_compliancePageFragment$data = {
  readonly canUpdate: boolean;
  readonly " $fragmentType": "CompliancePageFileListItem_compliancePageFragment";
};
export type CompliancePageFileListItem_compliancePageFragment$key = {
  readonly " $data"?: CompliancePageFileListItem_compliancePageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageFileListItem_compliancePageFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageFileListItem_compliancePageFragment",
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

(node as any).hash = "83a20628bcc627902e29c5515741efc5";

export default node;
