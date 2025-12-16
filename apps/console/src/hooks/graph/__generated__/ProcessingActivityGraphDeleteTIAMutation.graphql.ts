/**
 * @generated SignedSource<<0fa59581ecefda2fad4570e8e8c2f667>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteProcessingActivityTIAInput = {
  processingActivityTiaId: string;
};
export type ProcessingActivityGraphDeleteTIAMutation$variables = {
  input: DeleteProcessingActivityTIAInput;
};
export type ProcessingActivityGraphDeleteTIAMutation$data = {
  readonly deleteProcessingActivityTIA: {
    readonly deletedProcessingActivityTiaId: string;
  };
};
export type ProcessingActivityGraphDeleteTIAMutation = {
  response: ProcessingActivityGraphDeleteTIAMutation$data;
  variables: ProcessingActivityGraphDeleteTIAMutation$variables;
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
    "concreteType": "DeleteProcessingActivityTIAPayload",
    "kind": "LinkedField",
    "name": "deleteProcessingActivityTIA",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedProcessingActivityTiaId",
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
    "name": "ProcessingActivityGraphDeleteTIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphDeleteTIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "59c2e5775439e71d1c3546f094810389",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphDeleteTIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphDeleteTIAMutation(\n  $input: DeleteProcessingActivityTIAInput!\n) {\n  deleteProcessingActivityTIA(input: $input) {\n    deletedProcessingActivityTiaId\n  }\n}\n"
  }
};
})();

(node as any).hash = "358c23b60d25478b3a217ae16938caed";

export default node;
