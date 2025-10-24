/**
 * @generated SignedSource<<f788919caaf846a6a1bdc81f19841823>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteTrustCenterFileInput = {
  id: string;
};
export type TrustCenterFileGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteTrustCenterFileInput;
};
export type TrustCenterFileGraphDeleteMutation$data = {
  readonly deleteTrustCenterFile: {
    readonly deletedTrustCenterFileId: string;
  };
};
export type TrustCenterFileGraphDeleteMutation = {
  response: TrustCenterFileGraphDeleteMutation$data;
  variables: TrustCenterFileGraphDeleteMutation$variables;
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
  "name": "deletedTrustCenterFileId",
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
    "name": "TrustCenterFileGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTrustCenterFilePayload",
        "kind": "LinkedField",
        "name": "deleteTrustCenterFile",
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
    "name": "TrustCenterFileGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTrustCenterFilePayload",
        "kind": "LinkedField",
        "name": "deleteTrustCenterFile",
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
            "name": "deletedTrustCenterFileId",
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
    "cacheID": "ad78a064da9b58291653f8283012ffc6",
    "id": null,
    "metadata": {},
    "name": "TrustCenterFileGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterFileGraphDeleteMutation(\n  $input: DeleteTrustCenterFileInput!\n) {\n  deleteTrustCenterFile(input: $input) {\n    deletedTrustCenterFileId\n  }\n}\n"
  }
};
})();

(node as any).hash = "26c1ce69e74f3a4a39d060b38afbe3e0";

export default node;
