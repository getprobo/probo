/**
 * @generated SignedSource<<e42d456a9115e64c71c9ee6a89dcf246>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type TaskState = "DONE" | "TODO";
import { FragmentRefs } from "relay-runtime";
export type TaskFormDialogFragment$data = {
  readonly assignedTo: {
    readonly id: string;
  } | null | undefined;
  readonly deadline: string | null | undefined;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly measure: {
    readonly id: string;
  } | null | undefined;
  readonly name: string;
  readonly state: TaskState;
  readonly timeEstimate: string | null | undefined;
  readonly " $fragmentType": "TaskFormDialogFragment";
};
export type TaskFormDialogFragment$key = {
  readonly " $data"?: TaskFormDialogFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"TaskFormDialogFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = [
  (v0/*: any*/)
];
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "TaskFormDialogFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "description",
      "storageKey": null
    },
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
      "name": "state",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "timeEstimate",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "deadline",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Profile",
      "kind": "LinkedField",
      "name": "assignedTo",
      "plural": false,
      "selections": (v1/*: any*/),
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Measure",
      "kind": "LinkedField",
      "name": "measure",
      "plural": false,
      "selections": (v1/*: any*/),
      "storageKey": null
    }
  ],
  "type": "Task",
  "abstractKey": null
};
})();

(node as any).hash = "3a4bede2199df797a20a6d87358d41cf";

export default node;
