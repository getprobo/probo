/**
 * @generated SignedSource<<b449429f81b74e6da6c7c6ed523ca0ab>>
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
    readonly excludedUserNames: ReadonlyArray<string>;
    readonly id: string;
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
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "excludedUserNames",
          "storageKey": null
        },
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

(node as any).hash = "4f3898fca8d93369a46b50227f3df63d";

export default node;
