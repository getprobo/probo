/**
 * @generated SignedSource<<12b5695d6b5f46a7d4231abe93301a0b>>
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
export type DeleteCompliancePageDomainDialogMutation$variables = {
  input: DeleteCustomDomainInput;
};
export type DeleteCompliancePageDomainDialogMutation$data = {
  readonly deleteCustomDomain: {
    readonly deletedCustomDomainId: string;
  };
};
export type DeleteCompliancePageDomainDialogMutation = {
  response: DeleteCompliancePageDomainDialogMutation$data;
  variables: DeleteCompliancePageDomainDialogMutation$variables;
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
    "name": "DeleteCompliancePageDomainDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DeleteCompliancePageDomainDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "1d5e88cc48c7ce90d9110de6d4c6a4d9",
    "id": null,
    "metadata": {},
    "name": "DeleteCompliancePageDomainDialogMutation",
    "operationKind": "mutation",
    "text": "mutation DeleteCompliancePageDomainDialogMutation(\n  $input: DeleteCustomDomainInput!\n) {\n  deleteCustomDomain(input: $input) {\n    deletedCustomDomainId\n  }\n}\n"
  }
};
})();

(node as any).hash = "d65e8b9b66367509071b9349c828f5d0";

export default node;
