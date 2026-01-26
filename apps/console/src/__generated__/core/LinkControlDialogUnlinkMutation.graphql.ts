/**
 * @generated SignedSource<<c129078380fd3f65393dc60cb30e12f5>>
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
export type LinkControlDialogUnlinkMutation$variables = {
  input: DeleteApplicabilityStatementInput;
};
export type LinkControlDialogUnlinkMutation$data = {
  readonly deleteApplicabilityStatement: {
    readonly deletedApplicabilityStatementId: string;
  };
};
export type LinkControlDialogUnlinkMutation = {
  response: LinkControlDialogUnlinkMutation$data;
  variables: LinkControlDialogUnlinkMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "DeleteApplicabilityStatementPayload",
    "kind": "LinkedField",
    "name": "deleteApplicabilityStatement",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedApplicabilityStatementId",
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "LinkControlDialogUnlinkMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "LinkControlDialogUnlinkMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "560cb4342a2a277881705e91b5b6a17c",
    "id": null,
    "metadata": {},
    "name": "LinkControlDialogUnlinkMutation",
    "operationKind": "mutation",
    "text": "mutation LinkControlDialogUnlinkMutation(\n  $input: DeleteApplicabilityStatementInput!\n) {\n  deleteApplicabilityStatement(input: $input) {\n    deletedApplicabilityStatementId\n  }\n}\n"
  }
};
})();

(node as any).hash = "80dd4efb13f7b677ad62792150ca70d9";

export default node;
