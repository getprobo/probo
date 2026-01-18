/**
 * @generated SignedSource<<8df9edfe29c3c7432ed3420af7745be2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteApplicabilityStatementInput = {
  applicabilityStatementId: string;
};
export type FrameworkControlPageDetachStateOfApplicabilityControlMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteApplicabilityStatementInput;
};
export type FrameworkControlPageDetachStateOfApplicabilityControlMutation$data = {
  readonly deleteApplicabilityStatement: {
    readonly deletedApplicabilityStatementId: string;
  };
};
export type FrameworkControlPageDetachStateOfApplicabilityControlMutation = {
  response: FrameworkControlPageDetachStateOfApplicabilityControlMutation$data;
  variables: FrameworkControlPageDetachStateOfApplicabilityControlMutation$variables;
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
  "name": "deletedApplicabilityStatementId",
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
    "name": "FrameworkControlPageDetachStateOfApplicabilityControlMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "deleteApplicabilityStatement",
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
    "name": "FrameworkControlPageDetachStateOfApplicabilityControlMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "deleteApplicabilityStatement",
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
            "name": "deletedApplicabilityStatementId",
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
    "cacheID": "07debb27d1920dac0618dfec12fe673f",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDetachStateOfApplicabilityControlMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDetachStateOfApplicabilityControlMutation(\n  $input: DeleteApplicabilityStatementInput!\n) {\n  deleteApplicabilityStatement(input: $input) {\n    deletedApplicabilityStatementId\n  }\n}\n"
  }
};
})();

(node as any).hash = "2065e25abee426d8da7b6b19617265dd";

export default node;
