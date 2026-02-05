/**
 * @generated SignedSource<<81e8e025976a7d356fa1da47465d8ad0>>
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
  readonly canDelete: boolean;
  readonly date: string;
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
      "concreteType": "Profile",
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
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:meeting:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:meeting:delete\")"
    }
  ],
  "type": "Meeting",
  "abstractKey": null
};
})();

(node as any).hash = "22a8c7132716408a17d39ea5ff36a38d";

export default node;
