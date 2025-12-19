/**
 * @generated SignedSource<<d4b9efa7b9092ceb3f995507108a68e3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationDropdown_viewerFragment$data = {
  readonly pendingInvitations: {
    readonly totalCount: number;
  };
  readonly " $fragmentType": "OrganizationDropdown_viewerFragment";
};
export type OrganizationDropdown_viewerFragment$key = {
  readonly " $data"?: OrganizationDropdown_viewerFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"OrganizationDropdown_viewerFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "OrganizationDropdown_viewerFragment",
  "selections": [
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "InvitationConnection",
        "kind": "LinkedField",
        "name": "pendingInvitations",
        "plural": false,
        "selections": [
          {
            "kind": "RequiredField",
            "field": {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "totalCount",
              "storageKey": null
            },
            "action": "THROW"
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    }
  ],
  "type": "Identity",
  "abstractKey": null
};

(node as any).hash = "ce947f812c95b74e9f250b8d5821ebb0";

export default node;
