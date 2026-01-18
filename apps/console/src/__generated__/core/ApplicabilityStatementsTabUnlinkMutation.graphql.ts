/**
 * @generated SignedSource<<69bd0bf192ea8a01f2eaf33aac066529>>
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
export type ApplicabilityStatementsTabUnlinkMutation$variables = {
  input: DeleteApplicabilityStatementInput;
};
export type ApplicabilityStatementsTabUnlinkMutation$data = {
  readonly deleteApplicabilityStatement: {
    readonly deletedApplicabilityStatementId: string;
  };
};
export type ApplicabilityStatementsTabUnlinkMutation = {
  response: ApplicabilityStatementsTabUnlinkMutation$data;
  variables: ApplicabilityStatementsTabUnlinkMutation$variables;
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
    "name": "ApplicabilityStatementsTabUnlinkMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ApplicabilityStatementsTabUnlinkMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "18144c4ab5cfafdc156a523c47b67c40",
    "id": null,
    "metadata": {},
    "name": "ApplicabilityStatementsTabUnlinkMutation",
    "operationKind": "mutation",
    "text": "mutation ApplicabilityStatementsTabUnlinkMutation(\n  $input: DeleteApplicabilityStatementInput!\n) {\n  deleteApplicabilityStatement(input: $input) {\n    deletedApplicabilityStatementId\n  }\n}\n"
  }
};
})();

(node as any).hash = "30565a0f33a412db1e7c06db59cc096f";

export default node;
