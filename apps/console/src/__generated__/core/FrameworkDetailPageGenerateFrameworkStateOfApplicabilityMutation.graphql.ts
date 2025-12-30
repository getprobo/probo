/**
 * @generated SignedSource<<b7c54e98eb3ba12bf4619530b45fdbab>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation$variables = {
  frameworkId: string;
};
export type FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation$data = {
  readonly generateFrameworkStateOfApplicability: {
    readonly data: string;
  };
};
export type FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation = {
  response: FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation$data;
  variables: FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "frameworkId"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "fields": [
          {
            "kind": "Variable",
            "name": "frameworkId",
            "variableName": "frameworkId"
          }
        ],
        "kind": "ObjectValue",
        "name": "input"
      }
    ],
    "concreteType": "GenerateFrameworkStateOfApplicabilityPayload",
    "kind": "LinkedField",
    "name": "generateFrameworkStateOfApplicability",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "data",
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
    "name": "FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "80185d474f3a9c5d20f97a387b219511",
    "id": null,
    "metadata": {},
    "name": "FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkDetailPageGenerateFrameworkStateOfApplicabilityMutation(\n  $frameworkId: ID!\n) {\n  generateFrameworkStateOfApplicability(input: {frameworkId: $frameworkId}) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "eac70164d676cb8e19d7fe7f06b2b817";

export default node;
