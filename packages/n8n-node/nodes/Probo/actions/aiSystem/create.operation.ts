// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['aiSystem'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['aiSystem'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the AI system',
		required: true,
	},
	{
		displayName: 'Status',
		name: 'status',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['aiSystem'],
				operation: ['create'],
			},
		},
		options: [
			{ name: 'Active', value: 'ACTIVE' },
			{ name: 'In Development', value: 'IN_DEVELOPMENT' },
			{ name: 'Decommissioned', value: 'DECOMMISSIONED' },
		],
		default: 'IN_DEVELOPMENT',
		description: 'The status of the AI system',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['aiSystem'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Autonomy Level',
				name: 'autonomyLevel',
				type: 'string',
				default: '',
				description: 'The autonomy level of the AI system',
			},
			{
				displayName: 'Company Roles',
				name: 'companyRoles',
				type: 'string',
				default: '',
				description: 'Comma-separated company roles (PROVIDER, DEPLOYER, USER, DEVELOPER)',
			},
			{
				displayName: 'Data Sources And Type',
				name: 'dataSourcesAndType',
				type: 'string',
				default: '',
				description: 'Data sources and types used by the AI system',
			},
			{
				displayName: 'Deployment Date',
				name: 'deploymentDate',
				type: 'string',
				default: '',
				description: 'Deployment date (ISO 8601 datetime)',
			},
			{
				displayName: 'Human Oversight Mechanism',
				name: 'humanOversightMechanism',
				type: 'string',
				default: '',
				description: 'Human oversight mechanism description',
			},
			{
				displayName: 'Intended Use Cases',
				name: 'intendedUseCases',
				type: 'string',
				default: '',
				description: 'Intended use cases for the AI system',
			},
			{
				displayName: 'Key Stakeholders',
				name: 'keyStakeholders',
				type: 'string',
				default: '',
				description: 'Key stakeholders involved with the AI system',
			},
			{
				displayName: 'Last Review Date',
				name: 'lastReviewDate',
				type: 'string',
				default: '',
				description: 'Last review date (ISO 8601 datetime)',
			},
			{
				displayName: 'Next Review Date',
				name: 'nextReviewDate',
				type: 'string',
				default: '',
				description: 'Next review date (ISO 8601 datetime)',
			},
			{
				displayName: 'Notes',
				name: 'notes',
				type: 'string',
				default: '',
				description: 'Additional notes',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'Owner profile ID',
			},
			{
				displayName: 'Purpose',
				name: 'purpose',
				type: 'string',
				default: '',
				description: 'The purpose of the AI system',
			},
			{
				displayName: 'Risk Classification',
				name: 'riskClassification',
				type: 'string',
				default: '',
				description:
					'Risk classification (HIGH_RISK, LIMITED, MINIMAL, or GPAI)',
			},
			{
				displayName: 'Source',
				name: 'source',
				type: 'string',
				default: '',
				description: 'The source of the AI system',
			},
			{
				displayName: 'Version',
				name: 'version',
				type: 'string',
				default: '',
				description: 'The version of the AI system',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const status = this.getNodeParameter('status', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		version?: string;
		companyRoles?: string;
		ownerId?: string;
		source?: string;
		purpose?: string;
		intendedUseCases?: string;
		autonomyLevel?: string;
		humanOversightMechanism?: string;
		riskClassification?: string;
		keyStakeholders?: string;
		dataSourcesAndType?: string;
		deploymentDate?: string;
		lastReviewDate?: string;
		nextReviewDate?: string;
		notes?: string;
	};

	const query = `
		mutation CreateAiSystem($input: CreateAiSystemInput!) {
			createAiSystem(input: $input) {
				aiSystemEdge {
					node {
						id
						name
						version
						companyRoles
						status
						source
						purpose
						intendedUseCases
						autonomyLevel
						humanOversightMechanism
						riskClassification
						keyStakeholders
						dataSourcesAndType
						deploymentDate
						lastReviewDate
						nextReviewDate
						notes
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		name,
		status,
	};

	if (additionalFields.version) {
		input.version = additionalFields.version;
	}

	if (additionalFields.companyRoles) {
		input.companyRoles = additionalFields.companyRoles
			.split(',')
			.map((role) => role.trim())
			.filter(Boolean);
	}

	if (additionalFields.ownerId) {
		input.ownerId = additionalFields.ownerId;
	}

	if (additionalFields.source) {
		input.source = additionalFields.source;
	}

	if (additionalFields.purpose) {
		input.purpose = additionalFields.purpose;
	}

	if (additionalFields.intendedUseCases) {
		input.intendedUseCases = additionalFields.intendedUseCases;
	}

	if (additionalFields.autonomyLevel) {
		input.autonomyLevel = additionalFields.autonomyLevel;
	}

	if (additionalFields.humanOversightMechanism) {
		input.humanOversightMechanism = additionalFields.humanOversightMechanism;
	}

	if (additionalFields.riskClassification) {
		input.riskClassification = additionalFields.riskClassification;
	}

	if (additionalFields.keyStakeholders) {
		input.keyStakeholders = additionalFields.keyStakeholders;
	}

	if (additionalFields.dataSourcesAndType) {
		input.dataSourcesAndType = additionalFields.dataSourcesAndType;
	}

	if (additionalFields.deploymentDate) {
		input.deploymentDate = additionalFields.deploymentDate;
	}

	if (additionalFields.lastReviewDate) {
		input.lastReviewDate = additionalFields.lastReviewDate;
	}

	if (additionalFields.nextReviewDate) {
		input.nextReviewDate = additionalFields.nextReviewDate;
	}

	if (additionalFields.notes) {
		input.notes = additionalFields.notes;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
