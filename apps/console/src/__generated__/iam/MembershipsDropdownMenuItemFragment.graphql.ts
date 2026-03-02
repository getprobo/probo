/**
 * @generated SignedSource<<fe0de6febc77244c92f8d2413bbb0ca2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipsDropdownMenuItemFragment$data = {
  readonly id: string;
  readonly lastSession: {
    readonly expiresAt: string;
    readonly id: string;
  } | null | undefined;
  readonly " $fragmentType": "MembershipsDropdownMenuItemFragment";
};
export type MembershipsDropdownMenuItemFragment$key = {
  readonly " $data"?: MembershipsDropdownMenuItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdownMenuItemFragment">;
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
  "name": "MembershipsDropdownMenuItemFragment",
  "selections": [
    (v0/*: any*/),
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
    }
  ],
  "type": "Membership",
  "abstractKey": null
};
})();

(node as any).hash = "4f40e957597476f8e6179f6df68d386e";

export default node;
