/**
 * @generated SignedSource<<23a875a3299a51bdebef4d121b0b84ff>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateComplianceExternalURLInput = {
  name: string;
  trustCenterId: string;
  url: string;
};
export type CompliancePageExternalUrlsSection_createMutation$variables = {
  input: CreateComplianceExternalURLInput;
};
export type CompliancePageExternalUrlsSection_createMutation$data = {
  readonly createComplianceExternalURL: {
    readonly complianceExternalUrlEdge: {
      readonly node: {
        readonly id: string;
        readonly name: string;
        readonly rank: number;
        readonly url: string;
      };
    };
  };
};
export type CompliancePageExternalUrlsSection_createMutation = {
  response: CompliancePageExternalUrlsSection_createMutation$data;
  variables: CompliancePageExternalUrlsSection_createMutation$variables;
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
    "concreteType": "CreateComplianceExternalURLPayload",
    "kind": "LinkedField",
    "name": "createComplianceExternalURL",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ComplianceExternalURLEdge",
        "kind": "LinkedField",
        "name": "complianceExternalUrlEdge",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ComplianceExternalURL",
            "kind": "LinkedField",
            "name": "node",
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
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageExternalUrlsSection_createMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageExternalUrlsSection_createMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "946285e4ff57ca46030c3c2ec5b8cc22",
    "id": null,
    "metadata": {},
    "name": "CompliancePageExternalUrlsSection_createMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageExternalUrlsSection_createMutation(\n  $input: CreateComplianceExternalURLInput!\n) {\n  createComplianceExternalURL(input: $input) {\n    complianceExternalUrlEdge {\n      node {\n        id\n        name\n        url\n        rank\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8441156775b5070a2dc3f0e316d5bb0b";

export default node;
