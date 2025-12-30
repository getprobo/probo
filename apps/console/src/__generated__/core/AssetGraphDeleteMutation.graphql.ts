/**
 * @generated SignedSource<<067da1978a06c9feadeeff5ae6af7b50>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteAssetInput = {
  assetId: string;
};
export type AssetGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteAssetInput;
};
export type AssetGraphDeleteMutation$data = {
  readonly deleteAsset: {
    readonly deletedAssetId: string;
  };
};
export type AssetGraphDeleteMutation = {
  response: AssetGraphDeleteMutation$data;
  variables: AssetGraphDeleteMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedAssetId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "AssetGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteAssetPayload",
        "kind": "LinkedField",
        "name": "deleteAsset",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "AssetGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteAssetPayload",
        "kind": "LinkedField",
        "name": "deleteAsset",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedAssetId",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "bbda9e3ed8effba4140c130883a59d07",
    "id": null,
    "metadata": {},
    "name": "AssetGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation AssetGraphDeleteMutation(\n  $input: DeleteAssetInput!\n) {\n  deleteAsset(input: $input) {\n    deletedAssetId\n  }\n}\n"
  }
};
})();

(node as any).hash = "46b040e435edef42987b5524d823f592";

export default node;
