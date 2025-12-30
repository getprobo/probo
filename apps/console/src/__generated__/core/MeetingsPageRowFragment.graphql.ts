/**
 * @generated SignedSource<<7efb971ec8052cecbd96f86ea6d3af4c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MeetingsPageRowFragment$data = {
  readonly attendees: ReadonlyArray<{
    readonly fullName: string;
    readonly id: string;
  }>;
  readonly date: any;
  readonly id: string;
  readonly name: string;
  readonly " $fragmentType": "MeetingsPageRowFragment";
};
export type MeetingsPageRowFragment$key = {
  readonly " $data"?: MeetingsPageRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MeetingsPageRowFragment">;
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
  "name": "MeetingsPageRowFragment",
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
      "name": "date",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "attendees",
      "plural": true,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fullName",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Meeting",
  "abstractKey": null
};
})();

(node as any).hash = "91ba43abb559bed5acdda60b8f61bda9";

export default node;
