/**
 * @generated SignedSource<<39982b95a577811bc7de528129004def>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type GenerateRisksInput = {
  organizationId: string;
};
export type ListRiskViewGenerateRisksQuery$variables = {
  input: GenerateRisksInput;
};
export type ListRiskViewGenerateRisksQuery$data = {
  readonly generateRisks: {
    readonly risks: ReadonlyArray<string>;
  };
};
export type ListRiskViewGenerateRisksQuery = {
  response: ListRiskViewGenerateRisksQuery$data;
  variables: ListRiskViewGenerateRisksQuery$variables;
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
    "concreteType": "GenerateRisksPayload",
    "kind": "LinkedField",
    "name": "generateRisks",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "risks",
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
    "name": "ListRiskViewGenerateRisksQuery",
    "selections": (v1/*: any*/),
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ListRiskViewGenerateRisksQuery",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "4a36a09c10dce9a9186685b591b22858",
    "id": null,
    "metadata": {},
    "name": "ListRiskViewGenerateRisksQuery",
    "operationKind": "query",
    "text": "query ListRiskViewGenerateRisksQuery(\n  $input: GenerateRisksInput!\n) {\n  generateRisks(input: $input) {\n    risks\n  }\n}\n"
  }
};
})();

(node as any).hash = "aeb21fee9261e027ceb4d45bbd9dfba4";

export default node;
