/**
 * @generated SignedSource<<b655b8a8c70af9ceb259e2fed933f397>>
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
export type AddApplicabilityStatementDialogDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteApplicabilityStatementInput;
};
export type AddApplicabilityStatementDialogDeleteMutation$data = {
  readonly deleteApplicabilityStatement: {
    readonly deletedApplicabilityStatementId: string;
  };
};
export type AddApplicabilityStatementDialogDeleteMutation = {
  response: AddApplicabilityStatementDialogDeleteMutation$data;
  variables: AddApplicabilityStatementDialogDeleteMutation$variables;
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
    "name": "AddApplicabilityStatementDialogDeleteMutation",
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
    "name": "AddApplicabilityStatementDialogDeleteMutation",
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
    "cacheID": "33d1b54d91487d7c3fcbc8ff8bcb68ce",
    "id": null,
    "metadata": {},
    "name": "AddApplicabilityStatementDialogDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation AddApplicabilityStatementDialogDeleteMutation(\n  $input: DeleteApplicabilityStatementInput!\n) {\n  deleteApplicabilityStatement(input: $input) {\n    deletedApplicabilityStatementId\n  }\n}\n"
  }
};
})();

(node as any).hash = "965c5b50e3a2a077fb7dc287b70026f1";

export default node;
