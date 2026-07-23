// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import * as accessReview from './accessReview';
import * as asset from './asset';
import * as audit from './audit';
import * as auditLog from './auditLog';
import * as businessFunction from './businessFunction';
import * as control from './control';
import * as cookieBanner from './cookieBanner';
import * as cookieCategory from './cookieCategory';
import * as cookieConsentRecord from './cookieConsentRecord';
import * as trackerPattern from './trackerPattern';
import * as datum from './datum';
import * as document from './document';
import * as dpia from './dpia';
import * as evidence from './evidence';
import * as execute from './execute';
import * as finding from './finding';
import * as framework from './framework';
import * as measure from './measure';
import * as obligation from './obligation';
import * as organization from './organization';
import * as organizationContext from './organizationContext';
import * as processingActivity from './processingActivity';
import * as rightsRequest from './rightsRequest';
import * as resourceAlias from './resourceAlias';
import * as riskAssessment from './riskAssessment';
import * as user from './user';
import * as risk from './risk';
import * as statementOfApplicability from './statementOfApplicability';
import * as task from './task';
import * as tia from './tia';
import * as compliancePortal from './compliancePortal';
import * as thirdParty from './thirdParty';
import * as webhook from './webhook';

export interface ResourceModule {
	description: INodeProperties[];
	[key: string]: OperationModule | INodeProperties[];
}

export interface OperationModule {
	description: INodeProperties[];
	execute: (this: IExecuteFunctions, itemIndex: number) => Promise<INodeExecutionData>;
}

export const resources: Record<string, ResourceModule> = {
	accessReview: accessReview as ResourceModule,
	asset: asset as ResourceModule,
	audit: audit as ResourceModule,
	auditLog: auditLog as ResourceModule,
	businessFunction: businessFunction as ResourceModule,
	control: control as ResourceModule,
	cookieBanner: cookieBanner as ResourceModule,
	cookieCategory: cookieCategory as ResourceModule,
	cookieConsentRecord: cookieConsentRecord as ResourceModule,
	trackerPattern: trackerPattern as ResourceModule,
	datum: datum as ResourceModule,
	document: document as ResourceModule,
	dpia: dpia as ResourceModule,
	evidence: evidence as ResourceModule,
	execute: execute as ResourceModule,
	finding: finding as ResourceModule,
	framework: framework as ResourceModule,
	measure: measure as ResourceModule,
	obligation: obligation as ResourceModule,
	organization: organization as ResourceModule,
	organizationContext: organizationContext as ResourceModule,
	processingActivity: processingActivity as ResourceModule,
	rightsRequest: rightsRequest as ResourceModule,
	resourceAlias: resourceAlias as ResourceModule,
	riskAssessment: riskAssessment as ResourceModule,
	user: user as ResourceModule,
	risk: risk as ResourceModule,
	statementOfApplicability: statementOfApplicability as ResourceModule,
	task: task as ResourceModule,
	tia: tia as ResourceModule,
	compliancePortal: compliancePortal as ResourceModule,
	thirdParty: thirdParty as ResourceModule,
	webhook: webhook as ResourceModule,
};

export function getAllResourceOperations(): INodeProperties[] {
	const operations: INodeProperties[] = [];

	for (const resource of Object.values(resources)) {
		const operationProp = resource.description.find((prop) => prop.name === 'operation');
		if (operationProp) {
			operations.push(operationProp);
		}
	}

	return operations;
}

export function getAllResourceFields(): INodeProperties[] {
	const fields: INodeProperties[] = [];

	for (const resource of Object.values(resources)) {
		fields.push(...resource.description.filter((prop) => prop.name !== 'operation'));
	}

	return fields;
}

export function getExecuteFunction(resourceName: string, operationName: string) {
	const resource = resources[resourceName];
	if (!resource) {
		throw new Error(`Unknown resource: ${resourceName}`);
	}

	const operationKey = resourceName === 'execute' ? 'execute' : operationName;

	const operation = resource[operationKey] as OperationModule;

	if (!operation || typeof operation.execute !== 'function') {
		throw new Error(`Unknown operation: ${operationName} for resource: ${resourceName}`);
	}

	return operation.execute;
}
