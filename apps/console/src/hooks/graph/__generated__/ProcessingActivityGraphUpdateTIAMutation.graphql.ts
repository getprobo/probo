/**
 * @generated SignedSource<<a4fc26df235ee47b3cf31c1dc4ffb2c3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateProcessingActivityTIAInput = {
  dataSubjects?: string | null | undefined;
  id: string;
  legalMechanism?: string | null | undefined;
  localLawRisk?: string | null | undefined;
  supplementaryMeasures?: string | null | undefined;
  transfer?: string | null | undefined;
};
export type ProcessingActivityGraphUpdateTIAMutation$variables = {
  input: UpdateProcessingActivityTIAInput;
};
export type ProcessingActivityGraphUpdateTIAMutation$data = {
  readonly updateProcessingActivityTIA: {
    readonly processingActivityTia: {
      readonly createdAt: any;
      readonly dataSubjects: string | null | undefined;
      readonly id: string;
      readonly legalMechanism: string | null | undefined;
      readonly localLawRisk: string | null | undefined;
      readonly supplementaryMeasures: string | null | undefined;
      readonly transfer: string | null | undefined;
      readonly updatedAt: any;
    };
  };
};
export type ProcessingActivityGraphUpdateTIAMutation = {
  response: ProcessingActivityGraphUpdateTIAMutation$data;
  variables: ProcessingActivityGraphUpdateTIAMutation$variables;
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
    "concreteType": "UpdateProcessingActivityTIAPayload",
    "kind": "LinkedField",
    "name": "updateProcessingActivityTIA",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ProcessingActivityTIA",
        "kind": "LinkedField",
        "name": "processingActivityTia",
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
            "name": "dataSubjects",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "legalMechanism",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "transfer",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "localLawRisk",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "supplementaryMeasures",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
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
    "name": "ProcessingActivityGraphUpdateTIAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphUpdateTIAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "d596534ba9782fe47b15cd56e008db42",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphUpdateTIAMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphUpdateTIAMutation(\n  $input: UpdateProcessingActivityTIAInput!\n) {\n  updateProcessingActivityTIA(input: $input) {\n    processingActivityTia {\n      id\n      dataSubjects\n      legalMechanism\n      transfer\n      localLawRisk\n      supplementaryMeasures\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "a669ab652dd3687dff4069814b3ea0e8";

export default node;
