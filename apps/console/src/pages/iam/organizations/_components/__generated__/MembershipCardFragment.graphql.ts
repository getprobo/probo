/**
 * @generated SignedSource<<30ba0cd2d4738694d58bf346410d4d68>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipCardFragment$data = {
  readonly activeSession: {
    readonly expiresAt: any;
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
      "name": "activeSession",
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
    }
  ],
  "type": "Membership",
  "abstractKey": null
};
})();

(node as any).hash = "2145f6030b486447df7b7d9ff798ec06";

export default node;
