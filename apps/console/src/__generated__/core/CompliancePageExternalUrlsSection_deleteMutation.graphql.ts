/**
 * @generated SignedSource<<b014bdb0fa2614aa7df1efb00c5efb3a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteComplianceExternalURLInput = {
  id: string;
};
export type CompliancePageExternalUrlsSection_deleteMutation$variables = {
  input: DeleteComplianceExternalURLInput;
};
export type CompliancePageExternalUrlsSection_deleteMutation$data = {
  readonly deleteComplianceExternalURL: {
    readonly deletedComplianceExternalUrlId: string;
  };
};
export type CompliancePageExternalUrlsSection_deleteMutation = {
  response: CompliancePageExternalUrlsSection_deleteMutation$data;
  variables: CompliancePageExternalUrlsSection_deleteMutation$variables;
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
    "concreteType": "DeleteComplianceExternalURLPayload",
    "kind": "LinkedField",
    "name": "deleteComplianceExternalURL",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedComplianceExternalUrlId",
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
    "name": "CompliancePageExternalUrlsSection_deleteMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageExternalUrlsSection_deleteMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "e0715ce1d37b8854b0eeea3e6214973c",
    "id": null,
    "metadata": {},
    "name": "CompliancePageExternalUrlsSection_deleteMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageExternalUrlsSection_deleteMutation(\n  $input: DeleteComplianceExternalURLInput!\n) {\n  deleteComplianceExternalURL(input: $input) {\n    deletedComplianceExternalUrlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "394380c11f7ee240360143f8f71f058c";

export default node;
