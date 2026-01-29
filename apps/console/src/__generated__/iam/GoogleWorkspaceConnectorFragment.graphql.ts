/**
 * @generated SignedSource<<2beb895d75d5cbcbd64a59387e35c281>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ConnectorProvider = "GOOGLE_WORKSPACE" | "SLACK";
import { FragmentRefs } from "relay-runtime";
export type GoogleWorkspaceConnectorFragment$data = {
  readonly bridge: {
    readonly connector: {
      readonly createdAt: string;
      readonly id: string;
      readonly provider: ConnectorProvider;
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
              "name": "provider",
              "storageKey": null
            },
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

(node as any).hash = "180a97ae989c22060d815280ec2f7b67";

export default node;
