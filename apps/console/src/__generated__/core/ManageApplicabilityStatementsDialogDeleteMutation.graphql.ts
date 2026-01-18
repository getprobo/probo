/**
 * @generated SignedSource<<29841338b2cd037d804ec2c73469aa77>>
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
export type ManageApplicabilityStatementsDialogDeleteMutation$variables = {
  input: DeleteApplicabilityStatementInput;
};
export type ManageApplicabilityStatementsDialogDeleteMutation$data = {
  readonly deleteApplicabilityStatement: {
    readonly deletedApplicabilityStatementId: string;
  };
};
export type ManageApplicabilityStatementsDialogDeleteMutation = {
  response: ManageApplicabilityStatementsDialogDeleteMutation$data;
  variables: ManageApplicabilityStatementsDialogDeleteMutation$variables;
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
    "name": "ManageApplicabilityStatementsDialogDeleteMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ManageApplicabilityStatementsDialogDeleteMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "3812268dbfc77c8db556c68dc91fb36f",
    "id": null,
    "metadata": {},
    "name": "ManageApplicabilityStatementsDialogDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ManageApplicabilityStatementsDialogDeleteMutation(\n  $input: DeleteApplicabilityStatementInput!\n) {\n  deleteApplicabilityStatement(input: $input) {\n    deletedApplicabilityStatementId\n  }\n}\n"
  }
};
})();

(node as any).hash = "7b17936a836169398063bc57593cab38";

export default node;
