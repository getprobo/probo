/**
 * @generated SignedSource<<f8aca586d6c3f69c7b0c5d485bdcc101>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateComplianceExternalURLInput = {
  id: string;
  name: string;
  rank?: number | null | undefined;
  url: string;
};
export type CompliancePageExternalUrlsSection_updateMutation$variables = {
  input: UpdateComplianceExternalURLInput;
};
export type CompliancePageExternalUrlsSection_updateMutation$data = {
  readonly updateComplianceExternalURL: {
    readonly complianceExternalUrl: {
      readonly id: string;
      readonly name: string;
      readonly rank: number;
      readonly url: string;
    };
  };
};
export type CompliancePageExternalUrlsSection_updateMutation = {
  response: CompliancePageExternalUrlsSection_updateMutation$data;
  variables: CompliancePageExternalUrlsSection_updateMutation$variables;
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
    "concreteType": "UpdateComplianceExternalURLPayload",
    "kind": "LinkedField",
    "name": "updateComplianceExternalURL",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ComplianceExternalURL",
        "kind": "LinkedField",
        "name": "complianceExternalUrl",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "url",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "rank",
            "storageKey": null
          }
        ],
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
    "name": "CompliancePageExternalUrlsSection_updateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageExternalUrlsSection_updateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "f771e54a9a15080525576308c8277815",
    "id": null,
    "metadata": {},
    "name": "CompliancePageExternalUrlsSection_updateMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageExternalUrlsSection_updateMutation(\n  $input: UpdateComplianceExternalURLInput!\n) {\n  updateComplianceExternalURL(input: $input) {\n    complianceExternalUrl {\n      id\n      name\n      url\n      rank\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b6162bd7b54e784dff62db83e0ff32be";

export default node;
