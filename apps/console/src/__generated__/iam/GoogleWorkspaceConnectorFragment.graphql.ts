/**
 * @generated SignedSource<<3c7c629a3509891b6edc98178a376c2d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type GoogleWorkspaceConnectorFragment$data = {
  readonly bridge: {
    readonly connector: {
      readonly createdAt: string;
      readonly id: string;
    } | null | undefined;
  } | null | undefined;
  readonly id: string;
  readonly " $fragmentType": "GoogleWorkspaceConnectorFragment";
};
export type GoogleWorkspaceConnectorFragment$key = {
  readonly " $data"?: GoogleWorkspaceConnectorFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"GoogleWorkspaceConnectorFragment">;
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
  "name": "GoogleWorkspaceConnectorFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "concreteType": "SCIMBridge",
      "kind": "LinkedField",
      "name": "bridge",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "Connector",
          "kind": "LinkedField",
          "name": "connector",
          "plural": false,
          "selections": [
            (v0/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "createdAt",
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "SCIMConfiguration",
  "abstractKey": null
};
})();

(node as any).hash = "6ce8b7f55ff2a3d0fad04ad6b452a28a";

export default node;
