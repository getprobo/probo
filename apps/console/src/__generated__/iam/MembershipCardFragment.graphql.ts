/**
 * @generated SignedSource<<b545f1a231e042bd94c84430f3015112>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipCardFragment$data = {
  readonly lastSession: {
    readonly expiresAt: string;
    readonly id: string;
  } | null | undefined;
  readonly organization: {
    readonly id: string;
    readonly logoUrl: string | null | undefined;
    readonly name: string;
  };
  readonly " $fragmentType": "MembershipCardFragment";
};
export type MembershipCardFragment$key = {
  readonly " $data"?: MembershipCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MembershipCardFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MembershipCardFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Session",
      "kind": "LinkedField",
      "name": "lastSession",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "expiresAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "Organization",
        "kind": "LinkedField",
        "name": "organization",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "logoUrl",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    }
  ],
  "type": "Membership",
  "abstractKey": null
};
})();

(node as any).hash = "5a2532aed755a94df49c2f9ce0929a1f";

export default node;
