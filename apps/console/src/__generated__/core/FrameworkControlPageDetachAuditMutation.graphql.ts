/**
 * @generated SignedSource<<f3f6c7f7f6e09f53b111e39ce4176b1c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlAuditMappingInput = {
  auditId: string;
  controlId: string;
};
export type FrameworkControlPageDetachAuditMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlAuditMappingInput;
};
export type FrameworkControlPageDetachAuditMutation$data = {
  readonly deleteControlAuditMapping: {
    readonly deletedAuditId: string;
  };
};
export type FrameworkControlPageDetachAuditMutation = {
  response: FrameworkControlPageDetachAuditMutation$data;
  variables: FrameworkControlPageDetachAuditMutation$variables;
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
  "name": "deletedAuditId",
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
    "name": "FrameworkControlPageDetachAuditMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlAuditMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlAuditMapping",
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
    "name": "FrameworkControlPageDetachAuditMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlAuditMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlAuditMapping",
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
            "name": "deletedAuditId",
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
    "cacheID": "0ec6a285e6ffd0057602e30ef8411e16",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDetachAuditMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDetachAuditMutation(\n  $input: DeleteControlAuditMappingInput!\n) {\n  deleteControlAuditMapping(input: $input) {\n    deletedAuditId\n  }\n}\n"
  }
};
})();

(node as any).hash = "3c800764c7d5801bd228e8a61625bd75";

export default node;
