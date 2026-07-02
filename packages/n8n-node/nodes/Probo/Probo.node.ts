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

import {
	NodeConnectionTypes,
	NodeOperationError,
	type IExecuteFunctions,
	type INodeExecutionData,
	type INodeType,
	type INodeTypeDescription,
} from 'n8n-workflow';
import {
	getAllResourceOperations,
	getAllResourceFields,
	getExecuteFunction,
} from './actions';

export class Probo implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Probo',
		name: 'probo',
		icon: { light: 'file:../../icons/probo-light.svg', dark: 'file:../../icons/probo.svg' },
		group: ['input'],
		version: 1,
		subtitle: '={{$parameter["resource"]}} / {{$parameter["operation"]}}',
		description: 'Consume data from the Probo API',
		defaults: {
			name: 'Probo',
		},
		usableAsTool: true,
		inputs: [NodeConnectionTypes.Main],
		outputs: [NodeConnectionTypes.Main],
		credentials: [
			{
				name: 'proboApi',
				required: true,
				displayOptions: {
					show: {
						authentication: ['apiKey'],
					},
				},
			},
		],
		requestDefaults: {
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json',
			},
		},
		properties: [
			{
				displayName: 'Authentication',
				name: 'authentication',
				type: 'options',
				options: [
					{
						name: 'API Key',
						value: 'apiKey',
					},
				],
				default: 'apiKey',
			},
			{
				displayName: 'Resource',
				name: 'resource',
				type: 'options',
				noDataExpression: true,
				options: [
					{
						name: 'Access Review',
						value: 'accessReview',
						description: 'Manage access review campaigns',
					},
					{
						name: 'Asset',
						value: 'asset',
						description: 'Manage assets',
					},
					{
						name: 'Audit',
						value: 'audit',
						description: 'Manage audits',
					},
					{
						name: 'Audit Log',
						value: 'auditLog',
						description: 'View audit log entries',
					},
					{
						name: 'Control',
						value: 'control',
						description: 'Manage controls',
					},
					{
						name: 'Cookie Banner',
						value: 'cookieBanner',
						description: 'Manage cookie banners',
					},
					{
						name: 'Cookie Category',
						value: 'cookieCategory',
						description: 'Manage cookie categories',
					},
					{
						name: 'Cookie Consent Record',
						value: 'cookieConsentRecord',
						description: 'View cookie consent records',
					},
					{
						name: 'Data',
						value: 'datum',
						description: 'Manage data',
					},
					{
						name: 'Document',
						value: 'document',
						description: 'Manage documents, versions, and signatures',
					},
					{
						name: 'DPIA',
						value: 'dpia',
						description: 'Manage data protection impact assessments',
					},
					{
						name: 'Evidence',
						value: 'evidence',
						description: 'Manage evidences',
					},
					{
						name: 'Execute',
						value: 'execute',
						description: 'Execute a GraphQL query or mutation',
					},
					{
						name: 'Finding',
						value: 'finding',
						description: 'Manage findings',
					},
					{
						name: 'Framework',
						value: 'framework',
						description: 'Manage frameworks',
					},
					{
						name: 'Measure',
						value: 'measure',
						description: 'Manage measures',
					},
					{
						name: 'Obligation',
						value: 'obligation',
						description: 'Manage obligations',
					},
					{
						name: 'Organization',
						value: 'organization',
						description: 'Manage organizations',
					},
					{
						name: 'Organization Context',
						value: 'organizationContext',
						description: 'Manage organization context',
					},
					{
						name: 'Processing Activity',
						value: 'processingActivity',
						description: 'Manage processing activities',
					},
					{
						name: 'Resource Alias',
						value: 'resourceAlias',
						description: 'Manage resource aliases',
					},
					{
						name: 'Rights Request',
						value: 'rightsRequest',
						description: 'Manage rights requests',
					},
					{
						name: 'Risk',
						value: 'risk',
						description: 'Manage risks',
					},
					{
						name: 'Risk Assessment',
						value: 'riskAssessment',
						description: 'Manage risk assessments',
					},
					{
						name: 'Statement of Applicability',
						value: 'statementOfApplicability',
						description: 'Manage statements of applicability',
					},
					{
						name: 'Task',
						value: 'task',
						description: 'Manage tasks',
					},
					{
						name: 'Third Party',
						value: 'thirdParty',
						description: 'Manage third parties',
					},
					{
						name: 'TIA',
						value: 'tia',
						description: 'Manage transfer impact assessments',
					},
					{
						name: 'Tracker Pattern',
						value: 'trackerPattern',
						description: 'Manage tracker patterns',
					},
					{
						name: 'Trust Center',
						value: 'trustCenter',
						description: 'Manage trust center',
					},
					{
						name: 'User',
						value: 'user',
						description: 'Manage organization users (profiles)',
					},
					{
						name: 'Webhook',
						value: 'webhook',
						description: 'Manage webhook subscriptions',
					},
				],
				default: 'execute',
			},
			...getAllResourceOperations(),
			...getAllResourceFields(),
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const returnData: INodeExecutionData[] = [];

		for (let i = 0; i < items.length; i++) {
			try {
				const resource = this.getNodeParameter('resource', i) as string;
				const operation = this.getNodeParameter('operation', i, 'execute') as string;

				const executeFunction = getExecuteFunction(resource, operation);
				const result = await executeFunction.call(this, i);
				returnData.push(result);
			} catch (error) {
				if (this.continueOnFail()) {
					returnData.push({
						json: { error: error instanceof Error ? error.message : String(error) },
						pairedItem: { item: i },
					});
					continue;
				}
				throw new NodeOperationError(this.getNode(), error as Error, { itemIndex: i });
			}
		}

		return [returnData];
	}

	methods = {
		listSearch: {},
	};
}
