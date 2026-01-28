/**
 * @generated SignedSource<<883f3a053aa51e1c94ad7c75a2fb91d2>>
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
export type DeleteCompliancePageFileDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteTrustCenterFileInput;
};
export type DeleteCompliancePageFileDialogMutation$data = {
  readonly deleteTrustCenterFile: {
    readonly deletedTrustCenterFileId: string;
  };
};
export type DeleteCompliancePageFileDialogMutation = {
  response: DeleteCompliancePageFileDialogMutation$data;
  variables: DeleteCompliancePageFileDialogMutation$variables;
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
    "name": "DeleteCompliancePageFileDialogMutation",
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
    "name": "DeleteCompliancePageFileDialogMutation",
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
    "cacheID": "d3d7dcdb4201d8f3eeaa66973545bed0",
    "id": null,
    "metadata": {},
    "name": "DeleteCompliancePageFileDialogMutation",
    "operationKind": "mutation",
    "text": "mutation DeleteCompliancePageFileDialogMutation(\n  $input: DeleteTrustCenterFileInput!\n) {\n  deleteTrustCenterFile(input: $input) {\n    deletedTrustCenterFileId\n  }\n}\n"
  }
};
})();

(node as any).hash = "1bf5628f93e3ec6ee240290f5aaf5761";

export default node;
