/**
 * @generated SignedSource<<05d197ce59adad67cfb89dff4c502c58>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteFrameworkInput = {
  frameworkId: string;
};
export type FrameworkGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteFrameworkInput;
};
export type FrameworkGraphDeleteMutation$data = {
  readonly deleteFramework: {
    readonly deletedFrameworkId: string;
  };
};
export type FrameworkGraphDeleteMutation = {
  response: FrameworkGraphDeleteMutation$data;
  variables: FrameworkGraphDeleteMutation$variables;
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
  "name": "deletedFrameworkId",
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
    "name": "FrameworkGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteFrameworkPayload",
        "kind": "LinkedField",
        "name": "deleteFramework",
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
    "name": "FrameworkGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteFrameworkPayload",
        "kind": "LinkedField",
        "name": "deleteFramework",
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
            "name": "deletedFrameworkId",
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
    "cacheID": "761a391f497aa0e2e03d0db958089712",
    "id": null,
    "metadata": {},
    "name": "FrameworkGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkGraphDeleteMutation(\n  $input: DeleteFrameworkInput!\n) {\n  deleteFramework(input: $input) {\n    deletedFrameworkId\n  }\n}\n"
  }
};
})();

(node as any).hash = "6ebefeb003ce3a27bd134651e1b5c94b";

export default node;
