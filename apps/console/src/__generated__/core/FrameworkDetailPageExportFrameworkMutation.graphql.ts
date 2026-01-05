/**
 * @generated SignedSource<<05766c7969f5e78f91ca5121d2c30db8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type FrameworkDetailPageExportFrameworkMutation$variables = {
  frameworkId: string;
};
export type FrameworkDetailPageExportFrameworkMutation$data = {
  readonly exportFramework: {
    readonly exportJobId: string;
  };
};
export type FrameworkDetailPageExportFrameworkMutation = {
  response: FrameworkDetailPageExportFrameworkMutation$data;
  variables: FrameworkDetailPageExportFrameworkMutation$variables;
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
    "concreteType": "ExportFrameworkPayload",
    "kind": "LinkedField",
    "name": "exportFramework",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "exportJobId",
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
    "name": "FrameworkDetailPageExportFrameworkMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "FrameworkDetailPageExportFrameworkMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "d6fdd2bb2dacfc88248ccfc6f671eb4a",
    "id": null,
    "metadata": {},
    "name": "FrameworkDetailPageExportFrameworkMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkDetailPageExportFrameworkMutation(\n  $frameworkId: ID!\n) {\n  exportFramework(input: {frameworkId: $frameworkId}) {\n    exportJobId\n  }\n}\n"
  }
};
})();

(node as any).hash = "3866a75fb0aeaaa6cdf97934508116c1";

export default node;
