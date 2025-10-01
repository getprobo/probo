/**
 * @generated SignedSource<<b25e2b88094ba632cc451508af4df1e3>>
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
export type CustomDomainManagerDeleteMutation$variables = {
  input: DeleteCustomDomainInput;
};
export type CustomDomainManagerDeleteMutation$data = {
  readonly deleteCustomDomain: {
    readonly deletedCustomDomainId: string;
  };
};
export type CustomDomainManagerDeleteMutation = {
  response: CustomDomainManagerDeleteMutation$data;
  variables: CustomDomainManagerDeleteMutation$variables;
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
    "name": "CustomDomainManagerDeleteMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CustomDomainManagerDeleteMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "9f4a3ed02de61bc60a8370e3771ca7bd",
    "id": null,
    "metadata": {},
    "name": "CustomDomainManagerDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation CustomDomainManagerDeleteMutation(\n  $input: DeleteCustomDomainInput!\n) {\n  deleteCustomDomain(input: $input) {\n    deletedCustomDomainId\n  }\n}\n"
  }
};
})();

(node as any).hash = "e3878d11e361c3da2471363664150f3d";

export default node;
