/**
 * @generated SignedSource<<4f27f16c09a86fda0f17b4d33866fcb2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteCustomDomainInput = {
  organizationId: string;
};
export type DeleteDomainDialogMutation$variables = {
  input: DeleteCustomDomainInput;
};
export type DeleteDomainDialogMutation$data = {
  readonly deleteCustomDomain: {
    readonly deletedCustomDomainId: string;
  };
};
export type DeleteDomainDialogMutation = {
  response: DeleteDomainDialogMutation$data;
  variables: DeleteDomainDialogMutation$variables;
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
    "concreteType": "DeleteCustomDomainPayload",
    "kind": "LinkedField",
    "name": "deleteCustomDomain",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedCustomDomainId",
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
    "name": "DeleteDomainDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DeleteDomainDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "68e7fb0109af8d99c9947e5ef2314ad2",
    "id": null,
    "metadata": {},
    "name": "DeleteDomainDialogMutation",
    "operationKind": "mutation",
    "text": "mutation DeleteDomainDialogMutation(\n  $input: DeleteCustomDomainInput!\n) {\n  deleteCustomDomain(input: $input) {\n    deletedCustomDomainId\n  }\n}\n"
  }
};
})();

(node as any).hash = "4ff39445b000bd5416e04e1460f11571";

export default node;
