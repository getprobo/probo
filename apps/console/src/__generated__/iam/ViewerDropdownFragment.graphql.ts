/**
 * @generated SignedSource<<e7945358237630130f305751e60ad92a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ViewerDropdownFragment$data = {
  readonly canListAPIKeys: boolean;
  readonly email: string;
  readonly fullName: string;
  readonly " $fragmentType": "ViewerDropdownFragment";
};
export type ViewerDropdownFragment$key = {
  readonly " $data"?: ViewerDropdownFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"ViewerDropdownFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "ViewerDropdownFragment",
  "selections": [
    {
      "alias": "canListAPIKeys",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:personal-api-key:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:personal-api-key:list\")"
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "email",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fullName",
      "storageKey": null
    }
  ],
  "type": "Identity",
  "abstractKey": null
};

(node as any).hash = "e96fe96007f4d949403890fda7180cba";

export default node;
