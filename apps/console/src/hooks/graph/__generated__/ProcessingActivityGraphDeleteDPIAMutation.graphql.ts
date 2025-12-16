/**
 * @generated SignedSource<<be806f3431452b1a7a4d4945c8eda79b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteProcessingActivityDPIAInput = {
  processingActivityDpiaId: string;
};
export type ProcessingActivityGraphDeleteDPIAMutation$variables = {
  input: DeleteProcessingActivityDPIAInput;
};
export type ProcessingActivityGraphDeleteDPIAMutation$data = {
  readonly deleteProcessingActivityDPIA: {
    readonly deletedProcessingActivityDpiaId: string;
  };
};
export type ProcessingActivityGraphDeleteDPIAMutation = {
  response: ProcessingActivityGraphDeleteDPIAMutation$data;
  variables: ProcessingActivityGraphDeleteDPIAMutation$variables;
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
    "concreteType": "DeleteProcessingActivityDPIAPayload",
    "kind": "LinkedField",
    "name": "deleteProcessingActivityDPIA",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedProcessingActivityDpiaId",
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
    "name": "ProcessingActivityGraphDeleteDPIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphDeleteDPIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "0c27b0098614053d1a252f87e75f8a24",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphDeleteDPIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphDeleteDPIAMutation(\n  $input: DeleteProcessingActivityDPIAInput!\n) {\n  deleteProcessingActivityDPIA(input: $input) {\n    deletedProcessingActivityDpiaId\n  }\n}\n"
  }
};
})();

(node as any).hash = "9fc639df8d4ca6c8856acd9326c92997";

export default node;
