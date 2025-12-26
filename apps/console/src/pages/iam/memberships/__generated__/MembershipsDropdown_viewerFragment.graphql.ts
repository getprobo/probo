/**
 * @generated SignedSource<<842fffc896db6311ff3c5901eae70176>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipsDropdown_viewerFragment$data = {
  readonly pendingInvitations: {
    readonly totalCount: number;
  };
  readonly " $fragmentType": "MembershipsDropdown_viewerFragment";
};
export type MembershipsDropdown_viewerFragment$key = {
  readonly " $data"?: MembershipsDropdown_viewerFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdown_viewerFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MembershipsDropdown_viewerFragment",
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

(node as any).hash = "96b64fbc2481273728051e42d74bc8b5";

export default node;
